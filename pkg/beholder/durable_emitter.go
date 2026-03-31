package beholder

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
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
}

// DurableEmitterHooks records Publish vs Delete latency to locate pipeline bottlenecks.
type DurableEmitterHooks struct {
	// OnImmediatePublish is called after each async Publish in publishAndDelete (every attempt).
	OnImmediatePublish func(elapsed time.Duration, err error)
	// OnImmediateDelete is called after Delete following a successful immediate Publish.
	OnImmediateDelete func(elapsed time.Duration, err error)
	// OnRetransmitBatchPublish is called after each retransmit Publish (one RPC per queued event).
	OnRetransmitBatchPublish func(elapsed time.Duration, eventCount int, err error)
	// OnRetransmitBatchDeletes is called after a retransmit tick with total time and successful delete count.
	OnRetransmitBatchDeletes func(elapsed time.Duration, deleteCount int)
}

func DefaultDurableEmitterConfig() DurableEmitterConfig {
	return DurableEmitterConfig{
		RetransmitInterval:  5 * time.Second,
		RetransmitAfter:     10 * time.Second,
		RetransmitBatchSize: 100,
		ExpiryInterval:      1 * time.Minute,
		EventTTL:            24 * time.Hour,
		PublishTimeout:      5 * time.Second,
	}
}

// DurableEmitter implements Emitter with persistence-backed delivery guarantees.
//
// On Emit the event is serialized and written to a DurableEventStore. Once the
// insert succeeds Emit returns nil — the caller has a durable guarantee. An
// immediate async Publish is attempted; on success the record is deleted. If
// that fails a background retransmit loop will pick the event up and retry via
// Publish (one RPC per pending row per tick, up to RetransmitBatchSize).
//
// A separate expiry loop garbage-collects events older than EventTTL to bound
// table growth.
type DurableEmitter struct {
	store  DurableEventStore
	client chipingress.Client
	cfg    DurableEmitterConfig
	log    logger.Logger

	metrics       *durableEmitterMetrics
	persistFilter persistSourceFilter

	stopCh chan struct{}
	wg     sync.WaitGroup
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
	return &DurableEmitter{
		store:         store,
		client:        client,
		cfg:           cfg,
		log:           log,
		metrics:       m,
		persistFilter: newPersistSourceFilter(cfg.PersistCloudEventSources),
		stopCh:        make(chan struct{}),
	}, nil
}

