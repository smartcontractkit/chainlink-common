package capability

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

type TriggerCapabilityClient struct {
	*triggerExecutableClient
	*baseCapabilityClient
}

func NewTriggerCapabilityClient(brokerExt *net.BrokerExt, conn net.ClientConnInterface) capabilities.TriggerCapability {
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

func NewExecutableCapabilityClient(brokerExt *net.BrokerExt, conn net.ClientConnInterface) ExecutableCapability {
	return &ExecutableCapabilityClient{
		executableClient:     newExecutableClient(brokerExt, conn),
		baseCapabilityClient: newBaseCapabilityClient(brokerExt, conn),
	}
}

type CombinedCapabilityClient struct {
	*executableClient
	*baseCapabilityClient
	*triggerExecutableClient
}

func NewCombinedCapabilityClient(brokerExt *net.BrokerExt, conn net.ClientConnInterface) ExecutableCapability {
	return &CombinedCapabilityClient{
		executableClient:        newExecutableClient(brokerExt, conn),
		baseCapabilityClient:    newBaseCapabilityClient(brokerExt, conn),
		triggerExecutableClient: newTriggerExecutableClient(brokerExt, conn),
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
	case capabilities.CapabilityTypeCombined:
		ct = capabilitiespb.CapabilityType_CAPABILITY_TYPE_COMBINED
	case capabilities.CapabilityTypeUnknown:
		ct = capabilitiespb.CapabilityType_CAPABILITY_TYPE_UNKNOWN
	default:
		ct = capabilitiespb.CapabilityType_CAPABILITY_TYPE_UNKNOWN
	}

	types := make([]string, len(info.SpendTypes))
	for idx, sType := range info.SpendTypes {
		types[idx] = string(sType)
	}

	return &capabilitiespb.CapabilityInfoReply{
		Id:             info.ID,
		CapabilityType: ct,
		Description:    info.Description,
		IsLocal:        info.IsLocal,
		SpendTypes:     types,
	}
}

type baseCapabilityClient struct {
	c    net.ClientConnInterface
	grpc capabilitiespb.BaseCapabilityClient
	*net.BrokerExt
}

var _ capabilities.BaseCapability = (*baseCapabilityClient)(nil)

func newBaseCapabilityClient(brokerExt *net.BrokerExt, conn net.ClientConnInterface) *baseCapabilityClient {
	return &baseCapabilityClient{c: conn, grpc: capabilitiespb.NewBaseCapabilityClient(conn), BrokerExt: brokerExt}
}
func (c *baseCapabilityClient) GetState() connectivity.State {
	return c.c.GetState()
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
	case capabilitiespb.CapabilityTypeCombined:
		ct = capabilities.CapabilityTypeCombined
	case capabilitiespb.CapabilityTypeUnknown:
		return capabilities.CapabilityInfo{}, fmt.Errorf("invalid capability type: %s", ct)
	}

	types := make([]capabilities.CapabilitySpendType, len(resp.SpendTypes))
	for idx, sType := range resp.SpendTypes {
		types[idx] = capabilities.CapabilitySpendType(sType)
	}

	return capabilities.CapabilityInfo{
		ID:             resp.Id,
		CapabilityType: ct,
		Description:    resp.Description,
		IsLocal:        resp.IsLocal,
		SpendTypes:     types,
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
		// the first message sent to the client will be an ack or error message, this is done in order to syncronize the client and server and avoid
		// errors to unregister not found triggers. If the error is not nil, we send an error message to the client and return the error
		msg := &capabilitiespb.TriggerResponseMessage{
			Message: &capabilitiespb.TriggerResponseMessage_Response{
				Response: &capabilitiespb.TriggerResponse{
					Error: err.Error(),
				},
			},
		}
		return server.Send(msg)
	}

	// Send ACK response to client
	msg := &capabilitiespb.TriggerResponseMessage{
		Message: &capabilitiespb.TriggerResponseMessage_Ack{
			Ack: &emptypb.Empty{},
		},
	}
	if err = server.Send(msg); err != nil {
		return fmt.Errorf("failed sending ACK response for trigger registration %s: %w", request, err)
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

	// manage cancelation of gRPC client stream by trigger ID
	mu          sync.Mutex
	cancelFuncs map[string]func()
}

func (t *triggerExecutableClient) AckEvent(ctx context.Context, triggerId string, eventId string) error {
	req := &capabilitiespb.AckEventRequest{
		TriggerId: triggerId,
		EventId:   eventId,
	}
	_, err := t.grpc.AckEvent(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to call AckEvent: %w", err)
	}
	return nil
}

func (t *triggerExecutableClient) RegisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	ch, cancel, err := t.registerTrigger(ctx, req)
	if err != nil {
		cancel()
		return nil, err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// if exists, clean up previous stream spawned for matching trigger ID
	if prevCancel, ok := t.cancelFuncs[req.TriggerID]; ok {
		prevCancel()
		delete(t.cancelFuncs, req.TriggerID)
	}

	t.cancelFuncs[req.TriggerID] = cancel

	return ch, nil
}

// registerTrigger returns a cancel func for shutting down the returned channel of trigger responses.
func (t *triggerExecutableClient) registerTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, context.CancelFunc, error) {
	// ctx will outlive the parent ctx to keep the gRPC client connection stream alive.
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.WithoutCancel(ctx))
	responseStream, err := t.grpc.RegisterTrigger(ctx, pb.TriggerRegistrationRequestToProto(req))
	if err != nil {
		return nil, cancel, fmt.Errorf("error registering trigger: %w", err)
	}

	// In order to ensure the registration is successful, we need to wait for the first message from the server.
	// This will be an ack or error message. If the error is not nil, we return an error.
	ackMsg, err := responseStream.Recv()
	if err != nil {
		return nil, cancel, fmt.Errorf("failed to receive registering trigger ack message: %w", err)
	}

	if ackMsg.GetAck() == nil {
		return nil, cancel, errors.New(fmt.Sprintf("failed registering trigger: %s", ackMsg.GetResponse().GetError()))
	}

	ch, err := forwardTriggerResponsesToChannel(ctx, responseStream.Recv)
	if err != nil {
		return nil, cancel, fmt.Errorf("failed to start forwarding messages from stream: %w", err)
	}

	return ch, cancel, nil
}

