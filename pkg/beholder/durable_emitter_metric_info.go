package beholder

// Durable emitter OTel instruments (registered via beholder.GetMeter), matching the
// MetricInfo pattern used with beholder elsewhere in chainlink-common.

var (
	durableEmitterMetricEmitSuccess = MetricInfo{
		Name:        "beholder.durable_emitter.emit.success",
		Unit:        "{call}",
		Description: "Successful durable Emit calls (insert returned)",
	}
	durableEmitterMetricEmitFailure = MetricInfo{
		Name:        "beholder.durable_emitter.emit.failure",
		Unit:        "{call}",
		Description: "Failed Emit calls (before or during insert)",
	}
	durableEmitterMetricEmitDuration = MetricInfo{
		Name:        "beholder.durable_emitter.emit.duration",
		Unit:        "s",
		Description: "Emit insert path duration",
	}
	durableEmitterMetricPublishImmSuccess = MetricInfo{
		Name:        "beholder.durable_emitter.publish.immediate.success",
		Unit:        "{call}",
		Description: "Immediate Publish RPC successes",
	}
	durableEmitterMetricPublishImmFailure = MetricInfo{
		Name:        "beholder.durable_emitter.publish.immediate.failure",
		Unit:        "{call}",
		Description: "Immediate Publish RPC failures (events await retransmit)",
	}
	durableEmitterMetricPublishBatchSuccess = MetricInfo{
		Name:        "beholder.durable_emitter.publish.retransmit.batch.success",
		Unit:        "{call}",
		Description: "Successful retransmit PublishBatch calls",
	}
	durableEmitterMetricPublishBatchFailure = MetricInfo{
		Name:        "beholder.durable_emitter.publish.retransmit.batch.failure",
		Unit:        "{call}",
		Description: "Failed retransmit PublishBatch calls",
	}
	durableEmitterMetricPublishBatchEvSuccess = MetricInfo{
		Name:        "beholder.durable_emitter.publish.retransmit.events.success",
		Unit:        "{event}",
		Description: "Events delivered via successful PublishBatch",
	}
	durableEmitterMetricPublishBatchEvFailure = MetricInfo{
		Name:        "beholder.durable_emitter.publish.retransmit.events.failure",
		Unit:        "{event}",
		Description: "Events in failed PublishBatch attempts",
	}
	durableEmitterMetricDeliveryCompleted = MetricInfo{
		Name:        "beholder.durable_emitter.delivery.completed",
		Unit:        "{event}",
		Description: "Events removed from store after successful publish (immediate or batch)",
	}
	durableEmitterMetricExpiredPurged = MetricInfo{
		Name:        "beholder.durable_emitter.expired_purged",
		Unit:        "{event}",
		Description: "Events deleted by TTL expiry loop",
	}
	durableEmitterMetricStoreOperations = MetricInfo{
		Name:        "beholder.durable_emitter.store.operations",
		Unit:        "{op}",
		Description: "Durable store operations (proxy for DB load / IOPs)",
	}
	durableEmitterMetricStoreOpDuration = MetricInfo{
		Name:        "beholder.durable_emitter.store.operation.duration",
		Unit:        "s",
		Description: "Durable store operation latency",
	}
	durableEmitterMetricQueueDepth = MetricInfo{
		Name:        "beholder.durable_emitter.queue.depth",
		Unit:        "{row}",
		Description: "Pending rows in durable queue",
	}
	durableEmitterMetricQueuePayloadBytes = MetricInfo{
		Name:        "beholder.durable_emitter.queue.payload_bytes",
		Unit:        "By",
		Description: "Sum of payload bytes for pending rows",
	}
	durableEmitterMetricQueueOldestAgeSec = MetricInfo{
		Name:        "beholder.durable_emitter.queue.oldest_pending_age_seconds",
		Unit:        "s",
		Description: "Age of oldest pending row at last poll (longest wait)",
	}
	durableEmitterMetricQueueNearTTL = MetricInfo{
		Name:        "beholder.durable_emitter.queue.near_ttl",
		Unit:        "{row}",
		Description: "Rows within near-expiry window of EventTTL (DLQ pressure proxy; no separate DLQ table)",
	}
	durableEmitterMetricQueueCapacityRatio = MetricInfo{
		Name:        "beholder.durable_emitter.queue.capacity_usage_ratio",
		Unit:        "1",
		Description: "queue.payload_bytes / MaxQueuePayloadBytes when max > 0",
	}
	durableEmitterMetricProcHeapInuse = MetricInfo{
		Name:        "beholder.durable_emitter.process.memory.heap_inuse_bytes",
		Unit:        "By",
		Description: "Go runtime MemStats HeapInuse",
	}
	durableEmitterMetricProcHeapSys = MetricInfo{
		Name:        "beholder.durable_emitter.process.memory.heap_sys_bytes",
		Unit:        "By",
		Description: "Go runtime MemStats HeapSys",
	}
	durableEmitterMetricProcCPUUser = MetricInfo{
		Name:        "beholder.durable_emitter.process.cpu.user_seconds",
		Unit:        "s",
		Description: "Cumulative user CPU seconds (getrusage; Unix only)",
	}
	durableEmitterMetricProcCPUSys = MetricInfo{
		Name:        "beholder.durable_emitter.process.cpu.system_seconds",
		Unit:        "s",
		Description: "Cumulative system CPU seconds (getrusage; Unix only)",
	}
)
