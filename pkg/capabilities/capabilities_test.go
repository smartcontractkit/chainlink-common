package capabilities

import (
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

			return CapabilityResponse{val}, nil
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
