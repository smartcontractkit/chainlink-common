package durableemitter

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// BatchEmitter is the transport interface DurableEmitter delegates to for
// batched delivery of CloudEvents to Chip Ingress.
//
// *batch.Client from pkg/chipingress/batch satisfies this interface and
// handles seqnum stamping, gRPC size splitting, concurrency limiting, and
// graceful shutdown with a configurable timeout.
//
// The callback passed to QueueMessage is invoked once after the batch
// containing the event is sent. A nil error means the RPC succeeded; a
// non-nil error means the batch was dropped — the event remains in the DB
// and the retransmit loop will retry it.
type BatchEmitter interface {
	// QueueMessage enqueues a single CloudEvent for batched delivery.
	// Returns an error only if the internal buffer is full or the client
	// has been stopped. Callers must treat a non-nil return as a
	// drop (the event is still persisted; retransmit will retry).
	QueueMessage(event *chipingress.CloudEventPb, callback func(error)) error
	// Start begins background processing. Must be called before QueueMessage.
	Start(ctx context.Context)
	// Stop flushes any queued events, waits for all in-flight network calls
	// and callbacks to complete, then closes the underlying transport.
	Stop()
}

// Config configures the DurableEmitter behaviour.
type Config struct {
	// RetransmitInterval controls how often the retransmit loop ticks.
	RetransmitInterval time.Duration
	// RetransmitAfter is the minimum age of an event before the retransmit
	// loop considers it. This gives the batch publish path time to succeed.
	RetransmitAfter time.Duration
	// RetransmitBatchSize caps how many pending rows are listed per retransmit tick.
	RetransmitBatchSize int
	// ExpiryInterval controls how often the expiry loop ticks.
	ExpiryInterval time.Duration
	// EventTTL is the maximum age of an event before it is expired.
	EventTTL time.Duration
	// PublishTimeout is the deadline for DB operations in delivery callbacks
	// (BatchDelete). The actual gRPC publish timeout is configured on
	// the BatchEmitter (batch.Client) directly.
	PublishTimeout time.Duration
	// DisablePruning disables the background expiry (DeleteExpired) loop.
	// Events then remain in the DB even after they age past EventTTL. Useful for
	// post-test analysis of created_at timestamps.
	DisablePruning bool
	// Hooks is optional instrumentation (load tests, profiling). Nil fields are skipped.
	// Callbacks may run from many goroutines; implementations must be thread-safe.
	Hooks *Hooks
	// Metrics enables OpenTelemetry instruments (queue, publish, store, optional process stats).
	// When non-nil, a meter must be supplied to NewDurableEmitter; nil disables instrumentation.
	Metrics *DurableEmitterMetricsConfig
	// InsertBatchSize enables write coalescing when > 0 and the store implements
	// BatchInserter. Multiple concurrent Emit() calls are grouped into a single
	// multi-row INSERT, dramatically reducing per-event transaction overhead.
	// Each coalescer worker collects up to InsertBatchSize payloads before flushing.
	InsertBatchSize int
	// InsertBatchFlushInterval is the linger time after the first payload arrives
	// in a coalescing batch. Zero defaults to 2ms.
	InsertBatchFlushInterval time.Duration
	// InsertBatchWorkers is the number of concurrent batch-insert goroutines.
	// Zero defaults to 4.
	InsertBatchWorkers int
	// DeleteBatchSize enables delete coalescing when > 0. Instead of issuing one
	// DELETE per delivered event from the delivery callback, ids are funneled to
	// background workers that collapse many ids into a single BatchDelete,
	// drastically reducing per-event DELETE/connection churn.
	// Each worker collects up to DeleteBatchSize ids before flushing.
	DeleteBatchSize int
	// DeleteBatchFlushInterval is the linger time after the first id arrives in a
	// coalescing batch before it is flushed. Zero defaults to 100ms.
	DeleteBatchFlushInterval time.Duration
	// DeleteBatchWorkers is the number of concurrent batch-delete goroutines.
	// Zero defaults to 2.
	DeleteBatchWorkers int
}

