package rawsdk

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func GetRequest() *pb.ExecuteRequest {
	if len(os.Args) != 2 {
		SendError(errors.New("invalid request: request must contain a payload"))
	}

	request := os.Args[1]
	if request == "" {
		SendError(errors.New("invalid request: request cannot be empty"))
	}

	b := Must(base64.StdEncoding.DecodeString(request))

	req := &pb.ExecuteRequest{}
	if err := proto.Unmarshal(b, req); err != nil {
		SendError(err)
	}
	return req
}

func SendResponse(result any) {
	wrapped := values.Proto(Must(values.Wrap(result)))
	execResult := &pb.ExecutionResult{Result: &pb.ExecutionResult_Value{Value: wrapped}}
	bytes := Must(proto.Marshal(execResult))
	sendResponse(BufferToPointerLen(bytes))
	os.Exit(0)
}

func SendError(err error) {
	execResult := &pb.ExecutionResult{Result: &pb.ExecutionResult_Error{Error: err.Error()}}
	bytes := Must(proto.Marshal(execResult))
	sendResponse(BufferToPointerLen(bytes))
	os.Exit(0)
}

func SendSubscription(subscriptions *pb.TriggerSubscriptionRequest) {
	execResult := &pb.ExecutionResult{Result: &pb.ExecutionResult_TriggerSubscriptions{TriggerSubscriptions: subscriptions}}
	sendResponse(BufferToPointerLen(Must(proto.Marshal(execResult))))
}

var donCall = int32(0)
var nodeCall = int32(-1)

var NodeOutputConsensusDescriptor = &pb.ConsensusDescriptor{
	Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
		FieldsMap: &pb.FieldsMap{
			Fields: map[string]*pb.ConsensusDescriptor{
				"OutputThing": {
					Descriptor_: &pb.ConsensusDescriptor_Aggregation{
						Aggregation: pb.AggregationType_AGGREGATION_TYPE_MEDIAN,
					},
				},
			},
		},
	},
}

func DoRequestAsync(capabilityId, method string, mode pb.Mode, input proto.Message) int32 {
	var callbackId int32
	if mode == pb.Mode_MODE_NODE {
		callbackId = nodeCall
		nodeCall--
	} else {
		callbackId = donCall
		donCall++
	}

	req := &pb.CapabilityRequest{
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

func DoRequest[I, O proto.Message](capabilityId, method string, mode pb.Mode, input I, output O) {
	Await(DoRequestAsync(capabilityId, method, mode, input), output)
}

func GetSecret(id string) (string, error) {
	callbackId := donCall
	donCall++
	marshalled := Must(proto.Marshal(&pb.GetSecretsRequest{
		Requests: []*pb.SecretRequest{
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

	req := &pb.AwaitSecretsRequest{Ids: []int32{callbackId}}
	resp := &pb.AwaitSecretsResponse{}
	await(req, resp, awaitSecrets)
	if len(resp.Responses) != 1 {
		SendError(fmt.Errorf("expected 1 response, got %d", len(resp.Responses)))
	}

	responses := resp.Responses[callbackId].Responses
	if len(responses) != 1 {
		SendError(fmt.Errorf("expected 1 secret response, got %d", len(responses)))
	}

	switch r := responses[0].Response.(type) {
	case *pb.SecretResponse_Secret:
		return r.Secret.Value, nil
	case *pb.SecretResponse_Error:
		return "", errors.New(r.Error)
	default:
		SendError(fmt.Errorf("unexpected response type: %T", r))
	}

	return "", nil
}

func Await[O proto.Message](callbackId int32, output O) {
	resp := &pb.AwaitCapabilitiesResponse{}
	await(&pb.AwaitCapabilitiesRequest{Ids: []int32{callbackId}}, resp, awaitCapabilities)

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
		SendError(errors.New("awaitCapabilities returned an error"))
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
