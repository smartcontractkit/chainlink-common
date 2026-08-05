package client

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// The servers below expose an in-process capability at an address, so a registry
// that dials capabilities can reach it. RegisterCapability is the entry point.

type baseCapabilityServer struct {
	capabilitiespb.UnimplementedBaseCapabilityServer
	impl capabilities.BaseCapability
}

var _ capabilitiespb.BaseCapabilityServer = (*baseCapabilityServer)(nil)

func newBaseCapabilityServer(impl capabilities.BaseCapability) *baseCapabilityServer {
	return &baseCapabilityServer{impl: impl}
}

func (s *baseCapabilityServer) Info(ctx context.Context, _ *emptypb.Empty) (*capabilitiespb.CapabilityInfoReply, error) {
	info, err := s.impl.Info(ctx)
	if err != nil {
		return nil, err
	}

	ct, err := capabilitiespb.CapabilityTypeToProto(info.CapabilityType)
	if err != nil {
		return nil, err
	}

	spendTypes := make([]string, len(info.SpendTypes))
	for i, t := range info.SpendTypes {
		spendTypes[i] = string(t)
	}

	return &capabilitiespb.CapabilityInfoReply{
		Id:             info.ID,
		CapabilityType: ct,
		Description:    info.Description,
		IsLocal:        info.IsLocal,
		SpendTypes:     spendTypes,
	}, nil
}

type executableServer struct {
	capabilitiespb.UnimplementedExecutableServer
	impl capabilities.Executable
}

var _ capabilitiespb.ExecutableServer = (*executableServer)(nil)

func newExecutableServer(impl capabilities.Executable) *executableServer {
	return &executableServer{impl: impl}
}

func (s *executableServer) RegisterToWorkflow(ctx context.Context, req *capabilitiespb.RegisterToWorkflowRequest) (*emptypb.Empty, error) {
	config, err := values.FromMapValueProto(req.Config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config into map: %w", err)
	}
	err = s.impl.RegisterToWorkflow(ctx, capabilities.RegisterToWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID:  req.Metadata.GetWorkflowId(),
			ReferenceID: req.Metadata.GetReferenceId(),
		},
		Config: config,
	})
	return &emptypb.Empty{}, err
}

func (s *executableServer) UnregisterFromWorkflow(ctx context.Context, req *capabilitiespb.UnregisterFromWorkflowRequest) (*emptypb.Empty, error) {
	config, err := values.FromMapValueProto(req.Config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config into map: %w", err)
	}
	err = s.impl.UnregisterFromWorkflow(ctx, capabilities.UnregisterFromWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID:  req.Metadata.GetWorkflowId(),
			ReferenceID: req.Metadata.GetReferenceId(),
		},
		Config: config,
	})
	return &emptypb.Empty{}, err
}

func (s *executableServer) Execute(reqpb *capabilitiespb.CapabilityRequest, server capabilitiespb.Executable_ExecuteServer) error {
	req, err := capabilitiespb.CapabilityRequestFromProto(reqpb)
	if err != nil {
		return fmt.Errorf("could not unmarshal capability request: %w", err)
	}

	var responseMessage *capabilitiespb.CapabilityResponse
	response, err := s.impl.Execute(server.Context(), req)
	switch {
	case err == nil:
		responseMessage = capabilitiespb.CapabilityResponseToProto(response)
	default:
		var capabilityError caperrors.Error
		if errors.As(err, &capabilityError) {
			responseMessage = &capabilitiespb.CapabilityResponse{Error: capabilityError.SerializeToString()}
		} else {
			// Errors that are not capability errors are treated as private: they
			// are marked so an unclassified message cannot leak to a remote
			// caller.
			responseMessage = &capabilitiespb.CapabilityResponse{
				Error: caperrors.PrePendPrivateVisibilityIdentifier(err.Error()),
			}
		}
	}

	if err := server.Send(responseMessage); err != nil {
		return fmt.Errorf("error sending response for execute request: %w", err)
	}
	return nil
}

type triggerExecutableServer struct {
	capabilitiespb.UnimplementedTriggerExecutableServer
	lggr logger.Logger
	impl capabilities.TriggerExecutable
}

var _ capabilitiespb.TriggerExecutableServer = (*triggerExecutableServer)(nil)

func newTriggerExecutableServer(lggr logger.Logger, impl capabilities.TriggerExecutable) *triggerExecutableServer {
	return &triggerExecutableServer{lggr: lggr, impl: impl}
}

