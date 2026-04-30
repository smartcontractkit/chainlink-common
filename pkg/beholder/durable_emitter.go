package beholder

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	grpcEncoding "google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// DurableEmitterConfig configures the DurableEmitter behaviour.
type DurableEmitterConfig struct {
	// RetransmitInterval controls how often the retransmit loop ticks.
	RetransmitInterval time.Duration
	// RetransmitAfter is the minimum age of an event before the retransmit
	// loop considers it. This gives the batch publish path time to succeed.
	RetransmitAfter time.Duration
	// RetransmitBatchSize caps how many pending rows are listed per retransmit tick
	// (each row is enqueued for the batch publish workers).
	RetransmitBatchSize int
	// ExpiryInterval controls how often the expiry loop ticks.
	ExpiryInterval time.Duration
	// EventTTL is the maximum age of an event before it is expired.
	EventTTL time.Duration
	// PublishTimeout is the per-RPC deadline for each Publish call.
	PublishTimeout time.Duration
	// PurgeInterval is how often the purge loop runs to batch-delete rows that
	// were marked delivered (Postgres). Zero defaults to 250ms.
	PurgeInterval time.Duration
	// PurgeBatchSize is the maximum rows removed per PurgeDelivered call. Zero defaults to 500.
	PurgeBatchSize int
	// PublishBatchSize is the target number of events per PublishBatch RPC. Values below 1 are
	// clamped to 1 in NewDurableEmitter.
	PublishBatchSize int
	// PublishBatchWorkers is the number of concurrent goroutines that read from
	// the batch channel and call PublishBatch. More workers means higher
	// throughput (each worker handles one in-flight batch at a time). Zero defaults to 1.
	PublishBatchWorkers int
	// PublishBatchFlushInterval is the maximum time to wait for a full batch
	// before flushing a partial one. Zero defaults to 50ms.
	PublishBatchFlushInterval time.Duration
	// PublishBatchChannelSize overrides the publishCh buffer capacity. Zero
	// defaults to max(PublishBatchSize*2000, 200_000).
	PublishBatchChannelSize int
	// DisablePruning disables the background purge (PurgeDelivered) and expiry
	// (DeleteExpired) loops. Events remain in the DB after delivery. Useful for
	// post-test analysis of created_at / delivered_at timestamps.
	DisablePruning bool
	// Hooks is optional instrumentation (load tests, profiling). Nil fields are skipped.
	// Callbacks may run from many goroutines; implementations must be thread-safe.
	Hooks *DurableEmitterHooks
	// Metrics enables OpenTelemetry instruments on beholder.GetMeter() (queue, publish, store, optional process stats).
	// Nil disables.
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
}

// DurableEmitterHooks records Publish vs Delete latency to locate pipeline bottlenecks.
type DurableEmitterHooks struct {
	// OnEmitInsert is called after each store.Insert in Emit (the DB write that
	// blocks the caller). elapsed covers only the INSERT; err is nil on success.
	OnEmitInsert func(elapsed time.Duration, err error)
	// OnImmediatePublish is legacy; the durable emitter no longer uses unary Publish for durable emits.
	OnImmediatePublish func(elapsed time.Duration, err error)
	// OnImmediateDelete is unused; retained for API compatibility.
	OnImmediateDelete func(elapsed time.Duration, err error)
	// OnBatchPublish is called after each PublishBatch RPC in the batch publish loop.
	// batchSize is the number of events in the batch; err is nil on success.
	OnBatchPublish func(elapsed time.Duration, batchSize int, err error)
	// OnBatchMarkDelivered is called after MarkDeliveredBatch following a successful batch publish.
	OnBatchMarkDelivered func(elapsed time.Duration, count int)
	// OnRetransmitBatchPublish is not invoked (retransmit uses the same batch loop as Emit; use OnBatchPublish).
	OnRetransmitBatchPublish func(elapsed time.Duration, eventCount int, err error)
	// OnRetransmitBatchDeletes is not invoked (mark delivered is async per batch; hook retained for API compatibility).
	OnRetransmitBatchDeletes func(elapsed time.Duration, markedDeliveredCount int)
}

func DefaultDurableEmitterConfig() DurableEmitterConfig {
	return DurableEmitterConfig{
		RetransmitInterval:  5 * time.Second,
		RetransmitAfter:     10 * time.Second,
		RetransmitBatchSize: 100,
		ExpiryInterval:      1 * time.Minute,
		EventTTL:            24 * time.Hour,
		PublishTimeout:      5 * time.Second,
		PurgeInterval:       250 * time.Millisecond,
		PurgeBatchSize:      500,
		PublishBatchSize:    1,
	}
}