// Hooks records delivery latency to locate pipeline bottlenecks.
type Hooks struct {
	// OnEmitInsert is called after each store.Insert in Emit (the DB write that
	// blocks the caller). elapsed covers only the INSERT; err is nil on success.
	OnEmitInsert func(elapsed time.Duration, err error)
	// OnBatchPublish is called from the delivery callback after each event's
	// batch is sent. elapsed is measured from QueueMessage call to callback
	// invocation; batchSize is always 1 (one callback per event); err is nil
	// on success.
	OnBatchPublish func(elapsed time.Duration, batchSize int, err error)
	// OnBatchDelete is called after BatchDelete following a successful delivery.
	OnBatchDelete func(elapsed time.Duration, count int)
}

func DefaultConfig() Config {
	return Config{
		RetransmitInterval:       5 * time.Second,
		RetransmitAfter:          10 * time.Second,
		RetransmitBatchSize:      100,
		ExpiryInterval:           1 * time.Minute,
		EventTTL:                 1 * time.Hour,
		PublishTimeout:           5 * time.Second,
		InsertBatchFlushInterval: 500 * time.Millisecond,
		InsertBatchSize:          100,
		DeleteBatchSize:          100,
		DeleteBatchFlushInterval: 500 * time.Millisecond,
		DeleteBatchWorkers:       2,
		// Metrics is opt-in: callers who want instrumentation must set this
		// and pass a metric.Meter to NewDurableEmitter.
		Metrics: nil,
	}
}

// DurableEmitter implements Emitter with persistence-backed delivery guarantees.
//
// Emit writes to a DurableEventStore then hands the event to the BatchEmitter
// for async delivery. The delivery callback from BatchEmitter marks the row
// delivered, which under the Postgres store deletes it outright
// (delete-on-delivery) rather than tombstoning + purging — this avoids an extra
// write and index churn per event. When the batch emitter buffer is full or the
// network is down, a retransmit loop lists stale pending rows and re-enqueues
// them through the same BatchEmitter (up to RetransmitBatchSize per tick).
//
// A separate expiry loop garbage-collects events older than EventTTL to bound
// table growth.

// insertRequest is a single Emit() caller waiting for a coalesced batch INSERT.
type insertRequest struct {
	payload []byte
	result  chan insertResult
}

type insertResult struct {
	id  int64
	err error
}

type DurableEmitter struct {
	services.Service
	eng *services.Engine

	store        DurableEventStore
	batchEmitter BatchEmitter
	// retransmitEnabled controls whether this instance runs the retransmit and
	// cleanup loops. Should be set to false when initialized inside LOOP plugins.
	retransmitEnabled bool
	cfg               Config

	metrics *durableEmitterMetrics

	// batchInserter is non-nil when the store supports multi-row INSERTs
	// and InsertBatchSize > 0.
	batchInserter BatchInserter
	// insertCh buffers payloads for the write coalescer. Nil when batch
	// inserting is disabled.
	insertCh chan *insertRequest
	// insertShutdown stops new coalesced inserts; insertInFlight counts Emit
	// callers inside the coalesced path so stop can close(insertCh) after wait
	insertShutdown atomic.Bool
	insertInFlight atomic.Int32

	// deleteCh funnels delivered event ids to the delete coalescer. Nil when delete
	// coalescing is disabled (DeleteBatchSize <= 0). Its only producers are the
	// delivery callbacks, so it is closed during shutdown only after they have
	// quiesced (see stop). deleteWg tracks
	// the deleteBatchLoop workers, which must outlive d.wg so callbacks running
	// during the batch emitter's final flush can still enqueue deletes.
	deleteCh chan int64
	deleteWg sync.WaitGroup

	// stopCh signals background loops to exit.
	stopCh services.StopChan
	wg     sync.WaitGroup

	// retransmit paging cursor. The retransmit loop pages through the pending
	// backlog in (created_at, id) order, advancing the cursor each tick and
	// wrapping to zero at the end, so a persistently-failing ("poison") event
	// can't monopolise the head of the list and starve everything behind it.
	// Accessed only from the single retransmit-loop goroutine — no locking.
	retransmitCursorTs time.Time
	retransmitCursorID int64
}

