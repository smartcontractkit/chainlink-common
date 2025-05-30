package wasm

import (
	"errors"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"google.golang.org/protobuf/proto"
)

type runtimeInternals interface {
	callCapability(req unsafe.Pointer, reqLen int32) int64
	awaitCapabilities(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64
}

func newRuntime(internals runtimeInternals, mode sdkpb.Mode) sdkimpl.RuntimeBase {
	return sdkimpl.RuntimeBase{
		Call:   callCapabilityWasmWrapper(internals),
		Await:  awaitCapabilitiesWasmWrapper(internals),
		Writer: &writer{},
		Mode:   mode,
	}
}

func callCapabilityWasmWrapper(internals runtimeInternals) func(request *sdkpb.CapabilityRequest) error {
	return func(request *sdkpb.CapabilityRequest) error {
		marshalled, err := proto.Marshal(request)
		if err != nil {
			return err
		}

		marshalledPtr, marshalledLen, err := bufferToPointerLen(marshalled)
		if err != nil {
			return err
		}

		// TODO (CAPPL-846): callCapability should also have a response pointer and response pointer buffer
		result := internals.callCapability(marshalledPtr, marshalledLen)
		if result < 0 {
			return errors.New("cannot find capability " + request.Id)
		}

		return nil
	}
}

func awaitCapabilitiesWasmWrapper(internals runtimeInternals) func(request *sdkpb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*sdkpb.AwaitCapabilitiesResponse, error) {
	return func(request *sdkpb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*sdkpb.AwaitCapabilitiesResponse, error) {

		m, err := proto.Marshal(request)
		if err != nil {
			return nil, err
		}

		mptr, mlen, err := bufferToPointerLen(m)
		if err != nil {
			return nil, err
		}

		response := make([]byte, maxResponseSize)
		responsePtr, responseLen, err := bufferToPointerLen(response)
		if err != nil {
			return nil, err
		}

		bytes := internals.awaitCapabilities(mptr, mlen, responsePtr, responseLen)
		if bytes < 0 {
			return nil, errors.New(string(response[:-bytes]))
		}

		awaitResponse := &sdkpb.AwaitCapabilitiesResponse{}
		err = proto.Unmarshal(response[:bytes], awaitResponse)
		if err != nil {
			return nil, err
		}

		return awaitResponse, nil
	}
}
