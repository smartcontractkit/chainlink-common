package remote

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	registrypb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

var _ registry.CapabilitiesRegistry = (*capabilitiesRegistryClient)(nil)

type capabilitiesRegistryClient struct {
	lggr      logger.Logger
	transport Transport
	grpc      registrypb.CapabilitiesRegistryClient

	mu        sync.Mutex
	published map[string]io.Closer
}

func toDON(don *registrypb.DON) capabilities.DON {
	var members []p2ptypes.PeerID
	for _, m := range don.Members {
		members = append(members, p2ptypes.PeerID(m))
	}

	return capabilities.DON{
		ID:            don.Id,
		Name:          don.Name,
		Members:       members,
		F:             uint8(don.F),
		ConfigVersion: don.ConfigVersion,
		Families:      don.Families,
		Config:        don.Config,
	}
}

func toPbDON(don capabilities.DON) *registrypb.DON {
	membersBytes := make([][]byte, len(don.Members))
	for j, m := range don.Members {
		membersBytes[j] = m[:]
	}

	return &registrypb.DON{
		Id:            don.ID,
		Name:          don.Name,
		Members:       membersBytes,
		F:             uint32(don.F),
		ConfigVersion: don.ConfigVersion,
		Families:      don.Families,
		Config:        don.Config,
	}
}

func (cr *capabilitiesRegistryClient) LocalNode(ctx context.Context) (capabilities.Node, error) {
	res, err := cr.grpc.LocalNode(ctx, &emptypb.Empty{})
	if err != nil {
		return capabilities.Node{}, err
	}

	return cr.nodeFromNodeReply(res), nil
}

func (cr *capabilitiesRegistryClient) NodeByPeerID(ctx context.Context, peerID p2ptypes.PeerID) (capabilities.Node, error) {
	res, err := cr.grpc.NodeByPeerID(ctx, &registrypb.NodeRequest{PeerID: peerID[:]})
	if err != nil {
		return capabilities.Node{}, err
	}

	return cr.nodeFromNodeReply(res), nil
}

func (cr *capabilitiesRegistryClient) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	res, err := cr.grpc.DONsForCapability(ctx, &registrypb.DONForCapabilityRequest{CapabilityID: capabilityID})
	if err != nil {
		return nil, err
	}

	donsWithNodes := []capabilities.DONWithNodes{}
	for _, d := range res.Dons {
		don := toDON(d.Don)
		var nodes []capabilities.Node
		for _, n := range d.Nodes {
			nodes = append(nodes, cr.nodeFromNodeReply(n))
		}
		donsWithNodes = append(donsWithNodes, capabilities.DONWithNodes{
			DON:   don,
			Nodes: nodes,
		})
	}
	return donsWithNodes, nil
}

func (cr *capabilitiesRegistryClient) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	res, err := cr.grpc.DONByID(ctx, &registrypb.DONByIDRequest{DonID: donID})
	if err != nil {
		return capabilities.DON{}, err
	}
	return toDON(res.Don), nil
}

func (cr *capabilitiesRegistryClient) nodeFromNodeReply(nodeReply *registrypb.NodeReply) capabilities.Node {
	var pid *p2ptypes.PeerID
	if len(nodeReply.PeerID) > 0 {
		p := p2ptypes.PeerID(nodeReply.PeerID)
		pid = &p
	}

	cDONs := make([]capabilities.DON, len(nodeReply.CapabilityDONs))
	for i, don := range nodeReply.CapabilityDONs {
		cDONs[i] = toDON(don)
	}

	var signer32 [32]byte
	copy(signer32[:], nodeReply.Signer)
	var encryptionPublicKey32 [32]byte
	copy(encryptionPublicKey32[:], nodeReply.EncryptionPublicKey)
	return capabilities.Node{
		PeerID:              pid,
		NodeOperatorID:      nodeReply.NodeOperatorID,
		Signer:              signer32,
		EncryptionPublicKey: encryptionPublicKey32,
		WorkflowDON:         toDON(nodeReply.WorkflowDON),
		CapabilityDONs:      cDONs,
	}
}