// DurableEmitter implements Emitter with persistence-backed delivery guarantees.
//
// Emit writes to a DurableEventStore, returns nil after insert, and enqueues the
// row for async PublishBatch delivery. Successful publishes are followed by
// MarkDeliveredBatch; the purge loop removes delivered rows from Postgres. When
// Chip is down or the publish channel is full, a retransmit loop lists stale
// pending rows and re-enqueues them to the same batch workers (up to
// RetransmitBatchSize per tick).
//
// A separate expiry loop garbage-collects events older than EventTTL to bound
// table growth.
// publishWork is a unit of work for the batch publish channel.
type publishWork struct {
	id      int64
	payload []byte                    // serialized CloudEvent proto (always set)
	event   *chipingress.CloudEventPb // original struct; set from Emit(), nil from retransmit
}

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
	store  DurableEventStore
	client chipingress.Client
	cfg    DurableEmitterConfig
	log    logger.Logger

	metrics *durableEmitterMetrics

	// rawConn is the underlying gRPC connection when the client exposes it.
	// Non-nil enables zero-copy batch publishing (protowire + ForceCodec).
	rawConn *grpc.ClientConn

	// batchInserter is non-nil when the store supports multi-row INSERTs
	// and InsertBatchSize > 0.
	batchInserter BatchInserter
	// insertCh buffers payloads for the write coalescer. Nil when batch
	// inserting is disabled.
	insertCh chan *insertRequest

	// pendingCount is an exact, atomic count of rows inserted but not yet
	// delivered/deleted. Incremented on successful Insert, decremented on
	// MarkDelivered, Delete, or DeleteExpired. No polling required.
	pendingCount atomic.Int64
	pendingMax   atomic.Int64

	// publishCh buffers events for the batch publish loop.
	publishCh chan publishWork

	stopCh chan struct{}
	wg     sync.WaitGroup
	markWg sync.WaitGroup // tracks in-flight async MarkDelivered goroutines
}

// grpcConnProvider is an optional interface for clients that expose the
// underlying gRPC connection. When satisfied, DurableEmitter uses zero-copy
// batch publishing to avoid protobuf marshal/unmarshal overhead.
type grpcConnProvider interface {
	Conn() *grpc.ClientConn
}

// rawBytesCodec is a gRPC codec that passes []byte through without
// marshal/unmarshal. Name returns "proto" so the wire content-type is
// application/grpc+proto, matching what Chip Ingress expects.
type rawBytesCodec struct{}

func (rawBytesCodec) Marshal(v any) ([]byte, error) {
	b, ok := v.([]byte)
	if !ok {
		return nil, fmt.Errorf("rawBytesCodec.Marshal: expected []byte, got %T", v)
	}
	return b, nil
}

func (rawBytesCodec) Unmarshal(data []byte, v any) error {
	bp, ok := v.(*[]byte)
	if !ok {
		return fmt.Errorf("rawBytesCodec.Unmarshal: expected *[]byte, got %T", v)
	}
	*bp = data
	return nil
}

func (rawBytesCodec) Name() string { return "proto" }

var _ grpcEncoding.Codec = rawBytesCodec{}

// buildBatchBytes constructs the protobuf wire format for a CloudEventBatch
// directly from already-serialized CloudEvent payloads. Each payload is
// wrapped with the field-1 tag and length prefix — no unmarshal/re-marshal.
func buildBatchBytes(payloads [][]byte) []byte {
	size := 0
	for _, p := range payloads {
		size += protowire.SizeTag(1) + protowire.SizeBytes(len(p))
	}
	buf := make([]byte, 0, size)
	for _, p := range payloads {
		buf = protowire.AppendTag(buf, 1, protowire.BytesType)
		buf = protowire.AppendBytes(buf, p)
	}
	return buf
}

var _ Emitter = (*DurableEmitter)(nil)