// Compile-time assertion that *DurableEmitter exposes the canonical emit and
// close methods expected of an emitter.
var _ interface {
	Emit(ctx context.Context, body []byte, attrKVs ...any) error
	io.Closer
} = (*DurableEmitter)(nil)

// NewDurableEmitter constructs a DurableEmitter as a service.
//
// batchEmitter is the transport layer (typically *batch.Client from
// pkg/chipingress/batch) responsible for batched gRPC delivery, seqnum
// stamping, size splitting, and concurrency limiting.
//
// On a batch delivery failure the event is left in the DB and re-delivered by
// the DB-backed retransmit loop.
func NewDurableEmitter(
	store DurableEventStore,
	batchEmitter BatchEmitter,
	retransmitEnabled bool,
	cfg Config,
	lggr logger.Logger,
	meter metric.Meter,
) (*DurableEmitter, error) {
	if store == nil {
		return nil, errors.New("durable event store is nil")
	}
	if batchEmitter == nil {
		return nil, errors.New("batch emitter is nil")
	}
	if lggr == nil {
		return nil, errors.New("logger is nil")
	}
	var m *durableEmitterMetrics
	if cfg.Metrics != nil {
		if meter == nil {
			return nil, errors.New("durable emitter metrics enabled but meter is nil")
		}
		var err error
		m, err = newDurableEmitterMetrics(meter, "durable_emitter")
		if err != nil {
			return nil, fmt.Errorf("durable emitter metrics: %w", err)
		}
		store = newMetricsInstrumentedStore(store, m)
	}
	d := &DurableEmitter{
		store:             store,
		batchEmitter:      batchEmitter,
		retransmitEnabled: retransmitEnabled,
		cfg:               cfg,
		metrics:           m,
		stopCh:            make(chan struct{}),
	}
	d.Service, d.eng = services.Config{
		Name:  "DurableEmitter",
		Start: d.start,
		Close: d.stop,
	}.NewServiceEngine(lggr)

	if cfg.InsertBatchSize > 0 {
		if bi, ok := store.(BatchInserter); ok {
			d.batchInserter = bi
			chanSize := cfg.InsertBatchSize * 200
			if chanSize < 10_000 {
				chanSize = 10_000
			}
			d.insertCh = make(chan *insertRequest, chanSize)
			d.eng.Infow("DurableEmitter: write coalescing enabled",
				"insertBatchSize", cfg.InsertBatchSize,
				"insertBatchWorkers", cfg.InsertBatchWorkers,
				"insertBatchFlushInterval", cfg.InsertBatchFlushInterval)
		}
	}

	if cfg.DeleteBatchSize > 0 {
		chanSize := cfg.DeleteBatchSize * 200
		if chanSize < 10_000 {
			chanSize = 10_000
		}
		d.deleteCh = make(chan int64, chanSize)
		d.eng.Infow("DurableEmitter: delete coalescing enabled",
			"deleteBatchSize", cfg.DeleteBatchSize,
			"deleteBatchWorkers", cfg.DeleteBatchWorkers,
			"deleteBatchFlushInterval", cfg.DeleteBatchFlushInterval)
	}
	return d, nil
}

