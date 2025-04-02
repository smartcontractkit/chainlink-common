package write_target

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
	return fmt.Sprintf("platform_write_target_%s", name)
}

// Define metrics configuration
var (
	writeInitiated = struct {
		basic utils.MetricsInfoCapBasic
	}{
		basic: utils.NewMetricsInfoCapBasic(ns("write_initiated"), "platform.write-target.WriteInitiated"),
	}
	writeError = struct {
		basic utils.MetricsInfoCapBasic
	}{
		basic: utils.NewMetricsInfoCapBasic(ns("write_error"), "platform.write-target.WriteError"),
	}
	writeSent = struct {
		basic utils.MetricsInfoCapBasic
		// specific to WriteSent
		blockTimestamp utils.MetricInfo
		blockNumber    utils.MetricInfo
	}{
		basic: utils.NewMetricsInfoCapBasic(ns("write_sent"), "platform.write-target.WriteSent"),
		blockTimestamp: utils.MetricInfo{
			Name:        ns("write_sent_block_timestamp"),
			Unit:        "ms",
			Description: "The block timestamp at the latest sent write (as observed)",
		},
		blockNumber: utils.MetricInfo{
			Name:        ns("write_sent_block_number"),
			Unit:        "",
			Description: "The block number at the latest sent write (as observed)",
		},
	}
	writeConfirmed = struct {
		basic utils.MetricsInfoCapBasic
		// specific to WriteSent
		blockTimestamp utils.MetricInfo
		blockNumber    utils.MetricInfo
		signersNumber  utils.MetricInfo
	}{
		basic: utils.NewMetricsInfoCapBasic(ns("write_confirmed"), "platform.write-target.WriteConfirmed"),
		blockTimestamp: utils.MetricInfo{
			Name:        ns("write_confirmed_block_timestamp"),
			Unit:        "ms",
			Description: "The block timestamp for latest confirmed write (as observed)",
		},
		blockNumber: utils.MetricInfo{
			Name:        ns("write_confirmed_block_number"),
			Unit:        "",
			Description: "The block number for latest confirmed write (as observed)",
		},
		signersNumber: utils.MetricInfo{
			Name:        ns("write_confirmed_signers_number"),
			Unit:        "",
			Description: "The number of signers attached to the processed and confirmed write request",
		},
	}
)

// Define a new struct for metrics
type Metrics struct {
	// Define on WriteInitiated metrics
	writeInitiated struct {
		basic utils.MetricsCapBasic
	}
	// Define on WriteError metrics
	writeError struct {
		basic utils.MetricsCapBasic
	}
	// Define on WriteSent metrics
	writeSent struct {
		basic utils.MetricsCapBasic
		// specific to WriteSent
		blockTimestamp metric.Int64Gauge
		blockNumber    metric.Int64Gauge
	}
	// Define on WriteConfirmed metrics
	writeConfirmed struct {
		basic utils.MetricsCapBasic
		// specific to WriteConfirmed
		blockTimestamp metric.Int64Gauge
		blockNumber    metric.Int64Gauge
		signersNumber  metric.Int64Gauge
	}
}

func NewMetrics() (*Metrics, error) {
	// Define new metrics
	m := &Metrics{}

	meter := beholder.GetMeter()

	// Create new metrics
	var err error

	// WriteInitiated
	m.writeInitiated.basic, err = utils.NewMetricsCapBasic(writeInitiated.basic)
	if err != nil {
		return nil, fmt.Errorf("failed to create new basic metrics: %w", err)
	}

	// WriteError
	m.writeError.basic, err = utils.NewMetricsCapBasic(writeError.basic)
	if err != nil {
		return nil, fmt.Errorf("failed to create new basic metrics: %w", err)
	}

	// WriteSent
	m.writeSent.basic, err = utils.NewMetricsCapBasic(writeSent.basic)
	if err != nil {
		return nil, fmt.Errorf("failed to create new basic metrics: %w", err)
	}

	m.writeSent.blockTimestamp, err = writeSent.blockTimestamp.NewInt64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	m.writeSent.blockNumber, err = writeSent.blockNumber.NewInt64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	// WriteConfirmed
	m.writeConfirmed.basic, err = utils.NewMetricsCapBasic(writeConfirmed.basic)
	if err != nil {
		return nil, fmt.Errorf("failed to create new basic metrics: %w", err)
	}

	m.writeConfirmed.blockTimestamp, err = writeConfirmed.blockTimestamp.NewInt64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	m.writeConfirmed.blockNumber, err = writeConfirmed.blockNumber.NewInt64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	m.writeConfirmed.signersNumber, err = writeConfirmed.signersNumber.NewInt64Gauge(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge: %w", err)
	}

	return m, nil
}

