package dontime

import (
	"testing"
	"time"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
)

func newTestPluginOffchainConfig(t *testing.T) *pb.Config {
	return &pb.Config{
		MaxQueryLengthBytes:       defaultMaxPhaseOutputBytes,
		MaxObservationLengthBytes: defaultMaxPhaseOutputBytes,
		MaxReportLengthBytes:      defaultMaxPhaseOutputBytes,
		MaxBatchSize:              defaultBatchSize,
		MinTimeIncrease:           int64(defaultMinTimeIncrease),
		ExecutionRemovalTime:      durationpb.New(defaultExecutionRemovalTime),
	}
}

func newTestPluginConfig(t *testing.T) ocr3types.ReportingPluginConfig {
	offChainCfg := newTestPluginOffchainConfig(t)

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
	store := NewStore(DefaultRequestTimeout)
	config, offchainCfg := newTestPluginConfig(t), newTestPluginOffchainConfig(t)
	ctx := t.Context()

	plugin, err := NewPlugin(store, config, offchainCfg, lggr)
	require.NoError(t, err)

	outcomeCtx := ocr3types.OutcomeContext{
		PreviousOutcome: []byte(""),
	}

	query, err := plugin.Query(ctx, outcomeCtx)
	require.NoError(t, err)

	t.Run("Single request", func(t *testing.T) {
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
		store.deleteExecutionID("workflow-123")
	})
}

func TestPlugin_ValidateObservation(t *testing.T) {
	lggr := logger.Test(t)
	config, offchainCfg := newTestPluginConfig(t), newTestPluginOffchainConfig(t)
	ctx := t.Context()

	t.Run("Valid Observation", func(t *testing.T) {
		store := NewStore(DefaultRequestTimeout)
		plugin, err := NewPlugin(store, config, offchainCfg, lggr)
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
		store := NewStore(DefaultRequestTimeout)
		plugin, err := NewPlugin(store, config, offchainCfg, lggr)
		require.NoError(t, err)

		outcomeCtx := ocr3types.OutcomeContext{
			PreviousOutcome: []byte(""),
		}

		query, err := plugin.Query(ctx, outcomeCtx)
		require.NoError(t, err)

		// Add single request to queue
		executionID := "workflow-123"
		requestCh := store.RequestDonTime(executionID, 1)

		_, err = plugin.Observation(ctx, outcomeCtx, query)
		require.NoError(t, err)

		response := <-requestCh
		require.ErrorContains(t, response.Err, "requested seqNum 1 for executionID workflow-123 is greater than the number of observed don times 0")
	})
}

