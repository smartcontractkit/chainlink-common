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

func (p *Plugin) Observation(_ context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	previousOutcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, previousOutcome); err != nil {
		p.lggr.Errorf("failed to unmarshal previous outcome in Observation phase")
	}

	// Collect up to batchSize unexpired requests
	requests := map[string]int64{} // Maps executionID --> seqNum
	for batchOffset := 0; batchOffset < p.store.requests.Len() && len(requests) < p.batchSize; {
		batch, err := p.store.requests.RangeN(batchOffset, p.batchSize)
		if err != nil {
			p.lggr.Errorf("failed to get request batch at offset %d: %v", batchOffset, err)
			batchOffset += p.batchSize
			continue
		}
		if len(batch) == 0 {
			break
		}

		timeoutCheck := time.Now()
		for _, req := range batch {
			if req.ExpiryTime().Before(timeoutCheck) {
				// Request has been sitting in queue too long
				p.store.requests.Evict(req.WorkflowExecutionID)
				req.SendTimeout(nil)
				continue
			}

			// Validate request sequence number
			numObservedDonTimes := 0
			times, ok := previousOutcome.ObservedDonTimes[req.WorkflowExecutionID]
			if ok {
				// We have seen this workflow before so check against the sequence
				numObservedDonTimes = len(times.Timestamps)
			}

			if req.SeqNum > numObservedDonTimes {
				p.store.requests.Evict(req.WorkflowExecutionID)
				req.SendResponse(nil,
					Response{
						WorkflowExecutionID: req.WorkflowExecutionID,
						SeqNum:              req.SeqNum,
						Timestamp:           0,
						Err: fmt.Errorf("requested seqNum %d for executionID %s is greater than the number of observed don times %d",
							req.SeqNum, req.WorkflowExecutionID, numObservedDonTimes),
					})
				continue
			}

			requests[req.WorkflowExecutionID] = int64(req.SeqNum)
			batchOffset++
		}
	}

	observation := &pb.Observation{
		Timestamp: time.Now().UTC().UnixMilli(),
		Requests:  requests,
	}

	return proto.Marshal(observation)
}

func (p *Plugin) ValidateObservation(_ context.Context, oc ocr3types.OutcomeContext, _ types.Query, ao types.AttributedObservation) error {
	return nil
}

func (p *Plugin) ObservationQuorum(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (quorumReached bool, err error) {
	return quorumhelper.ObservationCountReachesObservationQuorum(quorumhelper.QuorumTwoFPlusOne, p.config.N, p.config.F, aos), nil
}

func (p *Plugin) Outcome(_ context.Context, outctx ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	observationCounts := map[string]int64{} // counts how many nodes reported where a new DON timestamp might be needed
	var times []int64

	prevOutcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, prevOutcome); err != nil {
		p.lggr.Errorf("failed to unmarshal previous outcome in Outcome phase")
	}
	if prevOutcome.ObservedDonTimes == nil {
		prevOutcome.ObservedDonTimes = make(map[string]*pb.ObservedDonTimes)
	}

	for _, ao := range aos {
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			p.lggr.Errorf("failed to unmarshal observation in Outcome phase")
			continue
		}

		for id, requestSeqNum := range observation.Requests {
			if _, ok := prevOutcome.ObservedDonTimes[id]; !ok {
				prevOutcome.ObservedDonTimes[id] = &pb.ObservedDonTimes{}
			}
			// We only count requests for the next sequence number and ignore all other ones.
			currSeqNum := int64(len(prevOutcome.ObservedDonTimes[id].Timestamps))
			if requestSeqNum == currSeqNum {
				observationCounts[id]++
			} else if requestSeqNum > currSeqNum {
				// This should never happen since we don't include out of sequence requests in the Observation phase
				p.lggr.Errorf("request seqNum %d for executionID %s is greater than the number of observed don times %d",
					requestSeqNum, id, currSeqNum)
			}
		}

		times = append(times, observation.Timestamp)
	}

	p.lggr.Debugw("Observed Node Timestamps", "timestamps", times)
	slices.Sort(times)
	donTime := times[len(times)/2]

	outcome := prevOutcome

	// Compare with prior outcome to ensure DON time never goes backward.
	if donTime < outcome.Timestamp+p.minTimeIncrease {
		p.lggr.Infow("DON Time incremented by minimum time increase to ensure time progression", "minTimeIncrease", p.minTimeIncrease)
		donTime = outcome.Timestamp + p.minTimeIncrease
	}

	p.lggr.Infow("New DON Time", "donTime", donTime)
	outcome.Timestamp = donTime

	for id, numRequests := range observationCounts {
		if numRequests > int64(p.config.F) {
			observedDonTimes, ok := outcome.ObservedDonTimes[id]
			if !ok {
				observedDonTimes = &pb.ObservedDonTimes{}
			}
			observedDonTimes.Timestamps = append(observedDonTimes.Timestamps, donTime)
			outcome.ObservedDonTimes[id] = observedDonTimes
		}
	}

	// Remove expired workflow executions
	for id, observedTimes := range outcome.ObservedDonTimes {
		if observedTimes != nil && len(observedTimes.Timestamps) > 0 {
			if donTime >= observedTimes.Timestamps[0]+p.offChainConfig.ExecutionRemovalTime.AsDuration().Milliseconds() {
				delete(outcome.ObservedDonTimes, id)
				p.store.deleteExecutionID(id)
			}
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
