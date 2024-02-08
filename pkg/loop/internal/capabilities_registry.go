package internal

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// In the long-term, registration of capabilities should happen entirely
// inside the core node, via well-defined registration paths. The canonical
// example of this would be the standard capability registration path defined in
// this design document:
// https://docs.google.com/document/d/1i2m1BEytNUR-yBU5mLFVQLGv9FFgYWkVUllFGB062UQ/edit#heading=h.4e7ng0tekbkc

// In the medium term, we need to define an OCR3 capability, and the best way
// to do this is to pass in a capability registry when creating reporting plugin
// loop.

// For that reason, we add this implementation here.
type CapabilitiesRegistryClient struct {
	*brokerExt
	client pb.CapabilitiesRegistryClient
}

var _ types.CapabilitiesRegistry = (*CapabilitiesRegistryClient)(nil)

func NewCapabilitiesRegistryClient(broker Broker, brokerCfg BrokerConfig, conn *grpc.ClientConn) (*CapabilitiesRegistryClient, error) {
	return &CapabilitiesRegistryClient{
		client:    pb.NewCapabilitiesRegistryClient(conn),
		brokerExt: &brokerExt{broker, brokerCfg},
	}, nil
}

func (c *CapabilitiesRegistryClient) Get(ctx context.Context, id string) (capabilities.BaseCapability, error) {
	resp, err := c.client.Get(ctx, &pb.GetRequest{Id: id})
	if err != nil {
		return nil, err
	}

	conn, err := c.dial(resp.CapabilityID)
	if err != nil {
		return nil, err
	}

	switch resp.Type {
	case pb.ExecuteAPIType_EXECUTE_API_TYPE_TRIGGER:
		return newTriggerCapabilityClient(c.brokerExt, conn), nil
	case pb.ExecuteAPIType_EXECUTE_API_TYPE_CALLBACK:
		return newCallbackCapabilityClient(c.brokerExt, conn), nil
	case pb.ExecuteAPIType_EXECUTE_API_TYPE_UNKNOWN:
	}

	return nil, fmt.Errorf("unknown api type %s", resp.Type)
}

func (c *CapabilitiesRegistryClient) GetTrigger(ctx context.Context, id string) (capabilities.TriggerCapability, error) {
	resp, err := c.client.GetTrigger(ctx, &pb.GetTriggerRequest{Id: id})
	if err != nil {
		return nil, err
	}

	conn, err := c.dial(resp.CapabilityID)
	if err != nil {
		return nil, err
	}

	return newTriggerCapabilityClient(c.brokerExt, conn), nil
}

func (c *CapabilitiesRegistryClient) GetAction(ctx context.Context, id string) (capabilities.ActionCapability, error) {
	resp, err := c.client.GetAction(ctx, &pb.GetActionRequest{Id: id})
	if err != nil {
		return nil, err
	}

	conn, err := c.dial(resp.CapabilityID)
	if err != nil {
		return nil, err
	}

	return newCallbackCapabilityClient(c.brokerExt, conn), nil
}

func (c *CapabilitiesRegistryClient) GetConsensus(ctx context.Context, id string) (capabilities.ConsensusCapability, error) {
	resp, err := c.client.GetConsensus(ctx, &pb.GetConsensusRequest{Id: id})
	if err != nil {
		return nil, err
	}

	conn, err := c.dial(resp.CapabilityID)
	if err != nil {
		return nil, err
	}

	return newCallbackCapabilityClient(c.brokerExt, conn), nil
}

func (c *CapabilitiesRegistryClient) GetTarget(ctx context.Context, id string) (capabilities.TargetCapability, error) {
	resp, err := c.client.GetTarget(ctx, &pb.GetTargetRequest{Id: id})
	if err != nil {
		return nil, err
	}

	conn, err := c.dial(resp.CapabilityID)
	if err != nil {
		return nil, err
	}

	return newCallbackCapabilityClient(c.brokerExt, conn), nil
}

func (c *CapabilitiesRegistryClient) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	return nil, nil
}

func (c *CapabilitiesRegistryClient) Add(ctx context.Context, cap capabilities.BaseCapability) error {
	var (
		capabilityID   uint32
		executeAPIType pb.ExecuteAPIType
		openedResource resource
	)
	switch tc := cap.(type) {
	case capabilities.TriggerExecutable:
		cid, res, err := c.serveNew("TriggerCapability", func(s *grpc.Server) {
			pb.RegisterBaseCapabilityServer(s, newBaseCapabilityServer(cap))
			pb.RegisterTriggerCapabilityServer(s, newTriggerCapabilityServer(c.brokerExt, tc))
		})
		if err != nil {
			return err
		}

		capabilityID = cid
		executeAPIType = pb.ExecuteAPIType_EXECUTE_API_TYPE_TRIGGER
		openedResource = res
	case capabilities.CallbackExecutable:
		cid, res, err := c.serveNew("CallbackCapability", func(s *grpc.Server) {
			pb.RegisterBaseCapabilityServer(s, newBaseCapabilityServer(cap))
			pb.RegisterCallbackExecutableServer(s, newCallbackExecutableServer(c.brokerExt, tc))
		})
		if err != nil {
			return err
		}

		capabilityID = cid
		executeAPIType = pb.ExecuteAPIType_EXECUTE_API_TYPE_CALLBACK
		openedResource = res
	default:
		return fmt.Errorf("could not add capability: unknown capability type: %T", cap)
	}

	_, err := c.client.Add(ctx, &pb.AddRequest{CapabilityID: capabilityID, Type: executeAPIType})
	if err != nil {
		c.closeAll(openedResource)
	}
	return err
}

func newCapabilitiesRegistryServer(b *brokerExt, cr types.CapabilitiesRegistry) *CapabilitiesRegistryServer {
	return &CapabilitiesRegistryServer{brokerExt: b, impl: cr}
}

