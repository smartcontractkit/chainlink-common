package host

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v28"
	"google.golang.org/protobuf/proto"

	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	wfpb "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

// Here we need to distinguish between variables that should survive a replay and
// variables that need to be re-initialized every replay.
type execution[T any] struct {
	fetchRequestsCounter int
	response             T
	ctx                  context.Context
	capabilityResponses  map[int32]*asyncResponse[sdkpb.CapabilityRequest, sdkpb.CapabilityResponse]
	secretsResponses     map[int32]<-chan *secretsResponse
	pendingCallsLimiter  limits.ResourcePoolLimiter[int]
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
	awaiting             []int32
	// peakMemoryBytes is the largest linear memory observed across (re)starts.
	// It is populated by callWasm and read by Execute to emit the memory metric.
	peakMemoryBytes int64
	// suspendOnAwait gates the suspend/resume behaviour. When false, the
	// execution behaves as it did before suspension was introduced:
	// awaitCapabilities blocks until each response is available and callCapAsync
	// always dispatches a fresh call. When true, awaitCapabilities returns
	// errSuspendExecution while responses are pending and callCapAsync replays
	// recorded calls instead of re-dispatching them.
	suspendOnAwait bool
}

type asyncResponse[I, O any] struct {
	mu   sync.Mutex
	ch   <-chan *O
	resp *O
	req  *I
}

func (a *asyncResponse[I, O]) wait(ctx context.Context) (*O, error) {
	if resp := a.getResp(ctx); resp != nil {
		return resp, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case o := <-a.ch:
		a.mu.Lock()
		defer a.mu.Unlock()
		a.resp = o
		return o, nil
	}
}

func (a *asyncResponse[I, O]) getResp(ctx context.Context) *O {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.resp
}