func TestPlugin_Outcome(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore(DefaultRequestTimeout)
	config, offchainCfg := newTestPluginConfig(t), newTestPluginOffchainConfig(t)
	ctx := t.Context()

	plugin, err := NewPlugin(store, config, offchainCfg, lggr)
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
		},
		{
			Timestamp: timestamp - int64(time.Second),
			Requests: map[string]int64{
				executionID: 0,
			},
		},
		{
			Timestamp: timestamp + int64(time.Second),
			Requests: map[string]int64{
				executionID: 0,
			},
		},
		{
			Timestamp: timestamp,
			Requests: map[string]int64{
				executionID: 0,
			},
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

func TestPlugin_Outcome_SequenceNumberHandling(t *testing.T) {
	lggr := logger.Test(t)
	config, offchainCfg := newTestPluginConfig(t), newTestPluginOffchainConfig(t)
	ctx := t.Context()

	makeObservations := func(t *testing.T, timestamp int64, requests map[string]int64, numNodes int) []types.AttributedObservation {
		t.Helper()
		aos := make([]types.AttributedObservation, numNodes)
		for i := 0; i < numNodes; i++ {
			obs := &pb.Observation{
				Timestamp: timestamp + int64(i),
				Requests:  requests,
			}
			rawObs, err := proto.Marshal(obs)
			require.NoError(t, err)
			aos[i] = types.AttributedObservation{
				Observation: rawObs,
				Observer:    commontypes.OracleID(i),
			}
		}
		return aos
	}

	t.Run("new execution ID not in previous outcome defaults currSeqNum to 0", func(t *testing.T) {
		store := NewStore(DefaultRequestTimeout)
		plugin, err := NewPlugin(store, config, offchainCfg, lggr)
		require.NoError(t, err)

		executionID := "new-workflow"
		_ = store.RequestDonTime(executionID, 0)

		timestamp := time.Now().UnixMilli()
		aos := makeObservations(t, timestamp, map[string]int64{executionID: 0}, 4)

		prevOutcome := &pb.Outcome{
			Timestamp:        0,
			ObservedDonTimes: map[string]*pb.ObservedDonTimes{},
		}
		prevOutcomeBytes, err := proto.Marshal(prevOutcome)
		require.NoError(t, err)

		query, err := plugin.Query(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes})
		require.NoError(t, err)

		outcome, err := plugin.Outcome(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes}, query, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		require.Contains(t, outcomeProto.ObservedDonTimes, executionID)
		require.Len(t, outcomeProto.ObservedDonTimes[executionID].Timestamps, 1)
	})

	t.Run("nil ObservedDonTimes in previous outcome does not panic", func(t *testing.T) {
		store := NewStore(DefaultRequestTimeout)
		plugin, err := NewPlugin(store, config, offchainCfg, lggr)
		require.NoError(t, err)

		executionID := "nil-map-workflow"
		_ = store.RequestDonTime(executionID, 0)

		timestamp := time.Now().UnixMilli()
		aos := makeObservations(t, timestamp, map[string]int64{executionID: 0}, 4)

		prevOutcome := &pb.Outcome{
			Timestamp:        0,
			ObservedDonTimes: nil,
		}
		prevOutcomeBytes, err := proto.Marshal(prevOutcome)
		require.NoError(t, err)

		query, err := plugin.Query(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes})
		require.NoError(t, err)

		outcome, err := plugin.Outcome(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes}, query, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		require.Contains(t, outcomeProto.ObservedDonTimes, executionID)
		require.Len(t, outcomeProto.ObservedDonTimes[executionID].Timestamps, 1)
	})

	t.Run("existing execution ID uses len(Timestamps) as currSeqNum", func(t *testing.T) {
		store := NewStore(DefaultRequestTimeout)
		plugin, err := NewPlugin(store, config, offchainCfg, lggr)
		require.NoError(t, err)

		executionID := "existing-workflow"
		_ = store.RequestDonTime(executionID, 1)

		timestamp := time.Now().UnixMilli()
		aos := makeObservations(t, timestamp, map[string]int64{executionID: 1}, 4)

		prevTimestamp := timestamp - 1000 // 1 second ago in millis
		prevOutcome := &pb.Outcome{
			Timestamp: prevTimestamp,
			ObservedDonTimes: map[string]*pb.ObservedDonTimes{
				executionID: {Timestamps: []int64{prevTimestamp}},
			},
		}
		prevOutcomeBytes, err := proto.Marshal(prevOutcome)
		require.NoError(t, err)

		query, err := plugin.Query(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes})
		require.NoError(t, err)

		outcome, err := plugin.Outcome(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes}, query, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		require.Contains(t, outcomeProto.ObservedDonTimes, executionID)
		require.Len(t, outcomeProto.ObservedDonTimes[executionID].Timestamps, 2)
	})

	t.Run("stale sequence number is ignored", func(t *testing.T) {
		store := NewStore(DefaultRequestTimeout)
		plugin, err := NewPlugin(store, config, offchainCfg, lggr)
		require.NoError(t, err)

		executionID := "stale-workflow"

		timestamp := time.Now().UnixMilli()
		// Observations report seqNum 0, but prevOutcome already has 2 timestamps (currSeqNum=2)
		aos := makeObservations(t, timestamp, map[string]int64{executionID: 0}, 4)

		prevTimestamp := timestamp - 1000 // 1 second ago in millis
		prevOutcome := &pb.Outcome{
			Timestamp: prevTimestamp,
			ObservedDonTimes: map[string]*pb.ObservedDonTimes{
				executionID: {Timestamps: []int64{
					prevTimestamp - 1000,
					prevTimestamp,
				}},
			},
		}
		prevOutcomeBytes, err := proto.Marshal(prevOutcome)
		require.NoError(t, err)

		query, err := plugin.Query(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes})
		require.NoError(t, err)

		outcome, err := plugin.Outcome(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes}, query, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		// Stale seqNum 0 should be ignored, so timestamps should remain unchanged at 2
		require.Len(t, outcomeProto.ObservedDonTimes[executionID].Timestamps, 2)
	})

	t.Run("mix of new and existing execution IDs", func(t *testing.T) {
		store := NewStore(DefaultRequestTimeout)
		plugin, err := NewPlugin(store, config, offchainCfg, lggr)
		require.NoError(t, err)

		existingID := "existing-workflow"
		newID := "new-workflow"
		_ = store.RequestDonTime(existingID, 1)
		_ = store.RequestDonTime(newID, 0)

		timestamp := time.Now().UnixMilli()
		requests := map[string]int64{
			existingID: 1,
			newID:      0,
		}
		aos := makeObservations(t, timestamp, requests, 4)

		prevTimestamp := timestamp - 1000 // 1 second ago in millis
		prevOutcome := &pb.Outcome{
			Timestamp: prevTimestamp,
			ObservedDonTimes: map[string]*pb.ObservedDonTimes{
				existingID: {Timestamps: []int64{prevTimestamp}},
			},
		}
		prevOutcomeBytes, err := proto.Marshal(prevOutcome)
		require.NoError(t, err)

		query, err := plugin.Query(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes})
		require.NoError(t, err)

		outcome, err := plugin.Outcome(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes}, query, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		require.Contains(t, outcomeProto.ObservedDonTimes, existingID)
		require.Len(t, outcomeProto.ObservedDonTimes[existingID].Timestamps, 2)
		require.Contains(t, outcomeProto.ObservedDonTimes, newID)
		require.Len(t, outcomeProto.ObservedDonTimes[newID].Timestamps, 1)
	})
}

