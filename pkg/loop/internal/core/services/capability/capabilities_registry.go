package capability

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/values"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"

	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

var _ core.CapabilitiesRegistry = (*capabilitiesRegistryClient)(nil)

type capabilitiesRegistryClient struct {
	*net.BrokerExt
	grpc pb.CapabilitiesRegistryClient
}

func toDON(don *pb.DON) capabilities.DON {
	var members []p2ptypes.PeerID
	for _, m := range don.Members {
		members = append(members, p2ptypes.PeerID(m))
	}

	return capabilities.DON{
		ID:            don.Id,
		Members:       members,
		F:             uint8(don.F),
		ConfigVersion: don.ConfigVersion,
	}
}

func toPbDON(don capabilities.DON) *pb.DON {
	membersBytes := make([][]byte, len(don.Members))
	for j, m := range don.Members {
		m := m
		membersBytes[j] = m[:]
	}

	return &pb.DON{
		Id:            don.ID,
		Members:       membersBytes,
		F:             uint32(don.F),
		ConfigVersion: don.ConfigVersion,
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
	res, err := cr.grpc.NodeByPeerID(ctx, &pb.NodeRequest{PeerID: peerID[:]})
	if err != nil {
		return capabilities.Node{}, err
	}

	return cr.nodeFromNodeReply(res), nil
}

func (cr *capabilitiesRegistryClient) nodeFromNodeReply(nodeReply *pb.NodeReply) capabilities.Node {
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
	res, err := cr.grpc.ConfigForCapability(ctx, &pb.ConfigForCapabilityRequest{
		CapabilityID: capabilityID,
		DonID:        donID,
	})
	if err != nil {
		return capabilities.CapabilityConfiguration{}, err
	}

	mc, err := values.FromMapValueProto(res.CapabilityConfig.DefaultConfig)
	if err != nil {
		return capabilities.CapabilityConfiguration{}, fmt.Errorf("could not convert map valueproto to map: %w", err)
	}

	var remoteTriggerConfig *capabilities.RemoteTriggerConfig
	var remoteTargetConfig *capabilities.RemoteTargetConfig
	var remoteExecutableConfig *capabilities.RemoteExecutableConfig

	switch res.CapabilityConfig.RemoteConfig.(type) {
	case *capabilitiespb.CapabilityConfig_RemoteTriggerConfig:
		prtc := res.CapabilityConfig.GetRemoteTriggerConfig()
		remoteTriggerConfig = &capabilities.RemoteTriggerConfig{}
		remoteTriggerConfig.RegistrationRefresh = prtc.RegistrationRefresh.AsDuration()
		remoteTriggerConfig.RegistrationExpiry = prtc.RegistrationExpiry.AsDuration()
		remoteTriggerConfig.MinResponsesToAggregate = prtc.MinResponsesToAggregate
		remoteTriggerConfig.MessageExpiry = prtc.MessageExpiry.AsDuration()
		remoteTriggerConfig.MaxBatchSize = prtc.MaxBatchSize
		remoteTriggerConfig.BatchCollectionPeriod = prtc.BatchCollectionPeriod.AsDuration()

	case *capabilitiespb.CapabilityConfig_RemoteTargetConfig:
		prtc := res.CapabilityConfig.GetRemoteTargetConfig()
		remoteTargetConfig = &capabilities.RemoteTargetConfig{}
		remoteTargetConfig.RequestHashExcludedAttributes = prtc.RequestHashExcludedAttributes
	case *capabilitiespb.CapabilityConfig_RemoteExecutableConfig:
		prtc := res.CapabilityConfig.GetRemoteExecutableConfig()
		remoteExecutableConfig = &capabilities.RemoteExecutableConfig{}
		remoteExecutableConfig.RequestHashExcludedAttributes = prtc.RequestHashExcludedAttributes
		remoteExecutableConfig.RegistrationRefresh = prtc.RegistrationRefresh.AsDuration()
		remoteExecutableConfig.RegistrationExpiry = prtc.RegistrationExpiry.AsDuration()
	}

	return capabilities.CapabilityConfiguration{
		DefaultConfig:          mc,
		RemoteTriggerConfig:    remoteTriggerConfig,
		RemoteTargetConfig:     remoteTargetConfig,
		RemoteExecutableConfig: remoteExecutableConfig,
	}, nil
}

func (cr *capabilitiesRegistryClient) Get(ctx context.Context, ID string) (capabilities.BaseCapability, error) {
	req := &pb.GetRequest{
		Id: ID,
	}

	res, err := cr.grpc.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	conn, err := cr.Dial(res.CapabilityID)
	if err != nil {
		return nil, net.ErrConnDial{Name: "Capability", ID: res.CapabilityID, Err: err}
	}
	client := newBaseCapabilityClient(cr.BrokerExt, conn)
	return client, nil
}

func (cr *capabilitiesRegistryClient) GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error) {
	req := &pb.GetTriggerRequest{
		Id: ID,
	}

	res, err := cr.grpc.GetTrigger(ctx, req)
	if err != nil {
		return nil, err
	}

	conn, err := cr.Dial(res.CapabilityID)
	if err != nil {
		return nil, net.ErrConnDial{Name: "GetTrigger", ID: res.CapabilityID, Err: err}
	}
	client := NewTriggerCapabilityClient(cr.BrokerExt, conn)
	return client, nil
}

