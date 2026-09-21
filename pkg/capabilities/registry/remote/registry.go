package remote

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	registrypb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

var _ registry.CapabilitiesRegistry = (*capabilitiesRegistryClient)(nil)

type capabilitiesRegistryClient struct {
	lggr      logger.Logger
	transport Transport
	grpc      registrypb.CapabilitiesRegistryClient
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

	_, err = cr.grpc.Add(ctx, &registrypb.AddRequest{
		CapabilityID: loc.LegacyHandle,
		Target:       loc.Target,
		Type:         ExecuteAPITypeFor(info.CapabilityType),
	})
	if err != nil {
		cRes.Close()
		return err
	}
	return nil
}

func (cr *capabilitiesRegistryClient) Remove(ctx context.Context, ID string) error {
	req := &registrypb.RemoveRequest{
		Id: ID,
	}

	_, err := cr.grpc.Remove(ctx, req)
	if err != nil {
		return err
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
	}
}

var _ registrypb.CapabilitiesRegistryServer = (*capabilitiesRegistryServer)(nil)

type capabilitiesRegistryServer struct {
	registrypb.UnimplementedCapabilitiesRegistryServer
	lggr      logger.Logger
	transport Transport
	impl      registry.CapabilitiesRegistry
}

func (c *capabilitiesRegistryServer) Get(ctx context.Context, request *registrypb.GetRequest) (*registrypb.GetReply, error) {
	capability, err := c.impl.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	info, err := capability.Info(ctx)
	if err != nil {
		return nil, err
	}

	loc, _, err := c.transport.Publish("Get", func(s *grpc.Server) {
		RegisterCapabilityServer(s, c.lggr, capability, info.CapabilityType)
	})
	if err != nil {
		return nil, err
	}

	return &registrypb.GetReply{
		CapabilityID: loc.LegacyHandle,
		Target:       loc.Target,
		Type:         ExecuteAPITypeFor(info.CapabilityType),
	}, nil
}

