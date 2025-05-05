package wasm

import (
	"context"
	"errors"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	basicactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction/basic_actionmock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRuntimeBase_CallCapability(t *testing.T) {
	t.Run("call capability returns host provided id and can be awaited", func(t *testing.T) {
		c, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)
		anyOutput := &basicaction.Outputs{AdaptedThing: "foo"}
		c.PerformAction = func(ctx context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
			return anyOutput, nil
		}

		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		capability := &basicaction.BasicAction{}
		response, err := capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()
		require.NoError(t, err)
		assert.True(t, proto.Equal(anyOutput, response))
	})

	t.Run("call capability host error", func(t *testing.T) {
		_, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)

		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		callCapabilityErr = true

		capability := &basicaction.BasicAction{}
		_, err = capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()
		require.Error(t, err)
	})

	t.Run("awaitCapabilities unparsable response", func(t *testing.T) {
		a, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)
		a.PerformAction = func(ctx context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
			return &basicaction.Outputs{AdaptedThing: "foo"}, nil
		}

		overrideCapabilityResponseForTest(t, func() ([]byte, error) { return []byte("invalid"), nil })

		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		capability := &basicaction.BasicAction{}
		_, err = capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()

		require.Error(t, err)
	})

	t.Run("awaitCapabilities error response", func(t *testing.T) {
		a, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)
		a.PerformAction = func(ctx context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
			return &basicaction.Outputs{AdaptedThing: "foo"}, nil
		}

		anyErr := errors.New("not this time")
		overrideCapabilityResponseForTest(t, func() ([]byte, error) { return nil, anyErr })

		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		capability := &basicaction.BasicAction{}
		_, err = capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()

		require.ErrorContains(t, err, anyErr.Error())
	})
}

func TestRuntimeBase_LogWriter(t *testing.T) {
	runtime := newTestRuntime(t)
	assert.IsType(t, &writer{}, runtime.LogWriter())
}

func newTestRuntime(t *testing.T) sdkimpl.RuntimeBase {
	initRunnerAndRuntimeForTest(t, anyExecutionId)
	runtime := newRuntime()
	runtime.ExecId = anyExecutionId
	runtime.ConfigBytes = anyConfig
	runtime.MaxResponseSize = sdk.DefaultMaxResponseSizeBytes
	return runtime
}