func (s *triggerExecutableServer) AckEvent(ctx context.Context, req *capabilitiespb.AckEventRequest) (*emptypb.Empty, error) {
	if err := s.impl.AckEvent(ctx, req.TriggerId, req.EventId, req.Method); err != nil {
		return nil, fmt.Errorf("error acking event: %w", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *triggerExecutableServer) RegisterTrigger(request *capabilitiespb.TriggerRegistrationRequest, server capabilitiespb.TriggerExecutable_RegisterTriggerServer) error {
	req, err := capabilitiespb.TriggerRegistrationRequestFromProto(request)
	if err != nil {
		return fmt.Errorf("could not unmarshal capability request: %w", err)
	}

	responseCh, err := s.impl.RegisterTrigger(server.Context(), req)
	if err != nil {
		// The first message is always an ack or an error, so the client can tell
		// a failed registration from an idle subscription and does not later try
		// to unregister a trigger that never existed.
		errorString := err.Error()
		var capErr caperrors.Error
		if errors.As(err, &capErr) {
			errorString = capErr.SerializeToString()
		}
		return server.Send(&capabilitiespb.TriggerResponseMessage{
			Message: &capabilitiespb.TriggerResponseMessage_Response{
				Response: &capabilitiespb.TriggerResponse{Error: errorString},
			},
		})
	}

	if err := server.Send(&capabilitiespb.TriggerResponseMessage{
		Message: &capabilitiespb.TriggerResponseMessage_Ack{Ack: &emptypb.Empty{}},
	}); err != nil {
		return fmt.Errorf("failed sending ACK response for trigger registration: %w", err)
	}

	defer func() {
		// Always unregister so the underlying capability releases resources even
		// if the client simply drops the stream.
		if err := s.impl.UnregisterTrigger(server.Context(), req); err != nil {
			s.lggr.Errorw("error unregistering trigger", "err", err)
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
			if err := server.Send(&capabilitiespb.TriggerResponseMessage{
				Message: &capabilitiespb.TriggerResponseMessage_Response{
					Response: capabilitiespb.TriggerResponseToProto(resp),
				},
			}); err != nil {
				return fmt.Errorf("error sending trigger response: %w", err)
			}
		}
	}
}

func (s *triggerExecutableServer) UnregisterTrigger(ctx context.Context, request *capabilitiespb.TriggerRegistrationRequest) (*emptypb.Empty, error) {
	req, err := capabilitiespb.TriggerRegistrationRequestFromProto(request)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal capability request: %w", err)
	}
	if err := s.impl.UnregisterTrigger(ctx, req); err != nil {
		return nil, fmt.Errorf("error unregistering trigger: %w", err)
	}
	return &emptypb.Empty{}, nil
}

// RegisterCapability serves impl on srv: BaseCapability always, plus Executable
// and/or TriggerExecutable according to t.
//
// A capability host calls this on its own gRPC server and then announces that
// server's address with Client.Add, which is the whole registration protocol.
// The set of services per capability type is kept here so the two sides of the
// registry cannot disagree about what a handle's type implies.
func RegisterCapability(lggr logger.Logger, srv grpc.ServiceRegistrar, impl capabilities.BaseCapability, t capabilities.CapabilityType) error {
	switch t {
	case capabilities.CapabilityTypeTrigger:
		trigger, ok := impl.(capabilities.TriggerCapability)
		if !ok {
			return errors.New("capability is typed as a trigger but does not implement TriggerCapability")
		}
		capabilitiespb.RegisterTriggerExecutableServer(srv, newTriggerExecutableServer(lggr, trigger))
	case capabilities.CapabilityTypeAction, capabilities.CapabilityTypeTarget, capabilities.CapabilityTypeConsensus:
		executable, ok := impl.(capabilities.ExecutableCapability)
		if !ok {
			return errors.New("capability is typed as executable but does not implement ExecutableCapability")
		}
		capabilitiespb.RegisterExecutableServer(srv, newExecutableServer(executable))
	case capabilities.CapabilityTypeCombined:
		trigger, ok := impl.(capabilities.TriggerCapability)
		if !ok {
			return errors.New("combined capability does not implement TriggerCapability")
		}
		executable, ok := impl.(capabilities.ExecutableCapability)
		if !ok {
			return errors.New("combined capability does not implement ExecutableCapability")
		}
		capabilitiespb.RegisterTriggerExecutableServer(srv, newTriggerExecutableServer(lggr, trigger))
		capabilitiespb.RegisterExecutableServer(srv, newExecutableServer(executable))
	case capabilities.CapabilityTypeUnknown:
		// Only the base capability service is registered.
	default:
		return fmt.Errorf("unknown capability type %s", t)
	}
	capabilitiespb.RegisterBaseCapabilityServer(srv, newBaseCapabilityServer(impl))
	return nil
}
