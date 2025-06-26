package workflowLib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/workflowLib/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func TestTransmitter_TransmitDonTimeRequest(t *testing.T) {
	lggr := logger.Test(t)
	store := newDonTimeStore(defaultRequestTimeout)
	ctx := t.Context()

	transmitter := NewTransmitter(lggr, store, defaultBatchSize)

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
	var err error
	r.Report, err = proto.Marshal(outcome)
	require.NoError(t, err)
	err = transmitter.Transmit(ctx, types.ConfigDigest{}, 0, r, []types.AttributedOnchainSignature{})
	require.NoError(t, err)

	select {
	case donTimeResp := <-timeRequest:
		require.Equal(t, timestamp, donTimeResp.Timestamp)
		require.Equal(t, executionID, donTimeResp.WorkflowExecutionID)
		require.Equal(t, 1, donTimeResp.SeqNum)
		require.NoError(t, donTimeResp.Err)
	case <-ctx.Done():
		t.Fatal("failed to retrieve donTime from request channel")
	}

	require.Empty(t, store.Requests.Get(executionID))
}

func TestTransmitter_ExpiredRequest(t *testing.T) {
	lggr := logger.Test(t)
	store := newDonTimeStore(0) // Expire requests instantly
	ctx := t.Context()

	transmitter := NewTransmitter(lggr, store, defaultBatchSize)

	// Create request for second donTime in sequence
	executionID := "workflow-123"
	timeRequest := store.RequestDonTime(executionID, 1)

	timestamp := time.Now().UnixMilli()
	outcome := &pb.Outcome{
		Timestamp:                     timestamp,
		ObservedDonTimes:              map[string]*pb.ObservedDonTimes{},
		FinishedExecutionRemovalTimes: make(map[string]int64),
		RemovedExecutionIDs:           make(map[string]bool),
	}

	r := ocr3types.ReportWithInfo[struct{}]{}
	var err error
	r.Report, err = proto.Marshal(outcome)
	require.NoError(t, err)
	err = transmitter.Transmit(ctx, types.ConfigDigest{}, 0, r, []types.AttributedOnchainSignature{})
	require.NoError(t, err)

	select {
	case donTimeResp := <-timeRequest:
		require.ErrorContains(t, donTimeResp.Err, "timeout exceeded: could not process request before expiry")
	case <-ctx.Done():
		t.Fatal("failed to retrieve donTime from request channel")
	}
}
