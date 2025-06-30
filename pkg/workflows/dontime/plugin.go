package dontime

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/smartcontractkit/libocr/quorumhelper"
	"google.golang.org/protobuf/proto"
)

type Plugin struct {
	mu sync.RWMutex

	store          *Store
	config         ocr3types.ReportingPluginConfig
	offChainConfig *pb.Config
	lggr           logger.Logger

	batchSize       int
	minTimeIncrease int64
}

var _ ocr3types.ReportingPlugin[struct{}] = (*Plugin)(nil)

func NewPlugin(store *Store, config ocr3types.ReportingPluginConfig, lggr logger.Logger) (*Plugin, error) {
	offchainCfg := &pb.Config{}
	err := proto.Unmarshal(config.OffchainConfig, offchainCfg)
	if err != nil {
		return nil, err
	}
	if offchainCfg.MaxBatchSize == 0 {
		return nil, errors.New("batch size cannot be 0")
	}
	if offchainCfg.MinTimeIncrease <= 0 {
		return nil, errors.New("minimum time increase must be positive")
	}
	if offchainCfg.ExecutionRemovalTime.AsDuration() <= 0 {
		return nil, errors.New("execution removal time must be positive")
	}

	return &Plugin{
		store:           store,
		config:          config,
		offChainConfig:  offchainCfg,
		lggr:            logger.Named(lggr, "WorkflowLibraryPlugin"),
		batchSize:       int(offchainCfg.MaxBatchSize),
		minTimeIncrease: offchainCfg.MinTimeIncrease,
	}, nil
}

func (p *Plugin) Query(_ context.Context, _ ocr3types.OutcomeContext) (types.Query, error) {
	return nil, nil
}

func (p *Plugin) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	previousOutcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, previousOutcome); err != nil {
		return nil, err
	}

	finishedExecutionIDs := p.store.GetFinishedExecutionIDs()
	var unscheduledFinishedExecutionIDs []string
	for id := range finishedExecutionIDs {
		if _, ok := previousOutcome.FinishedExecutionRemovalTimes[id]; !ok {
			unscheduledFinishedExecutionIDs = append(unscheduledFinishedExecutionIDs, id)
		}
	}

	// Collect up to batchSize unexpired requests
	requests := map[string]int64{} // Maps executionID --> seqNum
	for batchOffset := 0; batchOffset < p.store.requests.Len() && len(requests) < p.batchSize; {
		batch, err := p.store.requests.RangeN(batchOffset, p.batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to get request batch: %w", err)
		}
		if len(batch) == 0 {
			break
		}

		timeoutCheck := time.Now()
		for _, req := range batch {
			if req.ExpiryTime().Before(timeoutCheck) {
				// Request has been sitting in queue too long
				p.store.requests.Evict(req.ID())
				req.SendTimeout(ctx)
				continue
			}
			requests[req.WorkflowExecutionID] = int64(req.SeqNum)
			batchOffset++
		}
	}

	observation := &pb.Observation{
		Timestamp: time.Now().UTC().UnixMilli(),
		Requests:  requests,
		Finished:  unscheduledFinishedExecutionIDs,
	}

	return proto.Marshal(observation)
}

func (p *Plugin) ValidateObservation(_ context.Context, oc ocr3types.OutcomeContext, _ types.Query, ao types.AttributedObservation) error {
	observation := &pb.Observation{}
	if err := proto.Unmarshal(ao.Observation, observation); err != nil {
		return err
	}

	priorOutcome := &pb.Outcome{}
	if err := proto.Unmarshal(oc.PreviousOutcome, priorOutcome); err != nil {
		return err
	}
	if priorOutcome.ObservedDonTimes == nil {
		priorOutcome.ObservedDonTimes = map[string]*pb.ObservedDonTimes{}
	}

	newInvalidRequestError := func(requestSeqNum int64, id string, currSeqNum int) error {
		return fmt.Errorf("request number %d for id %s is greater than the number of observed don times %d", requestSeqNum, id, currSeqNum)
	}

	// A DON time beyond the expected number of requests is invalid, as there was never consensus on the prior request, which should be blocking.
	var err error
	for id, requestedSeqNumber := range observation.Requests {
		times, ok := priorOutcome.ObservedDonTimes[id]
		if !ok {
			if requestedSeqNumber != 0 {
				err = errors.Join(err, newInvalidRequestError(requestedSeqNumber, id, 0))
			}
			continue
		}
		seqNum := len(times.Timestamps)
		if requestedSeqNumber > int64(seqNum) {
			err = errors.Join(err, newInvalidRequestError(requestedSeqNumber, id, seqNum))
		}
	}

	return err
}

