package host

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

type stubModule struct {
	startFn   func()
	closeFn   func()
	legacyFn  func() bool
	executeFn func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error)
}

func (s *stubModule) Start()            { s.startFn() }
func (s *stubModule) Close()            { s.closeFn() }
func (s *stubModule) IsLegacyDAG() bool { return s.legacyFn() }
func (s *stubModule) Execute(ctx context.Context, req *sdk.ExecuteRequest, h ExecutionHelper) (*sdk.ExecutionResult, error) {
	return s.executeFn(ctx, req, h)
}

type requirementEnforcingStub struct {
	*stubModule
	setRequirementsFn func(string, *sdk.Requirements)
}

func (s *requirementEnforcingStub) SetRequirements(executionID string, requirements *sdk.Requirements) {
	s.setRequirementsFn(executionID, requirements)
}

func noop()      {}
func noopClose() {}

func triggerRequest(id uint64) *sdk.ExecuteRequest {
	return &sdk.ExecuteRequest{
		Request: &sdk.ExecuteRequest_Trigger{
			Trigger: &sdk.Trigger{Id: id},
		},
	}
}

func subscribeRequest() *sdk.ExecuteRequest {
	return &sdk.ExecuteRequest{
		Request: &sdk.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}},
	}
}

func subscribeResult(subs ...*sdk.TriggerSubscription) *sdk.ExecutionResult {
	return &sdk.ExecutionResult{
		Result: &sdk.ExecutionResult_TriggerSubscriptions{
			TriggerSubscriptions: &sdk.TriggerSubscriptionRequest{
				Subscriptions: subs,
			},
		},
	}
}

func subWithReqs(reqs *sdk.Requirements) *sdk.TriggerSubscription {
	return &sdk.TriggerSubscription{Requirements: reqs}
}

func TestRequirementSelectingModule_Start(t *testing.T) {
	t.Run("starts only main module", func(t *testing.T) {
		var mainStarted, additionalStarted bool
		main := ModuleAndHandler{Module: &stubModule{startFn: func() { mainStarted = true }}}
		add := ModuleAndHandler{Module: &stubModule{startFn: func() { additionalStarted = true }}}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		assert.True(t, mainStarted)
		assert.False(t, additionalStarted)
	})
}

func TestRequirementSelectingModule_Close(t *testing.T) {
	t.Run("closes main and no additional when none started", func(t *testing.T) {
		var mainClosed, addClosed bool
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop, closeFn: func() { mainClosed = true },
		}}
		add := ModuleAndHandler{Module: &stubModule{
			startFn: noop, closeFn: func() { addClosed = true },
		}}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()
		m.Close()

		assert.True(t, mainClosed)
		assert.False(t, addClosed)
	})

	t.Run("closes main and all started additional modules", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}

		var mainClosed, add0Closed, add1Closed bool
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			closeFn: func() { mainClosed = true },
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add0 := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				closeFn: func() { add0Closed = true },
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}
		add1 := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				closeFn: func() { add1Closed = true },
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return false }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add0, add1})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		m.Close()

		assert.True(t, mainClosed, "main should be closed")
		assert.True(t, add0Closed, "started additional should be closed")
		assert.False(t, add1Closed, "never-started additional should not be closed")
	})
}

func TestRequirementSelectingModule_IsLegacyDAG(t *testing.T) {
	main := ModuleAndHandler{Module: &stubModule{legacyFn: func() bool { return true }}}
	m := NewRequirementSelectingModule(main, nil)
	assert.True(t, m.IsLegacyDAG())
}

