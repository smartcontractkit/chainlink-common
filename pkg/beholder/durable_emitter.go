package beholder

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// DurableEmitterConfig configures the DurableEmitter behaviour.
type DurableEmitterConfig struct {
	// RetransmitInterval controls how often the retransmit loop ticks.
	RetransmitInterval time.Duration
	// RetransmitAfter is the minimum age of an event before the retransmit
	// loop considers it. This gives the immediate-publish path time to succeed.
	RetransmitAfter time.Duration
	// RetransmitBatchSize caps how many pending rows are listed per retransmit tick
	// (each row is sent with its own Publish RPC).
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
	// PublishBatchSize enables batched publishing via PublishBatch RPC when > 0.
	// Events are collected into batches of this size before a single PublishBatch
	// call is made. Zero disables batching (each Emit spawns its own goroutine
	// with an individual Publish RPC — the legacy behaviour).
	PublishBatchSize int
	// PublishBatchWorkers is the number of concurrent goroutines that read from
	// the batch channel and call PublishBatch. More workers means higher
	// throughput (each worker handles one in-flight batch at a time). Only used
	// when PublishBatchSize > 0. Zero defaults to 1.
	PublishBatchWorkers int
	// PublishBatchFlushInterval is the maximum time to wait for a full batch
	// before flushing a partial one. Only used when PublishBatchSize > 0.
	// Zero defaults to 50ms.
	PublishBatchFlushInterval time.Duration
	// PublishBatchChannelSize overrides the publishCh buffer capacity. Only
	// used when PublishBatchSize > 0. Zero defaults to max(PublishBatchSize*2000, 200_000).
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
	// PersistCloudEventSources limits durable persistence to these CloudEvent Source values
	// (the beholder_domain / ce_source). If nil, every source is persisted (library default).
	// If non-nil, only matching sources are inserted and retried; others get a single best-effort
	// Publish with no store insert. An empty slice persists nothing (all best-effort only).
	// A one-element slice containing only "*" is treated like nil (persist all).
	PersistCloudEventSources []string
	// QuietMode suppresses high-volume INFO-level logs (retransmit scan stats,
	// retransmit results, publish failures, expired event purges, etc.).
	// Error-level logs are never suppressed. Useful for load tests where the
	// logging overhead is measurable.
	QuietMode bool
}

// DurableEmitterHooks records Publish vs Delete latency to locate pipeline bottlenecks.
type DurableEmitterHooks struct {
	// OnEmitInsert is called after each store.Insert in Emit (the DB write that
	// blocks the caller). elapsed covers only the INSERT; err is nil on success.
	OnEmitInsert func(elapsed time.Duration, err error)
	// OnImmediatePublish is called after each async Publish in publishAndDelete (every attempt).
	// Only fires when PublishBatchSize == 0 (legacy per-event goroutine path).
	OnImmediatePublish func(elapsed time.Duration, err error)
	// OnImmediateDelete is called after MarkDelivered following a successful immediate Publish.
	// Only fires when PublishBatchSize == 0.
	OnImmediateDelete func(elapsed time.Duration, err error)
	// OnBatchPublish is called after each PublishBatch RPC in the batch publish loop.
	// batchSize is the number of events in the batch; err is nil on success.
	OnBatchPublish func(elapsed time.Duration, batchSize int, err error)
	// OnBatchMarkDelivered is called after MarkDeliveredBatch following a successful batch publish.
	OnBatchMarkDelivered func(elapsed time.Duration, count int)
	// OnRetransmitBatchPublish is called after each retransmit Publish (one RPC per queued event).
	OnRetransmitBatchPublish func(elapsed time.Duration, eventCount int, err error)
	// OnRetransmitBatchDeletes is called after a retransmit tick with total time and count of
	// successful MarkDelivered calls (mem store may delete rows; Postgres sets delivered_at).
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
	}
}

// DurableEmitter implements Emitter with persistence-backed delivery guarantees.
//
// On Emit the event is serialized and written to a DurableEventStore. Once the
// insert succeeds Emit returns nil — the caller has a durable guarantee. An
// immediate async Publish is attempted; on success the record is MarkDelivered
// (excluded from retries). Postgres stores then purge physical rows in batches;
// in-memory stores remove the row immediately. If Publish fails, a background
// retransmit loop retries via Publish (one RPC per pending row per tick, up to
// RetransmitBatchSize).
//
// A separate expiry loop garbage-collects events older than EventTTL to bound
// table growth.
// publishWork is a unit of work for the batch publish channel.
type publishWork struct {
	id    int64
	event *chipingress.CloudEventPb
}

