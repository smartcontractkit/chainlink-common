package dontime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
)

func TestTransmitter_TransmitDonTimeRequest(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore(DefaultRequestTimeout)
	ctx := t.Context()

	transmitter := NewTransmitter(lggr, store, "")

	timestamp := time.Now().UnixMilli()

	for _, tc := range []struct {
		name     string
		observed *pb.ObservedDonTimes
	}{
		{"unsequenced", &pb.ObservedDonTimes{Timestamps: []int64{timestamp - int64(time.Second), timestamp}}},
		{"sequenced", &pb.ObservedDonTimes{TimestampsBySequence: map[int64]int64{0: timestamp - int64(time.Second), 1: timestamp}}},
		{"both", &pb.ObservedDonTimes{Timestamps: []int64{timestamp - int64(time.Second), timestamp},
			TimestampsBySequence: map[int64]int64{0: timestamp - int64(time.Second), 1: timestamp}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Create request for second donTime in sequence
			executionID := "workflow-123"
			timeRequest := store.RequestDonTime(executionID, 1)

			outcome := &pb.Outcome{
				Timestamp:        timestamp,
				ObservedDonTimes: map[string]*pb.ObservedDonTimes{executionID: tc.observed},
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
		})
	}
}