func TestRequirementSelectingModule_Execute(t *testing.T) {
	t.Run("trigger with no cached entry goes to main", func(t *testing.T) {
		want := &sdk.ExecutionResult{}
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if req.GetTrigger() != nil {
					return want, nil
				}
				return subscribeResult(), nil
			},
		}}

		m := NewRequirementSelectingModule(main, nil)
		m.Start()

		got, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("main error on subscribe propagates", func(t *testing.T) {
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, assert.AnError
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					t.Fatal("additional module should not be called")
					return nil, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("subscribe with requirements routes trigger to additional", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}
		want := &sdk.ExecutionResult{}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return want, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		got, err := m.Execute(t.Context(), triggerRequest(0), nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("subscribe with unmatched requirements returns error", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module:              &stubModule{startFn: noop},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return false }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find a runner that can satisfy the requirements")
	})

	t.Run("subscribe skips non-matching and selects later additional", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}
		want := &sdk.ExecutionResult{}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add0 := ModuleAndHandler{
			Module:              &stubModule{startFn: noop},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return false }},
		}
		add1 := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return want, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add0, add1})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		got, err := m.Execute(t.Context(), triggerRequest(0), nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("additional module started lazily during subscribe", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}
		var addStartCount int32

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: func() { atomic.AddInt32(&addStartCount, 1) },
				closeFn: noopClose,
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()
		assert.Equal(t, int32(0), atomic.LoadInt32(&addStartCount))

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&addStartCount))

		// Second subscribe does not start additional again (sync.Once).
		_, err = m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&addStartCount))
	})

	t.Run("subscribe with no requirements returns main result", func(t *testing.T) {
		want := subscribeResult()

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return want, nil
			},
		}}

		m := NewRequirementSelectingModule(main, nil)
		m.Start()

		got, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("main module satisfying requirements keeps trigger on main", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}
		want := &sdk.ExecutionResult{}

		var mainTriggerCalls int32
		main := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
					if req.GetTrigger() != nil {
						atomic.AddInt32(&mainTriggerCalls, 1)
						return want, nil
					}
					return subscribeResult(subWithReqs(teeReqs)), nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					t.Fatal("additional module should not be called when main satisfies requirements")
					return nil, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		got, err := m.Execute(t.Context(), triggerRequest(0), nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Equal(t, int32(1), atomic.LoadInt32(&mainTriggerCalls), "trigger should run on main")
	})

	t.Run("cached trigger sets requirements before execute", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}
		want := &sdk.ExecutionResult{}
		executionID := "wf-exec-1"

		main := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
					return subscribeResult(subWithReqs(teeReqs)), nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return false }},
		}

		var calls []string
		var gotReqs *sdk.Requirements
		var gotExecutionID string
		enforcingAdd := &requirementEnforcingStub{
			stubModule: &stubModule{
				startFn: noop,
				closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					calls = append(calls, "execute")
					return want, nil
				},
			},
			setRequirementsFn: func(id string, requirements *sdk.Requirements) {
				calls = append(calls, "set")
				gotExecutionID = id
				gotReqs = requirements
			},
		}
		add := ModuleAndHandler{
			Module:              enforcingAdd,
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		helper := &MockExecutionHelper{}
		helper.On("GetWorkflowExecutionID").Return(executionID).Once()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		got, err := m.Execute(t.Context(), triggerRequest(0), helper)
		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Equal(t, []string{"set", "execute"}, calls)
		assert.Equal(t, executionID, gotExecutionID)
		assert.Same(t, teeReqs, gotReqs)
		helper.AssertExpectations(t)
	})
}

func TestRequirementSelectingModule_TriggerCache(t *testing.T) {
	t.Run("cached trigger skips main on subsequent calls", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}
		var mainTriggerCalls int32

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if req.GetTrigger() != nil {
					atomic.AddInt32(&mainTriggerCalls, 1)
				}
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return &sdk.ExecutionResult{}, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		_, err = m.Execute(t.Context(), triggerRequest(0), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(0), atomic.LoadInt32(&mainTriggerCalls), "cached trigger should skip main")

		_, err = m.Execute(t.Context(), triggerRequest(0), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(0), atomic.LoadInt32(&mainTriggerCalls), "cached trigger should skip main on repeat")
	})

	t.Run("trigger not in cache goes to main", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}
		var mainTriggerCalls int32

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if req.GetTrigger() != nil {
					atomic.AddInt32(&mainTriggerCalls, 1)
					return &sdk.ExecutionResult{}, nil
				}
				// subscription 0 has requirements; subscription 1 does not
				return subscribeResult(subWithReqs(teeReqs), subWithReqs(nil)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return &sdk.ExecutionResult{}, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		// trigger 1 has no requirements → goes to main
		_, err = m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&mainTriggerCalls))
	})

	t.Run("different triggers route to different modules", func(t *testing.T) {
		// subscription 0: TEE required → additional; subscription 1: no requirements → main
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{
			Item: &sdk.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdk.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdk.TeeTypeAndRegions{{Type: sdk.TeeType_TEE_TYPE_AWS_NITRO}},
			}},
		}}
		var mainTriggerCalls int32
		wantAdditional := &sdk.ExecutionResult{}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if req.GetTrigger() != nil {
					atomic.AddInt32(&mainTriggerCalls, 1)
					return &sdk.ExecutionResult{}, nil
				}
				return subscribeResult(subWithReqs(teeReqs), subWithReqs(nil)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop, closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return wantAdditional, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		// trigger 0 has TEE requirements → additional
		got, err := m.Execute(t.Context(), triggerRequest(0), nil)
		require.NoError(t, err)
		assert.Equal(t, wantAdditional, got)
		assert.Equal(t, int32(0), atomic.LoadInt32(&mainTriggerCalls))

		// trigger 1 has no requirements → main
		_, err = m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&mainTriggerCalls))
	})

	t.Run("no additional modules when subscribe has requirements returns error", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}

		m := NewRequirementSelectingModule(main, nil)
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find a runner")
	})
}
