package workflowLib

import (
	"context"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/workflowLib/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

var _ ocr3types.ContractTransmitter[struct{}] = (*Transmitter)(nil)

// Transmitter is a custom transmitter for the OCR3 capability.
// When called it will transmit DonTime requests back to the caller
// and handle deletion of finished executionIDs.
type Transmitter struct {
	lggr      logger.Logger
	store     *DonTimeStore
	batchSize int
}

func NewTransmitter(lggr logger.Logger, store *DonTimeStore, batchSize int) *Transmitter {
	return &Transmitter{lggr: lggr, store: store, batchSize: batchSize}
}

func (t *Transmitter) Transmit(ctx context.Context, _ types.ConfigDigest, _ uint64, r ocr3types.ReportWithInfo[struct{}], _ []types.AttributedOnchainSignature) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	outcome := &pb.Outcome{}
	if err := proto.Unmarshal(r.Report, outcome); err != nil {
		return err
	}

	for id, observedDonTimes := range outcome.ObservedDonTimes {
		t.store.setDonTimes(id, observedDonTimes.Timestamps)
	}
	t.store.setLastObservedDonTime(outcome.Timestamp)

	requests, err := t.store.Requests.FirstN(t.batchSize)
	if err != nil {
		return err
	}

	for _, request := range requests {
		id := request.WorkflowExecutionID
		if _, ok := outcome.ObservedDonTimes[id]; !ok {
			continue
		}
		if len(outcome.ObservedDonTimes[id].Timestamps) > request.SeqNum {
			donTime := outcome.ObservedDonTimes[id].Timestamps[request.SeqNum]
			t.store.Requests.Evict(id) // Make space for next request before delivering
			request.SendResponse(ctx, DonTimeResponse{
				WorkflowExecutionID: id,
				SeqNum:              request.SeqNum,
				Timestamp:           donTime,
				Err:                 nil,
			})
		}
	}

	for id := range outcome.RemovedExecutionIDs {
		t.store.deleteExecutionID(id)
	}

	return nil
}

func (t *Transmitter) FromAccount(ctx context.Context) (types.Account, error) {
	return "", nil
}