func (m *Metrics) OnWriteInitiated(ctx context.Context, msg *WriteInitiated, attrKVs ...any) error {
	// Emit basic metrics (count, timestamps)
	start, emit := msg.MetaCapabilityTimestampStart, msg.MetaCapabilityTimestampEmit
	m.writeInitiated.basic.RecordEmit(ctx, start, emit, msg.Attributes()...)
	return nil
}

func (m *Metrics) OnWriteError(ctx context.Context, msg *WriteError, attrKVs ...any) error {
	// Emit basic metrics (count, timestamps)
	start, emit := msg.MetaCapabilityTimestampStart, msg.MetaCapabilityTimestampEmit
	m.writeError.basic.RecordEmit(ctx, start, emit, msg.Attributes()...)
	return nil
}

func (m *Metrics) OnWriteSent(ctx context.Context, msg *WriteSent, attrKVs ...any) error {
	// Define attributes
	attrs := metric.WithAttributes(msg.Attributes()...)

	// Emit basic metrics (count, timestamps)
	start, emit := msg.MetaCapabilityTimestampStart, msg.MetaCapabilityTimestampEmit
	m.writeSent.basic.RecordEmit(ctx, start, emit, msg.Attributes()...)

	// Block timestamp
	m.writeSent.blockTimestamp.Record(ctx, int64(msg.BlockTimestamp), attrs)

	// Block number
	blockHeightVal, err := strconv.ParseInt(msg.BlockHeight, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse block height: %w", err)
	}
	m.writeSent.blockNumber.Record(ctx, blockHeightVal, attrs)
	return nil
}

func (m *Metrics) OnWriteConfirmed(ctx context.Context, msg *WriteConfirmed, attrKVs ...any) error {
	// Define attributes
	attrs := metric.WithAttributes(msg.Attributes()...)

	// Emit basic metrics (count, timestamps)
	start, emit := msg.MetaCapabilityTimestampStart, msg.MetaCapabilityTimestampEmit
	m.writeConfirmed.basic.RecordEmit(ctx, start, emit, msg.Attributes()...)

	// Signers number
	m.writeConfirmed.signersNumber.Record(ctx, int64(msg.SignersNum), attrs)

	// Block timestamp
	m.writeConfirmed.blockTimestamp.Record(ctx, int64(msg.BlockTimestamp), attrs)

	// Block number
	blockHeightVal, err := strconv.ParseInt(msg.BlockHeight, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse block height: %w", err)
	}
	m.writeConfirmed.blockNumber.Record(ctx, blockHeightVal, attrs)
	return nil
}

// Attributes returns the attributes for the WriteInitiated message to be used in metrics
func (m *WriteInitiated) Attributes() []attribute.KeyValue {
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
		attribute.String("node", m.Node),
		attribute.String("forwarder", m.Forwarder),
		attribute.String("receiver", m.Receiver),
		attribute.Int64("report_id", int64(m.ReportId)), // uint32 -> int64
	}

	return append(attrs, context.Attributes()...)
}

// Attributes returns the attributes for the WriteError message to be used in metrics
func (m *WriteError) Attributes() []attribute.KeyValue {
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
		attribute.String("node", m.Node),
		attribute.String("forwarder", m.Forwarder),
		attribute.String("receiver", m.Receiver),
		attribute.Int64("report_id", int64(m.ReportId)), // uint32 -> int64
		// Error information
		attribute.Int64("code", int64(m.Code)), // uint32 -> int64
		attribute.String("summary", m.Summary),
	}

	return append(attrs, context.Attributes()...)
}

// Attributes returns the attributes for the WriteSent message to be used in metrics
func (m *WriteSent) Attributes() []attribute.KeyValue {
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
		attribute.String("node", m.Node),
		attribute.String("forwarder", m.Forwarder),
		attribute.String("receiver", m.Receiver),
		attribute.Int64("report_id", int64(m.ReportId)), // uint32 -> int64
	}

	return append(attrs, context.Attributes()...)
}

// Attributes returns the attributes for the WriteConfirmed message to be used in metrics
func (m *WriteConfirmed) Attributes() []attribute.KeyValue {
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
		attribute.String("node", m.Node),
		attribute.String("forwarder", m.Forwarder),
		attribute.String("receiver", m.Receiver),
		attribute.Int64("report_id", int64(m.ReportId)), // uint32 -> int64
		attribute.String("transmitter", m.Transmitter),
		attribute.Bool("success", m.Success),
		// We mark confrmations by transmitter so we can query for only initial (fast) confirmations
		// with PromQL, and ignore the slower confirmations by other signers for SLA measurements.
		attribute.Bool("observed_by_transmitter", m.Transmitter == m.Node),
		// TODO: remove once NOT_SET bug with non-string labels is fixed
		attribute.String("observed_by_transmitter_str", strconv.FormatBool(m.Transmitter == m.Node)),
	}

	return append(attrs, context.Attributes()...)
}
