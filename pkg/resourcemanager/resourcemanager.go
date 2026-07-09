// Package resourcemanager emits metering.v1 billing data for durable billable
// resources such as trigger registrations, workflow specs, and log filters.
//
// It emits two complementary, load-bearing billing streams, each covering
// exactly one resource identified by its ResourceIdentity:
//
//   - MeterRecords: the durable, first-class billing event stream. Each record
//     captures one signed delta — the level change of one request against a
//     durable resource (EmitDelta -> METER_ACTION_UPDATE) or one instantaneous
//     occurrence of consumption (EmitUsage -> METER_ACTION_USAGE). Records give
//     the consumer precise request-time edges and the audit trail.
//   - MeterSnapshots: the periodic absolute level of each active resource,
//     emitted on a timer, one per resource a registered Meterable reports as
//     active. Snapshots carry the level and the liveness signal; a resource
//     that stops being snapshotted is released, and that absence is the only
//     lifecycle-cleanup mechanism.
//
// The emitter is stateless with respect to metering: a delta is derived only
// from the request being fielded plus the producer's own store. There is no
// pairing contract, no balance ledger, and no emission-history invariant.
// Producers emit only when fielding a request that changes desired state; there
// are no process-lifecycle emissions.
//
// event_id is generated fresh per emission (a UUIDv4) by this manager and is
// the consumer's dedup key: deltas have counter semantics, so at-least-once
// delivery must dedup by event_id, and snapshot reconciliation then bounds any
// residual level drift.
//
// Bucket semantics: An org draws down credits in a time bucket if the bucket
// contains any positive-delta UPDATE record, any USAGE record, or any nonzero
// snapshot for a resource attributed to that org. A +N and -N cancelling within
// one bucket still triggers drawdown via the positive delta.
//
// The ResourceManager is the single owner of the snapshot tick: each producer
// starts the manager as a sub-service and only Registers itself; producers
// never run their own snapshot loop.
//
// Emission is fail-open by design: EmitDelta, EmitUsage, and the snapshot loop
// return no error, and a metering failure must never gate, delay, or retry the
// resource operation being metered. Failures surface via error-level logs and
// the resource_manager_*_failure_total counters.
package resourcemanager

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

// Beholder/ChIP routing attributes. Each entity value is the contract shared with
// the CHiP schema registration and the consumer topic name; all must match
// exactly.
const (
	domain             = "cll-meter"
	entity             = "metering.v1.MeterRecord"
	dataSchema         = "metering.v1.meter_record"
	snapshotEntity     = "metering.v1.MeterSnapshot"
	snapshotDataSchema = "metering.v1.meter_snapshot"
)

// Counter names are a dashboard contract; do not rename.
//
// Success means the record was accepted for asynchronous export (enqueued
// with the OTel batch processor), not that it was delivered to Kafka;
// delivery failures past this point are invisible to these counters. Emit is
// non-blocking only while the batch processors are enabled, which is the
// default.
const (
	emitSuccessCounterName         = "resource_manager_emit_success_total"
	emitFailureCounterName         = "resource_manager_emit_failure_total"
	snapshotEmitSuccessCounterName = "resource_manager_snapshot_emit_success_total"
	snapshotEmitFailureCounterName = "resource_manager_snapshot_emit_failure_total"
)

// utilizationGaugeName is the per-resource utilization gauge. Unlike the
// low-cardinality emit counters, it is labeled with EVERY ResourceIdentity
// dimension (including node_id and resource_id) so dashboards can break
// utilization down by any dimension. That is high cardinality by construction;
// it is the intended, requested behavior.
const utilizationGaugeName = "resource_manager_utilization"

// DefaultSnapshotInterval is the recommended snapshot period. It is NOT
// applied implicitly: a zero ResourceManagerConfig.SnapshotInterval disables
// snapshots. Callers that want snapshots pass this (or their own value)
// explicitly.
const DefaultSnapshotInterval = 60 * time.Second

// Emitter delivers an encoded metering message with its routing attributes.
// beholder.Emitter satisfies it, so production wiring is
// beholder.GetEmitter(); tests substitute a fake.
type Emitter interface {
	Emit(ctx context.Context, body []byte, attrKVs ...any) error
}

