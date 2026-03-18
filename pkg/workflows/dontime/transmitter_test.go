package dontime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func TestTransmitter_TransmitDonTimeRequest(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore(DefaultRequestTimeout)
	ctx := t.Context()

	transmitter := NewTransmitter(lggr, store, "")

	// Create request for second donTime in sequence
	executionID := "workflow-123"
	timeRequest := store.RequestDonTime(executionID, 1)

	timestamp := time.Now().UnixMilli()
	outcome := &pb.Outcome{
		Timestamp: timestamp,
		ObservedDonTimes: map[string]*pb.ObservedDonTimes{
			executionID: {Timestamps: []int64{timestamp - int64(time.Second), timestamp}},
		},
	}

	r := ocr3types.ReportWithInfo[[]byte]{}
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

	require.Empty(t, store.GetRequest(executionID))
}

func TestTransmitter_TransmitPreservesCachedDonTimesForOmittedExecutionIDs(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore(DefaultRequestTimeout)
	ctx := t.Context()

	transmitter := NewTransmitter(lggr, store, "")

	store.setDonTimes("workflow-stale", []int64{11, 22})

	timestamp := time.Now().UnixMilli()
	outcome := &pb.Outcome{
		Timestamp: timestamp,
		ObservedDonTimes: map[string]*pb.ObservedDonTimes{
			"workflow-fresh": {Timestamps: []int64{timestamp}},
		},
	}

	r := ocr3types.ReportWithInfo[[]byte]{}
	var err error
	r.Report, err = proto.Marshal(outcome)
	require.NoError(t, err)

	err = transmitter.Transmit(ctx, types.ConfigDigest{}, 0, r, []types.AttributedOnchainSignature{})
	require.NoError(t, err)

	staleDonTimes, err := store.GetDonTimes("workflow-stale")
	require.NoError(t, err)
	require.Equal(t, []int64{11, 22}, staleDonTimes)

	freshDonTimes, err := store.GetDonTimes("workflow-fresh")
	require.NoError(t, err)
	require.Equal(t, []int64{timestamp}, freshDonTimes)
}
