package testutils

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func newRuntime(tb testing.TB, configBytes []byte) sdkimpl.RuntimeBase {
	tb.Cleanup(func() { delete(calls, tb) })
	registry := GetRegistry(tb)

	// If the user wants to mock consensus before creating a runtime, it's not harmful
	_ = registry.RegisterCapability(&consensusCapability{})

	return sdkimpl.RuntimeBase{
		ExecId:          tb.Name(),
		ConfigBytes:     configBytes,
		MaxResponseSize: sdk.DefaultMaxResponseSizeBytes,
		Call:            createCallCapability(tb),
		Await:           createAwaitCapabilities(tb),
		Writer:          &testWriter{},
	}
}

var calls = map[testing.TB]map[string]chan *pb.CapabilityResponse{}

func createCallCapability(tb testing.TB) func(request *pb.CapabilityRequest) ([]byte, error) {
	return func(request *pb.CapabilityRequest) ([]byte, error) {
		registry := GetRegistry(tb)
		capability, err := registry.GetCapability(request.Id)
		if err != nil {
			return nil, err
		}

		id := []byte(uuid.NewString())
		respCh := make(chan *pb.CapabilityResponse, 1)
		tbCalls, ok := calls[tb]
		if !ok {
			tbCalls = map[string]chan *pb.CapabilityResponse{}
			calls[tb] = tbCalls
		}
		tbCalls[string(id)] = respCh
		go func() {
			respCh <- capability.Invoke(tb.Context(), request)
		}()
		return id, nil
	}
}

func createAwaitCapabilities(tb testing.TB) sdkimpl.AwaitCapabilitiesFn {
	return func(request *pb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*pb.AwaitCapabilitiesResponse, error) {
		response := &pb.AwaitCapabilitiesResponse{Responses: map[string]*pb.CapabilityResponse{}}

		testCalls, ok := calls[tb]
		if !ok {
			return nil, errors.New("no calls found for this test")
		}

		var errs []error
		for _, id := range request.Ids {
			ch, ok := testCalls[id]
			if !ok {
				errs = append(errs, errors.New("no call found for "+id))
				continue
			}
			select {
			case resp := <-ch:
				response.Responses[id] = resp
			case <-tb.Context().Done():
				return nil, tb.Context().Err()
			}
		}

		bytes, _ := proto.Marshal(response)
		if len(bytes) > int(maxResponseSize) {
			return nil, errors.New(sdk.ResponseBufferTooSmall)
		}

		return response, errors.Join(errs...)
	}
}

// TODO https://smartcontract-it.atlassian.net/browse/CAPPL-816
type consensusCapability struct{}

func (c consensusCapability) Invoke(_ context.Context, request *pb.CapabilityRequest) *pb.CapabilityResponse {
	response := &pb.CapabilityResponse{}
	consensusRequest := &pb.BuiltInConsensusRequest{}
	if err := request.Payload.UnmarshalTo(consensusRequest); err != nil {
		response.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
	}

	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-798 determine what to do if values don't line up with consensus request
	switch o := consensusRequest.Observation.(type) {
	case *pb.BuiltInConsensusRequest_Value:
		return addValueToResponse(o.Value, response)
	case *pb.BuiltInConsensusRequest_Error:
		if consensusRequest.DefaultValue.Value != nil {
			return addValueToResponse(consensusRequest.DefaultValue, response)
		}
		response.Response = &pb.CapabilityResponse_Error{Error: o.Error}
	default:
		response.Response = &pb.CapabilityResponse_Error{Error: "unknown observation type"}

	}
	return response
}

func addValueToResponse(v *valuespb.Value, response *pb.CapabilityResponse) *pb.CapabilityResponse {
	a, err := anypb.New(v)
	if err == nil {
		response.Response = &pb.CapabilityResponse_Payload{Payload: a}
	} else {
		response.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
	}
	return response
}

func (c consensusCapability) InvokeTrigger(_ context.Context, request *pb.TriggerSubscription) (*pb.Trigger, error) {
	return nil, errors.New(fmt.Sprintf("Trigger %s not found", request.Method))
}

func (c consensusCapability) ID() string {
	return "consensus@1.0.0"
}
