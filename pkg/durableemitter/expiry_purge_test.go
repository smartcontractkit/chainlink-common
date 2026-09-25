package durableemitter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

const expiredPurgedMetric = "durable_emitter.expired_purged"

// makeCloudEventPayloadFrom is makeCloudEventPayload with an explicit domain
// (CloudEvent source), so purge attribution can be asserted per domain.
func makeCloudEventPayloadFrom(t *testing.T, domain, entity, body string) []byte {
	t.Helper()
	ev, err := chipingress.NewEvent(domain, entity, []byte(body), nil)
	require.NoError(t, err)
	evPb, err := chipingress.EventToProto(ev)
	require.NoError(t, err)
	payload, err := proto.Marshal(evPb)
	require.NoError(t, err)
	return payload
}

// waitForCounter polls the reader until the counter (filtered by want) reaches
// n. The expiry loop records the metric after the delete completes, so an empty
// store is not yet proof the count has landed.
func waitForCounter(t *testing.T, ctx context.Context, collect func(context.Context) metricdata.ResourceMetrics, name string, want map[string]string, n int64) metricdata.ResourceMetrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	require.Eventually(t, func() bool {
		rm = collect(ctx)
		return counterSumByAttrs(t, rm, name, want) >= n
	}, 3*time.Second, 10*time.Millisecond, "counter %s%v should reach %d", name, want, n)
	return rm
}

// counterSumByAttrs sums the data points of an int64 counter whose attributes
// include every key/value in want.
func counterSumByAttrs(t *testing.T, rm metricdata.ResourceMetrics, name string, want map[string]string) int64 {
	t.Helper()
	var total int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			require.True(t, ok, "expected Sum[int64] for %s", name)
			for _, dp := range sum.DataPoints {
				got := map[string]string{}
				for _, kv := range dp.Attributes.ToSlice() {
					got[string(kv.Key)] = kv.Value.AsString()
				}
				match := true
				for k, v := range want {
					if got[k] != v {
						match = false
						break
					}
				}
				if match {
					total += dp.Value
				}
			}
		}
	}
	return total
}

// insertAged inserts a payload and backdates it so the next expiry tick treats
// it as expired regardless of the configured TTL.
func insertAged(t *testing.T, store *MemDurableEventStore, payload []byte, age time.Duration) {
	t.Helper()
	id, err := store.Insert(context.Background(), payload)
	require.NoError(t, err)
	store.mu.Lock()
	store.events[id].CreatedAt = time.Now().Add(-age)
	store.mu.Unlock()
}

// plainStore hides ExpiredPurger from MemDurableEventStore so the emitter takes
// the unattributed DeleteExpired fallback.
type plainStore struct{ DurableEventStore }

func newPurgeTestEmitter(t *testing.T, store DurableEventStore, cfg Config) (*DurableEmitter, func(context.Context) metricdata.ResourceMetrics) {
	t.Helper()
	meter, reader := newTestMeter(t)
	cfg.RetransmitInterval = time.Hour // keep retransmit out of the picture
	cfg.Metrics = &DurableEmitterMetricsConfig{PollInterval: time.Hour}
	em, err := NewDurableEmitter(store, newTestBatchEmitter(), true, cfg, logger.Test(t), meter)
	require.NoError(t, err)
	collect := func(ctx context.Context) metricdata.ResourceMetrics {
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))
		return rm
	}
	return em, collect
}

// A purge must say what was lost: expired events are counted per CloudEvent
// source (domain) and type (subject), and an undecodable payload lands under
// unknown/unknown rather than vanishing from the count.
func TestDurableEmitter_ExpiryPurgeAttributesByDomainAndSubject(t *testing.T) {
	store := NewMemDurableEventStore()
	for range 2 {
		insertAged(t, store, makeCloudEventPayloadFrom(t, "platform", "workflow.execution", "a"), time.Hour)
	}
	insertAged(t, store, makeCloudEventPayloadFrom(t, "billing", "meter.record", "b"), time.Hour)
	insertAged(t, store, []byte("\xff\xfe not a proto"), time.Hour)

	cfg := DefaultConfig()
	cfg.ExpiryInterval = 20 * time.Millisecond
	cfg.EventTTL = time.Minute
	em, collect := newPurgeTestEmitter(t, store, cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	rm := waitForCounter(t, ctx, collect, expiredPurgedMetric, nil, 4)
	assert.Equal(t, 0, store.Len(), "expiry loop should purge all four aged events")
	assert.Equal(t, int64(2), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"domain": "platform", "subject": "workflow.execution"}))
	assert.Equal(t, int64(1), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"domain": "billing", "subject": "meter.record"}))
	assert.Equal(t, int64(1), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"domain": purgeUnknown, "subject": purgeUnknown}),
		"undecodable payload is still counted, under unknown attribution")
	assert.Equal(t, int64(4), counterSumByAttrs(t, rm, expiredPurgedMetric, nil), "every purged event is counted exactly once")
}

// One tick drains the whole expired backlog in bounded batches, so the count
// is complete even when the backlog exceeds ExpiryBatchSize.
func TestDurableEmitter_ExpiryPurgeDrainsBacklogInBatches(t *testing.T) {
	store := NewMemDurableEventStore()
	const n = 7
	for range n {
		insertAged(t, store, makeCloudEventPayloadFrom(t, "platform", "workflow.execution", "x"), time.Hour)
	}

	cfg := DefaultConfig()
	cfg.ExpiryInterval = 20 * time.Millisecond
	cfg.EventTTL = time.Minute
	cfg.ExpiryBatchSize = 3 // 7 rows -> 3 + 3 + 1
	em, collect := newPurgeTestEmitter(t, store, cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	rm := waitForCounter(t, ctx, collect, expiredPurgedMetric, nil, n)
	assert.Equal(t, 0, store.Len())
	assert.Equal(t, int64(n), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"domain": "platform", "subject": "workflow.execution"}))
}