func NewDurableEmitter(
	store DurableEventStore,
	client chipingress.Client,
	cfg DurableEmitterConfig,
	log logger.Logger,
) (*DurableEmitter, error) {
	if store == nil {
		return nil, fmt.Errorf("durable event store is nil")
	}
	if client == nil {
		return nil, fmt.Errorf("chipingress client is nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}
	if cfg.PublishBatchSize < 1 {
		cfg.PublishBatchSize = 1
	}
	var m *durableEmitterMetrics
	if cfg.Metrics != nil {
		var err error
		m, err = newDurableEmitterMetrics()
		if err != nil {
			return nil, fmt.Errorf("durable emitter metrics: %w", err)
		}
		store = newMetricsInstrumentedStore(store, m)
	}
	d := &DurableEmitter{
		store:   store,
		client:  client,
		cfg:     cfg,
		log:     log,
		metrics: m,
		stopCh:  make(chan struct{}),
	}
	if cp, ok := client.(grpcConnProvider); ok {
		d.rawConn = cp.Conn()
		log.Infow("DurableEmitter: raw-codec batch publishing enabled (zero-copy protowire)")
	}
	if cfg.InsertBatchSize > 0 {
		if bi, ok := store.(BatchInserter); ok {
			d.batchInserter = bi
			chanSize := cfg.InsertBatchSize * 200
			if chanSize < 10_000 {
				chanSize = 10_000
			}
			d.insertCh = make(chan *insertRequest, chanSize)
			log.Infow("DurableEmitter: write coalescing enabled",
				"insertBatchSize", cfg.InsertBatchSize,
				"insertBatchWorkers", cfg.InsertBatchWorkers,
				"insertBatchFlushInterval", cfg.InsertBatchFlushInterval)
		}
	}
	bufSize := cfg.PublishBatchChannelSize
	if bufSize <= 0 {
		bufSize = cfg.PublishBatchSize * 2000
		if bufSize < 200_000 {
			bufSize = 200_000
		}
	}
	d.publishCh = make(chan publishWork, bufSize)
	return d, nil
}

// Start launches the retransmit, expiry, purge, and (optionally) batch publish
// background loops. Cancel the supplied context or call Close to stop them.
func (d *DurableEmitter) Start(ctx context.Context) {
	n := 1 // retransmit always runs
	if !d.cfg.DisablePruning {
		n += 2 // expiry + purge
	}
	batchWorkers := d.cfg.PublishBatchWorkers
	if batchWorkers <= 0 {
		batchWorkers = 1
	}
	n += batchWorkers
	insertWorkers := d.cfg.InsertBatchWorkers
	if insertWorkers <= 0 {
		insertWorkers = 4
	}
	if d.insertCh != nil {
		n += insertWorkers
	}
	if d.metrics != nil && d.cfg.Metrics != nil {
		n++
	}
	d.wg.Add(n)
	go d.retransmitLoop(ctx)
	if !d.cfg.DisablePruning {
		go d.expiryLoop(ctx)
		go d.purgeLoop(ctx)
	}
	if d.insertCh != nil {
		for i := 0; i < insertWorkers; i++ {
			go d.insertBatchLoop(ctx)
		}
	}
	for i := 0; i < batchWorkers; i++ {
		go d.batchPublishLoop(ctx)
	}
	if d.metrics != nil && d.cfg.Metrics != nil {
		go d.metricsLoop(ctx)
	}
}

// Emit persists the event then enqueues it for async PublishBatch delivery. Returns nil once
// the insert is accepted (or the coalesced insert path completes successfully).
func (d *DurableEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
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
	sourceDomain, entityType, err := ExtractSourceAndType(attrKVs...)
	if err != nil {
		emitFail()
		return err
	}

	event, err := chipingress.NewEvent(sourceDomain, entityType, body, newAttributes(attrKVs...))
	if err != nil {
		emitFail()
		return err
	}

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
		tIns := time.Now()
		select {
		case d.insertCh <- req:
		case <-ctx.Done():
			emitFail()
			return ctx.Err()
		}
		res := <-req.result
		insElapsed = time.Since(tIns)
		if h := d.cfg.Hooks; h != nil && h.OnEmitInsert != nil {
			h.OnEmitInsert(insElapsed, res.err)
		}
		if d.metrics != nil {
			d.metrics.emitDuration.Record(ctx, insElapsed.Seconds())
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
			d.metrics.emitDuration.Record(ctx, insElapsed.Seconds())
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

	d.incPending(1)

	// Batch path: enqueue for batch publish loop (PublishBatchSize is always >= 1).
	work := publishWork{id: id, payload: payload}
	if d.rawConn == nil {
		work.event = eventPb
	}
	select {
	case d.publishCh <- work:
	default:
		// Channel full — event is safely in the DB; retransmit loop will deliver it.
		d.log.Warnw("DurableEmitter: batch publish channel full, relying on retransmit",
			"id", id, "ch_len", len(d.publishCh), "ch_cap", cap(d.publishCh))
	}

	return nil
}

// Close signals background loops to stop and waits for them to finish.
// Insert and publish channels are closed so workers can drain.
func (d *DurableEmitter) Close() error {
	close(d.stopCh)
	if d.insertCh != nil {
		close(d.insertCh)
	}
	close(d.publishCh)
	d.wg.Wait()
	d.markWg.Wait()
	return nil
}

// insertBatchLoop collects insertRequest items from insertCh and flushes them
// as multi-row INSERTs via BatchInserter.InsertBatch. Uses a linger pattern:
// blocks for the first request, then collects more until the batch is full or
// the flush interval elapses.
func (d *DurableEmitter) insertBatchLoop(ctx context.Context) {
	defer d.wg.Done()
	batchSize := d.cfg.InsertBatchSize
	linger := d.cfg.InsertBatchFlushInterval
	if linger <= 0 {
		linger = 2 * time.Millisecond
	}
	batch := make([]*insertRequest, 0, batchSize)

	for {
		batch = batch[:0]

		// Block until first request or shutdown.
		select {
		case req, ok := <-d.insertCh:
			if !ok {
				return
			}
			batch = append(batch, req)
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		}

		// Linger to collect more.
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

		// Flush: multi-row INSERT.
		payloads := make([][]byte, len(batch))
		for i, r := range batch {
			payloads[i] = r.payload
		}
		ids, batchErr := d.batchInserter.InsertBatch(ctx, payloads)
		for i, r := range batch {
			if batchErr != nil {
				r.result <- insertResult{err: batchErr}
			} else {
				r.result <- insertResult{id: ids[i]}
			}
		}
	}
}

// PendingDepth returns the current exact pending queue depth (inserted but not
// yet delivered/deleted). Thread-safe; no DB query required.
func (d *DurableEmitter) PendingDepth() int64 { return d.pendingCount.Load() }

// PendingMax returns the highest pending queue depth observed since Start.
func (d *DurableEmitter) PendingMax() int64 { return d.pendingMax.Load() }

func (d *DurableEmitter) incPending(n int64) {
	cur := d.pendingCount.Add(n)
	updated := false
	for {
		old := d.pendingMax.Load()
		if cur <= old {
			break
		}
		if d.pendingMax.CompareAndSwap(old, cur) {
			updated = true
			break
		}
	}
	if d.metrics != nil {
		d.metrics.queueDepth.Record(context.Background(), cur)
		if updated {
			d.metrics.queueDepthMax.Record(context.Background(), cur)
		}
	}
}

func (d *DurableEmitter) decPending(n int64) {
	cur := d.pendingCount.Add(-n)
	if d.metrics != nil {
		d.metrics.queueDepth.Record(context.Background(), cur)
	}
}

// batchPublishLoop reads events from publishCh, collects them into batches of
// PublishBatchSize, and sends each batch via PublishBatch RPC. It blocks until
// the batch is full or PublishBatchFlushInterval elapses after the first event
// arrives (linger pattern), guaranteeing full batches at high throughput.
func (d *DurableEmitter) batchPublishLoop(ctx context.Context) {
	defer d.wg.Done()

	batchSize := d.cfg.PublishBatchSize
	flushInterval := d.cfg.PublishBatchFlushInterval
	if flushInterval <= 0 {
		flushInterval = 50 * time.Millisecond
	}

	batch := make([]publishWork, 0, batchSize)

	for {
		// Stage 1: block until at least one event arrives (or shutdown).
		select {
		case w, ok := <-d.publishCh:
			if !ok {
				return
			}
			batch = append(batch, w)
		case <-ctx.Done():
			d.drainPublishCh(batch)
			return
		}

		// Stage 2: linger — keep reading until batch is full or deadline.
		linger := time.NewTimer(flushInterval)
	fill:
		for len(batch) < batchSize {
			select {
			case w, ok := <-d.publishCh:
				if !ok {
					linger.Stop()
					if len(batch) > 0 {
						d.flushBatch(batch)
					}
					return
				}
				batch = append(batch, w)
			case <-linger.C:
				break fill
			case <-ctx.Done():
				linger.Stop()
				d.drainPublishCh(batch)
				return
			}
		}
		linger.Stop()

		d.flushBatch(batch)
		batch = batch[:0]
	}
}

// drainPublishCh flushes the given partial batch plus any remaining items on
// publishCh in batchSize chunks. Called during shutdown (ctx cancelled or
// channel closed).
func (d *DurableEmitter) drainPublishCh(batch []publishWork) {
	for w := range d.publishCh {
		batch = append(batch, w)
		if len(batch) >= d.cfg.PublishBatchSize {
			d.flushBatch(batch)
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		d.flushBatch(batch)
	}
}

// flushBatch publishes a collected batch via PublishBatch and marks all events
// as delivered in a single MarkDeliveredBatch call.
//
// When rawConn is available, batch wire bytes are built directly from the
// already-serialized payloads using protowire — zero unmarshal/re-marshal.
// Otherwise, falls back to the typed PublishBatch RPC.
func (d *DurableEmitter) flushBatch(batch []publishWork) {
	ids := make([]int64, len(batch))
	for i, w := range batch {
		ids[i] = w.id
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), d.cfg.PublishTimeout)
	defer cancel()

	t0 := time.Now()
	var err error
	if d.rawConn != nil {
		err = d.flushBatchRaw(pubCtx, batch)
	} else {
		err = d.flushBatchTyped(pubCtx, batch)
	}
	elapsed := time.Since(t0)

	if h := d.cfg.Hooks; h != nil && h.OnBatchPublish != nil {
		h.OnBatchPublish(elapsed, len(batch), err)
	}
	d.metrics.recordPublish(context.Background(), elapsed, "batch", err)

	if err != nil {
		if d.metrics != nil {
			d.metrics.publishBatchEvErr.Add(context.Background(), int64(len(batch)))
		}
		d.log.Warnw("DurableEmitter: PublishBatch failed, events will be retransmitted",
			"batch_size", len(batch), "error", err,
			"elapsed_ms", elapsed.Milliseconds())
		return
	}

	if d.metrics != nil {
		d.metrics.publishBatchEvOK.Add(context.Background(), int64(len(batch)))
	}

	// Async MarkDelivered: the DB UPDATE runs in a background goroutine so
	// the batch worker can immediately start collecting the next batch.
	// If MarkDelivered fails, events stay pending and the retransmit loop
	// delivers them (at-least-once semantics are unchanged).
	d.markWg.Add(1)
	go func() {
		defer d.markWg.Done()
		tMark := time.Now()
		marked, markErr := d.store.MarkDeliveredBatch(context.Background(), ids)
		markElapsed := time.Since(tMark)
		if h := d.cfg.Hooks; h != nil && h.OnBatchMarkDelivered != nil {
			h.OnBatchMarkDelivered(markElapsed, int(marked))
		}
		if markErr != nil {
			d.log.Errorw("failed to batch-mark events delivered", "batch_size", len(ids), "error", markErr)
			return
		}
		d.decPending(marked)
		if d.metrics != nil {
			d.metrics.deliverComplete.Add(context.Background(), marked)
		}
	}()
}

// flushBatchRaw builds the CloudEventBatch wire format directly from
// already-serialized payloads and sends it via raw gRPC Invoke — zero
// protobuf unmarshal/re-marshal overhead.
func (d *DurableEmitter) flushBatchRaw(ctx context.Context, batch []publishWork) error {
	payloads := make([][]byte, len(batch))
	for i, w := range batch {
		payloads[i] = w.payload
	}
	batchBytes := buildBatchBytes(payloads)
	var respBytes []byte
	return d.rawConn.Invoke(
		ctx,
		pb.ChipIngress_PublishBatch_FullMethodName,
		batchBytes,
		&respBytes,
		grpc.ForceCodec(rawBytesCodec{}),
	)
}

// flushBatchTyped builds a typed CloudEventBatch and sends it via the
// standard gRPC client. Used as fallback when rawConn is not available.
func (d *DurableEmitter) flushBatchTyped(ctx context.Context, batch []publishWork) error {
	events := make([]*chipingress.CloudEventPb, len(batch))
	for i, w := range batch {
		if w.event != nil {
			events[i] = w.event
		} else {
			ev := new(chipingress.CloudEventPb)
			if err := proto.Unmarshal(w.payload, ev); err != nil {
				return fmt.Errorf("unmarshal event %d for typed publish: %w", i, err)
			}
			events[i] = ev
		}
	}
	batchPb := &chipingress.CloudEventBatch{Events: events}
	_, err := d.client.PublishBatch(ctx, batchPb)
	return err
}

func (d *DurableEmitter) retransmitLoop(ctx context.Context) {
	defer d.wg.Done()
	ticker := time.NewTicker(d.cfg.RetransmitInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.retransmitPending(ctx)
		}
	}
}

func (d *DurableEmitter) retransmitPending(ctx context.Context) {
	cutoff := time.Now().Add(-d.cfg.RetransmitAfter)
	pending, err := d.store.ListPending(ctx, cutoff, d.cfg.RetransmitBatchSize)
	if err != nil {
		d.log.Errorw("failed to list pending events", "error", err)
		return
	}

	if obs, ok := d.store.(DurableQueueObserver); ok {
		st, obsErr := obs.ObserveDurableQueue(ctx, d.cfg.EventTTL, d.queueStatsNearExpiryLead())
		if obsErr != nil {
			d.log.Warnw("DurableEmitter: retransmit scan ObserveDurableQueue failed", "error", obsErr)
		} else {
			d.log.Infow("DurableEmitter: retransmit pending scan",
				"pending_rows", st.Depth,
				"pending_payload_bytes", st.PayloadBytes,
				"oldest_pending_age", st.OldestPendingAge.String(),
				"near_ttl_rows", st.NearTTLCount,
				"retransmit_list_batch", len(pending),
				"retransmit_after", d.cfg.RetransmitAfter.String(),
				"list_limit", d.cfg.RetransmitBatchSize,
			)
		}
	}

	if len(pending) == 0 {
		return
	}

	d.retransmit(pending)
}

// retransmit enqueues pending DB rows to publishCh so the batch workers handle
// publishing. When rawConn is set, payloads are passed through without
// proto.Unmarshal — the batch workers use buildBatchBytes for the wire format.
func (d *DurableEmitter) retransmit(pending []DurableEvent) {
	var enqueued int

	for _, pe := range pending {
		select {
		case <-d.stopCh:
			return
		default:
		}
		work := publishWork{id: pe.ID, payload: pe.Payload}

		select {
		case <-d.stopCh:
			return
		case d.publishCh <- work:
			enqueued++
		default:
		}
	}

	d.log.Infow("DurableEmitter: retransmit enqueued to batch workers",
		"enqueued", enqueued,
		"skipped_ch_full", len(pending)-enqueued,
		"total_pending", len(pending),
		"ch_len", len(d.publishCh),
		"ch_cap", cap(d.publishCh),
	)
}

func (d *DurableEmitter) purgeLoop(ctx context.Context) {
	defer d.wg.Done()
	interval := d.cfg.PurgeInterval
	if interval <= 0 {
		interval = 250 * time.Millisecond
	}
	batch := d.cfg.PurgeBatchSize
	if batch <= 0 {
		batch = 500
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			for {
				n, err := d.store.PurgeDelivered(ctx, batch)
				if err != nil {
					d.log.Errorw("failed to purge delivered chip durable events", "error", err)
					break
				}
				if n == 0 {
					break
				}
			}
		}
	}
}

func (d *DurableEmitter) expiryLoop(ctx context.Context) {
	defer d.wg.Done()
	ticker := time.NewTicker(d.cfg.ExpiryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			deleted, err := d.store.DeleteExpired(ctx, d.cfg.EventTTL)
			if err != nil {
				d.log.Errorw("failed to delete expired events", "error", err)
				continue
			}
			if deleted > 0 {
				d.decPending(deleted)
				if d.metrics != nil {
					d.metrics.expiredPurged.Add(context.Background(), deleted)
				}
				d.log.Infow("purged expired events", "count", deleted)
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

func (d *DurableEmitter) metricsLoop(ctx context.Context) {
	defer d.wg.Done()
	mc := d.cfg.Metrics
	poll := mc.PollInterval
	if poll <= 0 {
		poll = 10 * time.Second
	}
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			if d.metrics == nil {
				return
			}
			bctx := context.Background()
			d.metrics.queueDepth.Record(bctx, d.pendingCount.Load())
			d.metrics.queueDepthMax.Record(bctx, d.pendingMax.Load())
			if obs, ok := d.store.(DurableQueueObserver); ok {
				d.metrics.pollQueueGauges(bctx, obs, d.cfg.EventTTL, d.queueStatsNearExpiryLead(), mc.MaxQueuePayloadBytes)
			}
		}
	}
}
