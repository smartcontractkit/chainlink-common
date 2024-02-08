package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"google.golang.org/grpc"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
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
		resource       resource
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
		resource = res
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
		resource = res
	default:
		return fmt.Errorf("could not add capability: unknown capability type: %T", cap)
	}

	_, err := c.client.Add(ctx, &pb.AddRequest{CapabilityID: capabilityID, Type: executeAPIType})
	if err != nil {
		c.closeAll(resource)
	}
	return err
}

type TriggerCapabilityClient struct {
	*triggerExecutableClient
	*baseCapabilityClient
}

func newTriggerCapabilityClient(brokerExt *brokerExt, conn *grpc.ClientConn) capabilities.TriggerCapability {
	return &TriggerCapabilityClient{
		triggerExecutableClient: newTriggerExecutableClient(brokerExt, conn),
		baseCapabilityClient:    newBaseCapabilityClient(brokerExt, conn),
	}
}

type CallbackCapabilityClient struct {
	*callbackExecutableClient
	*baseCapabilityClient
}

type callbackCapability interface {
	capabilities.CallbackExecutable
	capabilities.BaseCapability
}

func newCallbackCapabilityClient(brokerExt *brokerExt, conn *grpc.ClientConn) callbackCapability {
	return &CallbackCapabilityClient{
		callbackExecutableClient: newCallbackExecutableClient(brokerExt, conn),
		baseCapabilityClient:     newBaseCapabilityClient(brokerExt, conn),
	}
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

type baseCapabilityServer struct {
	pb.UnimplementedBaseCapabilityServer

	impl capabilities.BaseCapability
}

func newBaseCapabilityServer(impl capabilities.BaseCapability) *baseCapabilityServer {
	return &baseCapabilityServer{impl: impl}
}

var _ pb.BaseCapabilityServer = (*baseCapabilityServer)(nil)

func (c *baseCapabilityServer) Info(ctx context.Context, request *emptypb.Empty) (*pb.CapabilityInfoReply, error) {
	info, err := c.impl.Info(ctx)
	if err != nil {
		return nil, err
	}

	var ct pb.CapabilityType
	switch info.CapabilityType {
	case capabilities.CapabilityTypeTrigger:
		ct = pb.CapabilityType_CAPABILITY_TYPE_TRIGGER
	case capabilities.CapabilityTypeAction:
		ct = pb.CapabilityType_CAPABILITY_TYPE_ACTION
	case capabilities.CapabilityTypeConsensus:
		ct = pb.CapabilityType_CAPABILITY_TYPE_CONSENSUS
	case capabilities.CapabilityTypeTarget:
		ct = pb.CapabilityType_CAPABILITY_TYPE_TARGET
	}

	return &pb.CapabilityInfoReply{
		Id:             info.ID,
		CapabilityType: ct,
		Description:    info.Description,
		Version:        info.Version,
	}, nil
}

type baseCapabilityClient struct {
	grpc pb.BaseCapabilityClient
	*brokerExt
}

var _ capabilities.BaseCapability = (*baseCapabilityClient)(nil)

func newBaseCapabilityClient(brokerExt *brokerExt, conn *grpc.ClientConn) *baseCapabilityClient {
	return &baseCapabilityClient{grpc: pb.NewBaseCapabilityClient(conn), brokerExt: brokerExt}
}

func (c *baseCapabilityClient) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	resp, err := c.grpc.Info(ctx, &emptypb.Empty{})
	if err != nil {
		return capabilities.CapabilityInfo{}, err
	}

	var ct capabilities.CapabilityType
	switch resp.CapabilityType {
	case pb.CapabilityType_CAPABILITY_TYPE_TRIGGER:
		ct = capabilities.CapabilityTypeTrigger
	case pb.CapabilityType_CAPABILITY_TYPE_ACTION:
		ct = capabilities.CapabilityTypeAction
	case pb.CapabilityType_CAPABILITY_TYPE_CONSENSUS:
		ct = capabilities.CapabilityTypeConsensus
	case pb.CapabilityType_CAPABILITY_TYPE_TARGET:
		ct = capabilities.CapabilityTypeTarget
	case pb.CapabilityType_CAPABILITY_TYPE_UNKNOWN:
		return capabilities.CapabilityInfo{}, fmt.Errorf("invalid capability type: %s", ct)
	}

	return capabilities.CapabilityInfo{
		ID:             resp.Id,
		CapabilityType: ct,
		Description:    resp.Description,
		Version:        resp.Version,
	}, nil
}

