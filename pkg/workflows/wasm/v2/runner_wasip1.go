package v2

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

//go:wasmimport env callcap
func callcap(reqptr unsafe.Pointer, reqptrlen int32) int32

//go:wasmimport env awaitcaps
func awaitcaps(respptr unsafe.Pointer, resplenptr unsafe.Pointer, reqptr unsafe.Pointer, reqptrlen int32) int32

const uint32Size = int32(4)

func NewRunnerV2() *RunnerV2 {
	return &RunnerV2{
		sendResponse: wasm.SendResponseFn,
		runtimeFactory: func(sdkConfig *wasm.RuntimeConfig, refToResponse map[string]capabilities.CapabilityResponse, hostReqID string) *RuntimeV2 {
			return &RuntimeV2{
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
	ptr, ptrlen, err := wasm.BufferToPointerLen(pb)
	if err != nil {
		return nil, err
	}

	respBuffer := make([]byte, 100000) // TODO max size?
	respptr, _, err := wasm.BufferToPointerLen(respBuffer)
	if err != nil {
		return nil, err
	}

	resplenBuffer := make([]byte, uint32Size)
	resplenptr, _, err := wasm.BufferToPointerLen(resplenBuffer)
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

func callCapFn(response *wasmpb.CapabilityCall) error {
	pb, err := proto.Marshal(response)
	if err != nil {
		return err
	}
	ptr, ptrlen, err := wasm.BufferToPointerLen(pb)
	if err != nil {
		return err
	}
	errno := callcap(ptr, ptrlen)
	if errno != 0 {
		return fmt.Errorf("callcap failed with errno %d", errno)
	}
	return nil
}
