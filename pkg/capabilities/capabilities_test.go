package capabilities

import (
	"bytes"
	"context"
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

func Test_CapabilityInfo(t *testing.T) {
	ci, err := NewCapabilityInfo(
		"capability-id@1.0.0",
		CapabilityTypeAction,
		"This is a mock capability that doesn't do anything.",
	)
	require.NoError(t, err)

	gotCi, err := ci.Info(t.Context())
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

	gotCi, err = ci.Info(t.Context())
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

	gotCi, err = ci.Info(t.Context())
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
	resp, err := mcwe.Execute(t.Context(), req)

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
	_, err := mcwe.Execute(t.Context(), req)
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

	assert.Equal(t, DefaultExecutableRequestTimeout, rec.RequestTimeout)
	assert.Equal(t, DefaultServerMaxParallelRequests, rec.ServerMaxParallelRequests)
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
		invalidMap, err := values.NewMap(map[string]any{
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
		assert.ErrorContains(t, err, "nil map")
	})
	t.Run("nil underlying map", func(t *testing.T) {
		event := &OCRTriggerEvent{}
		err := event.FromMap(&values.Map{
			Underlying: nil,
		})
		assert.ErrorContains(t, err, "nil underlying map")
	})

}

func TestParseID(t *testing.T) {
	for _, tc := range []struct {
		id      string
		name    string
		labels  map[string]string
		version string
	}{
		{id: "foo", name: "foo"},
		{id: "foo@1.0.0", name: "foo", version: "1.0.0"},
		{id: "foo:k_v@1.0.0", name: "foo", labels: map[string]string{"k": "v"}, version: "1.0.0"},
		{id: "foo:k_v:k2_v2:k3@1.0.0", name: "foo", labels: map[string]string{"k": "v", "k2": "v2", "k3": ""}, version: "1.0.0"},
		//TODO more
	} {
		t.Run(tc.id, func(t *testing.T) {
			if tc.labels == nil {
				tc.labels = map[string]string{}
			}

			name, labels, version := ParseID(tc.id)
			assert.Equal(t, tc.name, name)
			assert.Equal(t, tc.labels, maps.Collect(labels))
			assert.Equal(t, tc.version, version)
		})
	}
}

func TestChainSelectorLabel(t *testing.T) {
	for _, tc := range []struct {
		id     string
		cs     *uint64
		errMsg string
	}{
		{"none@v1.0.0", nil, ""},
		{"kv:ChainSelector_1@v1.0.0", ptr[uint64](1), ""},
		{"kk:ChainSelector:1@v1.0.0", ptr[uint64](1), ""},
		{"kv-others:k_v:ChainSelector_1@v1.0.0", ptr[uint64](1), ""},
		{"kk-others:k_v:ChainSelector:1@v1.0.0", ptr[uint64](1), ""},

		{"kv:ChainSelector_foo@v1.0.0", ptr[uint64](1), "invalid chain selector"},
		{"kk:ChainSelector:bar@v1.0.0", ptr[uint64](1), "invalid chain selector"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			_, labels, _ := ParseID(tc.id)
			cs, err := ChainSelectorLabel(labels)
			if tc.errMsg != "" {
				require.ErrorContains(t, err, tc.errMsg)
			} else {
				require.Equal(t, tc.cs, cs)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }

func TestRequestMetadata_ContextWithCRE(t *testing.T) {
	ctx := t.Context()
	require.Equal(t, "", contexts.CREValue(ctx).Org)

	// set it
	ctx = contexts.WithCRE(ctx, contexts.CRE{Org: "org-id"})
	require.Equal(t, "org-id", contexts.CREValue(ctx).Org)

	// preserve it
	md := RequestMetadata{WorkflowOwner: "owner-id", WorkflowID: "workflow-id"}
	ctx = md.ContextWithCRE(ctx)
	require.Equal(t, "org-id", contexts.CREValue(ctx).Org)
}

func TestRegistrationMetadata_ContextWithCRE(t *testing.T) {
	ctx := t.Context()
	require.Equal(t, "", contexts.CREValue(ctx).Org)

	// set it
	ctx = contexts.WithCRE(ctx, contexts.CRE{Org: "org-id"})
	require.Equal(t, "org-id", contexts.CREValue(ctx).Org)

	// preserve it
	md := RegistrationMetadata{WorkflowOwner: "owner-id", WorkflowID: "workflow-id"}
	ctx = md.ContextWithCRE(ctx)
	require.Equal(t, "org-id", contexts.CREValue(ctx).Org)
}
