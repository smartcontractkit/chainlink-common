package capabilities

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func Test_CapabilityInfo(t *testing.T) {
	ci, err := NewCapabilityInfo(
		"capability-id@1.0.0",
		CapabilityTypeAction,
		"This is a mock capability that doesn't do anything.",
	)
	require.NoError(t, err)

	gotCi, err := ci.Info(tests.Context(t))
	require.NoError(t, err)
	require.Equal(t, ci.Version(), "1.0.0")
	assert.Equal(t, ci, gotCi)

	ci, err = NewCapabilityInfo(
		// add build metadata and sha
		"capability-id@1.0.0+build.1234.sha-5678",
		CapabilityTypeAction,
		"This is a mock capability that doesn't do anything.",
	)
	require.NoError(t, err)

	gotCi, err = ci.Info(tests.Context(t))
	require.NoError(t, err)
	require.Equal(t, ci.Version(), "1.0.0+build.1234.sha-5678")
	assert.Equal(t, ci, gotCi)

	// prerelease
	ci, err = NewCapabilityInfo(
		"capability-id@1.0.0-beta",
		CapabilityTypeAction,
		"This is a mock capability that doesn't do anything.",
	)
	require.NoError(t, err)

	gotCi, err = ci.Info(tests.Context(t))
	require.NoError(t, err)
	require.Equal(t, ci.Version(), "1.0.0-beta")
	assert.Equal(t, ci, gotCi)
}

func Test_CapabilityInfo_Invalid(t *testing.T) {
	_, err := NewCapabilityInfo(
		"capability-id@2.0.0",
		CapabilityTypeUnknown,
		"This is a mock capability that doesn't do anything.",
	)
	assert.ErrorContains(t, err, "invalid capability type")

	_, err = NewCapabilityInfo(
		"&!!!",
		CapabilityTypeAction,
		"This is a mock capability that doesn't do anything.",
	)
	assert.ErrorContains(t, err, "invalid id")

	_, err = NewCapabilityInfo(
		"mock-capability@v1.0.0",
		CapabilityTypeAction,
		"This is a mock capability that doesn't do anything.",
	)
	assert.ErrorContains(t, err, "invalid id")

	_, err = NewCapabilityInfo(
		"mock-capability@1.0",
		CapabilityTypeAction,
		"This is a mock capability that doesn't do anything.",
	)
	assert.ErrorContains(t, err, "invalid id")

	_, err = NewCapabilityInfo(
		"mock-capability@1",
		CapabilityTypeAction,
		"This is a mock capability that doesn't do anything.",
	)

	assert.ErrorContains(t, err, "invalid id")
	_, err = NewCapabilityInfo(
		strings.Repeat("n", 256),
		CapabilityTypeAction,
		"This is a mock capability that doesn't do anything.",
	)
	assert.ErrorContains(t, err, "exceeds max length 128")
}

type mockCapabilityWithExecute struct {
	Executable
	CapabilityInfo
	ExecuteFn func(ctx context.Context, req CapabilityRequest) (CapabilityResponse, error)
}

func (m *mockCapabilityWithExecute) Execute(ctx context.Context, req CapabilityRequest) (CapabilityResponse, error) {
	return m.ExecuteFn(ctx, req)
}

func Test_ExecuteSyncReturnValue(t *testing.T) {
	v := map[string]any{"hello": "world"}
	mcwe := &mockCapabilityWithExecute{
		ExecuteFn: func(ctx context.Context, req CapabilityRequest) (CapabilityResponse, error) {
			val, err := values.NewMap(v)
			if err != nil {
				return CapabilityResponse{}, err
			}

			return CapabilityResponse{Value: val}, nil
		},
	}
	req := CapabilityRequest{}
	resp, err := mcwe.Execute(tests.Context(t), req)

	require.NoError(t, err)
	unwrappedValue, err := resp.Value.Unwrap()
	require.NoError(t, err)
	assert.Equal(t, v, unwrappedValue)
}

