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
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}

		var mainClosed, add0Closed, add1Closed bool
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			closeFn: func() { mainClosed = true },
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
			},
		}}
		add0 := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop,
				closeFn: func() { add0Closed = true },
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return &sdk.ExecutionResult{}, nil
				},
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

		_, err := m.Execute(t.Context(), triggerRequest(1), nil)
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
	t.Run("main succeeds — returns result directly", func(t *testing.T) {
		want := &sdk.ExecutionResult{}
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return want, nil
			},
		}}

		m := NewRequirementSelectingModule(main, nil)
		m.Start()

		got, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("main non-RequirementsRerun error propagates", func(t *testing.T) {
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

		_, err := m.Execute(t.Context(), triggerRequest(1), nil)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("RequirementsRerun routes to matching additional", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		want := &sdk.ExecutionResult{}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
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

		got, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("RequirementsRerun with no matching additional returns error", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
			},
		}}
		add := ModuleAndHandler{
			Module:              &stubModule{startFn: noop},
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return false }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		_, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find a runner that can satisfy the requirements")
	})

	t.Run("RequirementsRerun skips non-matching and selects later match", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		want := &sdk.ExecutionResult{}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
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

		got, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("additional module started lazily", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		var addStartCount int32

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
			},
		}}
		add := ModuleAndHandler{
			Module: &stubModule{
				startFn: func() { atomic.AddInt32(&addStartCount, 1) },
				closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return &sdk.ExecutionResult{}, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return true }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{add})
		m.Start()

		// First execution starts the additional module.
		_, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&addStartCount))

		// Second execution with a different trigger still goes through main,
		// but the additional module is not started again (sync.Once).
		_, err = m.Execute(t.Context(), triggerRequest(2), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&addStartCount))
	})

	t.Run("subscribe request goes through main directly", func(t *testing.T) {
		want := &sdk.ExecutionResult{}

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
}

func TestRequirementSelectingModule_TriggerCache(t *testing.T) {
	t.Run("cached trigger skips main on subsequent calls", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		var mainCalls int32

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				atomic.AddInt32(&mainCalls, 1)
				return nil, rerunErr
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

		// First call triggers main.
		_, err := m.Execute(t.Context(), triggerRequest(42), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&mainCalls))

		// Second call with same trigger ID skips main.
		_, err = m.Execute(t.Context(), triggerRequest(42), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&mainCalls))
	})

	t.Run("different trigger IDs are cached independently", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}
		var mainCalls int32

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				atomic.AddInt32(&mainCalls, 1)
				return nil, rerunErr
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

		// Trigger 1 goes through main.
		_, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&mainCalls))

		// Trigger 2 also goes through main (different ID, not cached).
		_, err = m.Execute(t.Context(), triggerRequest(2), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(2), atomic.LoadInt32(&mainCalls))

		// Both are now cached — neither goes through main.
		_, err = m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		_, err = m.Execute(t.Context(), triggerRequest(2), nil)
		require.NoError(t, err)
		assert.Equal(t, int32(2), atomic.LoadInt32(&mainCalls))
	})

	t.Run("different triggers can route to different additional modules", func(t *testing.T) {
		teeRerun := &RequirementsRerun{Tee: &sdk.Tee{
			Type: &sdk.Tee_TypeSelection{TypeSelection: &sdk.TeeTypeSelection{
				Types: []*sdk.TeeTypeAndRegions{{Type: sdk.TeeType_TEE_TYPE_AWS_NITRO}},
			}},
		}}
		noReqRerun := &RequirementsRerun{}

		callCount := 0
		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(_ context.Context, req *sdk.ExecuteRequest, _ ExecutionHelper) (*sdk.ExecutionResult, error) {
				callCount++
				if req.GetTrigger().Id == 1 {
					return nil, teeRerun
				}
				return nil, noReqRerun
			},
		}}

		var addNitroResult, addDefaultResult sdk.ExecutionResult
		addNitro := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop, closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return &addNitroResult, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(tee *sdk.Tee) bool {
				return tee.GetTypeSelection() != nil
			}},
		}
		addDefault := ModuleAndHandler{
			Module: &stubModule{
				startFn: noop, closeFn: noopClose,
				executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
					return &addDefaultResult, nil
				},
			},
			RequirementsHandler: RequirementsHandler{Tee: func(*sdk.Tee) bool { return false }},
		}

		m := NewRequirementSelectingModule(main, []ModuleAndHandler{addNitro, addDefault})
		m.Start()

		got, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, &addNitroResult, got)

		// Trigger 1 is now cached to addNitro; verify second call skips main.
		got, err = m.Execute(t.Context(), triggerRequest(1), nil)
		require.NoError(t, err)
		assert.Equal(t, &addNitroResult, got)
		assert.Equal(t, 1, callCount, "main should only be called once for trigger 1")
	})

	t.Run("RequirementsRerun returned too late is rejected", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				time.Sleep(11 * time.Second)
				return nil, rerunErr
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

		_, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rerun requirement specified too late")
	})

	t.Run("RequirementsRerun with no additional modules returns error", func(t *testing.T) {
		rerunErr := &RequirementsRerun{Tee: &sdk.Tee{}}

		main := ModuleAndHandler{Module: &stubModule{
			startFn: noop,
			executeFn: func(context.Context, *sdk.ExecuteRequest, ExecutionHelper) (*sdk.ExecutionResult, error) {
				return nil, rerunErr
			},
		}}

		m := NewRequirementSelectingModule(main, nil)
		m.Start()

		_, err := m.Execute(t.Context(), triggerRequest(1), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find a runner")
	})
}
