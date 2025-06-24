package host

import (
	"context"
	"fmt"
	"sync"

	"github.com/bytecodealliance/wasmtime-go/v28"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type execution[T any] struct {
	fetchRequestsCounter int
	response             T
	ctx                  context.Context
	capabilityResponses  map[int32]<-chan *sdkpb.CapabilityResponse
	lock                 sync.RWMutex
	module               *module
	executor             ExecutionHelper
	hasRun               bool
	mode                 sdkpb.Mode
	donSeed              int64
	nodeSeed             int64
}

// callCapAsync async calls a capability by placing execution results onto a
// channel and storing each channel with a unique identifier for future
// retrieval on await.
func (e *execution[T]) callCapAsync(ctx context.Context, req *sdkpb.CapabilityRequest) error {
	ch := make(chan *sdkpb.CapabilityResponse, 1)
	e.lock.Lock()
	defer e.lock.Unlock()
	e.capabilityResponses[req.CallbackId] = ch

	go func() {
		resp, err := e.executor.CallCapability(ctx, req)

		if err != nil {
			resp = &sdkpb.CapabilityResponse{
				Response: &sdkpb.CapabilityResponse_Error{
					Error: err.Error(),
				},
			}
		}

		select {
		case <-ctx.Done():
		case ch <- resp:
		}
	}()

	return nil
}

func (e *execution[T]) awaitCapabilities(ctx context.Context, acr *sdkpb.AwaitCapabilitiesRequest) (*sdkpb.AwaitCapabilitiesResponse, error) {
	responses := make(map[int32]*sdkpb.CapabilityResponse, len(acr.Ids))

	e.lock.Lock()
	defer e.lock.Unlock()
	for _, callId := range acr.Ids {
		ch, ok := e.capabilityResponses[callId]
		if !ok {
			return nil, fmt.Errorf("failed to get call from store : %d", callId)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("failed to wait for capability response %d : %w", callId, ctx.Err())
		case resp := <-ch:
			responses[callId] = resp
		}

		delete(e.capabilityResponses, callId)
	}

	return &sdkpb.AwaitCapabilitiesResponse{
		Responses: responses,
	}, nil
}

func (e *execution[T]) log(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
	lggr := e.module.cfg.Logger
	b, innerErr := wasmRead(caller, ptr, ptrlen)
	if innerErr != nil {
		lggr.Errorf("error calling log: %s", innerErr)
		return
	}

	lggr.Info(string(b))
}

func (e *execution[T]) getSeed(mode int32) int64 {
	switch sdkpb.Mode(mode) {
	case sdkpb.Mode_DON:
		return e.donSeed
	case sdkpb.Mode_Node:
		return e.nodeSeed
	}

	return -1
}

func (e *execution[T]) switchModes(_ *wasmtime.Caller, mode int32) {
	e.hasRun = true
	e.mode = sdkpb.Mode(mode)
}
