package ocr3_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/datafeeds"
	pbtypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// mockCapability implements CapabilityIface for testing
type mockCapability struct {
	aggregators map[string]pbtypes.Aggregator
}

func (m *mockCapability) GetAggregator(workflowID string) (pbtypes.Aggregator, error) {
	return m.aggregators[workflowID], nil
}

func (m *mockCapability) GetEncoderByWorkflowID(workflowID string) (pbtypes.Encoder, error) {
	return nil, nil // Not used in benchmark
}

func (m *mockCapability) GetEncoderByName(encoderName string, config *values.Map) (pbtypes.Encoder, error) {
	return nil, nil // Not used in benchmark
}

func (m *mockCapability) GetRegisteredWorkflowsIDs() []string {
	ids := make([]string, 0, len(m.aggregators))
	for id := range m.aggregators {
		ids = append(ids, id)
	}
	return ids
}

func (m *mockCapability) UnregisterWorkflowID(workflowID string) {
	delete(m.aggregators, workflowID)
}

func BenchmarkReportingPlugin_Outcome_LLOAggregator(b *testing.B) {
	// Define test matrix parameters
	workflowCounts := []int{1, 2, 4, 8, 16, 32, 64, 128}
	streamCounts := []int{32, 64, 128, 256, 512, 1024}

	// Create one logger for all benchmarks to reduce setup overhead
	c := logger.Config{
		Level: zapcore.InfoLevel, // Set to InfoLevel for benchmarks to reduce log noise
	}
	lggr, err := c.New()
	require.NoError(b, err, "failed to create logger for benchmark")

	// Run benchmarks for each combination
	for _, numWorkflows := range workflowCounts {
		for _, numStreamsPerWorkflow := range streamCounts {
			benchName := fmt.Sprintf("workflows=%d/streams=%d", numWorkflows, numStreamsPerWorkflow)
			b.Run(benchName, func(b *testing.B) {
				runBenchmarkWithParams(b, lggr, numWorkflows, numStreamsPerWorkflow)
			})
		}
	}
}

func BenchmarkReportingPlugin_Observation_LLOAggregator(b *testing.B) {
	// Define test matrix parameters
	workflowCounts := []int{1, 2, 4, 8, 16, 32, 64, 128}
	streamCounts := []int{32, 64, 128, 256, 512, 1024}

	// Create one logger for all benchmarks to reduce setup overhead
	c := logger.Config{
		Level: zapcore.InfoLevel, // Set to InfoLevel for benchmarks to reduce log noise
	}
	lggr, err := c.New()
	require.NoError(b, err, "failed to create logger for benchmark")

	// Run benchmarks for each combination
	for _, numWorkflows := range workflowCounts {
		for _, numStreamsPerWorkflow := range streamCounts {
			benchName := fmt.Sprintf("workflows=%d/streams=%d", numWorkflows, numStreamsPerWorkflow)
			b.Run(benchName, func(b *testing.B) {
				runObservationBenchmarkWithParams(b, lggr, numWorkflows, numStreamsPerWorkflow)
			})
		}
	}
}