// Start launches the retransmit and expiry background loops.
// Cancel the supplied context or call Close to stop them.
func (d *DurableEmitter) Start(ctx context.Context) {
	n := 2
	if d.metrics != nil && d.cfg.Metrics != nil {
		n++
	}
	d.wg.Add(n)
	go d.retransmitLoop(ctx)
	go d.expiryLoop(ctx)
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
	if d.metrics != nil {
		d.metrics.emitDuration.Record(ctx, time.Since(tIns).Seconds())
		if err != nil {
			d.metrics.emitFail.Add(ctx, 1)
		} else {
			d.metrics.emitSuccess.Add(ctx, 1)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to persist event: %w", err)
	}

	// Fire-and-forget immediate delivery attempt.
	go d.publishAndDelete(id, eventPb)

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
func (d *DurableEmitter) Close() error {
	close(d.stopCh)
	d.wg.Wait()
	return nil
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
		d.log.Infow("DurableEmitter: Chip Ingress publish failed (immediate), retransmit loop will retry", failKVs...)
		return
	}

	pubOKKVs := append([]any{}, detailKVs...)
	pubOKKVs = append(pubOKKVs,
		"publish_rpc_elapsed", elapsed.String(),
		"publish_rpc_elapsed_ms", elapsed.Milliseconds(),
	)
	//d.log.Infow("DurableEmitter: Chip Ingress publish succeeded (immediate)", pubOKKVs...)

	t1 := time.Now()
	delErr := d.store.Delete(context.Background(), id)
	if h := d.cfg.Hooks; h != nil && h.OnImmediateDelete != nil {
		h.OnImmediateDelete(time.Since(t1), delErr)
	}
	if delErr == nil && d.metrics != nil {
		d.metrics.deliverComplete.Add(mctx, 1)
	}
	delElapsed := time.Since(t1)
	if delErr != nil {
		d.log.Errorw("failed to delete delivered event", "id", id, "error", delErr)
		return
	}
	delOKKVs := append([]any{}, detailKVs...)
	delOKKVs = append(delOKKVs,
		"publish_rpc_elapsed_ms", elapsed.Milliseconds(),
		"store_delete_elapsed", delElapsed.String(),
		"store_delete_elapsed_ms", delElapsed.Milliseconds(),
	)
	//d.log.Infow("DurableEmitter: durable row deleted after successful Chip publish (immediate)", delOKKVs...)
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
	if len(pending) == 0 {
		return
	}

	events := make([]*chipingress.CloudEventPb, 0, len(pending))
	ids := make([]int64, 0, len(pending))

	for _, pe := range pending {
		ev := new(chipingress.CloudEventPb)
		if err := proto.Unmarshal(pe.Payload, ev); err != nil {
			d.log.Errorw("corrupt pending event, deleting", "id", pe.ID, "error", err)
			_ = d.store.Delete(ctx, pe.ID)
			continue
		}
		if !d.persistFilter.allows(ev.GetSource()) {
			d.log.Infow("DurableEmitter: dropping queued event (ce_source not in PersistCloudEventSources)",
				"id", pe.ID, "ce_source", ev.GetSource(), "ce_type", ev.GetType())
			_ = d.store.Delete(ctx, pe.ID)
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
	var deleted int
	for i := range events {
		detailKVs := cloudEventPublishKVs(ids[i], "retransmit", d.cfg.PublishTimeout, events[i])
		d.log.Infow("DurableEmitter: Chip Ingress publish attempt (retransmit)", detailKVs...)

		tPub := time.Now()
		pubCtx, cancel := context.WithTimeout(context.Background(), d.cfg.PublishTimeout)
		_, pubErr := d.client.Publish(pubCtx, events[i])
		cancel()
		elapsed := time.Since(tPub)
		if h := d.cfg.Hooks; h != nil && h.OnRetransmitBatchPublish != nil {
			h.OnRetransmitBatchPublish(elapsed, 1, pubErr)
		}
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
			d.log.Infow("DurableEmitter: Chip Ingress publish failed (retransmit)", failKVs...)
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
		tDelOne := time.Now()
		if delErr := d.store.Delete(ctx, ids[i]); delErr != nil {
			d.log.Errorw("failed to delete retransmitted event", "id", ids[i], "error", delErr)
			continue
		}
		deleted++
		if d.metrics != nil {
			d.metrics.deliverComplete.Add(ctx, 1)
		}
		delElapsed := time.Since(tDelOne)
		delOKKVs := append([]any{}, detailKVs...)
		delOKKVs = append(delOKKVs,
			"publish_rpc_elapsed_ms", elapsed.Milliseconds(),
			"store_delete_elapsed", delElapsed.String(),
			"store_delete_elapsed_ms", delElapsed.Milliseconds(),
		)
		//d.log.Infow("DurableEmitter: durable row deleted after successful Chip publish (retransmit)", delOKKVs...)
	}
	if deleted > 0 {
		d.log.Debugw("retransmitted events", "deleted", deleted, "attempted", len(events))
	}
	if h := d.cfg.Hooks; h != nil && h.OnRetransmitBatchDeletes != nil && deleted > 0 {
		h.OnRetransmitBatchDeletes(time.Since(tDel), deleted)
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
				if d.metrics != nil {
					d.metrics.expiredPurged.Add(context.Background(), deleted)
				}
				d.log.Infow("purged expired events", "count", deleted)
			}
		}
	}
}

func (d *DurableEmitter) metricsLoop(ctx context.Context) {
	defer d.wg.Done()
	mc := d.cfg.Metrics
	poll := mc.PollInterval
	if poll <= 0 {
		poll = 10 * time.Second
	}
	lead := mc.NearExpiryLead
	if lead <= 0 {
		lead = 5 * time.Minute
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
			if obs, ok := d.store.(DurableQueueObserver); ok {
				d.metrics.pollQueueGauges(bctx, obs, d.cfg.EventTTL, lead, mc.MaxQueuePayloadBytes)
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
