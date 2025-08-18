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

// HTTPActionCapabilityMetrics contains metrics for HTTP actions on capability nodes
type HTTPActionCapabilityMetrics struct {
	requestCount            metric.Int64Counter
	inputValidationFailures metric.Int64Counter
	workflowOwnerThrottled  metric.Int64Counter
	nodeThrottled           metric.Int64Counter
	gatewayConnectionError  metric.Int64Counter
	gatewaySendError        metric.Int64Counter
	successfulResponse      metric.Int64Counter
	executionError          metric.Int64Counter
	gatewayNodeThrottled    metric.Int64Counter
	gatewayGlobalThrottled  metric.Int64Counter
	requestLatency          metric.Int64Histogram

	once sync.Once
	err  error
}

var httpActionCapabilityMetrics = &HTTPActionCapabilityMetrics{}

func (m *HTTPActionCapabilityMetrics) init() {
	meter := beholder.GetMeter()

	m.requestCount, m.err = meter.Int64Counter(
		"http_action_request_count",
		metric.WithDescription("Capability node metric. Number of HTTP action requests"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create request count metric: %w", m.err)
		return
	}

	m.inputValidationFailures, m.err = meter.Int64Counter(
		"http_action_validation_failure_count",
		metric.WithDescription("Capability node metric. Number of HTTP action input validation failures"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create validation failure metric: %w", m.err)
		return
	}

	m.workflowOwnerThrottled, m.err = meter.Int64Counter(
		"http_action_workflow_owner_throttled_count",
		metric.WithDescription("Capability node metric. Number of HTTP action requests exceeding per-workflow-owner rate limit"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create workflow owner throttled metric: %w", m.err)
		return
	}

	m.nodeThrottled, m.err = meter.Int64Counter(
		"http_action_node_throttled_count",
		metric.WithDescription("Capability node metric. Number of HTTP action requests exceeding global rate limit"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create node throttled metric: %w", m.err)
		return
	}

	m.gatewayConnectionError, m.err = meter.Int64Counter(
		"http_action_capability_gateway_connection_error_count",
		metric.WithDescription("Capability node metric. Number of HTTP action gateway connection errors."),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create capability gateway connection error metric: %w", m.err)
		return
	}

	m.gatewaySendError, m.err = meter.Int64Counter(
		"http_action_capability_gateway_send_error_count",
		metric.WithDescription("Capability node metric. Number of HTTP action gateway send errors."),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create capability gateway send error metric: %w", m.err)
		return
	}

	m.successfulResponse, m.err = meter.Int64Counter(
		"http_action_successful_response_count",
		metric.WithDescription("Capability node metric. Number of HTTP action successful responses"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create successful response metric: %w", m.err)
		return
	}

	m.executionError, m.err = meter.Int64Counter(
		"http_action_execution_error_count",
		metric.WithDescription("Capability node metric. Number of HTTP action execution errors"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create execution error metric: %w", m.err)
		return
	}

	m.gatewayNodeThrottled, m.err = meter.Int64Counter(
		"http_action_capability_gateway_node_throttled_count",
		metric.WithDescription("Capability node metric. Number of throttled requests while receiving HTTP action response from gateway. Per-gateway-node rate limit"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create capability gateway node throttled metric: %w", m.err)
		return
	}

	m.gatewayGlobalThrottled, m.err = meter.Int64Counter(
		"http_action_capability_gateway_global_throttled_count",
		metric.WithDescription("Capability node metric. Number of thro throttled requests while receiving HTTP action response from gateway. Global limit."),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create capability gateway global throttled metric: %w", m.err)
		return
	}

	m.requestLatency, m.err = meter.Int64Histogram(
		"http_action_request_latency_ms",
		metric.WithDescription("Capability node metric. HTTP action request latency in milliseconds"),
	)
	if m.err != nil {
		m.err = fmt.Errorf("failed to create request latency metric: %w", m.err)
		return
	}
}

func IncrementHTTPActionRequestCount(ctx context.Context, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action request count metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.requestCount.Add(ctx, 1)
}

func IncrementHTTPActionInputValidationFailures(ctx context.Context, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action validation failure metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.inputValidationFailures.Add(ctx, 1)
}

func IncrementHTTPActionWorkflowOwnerThrottled(ctx context.Context, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action workflow owner throttled metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.workflowOwnerThrottled.Add(ctx, 1)
}

func IncrementHTTPActionNodeThrottled(ctx context.Context, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action node throttled metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.nodeThrottled.Add(ctx, 1)
}

func IncrementHTTPActionCapabilityGatewayConnectionError(ctx context.Context, nodeAddress string, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action capability gateway connection error metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.gatewayConnectionError.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrNodeAddress, nodeAddress)))
}

func IncrementHTTPActionCapabilityGatewaySendError(ctx context.Context, nodeAddress string, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action capability gateway send error metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.gatewaySendError.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrNodeAddress, nodeAddress)))
}

func IncrementHTTPActionSuccessfulResponse(ctx context.Context, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action successful response metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.successfulResponse.Add(ctx, 1)
}

func IncrementHTTPActionExecutionError(ctx context.Context, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action internal execution error metric. Any errors that happen before or after receiving a response from the customer's endpoint", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.executionError.Add(ctx, 1)
}

func IncrementHTTPActionCapabilityGatewayNodeThrottled(ctx context.Context, nodeAddress string, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action capability gateway node throttled metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.gatewayNodeThrottled.Add(ctx, 1, metric.WithAttributes(attribute.String(AttrNodeAddress, nodeAddress)))
}

func IncrementHTTPActionCapabilityGatewayGlobalThrottled(ctx context.Context, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action capability gateway global throttled metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.gatewayGlobalThrottled.Add(ctx, 1)
}

func RecordHTTPActionRequestLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	httpActionCapabilityMetrics.once.Do(httpActionCapabilityMetrics.init)
	if httpActionCapabilityMetrics.err != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action request latency metric", "error", httpActionCapabilityMetrics.err)
		}
		return
	}
	httpActionCapabilityMetrics.requestLatency.Record(ctx, latencyMs)
}