// runObservationBenchmarkWithParams runs a benchmark with the specified parameters
func runObservationBenchmarkWithParams(b *testing.B, lggr logger.Logger, numWorkflows, numStreamsPerWorkflow int) {
	const (
		numOracles = 4 // Total nodes
		f          = 1 // Fault tolerance
	)

	// Create request store with requests for each workflow
	store := requests.NewStore[*ocr3.ReportRequest]()

	// Create capability with LLO aggregators for each workflow
	mockCap := &mockCapability{
		aggregators: make(map[string]pbtypes.Aggregator, numWorkflows),
	}

	// Create LLO aggregators for each workflow and populate the store
	for i := 0; i < numWorkflows; i++ {
		workflowID := fmt.Sprintf("workflow-%d", i)
		executionID := fmt.Sprintf("execution-%d", i)

		// Create aggregator
		agg, err := createLLOAggregator(b, numStreamsPerWorkflow)
		require.NoError(b, err)
		mockCap.aggregators[workflowID] = agg

		// Populate store with observation data
		lloEvent := createLLOEvent(b, numStreamsPerWorkflow, time.Now())
		wrappedEvent, err := values.Wrap(lloEvent)
		require.NoError(b, err)

		// Create list with the LLO event
		listVal, err := values.NewList([]interface{}{wrappedEvent})
		require.NoError(b, err)

		// Create and add request to store
		req := &ocr3.ReportRequest{
			WorkflowID:               workflowID,
			WorkflowExecutionID:      executionID,
			WorkflowName:             fmt.Sprintf("Workflow %d", i),
			WorkflowOwner:            "test-owner",
			WorkflowDonID:            1,
			WorkflowDonConfigVersion: 1,
			ReportID:                 fmt.Sprintf("report-%d", i),
			KeyID:                    "test-key",
			Observations:             listVal,
		}

		require.NoError(b, store.Add(req))
	}

	// Create reporting plugin
	plugin, err := ocr3.NewReportingPlugin(
		store,
		mockCap,
		numWorkflows, // batchSize matches numWorkflows
		ocr3types.ReportingPluginConfig{
			N: numOracles,
			F: f,
		},
		&pbtypes.ReportingPluginConfig{
			OutcomePruningThreshold: 100,
		},
		lggr,
	)
	require.NoError(b, err)

	// Create test query with workflow IDs
	query, err := createTestQuery(numWorkflows)
	require.NoError(b, err)

	// Create outcome context (not really used for Observation)
	outctx := ocr3types.OutcomeContext{
		SeqNr:           1,
		PreviousOutcome: nil, // Not needed for Observation benchmark
	}

	// Reset timer and enable memory allocation reporting
	b.ResetTimer()
	b.ReportAllocs()

	// Preallocate memory stats variables
	var memStatsBefore, memStatsAfter runtime.MemStats

	// Track cumulative metrics
	var totalMemUsage uint64
	var totalObservationSize int

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		runtime.GC() // Run garbage collection before measurement to reduce noise
		runtime.ReadMemStats(&memStatsBefore)

		// Call Observation function
		observation, err := plugin.Observation(context.Background(), outctx, query)
		require.NoError(b, err)

		// Measure memory usage
		runtime.ReadMemStats(&memStatsAfter)
		memUsage := memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc
		totalMemUsage += memUsage

		// Measure observation size
		observationSize := len(observation)
		totalObservationSize += observationSize

		// Basic validation of observation
		var parsedObservation pbtypes.Observations
		err = proto.Unmarshal(observation, &parsedObservation)
		require.NoError(b, err)
		require.Len(b, parsedObservation.Observations, numWorkflows)
	}

	// Report average metrics
	if b.N > 0 {
		b.ReportMetric(float64(totalMemUsage)/float64(b.N), "B/memory")
		b.ReportMetric(float64(totalObservationSize)/float64(b.N), "B/observation_size")
		// Report streams per second metric to understand throughput
		streamsProcessed := numWorkflows * numStreamsPerWorkflow
		b.ReportMetric(float64(streamsProcessed), "streams/op")
	}
}

