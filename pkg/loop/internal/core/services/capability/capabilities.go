package capability

import (
	"context"
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type ActionCapabilityClient struct {
	*executableClient
	*baseCapabilityClient
}

func NewActionCapabilityClient(brokerExt *net.BrokerExt, conn *grpc.ClientConn) capabilities.ActionCapability {
	return &ActionCapabilityClient{
		executableClient:     newExecutableClient(brokerExt, conn),
		baseCapabilityClient: newBaseCapabilityClient(brokerExt, conn),
	}
}

type ConsensusCapabilityClient struct {
	*executableClient
	*baseCapabilityClient
}

func NewConsensusCapabilityClient(brokerExt *net.BrokerExt, conn *grpc.ClientConn) capabilities.ConsensusCapability {
	return &ConsensusCapabilityClient{
		executableClient:     newExecutableClient(brokerExt, conn),
		baseCapabilityClient: newBaseCapabilityClient(brokerExt, conn),
	}
}

type TargetCapabilityClient struct {
	*executableClient
	*baseCapabilityClient
}

func NewTargetCapabilityClient(brokerExt *net.BrokerExt, conn *grpc.ClientConn) capabilities.TargetCapability {
	return &TargetCapabilityClient{
		executableClient:     newExecutableClient(brokerExt, conn),
		baseCapabilityClient: newBaseCapabilityClient(brokerExt, conn),
	}
}

type TriggerCapabilityClient struct {
	*triggerExecutableClient
	*baseCapabilityClient
}

func NewTriggerCapabilityClient(brokerExt *net.BrokerExt, conn *grpc.ClientConn) capabilities.TriggerCapability {
	return &TriggerCapabilityClient{
		triggerExecutableClient: newTriggerExecutableClient(brokerExt, conn),
		baseCapabilityClient:    newBaseCapabilityClient(brokerExt, conn),
	}
}

type ExecutableCapabilityClient struct {
	*executableClient
	*baseCapabilityClient
}

type ExecutableCapability interface {
	capabilities.Executable
	capabilities.BaseCapability
}

func NewExecutableCapabilityClient(brokerExt *net.BrokerExt, conn *grpc.ClientConn) ExecutableCapability {
	return &ExecutableCapabilityClient{
		executableClient:     newExecutableClient(brokerExt, conn),
		baseCapabilityClient: newBaseCapabilityClient(brokerExt, conn),
	}
}

func RegisterExecutableCapabilityServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl ExecutableCapability) error {
	bext := &net.BrokerExt{
		BrokerConfig: brokerCfg,
		Broker:       broker,
	}
	capabilitiespb.RegisterExecutableServer(server, newExecutableServer(bext, impl))
	capabilitiespb.RegisterBaseCapabilityServer(server, newBaseCapabilityServer(impl))
	return nil
}

func RegisterTriggerCapabilityServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl capabilities.TriggerCapability) error {
	bext := &net.BrokerExt{
		BrokerConfig: brokerCfg,
		Broker:       broker,
	}
	capabilitiespb.RegisterTriggerExecutableServer(server, newTriggerExecutableServer(bext, impl))
	capabilitiespb.RegisterBaseCapabilityServer(server, newBaseCapabilityServer(impl))
	return nil
}

type baseCapabilityServer struct {
	capabilitiespb.UnimplementedBaseCapabilityServer

	impl capabilities.BaseCapability
}

func newBaseCapabilityServer(impl capabilities.BaseCapability) *baseCapabilityServer {
	return &baseCapabilityServer{impl: impl}
}

var _ capabilitiespb.BaseCapabilityServer = (*baseCapabilityServer)(nil)

func (c *baseCapabilityServer) Info(ctx context.Context, request *emptypb.Empty) (*capabilitiespb.CapabilityInfoReply, error) {
	info, err := c.impl.Info(ctx)
	if err != nil {
		return nil, err
	}

	return InfoToReply(info), nil
}

