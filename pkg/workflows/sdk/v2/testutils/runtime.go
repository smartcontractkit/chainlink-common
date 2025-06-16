package testutils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/consensus/consensusmock"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils/registry"
	"google.golang.org/protobuf/proto"
)

func newRuntime(tb testing.TB, sourceFn func() rand.Source) sdkimpl.RuntimeBase {
	defaultConsensus, err := consensusmock.NewConsensusCapability(tb)

	// Do not override if the user provided their own consensus method
	if err == nil {
		defaultConsensus.Simple = defaultSimpleConsensus
	}

	return sdkimpl.RuntimeBase{
		MaxResponseSize: sdk.DefaultMaxResponseSizeBytes,
		RuntimeHelpers:  &runtimeHelpers{tb: tb, calls: map[int32]chan *pb.CapabilityResponse{}, sourceFn: sourceFn},
	}
}

func defaultSimpleConsensus(_ context.Context, input *pb.SimpleConsensusInputs) (*valuespb.Value, error) {
	switch o := input.Observation.(type) {
	case *pb.SimpleConsensusInputs_Value:
		return o.Value, nil
	case *pb.SimpleConsensusInputs_Error:
		if input.Default == nil || input.Default.Value == nil {
			return nil, errors.New(o.Error)
		}

		return input.Default, nil
	default:
		return nil, fmt.Errorf("unknown observation type %T", o)
	}
}

type runtimeHelpers struct {
	tb       testing.TB
	calls    map[int32]chan *pb.CapabilityResponse
	sourceFn func() rand.Source
}

func (rh *runtimeHelpers) GetSource(_ pb.Mode) rand.Source {
	return rh.sourceFn()
}

func (rh *runtimeHelpers) Call(request *pb.CapabilityRequest) error {
	reg := registry.GetRegistry(rh.tb)
	capability, err := reg.GetCapability(request.Id)
	if err != nil {
		return err
	}

	respCh := make(chan *pb.CapabilityResponse, 1)
	rh.calls[request.CallbackId] = respCh
	go func() {
		respCh <- capability.Invoke(rh.tb.Context(), request)
	}()
	return nil
}

func (rh *runtimeHelpers) Await(request *pb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*pb.AwaitCapabilitiesResponse, error) {
	response := &pb.AwaitCapabilitiesResponse{Responses: map[int32]*pb.CapabilityResponse{}}

	var errs []error
	for _, id := range request.Ids {
		ch, ok := rh.calls[id]
		if !ok {
			errs = append(errs, fmt.Errorf("no call found for %d", id))
			continue
		}
		select {
		case resp := <-ch:
			response.Responses[id] = resp
		case <-rh.tb.Context().Done():
			return nil, rh.tb.Context().Err()
		}
	}

	bytes, _ := proto.Marshal(response)
	if len(bytes) > int(maxResponseSize) {
		return nil, errors.New(sdk.ResponseBufferTooSmall)
	}

	return response, errors.Join(errs...)
}

func (rh *runtimeHelpers) SwitchModes(_ pb.Mode) {}
