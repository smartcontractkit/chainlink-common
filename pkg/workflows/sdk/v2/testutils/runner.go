package testutils

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils/registry"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type runner[C any] struct {
	baseRunner[C, sdk.Runtime]
}

func (r *runner[C]) Run(initFn func(env *sdk.Environment[C]) (sdk.Workflow[C], error)) {
	wfc := &sdk.Environment[C]{
		NodeEnvironment: sdk.NodeEnvironment[C]{
			Config:    r.config,
			LogWriter: r.writer,
			Logger:    slog.New(slog.NewTextHandler(r.writer, nil)),
		},
		SecretsProvider: &sdkimpl.Runtime{RuntimeBase: *r.baseRunner.base},
		ReportGenerator: &sdkimpl.Runtime{RuntimeBase: *r.baseRunner.base},
	}
	wfs, err := initFn(wfc)
	if err != nil {
		r.err = err
		return
	}

	r.baseRunner.run(wfs)
}

type baseRunner[C, T any] struct {
	tb             testing.TB
	ran            bool
	config         C
	result         any
	err            error
	workflowId     string
	registry       *registry.Registry
	strictTriggers bool
	runtime        T
	writer         *testWriter
	base           *sdkimpl.RuntimeBase
	source         rand.Source
	secrets        map[string]string
}

func (r *runner[T]) SetRandSource(source rand.Source) {
	r.source = source
}

type TestRunner interface {
	Result() (bool, any, error)

	// SetStrictTriggers causes the workflow to fail if a trigger isn'tb in the registry
	// this is useful for testing the workflow registrations.
	SetStrictTriggers(strict bool)

	// SetMaxResponseSizeBytes sets the maximum response size for the runtime.
	// Do not change unless you are working with a non-standard configuration.
	SetMaxResponseSizeBytes(maxResponseSizebytes uint64)

	Logs() []string

	SetRandSource(source rand.Source)

	SetSecret(id, namespace, value string) error
}

type Runner[C any] interface {
	sdk.Runner[C]
	TestRunner
}

func NewRunner[C any](tb testing.TB, config C) Runner[C] {
	drt := &sdkimpl.Runtime{}
	secrets := make(map[string]string)
	r := newBaseRunner(tb, config, drt, &drt.RuntimeBase, secrets)
	drt.RuntimeBase = newRuntime(tb, func() rand.Source { return r.source }, secrets)
	return &runner[C]{baseRunner: newBaseRunner[C, sdk.Runtime](tb, config, drt, &drt.RuntimeBase, secrets)}
}

func newBaseRunner[C, T any](tb testing.TB, config C, t T, base *sdkimpl.RuntimeBase, secrets map[string]string) baseRunner[C, T] {
	r := baseRunner[C, T]{
		tb:         tb,
		config:     config,
		workflowId: uuid.NewString(),
		registry:   registry.GetRegistry(tb),
		runtime:    t,
		writer:     &testWriter{},
		base:       base,
		source:     rand.NewSource(1),
		secrets:    secrets,
	}

	return r
}

func (r *baseRunner[C, T]) SetSecret(namespace, id, value string) error {
	if strings.Contains(namespace, "/") || strings.Contains(id, "/") {
		return fmt.Errorf("namespace and id cannot contain '/'")
	}

	key := fmt.Sprintf("%s/%s", namespace, id)
	r.secrets[key] = value
	return nil
}

func (r *baseRunner[C, T]) SetStrictTriggers(strict bool) {
	r.strictTriggers = strict
}

func (r *baseRunner[C, T]) SetMaxResponseSizeBytes(maxResponseSizeBytes uint64) {
	r.base.MaxResponseSize = maxResponseSizeBytes
}

func (r *baseRunner[C, T]) Logs() []string {
	logs := make([]string, len(r.writer.logs))
	for i, log := range r.writer.logs {
		logs[i] = string(log)
	}
	return logs
}

func (r *baseRunner[C, T]) run(workflows []sdk.ExecutionHandler[C, T]) {
	for _, handler := range workflows {
		trigger, err := r.registry.GetCapability(handler.CapabilityID())
		if err != nil {
			if r.strictTriggers {
				r.err = err
			}
			return
		}

		request := &pb.TriggerSubscription{
			Id:      uuid.NewString(),
			Payload: handler.TriggerCfg(),
			Method:  handler.Method(),
		}

		response, err := trigger.InvokeTrigger(r.tb.Context(), request)

		var nostub registry.ErrNoTriggerStub
		if err != nil && (r.strictTriggers || !errors.As(err, &nostub)) {
			r.err = err
			return
		}

		// trigger did not fire
		if response == nil {
			return
		}

		if r.ran {
			r.err = TooManyTriggers{}
			return
		}

		r.ran = true
		env := &sdk.Environment[C]{
			NodeEnvironment: sdk.NodeEnvironment[C]{
				Config:    r.config,
				LogWriter: r.writer,
				Logger:    slog.New(slog.NewTextHandler(r.writer, nil)),
			},
			SecretsProvider: &sdkimpl.Runtime{RuntimeBase: *r.base},
			ReportGenerator: &sdkimpl.Runtime{RuntimeBase: *r.base},
		}
		result, err := handler.Callback()(env, r.runtime, response.Payload)
		// If an error occurred during the callback (eg. via secrets fetching)
		// we don't want to override it with the result of the callback.
		if r.err != nil {
			return
		}
		r.result, r.err = result, err

		_, err = values.Wrap(r.result)
		if err != nil {
			r.result = nil
			r.err = err
			return
		}
	}
}

func (r *baseRunner[C, T]) Result() (bool, any, error) {
	return r.ran, r.result, r.err
}

var _ sdk.Runner[any] = &runner[any]{}

type TooManyTriggers struct{}

func (e TooManyTriggers) Error() string {
	return "too many triggers fired during execution"
}