// start launches the retransmit, expiry, purge, and insert-coalescing background
// loops, then starts the batch emitter transport. It is invoked by the
// services.Engine when the embedded Service is started.
func (d *DurableEmitter) start(ctx context.Context) error {
	d.batchEmitter.Start(ctx)

	insertWorkers := d.cfg.InsertBatchWorkers
	if insertWorkers <= 0 {
		insertWorkers = 4
	}
	if d.insertCh != nil {
		for i := 0; i < insertWorkers; i++ {
			d.wg.Go(d.insertBatchLoop)
		}
	}

	if d.deleteCh != nil {
		deleteWorkers := d.cfg.DeleteBatchWorkers
		if deleteWorkers <= 0 {
			deleteWorkers = 2
		}
		// deleteWg (not d.wg) so the workers keep draining while the batch
		// emitter flushes its final callbacks during stop().
		for i := 0; i < deleteWorkers; i++ {
			d.deleteWg.Go(d.deleteBatchLoop)
		}
	}

	if d.retransmitEnabled {
		d.wg.Go(d.retransmitLoop)
		if !d.cfg.DisablePruning {
			d.wg.Go(d.expiryLoop)
		}
	}
	if d.metrics != nil && d.cfg.Metrics != nil {
		d.wg.Go(d.metricsLoop)
	}
	return nil
}

// Emit persists the event then hands it to the BatchEmitter for async delivery.
// Returns nil once the insert is accepted (or the coalesced insert path
// completes successfully). Returns an error when the service is not in the
// Started state (e.g. before Start or after Close).
func (d *DurableEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	return d.eng.IfStarted(func() error {
		tEmitTotal := time.Now()
		defer func() {
			if d.metrics != nil {
				d.metrics.emitTotalDuration.Record(ctx, time.Since(tEmitTotal).Seconds())
			}
		}()
		emitFail := func() {
			if d.metrics != nil {
				d.metrics.emitFail.Add(ctx, 1)
			}
		}
		sourceDomain, entityType, err := extractSourceAndType(attrKVs...)
		if err != nil {
			emitFail()
			return err
		}

		attrs := parseAttrs(attrKVs...)
		ensureIdempotencyKey(attrs, sourceDomain, entityType, body)

		event, err := chipingress.NewEvent(sourceDomain, entityType, body, attrs)
		if err != nil {
			emitFail()
			return err
		}

		event.SetExtension("emitter", "DurableEmitter")

		eventPb, err := chipingress.EventToProto(event)
		if err != nil {
			emitFail()
			return fmt.Errorf("failed to convert event to proto: %w", err)
		}

		payload, err := proto.Marshal(eventPb)
		if err != nil {
			emitFail()
			return fmt.Errorf("failed to marshal event proto: %w", err)
		}

		var id int64
		var insElapsed time.Duration

		if d.insertCh != nil {
			// Write coalescing: send payload to the batch insert loop and block
			// until the multi-row INSERT completes.
			req := &insertRequest{
				payload: payload,
				result:  make(chan insertResult, 1),
			}
			var res insertResult
			var cerr error
			func() {
				d.insertInFlight.Add(1)
				defer d.insertInFlight.Add(-1)
				if d.insertShutdown.Load() {
					cerr = errors.New("durable emitter closed")
					return
				}
				tIns := time.Now()
				select {
				case d.insertCh <- req:
				case <-ctx.Done():
					cerr = ctx.Err()
					return
				}
				res = <-req.result
				insElapsed = time.Since(tIns)
			}()
			if cerr != nil {
				if errors.Is(cerr, context.Canceled) || errors.Is(cerr, context.DeadlineExceeded) {
					emitFail()
				}
				return cerr
			}
			if h := d.cfg.Hooks; h != nil && h.OnEmitInsert != nil {
				h.OnEmitInsert(insElapsed, res.err)
			}
			if d.metrics != nil {
				d.metrics.recordEmitDuration(ctx, insElapsed, res.err)
				if res.err != nil {
					d.metrics.emitFail.Add(ctx, 1)
				} else {
					d.metrics.emitSuccess.Add(ctx, 1)
				}
			}
			if res.err != nil {
				return fmt.Errorf("failed to persist event: %w", res.err)
			}
			id = res.id
		} else {
			tIns := time.Now()
			id, err = d.store.Insert(ctx, payload)
			insElapsed = time.Since(tIns)
			if h := d.cfg.Hooks; h != nil && h.OnEmitInsert != nil {
				h.OnEmitInsert(insElapsed, err)
			}
			if d.metrics != nil {
				d.metrics.recordEmitDuration(ctx, insElapsed, err)
				if err != nil {
					d.metrics.emitFail.Add(ctx, 1)
				} else {
					d.metrics.emitSuccess.Add(ctx, 1)
				}
			}
			if err != nil {
				return fmt.Errorf("failed to persist event: %w", err)
			}
		}

		// Hand off to the batch emitter. The callback fires once the batch
		// containing this event is sent (success or failure).
		t0Publish := time.Now()
		if qErr := d.batchEmitter.QueueMessage(eventPb, d.deliveryCallback(id, eventPb, t0Publish, publishPhaseBatch)); qErr != nil {
			d.eng.Warnw("DurableEmitter: batch emitter buffer full, relying on retransmit", "id", id)
			if d.metrics != nil {
				d.metrics.batchEnqueueBufferFull.Add(ctx, 1,
					metric.WithAttributes(attribute.String("phase", publishPhaseBatch.String())))
			}
		}
		return nil
	})
}

