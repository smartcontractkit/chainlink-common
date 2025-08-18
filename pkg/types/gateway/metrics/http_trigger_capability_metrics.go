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

// HTTPTriggerCapabilityMetrics contains metrics for HTTP triggers on capability nodes
type HTTPTriggerCapabilityMetrics struct {
	registerCount             metric.Int64Counter
	deregisterCount           metric.Int64Counter
	registerFailureCount      metric.Int64Counter
	deregisterFailureCount    metric.Int64Counter
	requestCacheCleanUpCount  metric.Int64Counter
	requestCount              metric.Int64Counter
	gatewayGlobalThrottled    metric.Int64Counter
	gatewayNodeThrottled      metric.Int64Counter
	requestSuccessCount       metric.Int64Counter
	gatewaySendError          metric.Int64Counter
	broadcastMetadataCount    metric.Int64Counter
	broadcastMetadataFailures metric.Int64Counter
	broadcastMetadataLatency  metric.Int64Histogram
	pullMetadataCount         metric.Int64Counter
	pullMetadataFailures      metric.Int64Counter
	pullMetadataLatency       metric.Int64Histogram
	requestLatency            metric.Int64Histogram

	once sync.Once
	err  error
}

var httpTriggerCapabilityMetrics = &HTTPTriggerCapabilityMetrics{}