func (c *capabilitiesRegistryServer) ConfigForCapability(ctx context.Context, req *registrypb.ConfigForCapabilityRequest) (*registrypb.ConfigForCapabilityReply, error) {
	cc, err := c.impl.ConfigForCapability(ctx, req.CapabilityID, req.DonID)
	if err != nil {
		return nil, err
	}

	ccp := &pb.CapabilityConfig{}

	if cc.DefaultConfig != nil {
		ccp.DefaultConfig = values.Proto(cc.DefaultConfig).GetMapValue()
	}

	// Handle method configs
	if cc.CapabilityMethodConfig != nil {
		ccp.MethodConfigs = make(map[string]*pb.CapabilityMethodConfig, len(cc.CapabilityMethodConfig))
		for mName, mConfig := range cc.CapabilityMethodConfig {
			pbMethodConfig := &pb.CapabilityMethodConfig{}

			// Handle remote trigger config for method
			if mConfig.RemoteTriggerConfig != nil {
				pbMethodConfig.RemoteConfig = &pb.CapabilityMethodConfig_RemoteTriggerConfig{
					RemoteTriggerConfig: &pb.RemoteTriggerConfig{
						RegistrationRefresh:     durationpb.New(mConfig.RemoteTriggerConfig.RegistrationRefresh),
						RegistrationExpiry:      durationpb.New(mConfig.RemoteTriggerConfig.RegistrationExpiry),
						MinResponsesToAggregate: mConfig.RemoteTriggerConfig.MinResponsesToAggregate,
						MessageExpiry:           durationpb.New(mConfig.RemoteTriggerConfig.MessageExpiry),
						MaxBatchSize:            mConfig.RemoteTriggerConfig.MaxBatchSize,
						BatchCollectionPeriod:   durationpb.New(mConfig.RemoteTriggerConfig.BatchCollectionPeriod),
					},
				}
			}

			// Handle remote executable config for method
			if mConfig.RemoteExecutableConfig != nil {
				pbMethodConfig.RemoteConfig = &pb.CapabilityMethodConfig_RemoteExecutableConfig{
					RemoteExecutableConfig: &pb.RemoteExecutableConfig{
						TransmissionSchedule:      pb.TransmissionSchedule(mConfig.RemoteExecutableConfig.TransmissionSchedule),
						DeltaStage:                durationpb.New(mConfig.RemoteExecutableConfig.DeltaStage),
						RequestTimeout:            durationpb.New(mConfig.RemoteExecutableConfig.RequestTimeout),
						ServerMaxParallelRequests: mConfig.RemoteExecutableConfig.ServerMaxParallelRequests,
						RequestHasherType:         pb.RequestHasherType(mConfig.RemoteExecutableConfig.RequestHasherType),
						MinResponsesToAggregate:   mConfig.RemoteExecutableConfig.MinResponsesToAggregate,
					},
				}
			}

			// Handle aggregator config for method
			if mConfig.AggregatorConfig != nil {
				pbMethodConfig.AggregatorConfig = &pb.AggregatorConfig{
					AggregatorType: pb.AggregatorType(mConfig.AggregatorConfig.AggregatorType),
				}
			}

			ccp.MethodConfigs[mName] = pbMethodConfig
		}
	}

	ccp.LocalOnly = cc.LocalOnly

	// Handle OCR3 configs
	if cc.Ocr3Configs != nil {
		ccp.Ocr3Configs = make(map[string]*pb.OCR3Config, len(cc.Ocr3Configs))
		for key, cfg := range cc.Ocr3Configs {
			signers := make([][]byte, len(cfg.Signers))
			for i, s := range cfg.Signers {
				signers[i] = []byte(s)
			}
			transmitters := make([][]byte, len(cfg.Transmitters))
			for i, t := range cfg.Transmitters {
				transmitters[i], err = hex.DecodeString(string(t))
				if err != nil {
					return nil, fmt.Errorf("failed to decode transmitter: %w", err)
				}
			}
			ccp.Ocr3Configs[key] = &pb.OCR3Config{
				ConfigCount:           cfg.ConfigCount,
				Signers:               signers,
				Transmitters:          transmitters,
				F:                     uint32(cfg.F),
				OnchainConfig:         cfg.OnchainConfig,
				OffchainConfigVersion: cfg.OffchainConfigVersion,
				OffchainConfig:        cfg.OffchainConfig,
				// NOTE: ConfigDigest is not passed in the proto, nor stored directly onchain.
			}
		}
	}

	// Handle Oracle factory configs
	if cc.OracleFactoryConfigs != nil {
		ccp.OracleFactoryConfigs = make(map[string]*valuespb.Map, len(cc.OracleFactoryConfigs))
		for key, m := range cc.OracleFactoryConfigs {
			ccp.OracleFactoryConfigs[key] = values.Proto(&m).GetMapValue()
		}
	}

	// Handle Spec config
	if cc.SpecConfig != nil {
		ccp.SpecConfig = values.Proto(cc.SpecConfig).GetMapValue()
	}

	return &registrypb.ConfigForCapabilityReply{
		CapabilityConfig: ccp,
	}, nil
}

func (c *capabilitiesRegistryServer) LocalNode(ctx context.Context, _ *emptypb.Empty) (*registrypb.NodeReply, error) {
	node, err := c.impl.LocalNode(ctx)
	if err != nil {
		return nil, err
	}

	return c.nodeReplyFromNode(node), nil
}

func (c *capabilitiesRegistryServer) NodeByPeerID(ctx context.Context, nodeRequest *registrypb.NodeRequest) (*registrypb.NodeReply, error) {
	node, err := c.impl.NodeByPeerID(ctx, p2ptypes.PeerID(nodeRequest.GetPeerID()))
	if err != nil {
		return nil, err
	}

	return c.nodeReplyFromNode(node), nil
}

