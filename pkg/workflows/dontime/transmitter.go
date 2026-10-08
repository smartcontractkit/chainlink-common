package dontime

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
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
		t.lggr.Errorf("failed to unmarshal report")
		return err
	}

	var total int
	currentDonTimes := make(map[string]map[int64]int64, len(outcome.ObservedDonTimes))
	for id, observedDonTimes := range outcome.ObservedDonTimes {
		if len(observedDonTimes.Timestamps) > 0 {
			m := make(map[int64]int64)
			for i, ts := range observedDonTimes.Timestamps {
				m[int64(i)] = ts
			}
			currentDonTimes[id] = m
		} else {
			currentDonTimes[id] = observedDonTimes.TimestampsBySequence
		}
		total += len(currentDonTimes[id])
	}
	t.store.replaceDonTimes(currentDonTimes)
	t.store.setLastObservedDonTime(outcome.Timestamp)

	t.lggr.Infow("Transmitting timestamps", "lastObservedDonTime", outcome.Timestamp, "executions", len(currentDonTimes), "total", total)

	for executionID, donTimes := range outcome.ObservedDonTimes {
		request := t.store.GetRequest(executionID)
		if request == nil {
			continue
		}

		// Nodes behind on multiple requests may wait one OCR round per request.
		// Caching future times locally could be added as an optimization.
		var donTime int64
		var ok bool
		if len(donTimes.TimestampsBySequence) > 0 {
			donTime, ok = donTimes.TimestampsBySequence[int64(request.SeqNum)]
		} else {
			ok = len(donTimes.Timestamps) > request.SeqNum
			if ok {
				donTime = donTimes.Timestamps[request.SeqNum]
				if donTime == 0 { // feature flag was disabled, and we had a gap in the sequence
					ok = false
				}
			}
		}
		if ok {
			t.store.RemoveRequest(executionID) // Make space for next request before delivering
			request.SendResponse(Response{
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
