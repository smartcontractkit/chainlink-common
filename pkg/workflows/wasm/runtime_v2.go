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

type donRuntime struct {
	callCapFn     func(payload *wasmpb.CapabilityCall) (int32, error)
	awaitCapsFn   func(payload *wasmpb.AwaitRequest) (*wasmpb.AwaitResponse, error)
	refToResponse map[int32]capabilities.CapabilityResponse
}

func (r *donRuntime) RunInNodeModeWithConsensus(fn func(nodeRuntime sdk.NodeRuntime) (values.Value, error), consensus sdk.Consensus) sdk.Promise[values.Value] {
	//TODO implement me
	panic("implement me")
}

func (r *donRuntime) CallCapability(capId string, request capabilities.CapabilityRequest) sdk.Promise[*values.Map] {
	id, err := r.callCapFn(&wasmpb.CapabilityCall{
		CapabilityId: capId,
		Request:      pb.CapabilityRequestToProto(request),
	})

	promise := &CapCall{
		ref:        id,
		capRequest: request,
		runtime:    r,
	}

	if err != nil {
		promise.Fulfill(capabilities.CapabilityResponse{}, err)
	}

	return promise
}

func (r *donRuntime) AwaitCapabilities(calls ...sdk.CapabilityCallPromise) error {
	var pendingRequests []int32
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
			call.Fulfill(capResp, err2)
		} else {
			return fmt.Errorf("missing response for ref %s", ref)
		}
	}
	return nil
}

var _ sdk.DonRuntime = (*donRuntime)(nil)

type CapCall struct {
	ref        int32
	capId      string
	capRequest capabilities.CapabilityRequest
	outputs    *values.Map
	err        error
	fulfilled  bool
	mu         sync.Mutex
	runtime    sdk.RuntimeBase
}

func (c *CapCall) Await() (*values.Map, error) {
	// lock not needed, if c is fulfilled nothing will change it.
	// if it's not, awaiting a second time won't cause problems.
	if !c.fulfilled {
		_ = c.runtime.AwaitCapabilities(c)
	}

	return c.outputs, c.err
}

func (c *CapCall) CallInfo() (ref int32, capId string, request capabilities.CapabilityRequest) {
	return c.ref, c.capId, c.capRequest
}

func (c *CapCall) Fulfill(response capabilities.CapabilityResponse, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fulfillNoLock(response, err)
}

func (c *CapCall) fulfillNoLock(response capabilities.CapabilityResponse, err error) {
	c.fulfilled = true
	c.err = err
	c.outputs = response.Value
}
