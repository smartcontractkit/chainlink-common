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

func bufferToPointerLen(buf []byte) (unsafe.Pointer, int32) {
	return unsafe.Pointer(&buf[0]), int32(len(buf))
}

func NewRunner() *Runner {
	return &Runner{
		sendResponse: func(response *wasmpb.Response) {
			pb, err := proto.Marshal(response)
			if err != nil {
				// We somehow couldn't marshal the response, so let's
				// exit with a special error code letting the host know
				// what happened.
				os.Exit(CodeInvalidResponse)
			}

			// unknownID will only be set when we've failed to parse
			// the request. Like before, let's bubble this up.
			if response.Id == unknownID {
				os.Exit(CodeInvalidRequest)
			}

			ptr, ptrlen := bufferToPointerLen(pb)
			errno := sendResponse(ptr, ptrlen)
			if errno != 0 {
				os.Exit(CodeHostErr)
			}

			code := CodeSuccess
			if response.ErrMsg != "" {
				code = CodeRunnerErr
			}

			os.Exit(code)
		},
		SDK: Runtime{
			Logger: logger.NewWithSync(&wasmWriteSyncer{}),
		},
		args: os.Args,
	}
}

type wasmWriteSyncer struct{}

// Write is used to proxy log requests from the WASM binary back to the host
func (wws *wasmWriteSyncer) Write(p []byte) (n int, err error) {
	ptr, ptrlen := bufferToPointerLen(p)
	log(ptr, ptrlen)
	return int(ptrlen), nil
}
