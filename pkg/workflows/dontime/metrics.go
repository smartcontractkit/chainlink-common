package dontime

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type BeholderMetrics struct {
	responsesDelivered metric.Int64Counter
	pendingRequests    metric.Int64Gauge
}

var sharedDontimeMetrics struct {
	once sync.Once
	m    *BeholderMetrics
}

func sharedBeholderMetrics(lggr logger.Logger) *BeholderMetrics {
	sharedDontimeMetrics.once.Do(func() {
		var err error
		sharedDontimeMetrics.m, err = NewBeholderMetrics()
		if err != nil {
			lggr.Warnw("failed to initialize dontime beholder metrics; continuing without", "err", err)
		}
	})
	return sharedDontimeMetrics.m
}

func NewBeholderMetrics() (*BeholderMetrics, error) {
	responsesDelivered, err := beholder.GetMeter().Int64Counter(
		"dontime_responses_delivered_total",
		metric.WithDescription("Total DON time responses delivered to callers after OCR3 transmit"),
	)
	if err != nil {
		return nil, err
	}
	pendingRequests, err := beholder.GetMeter().Int64Gauge(
		"dontime_pending_requests_in_store",
		metric.WithDescription("Pending DON time requests in the local store"),
	)
	if err != nil {
		return nil, err
	}
	return &BeholderMetrics{
		responsesDelivered: responsesDelivered,
		pendingRequests:    pendingRequests,
	}, nil
}

func (m *BeholderMetrics) AddResponsesDelivered(ctx context.Context, n int64) {
	if m == nil || n == 0 {
		return
	}
	m.responsesDelivered.Add(ctx, n)
}

func (m *BeholderMetrics) SetPendingRequestsInStore(ctx context.Context, n int64) {
	if m == nil {
		return
	}
	m.pendingRequests.Record(ctx, n)
}
