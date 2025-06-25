package workflowLib

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/workflowLib/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/smartcontractkit/libocr/quorumhelper"
	"google.golang.org/protobuf/proto"
)

type workflowLibPlugin struct {
	mu sync.RWMutex

	store          *DonTimeStore
	config         ocr3types.ReportingPluginConfig
	offChainConfig *pb.WorkflowLibConfig
	lggr           logger.Logger

	batchSize       int
	minTimeIncrease int64
}

func NewWorkflowLibPlugin(store *DonTimeStore, config ocr3types.ReportingPluginConfig, lggr logger.Logger) (*workflowLibPlugin, error) {
	offchainCfg := &pb.WorkflowLibConfig{}
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

	return &workflowLibPlugin{
		store:           store,
		config:          config,
		offChainConfig:  offchainCfg,
		lggr:            logger.Named(lggr, "WorkflowLibraryPlugin"),
		batchSize:       int(offchainCfg.MaxBatchSize),
		minTimeIncrease: offchainCfg.MinTimeIncrease,
	}, nil
}

var _ ocr3types.ReportingPlugin[struct{}] = (*workflowLibPlugin)(nil)
var _ ocr3types.ContractTransmitter[struct{}] = (*workflowLibPlugin)(nil)

func (p *workflowLibPlugin) Query(_ context.Context, _ ocr3types.OutcomeContext) (types.Query, error) {
	return nil, nil
}

func (p *workflowLibPlugin) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	priorOutcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, priorOutcome); err != nil {
		return nil, err
	}

	finishedExecutionIDs := p.store.GetFinishedExecutionIDs()
	var unscheduledFinishedExecutionIDs []string
	for id := range finishedExecutionIDs {
		if _, ok := priorOutcome.FinishedExecutionRemovalTimes[id]; !ok {
			unscheduledFinishedExecutionIDs = append(unscheduledFinishedExecutionIDs, id)
		}
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	requests := map[string]int64{} // Maps executionID --> seqNum
	nextRequestsBatch, err := p.store.Requests.FirstN(p.batchSize)
	if err != nil {
		return nil, err
	}

	for _, req := range nextRequestsBatch {
		requests[req.WorkflowExecutionID] = int64(req.SeqNum)
	}

	observation := &pb.Observation{
		Timestamp: time.Now().UTC().UnixMilli(),
		Requests:  requests,
		Finished:  unscheduledFinishedExecutionIDs,
	}

	return proto.Marshal(observation)
}

func (p *workflowLibPlugin) ValidateObservation(ctx context.Context, oc ocr3types.OutcomeContext, _ types.Query, ao types.AttributedObservation) error {
	observation := &pb.Observation{}
	if err := proto.Unmarshal(ao.Observation, observation); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
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
		if err := ctx.Err(); err != nil {
			return err
		}

		if _, ok := priorOutcome.ObservedDonTimes[id]; !ok {
			if requestedSeqNumber != 0 {
				err = errors.Join(err, newInvalidRequestError(requestedSeqNumber, id, 0))
			}
			continue
		}
		seqNum := len(priorOutcome.ObservedDonTimes[id].Timestamps)
		if requestedSeqNumber > int64(seqNum) {
			err = errors.Join(err, newInvalidRequestError(requestedSeqNumber, id, seqNum))
		}
	}

	return err
}

func (p *workflowLibPlugin) ObservationQuorum(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (quorumReached bool, err error) {
	return quorumhelper.ObservationCountReachesObservationQuorum(quorumhelper.QuorumTwoFPlusOne, p.config.N, p.config.F, aos), nil
}

func (p *workflowLibPlugin) Outcome(ctx context.Context, outctx ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	NumFinishedRequests := map[string]int64{} // counts how many nodes reported where a new DON timestamp might be needed
	finishedNodes := map[string]int64{}       // counts number of nodes finished with the workflow for executionID
	var times []int64

	for _, ao := range aos {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			return nil, err
		}

		prevOutcome := &pb.Outcome{}
		if err := proto.Unmarshal(outctx.PreviousOutcome, prevOutcome); err != nil {
			return nil, err
		}

		for id, currSeqNum := range observation.Requests {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
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

	if err := ctx.Err(); err != nil {
		return nil, err
	}

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
	if outcome.RemovedExecutionIDs == nil {
		outcome.RemovedExecutionIDs = make(map[string]bool)
	}

	// Compare with prior outcome to ensure DON time never goes backward.
	if donTime < outcome.Timestamp+p.minTimeIncrease {
		p.lggr.Infow("DON Time incremented by minimum time increase to ensure time progression", "minTimeIncrease", p.minTimeIncrease)
		donTime = outcome.Timestamp + p.minTimeIncrease
	}

	p.lggr.Infow("New DON Time", "donTime", donTime)
	outcome.Timestamp = donTime

	for id, numRequests := range NumFinishedRequests {
		fmt.Println("NumRequests: ", id, numRequests)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if numFinished >= int64(p.config.F) {
			if _, ok := outcome.FinishedExecutionRemovalTimes[id]; ok {
				continue
			}
			outcome.FinishedExecutionRemovalTimes[id] = donTime + int64(defaultRequestExpiry)
		}
	}

	for id, removeAt := range outcome.FinishedExecutionRemovalTimes {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if removeAt <= donTime {
			outcome.RemovedExecutionIDs[id] = true
			delete(outcome.FinishedExecutionRemovalTimes, id)
		}
	}

	return proto.Marshal(outcome)
}

func (p *workflowLibPlugin) Reports(ctx context.Context, seqNr uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportPlus[struct{}], error) {
	return nil, nil
}

func (p *workflowLibPlugin) ShouldAcceptAttestedReport(ctx context.Context, seqNr uint64, reportWithInfo ocr3types.ReportWithInfo[struct{}]) (bool, error) {
	return true, nil
}

func (p *workflowLibPlugin) ShouldTransmitAcceptedReport(ctx context.Context, seqNr uint64, reportWithInfo ocr3types.ReportWithInfo[struct{}]) (bool, error) {
	return true, nil
}

func (p *workflowLibPlugin) Close() error {
	return nil
}

func (p *workflowLibPlugin) Transmit(ctx context.Context, _ types.ConfigDigest, _ uint64, r ocr3types.ReportWithInfo[struct{}], _ []types.AttributedOnchainSignature) error {
	outcome := &pb.Outcome{}
	if err := proto.Unmarshal(r.Report, outcome); err != nil {
		return err
	}

	for id, observedDonTimes := range outcome.ObservedDonTimes {
		p.store.SetDonTimes(id, observedDonTimes.Timestamps)
	}
	p.store.SetLastObservedDonTime(outcome.Timestamp)

	requests, err := p.store.Requests.FirstN(p.batchSize)
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
			p.store.Requests.Evict(id) // Make space for next request before delivering
			request.SendResponse(ctx, DonTimeResponse{
				WorkflowExecutionID: id,
				seqNum:              request.SeqNum,
				timestamp:           donTime,
				Err:                 nil,
			})
		}
	}

	for id := range outcome.RemovedExecutionIDs {
		p.store.deleteExecutionID(id)
	}

	return nil
}

func (p *workflowLibPlugin) FromAccount(ctx context.Context) (types.Account, error) {
	return "", nil
}