func RegisterCapabilitiesRegistryServer(server *grpc.Server, broker Broker, brokerCfg BrokerConfig, impl types.CapabilitiesRegistry) error {
	pb.RegisterCapabilitiesRegistryServer(server, newCapabilitiesRegistryServer(&brokerExt{broker, brokerCfg}, impl))
	return nil
}

type CapabilitiesRegistryServer struct {
	pb.UnimplementedCapabilitiesRegistryServer

	impl types.CapabilitiesRegistry
	*brokerExt
}

var _ pb.CapabilitiesRegistryServer = (*CapabilitiesRegistryServer)(nil)

func (c *CapabilitiesRegistryServer) Get(ctx context.Context, request *pb.GetRequest) (*pb.GetReply, error) {
	id := request.Id
	cp, err := c.impl.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	var capabilityID uint32
	var executeAPIType pb.ExecuteAPIType
	switch tc := cp.(type) {
	case capabilities.TriggerExecutable:
		cid, _, err := c.serveNew("TriggerCapability", func(s *grpc.Server) {
			pb.RegisterBaseCapabilityServer(s, newBaseCapabilityServer(cp))
			pb.RegisterTriggerCapabilityServer(s, newTriggerCapabilityServer(c.brokerExt, tc))
		})
		if err != nil {
			return nil, err
		}

		capabilityID = cid
		executeAPIType = pb.ExecuteAPIType_EXECUTE_API_TYPE_TRIGGER
	case capabilities.CallbackExecutable:
		cid, _, err := c.serveNew("CallbackCapability", func(s *grpc.Server) {
			pb.RegisterBaseCapabilityServer(s, newBaseCapabilityServer(cp))
			pb.RegisterCallbackExecutableServer(s, newCallbackExecutableServer(c.brokerExt, tc))
		})
		if err != nil {
			return nil, err
		}

		capabilityID = cid
		executeAPIType = pb.ExecuteAPIType_EXECUTE_API_TYPE_CALLBACK
	default:
		return nil, fmt.Errorf("unknown capability type: %T", cp)
	}

	return &pb.GetReply{
		CapabilityID: capabilityID,
		Type:         executeAPIType,
	}, nil
}

func (c *CapabilitiesRegistryServer) GetTrigger(ctx context.Context, request *pb.GetTriggerRequest) (*pb.GetTriggerReply, error) {
	id := request.Id
	cp, err := c.impl.GetTrigger(ctx, id)
	if err != nil {
		return nil, err
	}

	cid, _, err := c.serveNew("TriggerCapability", func(s *grpc.Server) {
		pb.RegisterBaseCapabilityServer(s, newBaseCapabilityServer(cp))
		pb.RegisterTriggerCapabilityServer(s, newTriggerCapabilityServer(c.brokerExt, cp))
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetTriggerReply{CapabilityID: cid}, nil
}

func (c *CapabilitiesRegistryServer) GetAction(ctx context.Context, request *pb.GetActionRequest) (*pb.GetActionReply, error) {
	id := request.Id
	cp, err := c.impl.GetAction(ctx, id)
	if err != nil {
		return nil, err
	}

	cid, _, err := c.serveNew("ActionCapability", func(s *grpc.Server) {
		pb.RegisterBaseCapabilityServer(s, newBaseCapabilityServer(cp))
		pb.RegisterCallbackExecutableServer(s, newCallbackExecutableServer(c.brokerExt, cp))
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetActionReply{CapabilityID: cid}, nil
}

func (c *CapabilitiesRegistryServer) GetConsensus(ctx context.Context, request *pb.GetConsensusRequest) (*pb.GetConsensusReply, error) {
	id := request.Id
	cp, err := c.impl.GetConsensus(ctx, id)
	if err != nil {
		return nil, err
	}

	cid, _, err := c.serveNew("ConsensusCapability", func(s *grpc.Server) {
		pb.RegisterBaseCapabilityServer(s, newBaseCapabilityServer(cp))
		pb.RegisterCallbackExecutableServer(s, newCallbackExecutableServer(c.brokerExt, cp))
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetConsensusReply{CapabilityID: cid}, nil
}

func (c *CapabilitiesRegistryServer) GetTarget(ctx context.Context, request *pb.GetTargetRequest) (*pb.GetTargetReply, error) {
	id := request.Id
	cp, err := c.impl.GetTarget(ctx, id)
	if err != nil {
		return nil, err
	}

	cid, _, err := c.serveNew("TargetCapability", func(s *grpc.Server) {
		pb.RegisterBaseCapabilityServer(s, newBaseCapabilityServer(cp))
		pb.RegisterCallbackExecutableServer(s, newCallbackExecutableServer(c.brokerExt, cp))
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetTargetReply{CapabilityID: cid}, nil
}

func (c *CapabilitiesRegistryServer) Add(ctx context.Context, request *pb.AddRequest) (*emptypb.Empty, error) {
	conn, err := c.dial(request.CapabilityID)
	if err != nil {
		return nil, err
	}

	var bc capabilities.BaseCapability
	switch request.Type {
	case pb.ExecuteAPIType_EXECUTE_API_TYPE_TRIGGER:
		bc = newTriggerCapabilityClient(c.brokerExt, conn)
	case pb.ExecuteAPIType_EXECUTE_API_TYPE_CALLBACK:
		bc = newCallbackCapabilityClient(c.brokerExt, conn)
	case pb.ExecuteAPIType_EXECUTE_API_TYPE_UNKNOWN:
		return nil, fmt.Errorf("unknown capability type %s", request.Type)
	}

	return &emptypb.Empty{}, c.impl.Add(ctx, bc)
}
