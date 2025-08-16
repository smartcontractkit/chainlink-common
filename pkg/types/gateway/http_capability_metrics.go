package gateway

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var (
	httpActionRequestCount                         metric.Int64Counter
	httpActionInputValidationFailures              metric.Int64Counter
	httpActionWorkflowOwnerThrottled               metric.Int64Counter
	httpActionNodeThrottled                        metric.Int64Counter
	httpActionCapabilityGatewayConnectionError     metric.Int64Counter
	httpActionCapabilityGatewaySendError           metric.Int64Counter
	httpActionSuccessfulResponse                   metric.Int64Counter
	httpActionExecutionError                       metric.Int64Counter
	httpActionCapabilityGatewayNodeThrottled       metric.Int64Counter
	httpActionCapabilityGatewayGlobalThrottled     metric.Int64Counter
	httpActionRequestLatency                       metric.Int64Histogram
	httpTriggerRegisterCount                       metric.Int64Counter
	httpTriggerUnregisterCount                     metric.Int64Counter
	httpTriggerRegisterFailureCount                metric.Int64Counter
	httpTriggerUnregisterFailureCount              metric.Int64Counter
	httpTriggerRequestCacheCleanUpCount            metric.Int64Counter
	httpTriggerCapabilityRequestCount              metric.Int64Counter
	httpTriggerCapabilityGatewayGlobalThrottled    metric.Int64Counter
	httpTriggerCapabilityGatewayNodeThrottled      metric.Int64Counter
	httpTriggerCapabilityRequestSuccessCount       metric.Int64Counter
	httpTriggerCapabilityGatewaySendError          metric.Int64Counter
	httpTriggerCapabilityBroadcastMetadataCount    metric.Int64Counter
	httpTriggerCapabilityBroadcastMetadataFailures metric.Int64Counter
	httpTriggerCapabilityBroadcastMetadataLatency  metric.Int64Histogram
	httpTriggerCapabilityPullMetadataCount         metric.Int64Counter
	httpTriggerCapabilityPullMetadataFailures      metric.Int64Counter
	httpTriggerCapabilityPullMetadataLatency       metric.Int64Histogram
	httpTriggerCapabilityRequestLatency            metric.Int64Histogram
	metricsOnce                                    sync.Once
	metricsInitErr                                 error
)

