package remote

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	registrypb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ registrypb.CapabilitiesRegistryServer = (*capabilitiesRegistryServer)(nil)

type capabilitiesRegistryServer struct {
	registrypb.UnimplementedCapabilitiesRegistryServer
	lggr      logger.Logger
	transport Transport
	impl      registry.CapabilitiesRegistry

	mu        sync.Mutex
	published map[publishKey]publishedCapability
}

// publishKey includes the registered type because GetTrigger serves a combined
// capability as a trigger only, which is a different server from the one Get serves.
type publishKey struct {
	id  string
	typ capabilities.CapabilityType
}

type publishedCapability struct {
	loc Locator
	res io.Closer
}

// publish serves capability at most once per key, so repeated lookups share a server.
// Callers must hold c.mu from the impl lookup through publish, otherwise a concurrent
// Remove can land in between and the removed capability stays published.
func (c *capabilitiesRegistryServer) publish(id string, capability capabilities.BaseCapability, typ capabilities.CapabilityType) (Locator, error) {
	key := publishKey{id: id, typ: typ}
	if p, ok := c.published[key]; ok {
		return p.loc, nil
	}

	loc, res, err := c.transport.Publish(id, func(s *grpc.Server) {
		RegisterCapabilityServer(s, c.lggr, capability, typ)
	})
	if err != nil {
		return Locator{}, err
	}
	c.published[key] = publishedCapability{loc: loc, res: res}
	return loc, nil
}

func (c *capabilitiesRegistryServer) Get(ctx context.Context, request *registrypb.GetRequest) (*registrypb.GetReply, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	capability, err := c.impl.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	info, err := capability.Info(ctx)
	if err != nil {
		return nil, err
	}

	loc, err := c.publish(info.ID, capability, info.CapabilityType)
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
	c.mu.Lock()
	defer c.mu.Unlock()

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

	loc, err := c.publish(info.ID, capability, capabilities.CapabilityTypeTrigger)
	if err != nil {
		return nil, err
	}

	return &registrypb.GetTriggerReply{
		CapabilityID: loc.LegacyHandle,
		Target:       loc.Target,
	}, nil
}

func (c *capabilitiesRegistryServer) GetExecutable(ctx context.Context, request *registrypb.GetExecutableRequest) (*registrypb.GetExecutableReply, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

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

	loc, err := c.publish(info.ID, capability, info.CapabilityType)
	if err != nil {
		return nil, err
	}

	return &registrypb.GetExecutableReply{
		CapabilityID: loc.LegacyHandle,
		Target:       loc.Target,
	}, nil
}

func (c *capabilitiesRegistryServer) List(ctx context.Context, _ *emptypb.Empty) (*registrypb.ListReply, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	capabilities, err := c.impl.List(ctx)
	if err != nil {
		return nil, err
	}

	var locs []Locator
	for _, cap := range capabilities {
		info, err := cap.Info(ctx)
		if err != nil {
			return nil, err
		}

		loc, err := c.publish(info.ID, cap, info.CapabilityType)
		if err != nil {
			return nil, err
		}
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
	id := request.Id
	c.mu.Lock()
	if err := c.impl.Remove(ctx, id); err != nil {
		c.mu.Unlock()
		return &emptypb.Empty{}, err
	}
	var closers []io.Closer
	for key, p := range c.published {
		if key.id == id {
			closers = append(closers, p.res)
			delete(c.published, key)
		}
	}
	c.mu.Unlock()

	for _, cl := range closers {
		if err := cl.Close(); err != nil {
			c.lggr.Errorw("Error closing served capability", "id", id, "err", err)
		}
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
		published: map[publishKey]publishedCapability{},
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
