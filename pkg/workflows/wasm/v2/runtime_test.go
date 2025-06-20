package wasm

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	basicactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction/basic_actionmock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
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

		runtime := &sdkimpl.Runtime{RuntimeBase: newTestRuntime(t, false, nil)}
		capability := &basicaction.BasicAction{}
		response, err := capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()
		require.NoError(t, err)
		assert.True(t, proto.Equal(anyOutput, response))
	})

	t.Run("call capability host error", func(t *testing.T) {
		_, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)

		runtime := &sdkimpl.Runtime{RuntimeBase: newTestRuntime(t, true, nil)}

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

		override := func() ([]byte, error) { return []byte("invalid"), nil }

		runtime := &sdkimpl.Runtime{RuntimeBase: newTestRuntime(t, false, override)}
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
		override := func() ([]byte, error) { return nil, anyErr }

		runtime := &sdkimpl.Runtime{RuntimeBase: newTestRuntime(t, false, override)}
		capability := &basicaction.BasicAction{}
		_, err = capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()

		require.ErrorContains(t, err, anyErr.Error())
	})
}

func Test_runtimeInternals_UsesSeeds(t *testing.T) {
	anyDonSeed := int64(123456789)
	anyNodeSeed := int64(987654321)
	helper := &runtimeHelper{runtimeInternals: &runtimeInternalsTestHook{
		donSeed:  anyDonSeed,
		nodeSeed: anyNodeSeed,
	}}
	assertRnd(t, helper, sdkpb.Mode_MODE_DON, anyDonSeed)
	assertRnd(t, helper, sdkpb.Mode_MODE_NODE, anyNodeSeed)
}

func assertRnd(t *testing.T, helper *runtimeHelper, mode sdkpb.Mode, seed int64) {
	rnd := rand.New(helper.GetSource(mode))
	buff := make([]byte, 1000)
	n, err := rnd.Read(buff)
	require.NoError(t, err)
	assert.Equal(t, len(buff), n)
	expectedBuf := make([]byte, 1000)
	n, err = rand.New(rand.NewSource(seed)).Read(expectedBuf)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBuf), n)
	assert.Equal(t, string(expectedBuf), string(buff))
}

func newTestRuntime(t *testing.T, callCapabilityErr bool, awaitResponseOverride func() ([]byte, error)) sdkimpl.RuntimeBase {
	internals := testRuntimeInternals(t)
	internals.callCapabilityErr = callCapabilityErr
	internals.awaitResponseOverride = awaitResponseOverride
	runtime := newRuntime(internals, sdkpb.Mode_MODE_DON)
	runtime.MaxResponseSize = sdk.DefaultMaxResponseSizeBytes
	return runtime
}
