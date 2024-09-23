package wasm

import (
	"os"
	"unsafe"

	"google.golang.org/protobuf/proto"

	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

//go:wasmimport env sendResponse
func sendResponse(respptr unsafe.Pointer, respptrlen int32)

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
			sendResponse(ptr, ptrlen)

			code := CodeSuccess
			if response.ErrMsg != "" {
				code = CodeRunnerErr
			}

			os.Exit(code)
		},
		args: os.Args,
	}
}