type DurableEmitter struct {
	store  DurableEventStore
	client chipingress.Client
	cfg    DurableEmitterConfig
	log    logger.Logger

	metrics       *durableEmitterMetrics
	persistFilter persistSourceFilter

	// pendingCount is an exact, atomic count of rows inserted but not yet
	// delivered/deleted. Incremented on successful Insert, decremented on
	// MarkDelivered, Delete, or DeleteExpired. No polling required.
	pendingCount atomic.Int64
	pendingMax   atomic.Int64

	// publishCh buffers events for the batch publish loop. Nil when
	// PublishBatchSize == 0 (legacy per-goroutine mode).
	publishCh chan publishWork

	stopCh chan struct{}
	wg     sync.WaitGroup
	markWg sync.WaitGroup // tracks in-flight async MarkDelivered goroutines
}

// persistSourceFilter decides whether a CloudEvent source may be written to the durable store.
type persistSourceFilter struct {
	allowAll bool
	allowed  map[string]struct{}
}

func newPersistSourceFilter(sources []string) persistSourceFilter {
	if sources == nil {
		return persistSourceFilter{allowAll: true}
	}
	if len(sources) == 1 && strings.TrimSpace(sources[0]) == "*" {
		return persistSourceFilter{allowAll: true}
	}
	m := make(map[string]struct{}, len(sources))
	for _, s := range sources {
		m[strings.TrimSpace(s)] = struct{}{}
	}
	return persistSourceFilter{allowed: m}
}

func (f persistSourceFilter) allows(source string) bool {
	if f.allowAll {
		return true
	}
	_, ok := f.allowed[source]
	return ok
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
		store:         store,
		client:        client,
		cfg:           cfg,
		log:           log,
		metrics:       m,
		persistFilter: newPersistSourceFilter(cfg.PersistCloudEventSources),
		stopCh:        make(chan struct{}),
	}
	if cfg.PublishBatchSize > 0 {
		bufSize := cfg.PublishBatchChannelSize
		if bufSize <= 0 {
			bufSize = cfg.PublishBatchSize * 2000
			if bufSize < 200_000 {
				bufSize = 200_000
			}
		}
		d.publishCh = make(chan publishWork, bufSize)
	}
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
	if d.publishCh != nil {
		n += batchWorkers
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
	if d.publishCh != nil {
		for i := 0; i < batchWorkers; i++ {
			go d.batchPublishLoop(ctx)
		}
	}
	if d.metrics != nil && d.cfg.Metrics != nil {
		go d.metricsLoop(ctx)
	}
}

// Emit persists the event then attempts async delivery when the CloudEvent source is allowed
// by PersistCloudEventSources; otherwise it performs a single best-effort Publish with no
// persistence. Returns nil once processing is accepted (insert succeeded, or non-persist path started).
func (d *DurableEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
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

	if !d.persistFilter.allows(sourceDomain) {
		cl := proto.Clone(eventPb)
		evCopy, ok := cl.(*chipingress.CloudEventPb)
		if !ok {
			emitFail()
			return fmt.Errorf("proto.Clone event: got %T, want *chipingress.CloudEventPb", cl)
		}
		go d.publishBestEffortNoStore(evCopy)
		return nil
	}

	payload, err := proto.Marshal(eventPb)
	if err != nil {
		emitFail()
		return fmt.Errorf("failed to marshal event proto: %w", err)
	}

	tIns := time.Now()
	id, err := d.store.Insert(ctx, payload)
	insElapsed := time.Since(tIns)
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

	d.incPending(1)

	if d.publishCh != nil {
		// Batch mode: enqueue for batch publish loop.
		select {
		case d.publishCh <- publishWork{id: id, event: eventPb}:
		default:
			// Channel full — event is safely in the DB; retransmit loop will deliver it.
			if !d.cfg.QuietMode {
				d.log.Warnw("DurableEmitter: batch publish channel full, relying on retransmit",
					"id", id, "ch_len", len(d.publishCh), "ch_cap", cap(d.publishCh))
			}
		}
	} else {
		// Legacy mode: fire-and-forget immediate delivery attempt.
		go d.publishAndDelete(id, eventPb)
	}

	return nil
}

