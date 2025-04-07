package capabilities

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func UnwrapRequest(request CapabilityRequest, config proto.Message, value proto.Message) (bool, error) {
	migrated, err := FromValueOrAny(request.Inputs, request.Request, value)
	if err != nil {
		return migrated, err
	}

	// TODO config
	_, err = FromValueOrAny(request.Config /* TODO */, nil, config)
	if err != nil {
		return migrated, err
	}

	return migrated, nil
}

func UnwrapResponse(response CapabilityResponse, value proto.Message) (bool, error) {
	migrated, err := FromValueOrAny(response.Value, response.ResponseValue, value)
	if err != nil {
		return migrated, err
	}

	return migrated, nil
}

func SetResponse(response CapabilityResponse, migrated bool, value proto.Message) error {
	if migrated {
		wrapped, err := anypb.New(value)
		if err != nil {
			return err
		}
		response.ResponseValue = wrapped
	}

	wrapped, err := values.WrapMap(value)
	if err != nil {
		return err
	}

	response.Value = wrapped
	return nil
}

func FromValueOrAny(value values.Value, any *anypb.Any, into proto.Message) (bool, error) {
	if any != nil {
		if err := any.UnmarshalTo(into); err != nil {
			return false, err
		}
		return true, nil
	}

	err := value.UnwrapTo(into)
	return false, err
}

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

	if err = SetResponse(response, migrated, output); err != nil {
		return response, fmt.Errorf("error when setting response: %w", err)
	}

	return response, nil
}

type TriggerAndId[T proto.Message] struct {
	Trigger T
	Id      string
}

func RegisterTrigger[C, O proto.Message](
	ctx context.Context,
	triggerType string,
	request TriggerRegistrationRequest,
	message C,
	fn func(context.Context, RequestMetadata, C) (<-chan TriggerAndId[O], error)) (<-chan TriggerResponse, error) {
	migrated, err := FromValueOrAny(request.Config, request.Request, message)
	if err != nil {
		return nil, fmt.Errorf("error when unwrapping request: %w", err)
	}

	// TODO size?
	response := make(chan TriggerResponse, 100)
	respCh, err := fn(ctx, request.Metadata, message)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case resp := <-respCh:
				tr := TriggerResponse{
					Event: TriggerEvent{
						TriggerType: triggerType,
						ID:          resp.Id,
					},
				}
				if migrated {
					wrapped, err := anypb.New(resp.Trigger)
					tr.Err = err
					tr.Event.Value = wrapped
				} else {
					wrapped, err := values.WrapMap(resp.Trigger)
					tr.Err = err
					tr.Event.Outputs = wrapped
				}
				response <- tr
			case <-ctx.Done():
				close(response)
				return
			}
		}
	}()

	return response, nil
}
