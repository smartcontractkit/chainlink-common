package host

import (
	"testing"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
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

func makeRegressionTestModuleWithConfig(t *testing.T) *module {
	testName := strcase.ToSnake(t.Name()[len("TestRegression"):])
	return makeTestModuleByName(t, "./regression_tests", testName, nil, true)
}
