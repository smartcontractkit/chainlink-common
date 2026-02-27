package beholder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/timeutil"
)

// ChipIngressBatchEmitter buffers events per (domain, entity) and flushes them
// via chipingress.Client.PublishBatch on a periodic interval.
// It satisfies the Emitter interface so it can be used as a drop-in replacement
// for ChipIngressEmitter.
type ChipIngressBatchEmitter struct {
	services.Service
	eng *services.Engine

	client chipingress.Client

	workers      map[string]*chipIngressEmitterWorker
	workersMutex sync.RWMutex

	bufferSize   uint
	maxBatchSize uint
	sendInterval time.Duration
	sendTimeout  time.Duration
}

// NewChipIngressBatchEmitter creates a batch emitter backed by the given chipingress client.
// Call Start() to begin health monitoring, and Close() to stop all workers.
func NewChipIngressBatchEmitter(client chipingress.Client, lggr logger.Logger, cfg Config) (*ChipIngressBatchEmitter, error) {
	if client == nil {
		return nil, fmt.Errorf("chip ingress client is nil")
	}

	bufferSize := cfg.ChipIngressBufferSize
	if bufferSize == 0 {
		bufferSize = 100
	}
	maxBatchSize := cfg.ChipIngressMaxBatchSize
	if maxBatchSize == 0 {
		maxBatchSize = 50
	}
	sendInterval := cfg.ChipIngressSendInterval
	if sendInterval == 0 {
		sendInterval = 500 * time.Millisecond
	}
	sendTimeout := cfg.ChipIngressSendTimeout
	if sendTimeout == 0 {
		sendTimeout = 10 * time.Second
	}

	e := &ChipIngressBatchEmitter{
		client:       client,
		workers:      make(map[string]*chipIngressEmitterWorker),
		bufferSize:   bufferSize,
		maxBatchSize: maxBatchSize,
		sendInterval: sendInterval,
		sendTimeout:  sendTimeout,
	}

	e.Service, e.eng = services.Config{
		Name:  "ChipIngressBatchEmitter",
		Start: e.start,
	}.NewServiceEngine(lggr)

	return e, nil
}

func (e *ChipIngressBatchEmitter) start(_ context.Context) error {
	return nil
}

// Emit extracts (domain, entity) from the attributes, routes the event to the
// appropriate per-(domain, entity) worker, and returns immediately.
// If the worker's buffer is full, the event is dropped and a warning is logged.
func (e *ChipIngressBatchEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	domain, entity, err := ExtractSourceAndType(attrKVs...)
	if err != nil {
		return err
	}

	attributes := newAttributes(attrKVs...)

	worker := e.findOrCreateWorker(domain, entity)

	payload := emitterPayload{
		body:       body,
		attributes: attributes,
		domain:     domain,
		entity:     entity,
	}

	select {
	case worker.ch <- payload:
		worker.dropCount.Store(0)
	case <-ctx.Done():
		return ctx.Err()
	default:
		worker.logBufferFullWithExpBackoff(payload)
	}

	return nil
}

// findOrCreateWorker returns the worker for the given (domain, entity) pair,
// creating one with a new buffered channel and GoTick flush loop if it doesn't exist.
func (e *ChipIngressBatchEmitter) findOrCreateWorker(domain, entity string) *chipIngressEmitterWorker {
	workerKey := fmt.Sprintf("%s_%s", domain, entity)

	e.workersMutex.RLock()
	worker, found := e.workers[workerKey]
	e.workersMutex.RUnlock()

	if found {
		return worker
	}

	e.workersMutex.Lock()
	defer e.workersMutex.Unlock()

	// Double-check after acquiring write lock
	if worker, found = e.workers[workerKey]; found {
		return worker
	}

	worker = newChipIngressEmitterWorker(
		e.client,
		make(chan emitterPayload, e.bufferSize),
		domain,
		entity,
		e.maxBatchSize,
		e.sendTimeout,
		e.eng,
	)

	e.eng.GoTick(timeutil.NewTicker(func() time.Duration {
		return e.sendInterval
	}), worker.Send)

	e.workers[workerKey] = worker
	return worker
}