// ResourceManagerConfig configures a ResourceManager.
type ResourceManagerConfig struct {
	// MeterRecordsEnabled is the meter-record rollout gate. When false (the
	// default), EmitDelta and EmitUsage are no-ops.
	MeterRecordsEnabled bool

	// MeterSnapshotsEnabled gates snapshot emission. Snapshots are only emitted
	// when this is true AND MeterRecordsEnabled is true.
	MeterSnapshotsEnabled bool

	// Emitter delivers encoded records, typically beholder.GetEmitter(). A nil
	// Emitter makes EmitDelta and EmitUsage no-ops even when enabled and keeps
	// the snapshot loop from starting.
	Emitter Emitter

	// SnapshotInterval is the period between snapshots. Zero (the default)
	// DISABLES the snapshot loop; the manager still starts as a service and
	// EmitDelta / EmitUsage still work. Callers that want snapshots set a
	// positive value, e.g. DefaultSnapshotInterval. The default is not
	// substituted for zero — zero means off.
	SnapshotInterval time.Duration

	// Clock drives snapshot tick timing and record timestamps. Nil selects the
	// real clock.
	Clock clockwork.Clock
}

// registration is one registered Meterable. It is a pointer-identified handle
// so Register can return an idempotent unregister closure.
type registration struct {
	m Meterable
}

// ResourceManager emits MeterRecords and periodic MeterSnapshots for durable
// resources. It is a services.Service: callers start it (typically as a
// sub-service of the producer) and Register Meterables to be snapshotted. It
// is safe for concurrent use.
type ResourceManager struct {
	services.Service
	srvcEng *services.Engine

	lggr                  logger.SugaredLogger
	meterRecordsEnabled   bool
	meterSnapshotsEnabled bool
	emitter               Emitter
	snapshotInterval      time.Duration
	clock                 clockwork.Clock

	mu            sync.RWMutex
	registrations map[*registration]struct{}

	emitSuccess         metric.Int64Counter
	emitFailure         metric.Int64Counter
	snapshotEmitSuccess metric.Int64Counter
	snapshotEmitFailure metric.Int64Counter
	utilization         metric.Int64Gauge
}

// NewResourceManager returns a ResourceManager. A failure to create a metric
// instrument is logged and that instrument is skipped; it never prevents
// construction. The manager must be Started before its snapshot loop runs;
// EmitDelta and EmitUsage work regardless of Start.
func NewResourceManager(lggr logger.Logger, cfg ResourceManagerConfig) *ResourceManager {
	meter := beholder.GetMeter()
	sugared := logger.Sugared(lggr)
	newCounter := func(name string) metric.Int64Counter {
		c, err := meter.Int64Counter(name)
		if err != nil {
			sugared.Errorw("failed to create metering counter", "counter", name, "err", err)
			return nil
		}
		return c
	}
	gauge, err := meter.Int64Gauge(utilizationGaugeName)
	if err != nil {
		sugared.Errorw("failed to create metering gauge", "gauge", utilizationGaugeName, "err", err)
		gauge = nil
	}

	clock := cfg.Clock
	if clock == nil {
		clock = clockwork.NewRealClock()
	}

	meterSnapshotsEnabled := cfg.MeterRecordsEnabled && cfg.MeterSnapshotsEnabled && cfg.SnapshotInterval > 0
	if cfg.MeterSnapshotsEnabled && !cfg.MeterRecordsEnabled {
		sugared.Warn("MeterSnapshotsEnabled ignored because MeterRecordsEnabled is false")
	}

	rm := &ResourceManager{
		meterRecordsEnabled:   cfg.MeterRecordsEnabled,
		meterSnapshotsEnabled: meterSnapshotsEnabled,
		emitter:               cfg.Emitter,
		snapshotInterval:      cfg.SnapshotInterval,
		clock:                 clock,
		registrations:         make(map[*registration]struct{}),
		emitSuccess:           newCounter(emitSuccessCounterName),
		emitFailure:           newCounter(emitFailureCounterName),
		snapshotEmitSuccess:   newCounter(snapshotEmitSuccessCounterName),
		snapshotEmitFailure:   newCounter(snapshotEmitFailureCounterName),
		utilization:           gauge,
	}
	rm.Service, rm.srvcEng = services.Config{
		Name:  "ResourceManager",
		Start: rm.start,
		Close: rm.close,
	}.NewServiceEngine(lggr)
	rm.lggr = logger.Sugared(rm.srvcEng)
	return rm
}

