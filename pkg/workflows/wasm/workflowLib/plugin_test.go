package workflowLib

import (
	"testing"
	"time"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/workflowLib/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

func newTestPluginConfig(t *testing.T) ocr3types.ReportingPluginConfig {
	offChainCfg := &pb.WorkflowLibConfig{
		MaxQueryLengthBytes:       defaultMaxPhaseOutputBytes,
		MaxObservationLengthBytes: defaultMaxPhaseOutputBytes,
		MaxReportLengthBytes:      defaultMaxPhaseOutputBytes,
		MaxBatchSize:              defaultBatchSize,
		MinTimeIncrease:           int64(defaultMinTimeIncrease),
	}

	offChainCfgBytes, err := proto.Marshal(offChainCfg)
	if err != nil {
		t.Error(err)
	}

	return ocr3types.ReportingPluginConfig{
		N:                                       4,
		F:                                       1,
		OffchainConfig:                          offChainCfgBytes,
		MaxDurationObservation:                  defaultMinTimeIncrease,
		MaxDurationShouldAcceptAttestedReport:   0,
		MaxDurationShouldTransmitAcceptedReport: 0,
	}
}

func TestPlugin_Observation(t *testing.T) {
	lggr := logger.Test(t)
	store := NewDonTimeStore()
	config := newTestPluginConfig(t)
	ctx := t.Context()

	plugin, err := NewWorkflowLibPlugin(store, config, lggr)
	require.NoError(t, err)

	outcomeCtx := ocr3types.OutcomeContext{
		PreviousOutcome: []byte(""),
	}

	query, err := plugin.Query(ctx, outcomeCtx)
	require.NoError(t, err)

	// Add single request to queue
	executionID := "workflow-123"
	_ = store.RequestDonTime(executionID, 0)

	observation, err := plugin.Observation(ctx, outcomeCtx, query)
	require.NoError(t, err)

	// Validate Outcome from Observation
	obsProto := &pb.Observation{}
	err = proto.Unmarshal(observation, obsProto)
	require.NoError(t, err)
	require.NotEqual(t, 0, obsProto.Timestamp)

	expectedRequests := map[string]int64{
		executionID: 0,
	}
	require.Equal(t, expectedRequests, obsProto.Requests)
	require.Empty(t, obsProto.Finished)
}

func TestPlugin_ValidateObservation(t *testing.T) {
	lggr := logger.Test(t)
	config := newTestPluginConfig(t)
	ctx := t.Context()

	t.Run("Valid Observation", func(t *testing.T) {
		store := NewDonTimeStore()
		plugin, err := NewWorkflowLibPlugin(store, config, lggr)
		require.NoError(t, err)

		outcomeCtx := ocr3types.OutcomeContext{
			PreviousOutcome: []byte(""),
		}

		query, err := plugin.Query(ctx, outcomeCtx)
		require.NoError(t, err)

		// Add single request to queue
		executionID := "workflow-123"
		_ = store.RequestDonTime(executionID, 0)

		observation, err := plugin.Observation(ctx, outcomeCtx, query)
		require.NoError(t, err)

		ao := types.AttributedObservation{
			Observation: observation,
			Observer:    commontypes.OracleID(1),
		}

		err = plugin.ValidateObservation(ctx, outcomeCtx, query, ao)
		require.NoError(t, err)
	})

	t.Run("Invalid sequence number", func(t *testing.T) {
		store := NewDonTimeStore()
		plugin, err := NewWorkflowLibPlugin(store, config, lggr)
		require.NoError(t, err)

		outcomeCtx := ocr3types.OutcomeContext{
			PreviousOutcome: []byte(""),
		}

		query, err := plugin.Query(ctx, outcomeCtx)
		require.NoError(t, err)

		// Add single request to queue
		executionID := "workflow-123"
		_ = store.RequestDonTime(executionID, 1)

		observation, err := plugin.Observation(ctx, outcomeCtx, query)
		require.NoError(t, err)

		ao := types.AttributedObservation{
			Observation: observation,
			Observer:    commontypes.OracleID(1),
		}

		err = plugin.ValidateObservation(ctx, outcomeCtx, query, ao)
		require.ErrorContains(t, err, "request number 1 for id workflow-123 is greater than the number of observed don times 0")
	})
}

func TestPlugin_Outcome(t *testing.T) {
	lggr := logger.Test(t)
	store := NewDonTimeStore()
	config := newTestPluginConfig(t)
	ctx := t.Context()

	plugin, err := NewWorkflowLibPlugin(store, config, lggr)
	require.NoError(t, err)

	query, err := plugin.Query(ctx, ocr3types.OutcomeContext{PreviousOutcome: []byte("")})

	// Add single request to queue
	executionID := "workflow-123"
	_ = store.RequestDonTime(executionID, 0)

	timestamp := time.Now().UnixMilli()
	observations := []*pb.Observation{
		{
			Timestamp: timestamp,
			Requests: map[string]int64{
				executionID: 0,
			},
			Finished: []string{"workflow-abc"},
		},
		{
			Timestamp: timestamp - int64(time.Second),
			Requests: map[string]int64{
				executionID: 0,
			},
			Finished: []string{"workflow-abc"},
		},
		{
			Timestamp: timestamp + int64(time.Second),
			Requests: map[string]int64{
				executionID: 0,
			},
			Finished: []string{"workflow-abc"},
		},
		{
			Timestamp: timestamp,
			Requests: map[string]int64{
				executionID: 0,
			},
			Finished: []string{"workflow-abc"},
		},
	}

	aos := make([]types.AttributedObservation, 4)
	for i, observation := range observations {
		rawObs, err := proto.Marshal(observation)
		require.NoError(t, err)
		aos[i] = types.AttributedObservation{
			Observation: rawObs,
			Observer:    commontypes.OracleID(1),
		}
	}

	prevOutcome := &pb.Outcome{
		Timestamp: 0,
		ObservedDonTimes: map[string]*pb.ObservedDonTimes{
			executionID: {Timestamps: []int64{}},
		},
		FinishedExecutionRemovalTimes: make(map[string]int64),
		RemovedExecutionIDs:           make(map[string]bool),
	}

	prevOutcomeBytes, err := proto.Marshal(prevOutcome)
	require.NoError(t, err)

	outcome, err := plugin.Outcome(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes}, query, aos)
	require.NoError(t, err)

	outcomeProto := &pb.Outcome{}
	err = proto.Unmarshal(outcome, outcomeProto)
	require.NoError(t, err)
	require.Equal(t, timestamp, outcomeProto.Timestamp)
	require.Equal(t, []int64{timestamp}, outcomeProto.ObservedDonTimes[executionID].Timestamps)
}

