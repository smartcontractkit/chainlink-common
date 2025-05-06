package wasm

import (
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

type runnerInternals interface {
	args() []string
	sendResponse(response unsafe.Pointer, responseLen int32) int32
	versionV2()
}

func newDonRunner(runnerInternals runnerInternals, runtimeInternals runtimeInternals) sdk.DonRunner {
	drt := &sdkimpl.DonRuntime{RuntimeBase: newRuntime(runtimeInternals)}
	return getRunner(
		&subscriber[sdk.DonRuntime]{runnerInternals: runnerInternals},
		&runner[sdk.DonRuntime]{
			runtime:         drt,
			runnerInternals: runnerInternals,
			setRuntime: func(id string, config []byte, maxResponseSize uint64) {
				drt.ExecId = id
				drt.ConfigBytes = config
				drt.MaxResponseSize = maxResponseSize
			}})
}

func newNodeRunner(runnerInternals runnerInternals, runtimeInternals runtimeInternals) sdk.NodeRunner {
	nrt := &sdkimpl.NodeRuntime{RuntimeBase: newRuntime(runtimeInternals)}
	return getRunner(
		&subscriber[sdk.NodeRuntime]{runnerInternals: runnerInternals},
		&runner[sdk.NodeRuntime]{
			runnerInternals: runnerInternals,
			runtime:         nrt,
			setRuntime: func(id string, config []byte, maxResponseSize uint64) {
				nrt.ExecId = id
				nrt.ConfigBytes = config
				nrt.MaxResponseSize = maxResponseSize
			}})
}

type runner[T any] struct {
	runnerInternals
	trigger    *sdkpb.Trigger
	id         string
	runtime    T
	setRuntime func(id string, config []byte, maxResponseSize uint64)
	config     []byte
}

var _ sdk.DonRunner = &runner[sdk.DonRuntime]{}
var _ sdk.NodeRunner = &runner[sdk.NodeRuntime]{}

func (d *runner[T]) Run(args *sdk.WorkflowArgs[T]) {
	// used to ensure that the export isn't optimized away
	d.versionV2()
	for idx, handler := range args.Handlers {
		if uint64(idx) == d.trigger.Id {
			response, err := handler.Callback()(d.runtime, d.trigger.Payload)
			execResponse := &pb.ExecutionResult{Id: d.id}
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
			d.sendResponse(marshalledPtr, marshalledLen)
		}
	}
}

func (d *runner[T]) Config() []byte {
	return d.config
}

func (d *runner[T]) LogWriter() io.Writer {
	return &writer{}
}

type subscriber[T any] struct {
	runnerInternals
	id     string
	config []byte
}

var _ sdk.DonRunner = &subscriber[sdk.DonRuntime]{}
var _ sdk.NodeRunner = &subscriber[sdk.NodeRuntime]{}

func (d *subscriber[T]) Run(args *sdk.WorkflowArgs[T]) {
	subscriptions := make([]*sdkpb.TriggerSubscription, len(args.Handlers))
	for i, handler := range args.Handlers {
		subscriptions[i] = &sdkpb.TriggerSubscription{
			ExecId:  d.id,
			Id:      handler.Id(),
			Payload: handler.TriggerCfg(),
			Method:  handler.Method(),
		}
	}
	triggerSubscription := &sdkpb.TriggerSubscriptionRequest{Subscriptions: subscriptions}

	execResponse := &pb.ExecutionResult{
		Id:     d.id,
		Result: &pb.ExecutionResult_TriggerSubscriptions{TriggerSubscriptions: triggerSubscription},
	}

	configBytes, _ := proto.Marshal(execResponse)
	configPtr, configLen, _ := bufferToPointerLen(configBytes)

	result := d.sendResponse(configPtr, configLen)
	if result < 0 {
		panic(fmt.Sprintf("could not subscribe to triggers: %s", string(configBytes[:-result])))
	}
}

func (d *subscriber[T]) Config() []byte {
	return d.config
}

func (d *subscriber[T]) LogWriter() io.Writer {
	return &writer{}
}

type genericRunner[T any] interface {
	Run(args *sdk.WorkflowArgs[T])
	Config() []byte
	LogWriter() io.Writer
}

func getRunner[T any](subscribe *subscriber[T], run *runner[T]) genericRunner[T] {
	slog.SetDefault(slog.New(slog.NewTextHandler(&writer{}, nil)))

	args := run.args()

	// We expect exactly 2 args, i.e. `wasm <blob>`,
	// where <blob> is a base64 encoded protobuf message.
	if len(args) != 2 {
		panic("invalid request: request must contain a payload")
	}

	request := args[1]
	if request == "" {
		panic("invalid request: request cannot be empty")
	}

	b, err := base64.StdEncoding.DecodeString(request)
	if err != nil {
		panic("invalid request: could not decode request into bytes")
	}

	execRequest := &pb.ExecuteRequest{}
	if err = proto.Unmarshal(b, execRequest); err != nil {
		panic("invalid request: could not unmarshal request into ExecuteRequest")
	}

	switch req := execRequest.Request.(type) {
	case *pb.ExecuteRequest_Subscribe:
		subscribe.id = execRequest.Id
		subscribe.config = execRequest.Config
		return subscribe
	case *pb.ExecuteRequest_Trigger:
		run.trigger = req.Trigger
		run.id = execRequest.Id
		run.config = execRequest.Config
		run.setRuntime(execRequest.Id, execRequest.Config, execRequest.MaxResponseSize)
		return run
	}

	panic(fmt.Sprintf("invalid request: unknown request type %T", execRequest.Request))
}