func (cr *capabilitiesRegistryClient) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	res, err := cr.grpc.ConfigForCapability(ctx, &registrypb.ConfigForCapabilityRequest{
		CapabilityID: capabilityID,
		DonID:        donID,
	})
	if err != nil {
		return capabilities.CapabilityConfiguration{}, err
	}

	defaultConfig, err := values.FromMapValueProto(res.CapabilityConfig.DefaultConfig)
	if err != nil {
		return capabilities.CapabilityConfiguration{}, fmt.Errorf("could not decode default config: %w", err)
	}

	var methodConfig map[string]capabilities.CapabilityMethodConfig
	if res.CapabilityConfig.MethodConfigs != nil {
		methodConfig = make(map[string]capabilities.CapabilityMethodConfig, len(res.CapabilityConfig.MethodConfigs))
		for mName, mConfig := range res.CapabilityConfig.MethodConfigs {
			newCapCfg := capabilities.CapabilityMethodConfig{}
			switch mConfig.RemoteConfig.(type) {
			case *pb.CapabilityMethodConfig_RemoteTriggerConfig:
				newCapCfg.RemoteTriggerConfig = decodeRemoteTriggerConfig(mConfig.GetRemoteTriggerConfig())
			case *pb.CapabilityMethodConfig_RemoteExecutableConfig:
				newCapCfg.RemoteExecutableConfig = decodeRemoteExecutableConfig(mConfig.GetRemoteExecutableConfig())
			}
			if mConfig.AggregatorConfig != nil {
				newCapCfg.AggregatorConfig = &capabilities.AggregatorConfig{AggregatorType: capabilities.AggregatorType(mConfig.AggregatorConfig.AggregatorType)}
			}
			methodConfig[mName] = newCapCfg
		}
	}

	var ocr3Configs map[string]ocrtypes.ContractConfig
	if res.CapabilityConfig.Ocr3Configs != nil {
		ocr3Configs = make(map[string]ocrtypes.ContractConfig, len(res.CapabilityConfig.Ocr3Configs))
		for key, pbCfg := range res.CapabilityConfig.Ocr3Configs {
			ocr3Configs[key] = decodeOcr3Config(pbCfg)
		}
	}

	var oracleFactoryConfigs map[string]values.Map
	if res.CapabilityConfig.OracleFactoryConfigs != nil {
		oracleFactoryConfigs = make(map[string]values.Map, len(res.CapabilityConfig.OracleFactoryConfigs))
		for key, pbMap := range res.CapabilityConfig.OracleFactoryConfigs {
			m, err := values.FromMapValueProto(pbMap)
			if err != nil {
				return capabilities.CapabilityConfiguration{}, fmt.Errorf("could not decode oracle factory config for key %s: %w", key, err)
			}
			if m != nil {
				oracleFactoryConfigs[key] = *m
			}
		}
	}

	specConfig, err := values.FromMapValueProto(res.CapabilityConfig.SpecConfig)
	if err != nil {
		return capabilities.CapabilityConfiguration{}, fmt.Errorf("could not decode spec config: %w", err)
	}

	return capabilities.CapabilityConfiguration{
		DefaultConfig:          defaultConfig,
		CapabilityMethodConfig: methodConfig,
		LocalOnly:              res.CapabilityConfig.LocalOnly,
		Ocr3Configs:            ocr3Configs,
		OracleFactoryConfigs:   oracleFactoryConfigs,
		SpecConfig:             specConfig,
	}, nil
}

func decodeRemoteTriggerConfig(prtc *pb.RemoteTriggerConfig) *capabilities.RemoteTriggerConfig {
	remoteTriggerConfig := &capabilities.RemoteTriggerConfig{}
	remoteTriggerConfig.RegistrationRefresh = prtc.RegistrationRefresh.AsDuration()
	remoteTriggerConfig.RegistrationExpiry = prtc.RegistrationExpiry.AsDuration()
	remoteTriggerConfig.MinResponsesToAggregate = prtc.MinResponsesToAggregate
	remoteTriggerConfig.MessageExpiry = prtc.MessageExpiry.AsDuration()
	remoteTriggerConfig.MaxBatchSize = prtc.MaxBatchSize
	remoteTriggerConfig.BatchCollectionPeriod = prtc.BatchCollectionPeriod.AsDuration()
	return remoteTriggerConfig
}

func decodeRemoteExecutableConfig(prtc *pb.RemoteExecutableConfig) *capabilities.RemoteExecutableConfig {
	remoteExecutableConfig := &capabilities.RemoteExecutableConfig{}
	remoteExecutableConfig.TransmissionSchedule = capabilities.TransmissionSchedule(prtc.TransmissionSchedule)
	remoteExecutableConfig.DeltaStage = prtc.DeltaStage.AsDuration()
	remoteExecutableConfig.RequestTimeout = prtc.RequestTimeout.AsDuration()
	remoteExecutableConfig.ServerMaxParallelRequests = prtc.ServerMaxParallelRequests
	remoteExecutableConfig.RequestHasherType = capabilities.RequestHasherType(prtc.RequestHasherType)
	remoteExecutableConfig.MinResponsesToAggregate = prtc.MinResponsesToAggregate
	return remoteExecutableConfig
}

func decodeOcr3Config(pbCfg *pb.OCR3Config) ocrtypes.ContractConfig {
	signers := make([]ocrtypes.OnchainPublicKey, len(pbCfg.Signers))
	for i, s := range pbCfg.Signers {
		signers[i] = ocrtypes.OnchainPublicKey(s)
	}
	transmitters := make([]ocrtypes.Account, len(pbCfg.Transmitters))
	for i, t := range pbCfg.Transmitters {
		transmitters[i] = ocrtypes.Account(hex.EncodeToString(t))
	}
	return ocrtypes.ContractConfig{
		ConfigCount:           pbCfg.ConfigCount,
		Signers:               signers,
		Transmitters:          transmitters,
		F:                     uint8(pbCfg.F),
		OnchainConfig:         pbCfg.OnchainConfig,
		OffchainConfigVersion: pbCfg.OffchainConfigVersion,
		OffchainConfig:        pbCfg.OffchainConfig,
		// NOTE: ConfigDigest will be appended later by ContractConfigTracker.
	}
}

