// Package resourcemanager emits metering.v1 events for durable billable
// resources such as trigger registrations, workflow specs, and log filters.
//
// It emits two kinds of records, each covering exactly one resource identified
// by its ResourceIdentity:
//
//   - MeterRecord lifecycle edges, inline, via EmitMeterRecord at the point a
//     resource is reserved, released, or its utilization changes; and
//   - MeterSnapshot records, on a timer, one per resource a registered
//     Meterable reports as currently active. Snapshots are the
//     liveness/utilization-over-time signal that pure RESERVE/RELEASE cannot
//     provide (a node panic would otherwise leak a reservation forever).
//
// The ResourceManager is the single owner of the snapshot tick: each producer
// starts the manager as a sub-service and only Registers itself; producers
// never run their own snapshot loop.
//
// Emission is fail-open by design: EmitMeterRecord and the snapshot loop
// return no error, and a metering failure must never gate, delay, or retry the
// resource operation being metered. Failures surface via error-level logs and
// the resource_manager_*_failure_total counters; billing correctness is
// recovered downstream through idempotency keys and reconciliation.
package resourcemanager

import (
	"context"
	"strconv"
	"sync"
	"time"

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

// Beholder routing attributes. Each entity value is the contract shared with
// the CHiP schema registration and the consumer topic name; all must match
// exactly.
const (
	beholderDomain             = "platform"
	beholderEntity             = "metering.v1.MeterRecord"
	beholderDataSchema         = "metering.v1.meter_record"
	beholderSnapshotEntity     = "metering.v1.MeterSnapshot"
	beholderSnapshotDataSchema = "metering.v1.meter_snapshot"
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
	// default), EmitMeterRecord is a no-op.
	MeterRecordsEnabled bool

	// MeterSnapshotsEnabled gates snapshot emission. Snapshots are only emitted
	// when this is true AND MeterRecordsEnabled is true.
	MeterSnapshotsEnabled bool

	// Emitter delivers encoded records, typically beholder.GetEmitter(). A nil
	// Emitter makes EmitMeterRecord a no-op even when Enabled is true and keeps
	// the snapshot loop from starting.
	Emitter Emitter

	// SnapshotInterval is the period between snapshots. Zero (the default)
	// DISABLES the snapshot loop; the manager still starts as a service and
	// EmitMeterRecord still works. Callers that want snapshots set a positive
	// value, e.g. DefaultSnapshotInterval. The default is not substituted for
	// zero — zero means off.
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
// EmitMeterRecord works regardless of Start.
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
// EmitMeterRecord remains available.
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
		ticker := rm.clock.NewTicker(rm.snapshotInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.Chan():
				rm.emitSnapshots(ctx)
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

// emitSnapshots is the snapshot tick: it snapshots the registry under a read
// lock, then emits snapshots for every registered Meterable.
func (rm *ResourceManager) emitSnapshots(ctx context.Context) {
	rm.mu.RLock()
	ms := make([]Meterable, 0, len(rm.registrations))
	for reg := range rm.registrations {
		ms = append(ms, reg.m)
	}
	rm.mu.RUnlock()

	for _, m := range ms {
		rm.emitSnapshot(ctx, m)
	}
}

// emitSnapshot emits one MeterSnapshot per active resource reported by m. Each
// snapshot covers exactly one resource, fully identified by its
// ResourceIdentity. The interval bucket (timestamp truncated to the snapshot
// interval) keys the snapshot per interval. An empty utilization list emits
// nothing: billing zeroes a resource out by its absence from later snapshots.
// Fail-open: per-resource errors are logged and counted, never returned.
func (rm *ResourceManager) emitSnapshot(ctx context.Context, m Meterable) {
	if !rm.meterSnapshotsEnabled || rm.emitter == nil {
		return
	}

	now := rm.clock.Now()
	ts := timestamppb.New(now)
	interval := durationpb.New(rm.snapshotInterval)

	for _, e := range m.GetUtilization(ctx) {
		if len(e.Utilizations) == 0 {
			continue
		}
		for _, u := range e.Utilizations {
			rm.recordUtilization(ctx, e.Identity, u, meteringpb.MeterAction_METER_ACTION_UPDATE)
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
			beholder.AttrKeyDataSchema, beholderSnapshotDataSchema,
			beholder.AttrKeyDomain, beholderDomain,
			beholder.AttrKeyEntity, beholderSnapshotEntity,
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

// EmitMeterRecord emits a metering.v1.MeterRecord, timestamped now, for action
// on the one resource described by identity.
//
// EmitMeterRecord is fail-open and returns no error: when the manager is
// disabled or has no emitter it does nothing, and marshal or emit failures are
// recorded only via error-level logs and the failure counter. Callers must
// never gate resource allocation or deallocation on emission.
func (rm *ResourceManager) EmitMeterRecord(ctx context.Context, identity ResourceIdentity, action meteringpb.MeterAction, utilizations []*meteringpb.Utilization) {
	if !rm.meterRecordsEnabled {
		rm.lggr.Debugw("metering disabled; meter record not emitted",
			"service", identity.Service,
			"resourcePool", identity.ResourcePool,
			"action", action.String(),
		)
		return
	}

	for _, u := range utilizations {
		rm.recordUtilization(ctx, identity, u, action)
	}

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
		beholder.AttrKeyDataSchema, beholderDataSchema,
		beholder.AttrKeyDomain, beholderDomain,
		beholder.AttrKeyEntity, beholderEntity,
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
func (rm *ResourceManager) recordUtilization(ctx context.Context, id ResourceIdentity, u *meteringpb.Utilization, action meteringpb.MeterAction) {
	if rm.utilization == nil {
		return
	}
	if u == nil {
		return
	}

	var value int64
	if action == meteringpb.MeterAction_METER_ACTION_RELEASE {
		value = 0
	} else {
		parsed, err := strconv.ParseInt(u.GetValue(), 10, 64)
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
		value = parsed
	}
	rm.utilization.Record(ctx, value, metric.WithAttributes(
		attribute.String("product", id.Product),
		attribute.String("environment", id.Environment),
		attribute.String("zone", id.Zone),
		attribute.String("don_id", id.DONID),
		attribute.String("node_id", id.NodeID),
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