func (cr *capabilitiesRegistryClient) GetExecutable(ctx context.Context, ID string) (capabilities.ExecutableCapability, error) {
	req := &pb.GetExecutableRequest{
		Id: ID,
	}

	res, err := cr.grpc.GetExecutable(ctx, req)
	if err != nil {
		return nil, err
	}
	conn, err := cr.Dial(res.CapabilityID)
	if err != nil {
		return nil, net.ErrConnDial{Name: "GetExecutable", ID: res.CapabilityID, Err: err}
	}
	client := NewExecutableCapabilityClient(cr.BrokerExt, conn)
	return client, nil
}

func (cr *capabilitiesRegistryClient) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	res, err := cr.grpc.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var clients []capabilities.BaseCapability
	for _, id := range res.CapabilityID {
		conn, err := cr.Dial(id)
		if err != nil {
			return nil, net.ErrConnDial{Name: "List", ID: id, Err: err}
		}
		client := newBaseCapabilityClient(cr.BrokerExt, conn)
		clients = append(clients, client)
	}

	return clients, nil
}

func (cr *capabilitiesRegistryClient) Add(ctx context.Context, c capabilities.BaseCapability) error {
	info, err := c.Info(ctx)
	if err != nil {
		return err
	}

	var cRes net.Resource
	id, cRes, err := cr.ServeNew(info.ID, func(s *grpc.Server) {
		pbRegisterCapability(s, cr.BrokerExt, c, info.CapabilityType)
	})
	if err != nil {
		return err
	}

	_, err = cr.grpc.Add(ctx, &pb.AddRequest{
		CapabilityID: id,
		Type:         getExecuteAPIType(info.CapabilityType),
	})
	if err != nil {
		cRes.Close()
		return err
	}
	return nil
}

