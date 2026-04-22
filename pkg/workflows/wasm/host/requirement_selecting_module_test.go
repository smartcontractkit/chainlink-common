package host

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

type stubModuleV2 struct {
	startFn   func()
	closeFn   func()
	legacyFn  func() bool
	executeFn func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error)
}

func (s *stubModuleV2) Start()            { s.startFn() }
func (s *stubModuleV2) Close()            { s.closeFn() }
func (s *stubModuleV2) IsLegacyDAG() bool { return s.legacyFn() }
func (s *stubModuleV2) Execute(ctx context.Context, req *sdk.ExecuteRequest, h ExecutionHelper) (*sdk.ExecutionResult, error) {
	return s.executeFn(ctx, req, h)
}

func TestRequirementSelectingModule_Start(t *testing.T) {
	var started bool
	m0 := &stubModuleV2{startFn: func() { started = true }}
	m := NewRequirementSelectingModule([]ModuleAndHandler{{ModuleV2: m0}})
	m.Start()
	assert.True(t, started)
}

func TestRequirementSelectingModule_Close(t *testing.T) {
	t.Run("before execute closes first module", func(t *testing.T) {
		var closedIdx int
		m0 := &stubModuleV2{closeFn: func() { closedIdx = 0 }}
		m1 := &stubModuleV2{closeFn: func() { closedIdx = 1 }}
		m := NewRequirementSelectingModule([]ModuleAndHandler{
			{ModuleV2: m0},
			{ModuleV2: m1},
		})
		closedIdx = -1
		m.Close()
		assert.Equal(t, 0, closedIdx)
	})

	t.Run("after execute closes selected module", func(t *testing.T) {
		wantResult := &sdk.ExecutionResult{}
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		var closedIdx int

		m0 := &stubModuleV2{
			startFn: func() {},
			closeFn: func() { closedIdx = 0 },
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
			},
		}
		m1 := &stubModuleV2{
			closeFn: func() { closedIdx = 1 },
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return wantResult, nil
			},
		}

		m := NewRequirementSelectingModule([]ModuleAndHandler{
			{ModuleV2: m0},
			{ModuleV2: m1, RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }}},
		})

		_, err := m.Execute(t.Context(), &sdk.ExecuteRequest{}, nil)
		require.NoError(t, err)

		closedIdx = -1
		m.Close()
		assert.Equal(t, 1, closedIdx)
	})
}

func TestRequirementSelectingModule_IsLegacyDAG(t *testing.T) {
	t.Run("delegates", func(t *testing.T) {
		m0 := &stubModuleV2{legacyFn: func() bool { return true }}
		m := NewRequirementSelectingModule([]ModuleAndHandler{{ModuleV2: m0}})
		assert.True(t, m.IsLegacyDAG())
	})
}

func TestRequirementSelectingModule_Execute(t *testing.T) {
	t.Run("delegates when runOn already set", func(t *testing.T) {
		calls := 0
		wantResult := &sdk.ExecutionResult{}
		m0 := &stubModuleV2{
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				calls++
				return wantResult, nil
			},
		}

		m := NewRequirementSelectingModule([]ModuleAndHandler{{ModuleV2: m0}})

		_, err := m.Execute(t.Context(), &sdk.ExecuteRequest{}, nil)
		require.NoError(t, err)
		assert.Equal(t, 1, calls)

		got, err := m.Execute(t.Context(), &sdk.ExecuteRequest{}, nil)
		require.NoError(t, err)
		assert.Equal(t, wantResult, got)
		assert.Equal(t, 2, calls)
	})

	t.Run("first module succeeds sets runOn to zero", func(t *testing.T) {
		wantResult := &sdk.ExecutionResult{}
		numCalls := 0
		m0 := &stubModuleV2{
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				numCalls++
				return wantResult, nil
			},
		}

		m := NewRequirementSelectingModule([]ModuleAndHandler{{ModuleV2: m0}})

		got, err := m.Execute(t.Context(), &sdk.ExecuteRequest{}, nil)
		require.NoError(t, err)
		assert.Equal(t, 1, numCalls)
		assert.Equal(t, wantResult, got)
	})

	t.Run("non-RequirementsRerun error is propagated without additional executions", func(t *testing.T) {
		m0 := &stubModuleV2{
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, assert.AnError
			},
		}

		m1 := &stubModuleV2{executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
			assert.Fail(t, "second module should not be executed")
			return nil, nil
		}}

		m := NewRequirementSelectingModule([]ModuleAndHandler{{ModuleV2: m0}, {ModuleV2: m1}})

		_, err := m.Execute(t.Context(), &sdk.ExecuteRequest{}, nil)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("RequirementsRerun with matching handler not started", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		wantResult := &sdk.ExecutionResult{}
		var m1Started bool

		m0 := &stubModuleV2{
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
			},
		}
		m1 := &stubModuleV2{
			startFn: func() { m1Started = true },
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return wantResult, nil
			},
		}

		m := NewRequirementSelectingModule([]ModuleAndHandler{
			{ModuleV2: m0},
			{ModuleV2: m1, RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }}},
		})

		got, err := m.Execute(t.Context(), &sdk.ExecuteRequest{}, nil)
		require.NoError(t, err)
		assert.Equal(t, wantResult, got)
		assert.False(t, m1Started)
	})

	t.Run("RequirementsRerun with matching handler already started", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		wantResult := &sdk.ExecutionResult{}
		var m1Started bool

		m0 := &stubModuleV2{
			startFn: func() {},
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
			},
		}
		m1 := &stubModuleV2{
			startFn: func() { m1Started = true },
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return wantResult, nil
			},
		}

		m := NewRequirementSelectingModule([]ModuleAndHandler{
			{ModuleV2: m0},
			{ModuleV2: m1, RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }}},
		})

		m.Start()

		got, err := m.Execute(t.Context(), &sdk.ExecuteRequest{}, nil)
		require.NoError(t, err)
		assert.Equal(t, wantResult, got)
		assert.True(t, m1Started)
	})

	t.Run("RequirementsRerun with no matching handler returns error", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		m0 := &stubModuleV2{
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
			},
		}
		m1 := &stubModuleV2{}

		m := NewRequirementSelectingModule([]ModuleAndHandler{
			{ModuleV2: m0},
			{ModuleV2: m1, RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return false }}},
		})

		_, err := m.Execute(t.Context(), &sdk.ExecuteRequest{}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find a runner that can satisfy the requirements")
	})

	t.Run("RequirementsRerun skips non-matching selects later match", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		wantResult := &sdk.ExecutionResult{}

		m0 := &stubModuleV2{
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
			},
		}
		m1 := &stubModuleV2{}
		m2 := &stubModuleV2{
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return wantResult, nil
			},
		}

		m := NewRequirementSelectingModule([]ModuleAndHandler{
			{ModuleV2: m0},
			{ModuleV2: m1, RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return false }}},
			{ModuleV2: m2, RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }}},
		})

		got, err := m.Execute(t.Context(), &sdk.ExecuteRequest{}, nil)
		require.NoError(t, err)
		assert.Equal(t, wantResult, got)
	})
}