// deliveryCallback returns the function passed to BatchEmitter.QueueMessage.
// On success, it deletes the delivered event. On failure, it leaves the event
// in the DB for the retransmit loop.
func (d *DurableEmitter) deliveryCallback(id int64, eventPb *chipingress.CloudEventPb, t0Publish time.Time, phase publishPhase) func(error) {
	return func(sendErr error) {
		publishElapsed := time.Since(t0Publish)

		if h := d.cfg.Hooks; h != nil && h.OnBatchPublish != nil {
			h.OnBatchPublish(publishElapsed, 1, sendErr)
		}

		cbCtx, cbCancel := d.stopCh.NewCtx()
		defer cbCancel()

		if d.metrics != nil {
			d.metrics.recordPublish(cbCtx, publishElapsed, phase, sendErr)
			d.metrics.recordPublishBatchEvent(cbCtx, phase, sendErr)
		}

		if sendErr != nil {
			d.eng.Warnw("DurableEmitter: failed to deliver event. Relying on retransmit.", "eventID", eventPb.Id, "err", sendErr)
			return
		}

		// When delete coalescing is enabled the id is handed to the batch-delete
		// workers (one DELETE for many ids); otherwise delete it inline.
		if d.enqueueDelete(id) {
			return
		}

		tDelete := time.Now()
		deleted, deleteErr := d.store.BatchDelete(cbCtx, []int64{id})
		deleteElapsed := time.Since(tDelete)

		if h := d.cfg.Hooks; h != nil && h.OnBatchDelete != nil {
			h.OnBatchDelete(deleteElapsed, int(deleted))
		}
		if deleteErr != nil {
			d.eng.Errorw("failed to delete event from persistence", "id", id, "error", deleteErr)
			return
		}
		if d.metrics != nil {
			d.metrics.deliverComplete.Add(cbCtx, deleted)
		}
	}
}

// stop signals background loops to stop and waits for them to finish, then
// stops the batch emitter (which flushes any queued events and waits for all
// in-flight callbacks). It is invoked by the services.Engine when the embedded
// Service is closed.
func (d *DurableEmitter) stop() error {
	if d.insertCh != nil {
		d.insertShutdown.Store(true)
		for d.insertInFlight.Load() > 0 {
			time.Sleep(time.Millisecond)
		}
		close(d.insertCh)
	}
	close(d.stopCh)
	d.wg.Wait()
	// Stop the batch emitter: flushes remaining queued events, waits for all
	// in-flight PublishBatch RPCs, and waits for all delivery callbacks.
	d.batchEmitter.Stop()
	// The delivery callbacks (drained by batchEmitter.Stop) are the only producers
	// of deleteCh, so it is now safe to close it and let the workers flush any buffered deletes.
	if d.deleteCh != nil {
		close(d.deleteCh)
		d.deleteWg.Wait()
	}
	return nil
}

