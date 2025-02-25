package wasm

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
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

//go:wasmimport env callcap
func callcap(reqptr unsafe.Pointer, reqptrlen int32) int32

//go:wasmimport env awaitcaps
func awaitcaps(respptr unsafe.Pointer, resplenptr unsafe.Pointer, reqptr unsafe.Pointer, reqptrlen int32) int32

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
				fetchFn: createFetchFn(sdkConfig, l, fetch),
				emitFn:  createEmitFn(sdkConfig, l, emit),
			}
		},
		args: os.Args,
	}
}

func NewDonRunner() sdk.DonRunner {
	return &donRunner{
		sendResponse: sendResponseFn,
		runtimeFactory: func(sdkConfig *RuntimeConfig, refToResponse map[int32]capabilities.CapabilityResponse, hostReqID string) *donRuntime {
			return &donRuntime{
				callCapFn:     callCapFn,
				awaitCapsFn:   awaitCapsFn,
				refToResponse: refToResponse,
			}
		},
		args:     os.Args,
		triggers: map[string]triggerInfo{},
	}
}

func awaitCapsFn(payload *wasmpb.AwaitRequest) (*wasmpb.AwaitResponse, error) {
	pb, err := proto.Marshal(payload)
	if err != nil {
		return nil, err
	}
	ptr, ptrlen, err := bufferToPointerLen(pb)
	if err != nil {
		return nil, err
	}

	respBuffer := make([]byte, 100000) // TODO max size?
	respptr, _, err := bufferToPointerLen(respBuffer)
	if err != nil {
		return nil, err
	}

	resplenBuffer := make([]byte, uint32Size)
	resplenptr, _, err := bufferToPointerLen(resplenBuffer)
	if err != nil {
		return nil, err
	}

	errno := awaitcaps(respptr, resplenptr, ptr, ptrlen)
	if errno != 0 {
		return nil, fmt.Errorf("awaitcaps failed with errno %d", errno)
	}

	responseSize := binary.LittleEndian.Uint32(resplenBuffer)
	response := &wasmpb.AwaitResponse{}
	err = proto.Unmarshal(respBuffer[:responseSize], response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fetch response: %w", err)
	}
	return response, nil
}

func callCapFn(response *wasmpb.CapabilityCall) (int32, error) {
	pb, err := proto.Marshal(response)
	if err != nil {
		return -1, err
	}
	ptr, ptrlen, err := bufferToPointerLen(pb)
	if err != nil {
		return -1, err
	}
	id := callcap(ptr, ptrlen)
	if id < 0 {
		return -1, fmt.Errorf("callcap failed with errno %d", id)
	}
	return id, nil
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

	ptr, ptrlen, err := bufferToPointerLen(pb)
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
	ptr, ptrlen, err := bufferToPointerLen(p)
	if err != nil {
		return int(ptrlen), err
	}
	log(ptr, ptrlen)
	return int(ptrlen), nil
}
