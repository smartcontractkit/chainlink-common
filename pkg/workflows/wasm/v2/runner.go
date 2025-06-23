package wasm

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

type Config any

type runnerInternals interface {
	args() []string
	sendResponse(response unsafe.Pointer, responseLen int32) int32
	versionV2()
	switchModes(mode int32)
}

func newRunner[C Config](parse func(configBytes []byte) (C, error), runnerInternals runnerInternals, runtimeInternals runtimeInternals) sdk.Runner[C] {
	runnerInternals.versionV2()
	drt := &sdkimpl.Runtime{RuntimeBase: newRuntime(runtimeInternals, sdkpb.Mode_DON)}
	return runnerWrapper[C]{baseRunner: getRunner(
		parse,
		&subscriber[C, sdk.Runtime]{
			sp:              drt,
			runnerInternals: runnerInternals,
		},
		&runner[C, sdk.Runtime]{
			sp:              drt,
			runtime:         drt,
			runnerInternals: runnerInternals,
			setRuntime: func(config []byte, maxResponseSize uint64) {
				drt.MaxResponseSize = maxResponseSize
			},
		}),
	}
}

type runner[C, T any] struct {
	runnerInternals
	trigger    *sdkpb.Trigger
	id         string
	runtime    T
	setRuntime func(config []byte, maxResponseSize uint64)
	config     C
	sp         sdk.SecretsProvider
}

var _ baseRunner[any, sdk.Runtime] = (*runner[any, sdk.Runtime])(nil)

func (r *runner[C, T]) cfg() C {
	return r.config
}

func (r *runner[C, T]) secretsProvider() sdk.SecretsProvider {
	return r.sp
}

func (r *runner[C, T]) run(wfs []sdk.ExecutionHandler[C, T]) {
	env := &sdk.Environment[C]{
		NodeEnvironment: sdk.NodeEnvironment[C]{
			Config:    r.config,
			LogWriter: &writer{},
			Logger:    slog.New(slog.NewTextHandler(&writer{}, nil)),
		},
		SecretsProvider: r.secretsProvider(),
	}
	for idx, handler := range wfs {
		if uint64(idx) == r.trigger.Id {
			response, err := handler.Callback()(env, r.runtime, r.trigger.Payload)
			execResponse := &pb.ExecutionResult{}
			if err == nil {
				wrapped, err := values.Wrap(response)
				if err != nil {
					execResponse.Result = &pb.ExecutionResult_Error{Error: err.Error()}
				} else {
					execResponse.Result = &pb.ExecutionResult_Value{Value: values.Proto(wrapped)}
				}
			} else {
				execResponse.Result = &pb.ExecutionResult_Error{Error: err.Error()}
			}
			marshalled, _ := proto.Marshal(execResponse)
			marshalledPtr, marshalledLen, _ := bufferToPointerLen(marshalled)
			r.sendResponse(marshalledPtr, marshalledLen)
		}
	}
}

type subscriber[C, T any] struct {
	runnerInternals
	id     string
	config C
	sp     sdk.SecretsProvider
}

var _ baseRunner[any, sdk.Runtime] = &subscriber[any, sdk.Runtime]{}

func (s *subscriber[C, T]) cfg() C {
	return s.config
}

func (s *subscriber[C, T]) secretsProvider() sdk.SecretsProvider {
	return s.sp
}

func (s *subscriber[C, T]) run(wfs []sdk.ExecutionHandler[C, T]) {
	subscriptions := make([]*sdkpb.TriggerSubscription, len(wfs))
	for i, handler := range wfs {
		subscriptions[i] = &sdkpb.TriggerSubscription{
			Id:      handler.CapabilityID(),
			Payload: handler.TriggerCfg(),
			Method:  handler.Method(),
		}
	}
	triggerSubscription := &sdkpb.TriggerSubscriptionRequest{Subscriptions: subscriptions}

	execResponse := &pb.ExecutionResult{
		Result: &pb.ExecutionResult_TriggerSubscriptions{TriggerSubscriptions: triggerSubscription},
	}

	configBytes, _ := proto.Marshal(execResponse)
	configPtr, configLen, _ := bufferToPointerLen(configBytes)

	result := s.sendResponse(configPtr, configLen)
	if result < 0 {
		exitErr(fmt.Sprintf("could not subscribe to triggers: %s", string(configBytes[:-result])))
	}
}

func getWorkflows[C any](config C, secretsProvider sdk.SecretsProvider, initFn func(env *sdk.Environment[C]) (sdk.Workflow[C], error)) sdk.Workflow[C] {
	wfs, err := initFn(&sdk.Environment[C]{
		NodeEnvironment: sdk.NodeEnvironment[C]{
			Config:    config,
			LogWriter: &writer{},
			Logger:    slog.New(slog.NewTextHandler(&writer{}, nil)),
		},
		SecretsProvider: secretsProvider,
	})
	if err != nil {
		exitErr(err.Error())
	}
	return wfs
}

func getRunner[C, T any](parse func(configBytes []byte) (C, error), subscribe *subscriber[C, T], run *runner[C, T]) baseRunner[C, T] {
	args := run.args()

	// We expect exactly 2 args, i.e. `wasm <blob>`,
	// where <blob> is a base64 encoded protobuf message.
	if len(args) != 2 {
		exitErr("invalid request: request must contain a payload")
	}

	request := args[1]
	if request == "" {
		exitErr("invalid request: request cannot be empty")
	}

	b, err := base64.StdEncoding.DecodeString(request)
	if err != nil {
		exitErr("invalid request: could not decode request into bytes")
	}

	execRequest := &pb.ExecuteRequest{}
	if err = proto.Unmarshal(b, execRequest); err != nil {
		exitErr("invalid request: could not unmarshal request into ExecuteRequest")
	}

	c, err := parse(execRequest.Config)
	if err != nil {
		exitErr(err.Error())
	}

	switch req := execRequest.Request.(type) {
	case *pb.ExecuteRequest_Subscribe:
		subscribe.config = c
		return subscribe
	case *pb.ExecuteRequest_Trigger:
		run.trigger = req.Trigger
		run.config = c
		run.setRuntime(execRequest.Config, execRequest.MaxResponseSize)
		return run
	}

	exitErr(fmt.Sprintf("invalid request: unknown request type %T", execRequest.Request))
	return nil
}

func exitErr(msg string) {
	_, _ = (&writer{}).Write([]byte(msg))
	os.Exit(1)
}

type baseRunner[C, T any] interface {
	secretsProvider() sdk.SecretsProvider
	cfg() C
	run([]sdk.ExecutionHandler[C, T])
}

type runnerWrapper[C any] struct {
	baseRunner[C, sdk.Runtime]
}

func (r runnerWrapper[C]) Run(initFn func(env *sdk.Environment[C]) (sdk.Workflow[C], error)) {
	wfs := getWorkflows(r.baseRunner.cfg(), r.baseRunner.secretsProvider(), initFn)
	r.baseRunner.run(wfs)
}