func (cr *capabilitiesRegistryClient) Remove(ctx context.Context, ID string) error {
	req := &pb.RemoveRequest{
		Id: ID,
	}

	_, err := cr.grpc.Remove(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func NewCapabilitiesRegistryClient(cc grpc.ClientConnInterface, b *net.BrokerExt) *capabilitiesRegistryClient {
	return &capabilitiesRegistryClient{grpc: pb.NewCapabilitiesRegistryClient(cc), BrokerExt: b.WithName("CapabilitiesRegistryClient")}
}

var _ pb.CapabilitiesRegistryServer = (*capabilitiesRegistryServer)(nil)

type capabilitiesRegistryServer struct {
	pb.UnimplementedCapabilitiesRegistryServer
	*net.BrokerExt
	impl core.CapabilitiesRegistry
}

func (c *capabilitiesRegistryServer) Get(ctx context.Context, request *pb.GetRequest) (*pb.GetReply, error) {
	capability, err := c.impl.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	info, err := capability.Info(ctx)
	if err != nil {
		return nil, err
	}

	id, _, err := c.ServeNew("Get", func(s *grpc.Server) {
		pbRegisterCapability(s, c.BrokerExt, capability, info.CapabilityType)
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetReply{
		CapabilityID: id,
		Type:         getExecuteAPIType(info.CapabilityType),
	}, nil
}

func (c *capabilitiesRegistryServer) ConfigForCapability(ctx context.Context, req *pb.ConfigForCapabilityRequest) (*pb.ConfigForCapabilityReply, error) {
	cc, err := c.impl.ConfigForCapability(ctx, req.CapabilityID, req.DonID)
	if err != nil {
		return nil, err
	}

	ecm := values.Proto(cc.DefaultConfig).GetMapValue()

	ccp := &capabilitiespb.CapabilityConfig{
		DefaultConfig: ecm,
	}

	if cc.RemoteTriggerConfig != nil {
		ccp.RemoteConfig = &capabilitiespb.CapabilityConfig_RemoteTriggerConfig{
			RemoteTriggerConfig: &capabilitiespb.RemoteTriggerConfig{
				RegistrationRefresh:     durationpb.New(cc.RemoteTriggerConfig.RegistrationRefresh),
				RegistrationExpiry:      durationpb.New(cc.RemoteTriggerConfig.RegistrationExpiry),
				MinResponsesToAggregate: cc.RemoteTriggerConfig.MinResponsesToAggregate,
				MessageExpiry:           durationpb.New(cc.RemoteTriggerConfig.MessageExpiry),
				MaxBatchSize:            cc.RemoteTriggerConfig.MaxBatchSize,
				BatchCollectionPeriod:   durationpb.New(cc.RemoteTriggerConfig.BatchCollectionPeriod),
			},
		}
	}

	if cc.RemoteTargetConfig != nil {
		ccp.RemoteConfig = &capabilitiespb.CapabilityConfig_RemoteTargetConfig{
			RemoteTargetConfig: &capabilitiespb.RemoteTargetConfig{
				RequestHashExcludedAttributes: cc.RemoteTargetConfig.RequestHashExcludedAttributes,
			},
		}
	}

	if cc.RemoteExecutableConfig != nil {
		ccp.RemoteConfig = &capabilitiespb.CapabilityConfig_RemoteExecutableConfig{
			RemoteExecutableConfig: &capabilitiespb.RemoteExecutableConfig{
				RequestHashExcludedAttributes: cc.RemoteExecutableConfig.RequestHashExcludedAttributes,
				RegistrationRefresh:           durationpb.New(cc.RemoteExecutableConfig.RegistrationRefresh),
				RegistrationExpiry:            durationpb.New(cc.RemoteExecutableConfig.RegistrationExpiry),
			},
		}
	}

	return &pb.ConfigForCapabilityReply{
		CapabilityConfig: ccp,
	}, nil
}

func (c *capabilitiesRegistryServer) LocalNode(ctx context.Context, _ *emptypb.Empty) (*pb.NodeReply, error) {
	node, err := c.impl.LocalNode(ctx)
	if err != nil {
		return nil, err
	}

	return c.nodeReplyFromNode(node), nil
}

func (c *capabilitiesRegistryServer) NodeByPeerID(ctx context.Context, nodeRequest *pb.NodeRequest) (*pb.NodeReply, error) {
	node, err := c.impl.NodeByPeerID(ctx, p2ptypes.PeerID(nodeRequest.GetPeerID()))
	if err != nil {
		return nil, err
	}

	return c.nodeReplyFromNode(node), nil
}

func (c *capabilitiesRegistryServer) nodeReplyFromNode(node capabilities.Node) *pb.NodeReply {
	workflowDONpb := toPbDON(node.WorkflowDON)

	capabilityDONsPb := make([]*pb.DON, len(node.CapabilityDONs))
	for i, don := range node.CapabilityDONs {
		capabilityDONsPb[i] = toPbDON(don)
	}

	var pid []byte
	if node.PeerID != nil {
		pid = node.PeerID[:]
	}
	reply := &pb.NodeReply{
		PeerID:              pid,
		NodeOperatorID:      node.NodeOperatorID,
		Signer:              node.Signer[:],
		EncryptionPublicKey: node.EncryptionPublicKey[:],
		WorkflowDON:         workflowDONpb,
		CapabilityDONs:      capabilityDONsPb,
	}

	return reply
}

func (c *capabilitiesRegistryServer) GetTrigger(ctx context.Context, request *pb.GetTriggerRequest) (*pb.GetTriggerReply, error) {
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

	id, _, err := c.ServeNew("GetTrigger", func(s *grpc.Server) {
		pbRegisterCapability(s, c.BrokerExt, capability, capabilities.CapabilityTypeTrigger)
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetTriggerReply{
		CapabilityID: id,
	}, nil
}

func (c *capabilitiesRegistryServer) GetExecutable(ctx context.Context, request *pb.GetExecutableRequest) (*pb.GetExecutableReply, error) {
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

	id, _, err := c.ServeNew("GetExecutable", func(s *grpc.Server) {
		pbRegisterCapability(s, c.BrokerExt, capability, info.CapabilityType)
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetExecutableReply{
		CapabilityID: id,
	}, nil
}

func (c *capabilitiesRegistryServer) List(ctx context.Context, _ *emptypb.Empty) (*pb.ListReply, error) {
	capabilities, err := c.impl.List(ctx)
	if err != nil {
		return nil, err
	}

	reply := &pb.ListReply{}

	var resources []net.Resource
	for _, cap := range capabilities {
		info, err := cap.Info(ctx)
		if err != nil {
			c.CloseAll(resources...)
			return nil, err
		}

		id, res, err := c.ServeNew("List", func(s *grpc.Server) {
			pbRegisterCapability(s, c.BrokerExt, cap, info.CapabilityType)
		})
		if err != nil {
			c.CloseAll(resources...)
			return nil, err
		}
		resources = append(resources, res)
		reply.CapabilityID = append(reply.CapabilityID, id)
	}

	return reply, nil
}

func (c *capabilitiesRegistryServer) Add(ctx context.Context, request *pb.AddRequest) (*emptypb.Empty, error) {
	conn, err := c.Dial(request.CapabilityID)
	if err != nil {
		return &emptypb.Empty{}, net.ErrConnDial{Name: "Add", ID: request.CapabilityID, Err: err}
	}
	var client capabilities.BaseCapability

	switch request.Type {
	case pb.ExecuteAPIType_EXECUTE_API_TYPE_TRIGGER:
		client = NewTriggerCapabilityClient(c.BrokerExt, conn)
	case pb.ExecuteAPIType_EXECUTE_API_TYPE_EXECUTE:
		client = NewExecutableCapabilityClient(c.BrokerExt, conn)
	case pb.ExecuteAPIType_EXECUTE_API_TYPE_COMBINED:
		client = NewCombinedCapabilityClient(c.BrokerExt, conn)
	default:
		return nil, fmt.Errorf("unknown execute type %d", request.Type)
	}

	err = c.impl.Add(ctx, client)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (c *capabilitiesRegistryServer) Remove(ctx context.Context, request *pb.RemoveRequest) (*emptypb.Empty, error) {
	err := c.impl.Remove(ctx, request.Id)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func NewCapabilitiesRegistryServer(b *net.BrokerExt, i core.CapabilitiesRegistry) *capabilitiesRegistryServer {
	return &capabilitiesRegistryServer{
		BrokerExt: b.WithName("CapabilitiesRegistryServer"),
		impl:      i,
	}
}

// pbRegisterCapability registers the server with the correct capability based on capability type, this method assumes
// that the capability has already been validated with validateCapability.
func pbRegisterCapability(s *grpc.Server, b *net.BrokerExt, impl capabilities.BaseCapability, t capabilities.CapabilityType) {
	switch t {
	case capabilities.CapabilityTypeTrigger:
		i, _ := impl.(capabilities.TriggerCapability)
		capabilitiespb.RegisterTriggerExecutableServer(s, &triggerExecutableServer{
			BrokerExt: b,
			impl:      i,
		})
	case capabilities.CapabilityTypeCombined:
		t, _ := impl.(capabilities.TriggerCapability)
		capabilitiespb.RegisterTriggerExecutableServer(s, &triggerExecutableServer{
			BrokerExt: b,
			impl:      t,
		})
		e, _ := impl.(capabilities.ExecutableCapability)
		capabilitiespb.RegisterExecutableServer(s, &executableServer{
			BrokerExt:   b,
			impl:        e,
			cancelFuncs: map[string]func(){},
		})
	case capabilities.CapabilityTypeTarget, capabilities.CapabilityTypeAction, capabilities.CapabilityTypeConsensus:
		i, _ := impl.(capabilities.ExecutableCapability)
		capabilitiespb.RegisterExecutableServer(s, &executableServer{
			BrokerExt:   b,
			impl:        i,
			cancelFuncs: map[string]func(){},
		})
	case capabilities.CapabilityTypeUnknown:
		// Only register the base capability server
	}
	capabilitiespb.RegisterBaseCapabilityServer(s, newBaseCapabilityServer(impl))
}

func getExecuteAPIType(c capabilities.CapabilityType) pb.ExecuteAPIType {
	switch c {
	case capabilities.CapabilityTypeTrigger:
		return pb.ExecuteAPIType_EXECUTE_API_TYPE_TRIGGER
	case capabilities.CapabilityTypeAction, capabilities.CapabilityTypeConsensus, capabilities.CapabilityTypeTarget:
		return pb.ExecuteAPIType_EXECUTE_API_TYPE_EXECUTE
	case capabilities.CapabilityTypeCombined:
		return pb.ExecuteAPIType_EXECUTE_API_TYPE_COMBINED
	default:
		return pb.ExecuteAPIType_EXECUTE_API_TYPE_UNKNOWN
	}
}