func TestPlugin_FinishedExecutions(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore(DefaultRequestTimeout)
	config, offchainCfg := newTestPluginConfig(t), newTestPluginOffchainConfig(t)
	ctx := t.Context()

	transmitter := NewTransmitter(lggr, store, "")
	plugin, err := NewPlugin(store, config, offchainCfg, lggr)
	require.NoError(t, err)

	query, err := plugin.Query(ctx, ocr3types.OutcomeContext{PreviousOutcome: []byte("")})
	outcomeProto := &pb.Outcome{}

	t.Run("Outcome: remove expired workflow executions", func(t *testing.T) {
		timestamp := time.Now().UnixMilli()
		observations := []*pb.Observation{
			{
				Timestamp: timestamp,
				Requests:  map[string]int64{},
			},
			{
				Timestamp: timestamp - int64(time.Second),
				Requests:  map[string]int64{},
			},
			{
				Timestamp: timestamp + int64(time.Second),
				Requests:  map[string]int64{},
			},
			{
				Timestamp: timestamp,
				Requests:  map[string]int64{},
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

		// Set workflow-123 as expired
		prevDonTime := timestamp - int64(time.Second)
		prevOutcome := &pb.Outcome{
			Timestamp: prevDonTime,
			ObservedDonTimes: map[string]*pb.ObservedDonTimes{
				"workflow-123": {
					Timestamps: []int64{prevDonTime - defaultExecutionRemovalTime.Milliseconds()},
				},
			},
		}

		prevOutcomeBytes, err := proto.Marshal(prevOutcome)
		require.NoError(t, err)

		outcome, err := plugin.Outcome(ctx, ocr3types.OutcomeContext{PreviousOutcome: prevOutcomeBytes}, query, aos)
		require.NoError(t, err)

		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)
		require.NotContains(t, outcomeProto.ObservedDonTimes, "workflow-123")
	})

	t.Run("Transmit: delete removed executionIDs", func(t *testing.T) {
		store.setDonTimes("workflow-123", []int64{time.Now().UnixMilli()})

		r := ocr3types.ReportWithInfo[[]byte]{}
		r.Report, err = proto.Marshal(outcomeProto)
		require.NoError(t, err)
		err = transmitter.Transmit(ctx, types.ConfigDigest{}, 0, r, []types.AttributedOnchainSignature{})
		require.NoError(t, err)

		_, err = store.GetDonTimes("workflow-123")
		require.ErrorContains(t, err, "no don time for executionID workflow-123")
	})
}

func TestPlugin_ExpiredRequest(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore(0)
	config, offchainCfg := newTestPluginConfig(t), newTestPluginOffchainConfig(t)
	ctx := t.Context()

	plugin, err := NewPlugin(store, config, offchainCfg, lggr)
	require.NoError(t, err)

	outcomeCtx := ocr3types.OutcomeContext{
		PreviousOutcome: []byte(""),
	}

	query, err := plugin.Query(ctx, outcomeCtx)
	require.NoError(t, err)

	// Add single request to queue
	executionID := "workflow-123"
	timeRequest := store.RequestDonTime(executionID, 0)

	_, err = plugin.Observation(ctx, outcomeCtx, query)
	require.NoError(t, err)

	select {
	case donTimeResp := <-timeRequest:
		require.ErrorContains(t, donTimeResp.Err, "timeout exceeded: could not process request before expiry")
	case <-ctx.Done():
		t.Fatal("failed to retrieve donTime from request channel")
	}
}
