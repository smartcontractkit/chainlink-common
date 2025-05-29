package testhelpers

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	basicactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction/basic_actionmock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	basictriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger/basic_triggermock"
)

func TestWorkflowTrigger() *basictrigger.Outputs {
	return &basictrigger.Outputs{CoolOutput: "Hi"}
}

func TestWorkflowTriggerConfig() *basictrigger.Config {
	return &basictrigger.Config{
		Name:   "name",
		Number: 100,
	}
}

func SetupExpectedCalls(t *testing.T) {
	triggerMock, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	triggerMock.Trigger = func(ctx context.Context, input *basictrigger.Config) (*basictrigger.Outputs, error) {
		return TestWorkflowTrigger(), nil
	}

	basicAction, err := basicactionmock.NewBasicActionCapability(t)
	require.NoError(t, err)

	firstCall := true
	callLock := &sync.Mutex{}
	basicAction.PerformAction = func(ctx context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
		callLock.Lock()
		defer callLock.Unlock()
		assert.NotEqual(t, firstCall, input.InputThing, "failed first call assertion")
		firstCall = false
		if input.InputThing {
			return &basicaction.Outputs{AdaptedThing: "true"}, nil
		} else {
			return &basicaction.Outputs{AdaptedThing: "false"}, nil
		}
	}
}

func TestWorkflowExpectedResult() string {
	return "Hifalsetrue"
}