// publishBestEffortNoStore performs one Publish without persisting or retries.
func (d *DurableEmitter) publishBestEffortNoStore(eventPb *chipingress.CloudEventPb) {
	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.PublishTimeout)
	defer cancel()

	detailKVs := cloudEventPublishKVs(0, "best_effort_no_store", d.cfg.PublishTimeout, eventPb)
	//d.log.Infow("DurableEmitter: Chip Ingress publish attempt (best-effort, not persisted)", detailKVs...)

	t0 := time.Now()
	_, err := d.client.Publish(ctx, eventPb)
	elapsed := time.Since(t0)
	if h := d.cfg.Hooks; h != nil && h.OnImmediatePublish != nil {
		h.OnImmediatePublish(elapsed, err)
	}
	mctx := context.Background()
	d.metrics.recordPublish(mctx, elapsed, "best_effort", err)
	if d.metrics != nil {
		if err != nil {
			d.metrics.publishImmErr.Add(mctx, 1)
		} else {
			d.metrics.publishImmOK.Add(mctx, 1)
		}
	}
	if err != nil {
		failKVs := append([]any{}, detailKVs...)
		failKVs = append(failKVs,
			"error", err,
			"elapsed", elapsed.String(),
			"elapsed_ms", elapsed.Milliseconds(),
		)
		//d.log.Infow("DurableEmitter: best-effort Chip publish failed (not persisted, no retry)", failKVs...)
		return
	}
	okKVs := append([]any{}, detailKVs...)
	okKVs = append(okKVs, "publish_rpc_elapsed_ms", elapsed.Milliseconds())
	//d.log.Infow("DurableEmitter: best-effort Chip publish succeeded (not persisted)", okKVs...)
}

// Close signals background loops to stop and waits for them to finish.
// When batch publishing is enabled the channel is closed so the batch loop
// can drain remaining events before returning.
func (d *DurableEmitter) Close() error {
	close(d.stopCh)
	if d.publishCh != nil {
		close(d.publishCh)
	}
	d.wg.Wait()
	d.markWg.Wait()
	return nil
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

// publishAndDelete attempts a single Publish and deletes the record on success.
func (d *DurableEmitter) publishAndDelete(id int64, eventPb *chipingress.CloudEventPb) {
	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.PublishTimeout)
	defer cancel()

	detailKVs := cloudEventPublishKVs(id, "immediate", d.cfg.PublishTimeout, eventPb)
	//d.log.Infow("DurableEmitter: Chip Ingress publish attempt (immediate)", detailKVs...)

	t0 := time.Now()
	_, err := d.client.Publish(ctx, eventPb)
	elapsed := time.Since(t0)
	if h := d.cfg.Hooks; h != nil && h.OnImmediatePublish != nil {
		h.OnImmediatePublish(elapsed, err)
	}
	mctx := context.Background()
	d.metrics.recordPublish(mctx, elapsed, "immediate", err)
	if d.metrics != nil {
		if err != nil {
			d.metrics.publishImmErr.Add(mctx, 1)
		} else {
			d.metrics.publishImmOK.Add(mctx, 1)
		}
	}
	if err != nil {
		failKVs := append([]any{}, detailKVs...)
		failKVs = append(failKVs,
			"error", err,
			"elapsed", elapsed.String(),
			"elapsed_ms", elapsed.Milliseconds(),
		)
		if !d.cfg.QuietMode {
			d.log.Infow("DurableEmitter: Chip Ingress publish failed (immediate), retransmit loop will retry", failKVs...)
		}
		return
	}

	pubOKKVs := append([]any{}, detailKVs...)
	pubOKKVs = append(pubOKKVs,
		"publish_rpc_elapsed", elapsed.String(),
		"publish_rpc_elapsed_ms", elapsed.Milliseconds(),
	)
	//d.log.Infow("DurableEmitter: Chip Ingress publish succeeded (immediate)", pubOKKVs...)

	t1 := time.Now()
	markErr := d.store.MarkDelivered(context.Background(), id)
	if h := d.cfg.Hooks; h != nil && h.OnImmediateDelete != nil {
		h.OnImmediateDelete(time.Since(t1), markErr)
	}
	if markErr == nil {
		d.decPending(1)
		if d.metrics != nil {
			d.metrics.deliverComplete.Add(mctx, 1)
		}
	}
	markElapsed := time.Since(t1)
	if markErr != nil {
		d.log.Errorw("failed to mark delivered event", "id", id, "error", markErr)
		return
	}
	delOKKVs := append([]any{}, detailKVs...)
	delOKKVs = append(delOKKVs,
		"publish_rpc_elapsed_ms", elapsed.Milliseconds(),
		"store_mark_delivered_elapsed", markElapsed.String(),
		"store_mark_delivered_elapsed_ms", markElapsed.Milliseconds(),
	)
	//d.log.Infow("DurableEmitter: durable row marked delivered after successful Chip publish (immediate)", delOKKVs...)
}

