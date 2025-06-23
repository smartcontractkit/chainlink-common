package wasm

import (
	"errors"
	"math/rand"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"google.golang.org/protobuf/proto"
)

type runtimeInternals interface {
	callCapability(req unsafe.Pointer, reqLen int32) int64
	awaitCapabilities(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64
	getSecrets(req unsafe.Pointer, reqLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64
	awaitSecrets(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64
	switchModes(mode int32)
	getSeed(mode int32) int64
}

func newRuntime(internals runtimeInternals, mode sdkpb.Mode) sdkimpl.RuntimeBase {
	return sdkimpl.RuntimeBase{
		Mode:           mode,
		RuntimeHelpers: &runtimeHelper{runtimeInternals: internals},
	}
}

type runtimeHelper struct {
	runtimeInternals
	donSource  rand.Source
	nodeSource rand.Source
}

func (r *runtimeHelper) GetSource(mode sdkpb.Mode) rand.Source {
	switch mode {
	case sdkpb.Mode_DON:
		if r.donSource == nil {
			seed := r.getSeed(int32(mode))
			r.donSource = rand.NewSource(seed)
		}
		return r.donSource
	default:
		if r.nodeSource == nil {
			seed := r.getSeed(int32(mode))
			r.nodeSource = rand.NewSource(seed)
		}
		return r.nodeSource
	}
}

func (r *runtimeHelper) GetSecrets(request *sdkpb.GetSecretsRequest, maxResponseSize uint64) error {
	marshalled, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	marshalledPtr, marshalledLen, err := bufferToPointerLen(marshalled)
	if err != nil {
		return err
	}

	response := make([]byte, maxResponseSize)
	responsePtr, responseLen, err := bufferToPointerLen(response)
	if err != nil {
		return err
	}

	bytes := r.getSecrets(marshalledPtr, marshalledLen, responsePtr, responseLen)
	if bytes < 0 {
		return errors.New(string(response[:-bytes]))
	}

	return nil
}

func (r *runtimeHelper) AwaitSecrets(request *sdkpb.AwaitSecretsRequest, maxResponseSize uint64) (*pb.AwaitSecretsResponse, error) {
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

	bytes := r.awaitSecrets(mptr, mlen, responsePtr, responseLen)
	if bytes < 0 {
		return nil, errors.New(string(response[:-bytes]))
	}

	awaitResponse := &sdkpb.AwaitSecretsResponse{}
	err = proto.Unmarshal(response[:bytes], awaitResponse)
	if err != nil {
		return nil, err
	}

	return awaitResponse, nil
}

func (r *runtimeHelper) Call(request *sdkpb.CapabilityRequest) error {
	marshalled, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	marshalledPtr, marshalledLen, err := bufferToPointerLen(marshalled)
	if err != nil {
		return err
	}

	// TODO (CAPPL-846): callCapability should also have a response pointer and response pointer buffer
	result := r.callCapability(marshalledPtr, marshalledLen)
	if result < 0 {
		return errors.New("cannot find capability " + request.Id)
	}

	return nil
}

func (r *runtimeHelper) Await(request *sdkpb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*sdkpb.AwaitCapabilitiesResponse, error) {
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

	bytes := r.awaitCapabilities(mptr, mlen, responsePtr, responseLen)
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

func (r *runtimeHelper) SwitchModes(mode sdkpb.Mode) {
	r.switchModes(int32(mode))
}