func InfoToReply(info capabilities.CapabilityInfo) *capabilitiespb.CapabilityInfoReply {
	var ct capabilitiespb.CapabilityType
	switch info.CapabilityType {
	case capabilities.CapabilityTypeTrigger:
		ct = capabilitiespb.CapabilityType_CAPABILITY_TYPE_TRIGGER
	case capabilities.CapabilityTypeAction:
		ct = capabilitiespb.CapabilityType_CAPABILITY_TYPE_ACTION
	case capabilities.CapabilityTypeConsensus:
		ct = capabilitiespb.CapabilityType_CAPABILITY_TYPE_CONSENSUS
	case capabilities.CapabilityTypeTarget:
		ct = capabilitiespb.CapabilityType_CAPABILITY_TYPE_TARGET
	case capabilities.CapabilityTypeUnknown:
		ct = capabilitiespb.CapabilityType_CAPABILITY_TYPE_UNKNOWN
	default:
		ct = capabilitiespb.CapabilityType_CAPABILITY_TYPE_UNKNOWN
	}

	return &capabilitiespb.CapabilityInfoReply{
		Id:             info.ID,
		CapabilityType: ct,
		Description:    info.Description,
		IsLocal:        info.IsLocal,
	}
}

type baseCapabilityClient struct {
	grpc capabilitiespb.BaseCapabilityClient
	*net.BrokerExt
}

var _ capabilities.BaseCapability = (*baseCapabilityClient)(nil)

func newBaseCapabilityClient(brokerExt *net.BrokerExt, conn *grpc.ClientConn) *baseCapabilityClient {
	return &baseCapabilityClient{grpc: capabilitiespb.NewBaseCapabilityClient(conn), BrokerExt: brokerExt}
}

func (c *baseCapabilityClient) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	reply, err := c.grpc.Info(ctx, &emptypb.Empty{})
	if err != nil {
		return capabilities.CapabilityInfo{}, err
	}

	return InfoReplyToInfo(reply)
}

func InfoReplyToInfo(resp *capabilitiespb.CapabilityInfoReply) (capabilities.CapabilityInfo, error) {
	var ct capabilities.CapabilityType
	switch resp.CapabilityType {
	case capabilitiespb.CapabilityTypeTrigger:
		ct = capabilities.CapabilityTypeTrigger
	case capabilitiespb.CapabilityTypeAction:
		ct = capabilities.CapabilityTypeAction
	case capabilitiespb.CapabilityTypeConsensus:
		ct = capabilities.CapabilityTypeConsensus
	case capabilitiespb.CapabilityTypeTarget:
		ct = capabilities.CapabilityTypeTarget
	case capabilitiespb.CapabilityTypeUnknown:
		return capabilities.CapabilityInfo{}, fmt.Errorf("invalid capability type: %s", ct)
	}

	return capabilities.CapabilityInfo{
		ID:             resp.Id,
		CapabilityType: ct,
		Description:    resp.Description,
		IsLocal:        resp.IsLocal,
	}, nil
}

type triggerExecutableServer struct {
	capabilitiespb.UnimplementedTriggerExecutableServer
	*net.BrokerExt

	impl capabilities.TriggerExecutable
}

func newTriggerExecutableServer(brokerExt *net.BrokerExt, impl capabilities.TriggerExecutable) *triggerExecutableServer {
	return &triggerExecutableServer{
		impl:      impl,
		BrokerExt: brokerExt,
	}
}

var _ capabilitiespb.TriggerExecutableServer = (*triggerExecutableServer)(nil)

func (t *triggerExecutableServer) RegisterTrigger(request *capabilitiespb.TriggerRegistrationRequest,
	server capabilitiespb.TriggerExecutable_RegisterTriggerServer) error {
	req, err := pb.TriggerRegistrationRequestFromProto(request)
	if err != nil {
		return fmt.Errorf("could not unmarshal capability request: %w", err)
	}
	responseCh, err := t.impl.RegisterTrigger(server.Context(), req)
	if err != nil {
		return fmt.Errorf("error registering trigger: %w", err)
	}

	defer func() {
		// Always attempt to unregister the trigger to ensure any related resources are cleaned up
		err = t.impl.UnregisterTrigger(server.Context(), req)
		if err != nil {
			t.Logger.Error("error unregistering trigger", "err", err)
		}
	}()

	for {
		select {
		case <-server.Context().Done():
			return nil
		case resp, ok := <-responseCh:
			if !ok {
				return nil
			}

			msg := &capabilitiespb.TriggerResponseMessage{
				Message: &capabilitiespb.TriggerResponseMessage_Response{
					Response: pb.TriggerResponseToProto(resp),
				},
			}
			if err = server.Send(msg); err != nil {
				return fmt.Errorf("error sending response for trigger %s: %w", request, err)
			}
		}
	}
}