// insertBatchLoop collects insertRequest items from insertCh and flushes them
// as multi-row INSERTs via BatchInserter.InsertBatch.
func (d *DurableEmitter) insertBatchLoop() {
	batchSize := d.cfg.InsertBatchSize
	linger := d.cfg.InsertBatchFlushInterval
	if linger <= 0 {
		linger = 100 * time.Millisecond
	}
	batch := make([]*insertRequest, 0, batchSize)

	for {
		batch = batch[:0]

		req, ok := <-d.insertCh
		if !ok {
			return
		}
		batch = append(batch, req)

		timer := time.NewTimer(linger)
	collecting:
		for len(batch) < batchSize {
			select {
			case req, ok := <-d.insertCh:
				if !ok {
					timer.Stop()
					break collecting
				}
				batch = append(batch, req)
			case <-timer.C:
				break collecting
			}
		}
		timer.Stop()

		payloads := make([][]byte, len(batch))
		for i, r := range batch {
			payloads[i] = r.payload
		}
		ctx, cancel := context.WithTimeout(context.Background(), d.cfg.PublishTimeout)
		ids, batchErr := d.batchInserter.InsertBatch(ctx, payloads)
		cancel()
		if batchErr == nil {
			d.eng.Debugw("DurableEmitter: coalesced insert flushed", "count", len(payloads))
		}
		for i, r := range batch {
			if batchErr != nil {
				r.result <- insertResult{err: batchErr}
			} else {
				r.result <- insertResult{id: ids[i]}
			}
		}
	}
}

// enqueueDelete hands a delivered event id to the delete coalescer. It returns
// false when coalescing is disabled so the caller can delete inline.
// Send is a blocking hand-off; back-pressure here slows the delivery callback
// rather than dropping a delete.
func (d *DurableEmitter) enqueueDelete(id int64) bool {
	if d.deleteCh == nil {
		return false
	}
	d.deleteCh <- id
	return true
}

// deleteBatchLoop collects delivered event ids from deleteCh and flushes them as
// batched BatchDelete DELETEs, collapsing many single-row DELETEs into one and
// decoupling the delete from the delivery-callback path. On channel close it
// flushes any partially-collected batch before returning.
func (d *DurableEmitter) deleteBatchLoop() {
	batchSize := d.cfg.DeleteBatchSize
	linger := d.cfg.DeleteBatchFlushInterval
	if linger <= 0 {
		linger = 100 * time.Millisecond
	}
	batch := make([]int64, 0, batchSize)

	for {
		batch = batch[:0]

		id, ok := <-d.deleteCh
		if !ok {
			return
		}
		batch = append(batch, id)

		timer := time.NewTimer(linger)
	collecting:
		for len(batch) < batchSize {
			select {
			case id, ok := <-d.deleteCh:
				if !ok {
					timer.Stop()
					d.flushDeletes(batch)
					return
				}
				batch = append(batch, id)
			case <-timer.C:
				break collecting
			}
		}
		timer.Stop()
		d.flushDeletes(batch)
	}
}

func (d *DurableEmitter) flushDeletes(ids []int64) {
	if len(ids) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.PublishTimeout)
	defer cancel()

	tDelete := time.Now()
	deleted, err := d.store.BatchDelete(ctx, ids)
	deleteElapsed := time.Since(tDelete)

	if h := d.cfg.Hooks; h != nil && h.OnBatchDelete != nil {
		h.OnBatchDelete(deleteElapsed, int(deleted))
	}
	if err != nil {
		d.eng.Errorw("failed to batch delete delivered events", "count", len(ids), "error", err)
		return
	}
	d.eng.Debugw("DurableEmitter: coalesced delete flushed", "submitted", len(ids), "deleted", deleted)
	if d.metrics != nil {
		d.metrics.deliverComplete.Add(ctx, deleted)
	}
}

