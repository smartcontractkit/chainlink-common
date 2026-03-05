package capabilities

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

var ErrNeitherValueNorAny = errors.New("neither value nor any provided")

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
	var migrated bool
	if any == nil {
		// Check if the underlying concrete value is nil
		if v, ok := value.(*values.Map); ok && v == nil {
			return migrated, ErrNeitherValueNorAny
		}
		if value == nil {
			return migrated, ErrNeitherValueNorAny
		}
		if err := value.UnwrapTo(into); err != nil {
			return migrated, fmt.Errorf("failed to transform value to proto: %w", err)
		}
		return migrated, nil
	}

	migrated = true
	if err := any.UnmarshalTo(into); err != nil {
		return migrated, fmt.Errorf("failed to transform any to proto: %w", err)
	}

	return migrated, nil
}

// Execute is a helper function for capabilities that allows them to use their native types for input, config, and response
// while adhering to the standard capability interface.
func Execute[I, C, O proto.Message](
	ctx context.Context,
	request CapabilityRequest,
	input I,
	config C,
	exec func(context.Context, RequestMetadata, I, C) (O, ResponseMetadata, error)) (CapabilityResponse, error) {

	response := CapabilityResponse{}
	migrated, err := UnwrapRequest(request, config, input)
	if err != nil {
		return response, fmt.Errorf("error when unwrapping request: %w", err)
	}

	output, metadata, err := exec(ctx, request.Metadata, input, config)
	response.Metadata = metadata
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
	stop <-chan struct{},
	triggerType string,
	request TriggerRegistrationRequest,
	message I,
	fn func(context.Context, string, RequestMetadata, I) (<-chan TriggerAndId[O], caperrors.Error),
) (<-chan TriggerResponse, error) {
	migrated, err := FromValueOrAny(request.Config, request.Payload, message)
	if err != nil {
		return nil, fmt.Errorf("error when unwrapping request: %w", err)
	}

	response := make(chan TriggerResponse)
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
				case <-stop:
					return
				}
			case <-stop:
				return
			}
		}
	}()

	return response, nil
}
