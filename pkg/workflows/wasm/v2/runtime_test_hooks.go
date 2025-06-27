//go:build !wasip1

package wasm

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils/registry"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type runtimeInternalsTestHook struct {
	testTb                testing.TB
	awaitResponseOverride func() ([]byte, error)
	callCapabilityErr     bool
	outstandingCalls      map[int32]sdk.Promise[*sdkpb.CapabilityResponse]
	nodeSeed              int64
	donSeed               int64

	outstandingSecretsCalls map[int32]sdk.Promise[[]*sdkpb.SecretResponse]
	secrets                 map[string]*sdkpb.Secret
	mu                      sync.Mutex
}

func secretKey(namespace, id string) string {
	if namespace == "" {
		namespace = "default"
	}
	return fmt.Sprintf("%s.%s", namespace, id)
}

var _ runtimeInternals = (*runtimeInternalsTestHook)(nil)

func (r *runtimeInternalsTestHook) callCapability(req unsafe.Pointer, reqLen int32) int64 {
	if r.callCapabilityErr {
		return -1
	}

	reqBuff := unsafe.Slice((*byte)(req), reqLen)
	request := sdkpb.CapabilityRequest{}
	err := proto.Unmarshal(reqBuff, &request)
	require.NoError(r.testTb, err)

	reg := registry.GetRegistry(r.testTb)
	capability, err := reg.GetCapability(request.Id)
	require.NoError(r.testTb, err)

	respCh := make(chan *sdkpb.CapabilityResponse, 1)
	go func() {
		respCh <- capability.Invoke(r.testTb.Context(), &request)
	}()

	r.outstandingCalls[request.CallbackId] = sdk.NewBasicPromise(func() (*sdkpb.CapabilityResponse, error) {
		select {
		case resp := <-respCh:
			return resp, nil
		case <-r.testTb.Context().Done():
			return nil, r.testTb.Context().Err()
		}
	})

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

	responsepb := &sdkpb.AwaitCapabilitiesResponse{Responses: map[int32]*sdkpb.CapabilityResponse{}}
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

func (r *runtimeInternalsTestHook) getSecrets(req unsafe.Pointer, reqLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64 {
	require.Greater(r.testTb, maxResponseLen, int32(0))

	reqBuff := unsafe.Slice((*byte)(req), reqLen)
	request := sdkpb.GetSecretsRequest{}
	err := proto.Unmarshal(reqBuff, &request)
	require.NoError(r.testTb, err)

	respCh := make(chan []*sdkpb.SecretResponse, 1)
	go func() {
		resps := make([]*sdkpb.SecretResponse, 0, len(request.Requests))
		for _, req := range request.Requests {
			key := secretKey(req.Namespace, req.Id)
			var resp *sdkpb.SecretResponse
			if secret, ok := r.secrets[key]; ok {
				resp = &sdkpb.SecretResponse{
					Response: &sdkpb.SecretResponse_Secret{
						Secret: secret,
					},
				}
			} else {
				resp = &sdkpb.SecretResponse{
					Response: &sdkpb.SecretResponse_Error{
						Error: &sdkpb.SecretError{
							Namespace: req.Namespace,
							Id:        req.Id,
							Error:     fmt.Sprintf("secret %s not found", key),
						},
					},
				}
			}
			resps = append(resps, resp)
		}
		respCh <- resps
	}()

	r.outstandingSecretsCalls[request.CallbackId] = sdk.NewBasicPromise(func() ([]*sdkpb.SecretResponse, error) {
		select {
		case resp := <-respCh:
			return resp, nil
		case <-r.testTb.Context().Done():
			return nil, r.testTb.Context().Err()
		}
	})

	return 0
}

func (r *runtimeInternalsTestHook) awaitSecrets(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64 {
	require.Greater(r.testTb, maxResponseLen, int32(0))

	response := unsafe.Slice((*byte)(responseBuffer), maxResponseLen)

	awaitRequestBuff := unsafe.Slice((*byte)(awaitRequest), awaitRequestLen)
	requestpb := &sdkpb.AwaitSecretsRequest{}
	if err := proto.Unmarshal(awaitRequestBuff, requestpb); err != nil {
		msg := "failed to unmarshal await request"
		return readHostMessage(response, msg, true)
	}

	responsepb := &sdkpb.AwaitSecretsResponse{Responses: map[int32]*sdkpb.SecretResponses{}}
	for _, id := range requestpb.Ids {
		promise := r.outstandingSecretsCalls[id]
		result, err := promise.Await()
		if err != nil {
			msg := err.Error()
			return readHostMessage(response, msg, true)
		}
		responsepb.Responses[id] = &sdkpb.SecretResponses{Responses: result}
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

func (r *runtimeInternalsTestHook) switchModes(_ int32) {}

func (r *runtimeInternalsTestHook) getSeed(mode int32) int64 {
	switch mode {
	case int32(sdkpb.Mode_MODE_DON):
		return r.donSeed
	default:
		return r.nodeSeed
	}
}