func (m *HTTPTriggerCapabilityMetrics) init() {
	meter := beholder.GetMeter()

	m.registerCount, m.err = meter.Int64Counter(
		"http_trigger_register_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger registrations"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger register count metric: %w", m.err)
		return
	}

	m.deregisterCount, m.err = meter.Int64Counter(
		"http_trigger_deregister_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger deregistrations"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger deregister count metric: %w", m.err)
		return
	}

	m.registerFailureCount, m.err = meter.Int64Counter(
		"http_trigger_register_failure_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger registration failures"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger register failure count metric: %w", m.err)
		return
	}

	m.deregisterFailureCount, m.err = meter.Int64Counter(
		"http_trigger_deregister_failure_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger deregistration failures"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger deregister failure count metric: %w", m.err)
		return
	}

	m.requestCacheCleanUpCount, m.err = meter.Int64Counter(
		"http_trigger_request_cache_cleanup_count",
		metric.WithDescription("Capability node metric. Number of expired entries cleaned up from HTTP trigger request cache"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger request cache cleanup count metric: %w", m.err)
		return
	}

	m.requestCount, m.err = meter.Int64Counter(
		"http_trigger_capability_request_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger requests processed"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability request count metric: %w", m.err)
		return
	}

	m.gatewayGlobalThrottled, m.err = meter.Int64Counter(
		"http_trigger_capability_gateway_global_throttled",
		metric.WithDescription("Capability node metric. Number of HTTP trigger requests throttled due to global rate limit"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability gateway global throttled metric: %w", m.err)
		return
	}

	m.gatewayNodeThrottled, m.err = meter.Int64Counter(
		"http_trigger_capability_gateway_node_throttled",
		metric.WithDescription("Capability node metric. Number of HTTP trigger requests throttled due to per-node rate limit"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability gateway node throttled metric: %w", m.err)
		return
	}

	m.requestSuccessCount, m.err = meter.Int64Counter(
		"http_trigger_capability_request_success_count",
		metric.WithDescription("Capability node metric. Number of successful HTTP trigger responses sent to gateway"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability request success count metric: %w", m.err)
		return
	}

	m.gatewaySendError, m.err = meter.Int64Counter(
		"http_trigger_capability_gateway_send_error",
		metric.WithDescription("Capability node metric. Number of HTTP trigger gateway send errors"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability gateway send error metric: %w", m.err)
		return
	}

	m.broadcastMetadataCount, m.err = meter.Int64Counter(
		"http_trigger_capability_broadcast_metadata_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger broadcast metadata workflow operations"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability broadcast metadata count metric: %w", m.err)
		return
	}

	m.broadcastMetadataFailures, m.err = meter.Int64Counter(
		"http_trigger_capability_broadcast_metadata_failures",
		metric.WithDescription("Capability node metric. Number of HTTP trigger broadcast metadata workflow failures"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability broadcast metadata failures metric: %w", m.err)
		return
	}

	m.broadcastMetadataLatency, m.err = meter.Int64Histogram(
		"http_trigger_capability_broadcast_metadata_latency_ms",
		metric.WithDescription("Capability node metric. HTTP trigger broadcast metadata latency in milliseconds"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability broadcast metadata latency metric: %w", m.err)
		return
	}

	m.pullMetadataCount, m.err = meter.Int64Counter(
		"http_trigger_capability_pull_metadata_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger pull metadata workflow operations"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability pull metadata count metric: %w", m.err)
		return
	}

	m.pullMetadataFailures, m.err = meter.Int64Counter(
		"http_trigger_capability_pull_metadata_failures",
		metric.WithDescription("Capability node metric. Number of HTTP trigger pull metadata workflow failures"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability pull metadata failures metric: %w", m.err)
		return
	}

	m.pullMetadataLatency, m.err = meter.Int64Histogram(
		"http_trigger_capability_pull_metadata_latency_ms",
		metric.WithDescription("Capability node metric. HTTP trigger pull metadata latency in milliseconds"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability pull metadata latency metric: %w", m.err)
		return
	}

	m.requestLatency, m.err = meter.Int64Histogram(
		"http_trigger_capability_request_latency_ms",
		metric.WithDescription("Capability node metric. HTTP trigger capability request processing latency in milliseconds"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger capability request latency metric: %w", m.err)
		return
	}
}

func IncrementHTTPTriggerRegisterCount(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger register count metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.registerCount.Add(ctx, 1)
}

func IncrementHTTPTriggerDeregisterCount(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger deregister count metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.deregisterCount.Add(ctx, 1)
}

func IncrementHTTPTriggerRegisterFailureCount(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger register failure count metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.registerFailureCount.Add(ctx, 1)
}

func IncrementHTTPTriggerDeregisterFailureCount(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger deregister failure count metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.deregisterFailureCount.Add(ctx, 1)
}

func IncrementHTTPTriggerRequestCacheCleanUpCount(ctx context.Context, count int64, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger request cache cleanup count metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.requestCacheCleanUpCount.Add(ctx, count)
}

func IncrementHTTPTriggerCapabilityRequestCount(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability request count metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.requestCount.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityGatewayGlobalThrottled(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability gateway global throttled metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.gatewayGlobalThrottled.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityGatewayNodeThrottled(ctx context.Context, nodeAddress string, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability gateway node throttled metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.gatewayNodeThrottled.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrNodeAddress, nodeAddress)))
}

func IncrementHTTPTriggerCapabilityRequestSuccessCount(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability request success count metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.requestSuccessCount.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityGatewaySendError(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability gateway send error metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.gatewaySendError.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityBroadcastMetadataCount(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability broadcast metadata count metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.broadcastMetadataCount.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityBroadcastMetadataFailures(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability broadcast metadata failures metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.broadcastMetadataFailures.Add(ctx, 1)
}

func RecordHTTPTriggerCapabilityBroadcastMetadataLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability broadcast metadata latency metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.broadcastMetadataLatency.Record(ctx, latencyMs)
}

func IncrementHTTPTriggerCapabilityPullMetadataCount(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability pull metadata count metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.pullMetadataCount.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityPullMetadataFailures(ctx context.Context, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability pull metadata failures metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.pullMetadataFailures.Add(ctx, 1)
}

func RecordHTTPTriggerCapabilityPullMetadataLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability pull metadata latency metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.pullMetadataLatency.Record(ctx, latencyMs)
}

func RecordHTTPTriggerCapabilityRequestLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	httpTriggerCapabilityMetrics.once.Do(httpTriggerCapabilityMetrics.init)
	if httpTriggerCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability request latency metric", "error", httpTriggerCapabilityMetrics.err)
		}
		return
	}
	httpTriggerCapabilityMetrics.requestLatency.Record(ctx, latencyMs)
}
