package capability

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	registryremote "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote"
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

func toPbDON(don capabilities.DON) *pb.DON {
	membersBytes := make([][]byte, len(don.Members))
	for j, m := range don.Members {
		membersBytes[j] = m[:]
	}

	return &pb.DON{
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

	return registryremote.NodeFromProto(res, res.WorkflowDON, res.CapabilityDONs)
}

func (cr *capabilitiesRegistryClient) NodeByPeerID(ctx context.Context, peerID p2ptypes.PeerID) (capabilities.Node, error) {
	res, err := cr.grpc.NodeByPeerID(ctx, &pb.NodeRequest{PeerID: peerID[:]})
	if err != nil {
		return capabilities.Node{}, err
	}

	return registryremote.NodeFromProto(res, res.WorkflowDON, res.CapabilityDONs)
}

func (cr *capabilitiesRegistryClient) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	res, err := cr.grpc.DONsForCapability(ctx, &pb.DONForCapabilityRequest{CapabilityID: capabilityID})
	if err != nil {
		return nil, err
	}

	donsWithNodes := []capabilities.DONWithNodes{}
	for _, d := range res.Dons {
		var nodes []capabilities.Node
		for _, n := range d.Nodes {
			node, err := registryremote.NodeFromProto(n, n.WorkflowDON, n.CapabilityDONs)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
		donsWithNodes = append(donsWithNodes, capabilities.DONWithNodes{
			DON:   registryremote.DONFromProto(d.Don),
			Nodes: nodes,
		})
	}
	return donsWithNodes, nil
}

func (cr *capabilitiesRegistryClient) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	res, err := cr.grpc.DONByID(ctx, &pb.DONByIDRequest{DonID: donID})
	if err != nil {
		return capabilities.DON{}, err
	}
	return registryremote.DONFromProto(res.Don), nil
}

func (cr *capabilitiesRegistryClient) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	res, err := cr.grpc.ConfigForCapability(ctx, &pb.ConfigForCapabilityRequest{
		CapabilityID: capabilityID,
		DonID:        donID,
	})
	if err != nil {
		return capabilities.CapabilityConfiguration{}, err
	}

	return capabilitiespb.CapabilityConfigFromProto(res.CapabilityConfig)
}

func (cr *capabilitiesRegistryClient) Get(ctx context.Context, ID string) (capabilities.BaseCapability, error) {
	req := &pb.GetRequest{
		Id: ID,
	}

	conn := cr.NewClientConn("Capability", func(ctx context.Context) (id uint32, deps net.Resources, err error) {
		res, err := cr.grpc.Get(ctx, req)
		if err != nil {
			return 0, nil, err
		}
		return res.CapabilityID, nil, nil
	})
	client := registryremote.Wrap(cr.Logger, conn, capabilities.CapabilityTypeUnknown)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_, err := client.Info(ctx) // ensure exists by triggering lazy connection with reduced timeout
	return client, err
}

func (cr *capabilitiesRegistryClient) GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error) {
	req := &pb.GetTriggerRequest{
		Id: ID,
	}

	conn := cr.NewClientConn("Trigger", func(ctx context.Context) (id uint32, deps net.Resources, err error) {
		res, err := cr.grpc.GetTrigger(ctx, req)
		if err != nil {
			return 0, nil, err
		}
		return res.CapabilityID, nil, nil
	})
	client := NewTriggerCapabilityClient(cr.BrokerExt, conn)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_, err := client.Info(ctx) // ensure exists by triggering lazy connection with reduced timeout
	return client, err
}

func (cr *capabilitiesRegistryClient) GetExecutable(ctx context.Context, ID string) (capabilities.ExecutableCapability, error) {
	req := &pb.GetExecutableRequest{
		Id: ID,
	}

	conn := cr.NewClientConn("Executable", func(ctx context.Context) (id uint32, deps net.Resources, err error) {
		res, err := cr.grpc.GetExecutable(ctx, req)
		if err != nil {
			return 0, nil, err
		}
		return res.CapabilityID, nil, nil
	})
	client := NewExecutableCapabilityClient(cr.BrokerExt, conn)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_, err := client.Info(ctx) // ensure exists by triggering lazy connection with reduced timeout
	return client, err
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
		client := registryremote.Wrap(cr.Logger, conn, capabilities.CapabilityTypeUnknown)
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
		if err := registryremote.RegisterCapability(cr.Logger, s, c, info.CapabilityType); err != nil {
			cr.Logger.Errorw("failed to register capability", "capabilityID", info.ID, "err", err)
		}
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
		if err := registryremote.RegisterCapability(c.Logger, s, capability, info.CapabilityType); err != nil {
			c.Logger.Errorw("failed to register capability", "capabilityID", request.Id, "err", err)
		}
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

	ccp, err := capabilitiespb.CapabilityConfigToProto(cc)
	if err != nil {
		return nil, err
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

func (c *capabilitiesRegistryServer) DONsForCapability(ctx context.Context, req *pb.DONForCapabilityRequest) (*pb.DONForCapabilityReply, error) {
	dons, err := c.impl.DONsForCapability(ctx, req.CapabilityID)
	if err != nil {
		return nil, err
	}

	donWithNodes := []*pb.DONWithNodes{}
	for _, d := range dons {
		pbDon := toPbDON(d.DON)
		nodes := []*pb.NodeReply{}
		for _, n := range d.Nodes {
			nodes = append(nodes, c.nodeReplyFromNode(n))
		}
		donWithNodes = append(donWithNodes, &pb.DONWithNodes{
			Don:   pbDon,
			Nodes: nodes,
		})
	}

	return &pb.DONForCapabilityReply{
		Dons: donWithNodes,
	}, nil
}

func (c *capabilitiesRegistryServer) DONByID(ctx context.Context, req *pb.DONByIDRequest) (*pb.DONByIDReply, error) {
	don, err := c.impl.DONByID(ctx, req.DonID)
	if err != nil {
		return nil, err
	}
	return &pb.DONByIDReply{Don: toPbDON(don)}, nil
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
		if err := registryremote.RegisterCapability(c.Logger, s, capability, capabilities.CapabilityTypeTrigger); err != nil {
			c.Logger.Errorw("failed to register capability", "capabilityID", request.Id, "err", err)
		}
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
		if err := registryremote.RegisterCapability(c.Logger, s, capability, info.CapabilityType); err != nil {
			c.Logger.Errorw("failed to register capability", "capabilityID", request.Id, "err", err)
		}
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
			if err := registryremote.RegisterCapability(c.Logger, s, cap, info.CapabilityType); err != nil {
				c.Logger.Errorw("failed to register capability", "err", err)
			}
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

var _ registry.StateGetter = (*registryremote.TriggerCapabilityClient)(nil)
var _ registry.StateGetter = (*registryremote.ExecutableCapabilityClient)(nil)
var _ registry.StateGetter = (*registryremote.CombinedCapabilityClient)(nil)

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

// getExecuteAPIType maps a capability type to the broker API enum carried by
// GetReply and AddRequest.
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
