package wasm

import (
	"encoding/base64"
	"fmt"
	"os"
	"unsafe"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

//go:wasmimport env change_mode
func changeMode(mode int32)

//go:wasmimport env subscribe_to_trigger
func subscribeToTrigger(reqptr unsafe.Pointer, reqptrlen int32, configptr unsafe.Pointer, configLen int32) int32

//go:wasmimport env sendResponse
func sendResponse(respptr unsafe.Pointer, respptrlen int32) (errno int32)

func NewDonRunner() sdk.DonRunner {
	return getRunner(pb.Mode_DON, &subscriber[sdk.DonRuntime]{}, &runner[sdk.DonRuntime]{})
}

func NewNodeRunner() sdk.NodeRunner {
	return getRunner(pb.Mode_NODE, &subscriber[sdk.NodeRuntime]{}, &runner[sdk.NodeRuntime]{})
}

type runner[T any] struct {
	trigger *pb.Trigger
	id      string
	runtime T
}

var _ sdk.DonRunner = &runner[sdk.DonRuntime]{}
var _ sdk.NodeRunner = &runner[sdk.NodeRuntime]{}

// TODO callbacks to setup a trigger...
// TODO can't subscribe to a trigger more than once and differentiate the return value.

func (d *runner[T]) SubscribeToTrigger(id string, _ *anypb.Any, handler func(runtime T, triggerOutputs *anypb.Any) (any, error)) error {
	if id == d.trigger.Id {
		response, err := handler(d.runtime, d.trigger.Payload)
		execResponse := &pb.ExecutionResult{Id: d.id}
		if err == nil {
			wrapped, err := values.Wrap(response)
			if err != nil {
				tmp := err.Error()
				execResponse.Error = &tmp
			} else {
				execResponse.Value = values.Proto(wrapped)
			}
		} else {
			tmp := err.Error()
			execResponse.Error = &tmp
		}
		marshalled, _ := proto.Marshal(execResponse)
		sendResponse(unsafe.Pointer(&marshalled[0]), int32(len(marshalled)))
	}

	return nil
}

type subscriber[T any] struct{}

var _ sdk.DonRunner = &subscriber[sdk.DonRuntime]{}
var _ sdk.NodeRunner = &subscriber[sdk.NodeRuntime]{}

func (s *subscriber[T]) SubscribeToTrigger(id string, triggerCfg *anypb.Any, _ func(runtime T, triggerOutputs *anypb.Any) (any, error)) error {
	idBytes := []byte(id)
	configBytes, err := proto.Marshal(triggerCfg)
	if err != nil {
		return err
	}

	result := subscribeToTrigger(unsafe.Pointer(&idBytes[0]), int32(len(idBytes)), unsafe.Pointer(&configBytes[0]), int32(len(configBytes)))
	if result < 0 {
		return fmt.Errorf("could not subscribe to trigger: %s", id)
	}

	return nil
}

type genericRunner[T any] interface {
	SubscribeToTrigger(id string, triggerCfg *anypb.Any, handler func(runtime T, triggerOutputs *anypb.Any) (any, error)) error
}

func getRunner[T any](mode pb.Mode, subscribe *subscriber[T], run *runner[T]) genericRunner[T] {
	changeMode(int32(mode))
	// We expect exactly 2 args, i.e. `wasm <blob>`,
	// where <blob> is a base64 encoded protobuf message.
	if len(os.Args) != 2 {
		panic("invalid request: request must contain a payload")
	}

	request := os.Args[1]
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
		return subscribe
	case *pb.ExecuteRequest_Trigger:
		run.trigger = req.Trigger
		run.id = execRequest.Id
		return run
	}

	panic(fmt.Sprintf("invalid request: unknown request type %T", execRequest.Request))
}
