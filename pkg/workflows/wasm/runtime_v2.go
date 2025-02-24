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
	callCapFn     func(payload *wasmpb.CapabilityCall) (int32, error)
	awaitCapsFn   func(payload *wasmpb.AwaitRequest) (*wasmpb.AwaitResponse, error)
	refToResponse map[int32]capabilities.CapabilityResponse
}

var _ sdk.RuntimeV2 = (*RuntimeV2)(nil)

func (r *RuntimeV2) AwaitCapabilities(calls ...sdk.CapabilityCallPromise) error {
	pendingRequests := []int32{}
	for _, call := range calls {
		ref, _, _ := call.CallInfo()
		if response, ok := r.refToResponse[ref]; ok {
			call.Fulfill(response, nil)
		} else {
			pendingRequests = append(pendingRequests, ref)
		}
	}
	if len(pendingRequests) == 0 {
		// all already fulfilled
		return nil
	}

	resp, err := r.awaitCapsFn(&wasmpb.AwaitRequest{Refs: pendingRequests})
	if err != nil {
		return err
	}
	for _, call := range calls {
		ref, _, _ := call.CallInfo()
		if response, ok := resp.RefToResponse[ref]; ok {
			capResp, err2 := pb.CapabilityResponseFromProto(response)
			if err2 != nil {
				return err2
			}
			call.Fulfill(capResp, nil)
		} else {
			return fmt.Errorf("missing response for ref %s", ref)
		}
	}
	return nil
}

func (r *RuntimeV2) CallCapability(call sdk.CapabilityCallPromise) (int32, error) {
	ref, capId, request := call.CallInfo()
	if response, ok := r.refToResponse[ref]; ok {
		call.Fulfill(response, nil)
		return ref, nil
	}
	return r.callCapFn(&wasmpb.CapabilityCall{
		CapabilityId: capId,
		Request:      pb.CapabilityRequestToProto(request),
	})
}

type CapCall[Outputs any] struct {
	ref        int32
	capId      string
	capRequest capabilities.CapabilityRequest
	outputs    Outputs
	err        error
	fulfilled  bool
	mu         sync.Mutex
}

func (c *CapCall[Outputs]) CallInfo() (ref int32, capId string, request capabilities.CapabilityRequest) {
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
func CallCapability[Inputs any, Config any, Outputs any](runtime sdk.RuntimeV2, capId string, inputs Inputs, config Config) (*CapCall[Outputs], error) {
	inputsVal, err := values.CreateMapFromStruct(inputs)
	if err != nil {
		return nil, err
	}
	configVal, err := values.CreateMapFromStruct(config)
	if err != nil {
		return nil, err
	}
	call := &CapCall[Outputs]{
		capId: capId,
		capRequest: capabilities.CapabilityRequest{
			// TODO: Metadata?
			Inputs: inputsVal,
			Config: configVal,
		},
	}
	call.ref, err = runtime.CallCapability(call)
	return call, err
}