// callCapAsync async calls a capability by placing execution results onto a
// channel and storing each channel with a unique identifier for future
// retrieval on await.
func (e *execution[T]) callCapAsync(ctx context.Context, req *sdkpb.CapabilityRequest) error {
	if e.suspendOnAwait {
		// check if there is already an item in the capabilityResponses for a given callback id.
		// if there is -> integrity check (matching requests); return without firing goroutine.
		// else -> legacy path.
		if asyncResponse, ok := e.capabilityResponses[req.CallbackId]; ok {
			// there is already an item for this callback id; we must therefore be replaying
			// perform an integrity check to enforce determinism and return early.
			if !proto.Equal(asyncResponse.req, req) {
				return errors.New("non-determinism error")
			}
			return nil
		}
	}
	// Acquire a slot from the pool limiter to bound concurrency.
	free, err := e.pendingCallsLimiter.Wait(ctx, 1)
	if err != nil {
		return err
	}

	ch := make(chan *sdkpb.CapabilityResponse, 1)
	e.lock.Lock()
	defer e.lock.Unlock()
	e.capabilityResponses[req.CallbackId] = &asyncResponse[
		sdkpb.CapabilityRequest,
		sdkpb.CapabilityResponse,
	]{
		ch:  ch,
		req: req,
	}

	go func() {
		defer free()

		resp, err := e.executor.CallCapability(ctx, req)

		if err != nil {
			errString := err.Error()

			var caperror caperrors.Error
			if errors.As(err, &caperror) {
				errString = caperror.SerializeToString()
			}
			resp = &sdkpb.CapabilityResponse{
				Response: &sdkpb.CapabilityResponse_Error{
					Error: errString,
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

var errSuspendExecution = errors.New("__SUSPEND_EXECUTION__")

func (e *execution[T]) awaitCapabilities(ctx context.Context, acr *sdkpb.AwaitCapabilitiesRequest) (*sdkpb.AwaitCapabilitiesResponse, error) {
	responses := make(map[int32]*sdkpb.CapabilityResponse, len(acr.Ids))

	e.lock.Lock()
	defer e.lock.Unlock()

	if !e.suspendOnAwait {
		// Legacy behaviour: block until each requested response is available,
		// then consume and remove it from the store.
		for _, callId := range acr.Ids {
			ar, ok := e.capabilityResponses[callId]
			if !ok {
				return nil, fmt.Errorf("failed to get call from store : %d", callId)
			}

			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("failed to wait for capability response %d : %w", callId, ctx.Err())
			case resp := <-ar.ch:
				responses[callId] = resp
			}

			delete(e.capabilityResponses, callId)
		}

		return &sdkpb.AwaitCapabilitiesResponse{
			Responses: responses,
		}, nil
	}

	// for all ids, check whether we have a response
	// if yes: return the response
	// if no: return suspend execution error, record the ids for which we are still waiting
	// 	to be picked up by the runWasm routine
	responsesForAll := true
	for _, callId := range acr.Ids {
		ar, ok := e.capabilityResponses[callId]
		if !ok {
			return nil, fmt.Errorf("failed to get call from store : %d", callId)
		}

		resp := ar.getResp(ctx)
		if resp == nil {
			responsesForAll = false
			break
		}

		responses[callId] = resp
	}

	if !responsesForAll {
		e.awaiting = acr.Ids
		return nil, errSuspendExecution
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
	// Acquire a slot from the pool limiter to bound concurrency.
	free, err := e.pendingCallsLimiter.Wait(ctx, 1)
	if err != nil {
		return err
	}

	ch := make(chan *secretsResponse, 1)
	e.lock.Lock()
	defer e.lock.Unlock()
	e.secretsResponses[req.CallbackId] = ch

	go func() {
		defer free()

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

func (e *execution[T]) emitMetric(caller *wasmtime.Caller, ptr int32, ptrlen int32) int32 {
	if err := e.module.cfg.EnableUserMetricsLimiter.AllowErr(e.ctx); err != nil {
		return -1
	}

	if ptrlen <= 0 {
		return -1
	}

	if err := e.module.cfg.MaxUserMetricPayloadLimiter.Check(e.ctx, config.Size(ptrlen)); err != nil {
		e.module.cfg.Logger.Warnf("metric payload too large: %d bytes - dropping: %s", ptrlen, err)
		return -1
	}

	b, err := wasmRead(caller, ptr, ptrlen)
	if err != nil {
		e.module.cfg.Logger.Errorf("error reading metric payload: %s", err)
		return -1
	}

	metric := &wfpb.WorkflowUserMetric{}
	if err := proto.Unmarshal(b, metric); err != nil {
		e.module.cfg.Logger.Errorf("error unmarshaling metric: %s", err)
		return -1
	}

	if metric.Name == "" {
		e.module.cfg.Logger.Warnf("metric name cannot be empty - dropping")
		return -1
	}

	if err := e.module.cfg.MaxUserMetricNameLengthLimiter.Check(e.ctx, len(metric.Name)); err != nil {
		e.module.cfg.Logger.Warnf("metric name too long: %d chars - dropping: %s", len(metric.Name), err)
		return -1
	}

	if err := e.module.cfg.MaxUserMetricLabelsPerMetricLimiter.Check(e.ctx, len(metric.Labels)); err != nil {
		e.module.cfg.Logger.Warnf("too many labels on metric %q: %d - dropping: %s", metric.Name, len(metric.Labels), err)
		return -1
	}

	for k, v := range metric.Labels {
		if err := e.module.cfg.MaxUserMetricLabelValueLengthLimiter.Check(e.ctx, len(v)); err != nil {
			e.module.cfg.Logger.Warnf("label value too long for key %q on metric %q: %d chars - dropping: %s", k, metric.Name, len(v), err)
			return -1
		}
	}

	if err := e.executor.EmitUserMetric(e.ctx, metric); err != nil {
		e.module.cfg.Logger.Errorf("error emitting user metric: %s", err)
		return -1
	}

	return 0
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

// clockTimeGet is the default time.Now() which is also called by Go many times.
// This implementation uses Node Mode to not have to wait for OCR rounds.
func (e *execution[T]) clockTimeGet(caller *wasmtime.Caller, id int32, precision int64, resultTimestamp int32) int32 {
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
func (e *execution[T]) now(caller *wasmtime.Caller, resultTimestamp int32) int32 {
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
func (e *execution[T]) pollOneoff(caller *wasmtime.Caller, subscriptionptr int32, eventsptr int32, nsubscriptions int32, resultNevents int32) int32 {
	if nsubscriptions <= 0 || nsubscriptions > max(math.MaxInt32/subscriptionLen, math.MaxInt32/eventsLen) {
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
