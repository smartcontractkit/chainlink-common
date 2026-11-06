package host

import (
	"testing"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	whost "github.com/smartcontractkit/chainlink-common/pkg/workflows/host"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestRegressionInt32OverflowInWasmHostMemoryBoundsCheck(t *testing.T) {
	t.Parallel()
	mockExecutionHelper := mocks.NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id")
	// Some languages call time during initiation of the executable before the main is called.
	// This would be in unknown mode, which would call Node mode by default.
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()
	trigger := &basictrigger.Outputs{CoolOutput: anyTestTriggerValue}
	executeRequest := triggerExecuteRequest(t, 0, trigger)
	m := makeRegressionTestModuleWithConfig(t)

	// Bug was a panic here
	// we don't need to define the behaviour that happens when you don't use our SDK and avoid a return value
	// therefore, it's ok to not care if we have an empty result or an error.
	_, _ = m.Execute(t.Context(), executeRequest, mockExecutionHelper)
}

func TestRegressionMemoryExportInWasmModule(t *testing.T) {
	t.Parallel()

	// "missing" exports no memory at all; "wrong_type" exports a global named
	// "memory", which the export lookup finds but can't use as memory.
	for _, name := range []string{"missing", "wrong_type"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m, err := buildRegressionTestModuleWithConfig(t)
			if err != nil {
				// The host refused to admit the module. Nothing left to execute.
				t.Logf("module rejected at admission: %s", err)
				return
			}

			mockExecutionHelper := mocks.NewMockExecutionHelper(t)
			mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id").Maybe()
			mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
				return time.Now()
			}).Maybe()

			subscribe := &sdk.ExecuteRequest{Request: &sdk.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}}}

			_, err = m.Execute(t.Context(), subscribe, mockExecutionHelper)
			require.Error(t, err)
			require.NotContains(t, err.Error(), "invalid memory address or nil pointer dereference")
		})
	}
}

func TestRegressionPreHookNilResult(t *testing.T) {
	t.Parallel()
	m := makeRegressionTestModuleWithConfig(t)

	mockExecutionHelper := mocks.NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id").Maybe()
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()

	sm := whost.NewRequirementSelectingModule(whost.ModuleAndHandler{Module: m}, nil)

	subscribe := &sdk.ExecuteRequest{Request: &sdk.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}}}
	_, err := sm.Execute(t.Context(), subscribe, mockExecutionHelper)
	require.NoError(t, err)

	trigger := triggerExecuteRequest(t, 0, &basictrigger.Outputs{CoolOutput: anyTestTriggerValue})
	_, err = sm.Execute(t.Context(), trigger, mockExecutionHelper)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pre-hook")
}

func TestRegressionNoResponseIsError(t *testing.T) {
	t.Parallel()
	m := makeTestModuleByName(t, "./regression_tests", "pre_hook_nil_result", nil, true)

	mockExecutionHelper := mocks.NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("id").Maybe()
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()

	prehook := &sdk.ExecuteRequest{Request: &sdk.ExecuteRequest_PreHook{PreHook: &sdk.Trigger{}}}
	_, err := m.Execute(t.Context(), prehook, mockExecutionHelper)
	require.ErrorContains(t, err, "could not find response for execution")
}

func makeRegressionTestModuleWithConfig(t *testing.T) *module {
	m, err := buildRegressionTestModuleWithConfig(t)
	require.NoError(t, err)
	return m
}

func buildRegressionTestModuleWithConfig(t *testing.T) (*module, error) {
	testName := strcase.ToSnake(t.Name()[len("TestRegression"):])
	return buildTestModuleByName(t, "./regression_tests", testName, nil, true)
}
