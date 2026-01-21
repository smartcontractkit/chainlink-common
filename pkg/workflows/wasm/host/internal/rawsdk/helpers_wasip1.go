package rawsdk

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"
	"unsafe"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

const (
	ErrnoSuccess = 0
)

func GetRequest() *sdk.ExecuteRequest {
	if len(os.Args) != 2 {
		SendError(errors.New("invalid request: request must contain a payload"))
	}

	request := os.Args[1]
	if request == "" {
		SendError(errors.New("invalid request: request cannot be empty"))
	}

	b := Must(base64.StdEncoding.DecodeString(request))

	req := &sdk.ExecuteRequest{}
	if err := proto.Unmarshal(b, req); err != nil {
		SendError(err)
	}
	return req
}

func SendResponse(result any) {
	wrapped := values.Proto(Must(values.Wrap(result)))
	execResult := &sdk.ExecutionResult{Result: &sdk.ExecutionResult_Value{Value: wrapped}}
	bytes := Must(proto.Marshal(execResult))
	sendResponse(BufferToPointerLen(bytes))
	os.Exit(0)
}

func SendError(err error) {
	execResult := &sdk.ExecutionResult{Result: &sdk.ExecutionResult_Error{Error: err.Error()}}
	bytes := Must(proto.Marshal(execResult))
	sendResponse(BufferToPointerLen(bytes))
	os.Exit(0)
}

func SendSubscription(subscriptions *sdk.TriggerSubscriptionRequest) {
	execResult := &sdk.ExecutionResult{Result: &sdk.ExecutionResult_TriggerSubscriptions{TriggerSubscriptions: subscriptions}}
	sendResponse(BufferToPointerLen(Must(proto.Marshal(execResult))))
}

func Now() time.Time {
	var buf [8]byte // host writes UnixNano as little-endian uint64
	rc := now(unsafe.Pointer(&buf[0]))
	if rc != ErrnoSuccess {
		panic(fmt.Errorf("failed to fetch time from host: now() returned errno %d", rc))
	}
	ns := int64(binary.LittleEndian.Uint64(buf[:]))
	return time.Unix(0, ns)
}

var donCall = int32(0)
var nodeCall = int32(-1)

var NodeOutputConsensusDescriptor = &sdk.ConsensusDescriptor{
	Descriptor_: &sdk.ConsensusDescriptor_FieldsMap{
		FieldsMap: &sdk.FieldsMap{
			Fields: map[string]*sdk.ConsensusDescriptor{
				"OutputThing": {
					Descriptor_: &sdk.ConsensusDescriptor_Aggregation{
						Aggregation: sdk.AggregationType_AGGREGATION_TYPE_MEDIAN,
					},
				},
			},
		},
	},
}

func DoRequestAsync(capabilityId, method string, mode sdk.Mode, input proto.Message) int32 {
	var callbackId int32
	if mode == sdk.Mode_MODE_NODE {
		callbackId = nodeCall
		nodeCall--
	} else {
		callbackId = donCall
		donCall++
	}

	req := &sdk.CapabilityRequest{
		Id:         capabilityId,
		Payload:    Must(anypb.New(input)),
		Method:     method,
		CallbackId: callbackId,
	}

	if callCapability(BufferToPointerLen(Must(proto.Marshal(req)))) < 0 {
		SendError(errors.New("callCapability returned an error"))
	}

	return callbackId
}

func DoConsensusRequest(capabilityId string, input *sdk.SimpleConsensusInputs, output *valuespb.Value) {
	Await(DoRequestAsync(capabilityId, "Simple", sdk.Mode_MODE_DON, input), output)
}

func DoRequest[I, O proto.Message](capabilityId, method string, mode sdk.Mode, input I, output O) {
	Await(DoRequestAsync(capabilityId, method, mode, input), output)
}

func DoRequestErr[I proto.Message](capabilityId, method string, mode sdk.Mode, input I) error {
	callbackId := DoRequestAsync(capabilityId, method, mode, input)

	resp := &sdk.AwaitCapabilitiesResponse{}
	await(&sdk.AwaitCapabilitiesRequest{Ids: []int32{callbackId}}, resp, awaitCapabilities)

	errMsg := resp.Responses[callbackId].GetError()
	return errors.New(errMsg)
}

