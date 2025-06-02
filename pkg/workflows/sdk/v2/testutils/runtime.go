package testutils

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/consensus/consensusmock"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils/registry"
	"google.golang.org/protobuf/proto"
)

func newRuntime(tb testing.TB, configBytes []byte, writer *testWriter) sdkimpl.RuntimeBase {
	tb.Cleanup(func() { delete(calls, tb) })

	defaultConsensus, err := consensusmock.NewConsensusCapability(tb)

	// Do not override if the user provided their own consensus method
	if err == nil {
		defaultConsensus.Simple = defaultSimpleConsensus
	}

	return sdkimpl.RuntimeBase{
		ConfigBytes:     configBytes,
		MaxResponseSize: sdk.DefaultMaxResponseSizeBytes,
		Call:            createCallCapability(tb),
		Await:           createAwaitCapabilities(tb),
		Writer:          writer,
	}
}

func defaultSimpleConsensus(_ context.Context, input *pb.SimpleConsensusInputs) (*valuespb.Value, error) {
	switch o := input.Observation.(type) {
	case *pb.SimpleConsensusInputs_Value:
		return o.Value, nil
	case *pb.SimpleConsensusInputs_Error:
		if input.Default.Value == nil {
			return nil, errors.New(o.Error)
		}

		return input.Default, nil
	default:
		return nil, fmt.Errorf("unknown observation type %T", o)
	}
}

var calls = map[testing.TB]map[int32]chan *pb.CapabilityResponse{}

func createCallCapability(tb testing.TB) func(request *pb.CapabilityRequest) error {
	return func(request *pb.CapabilityRequest) error {
		reg := registry.GetRegistry(tb)
		capability, err := reg.GetCapability(request.Id)
		if err != nil {
			return err
		}

		respCh := make(chan *pb.CapabilityResponse, 1)
		tbCalls, ok := calls[tb]
		if !ok {
			tbCalls = map[int32]chan *pb.CapabilityResponse{}
			calls[tb] = tbCalls
		}
		tbCalls[request.CallbackId] = respCh
		go func() {
			respCh <- capability.Invoke(tb.Context(), request)
		}()
		return nil
	}
}

func createAwaitCapabilities(tb testing.TB) sdkimpl.AwaitCapabilitiesFn {
	return func(request *pb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*pb.AwaitCapabilitiesResponse, error) {
		response := &pb.AwaitCapabilitiesResponse{Responses: map[int32]*pb.CapabilityResponse{}}

		testCalls, ok := calls[tb]
		if !ok {
			return nil, errors.New("no calls found for this test")
		}

		var errs []error
		for _, id := range request.Ids {
			ch, ok := testCalls[id]
			if !ok {
				errs = append(errs, fmt.Errorf("no call found for %d", id))
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