func (d *DurableEmitter) retransmitLoop() {
	ticker := time.NewTicker(d.cfg.RetransmitInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.retransmitPending()
		}
	}
}

func (d *DurableEmitter) retransmitPending() {
	ctx, cancel := d.stopCh.NewCtx()
	defer cancel()

	cutoff := time.Now().Add(-d.cfg.RetransmitAfter)
	pending, err := d.store.ListPending(ctx, cutoff, d.retransmitCursorTs, d.retransmitCursorID, d.cfg.RetransmitBatchSize)
	if err != nil {
		d.eng.Errorw("failed to list pending events", "error", err)
		return
	}

	// Advance the paging cursor. A full page means there may be more rows ahead,
	// so continue after the last one next tick; a short page means we reached the
	// end of the currently-eligible set, so wrap back to the oldest. This lets the
	// loop page through (and past) poison rows instead of re-listing the same head
	// every tick — a poison row is revisited only once per full sweep.
	if len(pending) < d.cfg.RetransmitBatchSize {
		d.retransmitCursorTs = time.Time{}
		d.retransmitCursorID = 0
	} else {
		last := pending[len(pending)-1]
		d.retransmitCursorTs = last.CreatedAt
		d.retransmitCursorID = last.ID
	}

	d.eng.Debugw("DurableEmitter: retransmit pending scan",
		"retransmit_list_batch", len(pending),
		"retransmit_after", d.cfg.RetransmitAfter.String(),
		"list_limit", d.cfg.RetransmitBatchSize,
		"cursor_ts", d.retransmitCursorTs,
		"cursor_id", d.retransmitCursorID,
	)

	if len(pending) == 0 {
		return
	}

	d.retransmit(ctx, pending)
}

// retransmit re-enqueues pending DB rows through the batch emitter. Each row
// gets its own delivery callback that deletes it on success.
func (d *DurableEmitter) retransmit(ctx context.Context, pending []DurableEvent) {
	var enqueued, skipped int

	for _, pe := range pending {
		select {
		case <-d.stopCh:
			return
		default:
		}

		eventPb := new(chipingress.CloudEventPb)
		if err := proto.Unmarshal(pe.Payload, eventPb); err != nil {
			d.eng.Errorw("DurableEmitter: failed to unmarshal event for retransmit", "id", pe.ID, "error", err)
			continue
		}

		id := pe.ID
		if err := d.batchEmitter.QueueMessage(eventPb, d.deliveryCallback(id, eventPb, time.Now(), publishPhaseRetransmit)); err != nil {
			skipped++
			if d.metrics != nil {
				d.metrics.batchEnqueueBufferFull.Add(ctx, 1,
					metric.WithAttributes(attribute.String("phase", publishPhaseRetransmit.String())))
			}
		} else {
			enqueued++
		}
	}

	d.eng.Infow("DurableEmitter: retransmit queued to batch emitter",
		"enqueued", enqueued,
		"skipped_buffer_full", skipped,
		"total_pending", len(pending),
	)
}

func (d *DurableEmitter) expiryLoop() {
	ticker := time.NewTicker(d.cfg.ExpiryInterval)
	defer ticker.Stop()

	ctx, cancel := d.stopCh.NewCtx()
	defer cancel()
	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			deleted, err := d.store.DeleteExpired(ctx, d.cfg.EventTTL)
			if err != nil {
				d.eng.Errorw("failed to delete expired events", "error", err)
				continue
			}
			if deleted > 0 {
				if d.metrics != nil {
					d.metrics.expiredPurged.Add(ctx, deleted)
				}
				d.eng.Infow("purged expired events", "count", deleted)
			}
		}
	}
}

