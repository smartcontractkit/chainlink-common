package dontime

// NOTE: This plugin is not meant to be used in production.
// It has locks where we shouldn't use them, and exists to express functionality.
// I also haven't carefully vetter how I use the locking to ensure it's 100% correct.

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/dontime/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/smartcontractkit/libocr/quorumhelper"
	"google.golang.org/protobuf/proto"
)

type Plugin struct {
	finishedExecutionIds map[string]bool
	mu                   sync.RWMutex

	donTimes            map[string][]int64
	outstandingRequests map[string]*outstandingRequest
	current             int64

	n               int
	f               int
	minTimeIncrease int64
}

type outstandingRequest struct {
	resp chan int64
	on   int
}

func NewPlugin(n, f int, minTimeIncrease int64) *Plugin {
	return &Plugin{
		finishedExecutionIds: map[string]bool{},
		donTimes:             map[string][]int64{},
		outstandingRequests:  map[string]*outstandingRequest{},
		n:                    n,
		f:                    f,
		minTimeIncrease:      minTimeIncrease,
	}
}

// It was easier to demo this in one struct so everything is together.
// We would likely want to separate them

var _ ocr3types.ReportingPlugin[struct{}] = (*Plugin)(nil)
var _ ocr3types.ContractTransmitter[struct{}] = (*Plugin)(nil)

func (p *Plugin) RequestDonTime(executionId string, requestId int) chan<- int64 {
	ch := make(chan int64, 1)
	p.mu.Lock()
	defer p.mu.Unlock()
	if times, ok := p.donTimes[executionId]; ok {
		if len(times) > requestId {
			ch <- times[requestId]
			close(ch)
			return ch
		}
	}

	p.outstandingRequests[executionId] = &outstandingRequest{
		resp: ch,
		on:   requestId,
	}
	return ch
}

func (p *Plugin) FinishExecution(executionId string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.finishedExecutionIds[executionId] = true
}

func (p *Plugin) LastObservedDonTime() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current
}

func (p *Plugin) Query(_ context.Context, _ ocr3types.OutcomeContext) (types.Query, error) {
	return nil, nil
}

func (p *Plugin) Observation(_ context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	priorOutcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, priorOutcome); err != nil {
		return nil, err
	}

	finished := make([]string, 0, len(p.finishedExecutionIds))
	for id, _ := range priorOutcome.FinishedExecutionRemovalTimes {
		if _, ok := p.finishedExecutionIds[id]; ok {
			delete(p.finishedExecutionIds, id)
		}
	}

	for id := range p.finishedExecutionIds {
		finished = append(finished, id)
	}

	requests := map[string]int64{}
	for id, request := range p.outstandingRequests {
		requests[id] = int64(request.on)
	}

	observation := &pb.Observation{
		Timestamp: time.Now().UTC().UnixMilli(),
		Requests:  requests,
		Finished:  finished,
	}

	return proto.Marshal(observation)
}

func (p *Plugin) ValidateObservation(_ context.Context, oc ocr3types.OutcomeContext, _ types.Query, ao types.AttributedObservation) error {
	observation := &pb.Observation{}
	if err := proto.Unmarshal(ao.Observation, observation); err != nil {
		return err
	}

	priorObservation := &pb.Outcome{}
	if err := proto.Unmarshal(oc.PreviousOutcome, priorObservation); err != nil {
		return err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	// A DON time beyond the expected number of requests is invalid, as there was never consensus on the prior request, which should be blocking.
	var err error
	for id, requestNumber := range observation.Requests {
		on := len(priorObservation.ObservedDonTimes[id].Timestamps)
		if requestNumber > int64(on) {
			err = errors.Join(err, fmt.Errorf("request number %d for id %s is greater than the number of observed don times %d", requestNumber, id, on))
		}
	}

	return err
}

func (p *Plugin) ObservationQuorum(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (quorumReached bool, err error) {
	return quorumhelper.ObservationCountReachesObservationQuorum(quorumhelper.QuorumTwoFPlusOne, p.n, p.f, aos), nil
}

func (p *Plugin) Outcome(_ context.Context, outctx ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	requests := map[string]int64{}
	finished := map[string]int64{}
	var times []int64

	for _, ao := range aos {
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			return nil, err
		}

		for id, requestNumber := range observation.Requests {
			on := p.donTimes[id]
			if requestNumber == int64(len(on)) {
				requests[id]++
			}
		}

		for _, id := range observation.Finished {
			finished[id]++
		}

		times = append(times, observation.Timestamp)
	}

	slices.Sort(times)
	donTime := times[len(times)/2]

	outcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, outcome); err != nil {
		return nil, err
	}

	if donTime < outcome.Timestamp+p.minTimeIncrease {
		donTime = outcome.Timestamp + p.minTimeIncrease
	}

	for id, numRequests := range requests {
		if numRequests > int64(p.f) {
			observedDonTimes, ok := outcome.ObservedDonTimes[id]
			if !ok {
				observedDonTimes = &pb.ObservedDonTimes{}
			}
			observedDonTimes.Timestamps = append(observedDonTimes.Timestamps, donTime)
			outcome.ObservedDonTimes[id] = observedDonTimes
		}
	}

	for id, numFinished := range finished {
		if numFinished >= int64(p.f) {
			if _, ok := outcome.FinishedExecutionRemovalTimes[id]; ok {
				continue
			}

			outcome.FinishedExecutionRemovalTimes[id] = donTime + int64(10*time.Minute*time.Second*time.Millisecond)
		}
	}

	for id, removeAt := range outcome.FinishedExecutionRemovalTimes {
		if removeAt <= donTime {
			delete(outcome.FinishedExecutionRemovalTimes, id)
		}
	}

	return proto.Marshal(outcome)
}

func (p *Plugin) Reports(ctx context.Context, seqNr uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportPlus[struct{}], error) {
	// left out, but
	panic("implement me")
}

func (p *Plugin) ShouldAcceptAttestedReport(ctx context.Context, seqNr uint64, reportWithInfo ocr3types.ReportWithInfo[struct{}]) (bool, error) {
	return true, nil
}

func (p *Plugin) ShouldTransmitAcceptedReport(ctx context.Context, seqNr uint64, reportWithInfo ocr3types.ReportWithInfo[struct{}]) (bool, error) {
	return true, nil
}

func (p *Plugin) Close() error {
	return nil
}

func (p *Plugin) Transmit(_ context.Context, _ types.ConfigDigest, _ uint64, r ocr3types.ReportWithInfo[struct{}], _ []types.AttributedOnchainSignature) error {
	outcome := &pb.Outcome{}
	if err := proto.Unmarshal(r.Report, outcome); err != nil {
		return err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	for id, request := range p.outstandingRequests {
		if len(outcome.ObservedDonTimes[id].Timestamps) > request.on {
			request.resp <- outcome.ObservedDonTimes[id].Timestamps[request.on]
		}
	}

	return nil
}

// FromAccount is unused?
func (p *Plugin) FromAccount(ctx context.Context) (types.Account, error) {
	return "", nil
}