func (t *triggerExecutableServer) UnregisterTrigger(ctx context.Context, request *capabilitiespb.TriggerRegistrationRequest) (*emptypb.Empty, error) {
	req, err := pb.TriggerRegistrationRequestFromProto(request)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal capability request: %w", err)
	}
	if err := t.impl.UnregisterTrigger(ctx, req); err != nil {
		return nil, fmt.Errorf("error unregistering trigger: %w", err)
	}

	return &emptypb.Empty{}, nil
}

type triggerExecutableClient struct {
	grpc capabilitiespb.TriggerExecutableClient
	*net.BrokerExt
}

func (t *triggerExecutableClient) RegisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	responseStream, err := t.grpc.RegisterTrigger(ctx, pb.TriggerRegistrationRequestToProto(req))
	if err != nil {
		return nil, fmt.Errorf("error registering trigger: %w", err)
	}

	return forwardTriggerResponsesToChannel(ctx, t.Logger, req, responseStream.Recv)
}

func (t *triggerExecutableClient) UnregisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) error {
	_, err := t.grpc.UnregisterTrigger(ctx, pb.TriggerRegistrationRequestToProto(req))
	return err
}

func newTriggerExecutableClient(brokerExt *net.BrokerExt, conn *grpc.ClientConn) *triggerExecutableClient {
	return &triggerExecutableClient{grpc: capabilitiespb.NewTriggerExecutableClient(conn), BrokerExt: brokerExt}
}

type executableServer struct {
	capabilitiespb.UnimplementedExecutableServer
	*net.BrokerExt

	impl capabilities.Executable

	cancelFuncs map[string]func()
}

func newExecutableServer(brokerExt *net.BrokerExt, impl capabilities.Executable) *executableServer {
	return &executableServer{
		impl:        impl,
		BrokerExt:   brokerExt,
		cancelFuncs: map[string]func(){},
	}
}

var _ capabilitiespb.ExecutableServer = (*executableServer)(nil)

func (c *executableServer) RegisterToWorkflow(ctx context.Context, req *capabilitiespb.RegisterToWorkflowRequest) (*emptypb.Empty, error) {
	config, err := values.FromMapValueProto(req.Config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config into map: %w", err)
	}

	err = c.impl.RegisterToWorkflow(ctx, capabilities.RegisterToWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID: req.Metadata.WorkflowId,
		},
		Config: config,
	})
	return &emptypb.Empty{}, err
}

func (c *executableServer) UnregisterFromWorkflow(ctx context.Context, req *capabilitiespb.UnregisterFromWorkflowRequest) (*emptypb.Empty, error) {
	config, err := values.FromMapValueProto(req.Config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config into map: %w", err)
	}

	err = c.impl.UnregisterFromWorkflow(ctx, capabilities.UnregisterFromWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID: req.Metadata.WorkflowId,
		},
		Config: config,
	})
	return &emptypb.Empty{}, err
}

func (c *executableServer) Execute(reqpb *capabilitiespb.CapabilityRequest, server capabilitiespb.Executable_ExecuteServer) error {
	req, err := pb.CapabilityRequestFromProto(reqpb)
	if err != nil {
		return fmt.Errorf("could not unmarshal capability request: %w", err)
	}

	var responseMessage *capabilitiespb.ResponseMessage
	response, err := c.impl.Execute(server.Context(), req)
	if err != nil {
		responseMessage = &capabilitiespb.ResponseMessage{
			Message: &capabilitiespb.ResponseMessage_Response{
				Response: &capabilitiespb.CapabilityResponse{Error: err.Error()},
			},
		}
	} else {
		responseMessage = &capabilitiespb.ResponseMessage{
			Message: &capabilitiespb.ResponseMessage_Response{
				Response: pb.CapabilityResponseToProto(response),
			},
		}
	}

	if err = server.Send(responseMessage); err != nil {
		return fmt.Errorf("error sending response for execute request %s: %w", reqpb, err)
	}

	return nil
}

