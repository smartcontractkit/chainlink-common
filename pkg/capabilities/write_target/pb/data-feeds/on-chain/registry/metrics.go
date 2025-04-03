package registry

import (
	"context"
	"fmt"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/utils"
)

// ns returns a namespaced metric name
func ns(name string) string {
	return fmt.Sprintf("data_feeds_on_chain_registry_%s", name)
}

// Define metrics configuration
var (
	feedUpdated = struct {
		basic utils.MetricsInfoCapBasic
		// specific to FeedUpdated
		observationsTimestamp utils.MetricInfo
		duration              utils.MetricInfo // ts.emit - ts.observation
		benchmark             utils.MetricInfo
		blockTimestamp        utils.MetricInfo
		blockNumber           utils.MetricInfo
	}{
		basic: utils.NewMetricsInfoCapBasic(ns("feed_updated"), "data-feeds.on-chain.registry.FeedUpdated"),
		observationsTimestamp: utils.MetricInfo{
			Name:        ns("feed_updated_observations_timestamp"),
			Unit:        "ms",
			Description: "The observations timestamp for the latest confirmed update (as reported)",
		},
		duration: utils.MetricInfo{
			Name:        ns("feed_updated_duration"),
			Unit:        "ms",
			Description: "The duration (local) since observation to message: 'data-feeds.on-chain.registry.FeedUpdated' emit",
		},
		benchmark: utils.MetricInfo{
			Name:        ns("feed_updated_benchmark"),
			Unit:        "",
			Description: "The benchmark value for the latest confirmed update (as reported)",
		},
		blockTimestamp: utils.MetricInfo{
			Name:        ns("feed_updated_block_timestamp"),
			Unit:        "ms",
			Description: "The block timestamp at the latest confirmed update (as observed)",
		},
		blockNumber: utils.MetricInfo{
			Name:        ns("feed_updated_block_number"),
			Unit:        "",
			Description: "The block number at the latest confirmed update (as observed)",
		},
	}
)

// Define a new struct for metrics
type Metrics struct {
	// Define on FeedUpdated metrics
	feedUpdated struct {
		basic utils.MetricsCapBasic
		// specific to FeedUpdated
		observationsTimestamp metric.Int64Gauge
		duration              metric.Int64Gauge // ts.emit - ts.observation
		benchmark             metric.Float64Gauge
		blockTimestamp        metric.Int64Gauge
		blockNumber           metric.Int64Gauge
	}
}

func NewMetrics() (*Metrics, error) {
	// Define new metrics
	m := &Metrics{}

	meter := beholder.GetMeter()

	// Create new metrics
	var err error

	m.feedUpdated.basic, err = utils.NewMetricsCapBasic(feedUpdated.basic)
	if err != nil {
		return nil, fmt.Errorf("failed to create new basic metrics: %w", err)
	}

	m.feedUpdated.observationsTimestamp, err = feedUpdated.observationsTimestamp.NewInt64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	m.feedUpdated.duration, err = feedUpdated.duration.NewInt64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	m.feedUpdated.benchmark, err = feedUpdated.benchmark.NewFloat64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	m.feedUpdated.blockTimestamp, err = feedUpdated.blockTimestamp.NewInt64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	m.feedUpdated.blockNumber, err = feedUpdated.blockNumber.NewInt64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	return m, nil
}

func (m *Metrics) OnFeedUpdated(ctx context.Context, msg *FeedUpdated, attrKVs ...any) error {
	// Define attributes
	attrs := metric.WithAttributes(msg.Attributes()...)

	// Emit basic metrics (count, timestamps)
	start, emit := msg.MetaCapabilityTimestampStart, msg.MetaCapabilityTimestampEmit
	m.feedUpdated.basic.RecordEmit(ctx, start, emit, msg.Attributes()...)

	// Timestamp e2e observation update
	m.feedUpdated.observationsTimestamp.Record(ctx, int64(msg.ObservationsTimestamp), attrs)
	observation := uint64(msg.ObservationsTimestamp) * 1000 // convert to milliseconds
	m.feedUpdated.duration.Record(ctx, int64(emit-observation), attrs)

	// Benchmark
	m.feedUpdated.benchmark.Record(ctx, msg.BenchmarkVal, attrs)

	// Block timestamp
	m.feedUpdated.blockTimestamp.Record(ctx, int64(msg.BlockTimestamp), attrs)

	// Block number
	blockHeightVal, err := strconv.ParseInt(msg.BlockHeight, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse block height: %w", err)
	}
	m.feedUpdated.blockNumber.Record(ctx, blockHeightVal, attrs)

	return nil
}

// Attributes returns the attributes for the FeedUpdated message to be used in metrics
func (m *FeedUpdated) Attributes() []attribute.KeyValue {
	context := utils.ExecutionMetadata{
		// Execution Context - Source
		SourceId: m.MetaSourceId,
		// Execution Context - Chain
		ChainFamilyName: m.MetaChainFamilyName,
		ChainId:         m.MetaChainId,
		NetworkName:     m.MetaNetworkName,
		NetworkNameFull: m.MetaNetworkNameFull,
		// Execution Context - Workflow (capabilities.RequestMetadata)
		WorkflowId:               m.MetaWorkflowId,
		WorkflowOwner:            m.MetaWorkflowOwner,
		WorkflowExecutionId:      m.MetaWorkflowExecutionId,
		WorkflowName:             m.MetaWorkflowName,
		WorkflowDonId:            m.MetaWorkflowDonId,
		WorkflowDonConfigVersion: m.MetaWorkflowDonConfigVersion,
		ReferenceId:              m.MetaReferenceId,
		// Execution Context - Capability
		CapabilityType: m.MetaCapabilityType,
		CapabilityId:   m.MetaCapabilityId,
	}

	attrs := []attribute.KeyValue{
		// Transaction Data
		attribute.String("tx_sender", m.TxSender),
		attribute.String("tx_receiver", m.TxReceiver),

		// Event Data
		attribute.String("feed_id", m.FeedId),
		// TODO: do we need these attributes? (available in WriteConfirmed)
		// attribute.Int64("report_id", int64(m.ReportId)), // uint32 -> int64

		// We mark confrmations by transmitter so we can query for only initial (fast) confirmations
		// with PromQL, and ignore the slower confirmations by other signers for SLA measurements.
		attribute.Bool("observed_by_transmitter", m.TxSender == m.MetaSourceId), // source_id == node account
		// TODO: remove once NOT_SET bug with non-string labels is fixed
		attribute.String("observed_by_transmitter_str", strconv.FormatBool(m.TxSender == m.MetaSourceId)),
	}

	return append(attrs, context.Attributes()...)
}
