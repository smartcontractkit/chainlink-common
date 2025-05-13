package capabilities

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// UnwrapRequest extracts the input and config from the request, returning true if they were migrated to use pbany.Any values.
func UnwrapRequest(request CapabilityRequest, config proto.Message, value proto.Message) (bool, error) {
	migrated, err := FromValueOrAny(request.Inputs, request.Payload, value)
	if err != nil {
		return migrated, err
	}

	_, err = FromValueOrAny(request.Config, request.ConfigPayload, config)
	if err != nil {
		return migrated, err
	}

	return migrated, nil
}

// UnwrapResponse extracts the response, returning true if they were migrated to use pbany.Any values.
func UnwrapResponse(response CapabilityResponse, value proto.Message) (bool, error) {
	migrated, err := FromValueOrAny(response.Value, response.Payload, value)
	if err != nil {
		return migrated, err
	}

	return migrated, nil
}

// SetResponse sets the response payload based on whether it was migrated to use pbany.Any values.
func SetResponse(response *CapabilityResponse, migrated bool, value proto.Message) error {
	if migrated {
		wrapped, err := anypb.New(value)
		if err != nil {
			return err
		}
		response.Payload = wrapped
		return nil
	}

	wrapped, err := values.WrapMap(value)
	if err != nil {
		return err
	}

	response.Value = wrapped
	return nil
}

// FromValueOrAny extracts the value from either a values.Value or an anypb.Any, returning true if the value was migrated to use pbany.Any.
func FromValueOrAny(value values.Value, any *anypb.Any, into proto.Message) (bool, error) {
	if any == nil {
		if value == nil {
			return false, errors.New("neither value nor any provided")
		}
		err := value.UnwrapTo(into)
		return false, err
	}

	err := any.UnmarshalTo(into)
	return true, err
}

// Execute is a helper function for capabilities that allows them to use their native types for input, config, and response
// while adhering to the standard capability interface.
func Execute[I, C, O proto.Message](
	ctx context.Context,
	request CapabilityRequest,
	input I,
	config C,
	exec func(context.Context, RequestMetadata, I, C) (O, error)) (CapabilityResponse, error) {

	response := CapabilityResponse{}
	migrated, err := UnwrapRequest(request, config, input)
	if err != nil {
		return response, fmt.Errorf("error when unwrapping request: %w", err)
	}

	output, err := exec(ctx, request.Metadata, input, config)
	if err != nil {
		return response, err
	}

	if err = SetResponse(&response, migrated, output); err != nil {
		return response, fmt.Errorf("error when setting response: %w", err)
	}

	return response, nil
}

type TriggerAndId[T proto.Message] struct {
	Trigger T
	Id      string
}

// RegisterTrigger is a helper function for capabilities that allows them to use their native types for input, config, and response
// while adhering to the standard capability interface.
func RegisterTrigger[I, O proto.Message](
	ctx context.Context,
	triggerType string,
	request TriggerRegistrationRequest,
	message I,
	fn func(context.Context, string, RequestMetadata, I) (<-chan TriggerAndId[O], error),
) (<-chan TriggerResponse, error) {
	migrated, err := FromValueOrAny(request.Config, request.Payload, message)
	if err != nil {
		return nil, fmt.Errorf("error when unwrapping request: %w", err)
	}

	response := make(chan TriggerResponse, 100)
	respCh, err := fn(ctx, request.TriggerID, request.Metadata, message)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(response)
		for {
			select {
			case resp, open := <-respCh:
				if !open {
					return
				}

				tr := TriggerResponse{
					Event: TriggerEvent{
						TriggerType: triggerType,
						ID:          resp.Id,
					},
				}
				if migrated {
					wrapped, err := anypb.New(resp.Trigger)
					tr.Err = err
					tr.Event.Payload = wrapped
				} else {
					wrapped, err := values.WrapMap(resp.Trigger)
					tr.Err = err
					tr.Event.Outputs = wrapped
				}
				select {
				case response <- tr:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return response, nil
}
