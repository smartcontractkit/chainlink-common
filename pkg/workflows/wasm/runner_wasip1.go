package wasm

import (
	"encoding/binary"
	"os"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

//go:wasmimport env sendResponse
func sendResponse(respptr unsafe.Pointer, respptrlen int32) (errno int32)

//go:wasmimport env log
func log(respptr unsafe.Pointer, respptrlen int32)

//go:wasmimport env fetch
func fetch(respptr unsafe.Pointer, resplenptr unsafe.Pointer, reqptr unsafe.Pointer, reqptrlen int32) int32

func bufferToPointerLen(buf []byte) (unsafe.Pointer, int32) {
	return unsafe.Pointer(&buf[0]), int32(len(buf))
}

func NewRunner() *Runner {
	l := logger.NewWithSync(&wasmWriteSyncer{})

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
			Logger: l,
			Fetch: func(req FetchRequest) FetchResponse {

				headerspb, err := values.NewMap(req.Headers)
				if err != nil {
					os.Exit(CodeRunnerErr)
				}

				b, err := proto.Marshal(&wasmpb.FetchRequest{
					Url:       req.URL,
					Method:    req.Method,
					Headers:   values.ProtoMap(headerspb),
					Body:      req.Body,
					TimeoutMs: req.TimeoutMs,
				})
				if err != nil {
					os.Exit(CodeRunnerErr)
				}
				reqptr, reqptrlen := bufferToPointerLen(b)

				respBuffer := make([]byte, 1024*5)
				respptr, _ := bufferToPointerLen(respBuffer)

				resplenBuffer := make([]byte, 4)
				resplenptr, _ := bufferToPointerLen(resplenBuffer)

				errno := fetch(respptr, resplenptr, reqptr, reqptrlen)
				if errno != 0 {
					l.Error("Error number: ", errno)
					os.Exit(CodeHostErr)
				}

				responseSize := binary.LittleEndian.Uint32(resplenBuffer)
				response := &wasmpb.FetchResponse{}
				err = proto.Unmarshal(respBuffer[:responseSize], response)
				if err != nil {
					l.Error("Error: ", err.Error())
					os.Exit(CodeRunnerErr)
				}

				fields := response.Headers.GetFields()
				headersResp := make(map[string]any, len(fields))
				for k, v := range fields {
					headersResp[k] = v
				}

				return FetchResponse{
					Success:      response.Success,
					ErrorMessage: response.ErrorMessage,
					StatusCode:   uint8(response.StatusCode),
					Headers:      headersResp,
					Body:         response.Body,
				}
			},
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
