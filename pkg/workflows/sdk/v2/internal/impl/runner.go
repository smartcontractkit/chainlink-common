package impl

import (
	"io"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

func NewDonRunner() sdk.DonRunner {
	return &runner[sdk.DonRuntime]{runtime: drt, setRuntime: func(id string, config []byte, maxResponseSize uint64) {
		drt.execId = id
		drt.config = config
		drt.maxResponseSize = maxResponseSize
	}}
}

type runner[T any] struct {
	trigger      *pb.Trigger
	id           string
	runtime      T
	setRuntime   func(id string, config []byte, maxResponseSize uint64)
	config       []byte
	versionV2    func()
	sendResponse func(response unsafe.Pointer, responseLen int32) int32
	writer       io.Writer
}

var _ sdk.DonRunner = &runner[sdk.DonRuntime]{}
var _ sdk.NodeRunner = &runner[sdk.NodeRuntime]{}

func (d *runner[T]) Run(args *sdk.WorkflowArgs[T]) {
	// used to ensure that the export isn't optimized away
	d.versionV2()
	for _, handler := range args.Handlers {
		// TODO multiple subscriptions the the same trigger
		if handler.Id() == d.trigger.Id {
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
			runner.sendResponse(marshalledPtr, marshalledLen)
		}
	}
}

func (d *runner[T]) Config() []byte {
	return d.config
}

func (d *runner[T]) LogWriter() io.Writer {
	return runner.writer
}
