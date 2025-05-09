//go:build !wasip1

package wasm

import (
	"testing"
	"unsafe"

	"github.com/google/uuid"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type runtimeInternalsTestHook struct {
	testTb                testing.TB
	awaitResponseOverride func() ([]byte, error)
	callCapabilityErr     bool
	executionId           string
	outstandingCalls      map[string]sdk.Promise[*sdkpb.CapabilityResponse]
}

var _ runtimeInternals = (*runtimeInternalsTestHook)(nil)

func (r *runtimeInternalsTestHook) callCapability(req unsafe.Pointer, reqLen int32, id unsafe.Pointer) int64 {
	if r.callCapabilityErr {
		return -1
	}

	reqBuff := unsafe.Slice((*byte)(req), reqLen)
	request := sdkpb.CapabilityRequest{}
	err := proto.Unmarshal(reqBuff, &request)
	require.NoError(r.testTb, err)

	assert.Equal(r.testTb, r.executionId, request.ExecutionId)

	registry := testutils.GetRegistry(r.testTb)
	capability, err := registry.GetCapability(request.Id)
	require.NoError(r.testTb, err)

	callId := uuid.New()
	respCh := make(chan *sdkpb.CapabilityResponse, 1)
	go func() {
		respCh <- capability.Invoke(r.testTb.Context(), &request)
	}()

	r.outstandingCalls[callId.String()] = sdk.NewBasicPromise(func() (*sdkpb.CapabilityResponse, error) {
		select {
		case resp := <-respCh:
			return resp, nil
		case <-r.testTb.Context().Done():
			return nil, r.testTb.Context().Err()
		}
	})

	idBuffer := unsafe.Slice((*byte)(id), sdk.IdLen)
	copy(idBuffer, callId.String())

	return 0
}

func (r *runtimeInternalsTestHook) awaitCapabilities(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64 {
	if r.awaitResponseOverride != nil {
		awaitResponse, err := r.awaitResponseOverride()
		if err != nil {
			awaitResponse = []byte(err.Error())
		}

		copy(unsafe.Slice((*byte)(responseBuffer), maxResponseLen), awaitResponse)
		responseLen := int64(len(awaitResponse))
		if err != nil {
			responseLen = -responseLen
		}

		return responseLen
	}

	response := unsafe.Slice((*byte)(responseBuffer), maxResponseLen)

	awaitRequestBuff := unsafe.Slice((*byte)(awaitRequest), awaitRequestLen)
	requestpb := &sdkpb.AwaitCapabilitiesRequest{}
	if err := proto.Unmarshal(awaitRequestBuff, requestpb); err != nil {
		msg := "failed to unmarshal await request"
		return readHostMessage(response, msg, true)
	}

	responsepb := &sdkpb.AwaitCapabilitiesResponse{Responses: map[string]*sdkpb.CapabilityResponse{}}
	for _, id := range requestpb.Ids {
		promise := r.outstandingCalls[id]
		result, err := promise.Await()
		if err != nil {
			result = &sdkpb.CapabilityResponse{
				Response: &sdkpb.CapabilityResponse_Error{Error: err.Error()},
			}
		}
		responsepb.Responses[id] = result
	}

	responseBytes, err := proto.Marshal(responsepb)
	if err != nil {
		msg := "failed to marshal response"
		return readHostMessage(response, msg, true)
	}

	if len(responseBytes) > int(maxResponseLen) {
		msg := "response too large"
		return readHostMessage(response, msg, true)
	}
	copy(response, responseBytes)
	return int64(len(responseBytes))
}

func readHostMessage(response []byte, msg string, isError bool) int64 {
	if len(msg) > len(response) {
		msg = msg[:len(response)]
	}
	copy(response, msg)

	written := int64(len(msg))
	if isError {
		return -written
	}

	return written
}