type triggerCapabilityServer struct {
	pb.UnimplementedTriggerCapabilityServer
	*brokerExt

	impl capabilities.TriggerExecutable

	cancelFuncs map[string]func()
}

func newTriggerCapabilityServer(brokerExt *brokerExt, impl capabilities.TriggerExecutable) *triggerCapabilityServer {
	return &triggerCapabilityServer{
		impl:        impl,
		brokerExt:   brokerExt,
		cancelFuncs: map[string]func(){},
	}
}

var _ pb.TriggerCapabilityServer = (*triggerCapabilityServer)(nil)

func (t *triggerCapabilityServer) RegisterTrigger(ctx context.Context, request *pb.RegisterTriggerRequest) (*emptypb.Empty, error) {
	ch := make(chan capabilities.CapabilityResponse)

	conn, err := t.dial(request.CallbackID)
	if err != nil {
		return nil, err
	}

	connCtx, connCancel := context.WithCancel(context.Background())
	go callbackIssuer(connCtx, pb.NewCallbackClient(conn), ch, t.Logger)

	cr := request.CapabilityRequest
	md := cr.Metadata

	config, err := values.FromProto(cr.Config)
	if err != nil {
		connCancel()
		return nil, err
	}

	inputs, err := values.FromProto(cr.Inputs)
	if err != nil {
		connCancel()
		return nil, err
	}

	req := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:          md.WorkflowID,
			WorkflowExecutionID: md.WorkflowExecutionID,
		},
		Config: config.(*values.Map),
		Inputs: inputs.(*values.Map),
	}

	err = t.impl.RegisterTrigger(ctx, ch, req)
	if err != nil {
		connCancel()
		return nil, err
	}

	t.cancelFuncs[md.WorkflowID] = connCancel
	return &emptypb.Empty{}, nil
}

func (t *triggerCapabilityServer) UnregisterTrigger(ctx context.Context, request *pb.UnregisterTriggerRequest) (*emptypb.Empty, error) {
	req := request.CapabilityRequest
	md := req.Metadata

	config, err := values.FromProto(req.Config)
	if err != nil {
		return nil, err
	}

	inputs, err := values.FromProto(req.Inputs)
	if err != nil {
		return nil, err
	}

	err = t.impl.UnregisterTrigger(ctx, capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:          md.WorkflowID,
			WorkflowExecutionID: md.WorkflowExecutionID,
		},
		Inputs: inputs.(*values.Map),
		Config: config.(*values.Map),
	})
	if err != nil {
		return nil, err
	}

	cancelFunc := t.cancelFuncs[md.WorkflowID]
	if cancelFunc != nil {
		cancelFunc()
	}

	return &emptypb.Empty{}, nil
}

type triggerExecutableClient struct {
	grpc pb.TriggerCapabilityClient
	*brokerExt
}

var _ capabilities.TriggerExecutable = (*triggerExecutableClient)(nil)

func (t *triggerExecutableClient) RegisterTrigger(ctx context.Context, callback chan<- capabilities.CapabilityResponse, req capabilities.CapabilityRequest) error {
	cid, res, err := t.serveNew("Callback", func(s *grpc.Server) {
		pb.RegisterCallbackServer(s, newCallbackServer(callback))
	})
	if err != nil {
		return err
	}

	reqPb, err := toProto(req)
	if err != nil {
		t.closeAll(res)
		return err
	}

	r := &pb.RegisterTriggerRequest{
		CallbackID:        cid,
		CapabilityRequest: reqPb,
	}

	_, err = t.grpc.RegisterTrigger(ctx, r)
	if err != nil {
		t.closeAll(res)
	}
	return err
}

func (t *triggerExecutableClient) UnregisterTrigger(ctx context.Context, req capabilities.CapabilityRequest) error {
	reqPb, err := toProto(req)
	if err != nil {
		return err
	}

	r := &pb.UnregisterTriggerRequest{
		CapabilityRequest: reqPb,
	}

	_, err = t.grpc.UnregisterTrigger(ctx, r)
	return err
}

