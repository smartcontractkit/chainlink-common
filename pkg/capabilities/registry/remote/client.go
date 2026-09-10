package remote

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
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

type baseCapabilityClient struct {
	grpc pb.BaseCapabilityClient
	conn grpc.ClientConnInterface
}

var _ capabilities.BaseCapability = (*baseCapabilityClient)(nil)

func (c *baseCapabilityClient) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	reply, err := c.grpc.Info(ctx, &emptypb.Empty{})
	if err != nil {
		return capabilities.CapabilityInfo{}, err
	}
	return pb.InfoReplyToInfo(reply)
}

// GetState reports the state of the underlying connection when it exposes one
// (a [grpc.ClientConn] or the go-plugin broker's conn wrapper both do). The
// registry's atomic capability wrappers use it to decide whether a registered
// capability may be replaced.
func (c *baseCapabilityClient) GetState() connectivity.State {
	if sg, ok := c.conn.(interface{ GetState() connectivity.State }); ok {
		return sg.GetState()
	}
	return connectivity.Shutdown
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
	return &baseCapabilityClient{grpc: pb.NewBaseCapabilityClient(cc), conn: cc}
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

// TriggerCapabilityClient serves BaseCapability + TriggerExecutable over conn.
type TriggerCapabilityClient struct {
	*baseCapabilityClient
	*triggerExecutableClient
}

var _ capabilities.TriggerCapability = (*TriggerCapabilityClient)(nil)

// ExecutableCapabilityClient serves BaseCapability + Executable over conn.
type ExecutableCapabilityClient struct {
	*baseCapabilityClient
	*executableClient
}

var _ capabilities.ExecutableCapability = (*ExecutableCapabilityClient)(nil)

// CombinedCapabilityClient serves all three surfaces over conn.
type CombinedCapabilityClient struct {
	*baseCapabilityClient
	*executableClient
	*triggerExecutableClient
}

var _ capabilities.ExecutableAndTriggerCapability = (*CombinedCapabilityClient)(nil)

// Wrap builds the capability surface capType promises, served over conn.
//
// Exported so a registry holding a Handle (an ID, a type and a callback
// address) can dial that address itself and get back a real
// capabilities.BaseCapability, and so the go-plugin registry client
// (pkg/loop/internal/core/services/capability) can build the same surfaces over
// its broker connections.
//
// Mirrors RegisterCapability on the serving side: same type, same services.
func Wrap(lggr logger.Logger, conn grpc.ClientConnInterface, capType capabilities.CapabilityType) capabilities.BaseCapability {
	base := newBaseCapabilityClient(conn)
	switch capType {
	case capabilities.CapabilityTypeTrigger:
		return &TriggerCapabilityClient{
			baseCapabilityClient:    base,
			triggerExecutableClient: newTriggerExecutableClient(lggr, conn),
		}
	case capabilities.CapabilityTypeAction,
		capabilities.CapabilityTypeTarget,
		capabilities.CapabilityTypeConsensus:
		return &ExecutableCapabilityClient{
			baseCapabilityClient: base,
			executableClient:     newExecutableClient(conn),
		}
	case capabilities.CapabilityTypeCombined:
		return &CombinedCapabilityClient{
			baseCapabilityClient:    base,
			executableClient:        newExecutableClient(conn),
			triggerExecutableClient: newTriggerExecutableClient(lggr, conn),
		}
	case capabilities.CapabilityTypeUnknown:
		// Only the base capability service is registered, so only the base
		// surface exists to wrap.
		return base
	default:
		panic(fmt.Sprintf("unknown capability type %s", capType))
	}
}
