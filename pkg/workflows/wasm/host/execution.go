package host

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v28"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type execution[T any] struct {
	fetchRequestsCounter int
	response             T
	ctx                  context.Context
	capabilityResponses  map[int32]<-chan *sdkpb.CapabilityResponse
	secretsResponses     map[int32]<-chan *secretsResponse
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

type secretsResponse struct {
	responses []*sdkpb.SecretResponse
	err       error
}

func (e *execution[T]) getSecretsAsync(ctx context.Context, req *sdkpb.GetSecretsRequest) error {
	ch := make(chan *secretsResponse, 1)
	e.lock.Lock()
	defer e.lock.Unlock()
	e.secretsResponses[req.CallbackId] = ch

	go func() {
		resp, err := e.executor.GetSecrets(ctx, req)
		sr := &secretsResponse{responses: resp, err: err}

		select {
		case <-ctx.Done():
		case ch <- sr:
		}
	}()

	return nil
}

func (e *execution[T]) awaitSecrets(ctx context.Context, acr *sdkpb.AwaitSecretsRequest) (*sdkpb.AwaitSecretsResponse, error) {
	responses := make(map[int32]*sdkpb.SecretResponses, len(acr.Ids))

	e.lock.Lock()
	defer e.lock.Unlock()
	for _, callId := range acr.Ids {
		ch, ok := e.secretsResponses[callId]
		if !ok {
			return nil, fmt.Errorf("failed to get call from store : %d", callId)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("failed to wait for capability response %d : %w", callId, ctx.Err())
		case resp := <-ch:
			if resp.err != nil {
				return nil, fmt.Errorf("failed to get secrets for call %d: %w", callId, resp.err)
			}

			responses[callId] = &sdkpb.SecretResponses{Responses: resp.responses}
		}

		delete(e.secretsResponses, callId)
	}

	return &sdkpb.AwaitSecretsResponse{
		Responses: responses,
	}, nil
}

func (e *execution[T]) log(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
	b, innerErr := wasmRead(caller, ptr, ptrlen)
	if innerErr != nil {
		e.module.cfg.Logger.Errorf("error calling log: %s", innerErr)
		return
	}
	innerErr = e.executor.EmitUserLog(string(b))
	if innerErr != nil {
		e.module.cfg.Logger.Errorf("error emitting user log: %s", innerErr)
		return
	}
}

func (e *execution[T]) getSeed(mode int32) int64 {
	switch sdkpb.Mode(mode) {
	case sdkpb.Mode_MODE_DON:
		return e.donSeed
	case sdkpb.Mode_MODE_NODE:
		return e.nodeSeed
	}

	return -1
}

func (e *execution[T]) switchModes(_ *wasmtime.Caller, mode int32) {
	e.hasRun = true
	e.mode = sdkpb.Mode(mode)
}

func (e *execution[T]) getTime(caller *wasmtime.Caller, resultTimestamp int32) int32 {
	var donTime time.Time
	switch e.mode {
	case sdkpb.Mode_MODE_DON:
		var err error
		donTime, err = e.executor.GetDONTime(context.TODO())
		if err != nil {
			return ErrnoInval
		}
	case sdkpb.Mode_MODE_NODE:
		donTime = e.executor.GetNodeTime()
	default:
		return ErrnoInval
	}

	uint64Size := int32(8)
	trg := make([]byte, uint64Size)
	binary.LittleEndian.PutUint64(trg, uint64(donTime.UnixNano()))
	wasmWrite(caller, trg, resultTimestamp, uint64Size)
	return ErrnoSuccess
}