func newTriggerExecutableClient(brokerExt *brokerExt, conn *grpc.ClientConn) *triggerExecutableClient {
	return &triggerExecutableClient{grpc: pb.NewTriggerCapabilityClient(conn), brokerExt: brokerExt}
}

type callbackExecutableServer struct {
	pb.UnimplementedCallbackExecutableServer
	*brokerExt

	impl capabilities.CallbackExecutable

	cancelFuncs map[string]func()
}

func newCallbackExecutableServer(brokerExt *brokerExt, impl capabilities.CallbackExecutable) *callbackExecutableServer {
	return &callbackExecutableServer{
		impl:        impl,
		brokerExt:   brokerExt,
		cancelFuncs: map[string]func(){},
	}
}

var _ pb.CallbackExecutableServer = (*callbackExecutableServer)(nil)

func (c *callbackExecutableServer) RegisterToWorkflow(ctx context.Context, req *pb.RegisterToWorkflowRequest) (*emptypb.Empty, error) {
	config, err := values.FromProto(req.Config)
	if err != nil {
		return nil, err
	}

	err = c.impl.RegisterToWorkflow(ctx, capabilities.RegisterToWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID: req.Metadata.WorkflowID,
		},
		Config: config.(*values.Map),
	})
	return &emptypb.Empty{}, err
}

func (c *callbackExecutableServer) UnregisterFromWorkflow(ctx context.Context, req *pb.UnregisterFromWorkflowRequest) (*emptypb.Empty, error) {
	config, err := values.FromProto(req.Config)
	if err != nil {
		return nil, err
	}

	err = c.impl.UnregisterFromWorkflow(ctx, capabilities.UnregisterFromWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID: req.Metadata.WorkflowID,
		},
		Config: config.(*values.Map),
	})
	return &emptypb.Empty{}, err
}

func (c *callbackExecutableServer) Execute(ctx context.Context, req *pb.ExecuteRequest) (*emptypb.Empty, error) {
	ch := make(chan capabilities.CapabilityResponse)

	conn, err := c.dial(req.CallbackID)
	if err != nil {
		return nil, err
	}

	connCtx, connCancel := context.WithCancel(context.Background())
	go callbackIssuer(connCtx, pb.NewCallbackClient(conn), ch, c.Logger)

	cr := req.CapabilityRequest
	md := cr.Metadata

	config, err := values.FromProto(cr.Config)
	if err != nil {
		connCancel()
		return nil, err
	}

	inputs, err := values.FromProto(cr.Inputs)
	if err != nil {
		connCancel()
		return nil, err
	}

	r := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:          md.WorkflowID,
			WorkflowExecutionID: md.WorkflowExecutionID,
		},
		Config: config.(*values.Map),
		Inputs: inputs.(*values.Map),
	}

	err = c.impl.Execute(ctx, ch, r)
	if err != nil {
		connCancel()
		return nil, err
	}

	c.cancelFuncs[md.WorkflowID] = connCancel
	return &emptypb.Empty{}, nil
}

type callbackExecutableClient struct {
	grpc pb.CallbackExecutableClient
	*brokerExt
}

func newCallbackExecutableClient(brokerExt *brokerExt, conn *grpc.ClientConn) *callbackExecutableClient {
	return &callbackExecutableClient{
		grpc:      pb.NewCallbackExecutableClient(conn),
		brokerExt: brokerExt,
	}
}

var _ capabilities.CallbackExecutable = (*callbackExecutableClient)(nil)

func toProto(req capabilities.CapabilityRequest) (*pb.CapabilityRequest, error) {
	inputs := &values.Map{Underlying: map[string]values.Value{}}
	if req.Inputs != nil {
		inputs = req.Inputs
	}

	inputsPb, err := inputs.Proto()
	if err != nil {
		return nil, err
	}

	config := &values.Map{Underlying: map[string]values.Value{}}
	if req.Config != nil {
		config = req.Config
	}

	configPb, err := config.Proto()
	if err != nil {
		return nil, err
	}

	return &pb.CapabilityRequest{
		Metadata: &pb.RequestMetadata{
			WorkflowID:          req.Metadata.WorkflowID,
			WorkflowExecutionID: req.Metadata.WorkflowExecutionID,
		},
		Inputs: inputsPb,
		Config: configPb,
	}, nil
}