// start launches the snapshot loop. The loop runs only when metering is
// enabled, an emitter is configured, and a positive SnapshotInterval is set;
// otherwise the service starts cleanly with snapshots disabled and
// EmitDelta / EmitUsage remain available.
func (rm *ResourceManager) start(_ context.Context) error {
	if !rm.meterSnapshotsEnabled || rm.emitter == nil {
		rm.lggr.Infow("snapshot loop disabled",
			"meterRecordsEnabled", rm.meterRecordsEnabled,
			"meterSnapshotsEnabled", rm.meterSnapshotsEnabled,
			"hasEmitter", rm.emitter != nil,
			"snapshotInterval", rm.snapshotInterval,
		)
		return nil
	}
	rm.srvcEng.Go(func(ctx context.Context) {
		// Align the first tick to the next wall-clock multiple of the interval,
		// then run on the interval boundary thereafter. For DonTime-synced
		// clocks this makes cross-node snapshot buckets agree structurally: the
		// emitted snapshot timestamp is the tick time truncated to the interval.
		now := rm.clock.Now()
		firstTick := now.Truncate(rm.snapshotInterval).Add(rm.snapshotInterval)
		timer := rm.clock.NewTimer(firstTick.Sub(now))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.Chan():
		}
		rm.emitSnapshots(ctx, firstTick)

		ticker := rm.clock.NewTicker(rm.snapshotInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case tickTime := <-ticker.Chan():
				rm.emitSnapshots(ctx, tickTime)
			}
		}
	})
	return nil
}

func (rm *ResourceManager) close() error { return nil }

// Register adds m to the snapshot registry and returns an idempotent function
// that removes it. Calling the returned function more than once is a no-op.
// The returned closure is safe for concurrent use.
func (rm *ResourceManager) Register(m Meterable) (unregister func()) {
	reg := &registration{m: m}

	rm.mu.Lock()
	rm.registrations[reg] = struct{}{}
	rm.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			rm.mu.Lock()
			delete(rm.registrations, reg)
			rm.mu.Unlock()
		})
	}
}

// snapshotRegistrantTimeoutFloor bounds how long a single Meterable may take
// (GetUtilization + marshal + emit) before it is skipped for this tick.
const snapshotRegistrantTimeoutFloor = 5 * time.Second

// emitSnapshots is the snapshot tick: it snapshots the registry under a read
// lock, then emits snapshots for every registered Meterable. tickTime is the
// interval boundary this tick covers; each snapshot timestamp is tickTime
// truncated to the snapshot interval so cross-node buckets agree.
//
// Each registrant runs under its own context deadline (SnapshotInterval/4,
// floored at 5s). A registrant that exceeds it is logged and skipped so one
// slow or stuck Meterable can never stall the tick for the others.
func (rm *ResourceManager) emitSnapshots(ctx context.Context, tickTime time.Time) {
	rm.mu.RLock()
	ms := make([]Meterable, 0, len(rm.registrations))
	for reg := range rm.registrations {
		ms = append(ms, reg.m)
	}
	rm.mu.RUnlock()

	timeout := rm.snapshotInterval / 4
	if timeout < snapshotRegistrantTimeoutFloor {
		timeout = snapshotRegistrantTimeoutFloor
	}

	for _, m := range ms {
		func(m Meterable) {
			rctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			rm.emitSnapshot(rctx, m, tickTime)
			if rctx.Err() != nil && ctx.Err() == nil {
				rm.lggr.Warnw("snapshot registrant exceeded its per-tick deadline; skipped",
					"timeout", timeout,
					"err", rctx.Err(),
				)
			}
		}(m)
	}
}