type executableClient struct {
	grpc capabilitiespb.ExecutableClient
	*net.BrokerExt
}

func newExecutableClient(brokerExt *net.BrokerExt, conn *grpc.ClientConn) *executableClient {
	return &executableClient{
		grpc:      capabilitiespb.NewExecutableClient(conn),
		BrokerExt: brokerExt,
	}
}

var _ capabilities.Executable = (*executableClient)(nil)

func (c *executableClient) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	responseStream, err := c.grpc.Execute(ctx, pb.CapabilityRequestToProto(req))
	if err != nil {
		return capabilities.CapabilityResponse{}, fmt.Errorf("error executing capability request: %w", err)
	}

	message, err := responseStream.Recv()
	if err != nil {
		return capabilities.CapabilityResponse{}, fmt.Errorf("error waiting for response message: %w", err)
	}

	resp := message.GetResponse()
	if resp == nil {
		return capabilities.CapabilityResponse{}, fmt.Errorf("nil message when receiving response")
	}

	if resp.Error != "" {
		return capabilities.CapabilityResponse{}, errors.New(resp.Error)
	}

	r, err := pb.CapabilityResponseFromProto(resp)
	if err != nil {
		return capabilities.CapabilityResponse{}, fmt.Errorf("could not unmarshal response: %w", err)
	}

	return r, err
}

func (c *executableClient) UnregisterFromWorkflow(ctx context.Context, req capabilities.UnregisterFromWorkflowRequest) error {
	config := &values.Map{Underlying: map[string]values.Value{}}
	if req.Config != nil {
		config = req.Config
	}

	r := &capabilitiespb.UnregisterFromWorkflowRequest{
		Config: values.ProtoMap(config),
		Metadata: &capabilitiespb.RegistrationMetadata{
			WorkflowId: req.Metadata.WorkflowID,
		},
	}

	_, err := c.grpc.UnregisterFromWorkflow(ctx, r)
	return err
}

func (c *executableClient) RegisterToWorkflow(ctx context.Context, req capabilities.RegisterToWorkflowRequest) error {
	config := &values.Map{Underlying: map[string]values.Value{}}
	if req.Config != nil {
		config = req.Config
	}

	r := &capabilitiespb.RegisterToWorkflowRequest{
		Config: values.ProtoMap(config),
		Metadata: &capabilitiespb.RegistrationMetadata{
			WorkflowId: req.Metadata.WorkflowID,
		},
	}

	_, err := c.grpc.RegisterToWorkflow(ctx, r)
	return err
}

func forwardTriggerResponsesToChannel(ctx context.Context, logger logger.Logger, req capabilities.TriggerRegistrationRequest, receive func() (*capabilitiespb.TriggerResponseMessage, error)) (<-chan capabilities.TriggerResponse, error) {
	responseCh := make(chan capabilities.TriggerResponse)

	go func() {
		defer close(responseCh)
		for {
			message, err := receive()
			if errors.Is(err, io.EOF) {
				return
			}

			if err != nil {
				resp := capabilities.TriggerResponse{
					Err: err,
				}
				select {
				case responseCh <- resp:
				case <-ctx.Done():
				}
				return
			}

			resp := message.GetResponse()
			if resp == nil {
				resp := capabilities.TriggerResponse{
					Err: errors.New("unexpected message type when receiving response: expected response"),
				}
				select {
				case responseCh <- resp:
				case <-ctx.Done():
				}
				return
			}

			r, err := pb.TriggerResponseFromProto(resp)
			if err != nil {
				r = capabilities.TriggerResponse{
					Err: err,
				}
			}

			select {
			case responseCh <- r:
			case <-ctx.Done():
				return
			}
		}
	}()

	return responseCh, nil
}
