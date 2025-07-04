package dontime

import (
	"fmt"
	"testing"
	"time"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

func newTestPluginConfig(t *testing.T) ocr3types.ReportingPluginConfig {
	offChainCfg := &pb.Config{
		MaxQueryLengthBytes:       defaultMaxPhaseOutputBytes,
		MaxObservationLengthBytes: defaultMaxPhaseOutputBytes,
		MaxReportLengthBytes:      defaultMaxPhaseOutputBytes,
		MaxBatchSize:              defaultBatchSize,
		MinTimeIncrease:           int64(defaultMinTimeIncrease),
		ExecutionRemovalTime:      durationpb.New(defaultExecutionRemovalTime),
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
	store := NewStore(DefaultRequestTimeout)
	config := newTestPluginConfig(t)
	ctx := t.Context()

	plugin, err := NewPlugin(store, config, lggr)
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

	t.Run("Batching with expired requests", func(t *testing.T) {
		// Generate request queue: 1-2(expired)-3-4(expired)-5-6(expired)
		var expiredRequestChs []chan Response
		for i := range 6 {
			executionID := fmt.Sprintf("workflow-%d", i)
			ch := make(chan Response, 1)
			request := &Request{
				ExpiresAt:           time.Now().Add(defaultExecutionRemovalTime),
				CallbackCh:          ch,
				WorkflowExecutionID: executionID,
				SeqNum:              0,
			}
			if i%2 == 0 {
				request.ExpiresAt = time.Now()
				expiredRequestChs = append(expiredRequestChs, ch)
			}
			err := store.requests.Add(request)
			require.NoError(t, err)
		}

		// Batch 3 requests and verify removal of expired requests
		plugin.batchSize = 3

		observation, err := plugin.Observation(ctx, outcomeCtx, query)
		require.NoError(t, err)

		// Validate Outcome from Observation
		obsProto := &pb.Observation{}
		err = proto.Unmarshal(observation, obsProto)
		require.NoError(t, err)
		require.NotEqual(t, 0, obsProto.Timestamp)

		expectedRequests := map[string]int64{
			"workflow-1": 0,
			"workflow-3": 0,
			"workflow-5": 0,
		}
		require.Equal(t, expectedRequests, obsProto.Requests)

		for _, ch := range expiredRequestChs {
			resp := <-ch
			require.Contains(t, resp.Err.Error(), "timeout exceeded: could not process request before expiry")
		}
	})
}

func TestPlugin_ValidateObservation(t *testing.T) {
	lggr := logger.Test(t)
	config := newTestPluginConfig(t)
	ctx := t.Context()

	t.Run("Valid Observation", func(t *testing.T) {
		store := NewStore(DefaultRequestTimeout)
		plugin, err := NewPlugin(store, config, lggr)
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
		plugin, err := NewPlugin(store, config, lggr)
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
	config := newTestPluginConfig(t)
	ctx := t.Context()

	plugin, err := NewPlugin(store, config, lggr)
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

func TestPlugin_FinishedExecutions(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore(DefaultRequestTimeout)
	config := newTestPluginConfig(t)
	ctx := t.Context()

	transmitter := NewTransmitter(lggr, store)
	plugin, err := NewPlugin(store, config, lggr)
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

	// TODO: Transmit should just delete expired requests
	t.Run("Transmit: delete removed executionIDs", func(t *testing.T) {
		r := ocr3types.ReportWithInfo[struct{}]{}
		r.Report, err = proto.Marshal(outcomeProto)
		require.NoError(t, err)
		err = transmitter.Transmit(ctx, types.ConfigDigest{}, 0, r, []types.AttributedOnchainSignature{})
		require.NoError(t, err)
	})
}

func TestPlugin_ExpiredRequest(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore(0)
	config := newTestPluginConfig(t)
	ctx := t.Context()

	plugin, err := NewPlugin(store, config, lggr)
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