// emitSnapshot emits one MeterSnapshot per active resource reported by m. Each
// snapshot covers exactly one resource, fully identified by its
// ResourceIdentity. The snapshot timestamp is tickTime truncated to the
// snapshot interval, so retries of the same interval dedup and cross-node
// buckets agree. An empty utilization list emits nothing: a resource is
// released by its absence from later snapshots. Fail-open: per-resource errors
// are logged and counted, never returned.
func (rm *ResourceManager) emitSnapshot(ctx context.Context, m Meterable, tickTime time.Time) {
	if !rm.meterSnapshotsEnabled || rm.emitter == nil {
		return
	}

	ts := timestamppb.New(tickTime.Truncate(rm.snapshotInterval))
	interval := durationpb.New(rm.snapshotInterval)

	for _, e := range m.GetUtilization(ctx) {
		if len(e.Utilizations) == 0 {
			continue
		}
		// event_id is generated by the manager, never the producer: one fresh
		// UUIDv4 per snapshot emission, stamped on each utilization.
		eventID := uuid.NewString()
		for _, u := range e.Utilizations {
			if u != nil {
				u.EventId = eventID
			}
			rm.recordUtilization(ctx, e.Identity, u)
		}

		snapshot := &meteringpb.MeterSnapshot{
			Timestamp:   ts,
			Identity:    e.Identity.toProto(),
			Utilization: e.Utilizations,
			Interval:    interval,
		}

		body, err := proto.Marshal(snapshot)
		if err != nil {
			rm.lggr.Errorw("failed to marshal snapshot",
				"service", e.Identity.Service,
				"resourcePool", e.Identity.ResourcePool,
				"resourcePoolID", e.Identity.ResourcePoolID,
				"err", err,
			)
			rm.countSnapshot(ctx, rm.snapshotEmitFailure, e.Identity, attribute.String("reason", "marshal"))
			continue
		}

		if err := rm.emitter.Emit(ctx, body,
			beholder.AttrKeyDataSchema, snapshotDataSchema,
			beholder.AttrKeyDomain, domain,
			beholder.AttrKeyEntity, snapshotEntity,
		); err != nil {
			rm.lggr.Errorw("failed to emit snapshot",
				"service", e.Identity.Service,
				"resourcePool", e.Identity.ResourcePool,
				"resourcePoolID", e.Identity.ResourcePoolID,
				"err", err,
			)
			rm.countSnapshot(ctx, rm.snapshotEmitFailure, e.Identity, attribute.String("reason", "emit"))
			continue
		}

		rm.countSnapshot(ctx, rm.snapshotEmitSuccess, e.Identity)
	}
}

// EmitDelta emits a metering.v1.MeterRecord with METER_ACTION_UPDATE: a signed
// delta to the durable resource's level (register = +N, unregister = -N, resize
// = ±delta). delta may be negative. The record is timestamped with the raw
// clock (records are request-time edges, not bucket-aligned).
//
// EmitDelta is fail-open and returns no error: when the manager is disabled or
// has no emitter it does nothing, and marshal or emit failures are recorded
// only via error-level logs and the failure counter. Callers must never gate
// the resource operation being metered on emission.
func (rm *ResourceManager) EmitDelta(ctx context.Context, identity ResourceIdentity, delta int64, fields UtilizationFields) {
	rm.emitRecord(ctx, identity, meteringpb.MeterAction_METER_ACTION_UPDATE, NewUtilizationInt(delta, fields))
}

// EmitUsage emits a metering.v1.MeterRecord with METER_ACTION_USAGE: a one-off
// instantaneous consumption event billed per occurrence (no level, never
// snapshotted). Same fail-open semantics as EmitDelta.
func (rm *ResourceManager) EmitUsage(ctx context.Context, identity ResourceIdentity, quantity int64, fields UtilizationFields) {
	rm.emitRecord(ctx, identity, meteringpb.MeterAction_METER_ACTION_USAGE, NewUtilizationInt(quantity, fields))
}

