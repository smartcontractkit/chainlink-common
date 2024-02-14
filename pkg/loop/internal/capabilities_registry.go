package internal

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.CapabilitiesRegistry = (*capabilitiesRegistryClient)(nil)

type capabilitiesRegistryClient struct {
	*BrokerExt
	grpc pb.CapabilitiesRegistryClient
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
		return nil, ErrConnDial{Name: "Capability", ID: res.CapabilityID, Err: err}
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
		return nil, ErrConnDial{Name: "GetTrigger", ID: res.CapabilityID, Err: err}
	}
	client := NewTriggerCapabilityClient(cr.BrokerExt, conn)
	return client, nil
}

func (cr *capabilitiesRegistryClient) GetAction(ctx context.Context, ID string) (capabilities.ActionCapability, error) {
	req := &pb.GetActionRequest{
		Id: ID,
	}

	res, err := cr.grpc.GetAction(ctx, req)
	if err != nil {
		return nil, err
	}
	conn, err := cr.Dial(res.CapabilityID)
	if err != nil {
		return nil, ErrConnDial{Name: "GetAction", ID: res.CapabilityID, Err: err}
	}
	client := NewActionCapabilityClient(cr.BrokerExt, conn)
	return client, nil

}

func (cr *capabilitiesRegistryClient) GetConsensus(ctx context.Context, ID string) (capabilities.ConsensusCapability, error) {
	req := &pb.GetConsensusRequest{
		Id: ID,
	}

	res, err := cr.grpc.GetConsensus(ctx, req)
	if err != nil {
		return nil, err
	}

	conn, err := cr.Dial(res.CapabilityID)
	if err != nil {
		return nil, ErrConnDial{Name: "GetAction", ID: res.CapabilityID, Err: err}
	}
	client := NewConsensusCapabilityClient(cr.BrokerExt, conn)
	return client, nil
}

func (cr *capabilitiesRegistryClient) GetTarget(ctx context.Context, ID string) (capabilities.TargetCapability, error) {
	req := &pb.GetTargetRequest{
		Id: ID,
	}

	res, err := cr.grpc.GetTarget(ctx, req)
	if err != nil {
		return nil, err
	}

	conn, err := cr.Dial(res.CapabilityID)
	if err != nil {
		return nil, ErrConnDial{Name: "GetAction", ID: res.CapabilityID, Err: err}
	}
	client := NewTargetCapabilityClient(cr.BrokerExt, conn)
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
			return nil, ErrConnDial{Name: "List", ID: id, Err: err}
		}
		client := newBaseCapabilityClient(cr.BrokerExt, conn)
		clients = append(clients, client)
	}

	return clients, nil
}

func (cr *capabilitiesRegistryClient) Add(ctx context.Context, c capabilities.BaseCapability) error {
	var cRes Resource
	id, cRes, err := cr.ServeNew("someCapability", func(s *grpc.Server) {
		pb.RegisterBaseCapabilityServer(s, &baseCapabilityServer{impl: c})
	})

	if err != nil {
		return err
	}

	info, err := c.Info(ctx)
	if err != nil {
		cRes.Close()
		return err
	}

	req := &pb.AddRequest{
		CapabilityID: id,
		Type:         pb.ExecuteAPIType(info.CapabilityType), // TODO: probably not right, figure out what this is
	}

	_, err = cr.grpc.Add(ctx, req)
	if err != nil {
		cRes.Close()
		return err
	}
	return nil
}

func NewCapabilitiesRegistryClient(cc grpc.ClientConnInterface) *capabilitiesRegistryClient {
	return &capabilitiesRegistryClient{grpc: pb.NewCapabilitiesRegistryClient(cc)}
}

var _ pb.CapabilitiesRegistryServer = (*capabilitiesRegistryServer)(nil)

type capabilitiesRegistryServer struct {
	pb.UnimplementedCapabilitiesRegistryServer
	*BrokerExt
	impl types.CapabilitiesRegistry
}