// initMetrics initializes all HTTP action metrics using beholder (called once)
func initMetrics() {
	meter := beholder.GetMeter()

	httpActionRequestCount, metricsInitErr = meter.Int64Counter(
		"http_action_request_count",
		metric.WithDescription("Capability node metric. Number of HTTP action requests"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create request count metric: %w", metricsInitErr)
		return
	}

	httpActionInputValidationFailures, metricsInitErr = meter.Int64Counter(
		"http_action_validation_failure_count",
		metric.WithDescription("Capability node metric. Number of HTTP action validation failures"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create validation failure metric: %w", metricsInitErr)
		return
	}

	httpActionWorkflowOwnerThrottled, metricsInitErr = meter.Int64Counter(
		"http_action_workflow_owner_throttled_count",
		metric.WithDescription("Capability node metric. Number of HTTP action requests exceeding per-workflow-owner rate limit"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create workflow owner throttled metric: %w", metricsInitErr)
		return
	}

	httpActionNodeThrottled, metricsInitErr = meter.Int64Counter(
		"http_action_node_throttled_count",
		metric.WithDescription("Capability node metric. Number of HTTP action requests exceeding global rate limit"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create node throttled metric: %w", metricsInitErr)
		return
	}

	httpActionCapabilityGatewayConnectionError, metricsInitErr = meter.Int64Counter(
		"http_action_capability_gateway_connection_error_count",
		metric.WithDescription("Capability node metric. Number of HTTP action gateway connection errors."),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create capability gateway connection error metric: %w", metricsInitErr)
		return
	}

	httpActionCapabilityGatewaySendError, metricsInitErr = meter.Int64Counter(
		"http_action_capability_gateway_send_error_count",
		metric.WithDescription("Capability node metric. Number of HTTP action gateway send errors."),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create capability gateway send error metric: %w", metricsInitErr)
		return
	}

	httpActionSuccessfulResponse, metricsInitErr = meter.Int64Counter(
		"http_action_successful_response_count",
		metric.WithDescription("Capability node metric. Number of HTTP action successful responses"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create successful response metric: %w", metricsInitErr)
		return
	}

	httpActionExecutionError, metricsInitErr = meter.Int64Counter(
		"http_action_execution_error_count",
		metric.WithDescription("Capability node metric. Number of HTTP action execution errors"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create execution error metric: %w", metricsInitErr)
		return
	}

	httpActionCapabilityGatewayNodeThrottled, metricsInitErr = meter.Int64Counter(
		"http_action_capability_gateway_node_throttled_count",
		metric.WithDescription("Capability node metric. Number of HTTP action capability throttled requests by gateway while receiving response from gateway"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create capability gateway node throttled metric: %w", metricsInitErr)
		return
	}

	httpActionCapabilityGatewayGlobalThrottled, metricsInitErr = meter.Int64Counter(
		"http_action_capability_gateway_global_throttled_count",
		metric.WithDescription("Capability node metric. Number of HTTP action capability throttled requests globally while receiving response from gateway"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create capability gateway global throttled metric: %w", metricsInitErr)
		return
	}

	httpActionRequestLatency, metricsInitErr = meter.Int64Histogram(
		"http_action_request_latency_ms",
		metric.WithDescription("Capability node metric. HTTP action request latency in milliseconds"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create request latency metric: %w", metricsInitErr)
		return
	}

	httpTriggerRegisterCount, metricsInitErr = meter.Int64Counter(
		"http_trigger_register_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger registrations"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger register count metric: %w", metricsInitErr)
		return
	}

	httpTriggerUnregisterCount, metricsInitErr = meter.Int64Counter(
		"http_trigger_unregister_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger unregistrations"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger unregister count metric: %w", metricsInitErr)
		return
	}

	httpTriggerRegisterFailureCount, metricsInitErr = meter.Int64Counter(
		"http_trigger_register_failure_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger registration failures"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger register failure count metric: %w", metricsInitErr)
		return
	}

	httpTriggerUnregisterFailureCount, metricsInitErr = meter.Int64Counter(
		"http_trigger_unregister_failure_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger unregistration failures"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger unregister failure count metric: %w", metricsInitErr)
		return
	}

	httpTriggerRequestCacheCleanUpCount, metricsInitErr = meter.Int64Counter(
		"http_trigger_request_cache_cleanup_count",
		metric.WithDescription("Capability node metric. Number of expired entries cleaned up from HTTP trigger request cache"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger request cache cleanup count metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityRequestCount, metricsInitErr = meter.Int64Counter(
		"http_trigger_capability_request_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger requests processed"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability request count metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityGatewayGlobalThrottled, metricsInitErr = meter.Int64Counter(
		"http_trigger_capability_gateway_global_throttled",
		metric.WithDescription("Capability node metric. Number of HTTP trigger requests throttled due to global rate limit"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability gateway global throttled metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityGatewayNodeThrottled, metricsInitErr = meter.Int64Counter(
		"http_trigger_capability_gateway_node_throttled",
		metric.WithDescription("Capability node metric. Number of HTTP trigger requests throttled due to per-node rate limit"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability gateway node throttled metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityRequestSuccessCount, metricsInitErr = meter.Int64Counter(
		"http_trigger_capability_request_success_count",
		metric.WithDescription("Capability node metric. Number of successful HTTP trigger responses sent to gateway"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability request success count metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityGatewaySendError, metricsInitErr = meter.Int64Counter(
		"http_trigger_capability_gateway_send_error",
		metric.WithDescription("Capability node metric. Number of HTTP trigger gateway send errors"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability gateway send error metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityBroadcastMetadataCount, metricsInitErr = meter.Int64Counter(
		"http_trigger_capability_broadcast_metadata_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger broadcast metadata workflow operations"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability broadcast metadata count metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityBroadcastMetadataFailures, metricsInitErr = meter.Int64Counter(
		"http_trigger_capability_broadcast_metadata_failures",
		metric.WithDescription("Capability node metric. Number of HTTP trigger broadcast metadata workflow failures"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability broadcast metadata failures metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityBroadcastMetadataLatency, metricsInitErr = meter.Int64Histogram(
		"http_trigger_capability_broadcast_metadata_latency_ms",
		metric.WithDescription("Capability node metric. HTTP trigger broadcast metadata latency in milliseconds"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability broadcast metadata latency metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityPullMetadataCount, metricsInitErr = meter.Int64Counter(
		"http_trigger_capability_pull_metadata_count",
		metric.WithDescription("Capability node metric. Number of HTTP trigger pull metadata workflow operations"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability pull metadata count metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityPullMetadataFailures, metricsInitErr = meter.Int64Counter(
		"http_trigger_capability_pull_metadata_failures",
		metric.WithDescription("Capability node metric. Number of HTTP trigger pull metadata workflow failures"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability pull metadata failures metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityPullMetadataLatency, metricsInitErr = meter.Int64Histogram(
		"http_trigger_capability_pull_metadata_latency_ms",
		metric.WithDescription("Capability node metric. HTTP trigger pull metadata latency in milliseconds"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability pull metadata latency metric: %w", metricsInitErr)
		return
	}

	httpTriggerCapabilityRequestLatency, metricsInitErr = meter.Int64Histogram(
		"http_trigger_capability_request_latency_ms",
		metric.WithDescription("Capability node metric. HTTP trigger capability request processing latency in milliseconds"),
	)
	if metricsInitErr != nil {
		metricsInitErr = fmt.Errorf("failed to create HTTP trigger capability request latency metric: %w", metricsInitErr)
		return
	}
}

// IncrementHTTPActionRequestCount increments the HTTP action request count metric
func IncrementHTTPActionRequestCount(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action request count metric", "error", metricsInitErr)
		}
		return
	}
	httpActionRequestCount.Add(ctx, 1)
}

// IncrementHTTPActionInputValidationFailures increments the HTTP action input validation failures metric
func IncrementHTTPActionInputValidationFailures(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action validation failure metric", "error", metricsInitErr)
		}
		return
	}
	httpActionInputValidationFailures.Add(ctx, 1)
}

func IncrementHTTPActionWorkflowOwnerThrottled(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action workflow owner throttled metric", "error", metricsInitErr)
		}
		return
	}
	httpActionWorkflowOwnerThrottled.Add(ctx, 1)
}

func IncrementHTTPActionNodeThrottled(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action node throttled metric", "error", metricsInitErr)
		}
		return
	}
	httpActionNodeThrottled.Add(ctx, 1)
}

func IncrementHTTPActionCapabilityGatewayConnectionError(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action capability gateway connection error metric", "error", metricsInitErr)
		}
		return
	}
	httpActionCapabilityGatewayConnectionError.Add(ctx, 1)
}

func IncrementHTTPActionCapabilityGatewaySendError(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action capability gateway send error metric", "error", metricsInitErr)
		}
		return
	}
	httpActionCapabilityGatewaySendError.Add(ctx, 1)
}

// IncrementHTTPActionSuccessfulResponse increments regardless of status code returned from customer's endpoint
func IncrementHTTPActionSuccessfulResponse(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action successful response metric", "error", metricsInitErr)
		}
		return
	}
	httpActionSuccessfulResponse.Add(ctx, 1)
}

// IncrementHTTPActionExecutionError increments if there were errors not related to the customer's endpoint
func IncrementHTTPActionExecutionError(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action execution error metric", "error", metricsInitErr)
		}
		return
	}
	httpActionExecutionError.Add(ctx, 1)
}

func IncrementHTTPActionCapabilityGatewayNodeThrottled(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action capability gateway node throttled metric", "error", metricsInitErr)
		}
		return
	}
	httpActionCapabilityGatewayNodeThrottled.Add(ctx, 1)
}

func IncrementHTTPActionCapabilityGatewayGlobalThrottled(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action capability gateway global throttled metric", "error", metricsInitErr)
		}
		return
	}
	httpActionCapabilityGatewayGlobalThrottled.Add(ctx, 1)
}

func RecordHTTPActionRequestLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP action request latency metric", "error", metricsInitErr)
		}
		return
	}
	httpActionRequestLatency.Record(ctx, latencyMs)
}

func IncrementHTTPTriggerRegisterCount(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger register count metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerRegisterCount.Add(ctx, 1)
}

func IncrementHTTPTriggerUnregisterCount(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger unregister count metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerUnregisterCount.Add(ctx, 1)
}

func IncrementHTTPTriggerRegisterFailureCount(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger register failure count metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerRegisterFailureCount.Add(ctx, 1)
}

func IncrementHTTPTriggerUnregisterFailureCount(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger unregister failure count metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerUnregisterFailureCount.Add(ctx, 1)
}

func IncrementHTTPTriggerRequestCacheCleanUpCount(ctx context.Context, count int64, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger request cache cleanup count metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerRequestCacheCleanUpCount.Add(ctx, count)
}

func IncrementHTTPTriggerCapabilityRequestCount(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability request count metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityRequestCount.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityGatewayGlobalThrottled(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability gateway global throttled metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityGatewayGlobalThrottled.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityGatewayNodeThrottled(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability gateway node throttled metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityGatewayNodeThrottled.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityRequestSuccessCount(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability request success count metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityRequestSuccessCount.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityGatewaySendError(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability gateway send error metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityGatewaySendError.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityBroadcastMetadataCount(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability broadcast metadata count metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityBroadcastMetadataCount.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityBroadcastMetadataFailures(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability broadcast metadata failures metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityBroadcastMetadataFailures.Add(ctx, 1)
}

func RecordHTTPTriggerCapabilityBroadcastMetadataLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability broadcast metadata latency metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityBroadcastMetadataLatency.Record(ctx, latencyMs)
}

func IncrementHTTPTriggerCapabilityPullMetadataCount(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability pull metadata count metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityPullMetadataCount.Add(ctx, 1)
}

func IncrementHTTPTriggerCapabilityPullMetadataFailures(ctx context.Context, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability pull metadata failures metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityPullMetadataFailures.Add(ctx, 1)
}

func RecordHTTPTriggerCapabilityPullMetadataLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability pull metadata latency metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityPullMetadataLatency.Record(ctx, latencyMs)
}

func RecordHTTPTriggerCapabilityRequestLatency(ctx context.Context, latencyMs int64, lggr logger.Logger) {
	metricsOnce.Do(initMetrics)
	if metricsInitErr != nil {
		if lggr != nil {
			lggr.Errorw("Failed to initialize HTTP trigger capability request latency metric", "error", metricsInitErr)
		}
		return
	}
	httpTriggerCapabilityRequestLatency.Record(ctx, latencyMs)
}