func TestPlugin_FinishedExecutions(t *testing.T) {
	lggr := logger.Test(t)
	store := NewDonTimeStore()
	config := newTestPluginConfig(t)
	ctx := t.Context()

	plugin, err := NewWorkflowLibPlugin(store, config, lggr)
	require.NoError(t, err)

	query, err := plugin.Query(ctx, ocr3types.OutcomeContext{PreviousOutcome: []byte("")})

	store.ExecutionFinished("workflow-123")
	store.ExecutionFinished("workflow-abc")
	outcomeProto := &pb.Outcome{}

	t.Run("Observation: new finished executionIDs", func(t *testing.T) {
		prevOutcome := &pb.Outcome{
			Timestamp:        0,
			ObservedDonTimes: nil,
			// We have already scheduled workflow-123 for removal
			FinishedExecutionRemovalTimes: map[string]int64{
				"workflow-123": time.Now().UnixMilli(),
			},
			RemovedExecutionIDs: nil,
		}

		prevOutcomeBytes, err := proto.Marshal(prevOutcome)
		require.NoError(t, err)
		outcomeCtx := ocr3types.OutcomeContext{
			PreviousOutcome: prevOutcomeBytes,
		}

		query, err := plugin.Query(ctx, outcomeCtx)
		require.NoError(t, err)

		observation, err := plugin.Observation(ctx, outcomeCtx, query)
		require.NoError(t, err)

		obsProto := &pb.Observation{}
		err = proto.Unmarshal(observation, obsProto)
		require.NoError(t, err)
		require.Equal(t, []string{"workflow-abc"}, obsProto.Finished)

	})

	t.Run("Outcome: schedule for removal", func(t *testing.T) {
		timestamp := time.Now().UnixMilli()
		observations := []*pb.Observation{
			{
				Timestamp: timestamp,
				Requests:  map[string]int64{},
				Finished:  []string{"workflow-abc"},
			},
			{
				Timestamp: timestamp - int64(time.Second),
				Requests:  map[string]int64{},
				Finished:  []string{"workflow-abc"},
			},
			{
				Timestamp: timestamp + int64(time.Second),
				Requests:  map[string]int64{},
				Finished:  []string{"workflow-abc"},
			},
			{
				Timestamp: timestamp,
				Requests:  map[string]int64{},
				Finished:  []string{"workflow-abc"},
			},
		}

		aos := make([]types.AttributedObservation, 4)
		for i, observation := range observations {
			rawObs, err := proto.Marshal(observation)
			require.NoError(t, err)
			aos[i] = types.AttributedObservation{
				Observation: rawObs,
				Observer:    commontypes.OracleID(1),
			}
		}

		prevOutcome := &pb.Outcome{
			Timestamp:        0,
			ObservedDonTimes: map[string]*pb.ObservedDonTimes{},
			FinishedExecutionRemovalTimes: map[string]int64{
				"workflow-123": timestamp - int64(time.Second),
			},
			RemovedExecutionIDs: make(map[string]bool),
		}

		prevOutcomeBytes, err := proto.Marshal(prevOutcome)
		require.NoError(t, err)

		outcome, err := plugin.Outcome(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes}, query, aos)
		require.NoError(t, err)

		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)
		require.Contains(t, outcomeProto.FinishedExecutionRemovalTimes, "workflow-abc")
		// workflow-123 should be considered removed now and will be deleted from store during transmit
		require.NotContains(t, outcomeProto.FinishedExecutionRemovalTimes, "workflow-123")
		require.Contains(t, outcomeProto.RemovedExecutionIDs, "workflow-123")
	})

	t.Run("Transmit: delete removed executionIDs", func(t *testing.T) {
		r := ocr3types.ReportWithInfo[struct{}]{}
		r.Report, err = proto.Marshal(outcomeProto)
		require.NoError(t, err)
		err = plugin.Transmit(ctx, types.ConfigDigest{}, 0, r, []types.AttributedOnchainSignature{})
		require.NoError(t, err)
		require.Contains(t, store.finishedExecutionIDs, "workflow-abc")
		require.NotContains(t, store.finishedExecutionIDs, "workflow-123")
	})
}

