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
func emit(respptr unsafe.Pointer, resplenptr unsafe.Pointer, reqptr unsafe.Pointer, reqptrlen int32) int32

func NewRunner() *Runner {
	l := logger.NewWithSync(&wasmWriteSyncer{})

	return &Runner{
		sendResponse: sendResponseFn,
		sdkFactory: func(sdkConfig *RuntimeConfig, opts ...func(*RuntimeConfig)) *Runtime {
			for _, opt := range opts {
				opt(sdkConfig)
			}

			return &Runtime{
				logger:  l,
				fetchFn: createFetchFn(sdkConfig, l),
				emitFn:  createEmitFn(sdkConfig, l, emit),
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
			StatusCode: uint8(response.StatusCode),
			Headers:    headersResp,
			Body:       response.Body,
		}, nil
	}
	return fetchFn
}

type wasmWriteSyncer struct{}

// Write is used to proxy log requests from the WASM binary back to the host
func (wws *wasmWriteSyncer) Write(p []byte) (n int, err error) {
	ptr, ptrlen := bufferToPointerLen(p)
	log(ptr, ptrlen)
	return int(ptrlen), nil
}
