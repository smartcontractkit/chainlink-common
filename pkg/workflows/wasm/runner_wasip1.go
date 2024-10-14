package wasm

import (
	"encoding/binary"
	"errors"
	"os"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

//go:wasmimport env sendResponse
func sendResponse(respptr unsafe.Pointer, respptrlen int32) (errno int32)

//go:wasmimport env log
func log(respptr unsafe.Pointer, respptrlen int32)

//go:wasmimport env fetch
func fetch(respptr unsafe.Pointer, resplenptr unsafe.Pointer, reqptr unsafe.Pointer, reqptrlen int32) int32

//go:wasmimport env emit
func emit(pbptr unsafe.Pointer, pblen int32) int32

func NewRunner() *Runner {
	l := logger.NewWithSync(&wasmWriteSyncer{})

	return &Runner{
		sendResponse: sendResponseFn,
		sdkFactory: func(sdkConfig *RuntimeConfig) *Runtime {
			return &Runtime{
				logger:  l,
				fetchFn: createFetchFn(sdkConfig, l),
				emitFn:  emitFn,
			}
		},
		args: os.Args,
	}
}

// sendResponseFn implements sendResponse for import into WASM.
func sendResponseFn(response *wasmpb.Response) {
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
}

// createFetchFn injects dependencies and creates a fetch function that can be used by the WASM
// binary.
func createFetchFn(
	sdkConfig *RuntimeConfig,
	l logger.Logger,
) func(sdk.FetchRequest) (sdk.FetchResponse, error) {
	fetchFn := func(req sdk.FetchRequest) (sdk.FetchResponse, error) {
		headerspb, err := values.NewMap(req.Headers)
		if err != nil {
			os.Exit(CodeInvalidRequest)
		}

		b, err := proto.Marshal(&wasmpb.FetchRequest{
			Url:       req.URL,
			Method:    req.Method,
			Headers:   values.ProtoMap(headerspb),
			Body:      req.Body,
			TimeoutMs: req.TimeoutMs,
		})
		if err != nil {
			os.Exit(CodeInvalidRequest)
		}
		reqptr, reqptrlen := bufferToPointerLen(b)

		respBuffer := make([]byte, sdkConfig.MaxFetchResponseSizeBytes)
		respptr, _ := bufferToPointerLen(respBuffer)

		resplenBuffer := make([]byte, uint32Size)
		resplenptr, _ := bufferToPointerLen(resplenBuffer)

		errno := fetch(respptr, resplenptr, reqptr, reqptrlen)
		if errno != 0 {
			os.Exit(CodeRunnerErr)
		}

		responseSize := binary.LittleEndian.Uint32(resplenBuffer)
		response := &wasmpb.FetchResponse{}
		err = proto.Unmarshal(respBuffer[:responseSize], response)
		if err != nil {
			l.Errorw("failed to unmarshal fetch response", "error", err.Error())
			os.Exit(CodeInvalidResponse)
		}

		fields := response.Headers.GetFields()
		headersResp := make(map[string]any, len(fields))
		for k, v := range fields {
			headersResp[k] = v
		}

		if response.ErrorMessage != "" {
			return sdk.FetchResponse{}, errors.New(response.ErrorMessage)
		}

		return sdk.FetchResponse{
			Success:    response.Success,
			StatusCode: uint8(response.StatusCode),
			Headers:    headersResp,
			Body:       response.Body,
		}, nil
	}
	return fetchFn
}

func emitFn(msg string, labels map[string]any) error {
	vm, err := values.NewMap(labels)
	if err != nil {
		return err
	}

	b, err := proto.Marshal(&wasmpb.CustomEmitMessage{
		Message: msg,
		Labels:  values.ProtoMap(vm),
	})

	ptr, ptrlen := bufferToPointerLen(b)
	errno := emit(ptr, ptrlen)
	if errno != 0 {
		os.Exit(CodeRunnerErr)
	}

	return nil
}

type wasmWriteSyncer struct{}

// Write is used to proxy log requests from the WASM binary back to the host
func (wws *wasmWriteSyncer) Write(p []byte) (n int, err error) {
	ptr, ptrlen := bufferToPointerLen(p)
	log(ptr, ptrlen)
	return int(ptrlen), nil
}

const uint32Size = int32(4)

// bufferToPointerLen returns a pointer to the first element of the buffer and the length of the buffer.
func bufferToPointerLen(buf []byte) (unsafe.Pointer, int32) {
	return unsafe.Pointer(&buf[0]), int32(len(buf))
}