func TestReportingPlugin_Transmit(t *testing.T) {
	lggr := logger.Test(t)
	store := NewDonTimeStore()
	config := newTestPluginConfig(t)
	ctx := t.Context()

	plugin, err := NewWorkflowLibPlugin(store, config, lggr)
	require.NoError(t, err)

	// Create request for second donTime in sequence
	executionID := "workflow-123"
	timeRequest := store.RequestDonTime(executionID, 1)

	timestamp := time.Now().UnixMilli()
	outcome := &pb.Outcome{
		Timestamp: timestamp,
		ObservedDonTimes: map[string]*pb.ObservedDonTimes{
			executionID: {Timestamps: []int64{timestamp - int64(time.Second), timestamp}},
		},
		FinishedExecutionRemovalTimes: make(map[string]int64),
		RemovedExecutionIDs:           make(map[string]bool),
	}

	r := ocr3types.ReportWithInfo[struct{}]{}
	r.Report, err = proto.Marshal(outcome)
	require.NoError(t, err)
	err = plugin.Transmit(ctx, types.ConfigDigest{}, 0, r, []types.AttributedOnchainSignature{})
	require.NoError(t, err)

	select {
	case donTimeResp := <-timeRequest:
		require.Equal(t, timestamp, donTimeResp.timestamp)
		require.Equal(t, executionID, donTimeResp.WorkflowExecutionID)
		require.Equal(t, 1, donTimeResp.seqNum)
		require.NoError(t, donTimeResp.Err)
	case <-ctx.Done():
		t.Fatal("failed to retrieve donTime from request channel")
	}

	require.Empty(t, store.Requests.Get(executionID))
}
