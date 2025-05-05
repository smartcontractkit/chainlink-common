package wasm

import (
	"errors"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"google.golang.org/protobuf/proto"
)

func newRuntime() sdkimpl.RuntimeBase {
	return sdkimpl.RuntimeBase{
		Call:   callCapabilityWasmWrapper,
		Await:  awaitCapabilitiesWasmWrapper,
		Writer: &writer{},
	}
}

func callCapabilityWasmWrapper(request *sdkpb.CapabilityRequest) ([]byte, error) {
	marshalled, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	id := make([]byte, sdk.IdLen)
	marshalledPtr, marshalledLen, err := bufferToPointerLen(marshalled)
	if err != nil {
		return nil, err
	}

	result := callCapability(marshalledPtr, marshalledLen, unsafe.Pointer(&id[0]))
	if result < 0 {
		return nil, errors.New("cannot find capability " + request.Id)
	}

	return id, nil
}

func awaitCapabilitiesWasmWrapper(request *sdkpb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*sdkpb.AwaitCapabilitiesResponse, error) {
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

	bytes := awaitCapabilities(mptr, mlen, responsePtr, responseLen)
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
