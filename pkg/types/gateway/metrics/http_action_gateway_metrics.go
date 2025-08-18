package metrics

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// HTTPActionGatewayMetrics contains metrics for HTTP actions on gateway nodes
type HTTPActionGatewayMetrics struct {
	requestCount                   metric.Int64Counter
	requestFailures                metric.Int64Counter
	capabilityNodeThrottled        metric.Int64Counter
	gatewayGlobalThrottled         metric.Int64Counter
	requestLatency                 metric.Int64Histogram
	customerEndpointRequestLatency metric.Int64Histogram
	customerEndpointResponseCount  metric.Int64Counter
	cacheReadCount                 metric.Int64Counter
	cacheHitCount                  metric.Int64Counter
	cacheCleanUpCount              metric.Int64Counter
	cacheSize                      metric.Int64Gauge
	capabilityRequestCount         metric.Int64Counter
	capabilityFailures             metric.Int64Counter

	once sync.Once
	err  error
}

var httpActionGatewayMetrics = &HTTPActionGatewayMetrics{}

func (m *HTTPActionGatewayMetrics) init() {
	meter := beholder.GetMeter()

	m.requestCount, m.err = meter.Int64Counter(
		"http_action_gateway_request_count",
		metric.WithDescription("Gateway node metric. Number of HTTP action requests received by the gateway"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action gateway request count metric: %w", m.err)
		return
	}

	m.requestFailures, m.err = meter.Int64Counter(
		"http_action_gateway_request_failures",
		metric.WithDescription("Gateway node metric. Number of HTTP action request failures in the gateway"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action gateway request failures metric: %w", m.err)
		return
	}

	m.capabilityNodeThrottled, m.err = meter.Int64Counter(
		"http_action_gateway_capability_node_throttled",
		metric.WithDescription("Gateway node metric. Number of HTTP action gateway requests throttled due to per-capability-node rate limit"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action gateway capability node throttled metric: %w", m.err)
		return
	}

	m.gatewayGlobalThrottled, m.err = meter.Int64Counter(
		"http_action_gateway_global_throttled",
		metric.WithDescription("Gateway node metric. Number of HTTP action gateway requests throttled due to global rate limit"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action gateway global throttled metric: %w", m.err)
		return
	}

	m.requestLatency, m.err = meter.Int64Histogram(
		"http_action_gateway_request_latency_ms",
		metric.WithDescription("Gateway node metric. HTTP action request latency in milliseconds in the gateway"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action gateway request latency metric: %w", m.err)
		return
	}

	m.customerEndpointRequestLatency, m.err = meter.Int64Histogram(
		"http_action_customer_endpoint_request_latency_ms",
		metric.WithDescription("Gateway node metric. Request latency while calling customer endpoint in milliseconds"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action customer endpoint request latency metric: %w", m.err)
		return
	}

	m.customerEndpointResponseCount, m.err = meter.Int64Counter(
		"http_action_customer_endpoint_response_count",
		metric.WithDescription("Gateway node metric. Number of external calls to customer endpoints"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action customer endpoint response count metric: %w", m.err)
		return
	}

	m.cacheReadCount, m.err = meter.Int64Counter(
		"http_action_cache_read_count",
		metric.WithDescription("Gateway node metric. Number of HTTP action cache read operations"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action response cache read count metric: %w", m.err)
		return
	}

	m.cacheHitCount, m.err = meter.Int64Counter(
		"http_action_cache_hit_count",
		metric.WithDescription("Gateway node metric. Number of HTTP action response cache hits"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action cache hit count metric: %w", m.err)
		return
	}

	m.cacheCleanUpCount, m.err = meter.Int64Counter(
		"http_action_cache_cleanup_count",
		metric.WithDescription("Gateway node metric. Number of HTTP action response cache entries cleaned up"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action cache cleanup count metric: %w", m.err)
		return
	}

	m.cacheSize, m.err = meter.Int64Gauge(
		"http_action_cache_size",
		metric.WithDescription("Gateway node metric. Current number of entries in HTTP action response cache"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action cache size metric: %w", m.err)
		return
	}

	m.capabilityRequestCount, m.err = meter.Int64Counter(
		"http_action_gateway_capability_request_count",
		metric.WithDescription("Gateway node metric. Number of gateway responses to the capability nodes for HTTP action"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action gateway capability request count metric: %w", m.err)
		return
	}

	m.capabilityFailures, m.err = meter.Int64Counter(
		"http_action_gateway_capability_failures",
		metric.WithDescription("Gateway node metric. Number of errors while responding to the capability nodes for HTTP action"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP action gateway capability failures metric: %w", m.err)
		return
	}
}

func IncrementHTTPActionGatewayRequestCount(ctx context.Context, nodeAddress string, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action gateway request count metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.requestCount.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrNodeAddress, nodeAddress)))
}

func IncrementHTTPActionGatewayRequestFailures(ctx context.Context, nodeAddress string, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action gateway request failures metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.requestFailures.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrNodeAddress, nodeAddress)))
}

func IncrementHTTPActionGatewayCapabilityNodeThrottled(ctx context.Context, nodeAddress string, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action gateway capability node throttled metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.capabilityNodeThrottled.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrNodeAddress, nodeAddress)))
}

func IncrementHTTPActionGatewayGlobalThrottled(ctx context.Context, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action gateway global throttled metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.gatewayGlobalThrottled.Add(ctx, 1)
}

func RecordHTTPActionGatewayRequestLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action gateway request latency metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.requestLatency.Record(ctx, latencyMs)
}

func RecordHTTPActionCustomerEndpointRequestLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action customer endpoint request latency metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.customerEndpointRequestLatency.Record(ctx, latencyMs)
}

func IncrementHTTPActionCustomerEndpointResponseCount(ctx context.Context, statusCode string, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action customer endpoint response count metric. Does not count retries", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.customerEndpointResponseCount.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrStatusCode, statusCode)))
}

func IncrementHTTPActionCacheReadCount(ctx context.Context, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action cache read count metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.cacheReadCount.Add(ctx, 1)
}

func IncrementHTTPActionCacheHitCount(ctx context.Context, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action cache hit count metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.cacheHitCount.Add(ctx, 1)
}

func IncrementHTTPActionCacheCleanUpCount(ctx context.Context, count int64, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action cache cleanup count metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.cacheCleanUpCount.Add(ctx, count)
}

func RecordHTTPActionCacheSize(ctx context.Context, size int64, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action cache size metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.cacheSize.Record(ctx, size)
}

func IncrementHTTPActionGatewayCapabilityRequestCount(ctx context.Context, nodeAddress string, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action gateway capability request count metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.capabilityRequestCount.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrNodeAddress, nodeAddress)))
}

func IncrementHTTPActionGatewayCapabilityFailures(ctx context.Context, nodeAddress string, lggr logger.Logger) {
	httpActionGatewayMetrics.once.Do(httpActionGatewayMetrics.init)
	if httpActionGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action gateway capability failures metric", "error", httpActionGatewayMetrics.err)
		}
		return
	}
	httpActionGatewayMetrics.capabilityFailures.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrNodeAddress, nodeAddress)))
}