func (t *triggerExecutableClient) UnregisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) error {
	_, err := t.grpc.UnregisterTrigger(ctx, pb.TriggerRegistrationRequestToProto(req))
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if cancel, ok := t.cancelFuncs[req.TriggerID]; ok {
		cancel()
		return nil
	}

	t.Logger.Warnw("attempted to cleanup stream that was not found", "triggerID", req.TriggerID, "workflowID", req.Metadata.WorkflowID)
	return nil
}

func newTriggerExecutableClient(brokerExt *net.BrokerExt, conn grpc.ClientConnInterface) *triggerExecutableClient {
	return &triggerExecutableClient{
		grpc:        capabilitiespb.NewTriggerExecutableClient(conn),
		BrokerExt:   brokerExt,
		cancelFuncs: make(map[string]func()),
	}
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
			WorkflowID:  req.Metadata.WorkflowId,
			ReferenceID: req.Metadata.ReferenceId,
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
			WorkflowID:  req.Metadata.WorkflowId,
			ReferenceID: req.Metadata.ReferenceId,
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

	var responseMessage *capabilitiespb.CapabilityResponse
	response, err := c.impl.Execute(server.Context(), req)
	if err != nil {
		var capabilityError caperrors.Error
		if errors.As(err, &capabilityError) {
			responseMessage = &capabilitiespb.CapabilityResponse{Error: capabilityError.SerializeToString()}
		} else {
			// All other errors are treated as private visibility and are marked as such to prevent accidental or malicious
			// reporting of sensitive information by prefixing the error message with the remote reportable identifier.
			responseMessage = &capabilitiespb.CapabilityResponse{Error: caperrors.PrePendPrivateVisibilityIdentifier(err.Error())}
		}
	} else {
		responseMessage = pb.CapabilityResponseToProto(response)
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

func newExecutableClient(brokerExt *net.BrokerExt, conn grpc.ClientConnInterface) *executableClient {
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

	resp, err := responseStream.Recv()
	if err != nil {
		return capabilities.CapabilityResponse{}, fmt.Errorf("error waiting for response message: %w", err)
	}

	if resp.Error != "" {
		return capabilities.CapabilityResponse{}, caperrors.DeserializeErrorFromString(resp.Error)
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
			WorkflowId:  req.Metadata.WorkflowID,
			ReferenceId: req.Metadata.ReferenceID,
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
			WorkflowId:  req.Metadata.WorkflowID,
			ReferenceId: req.Metadata.ReferenceID,
		},
	}

	_, err := c.grpc.RegisterToWorkflow(ctx, r)
	return err
}

func forwardTriggerResponsesToChannel(
	ctx context.Context,
	receive func() (*capabilitiespb.TriggerResponseMessage, error),
) (<-chan capabilities.TriggerResponse, error) {
	responseCh := make(chan capabilities.TriggerResponse)

	send := func(resp capabilities.TriggerResponse) {
		select {
		case responseCh <- resp:
		case <-ctx.Done():
		}
	}

	go func() {
		defer close(responseCh)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			message, err := receive()
			if errors.Is(err, io.EOF) {
				return
			}

			if err != nil {
				resp := capabilities.TriggerResponse{
					Err: err,
				}
				send(resp)
				return
			}

			resp := message.GetResponse()
			if resp == nil {
				resp := capabilities.TriggerResponse{
					Err: errors.New("unexpected message type when receiving response: expected response"),
				}
				send(resp)
				return
			}

			r, err := pb.TriggerResponseFromProto(resp)
			if err != nil {
				r = capabilities.TriggerResponse{
					Err: err,
				}
			}
			send(r)
		}
	}()

	return responseCh, nil
}