// Events younger than EventTTL survive a tick; only the expired ones are counted.
func TestDurableEmitter_ExpiryPurgeLeavesFreshEventsAlone(t *testing.T) {
	store := NewMemDurableEventStore()
	insertAged(t, store, makeCloudEventPayloadFrom(t, "platform", "old", "o"), time.Hour)
	insertAged(t, store, makeCloudEventPayloadFrom(t, "platform", "fresh", "f"), 0)

	cfg := DefaultConfig()
	cfg.ExpiryInterval = 20 * time.Millisecond
	cfg.EventTTL = time.Minute
	em, collect := newPurgeTestEmitter(t, store, cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	rm := waitForCounter(t, ctx, collect, expiredPurgedMetric, map[string]string{"subject": "old"}, 1)
	time.Sleep(60 * time.Millisecond) // a couple more ticks: the fresh row must stay
	assert.Equal(t, 1, store.Len())
	rm = collect(ctx)
	assert.Equal(t, int64(1), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"subject": "old"}))
	assert.Equal(t, int64(0), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"subject": "fresh"}))
}

// A store that cannot return payloads still purges and still counts, with
// unknown attribution, so the metric never goes dark on a legacy store.
func TestDurableEmitter_ExpiryPurgeFallsBackWithoutPurger(t *testing.T) {
	mem := NewMemDurableEventStore()
	for range 3 {
		insertAged(t, mem, makeCloudEventPayloadFrom(t, "platform", "workflow.execution", "x"), time.Hour)
	}
	store := plainStore{mem}
	_, isPurger := DurableEventStore(store).(ExpiredPurger)
	require.False(t, isPurger, "test precondition: wrapper must hide ExpiredPurger")

	cfg := DefaultConfig()
	cfg.ExpiryInterval = 20 * time.Millisecond
	cfg.EventTTL = time.Minute
	em, collect := newPurgeTestEmitter(t, store, cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	rm := waitForCounter(t, ctx, collect, expiredPurgedMetric, nil, 3)
	assert.Equal(t, 0, mem.Len())
	assert.Equal(t, int64(3), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"domain": purgeUnknown, "subject": purgeUnknown}))
}

// The Postgres store advertises ExpiredPurger so production nodes take the
// attributed path; the query itself is exercised by the DB-backed tests in
// chainlink/core/services/durableemitter.
func TestPgDurableEventStore_ImplementsExpiredPurger(t *testing.T) {
	var s DurableEventStore = NewPgDurableEventStore(nil)
	_, ok := s.(ExpiredPurger)
	assert.True(t, ok)
}

// proto.Unmarshal accepts arbitrary bytes as an empty message, so a payload that
// decodes but carries no source/type must still be attributed as unknown, not
// as an empty label (review finding on #2411).
func TestDurableEmitter_ExpiryPurgeFieldlessPayloadIsUnknown(t *testing.T) {
	fieldless, err := proto.Marshal(&chipingress.CloudEventPb{Id: "no-source-no-type"})
	require.NoError(t, err)
	onlySource, err := proto.Marshal(&chipingress.CloudEventPb{Id: "x", Source: "platform"})
	require.NoError(t, err)

	store := NewMemDurableEventStore()
	insertAged(t, store, fieldless, time.Hour)
	insertAged(t, store, onlySource, time.Hour)
	insertAged(t, store, []byte{}, time.Hour) // empty payload decodes as an empty message too

	cfg := DefaultConfig()
	cfg.ExpiryInterval = 20 * time.Millisecond
	cfg.EventTTL = time.Minute
	em, collect := newPurgeTestEmitter(t, store, cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	rm := waitForCounter(t, ctx, collect, expiredPurgedMetric, nil, 3)
	assert.Equal(t, int64(2), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"domain": purgeUnknown, "subject": purgeUnknown}))
	assert.Equal(t, int64(1), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"domain": "platform", "subject": purgeUnknown}),
		"a present source is kept, the missing type is reported as unknown")
	assert.Equal(t, int64(0), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"domain": ""}), "no empty-string labels")
	assert.Equal(t, int64(0), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"subject": ""}), "no empty-string labels")
}

// A shutdown mid-drain stops deleting but must still record the batches that
// were already deleted; those rows are gone from the store regardless (review
// finding on #2411).
func TestDurableEmitter_ExpiryPurgeRecordsPartialDrainOnShutdown(t *testing.T) {
	store := NewMemDurableEventStore()
	const n = 7
	for range n {
		insertAged(t, store, makeCloudEventPayloadFrom(t, "platform", "workflow.execution", "x"), time.Hour)
	}

	cfg := DefaultConfig()
	cfg.EventTTL = time.Minute
	cfg.ExpiryBatchSize = 3
	em, collect := newPurgeTestEmitter(t, store, cfg)
	// Not started: drive one pass by hand with the stop signal already raised,
	// so the loop deletes exactly one batch and then observes shutdown.
	close(em.stopCh)
	em.purgeExpired(t.Context())

	assert.Equal(t, n-3, store.Len(), "one batch deleted before shutdown was observed")
	rm := collect(t.Context())
	assert.Equal(t, int64(3), counterSumByAttrs(t, rm, expiredPurgedMetric, map[string]string{"domain": "platform", "subject": "workflow.execution"}),
		"the deleted batch is counted even though the drain was interrupted")
}