func (p *Plugin) ObservationQuorum(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (quorumReached bool, err error) {
	return quorumhelper.ObservationCountReachesObservationQuorum(quorumhelper.QuorumTwoFPlusOne, p.config.N, p.config.F, aos), nil
}

func (p *Plugin) Outcome(ctx context.Context, outctx ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	NumFinishedRequests := map[string]int64{} // counts how many nodes reported where a new DON timestamp might be needed
	finishedNodes := map[string]int64{}       // counts number of nodes finished with the workflow for executionID
	var times []int64

	prevOutcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, prevOutcome); err != nil {
		return nil, err
	}

	for _, ao := range aos {
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			return nil, err
		}

		for id, currSeqNum := range observation.Requests {
			if _, ok := prevOutcome.ObservedDonTimes[id]; !ok {
				prevOutcome.ObservedDonTimes[id] = &pb.ObservedDonTimes{}
			}
			if currSeqNum == int64(len(prevOutcome.ObservedDonTimes[id].Timestamps)) {
				NumFinishedRequests[id]++
			}
		}

		for _, id := range observation.Finished {
			finishedNodes[id]++
		}

		times = append(times, observation.Timestamp)
	}

	p.lggr.Debugw("Observed Node Timestamps", "timestamps", times)
	slices.Sort(times)
	donTime := times[len(times)/2]

	outcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, outcome); err != nil {
		return nil, err
	}

	if outcome.FinishedExecutionRemovalTimes == nil {
		outcome.FinishedExecutionRemovalTimes = make(map[string]int64)
	}
	if outcome.ObservedDonTimes == nil {
		outcome.ObservedDonTimes = make(map[string]*pb.ObservedDonTimes)
	}

	// Compare with prior outcome to ensure DON time never goes backward.
	if donTime < outcome.Timestamp+p.minTimeIncrease {
		p.lggr.Infow("DON Time incremented by minimum time increase to ensure time progression", "minTimeIncrease", p.minTimeIncrease)
		donTime = outcome.Timestamp + p.minTimeIncrease
	}

	p.lggr.Infow("New DON Time", "donTime", donTime)
	outcome.Timestamp = donTime

	for id, numRequests := range NumFinishedRequests {
		p.lggr.Debugw("Checking finished requests", "executionID", id, "numRequests", numRequests)
		if numRequests > int64(p.config.F) {
			observedDonTimes, ok := outcome.ObservedDonTimes[id]
			if !ok {
				observedDonTimes = &pb.ObservedDonTimes{}
			}
			observedDonTimes.Timestamps = append(observedDonTimes.Timestamps, donTime)
			outcome.ObservedDonTimes[id] = observedDonTimes
		}
	}

	// Check if consensus is reached on the workflow execution being finished
	for id, numFinished := range finishedNodes {
		if numFinished >= int64(p.config.F) {
			if _, ok := outcome.FinishedExecutionRemovalTimes[id]; ok {
				continue
			}
			outcome.FinishedExecutionRemovalTimes[id] = donTime + p.offChainConfig.ExecutionRemovalTime.AsDuration().Milliseconds()
		}
	}

	for id, removeAt := range outcome.FinishedExecutionRemovalTimes {
		if removeAt <= donTime {
			delete(outcome.FinishedExecutionRemovalTimes, id)
			p.store.deleteExecutionID(id)
		}
	}

	return proto.Marshal(outcome)
}

func (p *Plugin) Reports(_ context.Context, _ uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportPlus[struct{}], error) {
	return []ocr3types.ReportPlus[struct{}]{
		{
			ReportWithInfo: ocr3types.ReportWithInfo[struct{}]{
				Report: types.Report(outcome),
				Info:   struct{}{},
			},
			TransmissionScheduleOverride: nil,
		},
	}, nil
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
