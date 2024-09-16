package wasm

import (
	"os"
	"unsafe"

	"google.golang.org/protobuf/proto"

	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

var (
	codeMarshalErr = 222
	codeRunnerErr  = 223
	codeSuccess    = 0
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
				os.Exit(codeMarshalErr)
			}

			ptr, ptrlen := bufferToPointerLen(pb)
			sendResponse(ptr, ptrlen)

			code := codeSuccess
			if response.ErrMsg != "" {
				code = codeRunnerErr
			}

			os.Exit(code)
		},
		args: os.Args,
	}
}
