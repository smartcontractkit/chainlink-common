package host

import (
	"context"
	"errors"
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
	executeFn  func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error)
	startCount atomic.Int32
	closeCount atomic.Int32
	legacy     bool
}

func (s *stubModule) Start()            { s.startCount.Add(1) }
func (s *stubModule) Close()            { s.closeCount.Add(1) }
func (s *stubModule) IsLegacyDAG() bool { return s.legacy }
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

type restrictionAwareStub struct {
	*stubModule
	setRestrictionsFn func(string, *sdk.Restrictions)
}

func (s *restrictionAwareStub) SetRestrictions(executionID string, restrictions *sdk.Restrictions) {
	s.setRestrictionsFn(executionID, restrictions)
}

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
		main := &stubModule{}
		unused := &stubModule{}

		m := NewRequirementSelectingModule(
			ModuleAndHandler{Module: main},
			[]ModuleAndHandler{{Module: unused}},
		)
		m.Start()

		assert.Equal(t, int32(1), main.startCount.Load())
		assert.Equal(t, int32(0), unused.startCount.Load())
	})
}

func TestRequirementSelectingModule_Close(t *testing.T) {
	t.Run("closes main and no additional when none started", func(t *testing.T) {
		main := &stubModule{}
		unused := &stubModule{}

		m := NewRequirementSelectingModule(
			ModuleAndHandler{Module: main},
			[]ModuleAndHandler{{Module: unused}},
		)
		m.Start()
		m.Close()

		assert.Equal(t, int32(1), main.closeCount.Load())
		assert.Equal(t, int32(0), unused.closeCount.Load())
	})

	t.Run("closes main and all started additional modules", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}

		main := &stubModule{
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}
		requirementsSatisfier := &stubModule{}
		nonMatcher := &stubModule{}

		m := NewRequirementSelectingModule(
			ModuleAndHandler{Module: main},
			[]ModuleAndHandler{
				{
					Module:              requirementsSatisfier,
					RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
				},
				{
					Module:              nonMatcher,
					RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return false }},
				},
			},
		)
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		m.Close()

		assert.Equal(t, int32(1), main.closeCount.Load(), "main should be closed")
		assert.Equal(t, int32(1), requirementsSatisfier.closeCount.Load(), "started additional should be closed")
		assert.Equal(t, int32(0), nonMatcher.closeCount.Load(), "never-started additional should not be closed")
	})
}

func TestRequirementSelectingModule_IsLegacyDAG(t *testing.T) {
	main := &stubModule{legacy: true}
	m := NewRequirementSelectingModule(ModuleAndHandler{Module: main}, nil)
	assert.True(t, m.IsLegacyDAG())
}