// batchPublishLoop reads events from publishCh, collects them into batches of
// PublishBatchSize, and sends each batch via PublishBatch RPC. A partial batch
// is flushed when PublishBatchFlushInterval elapses. On channel close (during
// Close), remaining items are drained and published.
func (d *DurableEmitter) batchPublishLoop(ctx context.Context) {
	defer d.wg.Done()

	batchSize := d.cfg.PublishBatchSize
	flushInterval := d.cfg.PublishBatchFlushInterval
	if flushInterval <= 0 {
		flushInterval = 50 * time.Millisecond
	}

	batch := make([]publishWork, 0, batchSize)
	flush := time.NewTicker(flushInterval)
	defer flush.Stop()

	for {
		select {
		case w, ok := <-d.publishCh:
			if !ok {
				// Channel closed — drain remaining items.
				if len(batch) > 0 {
					d.flushBatch(batch)
				}
				return
			}
			batch = append(batch, w)
			if len(batch) >= batchSize {
				d.flushBatch(batch)
				batch = batch[:0]
			}
		case <-flush.C:
			if len(batch) > 0 {
				d.flushBatch(batch)
				batch = batch[:0]
			}
		case <-ctx.Done():
			// Drain channel on context cancellation.
			for w := range d.publishCh {
				batch = append(batch, w)
				if len(batch) >= batchSize {
					d.flushBatch(batch)
					batch = batch[:0]
				}
			}
			if len(batch) > 0 {
				d.flushBatch(batch)
			}
			return
		}
	}
}

// flushBatch publishes a collected batch via PublishBatch and marks all events
// as delivered in a single MarkDeliveredBatch call.
func (d *DurableEmitter) flushBatch(batch []publishWork) {
	events := make([]*chipingress.CloudEventPb, len(batch))
	ids := make([]int64, len(batch))
	for i, w := range batch {
		events[i] = w.event
		ids[i] = w.id
	}

	batchPb := &chipingress.CloudEventBatch{Events: events}
	pubCtx, cancel := context.WithTimeout(context.Background(), d.cfg.PublishTimeout)
	defer cancel()

	t0 := time.Now()
	_, err := d.client.PublishBatch(pubCtx, batchPb)
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
		} else if !d.cfg.QuietMode {
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

	events := make([]*chipingress.CloudEventPb, 0, len(pending))
	ids := make([]int64, 0, len(pending))

	for _, pe := range pending {
		ev := new(chipingress.CloudEventPb)
		if err := proto.Unmarshal(pe.Payload, ev); err != nil {
			d.log.Errorw("corrupt pending event, deleting", "id", pe.ID, "error", err)
			if delErr := d.store.Delete(ctx, pe.ID); delErr == nil {
				d.decPending(1)
			}
			continue
		}
		if !d.persistFilter.allows(ev.GetSource()) {
			if !d.cfg.QuietMode {
				d.log.Infow("DurableEmitter: dropping queued event (ce_source not in PersistCloudEventSources)",
					"id", pe.ID, "ce_source", ev.GetSource(), "ce_type", ev.GetType())
			}
			if delErr := d.store.Delete(ctx, pe.ID); delErr == nil {
				d.decPending(1)
			}
			continue
		}
		events = append(events, ev)
		ids = append(ids, pe.ID)
	}
	if len(events) == 0 {
		return
	}

	// One Publish per row so a single bad or rejected event does not block the rest of the slice.
	tDel := time.Now()
	var markedDelivered int
	for i := range events {
		detailKVs := cloudEventPublishKVs(ids[i], "retransmit", d.cfg.PublishTimeout, events[i])
		//d.log.Infow("DurableEmitter: Chip Ingress publish attempt (retransmit)", detailKVs...)

		tPub := time.Now()
		pubCtx, cancel := context.WithTimeout(context.Background(), d.cfg.PublishTimeout)
		_, pubErr := d.client.Publish(pubCtx, events[i])
		cancel()
		elapsed := time.Since(tPub)
		if h := d.cfg.Hooks; h != nil && h.OnRetransmitBatchPublish != nil {
			h.OnRetransmitBatchPublish(elapsed, 1, pubErr)
		}
		d.metrics.recordPublish(context.Background(), elapsed, "retransmit", pubErr)
		if pubErr != nil {
			if d.metrics != nil {
				d.metrics.publishBatchEvErr.Add(ctx, 1)
			}
			failKVs := append([]any{}, detailKVs...)
			failKVs = append(failKVs,
				"error", pubErr,
				"elapsed", elapsed.String(),
				"elapsed_ms", elapsed.Milliseconds(),
			)
			if !d.cfg.QuietMode {
				d.log.Infow("DurableEmitter: Chip Ingress publish failed (retransmit)", failKVs...)
			}
			continue
		}
		pubOKKVs := append([]any{}, detailKVs...)
		pubOKKVs = append(pubOKKVs,
			"publish_rpc_elapsed", elapsed.String(),
			"publish_rpc_elapsed_ms", elapsed.Milliseconds(),
		)
		//d.log.Infow("DurableEmitter: Chip Ingress publish succeeded (retransmit)", pubOKKVs...)
		if d.metrics != nil {
			d.metrics.publishBatchEvOK.Add(ctx, 1)
		}
		tMarkOne := time.Now()
		if markErr := d.store.MarkDelivered(ctx, ids[i]); markErr != nil {
			d.log.Errorw("failed to mark retransmitted event delivered", "id", ids[i], "error", markErr)
			continue
		}
		d.decPending(1)
		markedDelivered++
		if d.metrics != nil {
			d.metrics.deliverComplete.Add(ctx, 1)
		}
		markElapsed := time.Since(tMarkOne)
		delOKKVs := append([]any{}, detailKVs...)
		delOKKVs = append(delOKKVs,
			"publish_rpc_elapsed_ms", elapsed.Milliseconds(),
			"store_mark_delivered_elapsed", markElapsed.String(),
			"store_mark_delivered_elapsed_ms", markElapsed.Milliseconds(),
		)
		//d.log.Infow("DurableEmitter: durable row deleted after successful Chip publish (retransmit)", delOKKVs...)
	}
	if markedDelivered > 0 && !d.cfg.QuietMode {
		d.log.Infow("retransmitted events",
			"marked_delivered", markedDelivered,
			"attempted", len(events),
		)
	}
	if h := d.cfg.Hooks; h != nil && h.OnRetransmitBatchDeletes != nil && markedDelivered > 0 {
		h.OnRetransmitBatchDeletes(time.Since(tDel), markedDelivered)
	}
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
				if !d.cfg.QuietMode {
					d.log.Infow("purged expired events", "count", deleted)
				}
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
			if mc.RecordProcessStats {
				d.metrics.recordProcessMem(bctx)
				d.metrics.recordProcessCPU(bctx)
			}
		}
	}
}