// runBenchmarkWithParams runs a benchmark with the specified parameters
func runBenchmarkWithParams(b *testing.B, lggr logger.Logger, numWorkflows, numStreamsPerWorkflow int) {
	// Test parameters
	const (
		numOracles = 4 // Total nodes
		f          = 1 // Fault tolerance
	)

	// Create request store
	store := requests.NewStore[*ocr3.ReportRequest]()

	// Create capability with LLO aggregators for each workflow
	mockCap := &mockCapability{
		aggregators: make(map[string]pbtypes.Aggregator, numWorkflows),
	}

	// Create LLO aggregators for each workflow
	for i := 0; i < numWorkflows; i++ {
		workflowID := fmt.Sprintf("workflow-%d", i)
		agg, err := createLLOAggregator(b, numStreamsPerWorkflow)
		require.NoError(b, err)
		mockCap.aggregators[workflowID] = agg
	}

	// Create reporting plugin
	plugin, err := ocr3.NewReportingPlugin(
		store,
		mockCap,
		numWorkflows, // batchSize
		ocr3types.ReportingPluginConfig{
			N: numOracles,
			F: f,
		},
		&pbtypes.ReportingPluginConfig{
			OutcomePruningThreshold: 100,
		},
		lggr,
	)
	require.NoError(b, err)

	// Create test query with 10 workflow IDs
	query, err := createTestQuery(numWorkflows)
	require.NoError(b, err)

	// Create previous outcome with the same 10 workflow IDs
	previousOutcome, err := createTestPreviousOutcome(numWorkflows, numStreamsPerWorkflow)
	require.NoError(b, err)

	// Create attributed observations from all oracles
	aos := createTestAttributedObservations(b, numOracles, numWorkflows, numStreamsPerWorkflow)

	// Create outcome context
	outctx := ocr3types.OutcomeContext{
		SeqNr:           1,
		PreviousOutcome: previousOutcome,
	}

	// Reset timer and enable memory allocation reporting
	b.ResetTimer()
	b.ReportAllocs()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		var memStatsBefore, memStatsAfter runtime.MemStats
		runtime.ReadMemStats(&memStatsBefore)

		// Call Outcome function
		outcome, err := plugin.Outcome(context.Background(), outctx, query, aos)
		require.NoError(b, err)

		// Measure memory usage
		runtime.ReadMemStats(&memStatsAfter)
		memUsage := memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc

		// Measure outcome size
		outcomeSize := len(outcome)

		// Report custom metrics
		b.ReportMetric(float64(memUsage), "B/memory")
		b.ReportMetric(float64(outcomeSize), "B/outcome_size")

		// Validate outcome contents
		var parsedOutcome pbtypes.Outcome
		err = proto.Unmarshal(outcome, &parsedOutcome)
		require.NoError(b, err)
		require.Len(b, parsedOutcome.Outcomes, numWorkflows)
	}
}

// Helper functions

// createTestQuery generates a query with the specified number of workflow IDs
func createTestQuery(numWorkflows int) ([]byte, error) {
	ids := make([]*pbtypes.Id, numWorkflows)
	for i := 0; i < numWorkflows; i++ {
		ids[i] = &pbtypes.Id{
			WorkflowExecutionId:      fmt.Sprintf("execution-%d", i),
			WorkflowId:               fmt.Sprintf("workflow-%d", i),
			WorkflowOwner:            "test-owner",
			WorkflowName:             fmt.Sprintf("Workflow %d", i),
			WorkflowDonId:            1,
			WorkflowDonConfigVersion: 1,
			ReportId:                 fmt.Sprintf("report-%d", i),
			KeyId:                    "test-key",
		}
	}

	query := &pbtypes.Query{
		Ids: ids,
	}

	return proto.MarshalOptions{Deterministic: true}.Marshal(query)
}

// createTestPreviousOutcome generates a previous outcome with consistent LLOOutcomeMetadata
func createTestPreviousOutcome(numWorkflows, numStreamsPerWorkflow int) ([]byte, error) {
	outcome := &pbtypes.Outcome{
		Outcomes:       make(map[string]*pbtypes.AggregationOutcome, numWorkflows),
		CurrentReports: []*pbtypes.Report{},
	}

	// Create an identical LLOOutcomeMetadata for all workflows
	baseMetadata := &datafeeds.LLOOutcomeMetadata{
		StreamInfo: make(map[uint32]*datafeeds.LLOStreamInfo, numStreamsPerWorkflow),
	}

	// Populate with stream info
	baseTime := time.Now().Add(-10 * time.Minute).UnixNano()
	zeroPrice, _ := decimal.Zero.MarshalBinary()

	for i := 0; i < numStreamsPerWorkflow; i++ {
		streamID := uint32(i)
		baseMetadata.StreamInfo[streamID] = &datafeeds.LLOStreamInfo{
			Timestamp: baseTime,
			Price:     zeroPrice,
		}
	}

	// Marshal once
	metadataBytes, err := proto.Marshal(baseMetadata)
	if err != nil {
		return nil, err
	}

	// Create outcome entries for each workflow, using the same metadata
	for i := 0; i < numWorkflows; i++ {
		workflowID := fmt.Sprintf("workflow-%d", i)
		outcome.Outcomes[workflowID] = &pbtypes.AggregationOutcome{
			Metadata:         metadataBytes,
			LastSeenAt:       1,
			ShouldReport:     false,
			Timestamp:        timestamppb.Now(),
			EncodableOutcome: nil, // Not needed for benchmark
		}
	}

	return proto.MarshalOptions{Deterministic: true}.Marshal(outcome)
}

