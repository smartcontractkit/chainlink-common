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

// HTTPTriggerGatewayMetrics contains metrics for HTTP triggers on gateway nodes
type HTTPTriggerGatewayMetrics struct {
	requestCount                         metric.Int64Counter
	requestErrors                        metric.Int64Counter
	workflowOwnerThrottled               metric.Int64Counter
	globalThrottled                      metric.Int64Counter
	pendingRequestsCleanUpCount          metric.Int64Counter
	pendingRequestsCount                 metric.Int64Gauge
	requestHandlerLatency                metric.Int64Histogram
	capabilityRequestCount               metric.Int64Counter
	capabilityRequestFailures            metric.Int64Counter
	capabilityMetadataProcessingFailures metric.Int64Counter
	capabilityMetadataRequestCount       metric.Int64Counter
	metadataObservationsCleanUpCount     metric.Int64Counter
	metadataObservationsCount            metric.Int64Gauge

	once sync.Once
	err  error
}

var httpTriggerGatewayMetrics = &HTTPTriggerGatewayMetrics{}

func (m *HTTPTriggerGatewayMetrics) init() {
	meter := beholder.GetMeter()

	m.requestCount, m.err = meter.Int64Counter(
		"http_trigger_gateway_request_count",
		metric.WithDescription("Gateway node metric. Number of user HTTP trigger requests received by the gateway"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway request count metric: %w", m.err)
		return
	}

	m.requestErrors, m.err = meter.Int64Counter(
		"http_trigger_gateway_request_errors",
		metric.WithDescription("Gateway node metric. Number of HTTP trigger gateway request errors"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway request errors metric: %w", m.err)
		return
	}

	m.workflowOwnerThrottled, m.err = meter.Int64Counter(
		"http_trigger_gateway_workflow_owner_throttled",
		metric.WithDescription("Gateway node metric. Number of HTTP trigger gateway requests throttled per workflow owner"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway workflow owner throttled metric: %w", m.err)
		return
	}

	m.globalThrottled, m.err = meter.Int64Counter(
		"http_trigger_gateway_global_throttled",
		metric.WithDescription("Gateway node metric. Number of HTTP trigger gateway requests throttled globally"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway global throttled metric: %w", m.err)
		return
	}

	m.pendingRequestsCleanUpCount, m.err = meter.Int64Counter(
		"http_trigger_gateway_pending_requests_cleanup_count",
		metric.WithDescription("Gateway node metric. Number of pending HTTP trigger gateway requests cleaned up"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway pending requests cleanup count metric: %w", m.err)
		return
	}

	m.pendingRequestsCount, m.err = meter.Int64Gauge(
		"http_trigger_gateway_pending_requests_count",
		metric.WithDescription("Gateway node metric. Current number of pending HTTP trigger gateway requests"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway pending requests count metric: %w", m.err)
		return
	}

	m.requestHandlerLatency, m.err = meter.Int64Histogram(
		"http_trigger_gateway_request_handler_latency_ms",
		metric.WithDescription("Gateway node metric. HTTP trigger gateway request handler latency in milliseconds"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway request latency metric: %w", m.err)
		return
	}

	m.capabilityRequestCount, m.err = meter.Int64Counter(
		"http_trigger_gateway_capability_request_count",
		metric.WithDescription("Gateway node metric. Number of HTTP trigger requests sent from gateway node to capability nodes"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway capability request count metric: %w", m.err)
		return
	}

	m.capabilityRequestFailures, m.err = meter.Int64Counter(
		"http_trigger_gateway_capability_request_failures",
		metric.WithDescription("Gateway node metric. Number of errors while sending HTTP trigger requests from gateway node to capability nodes"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway capability request failures metric: %w", m.err)
		return
	}

	m.capabilityMetadataProcessingFailures, m.err = meter.Int64Counter(
		"http_trigger_gateway_capability_metadata_processing_failures",
		metric.WithDescription("Gateway node metric. Number of HTTP trigger gateway metadata processing failures"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway metadata processing failures metric: %w", m.err)
		return
	}

	m.capabilityMetadataRequestCount, m.err = meter.Int64Counter(
		"http_trigger_gateway_capability_metadata_request_count",
		metric.WithDescription("Gateway node metric. Number of HTTP trigger gateway metadata requests"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create HTTP trigger gateway metadata request count metric: %w", m.err)
		return
	}

	m.metadataObservationsCleanUpCount, m.err = meter.Int64Counter(
		"http_trigger_metadata_observations_clean_count",
		metric.WithDescription("Gateway node metric. Number of workflow metadata observations cleaned"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create workflow metadata observations clean count metric: %w", m.err)
		return
	}

	m.metadataObservationsCount, m.err = meter.Int64Gauge(
		"http_trigger_metadata_observations_count",
		metric.WithDescription("Gateway node metric. Current number of workflow metadata observations in memory"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create workflow metadata observations count metric: %w", m.err)
		return
	}
}

func IncrementHTTPTriggerGatewayRequestCount(ctx context.Context, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway request count metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.requestCount.Add(ctx, 1)
}

func IncrementHTTPTriggerGatewayRequestErrors(ctx context.Context, errorCode string, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway request errors metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.requestErrors.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrErrorCode, errorCode)))
}

func IncrementHTTPTriggerGatewayWorkflowOwnerThrottled(ctx context.Context, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway workflow owner throttled metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.workflowOwnerThrottled.Add(ctx, 1)
}

func IncrementHTTPTriggerGatewayGlobalThrottled(ctx context.Context, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway global throttled metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.globalThrottled.Add(ctx, 1)
}

func IncrementHTTPTriggerGatewayPendingRequestsCleanUpCount(ctx context.Context, count int64, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway pending requests cleanup count metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.pendingRequestsCleanUpCount.Add(ctx, count)
}

func RecordHTTPTriggerGatewayPendingRequestsCount(ctx context.Context, count int64, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway pending requests count metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.pendingRequestsCount.Record(ctx, count)
}

func RecordHTTPTriggerGatewayRequestHandlerLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway request latency metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.requestHandlerLatency.Record(ctx, latencyMs)
}

func IncrementHTTPTriggerGatewayCapabilityRequestCount(ctx context.Context, nodeAddress string, methodName string, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway capability request count metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.capabilityRequestCount.Add(ctx, 1, metric.WithAttributes(
		attribute.String(AttrNodeAddress, nodeAddress),
		attribute.String(AttrMethodName, methodName),
	))
}

func IncrementHTTPTriggerGatewayCapabilityRequestFailures(ctx context.Context, nodeAddress string, methodName string, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway capability request failures metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.capabilityRequestFailures.Add(ctx, 1, metric.WithAttributes(
		attribute.String(AttrNodeAddress, nodeAddress),
		attribute.String(AttrMethodName, methodName),
	))
}

func IncrementHTTPTriggerGatewayCapabilityMetadataProcessingFailures(ctx context.Context, nodeAddress string, methodName string, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway metadata processing failures metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.capabilityMetadataProcessingFailures.Add(ctx, 1, metric.WithAttributes(
		attribute.String(AttrNodeAddress, nodeAddress),
		attribute.String(AttrMethodName, methodName),
	))
}

func IncrementHTTPTriggerGatewayCapabilityMetadataRequestCount(ctx context.Context, nodeAddress string, methodName string, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger gateway metadata request count metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.capabilityMetadataRequestCount.Add(ctx, 1, metric.WithAttributes(
		attribute.String(AttrNodeAddress, nodeAddress),
		attribute.String(AttrMethodName, methodName),
	))
}

func IncrementHTTPTriggerMetadataObservationsCleanUpCount(ctx context.Context, count int64, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize workflow metadata observations clean count metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.metadataObservationsCleanUpCount.Add(ctx, count)
}

func RecordHTTPTriggerMetadataObservationsCount(ctx context.Context, count int64, lggr logger.Logger) {
	httpTriggerGatewayMetrics.once.Do(httpTriggerGatewayMetrics.init)
	if httpTriggerGatewayMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger metadata observations count metric", "error", httpTriggerGatewayMetrics.err)
		}
		return
	}
	httpTriggerGatewayMetrics.metadataObservationsCount.Record(ctx, count)
}