func GetSecret(id string) (string, error) {
	callbackId := donCall
	donCall++
	marshalled := Must(proto.Marshal(&sdk.GetSecretsRequest{
		Requests: []*sdk.SecretRequest{
			{
				Id: id,
			},
		},
		CallbackId: callbackId,
	}))

	marshalledPtr, marshalledLen := BufferToPointerLen(marshalled)

	response := make([]byte, 1024*1024)
	responsePtr, responseLen := BufferToPointerLen(response)

	bytes := getSecrets(marshalledPtr, marshalledLen, responsePtr, responseLen)
	if bytes < 0 {
		SendError(errors.New("callCapability returned an error"))
	}

	req := &sdk.AwaitSecretsRequest{Ids: []int32{callbackId}}
	resp := &sdk.AwaitSecretsResponse{}
	await(req, resp, awaitSecrets)
	if len(resp.Responses) != 1 {
		SendError(fmt.Errorf("expected 1 response, got %d", len(resp.Responses)))
	}

	responses := resp.Responses[callbackId].Responses
	if len(responses) != 1 {
		SendError(fmt.Errorf("expected 1 secret response, got %d", len(responses)))
	}

	switch r := responses[0].Response.(type) {
	case *sdk.SecretResponse_Secret:
		return r.Secret.Value, nil
	case *sdk.SecretResponse_Error:
		return "", errors.New(r.Error.Error)
	default:
		SendError(fmt.Errorf("unexpected response type: %T", r))
	}

	return "", nil
}

func Await[O proto.Message](callbackId int32, output O) {
	resp := &sdk.AwaitCapabilitiesResponse{}
	await(&sdk.AwaitCapabilitiesRequest{Ids: []int32{callbackId}}, resp, awaitCapabilities)

	payload := resp.Responses[callbackId].GetPayload()
	if payload.UnmarshalTo(output) != nil {
		SendError(fmt.Errorf("failed to unmarshal capability response payload %s into %T", payload.TypeUrl, payload))
	}
}

type awaitFn func(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64

func await[I, O proto.Message](input I, output O, fn awaitFn) {
	awaitRequest := Must(proto.Marshal(input))

	mptr, mlen := BufferToPointerLen(awaitRequest)
	response := make([]byte, 1024*1024)
	responsePtr, responseLen := BufferToPointerLen(response)

	bytes := fn(mptr, mlen, responsePtr, responseLen)

	if bytes < 0 {
		response = response[:-bytes]
		SendError(fmt.Errorf("awaitCapabilities returned an error %s", string(response)))
	}

	if proto.Unmarshal(response[:bytes], output) != nil {
		SendError(errors.New("failed to proto unmarshal await response"))
	}
}

// BufferToPointerLen returns a pointer to the first element of the buffer and the length of the buffer.
func BufferToPointerLen(buf []byte) (unsafe.Pointer, int32) {
	if len(buf) == 0 {
		SendError(errors.New("buffer cannot be empty"))
	}
	return unsafe.Pointer(&buf[0]), int32(len(buf))
}

func Must[T any](v T, err error) T {
	if err != nil {
		SendError(err)
		os.Exit(0)
	}
	return v
}

//go:wasmimport env send_response
func sendResponse(response unsafe.Pointer, responseLen int32) int32

//go:wasmimport env switch_modes
func SwitchModes(mode int32)

//go:wasmimport env now
func now(response unsafe.Pointer) int32

//go:wasmimport env call_capability
func callCapability(req unsafe.Pointer, reqLen int32) int64

//go:wasmimport env await_capabilities
func awaitCapabilities(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64

//go:wasmimport env random_seed
func GetSeed(mode int32) int64

//go:wasmimport env log
func Log(message unsafe.Pointer, messageLen int32)

//go:wasmimport env get_secrets
func getSecrets(req unsafe.Pointer, reqLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64

//go:wasmimport env await_secrets
func awaitSecrets(req unsafe.Pointer, reqLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64