func (d *DurableEmitter) queueStatsNearExpiryLead() time.Duration {
	lead := 5 * time.Minute
	if d.cfg.Metrics != nil && d.cfg.Metrics.NearExpiryLead > 0 {
		lead = d.cfg.Metrics.NearExpiryLead
	}
	return lead
}

func (d *DurableEmitter) metricsLoop() {
	mc := d.cfg.Metrics
	poll := mc.PollInterval
	if poll <= 0 {
		poll = 500 * time.Millisecond
	}

	ctx, cancel := d.stopCh.NewCtx()
	defer cancel()

	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			if obs, ok := d.store.(DurableQueueObserver); ok {
				st, err := obs.ObserveDurableQueue(ctx, d.cfg.EventTTL, d.queueStatsNearExpiryLead())
				if err != nil {
					d.eng.Debugw("DurableEmitter: queue observe failed; keeping last depth", "error", err)
				} else {
					d.metrics.queueDepth.Record(ctx, st.TotalRows)
					d.metrics.recordQueueStats(ctx, st, mc.MaxQueuePayloadBytes)
				}
			}
			if d.insertCh != nil {
				if c := cap(d.insertCh); c > 0 {
					d.metrics.insertCoalescerFill.Record(ctx, float64(len(d.insertCh))/float64(c))
				}
			} else {
				d.metrics.insertCoalescerFill.Record(ctx, 0)
			}
			if d.deleteCh != nil {
				if c := cap(d.deleteCh); c > 0 {
					d.metrics.deleteCoalescerFill.Record(ctx, float64(len(d.deleteCh))/float64(c))
				}
			} else {
				d.metrics.deleteCoalescerFill.Record(ctx, 0)
			}
			d.metrics.pollProcessGauges(ctx)
		}
	}
}

// ensureIdempotencyKey sets attrs[IdempotencyKeyAttr] to a deterministic hash
// of source, type, and body when the caller did not supply a non-empty key.
func ensureIdempotencyKey(attrs map[string]any, sourceDomain, entityType string, body []byte) {
	if val, ok := attrs[chipingress.IdempotencyKeyAttr].(string); ok && val != "" {
		return
	}
	attrs[chipingress.IdempotencyKeyAttr] = defaultIdempotencyKey(sourceDomain, entityType, body)
}

func defaultIdempotencyKey(sourceDomain, entityType string, body []byte) string {
	h := sha256.New()
	for _, s := range []string{sourceDomain, entityType} {
		h.Write(binary.BigEndian.AppendUint32(nil, uint32(len(s))))
		h.Write([]byte(s))
	}
	h.Write(binary.BigEndian.AppendUint32(nil, uint32(len(body))))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// parseAttrs converts a variadic slice of (key, value) pairs (with optional
// embedded map[string]any) into a flat attributes map.
func parseAttrs(attrKVs ...any) map[string]any {
	a := make(map[string]any, len(attrKVs)/2)
	l := len(attrKVs)
	for i := 0; i < l; {
		switch t := attrKVs[i].(type) {
		case map[string]any:
			maps.Copy(a, t)
			i++
		case string:
			if i+1 >= l {
				break
			}
			a[t] = attrKVs[i+1]
			i += 2
		default:
			return a
		}
	}
	return a
}

// extractSourceAndType returns the CloudEvent source domain and entity type
// from the supplied attributes. Callers must provide the canonical CloudEvents
// keys "source" and "type". Both must be non-empty strings.
func extractSourceAndType(attrKVs ...any) (sourceDomain, entityType string, err error) {
	attrs := parseAttrs(attrKVs...)
	if v, ok := attrs["source"].(string); ok {
		sourceDomain = v
	}
	if v, ok := attrs["type"].(string); ok {
		entityType = v
	}
	if sourceDomain == "" {
		return "", "", errors.New(`"source" not found in provided key/value attributes`)
	}
	if entityType == "" {
		return "", "", errors.New(`"type" not found in provided key/value attributes`)
	}
	return sourceDomain, entityType, nil
}
