package host

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	wfpb "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
		}
		add1 := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				closeFn: func() { add1Closed = true },
			},
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return false }},
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return false }},
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return false }},
		}
		add1 := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return want, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
		}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					t.Fatal("additional module should not be called when main satisfies requirements")
					return nil, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
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
			Type: &sdk.Tee_TypeSelection{TypeSelection: &sdk.TeeTypeSelection{
				Types: []*sdk.TeeTypeAndRegions{{Type: sdk.TeeType_TEE_TYPE_AWS_NITRO}},
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
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
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

func subWithReqsAndPreHook(reqs *sdk.Requirements) *sdk.TriggerSubscription {
	return &sdk.TriggerSubscription{Requirements: reqs, PreHook: true}
}

func restrictionsResult(r *sdk.Restrictions) *sdk.ExecutionResult {
	return &sdk.ExecutionResult{
		Result: &sdk.ExecutionResult_Restrictions{Restrictions: r},
	}
}

func TestRequirementSelectingModule_PreHook(t *testing.T) {
	teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}

	t.Run("pre-hook runs in main, trigger runs in additional with restricted helper", func(t *testing.T) {
		restrictions := &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 1,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_OPEN,
			},
		}

		var helperSeenByAdditional ExecutionHelper
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if _, ok := req.Request.(*sdk.ExecuteRequest_PreHook); ok {
					return restrictionsResult(restrictions), nil
				}
				return subscribeResult(subWithReqsAndPreHook(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop, closeFn: noopClose,
				executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, h ExecutionHelper) (*sdk.ExecutionResult, error) {
					helperSeenByAdditional = h
					return &sdk.ExecutionResult{}, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		_, err = m.Execute(t.Context(), triggerRequest(0), nil)
		require.NoError(t, err)

		_, isRestricted := helperSeenByAdditional.(*executionRestrictions)
		assert.True(t, isRestricted, "additional module should receive a restricted helper")
	})

	t.Run("pre-hook error propagates", func(t *testing.T) {
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if _, ok := req.Request.(*sdk.ExecuteRequest_PreHook); ok {
					return nil, assert.AnError
				}
				return subscribeResult(subWithReqsAndPreHook(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop, closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					t.Fatal("additional module should not be called when pre-hook fails")
					return nil, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		_, err = m.Execute(t.Context(), triggerRequest(0), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pre-hook execution failed")
	})

	t.Run("pre-hook on main-routed trigger applies restrictions to main", func(t *testing.T) {
		restrictions := &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{MaxTotalCalls: 0},
		}
		var helperSeenByMain ExecutionHelper
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, h ExecutionHelper) (*sdk.ExecutionResult, error) {
				if _, ok := req.Request.(*sdk.ExecuteRequest_PreHook); ok {
					return restrictionsResult(restrictions), nil
				}
				if req.GetTrigger() != nil {
					helperSeenByMain = h
					return &sdk.ExecutionResult{}, nil
				}
				// Subscribe: no requirements, PreHook=true
				return subscribeResult(&sdk.TriggerSubscription{PreHook: true}), nil
			},
		}}

		m := NewRequirementSelectingModule(main, nil)
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		_, err = m.Execute(t.Context(), triggerRequest(0), nil)
		require.NoError(t, err)

		_, isRestricted := helperSeenByMain.(*executionRestrictions)
		assert.True(t, isRestricted, "main should receive a restricted helper when pre-hook is set")
	})

	t.Run("no pre-hook passes original helper to additional", func(t *testing.T) {
		var helperSeenByAdditional ExecutionHelper
		inner := &stubExecutionHelper{}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if req.GetTrigger() != nil {
					t.Fatal("main should not be called for trigger when cached in additional")
				}
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop, closeFn: noopClose,
				executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, h ExecutionHelper) (*sdk.ExecutionResult, error) {
					helperSeenByAdditional = h
					return &sdk.ExecutionResult{}, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		_, err = m.Execute(t.Context(), triggerRequest(0), inner)
		require.NoError(t, err)

		assert.Same(t, inner, helperSeenByAdditional, "without pre-hook, original helper should be passed unchanged")
	})
}

// stubExecutionHelper is a minimal ExecutionHelper implementation for testing.
type stubExecutionHelper struct{}

func (s *stubExecutionHelper) CallCapability(context.Context, *sdk.CapabilityRequest) (*sdk.CapabilityResponse, error) {
	return nil, nil
}
func (s *stubExecutionHelper) GetSecrets(context.Context, *sdk.GetSecretsRequest) ([]*sdk.SecretResponse, error) {
	return nil, nil
}
func (s *stubExecutionHelper) GetWorkflowExecutionID() string { return "" }
func (s *stubExecutionHelper) GetNodeTime() time.Time         { return time.Time{} }
func (s *stubExecutionHelper) GetDONTime() (time.Time, error) { return time.Time{}, nil }
func (s *stubExecutionHelper) EmitUserLog(string) error       { return nil }
func (s *stubExecutionHelper) EmitUserMetric(context.Context, *wfpb.WorkflowUserMetric) error {
	return nil
}
