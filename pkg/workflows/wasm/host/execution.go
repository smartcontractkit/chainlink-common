package host

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/engine"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
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
	timeFetcher          *timeFetcher
	baseTime             *time.Time
	hasRun               bool
	mode                 sdkpb.Mode
	donSeed              int64
	nodeSeed             int64
	donLogCount          uint32
	nodeLogCount         uint32
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

func (e *execution[T]) log(caller engine.MemoryAccessor, ptr int32, ptrlen int32) {
	switch e.mode {
	case sdkpb.Mode_MODE_DON:
		e.donLogCount++
		if e.donLogCount == e.module.cfg.MaxLogCountDONMode {
			e.module.cfg.Logger.Warnf("max log count for don mode reached: %d - all subsequent logs will be dropped", e.donLogCount)
		}
		if e.donLogCount > e.module.cfg.MaxLogCountDONMode {
			// silently drop to avoid spamming logs
			return
		}
	case sdkpb.Mode_MODE_NODE:
		e.nodeLogCount++
		if e.nodeLogCount == e.module.cfg.MaxLogCountNodeMode {
			e.module.cfg.Logger.Warnf("max log count for node mode reached: %d - all subsequent logs will be dropped", e.nodeLogCount)
		}
		if e.nodeLogCount > e.module.cfg.MaxLogCountNodeMode {
			// silently drop to avoid spamming logs
			return
		}
	default:
		// unexpected / malicious
		return
	}

	if ptrlen > int32(e.module.cfg.MaxLogLenBytes) {
		e.module.cfg.Logger.Warnf("log message too long: %d - dropping", ptrlen)
		return
	}

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

func (e *execution[T]) switchModes(_ engine.MemoryAccessor, mode int32) {
	e.hasRun = true
	e.mode = sdkpb.Mode(mode)
}

// clockTimeGet is the default time.Now() which is also called by Go many times.
// This implementation uses Node Mode to not have to wait for OCR rounds.
func (e *execution[T]) clockTimeGet(caller engine.MemoryAccessor, id int32, precision int64, resultTimestamp int32) int32 {
	donTime, err := e.timeFetcher.GetTime(sdkpb.Mode_MODE_NODE)
	if err != nil {
		return ErrnoInval
	}

	if e.baseTime == nil {
		// baseTime must be before the first poll or Go panics
		t := donTime.Add(-time.Nanosecond)
		e.baseTime = &t
	}

	var val int64
	switch id {
	case clockIDMonotonic:
		val = donTime.Sub(*e.baseTime).Nanoseconds()
	case clockIDRealtime:
		val = donTime.UnixNano()
	default:
		return ErrnoInval
	}

	uint64Size := int32(8)
	trg := make([]byte, uint64Size)
	binary.LittleEndian.PutUint64(trg, uint64(val))
	wasmWrite(caller, trg, resultTimestamp, uint64Size)
	return ErrnoSuccess
}

// now is used by rawsdk for Workflows and should be called instead of Go's time.Now().
func (e *execution[T]) now(caller engine.MemoryAccessor, resultTimestamp int32) int32 {
	donTime, err := e.timeFetcher.GetTime(e.mode)
	if err != nil {
		return ErrnoInval
	}

	val := donTime.UnixNano()
	uint64Size := int32(8)
	trg := make([]byte, uint64Size)
	binary.LittleEndian.PutUint64(trg, uint64(val))
	wasmWrite(caller, trg, resultTimestamp, uint64Size)
	return ErrnoSuccess
}

// Loosely based off the implementation here:
// https://github.com/tetratelabs/wazero/blob/main/imports/wasi_snapshot_preview1/poll.go#L52
// For an overview of the spec, including the datatypes being referred to, see:
// https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md
// This implementation only responds to clock events, not to file descriptor notifications.
// It sleeps based on the largest timeout
func (e *execution[T]) pollOneoff(caller engine.MemoryAccessor, subscriptionptr int32, eventsptr int32, nsubscriptions int32, resultNevents int32) int32 {
	if nsubscriptions == 0 {
		return ErrnoInval
	}

	subs, err := wasmRead(caller, subscriptionptr, nsubscriptions*subscriptionLen)
	if err != nil {
		return ErrnoFault
	}

	events := make([]byte, nsubscriptions*eventsLen)
	timeout := time.Duration(0)

	for i := range nsubscriptions {
		inOffset := i * subscriptionLen
		userData := subs[inOffset : inOffset+8]
		eventType := subs[inOffset+8]
		argBuf := subs[inOffset+8+8:]

		slot, err := getSlot(events, i)
		if err != nil {
			return ErrnoFault
		}

		switch eventType {
		case eventTypeClock:
			newTimeout := binary.LittleEndian.Uint64(argBuf[8:16])
			flag := binary.LittleEndian.Uint16(argBuf[24:32])

			var errno Errno
			switch flag {
			case 0: // relative
				errno = ErrnoSuccess
				if timeout < time.Duration(newTimeout) {
					timeout = time.Duration(newTimeout)
				}
			default:
				errno = ErrnoNotsup
			}
			writeEvent(slot, userData, errno, eventTypeClock)

		case eventTypeFDRead, eventTypeFDWrite:
			writeEvent(slot, userData, ErrnoBadf, int(eventType))

		default:
			writeEvent(slot, userData, ErrnoInval, int(eventType))
		}
	}

	if timeout > 0 {
		select {
		case <-time.After(timeout):
		case <-e.ctx.Done():
			// If context was cancelled, there will be a trap from the engine
			// which will halt execution, therefore the return value isn't read
			return 0
		}
	}

	uint32Size := int32(4)
	rne := make([]byte, uint32Size)
	binary.LittleEndian.PutUint32(rne, uint32(nsubscriptions))

	if wasmWrite(caller, rne, resultNevents, uint32Size) == -1 {
		return ErrnoFault
	}
	if wasmWrite(caller, events, eventsptr, nsubscriptions*eventsLen) == -1 {
		return ErrnoFault
	}

	return ErrnoSuccess
}