func (c *capabilitiesRegistryServer) DONsForCapability(ctx context.Context, req *registrypb.DONForCapabilityRequest) (*registrypb.DONForCapabilityReply, error) {
	dons, err := c.impl.DONsForCapability(ctx, req.CapabilityID)
	if err != nil {
		return nil, err
	}

	donWithNodes := []*registrypb.DONWithNodes{}
	for _, d := range dons {
		pbDon := toPbDON(d.DON)
		nodes := []*registrypb.NodeReply{}
		for _, n := range d.Nodes {
			nodes = append(nodes, c.nodeReplyFromNode(n))
		}
		donWithNodes = append(donWithNodes, &registrypb.DONWithNodes{
			Don:   pbDon,
			Nodes: nodes,
		})
	}

	return &registrypb.DONForCapabilityReply{
		Dons: donWithNodes,
	}, nil
}

func (c *capabilitiesRegistryServer) DONByID(ctx context.Context, req *registrypb.DONByIDRequest) (*registrypb.DONByIDReply, error) {
	don, err := c.impl.DONByID(ctx, req.DonID)
	if err != nil {
		return nil, err
	}
	return &registrypb.DONByIDReply{Don: toPbDON(don)}, nil
}

func (c *capabilitiesRegistryServer) nodeReplyFromNode(node capabilities.Node) *registrypb.NodeReply {
	workflowDONpb := toPbDON(node.WorkflowDON)

	capabilityDONsPb := make([]*registrypb.DON, len(node.CapabilityDONs))
	for i, don := range node.CapabilityDONs {
		capabilityDONsPb[i] = toPbDON(don)
	}

	var pid []byte
	if node.PeerID != nil {
		pid = node.PeerID[:]
	}
	reply := &registrypb.NodeReply{
		PeerID:              pid,
		NodeOperatorID:      node.NodeOperatorID,
		Signer:              node.Signer[:],
		EncryptionPublicKey: node.EncryptionPublicKey[:],
		WorkflowDON:         workflowDONpb,
		CapabilityDONs:      capabilityDONsPb,
	}

	return reply
}

