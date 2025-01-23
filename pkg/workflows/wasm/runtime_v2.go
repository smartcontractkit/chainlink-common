package wasm

import (
	"fmt"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type RuntimeV2 struct {
	sendResponseFn func(payload *wasmpb.Response)
	refToResponse  map[string]capabilities.CapabilityResponse
	hostRequestID  string
}

var _ sdk.RuntimeV2 = (*RuntimeV2)(nil)

func (r *RuntimeV2) CallCapabilities(calls ...sdk.CapabilityCallPromise) error {
	missingRequests := make(map[string]*pb.CapabilityRequest)
	for _, call := range calls {
		ref, _, request := call.CallInfo()
		if response, ok := r.refToResponse[ref]; ok {
			call.Fulfill(response, nil)
		} else {
			missingRequests[ref] = pb.CapabilityRequestToProto(request)
		}
	}
	if len(missingRequests) == 0 {
		// all already fulfilled
		return nil
	}
	if len(missingRequests) != len(calls) {
		// only all-or-nothing
		return fmt.Errorf("partially missing responses")
	}
	// send back a response with all pending capability calls and terminate execution
	capCallsProtos := map[string]*wasmpb.CapabilityCall{}
	for _, call := range calls {
		ref, capId, request := call.CallInfo()
		capCallsProtos[ref] = &wasmpb.CapabilityCall{
			CapabilityId: capId,
			Request:      pb.CapabilityRequestToProto(request),
		}
	}
	// this will never return
	r.sendResponseFn(&wasmpb.Response{
		Id: r.hostRequestID,
		Message: &wasmpb.Response_RunResponse{
			RunResponse: &wasmpb.RunResponse{
				RefToCapCall: capCallsProtos,
			},
		},
	})
	return fmt.Errorf("should never reach here")
}

type CapCall[Outputs any] struct {
	ref        string
	capId      string
	capRequest capabilities.CapabilityRequest
	outputs    Outputs
	err        error
	fulfilled  bool
	mu         sync.Mutex
}

func (c *CapCall[Outputs]) CallInfo() (ref string, capId string, request capabilities.CapabilityRequest) {
	return c.ref, c.capId, c.capRequest
}

func (c *CapCall[Outputs]) Fulfill(response capabilities.CapabilityResponse, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = response.Value.UnwrapTo(&c.outputs)
}

func (c *CapCall[Outputs]) Result() (Outputs, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.fulfilled {
		return c.outputs, fmt.Errorf("not yet fulfilled")
	}
	return c.outputs, c.err
}

// TODO: maybe we could generate those for every capability individually?
func NewCapabilityCall[Inputs any, Config any, Outputs any](ref string, capId string, inputs Inputs, config Config) (*CapCall[Outputs], error) {
	inputsVal, err := values.CreateMapFromStruct(inputs)
	if err != nil {
		return nil, err
	}
	configVal, err := values.CreateMapFromStruct(config)
	if err != nil {
		return nil, err
	}

	return &CapCall[Outputs]{
		ref:   ref,
		capId: capId,
		capRequest: capabilities.CapabilityRequest{
			// TODO: Metadata?
			Inputs: inputsVal,
			Config: configVal,
		},
	}, nil
}