func Test_ExecuteSyncCapabilitySetupErrors(t *testing.T) {
	expectedErr := errors.New("something went wrong during setup")
	mcwe := &mockCapabilityWithExecute{
		ExecuteFn: func(ctx context.Context, req CapabilityRequest) (CapabilityResponse, error) {
			return CapabilityResponse{}, expectedErr
		},
	}
	req := CapabilityRequest{}
	_, err := mcwe.Execute(tests.Context(t), req)
	assert.ErrorContains(t, err, expectedErr.Error())
}

func Test_MustNewCapabilityInfo(t *testing.T) {
	assert.Panics(t, func() {
		MustNewCapabilityInfo(
			"capability-id",
			CapabilityTypeAction,
			"This is a mock capability that doesn't do anything.",
		)
	})
}

func Test_RemoteExecutableConfig_ApplyDefaults(t *testing.T) {
	rec := &RemoteExecutableConfig{}
	rec.ApplyDefaults()

	assert.Equal(t, DefaultRegistrationRefresh, rec.RegistrationRefresh)
	assert.Equal(t, DefaultRegistrationExpiry, rec.RegistrationExpiry)
}

func TestOCRTriggerEvent_ToMapFromMap(t *testing.T) {
	// Create test signatures
	sigs := []OCRAttributedOnchainSignature{
		{
			Signature: []byte("first_signature_data"),
			Signer:    1,
		},
		{
			Signature: []byte("second_signature_data"),
			Signer:    2,
		},
		{
			Signature: []byte("third_signature_data"),
			Signer:    3,
		},
	}

	testCases := []struct {
		name  string
		event *OCRTriggerEvent
	}{
		{
			name: "typical event with all fields populated",
			event: &OCRTriggerEvent{
				ConfigDigest: []byte("test_config_digest_data"),
				SeqNr:        123456789,
				Report:       []byte("marshaled_report_payload_data"),
				Sigs:         sigs,
			},
		},

		{
			name: "event with empty slices",
			event: &OCRTriggerEvent{
				ConfigDigest: []byte{},
				SeqNr:        987654321,
				Report:       []byte{},
				Sigs:         []OCRAttributedOnchainSignature{},
			},
		},
		{
			name: "event with nil slices",
			event: &OCRTriggerEvent{
				ConfigDigest: nil,
				SeqNr:        555555,
				Report:       nil,
				Sigs:         nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert the original event to a values.Map
			valueMap, err := tc.event.ToMap()
			require.NoError(t, err, "ToMap should not error")
			require.NotNil(t, valueMap, "ToMap should return a non-nil map")

			// Create a new empty event
			reconstructedEvent := &OCRTriggerEvent{}

			// Convert the values.Map back to an event
			err = reconstructedEvent.FromMap(valueMap)
			require.NoError(t, err, "FromMap should not error")

			// Validate that the reconstructed event matches the original
			assert.Equal(t, tc.event.SeqNr, reconstructedEvent.SeqNr, "SeqNr should match")

			// Compare byte slices
			assert.True(t, bytes.Equal(tc.event.ConfigDigest, reconstructedEvent.ConfigDigest),
				"ConfigDigest should match")
			assert.True(t, bytes.Equal(tc.event.Report, reconstructedEvent.Report),
				"Report should match")

			// Compare signatures
			assert.Equal(t, len(tc.event.Sigs), len(reconstructedEvent.Sigs),
				"Number of signatures should match")

			for i := 0; i < len(tc.event.Sigs); i++ {
				if i < len(reconstructedEvent.Sigs) {
					assert.Equal(t, tc.event.Sigs[i].Signer, reconstructedEvent.Sigs[i].Signer,
						"Signature signer should match at index %d", i)
					assert.True(t, bytes.Equal(tc.event.Sigs[i].Signature, reconstructedEvent.Sigs[i].Signature),
						"Signature data should match at index %d", i)
				}
			}
		})
	}

	// Test error handling

	t.Run("invalid map missing key", func(t *testing.T) {
		invalidMap, err := values.NewMap(map[string]interface{}{
			"WrongKey": "value",
		})
		require.NoError(t, err)

		event := &OCRTriggerEvent{}
		err = event.FromMap(invalidMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing key")
	})

	t.Run("nil map", func(t *testing.T) {
		event := &OCRTriggerEvent{}
		err := event.FromMap(nil)
		assert.Error(t, err)
	})

}