func (c *capabilitiesRegistryServer) GetTrigger(ctx context.Context, request *registrypb.GetTriggerRequest) (*registrypb.GetTriggerReply, error) {
	capability, err := c.impl.GetTrigger(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	info, err := capability.Info(ctx)
	if err != nil {
		return nil, err
	}

	switch info.CapabilityType {
	case capabilities.CapabilityTypeTrigger, capabilities.CapabilityTypeCombined:
	default:
		return nil, fmt.Errorf("capability with id: %s does not satisfy the capability interface", request.Id)
	}

	loc, _, err := c.transport.Publish("GetTrigger", func(s *grpc.Server) {
		RegisterCapabilityServer(s, c.lggr, capability, capabilities.CapabilityTypeTrigger)
	})
	if err != nil {
		return nil, err
	}

	return &registrypb.GetTriggerReply{
		CapabilityID: loc.LegacyHandle,
		Target:       loc.Target,
	}, nil
}

func (c *capabilitiesRegistryServer) GetExecutable(ctx context.Context, request *registrypb.GetExecutableRequest) (*registrypb.GetExecutableReply, error) {
	capability, err := c.impl.GetExecutable(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	info, err := capability.Info(ctx)
	if err != nil {
		return nil, err
	}

	switch info.CapabilityType {
	case capabilities.CapabilityTypeAction, capabilities.CapabilityTypeConsensus, capabilities.CapabilityTypeTarget, capabilities.CapabilityTypeCombined:
	default:
		return nil, fmt.Errorf("capability with id: %s does not satisfy the capability interface", request.Id)
	}

	loc, _, err := c.transport.Publish("GetExecutable", func(s *grpc.Server) {
		RegisterCapabilityServer(s, c.lggr, capability, info.CapabilityType)
	})
	if err != nil {
		return nil, err
	}

	return &registrypb.GetExecutableReply{
		CapabilityID: loc.LegacyHandle,
		Target:       loc.Target,
	}, nil
}

func (c *capabilitiesRegistryServer) List(ctx context.Context, _ *emptypb.Empty) (*registrypb.ListReply, error) {
	capabilities, err := c.impl.List(ctx)
	if err != nil {
		return nil, err
	}

	var (
		locs      []Locator
		resources []io.Closer
	)
	for _, cap := range capabilities {
		info, err := cap.Info(ctx)
		if err != nil {
			c.closeAll(resources...)
			return nil, err
		}

		loc, res, err := c.transport.Publish("List", func(s *grpc.Server) {
			RegisterCapabilityServer(s, c.lggr, cap, info.CapabilityType)
		})
		if err != nil {
			c.closeAll(resources...)
			return nil, err
		}
		resources = append(resources, res)
		locs = append(locs, loc)
	}

	return listReplyFor(locs), nil
}

func (c *capabilitiesRegistryServer) Add(ctx context.Context, request *registrypb.AddRequest) (*emptypb.Empty, error) {
	loc := Locator{Target: request.Target, LegacyHandle: request.CapabilityID}
	conn, err := c.transport.Dial(ctx, "Add", func(context.Context) (Locator, error) {
		return loc, nil
	})
	if err != nil {
		return nil, err
	}

	var client capabilities.BaseCapability

	switch request.Type {
	case registrypb.ExecuteAPIType_EXECUTE_API_TYPE_TRIGGER:
		client = NewTriggerCapabilityClient(c.lggr, conn)
	case registrypb.ExecuteAPIType_EXECUTE_API_TYPE_EXECUTE:
		client = NewExecutableCapabilityClient(c.lggr, conn)
	case registrypb.ExecuteAPIType_EXECUTE_API_TYPE_COMBINED:
		client = NewCombinedCapabilityClient(c.lggr, conn)
	default:
		return nil, fmt.Errorf("unknown execute type %d", request.Type)
	}

	if err := c.impl.Add(ctx, client); err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (c *capabilitiesRegistryServer) Remove(ctx context.Context, request *registrypb.RemoveRequest) (*emptypb.Empty, error) {
	err := c.impl.Remove(ctx, request.Id)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

// RegisterCapabilitiesRegistryServer serves i as the CapabilitiesRegistry service on s.
// Capabilities it hands out are published over t, and local ones passed to Add are
// dialled over it.
func RegisterCapabilitiesRegistryServer(s *grpc.Server, lggr logger.Logger, i registry.CapabilitiesRegistry, t Transport) {
	registrypb.RegisterCapabilitiesRegistryServer(s, NewCapabilitiesRegistryServer(lggr, i, t))
}

func NewCapabilitiesRegistryServer(lggr logger.Logger, i registry.CapabilitiesRegistry, t Transport) registrypb.CapabilitiesRegistryServer {
	return &capabilitiesRegistryServer{
		lggr:      logger.Named(lggr, "CapabilitiesRegistryServer"),
		transport: t,
		impl:      i,
	}
}

func (c *capabilitiesRegistryServer) closeAll(closers ...io.Closer) {
	for _, cl := range closers {
		if err := cl.Close(); err != nil {
			c.lggr.Errorw("Error closing served capability", "err", err)
		}
	}
}

// listReplyFor names each published capability both ways. Targets replace the
// deprecated handles wholesale rather than supplementing them, so a reply carries
// them only when every capability has one; a partial list would be read as complete.
func listReplyFor(locs []Locator) *registrypb.ListReply {
	reply := &registrypb.ListReply{}
	targets := make([]string, 0, len(locs))
	complete := true
	for _, loc := range locs {
		reply.CapabilityID = append(reply.CapabilityID, loc.LegacyHandle)
		if loc.Target == "" {
			complete = false
		}
		targets = append(targets, loc.Target)
	}
	if complete && len(targets) > 0 {
		reply.Targets = targets
	}
	return reply
}

// locatorsFromListReply reads the locators a List reply names, preferring targets.
func locatorsFromListReply(res *registrypb.ListReply) []Locator {
	if len(res.Targets) > 0 {
		locs := make([]Locator, len(res.Targets))
		for i, t := range res.Targets {
			locs[i] = Locator{Target: t}
			if i < len(res.CapabilityID) {
				locs[i].LegacyHandle = res.CapabilityID[i]
			}
		}
		return locs
	}
	locs := make([]Locator, len(res.CapabilityID))
	for i, id := range res.CapabilityID {
		locs[i] = Locator{LegacyHandle: id}
	}
	return locs
}
