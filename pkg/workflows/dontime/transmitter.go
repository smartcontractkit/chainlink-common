package dontime

import (
	"context"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

var _ ocr3types.ContractTransmitter[[]byte] = (*Transmitter)(nil)

// Transmitter is a custom transmitter for the OCR3 capability.
// When called it will transmit DonTime requests back to the caller
// and handle deletion of finished executionIDs.
type Transmitter struct {
	lggr        logger.Logger
	store       *Store
	fromAccount types.Account
}

func NewTransmitter(lggr logger.Logger, store *Store, fromAccount types.Account) *Transmitter {
	return &Transmitter{lggr: lggr, store: store, fromAccount: fromAccount}
}

func (t *Transmitter) Transmit(_ context.Context, _ types.ConfigDigest, _ uint64, r ocr3types.ReportWithInfo[[]byte], _ []types.AttributedOnchainSignature) error {
	outcome := &pb.Outcome{}
	if err := proto.Unmarshal(r.Report, outcome); err != nil {
		return err
	}

	for id, observedDonTimes := range outcome.ObservedDonTimes {
		t.store.setDonTimes(id, observedDonTimes.Timestamps)
	}
	t.store.setLastObservedDonTime(outcome.Timestamp)

	t.lggr.Infow("Transmitting timestamps", "lastObservedDonTime", outcome.Timestamp)

	for executionID, donTimes := range outcome.ObservedDonTimes {
		request := t.store.GetRequest(executionID)
		if request == nil {
			continue
		}

		// Nodes behind on multiple requests may wait one OCR round per request.
		// Caching future times locally could be added as an optimization.
		if len(donTimes.Timestamps) > request.SeqNum {
			donTime := donTimes.Timestamps[request.SeqNum]
			t.store.requests.Evict(executionID) // Make space for next request before delivering
			request.SendResponse(nil, Response{
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
	return t.fromAccount, nil
}