func (c *callbackExecutableClient) Execute(ctx context.Context, callback chan<- capabilities.CapabilityResponse, req capabilities.CapabilityRequest) error {
	cid, res, err := c.serveNew("Callback", func(s *grpc.Server) {
		pb.RegisterCallbackServer(s, newCallbackServer(callback))
	})
	if err != nil {
		return err
	}

	reqPb, err := toProto(req)
	if err != nil {
		c.closeAll(res)
		return nil
	}

	r := &pb.ExecuteRequest{
		CallbackID:        cid,
		CapabilityRequest: reqPb,
	}

	_, err = c.grpc.Execute(ctx, r)
	if err != nil {
		c.closeAll(res)
	}
	return err
}

func (c *callbackExecutableClient) UnregisterFromWorkflow(ctx context.Context, req capabilities.UnregisterFromWorkflowRequest) error {
	config := &values.Map{Underlying: map[string]values.Value{}}
	if req.Config != nil {
		config = req.Config
	}

	configPb, err := config.Proto()
	if err != nil {
		return err
	}

	r := &pb.UnregisterFromWorkflowRequest{
		Config: configPb,
		Metadata: &pb.RegistrationMetadata{
			WorkflowID: req.Metadata.WorkflowID,
		},
	}

	_, err = c.grpc.UnregisterFromWorkflow(ctx, r)
	return err
}

func (c *callbackExecutableClient) RegisterToWorkflow(ctx context.Context, req capabilities.RegisterToWorkflowRequest) error {
	config := &values.Map{Underlying: map[string]values.Value{}}
	if req.Config != nil {
		config = req.Config
	}

	configPb, err := config.Proto()
	if err != nil {
		return err
	}

	r := &pb.RegisterToWorkflowRequest{
		Config: configPb,
		Metadata: &pb.RegistrationMetadata{
			WorkflowID: req.Metadata.WorkflowID,
		},
	}

	_, err = c.grpc.RegisterToWorkflow(ctx, r)
	return err
}

type callbackServer struct {
	pb.UnimplementedCallbackServer
	ch chan<- capabilities.CapabilityResponse

	isClosed bool
	mu       sync.RWMutex
}

func newCallbackServer(ch chan<- capabilities.CapabilityResponse) *callbackServer {
	return &callbackServer{ch: ch}
}

func (c *callbackServer) SendResponse(ctx context.Context, req *pb.CapabilityResponse) (*emptypb.Empty, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosed {
		return nil, errors.New("cannot send response: the underlying channel has been closed")
	}

	val, err := values.FromProto(req.Value)
	if err != nil {
		return nil, err
	}

	err = nil
	if req.Error != "" {
		err = errors.New(req.Error)
	}
	resp := capabilities.CapabilityResponse{
		Value: val,
		Err:   err,
	}
	c.ch <- resp
	return &emptypb.Empty{}, nil
}

func (c *callbackServer) CloseCallback(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	close(c.ch)
	c.isClosed = true
	return &emptypb.Empty{}, nil
}

func callbackIssuer(ctx context.Context, client pb.CallbackClient, callbackChannel chan capabilities.CapabilityResponse, logger logger.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		case resp, isOpen := <-callbackChannel:
			if !isOpen {
				_, err := client.CloseCallback(ctx, &emptypb.Empty{})
				if err != nil {
					logger.Error("could not close upstream callback", err)
				}
				return
			}
			var (
				val    *valuespb.Value
				errStr string
			)
			if resp.Err != nil {
				errStr = resp.Err.Error()
			}
			if resp.Value != nil {
				v, err := resp.Value.Proto()
				if err != nil {
					errStr = err.Error()
				} else {
					val = v
				}
			}
			cr := &pb.CapabilityResponse{
				Error: errStr,
				Value: val,
			}

			_, err := client.SendResponse(ctx, cr)
			if err != nil {
				logger.Error("error sending callback response", err)
			}
		}
	}
}