func (cr *capabilitiesRegistryClient) Get(ctx context.Context, ID string) (capabilities.BaseCapability, error) {
	req := &registrypb.GetRequest{
		Id: ID,
	}

	conn, err := cr.transport.Dial(ctx, "Capability", func(ctx context.Context) (Locator, error) {
		res, err := cr.grpc.Get(ctx, req)
		if err != nil {
			return Locator{}, err
		}
		return Locator{Target: res.Target, LegacyHandle: res.CapabilityID}, nil
	})
	if err != nil {
		return nil, err
	}

	client := NewBaseCapabilityClient(cr.lggr, conn)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_, err = client.Info(ctx) // ensure the capability is reachable, with a reduced timeout
	return client, err
}

func (cr *capabilitiesRegistryClient) GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error) {
	req := &registrypb.GetTriggerRequest{
		Id: ID,
	}

	conn, err := cr.transport.Dial(ctx, "Trigger", func(ctx context.Context) (Locator, error) {
		res, err := cr.grpc.GetTrigger(ctx, req)
		if err != nil {
			return Locator{}, err
		}
		return Locator{Target: res.Target, LegacyHandle: res.CapabilityID}, nil
	})
	if err != nil {
		return nil, err
	}

	client := NewTriggerCapabilityClient(cr.lggr, conn)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_, err = client.Info(ctx) // ensure the capability is reachable, with a reduced timeout
	return client, err
}

func (cr *capabilitiesRegistryClient) GetExecutable(ctx context.Context, ID string) (capabilities.ExecutableCapability, error) {
	req := &registrypb.GetExecutableRequest{
		Id: ID,
	}

	conn, err := cr.transport.Dial(ctx, "Executable", func(ctx context.Context) (Locator, error) {
		res, err := cr.grpc.GetExecutable(ctx, req)
		if err != nil {
			return Locator{}, err
		}
		return Locator{Target: res.Target, LegacyHandle: res.CapabilityID}, nil
	})
	if err != nil {
		return nil, err
	}

	client := NewExecutableCapabilityClient(cr.lggr, conn)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_, err = client.Info(ctx) // ensure the capability is reachable, with a reduced timeout
	return client, err
}

func (cr *capabilitiesRegistryClient) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	res, err := cr.grpc.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var clients []capabilities.BaseCapability
	for i, loc := range locatorsFromListReply(res) {
		conn, err := cr.transport.Dial(ctx, fmt.Sprintf("List[%d]", i), func(context.Context) (Locator, error) {
			return loc, nil
		})
		if err != nil {
			return nil, err
		}
		clients = append(clients, NewBaseCapabilityClient(cr.lggr, conn))
	}

	return clients, nil
}

func (cr *capabilitiesRegistryClient) Add(ctx context.Context, c capabilities.BaseCapability) error {
	info, err := c.Info(ctx)
	if err != nil {
		return err
	}

	loc, cRes, err := cr.transport.Publish(info.ID, func(s *grpc.Server) {
		RegisterCapabilityServer(s, cr.lggr, c, info.CapabilityType)
	})
	if err != nil {
		return err
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	_, err = cr.grpc.Add(ctx, &registrypb.AddRequest{
		CapabilityID: loc.LegacyHandle,
		Target:       loc.Target,
		Type:         ExecuteAPITypeFor(info.CapabilityType),
	})
	if err != nil {
		cRes.Close()
		return err
	}

	cr.published[info.ID] = cRes

	return nil
}

func (cr *capabilitiesRegistryClient) Remove(ctx context.Context, ID string) error {
	req := &registrypb.RemoveRequest{
		Id: ID,
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()
	_, err := cr.grpc.Remove(ctx, req)
	if err != nil {
		return err
	}

	res, ok := cr.published[ID]
	delete(cr.published, ID)

	if ok {
		return res.Close()
	}
	return nil
}

// NewCapabilitiesRegistryClient returns a CapabilitiesRegistry backed by the service
// on cc. Capabilities it resolves are reached over t, which also publishes the local
// capabilities handed to Add.
func NewCapabilitiesRegistryClient(lggr logger.Logger, cc grpc.ClientConnInterface, t Transport) registry.CapabilitiesRegistry {
	return &capabilitiesRegistryClient{
		lggr:      logger.Named(lggr, "CapabilitiesRegistryClient"),
		transport: t,
		grpc:      registrypb.NewCapabilitiesRegistryClient(cc),
		published: map[string]io.Closer{},
	}
}