func (c *capabilitiesRegistryServer) Get(ctx context.Context, request *pb.GetRequest) (*pb.GetReply, error) {
	capability, err := c.impl.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	var cRes Resource
	id, cRes, err := c.ServeNew("someCapability", func(s *grpc.Server) {
		pb.RegisterBaseCapabilityServer(s, &baseCapabilityServer{impl: capability})
	})

	if err != nil {
		return nil, err
	}

	info, err := capability.Info(ctx)
	if err != nil {
		cRes.Close()
		return nil, err
	}

	return &pb.GetReply{
		CapabilityID: id,
		Type:         pb.ExecuteAPIType(info.CapabilityType), // TODO: probably not right, figure out what this is
	}, nil
}

func (c *capabilitiesRegistryServer) GetTrigger(ctx context.Context, request *pb.GetTriggerRequest) (*pb.GetTriggerReply, error) {
	capability, err := c.impl.GetTrigger(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	id, _, err := c.ServeNew("someTrigger", func(s *grpc.Server) {
		pb.RegisterTriggerExecutableServer(s, &triggerExecutableServer{impl: capability})
	})

	if err != nil {
		return nil, err
	}

	return &pb.GetTriggerReply{
		CapabilityID: id,
	}, nil
}

func (c *capabilitiesRegistryServer) GetAction(ctx context.Context, request *pb.GetActionRequest) (*pb.GetActionReply, error) {
	capability, err := c.impl.GetAction(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	id, _, err := c.ServeNew("someAction", func(s *grpc.Server) {
		pb.RegisterCallbackExecutableServer(s, &callbackExecutableServer{impl: capability})
	})

	if err != nil {
		return nil, err
	}

	return &pb.GetActionReply{
		CapabilityID: id,
	}, nil

}

func (c *capabilitiesRegistryServer) GetConsensus(ctx context.Context, request *pb.GetConsensusRequest) (*pb.GetConsensusReply, error) {
	capability, err := c.impl.GetConsensus(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	id, _, err := c.ServeNew("someConsensus", func(s *grpc.Server) {
		pb.RegisterCallbackExecutableServer(s, &callbackExecutableServer{impl: capability})
	})

	if err != nil {
		return nil, err
	}

	return &pb.GetConsensusReply{
		CapabilityID: id,
	}, nil
}

func (c *capabilitiesRegistryServer) GetTarget(ctx context.Context, request *pb.GetTargetRequest) (*pb.GetTargetReply, error) {
	capability, err := c.impl.GetTarget(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	id, _, err := c.ServeNew("someTarget", func(s *grpc.Server) {
		pb.RegisterCallbackExecutableServer(s, &callbackExecutableServer{impl: capability})
	})

	if err != nil {
		return nil, err
	}

	return &pb.GetTargetReply{
		CapabilityID: id,
	}, nil
}

func (c *capabilitiesRegistryServer) List(ctx context.Context, _ *emptypb.Empty) (*pb.ListReply, error) {
	capabilities, err := c.impl.List(ctx)
	if err != nil {
		return nil, err
	}

	reply := &pb.ListReply{}

	var resources []Resource
	for _, cap := range capabilities {
		id, res, err := c.ServeNew("List", func(s *grpc.Server) {
			pb.RegisterBaseCapabilityServer(s, &baseCapabilityServer{impl: cap})
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
		return &emptypb.Empty{}, ErrConnDial{Name: "Add", ID: request.CapabilityID, Err: err}
	}
	client := newBaseCapabilityClient(c.BrokerExt, conn)

	err = c.impl.Add(ctx, client)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func NewCapabilitiesRegistryServer(b *BrokerExt, i types.CapabilitiesRegistry) *capabilitiesRegistryServer {
	return &capabilitiesRegistryServer{
		BrokerExt: b,
		impl:      i,
	}
}
