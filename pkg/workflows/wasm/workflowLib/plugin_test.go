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

// TODO: Test validate observations
// TODO: Test edge cases
// TODO: Test executionID Removal

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
	t.Run("Removal Scheduling", func(t *testing.T) {

	})

	t.Run("Deletion", func(t *testing.T) {

	})
}

func TestReportingPlugin_Transmit(t *testing.T) {
	// TODO: Verify don time requests are fulfilled to their channels

}
