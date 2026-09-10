package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

type baseCapabilityClient struct {
	grpc pb.BaseCapabilityClient
}

var _ capabilities.BaseCapability = (*baseCapabilityClient)(nil)

func (c *baseCapabilityClient) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	reply, err := c.grpc.Info(ctx, &emptypb.Empty{})
	if err != nil {
		return capabilities.CapabilityInfo{}, err
	}
	return pb.InfoReplyToInfo(reply)
}

type executableClient struct {
	grpc pb.ExecutableClient
}

var _ capabilities.Executable = (*executableClient)(nil)

func (c *executableClient) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	responseStream, err := c.grpc.Execute(ctx, pb.CapabilityRequestToProto(req))
	if err != nil {
		return capabilities.CapabilityResponse{}, caperrors.NewPublicSystemError(
			fmt.Errorf("error executing capability request: %w", err), caperrors.Unavailable)
	}

	resp, err := responseStream.Recv()
	if err != nil {
		return capabilities.CapabilityResponse{}, caperrors.NewPublicSystemError(
			fmt.Errorf("error waiting for response message: %w", err), caperrors.Unavailable)
	}

	if resp.Error != "" {
		return capabilities.CapabilityResponse{}, caperrors.DeserializeErrorFromString(resp.Error)
	}

	r, err := pb.CapabilityResponseFromProto(resp)
	if err != nil {
		return capabilities.CapabilityResponse{}, caperrors.NewPublicSystemError(
			fmt.Errorf("could not unmarshal response: %w", err), caperrors.Internal)
	}
	return r, nil
}

func (c *executableClient) RegisterToWorkflow(ctx context.Context, req capabilities.RegisterToWorkflowRequest) error {
	_, err := c.grpc.RegisterToWorkflow(ctx, &pb.RegisterToWorkflowRequest{
		Config: values.ProtoMap(orEmptyMap(req.Config)),
		Metadata: &pb.RegistrationMetadata{
			WorkflowId:  req.Metadata.WorkflowID,
			ReferenceId: req.Metadata.ReferenceID,
		},
	})
	return err
}

func (c *executableClient) UnregisterFromWorkflow(ctx context.Context, req capabilities.UnregisterFromWorkflowRequest) error {
	_, err := c.grpc.UnregisterFromWorkflow(ctx, &pb.UnregisterFromWorkflowRequest{
		Config: values.ProtoMap(orEmptyMap(req.Config)),
		Metadata: &pb.RegistrationMetadata{
			WorkflowId:  req.Metadata.WorkflowID,
			ReferenceId: req.Metadata.ReferenceID,
		},
	})
	return err
}

func orEmptyMap(m *values.Map) *values.Map {
	if m != nil {
		return m
	}
	return &values.Map{Underlying: map[string]values.Value{}}
}

type triggerExecutableClient struct {
	grpc pb.TriggerExecutableClient
	lggr logger.Logger

	// cancelFuncs tracks the stream backing each trigger registration, keyed by
	// trigger ID, so UnregisterTrigger can tear it down.
	mu          sync.Mutex
	cancelFuncs map[string]func()
}

var _ capabilities.TriggerExecutable = (*triggerExecutableClient)(nil)

func (t *triggerExecutableClient) RegisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	ch, cancel, err := t.registerTrigger(ctx, req)
	if err != nil {
		cancel()
		return nil, err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Re-registering the same trigger ID replaces the previous stream.
	if prevCancel, ok := t.cancelFuncs[req.TriggerID]; ok {
		prevCancel()
		delete(t.cancelFuncs, req.TriggerID)
	}
	t.cancelFuncs[req.TriggerID] = cancel
	return ch, nil
}

func (t *triggerExecutableClient) registerTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, context.CancelFunc, error) {
	// The stream must outlive the calling ctx: the caller's ctx bounds the
	// registration call, not the subscription it creates.
	streamCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))

	responseStream, err := t.grpc.RegisterTrigger(streamCtx, pb.TriggerRegistrationRequestToProto(req))
	if err != nil {
		return nil, cancel, fmt.Errorf("error registering trigger: %w", err)
	}

	// The server's first message is an ack or an error, so registration failure
	// surfaces here rather than as a silent empty stream.
	ackMsg, err := responseStream.Recv()
	if err != nil {
		return nil, cancel, fmt.Errorf("failed to receive registering trigger ack message: %w", err)
	}
	if ackMsg.GetAck() == nil {
		return nil, cancel, caperrors.DeserializeErrorFromString(ackMsg.GetResponse().GetError())
	}

	return forwardTriggerResponseStream(streamCtx, responseStream.Recv), cancel, nil
}

func (t *triggerExecutableClient) UnregisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) error {
	if _, err := t.grpc.UnregisterTrigger(ctx, pb.TriggerRegistrationRequestToProto(req)); err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if cancel, ok := t.cancelFuncs[req.TriggerID]; ok {
		cancel()
		delete(t.cancelFuncs, req.TriggerID)
		return nil
	}

	t.lggr.Warnw("attempted to clean up stream that was not found",
		"triggerID", req.TriggerID, "workflowID", req.Metadata.WorkflowID)
	return nil
}

func (t *triggerExecutableClient) AckEvent(ctx context.Context, triggerID, eventID, method string) error {
	_, err := t.grpc.AckEvent(ctx, &pb.AckEventRequest{
		TriggerId: triggerID,
		EventId:   eventID,
		Method:    method,
	})
	if err != nil {
		return fmt.Errorf("failed to call AckEvent: %w", err)
	}
	return nil
}

func forwardTriggerResponseStream(ctx context.Context, receive func() (*pb.TriggerResponseMessage, error)) <-chan capabilities.TriggerResponse {
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
				send(capabilities.TriggerResponse{Err: err})
				return
			}

			resp := message.GetResponse()
			if resp == nil {
				send(capabilities.TriggerResponse{
					Err: errors.New("unexpected message type when receiving response: expected response"),
				})
				return
			}

			r, err := pb.TriggerResponseFromProto(resp)
			if err != nil {
				send(capabilities.TriggerResponse{Err: err})
				return
			}
			send(r)
		}
	}()

	return responseCh
}

func newBaseCapabilityClient(cc grpc.ClientConnInterface) *baseCapabilityClient {
	return &baseCapabilityClient{grpc: pb.NewBaseCapabilityClient(cc)}
}

func newExecutableClient(cc grpc.ClientConnInterface) *executableClient {
	return &executableClient{grpc: pb.NewExecutableClient(cc)}
}

func newTriggerExecutableClient(lggr logger.Logger, cc grpc.ClientConnInterface) *triggerExecutableClient {
	return &triggerExecutableClient{
		grpc:        pb.NewTriggerExecutableClient(cc),
		lggr:        lggr,
		cancelFuncs: map[string]func(){},
	}
}
