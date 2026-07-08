package vault

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestObservationErrorField_SHAStableForOkObservations(t *testing.T) {
	t.Parallel()

	obs := &Observation{
		Id:          "request-1",
		RequestType: RequestType_DELETE_SECRETS,
		Response: &Observation_DeleteSecretsResponse{
			DeleteSecretsResponse: &DeleteSecretsResponse{
				Responses: []*DeleteSecretResponse{{Success: false}},
			},
		},
	}

	before, err := proto.MarshalOptions{Deterministic: true}.Marshal(obs)
	require.NoError(t, err)

	roundTrip := &Observation{}
	require.NoError(t, proto.Unmarshal(before, roundTrip))
	require.Nil(t, roundTrip.GetError())

	after, err := proto.MarshalOptions{Deterministic: true}.Marshal(roundTrip)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestObservationErrorField_RoundTrip(t *testing.T) {
	t.Parallel()

	obs := &Observation{
		Id:          "request-1",
		RequestType: RequestType_CREATE_SECRETS,
		Error: &ObservationError{
			Message: "request is not valid",
		},
	}

	b, err := proto.MarshalOptions{Deterministic: true}.Marshal(obs)
	require.NoError(t, err)

	out := &Observation{}
	require.NoError(t, proto.Unmarshal(b, out))
	require.Equal(t, "request is not valid", out.GetError().GetMessage())
	require.Nil(t, out.GetCreateSecretsResponse())
}