// cloudEventPublishKVs returns structured fields for logging a Chip Ingress Publish RPC.
func cloudEventPublishKVs(durableRowID int64, phase string, timeout time.Duration, ev *chipingress.CloudEventPb) []any {
	if ev == nil {
		return []any{
			"durable_row_id", durableRowID,
			"publish_phase", phase,
			"publish_timeout", timeout.String(),
			"ce_nil", true,
		}
	}

	attrs := ev.GetAttributes()
	bin := ev.GetBinaryData()
	text := ev.GetTextData()
	pd := ev.GetProtoData()
	var protoTypeURL string
	if pd != nil {
		protoTypeURL = pd.GetTypeUrl()
	}

	attrKeys := make([]string, 0, len(attrs))
	for k := range attrs {
		attrKeys = append(attrKeys, k)
	}
	slices.Sort(attrKeys)

	kvs := []any{
		"durable_row_id", durableRowID,
		"publish_phase", phase,
		"publish_timeout", timeout.String(),
		"ce_id", ev.GetId(),
		"ce_source", ev.GetSource(),
		"ce_type", ev.GetType(),
		"ce_spec_version", ev.GetSpecVersion(),
		"ce_data_binary_bytes", len(bin),
		"ce_data_text_bytes", len(text),
		"ce_proto_data_type_url", protoTypeURL,
		"ce_attribute_count", len(attrs),
		"ce_attribute_keys", strings.Join(attrKeys, ","),
		"ce_attr_datacontenttype", cloudEventAttrString(attrs, "datacontenttype"),
		"ce_attr_dataschema", cloudEventAttrString(attrs, "dataschema"),
		"ce_attr_subject", cloudEventAttrString(attrs, "subject"),
	}
	return kvs
}

func cloudEventAttrString(attrs map[string]*cepb.CloudEventAttributeValue, key string) string {
	if attrs == nil {
		return ""
	}
	v := attrs[key]
	if v == nil {
		return ""
	}
	if s := v.GetCeString(); s != "" {
		return s
	}
	if s := v.GetCeUri(); s != "" {
		return s
	}
	return ""
}
