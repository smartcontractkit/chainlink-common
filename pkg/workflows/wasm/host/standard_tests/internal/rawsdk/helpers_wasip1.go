package rawsdk

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
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

func SendSubscription(subscriptions *sdkpb.TriggerSubscriptionRequest) {
	execResult := &pb.ExecutionResult{Result: &pb.ExecutionResult_TriggerSubscriptions{TriggerSubscriptions: subscriptions}}
	sendResponse(BufferToPointerLen(Must(proto.Marshal(execResult))))
}

var donCall = int32(0)
var nodeCall = int32(-1)

var NodeOutputConsensusDescriptor = &sdkpb.ConsensusDescriptor{
	Descriptor_: &sdkpb.ConsensusDescriptor_FieldsMap{
		FieldsMap: &sdkpb.FieldsMap{
			Fields: map[string]*sdkpb.ConsensusDescriptor{
				"OutputThing": {
					Descriptor_: &sdkpb.ConsensusDescriptor_Aggregation{
						Aggregation: sdkpb.AggregationType_IDENTICAL,
					},
				},
			},
		},
	},
}

func MakeRequestAsync(capabilityId, method string, mode sdkpb.Mode, input proto.Message) int32 {
	var callbackId int32
	if mode == sdkpb.Mode_Node {
		callbackId = nodeCall
		nodeCall--
	} else {
		callbackId = donCall
		donCall++
	}

	req := &sdkpb.CapabilityRequest{
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

func MakeRequest[I, O proto.Message](capabilityId, method string, mode sdkpb.Mode, input I, output O) {
	Await(MakeRequestAsync(capabilityId, method, mode, input), output)
}

func Await[O proto.Message](callbackId int32, output O) {
	await := Must(proto.Marshal(&sdkpb.AwaitCapabilitiesRequest{Ids: []int32{callbackId}}))

	mptr, mlen := BufferToPointerLen(await)
	response := make([]byte, 1024*1024)
	responsePtr, responseLen := BufferToPointerLen(response)

	bytes := awaitCapabilities(mptr, mlen, responsePtr, responseLen)

	if bytes < 0 {
		SendError(errors.New("awaitCapabilities returned an error"))
	}

	awaitResponse := &sdkpb.AwaitCapabilitiesResponse{}

	if proto.Unmarshal(response[:bytes], awaitResponse) != nil {
		SendError(errors.New("failed to proto unmarshal await response"))
	}

	payload := awaitResponse.Responses[callbackId].GetPayload()
	if payload.UnmarshalTo(output) != nil {
		SendError(fmt.Errorf("failed to unmarshal capability response payload %s into %T", payload.TypeUrl, output))
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
