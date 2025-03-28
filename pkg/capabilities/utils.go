package capabilities

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func UnwrapRequest(request CapabilityRequest, config proto.Message, value proto.Message) (bool, error) {
	migrated, err := fromValueOrAny(request.Inputs, request.Request, value)
	if err != nil {
		return migrated, err
	}

	// TODO config
	_, err = fromValueOrAny(request.Config /* TODO */, nil, config)
	if err != nil {
		return migrated, err
	}

	return migrated, nil
}

func UnwrapResponse(response CapabilityResponse, value proto.Message) (bool, error) {
	migrated, err := fromValueOrAny(response.Value, response.ResponseValue, value)
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

func fromValueOrAny(value values.Value, any *anypb.Any, into proto.Message) (bool, error) {
	if any != nil {
		if err := any.UnmarshalTo(into); err != nil {
			return false, err
		}
		return true, nil
	}

	err := value.UnwrapTo(into)
	return false, err
}