// createTestAttributedObservations generates attributed observations from multiple oracles
func createTestAttributedObservations(b *testing.B, numOracles, numWorkflows, numStreamsPerWorkflow int) []types.AttributedObservation {
	aos := make([]types.AttributedObservation, numOracles)
	ts := timestamppb.Now() // Use a consistent timestamp for all observations to ensure consensus
	for oracle := 0; oracle < numOracles; oracle++ {
		observationsProto := &pbtypes.Observations{
			Observations:          make([]*pbtypes.Observation, numWorkflows),
			RegisteredWorkflowIds: make([]string, numWorkflows),
			Timestamp:             ts,
		}

		// Create an observation for each workflow
		for i := 0; i < numWorkflows; i++ {
			workflowID := fmt.Sprintf("workflow-%d", i)
			executionID := fmt.Sprintf("execution-%d", i)
			observationsProto.RegisteredWorkflowIds[i] = workflowID

			// Create LLO events
			lloEvent := createLLOEvent(b, numStreamsPerWorkflow, ts.AsTime())
			wrappedEvent, err := values.Wrap(lloEvent)
			require.NoError(b, err)

			// Create list value with the LLO event
			listVal, err := values.NewList([]interface{}{wrappedEvent})
			require.NoError(b, err)

			listProto := values.Proto(listVal).GetListValue()
			require.NotNil(b, listProto, "listProto should not be nil") // Ensure listProto is not nil

			// Add observation for this workflow
			observationsProto.Observations[i] = &pbtypes.Observation{
				Id: &pbtypes.Id{
					WorkflowExecutionId:      executionID,
					WorkflowId:               workflowID,
					WorkflowOwner:            "test-owner",
					WorkflowName:             fmt.Sprintf("Workflow %d", i),
					WorkflowDonId:            1,
					WorkflowDonConfigVersion: 1,
					ReportId:                 fmt.Sprintf("report-%d", i),
					KeyId:                    "test-key",
				},
				Observations: listProto,
			}
		}

		// Marshal the observations
		obsBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(observationsProto)
		require.NoError(b, err)

		// Create attributed observation
		aos[oracle] = types.AttributedObservation{
			Observation: obsBytes,
			Observer:    ocrcommon.OracleID(oracle),
		}
	}

	return aos
}

// createLLOEvent creates an LLO event with the specified number of streams
func createLLOEvent(b *testing.B, numStreams int, ts time.Time) *datastreams.LLOStreamsTriggerEvent {
	timestamp := uint64(ts.UnixNano())
	event := &datastreams.LLOStreamsTriggerEvent{
		ObservationTimestampNanoseconds: timestamp,
		Payload:                         make([]*datastreams.LLOStreamDecimal, 0, numStreams),
	}

	// Create stream values with consistent prices
	for i := 0; i < numStreams; i++ {
		price := decimal.NewFromInt(int64(100 + i%10)) // Use a few different price values
		binary, err := price.MarshalBinary()
		require.NoError(b, err)

		event.Payload = append(event.Payload, &datastreams.LLOStreamDecimal{
			StreamID: uint32(i),
			Decimal:  binary,
		})
	}

	return event
}

// createLLOAggregator creates an LLO aggregator with the specified number of streams
func createLLOAggregator(b *testing.B, numStreams int) (pbtypes.Aggregator, error) {
	// Create feed configs for all streams
	streamConfigs := make(map[string]datafeeds.FeedConfig, numStreams)
	for i := 0; i < numStreams; i++ {
		streamConfigs[fmt.Sprintf("%d", i)] = datafeeds.FeedConfig{
			//	Deviation:     decimal.NewFromFloat(0.01),     // 1% deviation threshold
			Heartbeat:     3600,                           // 1 hour heartbeat
			RemappedIDHex: fmt.Sprintf("0x%064x", i+1000), // Unique remapped ID
		}
	}

	// Create LLO config
	c := datafeeds.LLOAggregatorConfig{
		Streams: streamConfigs,
	}

	// Create LLO aggregator
	//return datafeeds.NewLLOAggregator(configMap)
	m, err := c.ToMap()
	if err != nil {
		// Handle error in creating LLO aggregator
		return nil, fmt.Errorf("failed to create LLO aggregator: %w", err)
	}
	return datafeeds.NewLLOAggregator(*m)
}
