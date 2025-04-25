//go:build !wasip1

package wasm

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

var lock sync.Mutex

var awaitResponseOverride func() ([]byte, error)
var callCapabilityErr bool

func overrideCapabilityResponseForTest(t *testing.T, awaitResponse func() ([]byte, error)) {
	lock.Lock()
	defer lock.Unlock()
	awaitResponseOverride = awaitResponse
	t.Cleanup(func() {
		awaitResponseOverride = nil
	})
}

func callCapability(req unsafe.Pointer, reqLen int32, id unsafe.Pointer) int64 {
	lock.Lock()
	defer lock.Unlock()

	if callCapabilityErr {
		return -1
	}

	reqBuff := unsafe.Slice((*byte)(req), reqLen)
	request := sdkpb.CapabilityRequest{}
	err := proto.Unmarshal(reqBuff, &request)
	require.NoError(testTb, err)

	assert.Equal(testTb, executionId, request.ExecutionId)

	capability, err := registry.GetCapability(request.Id)
	require.NoError(testTb, err)

	callId := uuid.New()
	respCh := make(chan *sdkpb.CapabilityResponse, 1)
	go func() {
		// Test ended before the capability was called
		tmp := testTb
		if tmp == nil {
			return
		}
		respCh <- capability.Invoke(tmp.Context(), &request)
	}()

	outstandingCalls[callId.String()] = sdk.NewBasicPromise(func() (*sdkpb.CapabilityResponse, error) {
		select {
		case resp := <-respCh:
			return resp, nil
		case <-testTb.Context().Done():
			return nil, testTb.Context().Err()
		}
	})

	idBuffer := unsafe.Slice((*byte)(id), sdk.IdLen)
	copy(idBuffer, callId.String())

	return 0
}

func awaitCapabilities(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64 {
	lock.Lock()
	defer lock.Unlock()

	if awaitResponseOverride != nil {
		awaitResponse, err := awaitResponseOverride()
		if err != nil {
			awaitResponse = []byte(err.Error())
		}

		copy(unsafe.Slice((*byte)(responseBuffer), maxResponseLen), awaitResponse)
		return int64(-len(awaitResponse))
	}

	if maxResponseLen < 0 {
		return 0
	}

	response := unsafe.Slice((*byte)(responseBuffer), maxResponseLen)

	if registry == nil {
		msg := "test hook has not bee initialized"
		readHostMessage(response, msg, true)
	}

	awaitRequestBuff := unsafe.Slice((*byte)(awaitRequest), awaitRequestLen)
	requestpb := &sdkpb.AwaitCapabilitiesRequest{}
	if err := proto.Unmarshal(awaitRequestBuff, requestpb); err != nil {
		msg := "failed to unmarshal await request"
		return readHostMessage(response, msg, true)
	}

	responsepb := &sdkpb.AwaitCapabilitiesResponse{Responses: map[string]*sdkpb.CapabilityResponse{}}
	for _, id := range requestpb.Ids {
		promise, ok := outstandingCalls[id]
		if !ok {
			msg := "cannot find capability " + id
			return readHostMessage(response, msg, true)
		}
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
