package wasm

import (
	"os"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

//go:wasmimport env sendResponse
func sendResponse(respptr unsafe.Pointer, respptrlen int32) (errno int32)

//go:wasmimport env log
func log(respptr unsafe.Pointer, respptrlen int32)

//go:wasmimport env fetch
func fetch(respptr unsafe.Pointer, resplenptr unsafe.Pointer, reqptr unsafe.Pointer, reqptrlen int32) int32

//go:wasmimport env emit
func emit(respptr unsafe.Pointer, resplenptr unsafe.Pointer, reqptr unsafe.Pointer, reqptrlen int32) int32

func NewRunner() *Runner {
	l := logger.NewWithSync(&wasmWriteSyncer{})

	return &Runner{
		sendResponse: SendResponseFn,
		sdkFactory: func(sdkConfig *RuntimeConfig, opts ...func(*RuntimeConfig)) *Runtime {
			for _, opt := range opts {
				opt(sdkConfig)
			}

			return &Runtime{
				logger:  l,
				fetchFn: createFetchFn(sdkConfig, l, fetch),
				emitFn:  createEmitFn(sdkConfig, l, emit),
			}
		},
		args: os.Args,
	}
}

// SendResponseFn implements sendResponse for import into WASM.
func SendResponseFn(response *wasmpb.Response) {
	pb, err := proto.Marshal(response)
	if err != nil {
		// We somehow couldn't marshal the response, so let's
		// exit with a special error code letting the host know
		// what happened.
		os.Exit(CodeInvalidResponse)
	}

	// UnknownID will only be set when we've failed to parse
	// the request. Like before, let's bubble this up.
	if response.Id == UnknownID {
		os.Exit(CodeInvalidRequest)
	}

	ptr, ptrlen, err := BufferToPointerLen(pb)
	if err != nil {
		os.Exit(CodeInvalidResponse)
	}
	errno := sendResponse(ptr, ptrlen)
	if errno != 0 {
		os.Exit(CodeHostErr)
	}

	code := CodeSuccess
	if response.ErrMsg != "" {
		code = CodeRunnerErr
	}

	os.Exit(code)
}

type wasmWriteSyncer struct{}

// Write is used to proxy log requests from the WASM binary back to the host
func (wws *wasmWriteSyncer) Write(p []byte) (n int, err error) {
	ptr, ptrlen, err := BufferToPointerLen(p)
	if err != nil {
		return int(ptrlen), err
	}
	log(ptr, ptrlen)
	return int(ptrlen), nil
}