// emitRecord builds and emits a single-utilization MeterRecord for action.
// event_id is generated here (a fresh UUIDv4 per emission), never supplied by
// the caller, and stamped on the utilization.
func (rm *ResourceManager) emitRecord(ctx context.Context, identity ResourceIdentity, action meteringpb.MeterAction, u *meteringpb.Utilization) {
	if !rm.meterRecordsEnabled {
		rm.lggr.Debugw("metering disabled; meter record not emitted",
			"service", identity.Service,
			"resourcePool", identity.ResourcePool,
			"action", action.String(),
		)
		return
	}

	if u != nil {
		u.EventId = uuid.NewString()
	}
	utilizations := []*meteringpb.Utilization{u}
	rm.recordUtilization(ctx, identity, u)

	if rm.emitter == nil {
		return
	}

	record := &meteringpb.MeterRecord{
		Timestamp:    timestamppb.New(rm.clock.Now()),
		Identity:     identity.toProto(),
		Action:       action,
		Utilizations: utilizations,
	}

	body, err := proto.Marshal(record)
	if err != nil {
		rm.lggr.Errorw("failed to marshal meter record",
			"service", identity.Service,
			"resourcePool", identity.ResourcePool,
			"action", action.String(),
			"err", err,
		)
		rm.countRecord(ctx, rm.emitFailure, identity, action, attribute.String("reason", "marshal"))
		return
	}

	if err := rm.emitter.Emit(ctx, body,
		beholder.AttrKeyDataSchema, dataSchema,
		beholder.AttrKeyDomain, domain,
		beholder.AttrKeyEntity, entity,
	); err != nil {
		rm.lggr.Errorw("failed to emit meter record",
			"service", identity.Service,
			"resourcePool", identity.ResourcePool,
			"action", action.String(),
			"err", err,
		)
		rm.countRecord(ctx, rm.emitFailure, identity, action, attribute.String("reason", "emit"))
		return
	}

	rm.countRecord(ctx, rm.emitSuccess, identity, action)
}

// recordUtilization records value to the per-resource utilization gauge,
// labeled with every ResourceIdentity dimension plus utilization identity.
func (rm *ResourceManager) recordUtilization(ctx context.Context, id ResourceIdentity, u *meteringpb.Utilization) {
	if rm.utilization == nil {
		return
	}
	if u == nil {
		return
	}

	value, err := strconv.ParseInt(u.GetValue(), 10, 64)
	if err != nil {
		rm.lggr.Debugw("skipping utilization gauge record for non-int64 value",
			"value", u.GetValue(),
			"resourceType", u.GetResourceType(),
			"resourceID", u.GetResourceId(),
			"orgID", u.GetOrgId(),
			"err", err,
		)
		return
	}
	rm.utilization.Record(ctx, value, metric.WithAttributes(
		attribute.String("product", id.Product),
		attribute.String("tenant", id.Tenant),
		attribute.String("numeric_tenant_id", id.NumericTenantID),
		attribute.String("environment", id.Environment),
		attribute.String("zone", id.Zone),
		attribute.String("don_id", id.DonID()),
		attribute.String("node_id", id.NodeID()),
		attribute.String("service", id.Service),
		attribute.String("resource_pool", id.ResourcePool),
		attribute.String("resource_pool_id", id.ResourcePoolID),
		attribute.String("resource_type", u.GetResourceType()),
		attribute.String("resource_id", u.GetResourceId()),
		attribute.String("org_id", u.GetOrgId()),
	))
}

// countRecord records a MeterRecord emit outcome. Labels are intentionally
// low-cardinality: service, resource, action (+ reason on failure). node_id
// and resource_id are deliberately excluded here to keep the time series
// bounded; the utilization gauge carries the full identity instead.
func (rm *ResourceManager) countRecord(ctx context.Context, c metric.Int64Counter, identity ResourceIdentity, action meteringpb.MeterAction, extra ...attribute.KeyValue) {
	if c == nil {
		return
	}
	attrs := append([]attribute.KeyValue{
		attribute.String("service", identity.Service),
		attribute.String("resource_pool", identity.ResourcePool),
		attribute.String("action", action.String()),
	}, extra...)
	c.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// countSnapshot records a MeterSnapshot emit outcome. Labels are intentionally
// low-cardinality: service, resource (+ reason on failure).
func (rm *ResourceManager) countSnapshot(ctx context.Context, c metric.Int64Counter, identity ResourceIdentity, extra ...attribute.KeyValue) {
	if c == nil {
		return
	}
	attrs := append([]attribute.KeyValue{
		attribute.String("service", identity.Service),
		attribute.String("resource_pool", identity.ResourcePool),
	}, extra...)
	c.Add(ctx, 1, metric.WithAttributes(attrs...))
}