func TestRequirementSelectingModule_Execute(t *testing.T) {
	t.Run("trigger with no cached entry errors", func(t *testing.T) {
		main := ModuleAndHandler{Module: &stubModule{
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				assert.Fail(t, "main should not be called for trigger when no subscriptions")
				return nil, errors.New("unexpected callback")
			},
		}}

		m := NewRequirementSelectingModule(main, nil)
		m.Start()

		_, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.ErrorContains(t, err, "cannot trigger before gathering subscriptions")
	})

	t.Run("main error on subscribe propagates", func(t *testing.T) {
		main := ModuleAndHandler{Module: &stubModule{
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, assert.AnError
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
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
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
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
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module:              &stubModule{},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return false }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrRunnerUnavailable)
	})

	t.Run("subscribe skips non-matching and selects later additional", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}
		want := &sdk.ExecutionResult{}

		main := ModuleAndHandler{Module: &stubModule{
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add0 := ModuleAndHandler{
			Module:              &stubModule{},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return false }},
		}
		add1 := ModuleAndHandler{
			Module: &stubModule{
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

		main := ModuleAndHandler{Module: &stubModule{
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		requirementsSatisfier := &stubModule{}
		add := ModuleAndHandler{
			Module:              requirementsSatisfier,
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()
		assert.Equal(t, int32(0), requirementsSatisfier.startCount.Load())

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), requirementsSatisfier.startCount.Load())

		// Second subscribe does not start additional again (sync.Once).
		_, err = m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), requirementsSatisfier.startCount.Load())
	})

	t.Run("subscribe with no requirements returns main result", func(t *testing.T) {
		want := subscribeResult()

		main := ModuleAndHandler{Module: &stubModule{
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

		helper := &stubExecutionHelper{executionID: executionID}

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		got, err := m.Execute(t.Context(), triggerRequest(0), helper)
		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Equal(t, []string{"set", "execute"}, calls)
		assert.Equal(t, executionID, gotExecutionID)
		assert.Same(t, teeReqs, gotReqs)
	})
}

func TestRequirementSelectingModule_TriggerCache(t *testing.T) {
	t.Run("cached trigger skips main on subsequent calls", func(t *testing.T) {
		teeReqs := &sdk.Requirements{Tee: &sdk.Tee{}}
		var mainTriggerCalls int32

		main := ModuleAndHandler{Module: &stubModule{
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if req.GetTrigger() != nil {
					atomic.AddInt32(&mainTriggerCalls, 1)
				}
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
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
			executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}

		m := NewRequirementSelectingModule(main, nil)
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find a runner")
		assert.NotErrorIs(t, err, ErrRunnerUnavailable)
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
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if _, ok := req.Request.(*sdk.ExecuteRequest_PreHook); ok {
					return restrictionsResult(restrictions), nil
				}
				return subscribeResult(subWithReqsAndPreHook(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, h ExecutionHelper) (*sdk.ExecutionResult, error) {
					helperSeenByAdditional = h
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

		_, isRestricted := helperSeenByAdditional.(*executionRestrictions)
		assert.True(t, isRestricted, "additional module should receive a restricted helper")
	})

	t.Run("pre-hook error result is returned directly without running the trigger", func(t *testing.T) {
		errResult := &sdk.ExecutionResult{
			Result: &sdk.ExecutionResult_Error{Error: "denied by pre-hook"},
		}

		var helperSeenByAdditional ExecutionHelper
		main := ModuleAndHandler{Module: &stubModule{
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if _, ok := req.Request.(*sdk.ExecuteRequest_PreHook); ok {
					return errResult, nil
				}
				return subscribeResult(subWithReqsAndPreHook(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, h ExecutionHelper) (*sdk.ExecutionResult, error) {
					helperSeenByAdditional = h
					t.Fatal("additional module should not be called when pre-hook returns an error result")
					return nil, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		got, err := m.Execute(t.Context(), triggerRequest(0), &stubExecutionHelper{})
		require.NoError(t, err)
		assert.Same(t, errResult, got, "pre-hook error result should be returned unchanged")
		assert.Nil(t, helperSeenByAdditional, "additional module must not be invoked")
	})

	t.Run("pre-hook error propagates", func(t *testing.T) {
		main := ModuleAndHandler{Module: &stubModule{
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if _, ok := req.Request.(*sdk.ExecuteRequest_PreHook); ok {
					return nil, assert.AnError
				}
				return subscribeResult(subWithReqsAndPreHook(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					t.Fatal("additional module should not be called when pre-hook fails")
					return nil, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
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
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if req.GetTrigger() != nil {
					t.Fatal("main should not be called for trigger when cached in additional")
				}
				return subscribeResult(subWithReqs(teeReqs)), nil
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				executeFn: func(_ context.Context, _ *sdk.ExecuteRequest, h ExecutionHelper) (*sdk.ExecutionResult, error) {
					helperSeenByAdditional = h
					return &sdk.ExecutionResult{}, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		_, err = m.Execute(t.Context(), triggerRequest(0), inner)
		require.NoError(t, err)

		assert.Same(t, inner, helperSeenByAdditional, "without pre-hook, original helper should be passed unchanged")
	})

	t.Run("pre-hook restrictions are forwarded to RestrictionAwareModule", func(t *testing.T) {
		restrictions := &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{MaxTotalCalls: 3},
		}
		executionID := "wf-exec-restricted"

		main := ModuleAndHandler{Module: &stubModule{
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if _, ok := req.Request.(*sdk.ExecuteRequest_PreHook); ok {
					return restrictionsResult(restrictions), nil
				}
				return subscribeResult(subWithReqsAndPreHook(teeReqs)), nil
			},
		}}

		var calls []string
		var gotExecutionID string
		var gotRestrictions *sdk.Restrictions
		awareAdd := &restrictionAwareStub{
			stubModule: &stubModule{
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					calls = append(calls, "execute")
					return &sdk.ExecutionResult{}, nil
				},
			},
			setRestrictionsFn: func(id string, r *sdk.Restrictions) {
				calls = append(calls, "setRestrictions")
				gotExecutionID = id
				gotRestrictions = r
			},
		}
		add := ModuleAndHandler{
			Module:              awareAdd,
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		helper := &stubExecutionHelper{executionID: executionID}

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		_, err = m.Execute(t.Context(), triggerRequest(0), helper)
		require.NoError(t, err)

		assert.Equal(t, []string{"setRestrictions", "execute"}, calls)
		assert.Equal(t, executionID, gotExecutionID)
		assert.Same(t, restrictions, gotRestrictions)
	})

	t.Run("pre-hook restrictions are forwarded to module implementing both Restriction- and Requirement-aware interfaces", func(t *testing.T) {
		restrictions := &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{MaxTotalCalls: 5},
		}
		executionID := "wf-exec-both"

		main := ModuleAndHandler{Module: &stubModule{
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				if _, ok := req.Request.(*sdk.ExecuteRequest_PreHook); ok {
					return restrictionsResult(restrictions), nil
				}
				return subscribeResult(subWithReqsAndPreHook(teeReqs)), nil
			},
		}}

		var calls []string
		bothAware := &requirementAndRestrictionAwareStub{
			restrictionAwareStub: &restrictionAwareStub{
				stubModule: &stubModule{
					executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
						calls = append(calls, "execute")
						return &sdk.ExecutionResult{}, nil
					},
				},
				setRestrictionsFn: func(string, *sdk.Restrictions) { calls = append(calls, "setRestrictions") },
			},
			setRequirementsFn: func(string, *sdk.Requirements) { calls = append(calls, "setRequirements") },
		}
		add := ModuleAndHandler{
			Module:              bothAware,
			RequirementsHandler: RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		helper := &stubExecutionHelper{executionID: executionID}

		_, err := m.Execute(t.Context(), subscribeRequest(), nil)
		require.NoError(t, err)

		_, err = m.Execute(t.Context(), triggerRequest(0), helper)
		require.NoError(t, err)

		// Restrictions must be set before requirements, both before execute.
		assert.Equal(t, []string{"setRestrictions", "setRequirements", "execute"}, calls)
	})
}

// requirementAndRestrictionAwareStub implements both RestrictionAwareModule and RequirementEnforcingModule.
type requirementAndRestrictionAwareStub struct {
	*restrictionAwareStub
	setRequirementsFn func(string, *sdk.Requirements)
}

func (s *requirementAndRestrictionAwareStub) SetRequirements(executionID string, requirements *sdk.Requirements) {
	s.setRequirementsFn(executionID, requirements)
}

// stubExecutionHelper is a minimal ExecutionHelper implementation for testing.
type stubExecutionHelper struct{ executionID string }

func (s *stubExecutionHelper) CallCapability(context.Context, *sdk.CapabilityRequest) (*sdk.CapabilityResponse, error) {
	return nil, nil
}
func (s *stubExecutionHelper) GetSecrets(context.Context, *sdk.GetSecretsRequest) ([]*sdk.SecretResponse, error) {
	return nil, nil
}
func (s *stubExecutionHelper) GetWorkflowExecutionID() string { return s.executionID }
func (s *stubExecutionHelper) GetNodeTime() time.Time         { return time.Time{} }
func (s *stubExecutionHelper) GetDONTime() (time.Time, error) { return time.Time{}, nil }
func (s *stubExecutionHelper) EmitUserLog(string) error       { return nil }
func (s *stubExecutionHelper) EmitUserMetric(context.Context, *wfpb.WorkflowUserMetric) error {
	return nil
}
