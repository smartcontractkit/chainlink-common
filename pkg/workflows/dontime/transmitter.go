package dontime

import (
	"context"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

var _ ocr3types.ContractTransmitter[struct{}] = (*Transmitter)(nil)

// Transmitter is a custom transmitter for the OCR3 capability.
// When called it will transmit DonTime requests back to the caller
// and handle deletion of finished executionIDs.
type Transmitter struct {
	lggr      logger.Logger
	store     *Store
	batchSize int
}

func NewTransmitter(lggr logger.Logger, store *Store, batchSize int) *Transmitter {
	return &Transmitter{lggr: lggr, store: store, batchSize: batchSize}
}

func (t *Transmitter) Transmit(ctx context.Context, _ types.ConfigDigest, _ uint64, r ocr3types.ReportWithInfo[struct{}], _ []types.AttributedOnchainSignature) error {
	outcome := &pb.Outcome{}
	if err := proto.Unmarshal(r.Report, outcome); err != nil {
		return err
	}

	for id, observedDonTimes := range outcome.ObservedDonTimes {
		t.store.setDonTimes(id, observedDonTimes.Timestamps)
	}
	t.store.setLastObservedDonTime(outcome.Timestamp)

	for executionID, donTimes := range outcome.ObservedDonTimes {
		request := t.store.GetRequest(executionID)
		if request == nil {
			continue
		}

		if len(donTimes.Timestamps) > request.SeqNum {
			donTime := donTimes.Timestamps[request.SeqNum]
			t.store.requests.Evict(executionID) // Make space for next request before delivering
			request.SendResponse(ctx, DonTimeResponse{
				WorkflowExecutionID: executionID,
				SeqNum:              request.SeqNum,
				Timestamp:           donTime,
				Err:                 nil,
			})
		}
	}

	return nil
}

func (t *Transmitter) FromAccount(ctx context.Context) (types.Account, error) {
	return "", nil
}
