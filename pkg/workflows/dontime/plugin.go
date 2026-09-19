package dontime

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"go.opentelemetry.io/otel/metric"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/smartcontractkit/libocr/quorumhelper"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
)

type pluginMetrics struct {
	donTime                  metric.Int64Gauge
	donTimeEntries           metric.Int64Gauge
	outcomeSize              metric.Int64Gauge
	observationBatchOverflow metric.Int64Gauge
	outcomeBatchOverflow     metric.Int64Gauge
}

func newPluginMetrics() (pluginMetrics, error) {
	meter := beholder.GetMeter()

	donTime, err := meter.Int64Gauge("platform_dontime_outcome_don_time_ms",
		metric.WithDescription("DON consensus timestamp included in the latest outcome, in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return pluginMetrics{}, fmt.Errorf("failed to create don_time gauge: %w", err)
	}

	donTimeEntries, err := meter.Int64Gauge("platform_dontime_outcome_entries",
		metric.WithDescription("Number of workflow execution entries tracked in the latest outcome"),
		metric.WithUnit("{entry}"),
	)
	if err != nil {
		return pluginMetrics{}, fmt.Errorf("failed to create don_time_entries gauge: %w", err)
	}

	outcomeSize, err := meter.Int64Gauge("platform_dontime_outcome_size_bytes",
		metric.WithDescription("Serialised size of the latest outcome in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return pluginMetrics{}, fmt.Errorf("failed to create outcome_size gauge: %w", err)
	}

	observationBatchOverflow, err := meter.Int64Gauge("platform_dontime_observation_batch_overflow",
		metric.WithDescription("Number of pending requests excluded from the observation due to batch size limit"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return pluginMetrics{}, fmt.Errorf("failed to create observation_batch_overflow gauge: %w", err)
	}

	outcomeBatchOverflow, err := meter.Int64Gauge("platform_dontime_outcome_batch_overflow",
		metric.WithDescription("Number of workflow execution entries removed from the outcome due to batch size limit"),
		metric.WithUnit("{entry}"),
	)
	if err != nil {
		return pluginMetrics{}, fmt.Errorf("failed to create outcome_batch_overflow gauge: %w", err)
	}

	return pluginMetrics{
		donTime:                  donTime,
		donTimeEntries:           donTimeEntries,
		outcomeSize:              outcomeSize,
		observationBatchOverflow: observationBatchOverflow,
		outcomeBatchOverflow:     outcomeBatchOverflow,
	}, nil
}

type Plugin struct {
	store          *Store
	config         ocr3types.ReportingPluginConfig
	offChainConfig *pb.Config
	lggr           logger.Logger

	batchSize       int
	minTimeIncrease int64

	metrics pluginMetrics

	sequencedTSEnabled limits.RangeLimiter[config.Timestamp]
}

var _ ocr3types.ReportingPlugin[[]byte] = (*Plugin)(nil)

func NewPlugin(store *Store, config ocr3types.ReportingPluginConfig, offchainCfg *pb.Config, lggr logger.Logger) (*Plugin, error) {
	if offchainCfg.MaxBatchSize == 0 {
		return nil, errors.New("batch size cannot be 0")
	}
	if offchainCfg.MinTimeIncrease <= 0 {
		return nil, errors.New("minimum time increase must be positive")
	}
	if offchainCfg.ExecutionRemovalTime.AsDuration() <= 0 {
		return nil, errors.New("execution removal time must be positive")
	}

	metrics, err := newPluginMetrics()
	if err != nil {
		return nil, err
	}

	return &Plugin{
		store:              store,
		config:             config,
		offChainConfig:     offchainCfg,
		lggr:               logger.Named(lggr, "DONTimePlugin"),
		batchSize:          int(offchainCfg.MaxBatchSize),
		minTimeIncrease:    offchainCfg.MinTimeIncrease / int64(time.Millisecond),
		metrics:            metrics,
		sequencedTSEnabled: limits.NewRangeLimiter(cresettings.Default.DonTimeSequencedTimestampsEnabled.DefaultValue),
	}, nil
}

func (p *Plugin) setSequencedTSEnabled(enabledRange limits.RangeLimiter[config.Timestamp]) {
	p.sequencedTSEnabled = enabledRange
}

func (p *Plugin) Query(_ context.Context, _ ocr3types.OutcomeContext) (types.Query, error) {
	return nil, nil
}

func sortedRequests(requests map[string]*Request) []*Request {
	if len(requests) == 0 {
		return nil
	}

	ids := slices.Sorted(maps.Keys(requests))

	sorted := make([]*Request, 0, len(ids))
	for _, id := range ids {
		sorted = append(sorted, requests[id])
	}
	return sorted
}

func (p *Plugin) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	sortedRequests := sortedRequests(p.store.GetRequests())
	requests := map[string]int64{} // Maps executionID --> seqNum
	for _, req := range sortedRequests {
		requests[req.WorkflowExecutionID] = int64(req.SeqNum)
		if len(requests) >= p.batchSize {
			break
		}
	}

	overflowCount := len(sortedRequests) - len(requests)
	p.lggr.Debugw("Observation batch processed",
		"inputRequests", len(sortedRequests),
		"batchSize", p.batchSize,
		"includedRequests", len(requests),
		"overflowRequests", overflowCount,
	)
	if overflowCount > 0 {
		p.lggr.Warnw("Observation batch overflow", "overflowRequests", overflowCount)
	}
	p.metrics.observationBatchOverflow.Record(ctx, int64(overflowCount))

	observation := &pb.Observation{
		Timestamp:            time.Now().UTC().UnixMilli(),
		Requests:             requests,
		LimitByBatchSizeFlag: true,
	}

	return proto.MarshalOptions{Deterministic: true}.Marshal(observation)
}

func (p *Plugin) ValidateObservation(_ context.Context, oc ocr3types.OutcomeContext, _ types.Query, ao types.AttributedObservation) error {
	return nil
}

func (p *Plugin) ObservationQuorum(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (quorumReached bool, err error) {
	return quorumhelper.ObservationCountReachesObservationQuorum(quorumhelper.QuorumTwoFPlusOne, p.config.N, p.config.F, aos), nil
}

func (p *Plugin) Outcome(ctx context.Context, outctx ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	type timestampNodePair struct {
		Timestamp        int64
		NodeID           int
		OffsetFromMedian int64
	}
	var timestampNodePairs []timestampNodePair
	for idx, ao := range aos {
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			p.lggr.Errorf("failed to unmarshal observation in Outcome phase")
			continue
		}

		timestampNodePairs = append(timestampNodePairs, timestampNodePair{Timestamp: observation.Timestamp, NodeID: idx})
	}

	if len(timestampNodePairs) == 0 {
		return nil, errors.New("no observation contains a valid timestamp")
	}

	slices.SortFunc(timestampNodePairs, func(a, b timestampNodePair) int {
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})
	donTime := timestampNodePairs[len(timestampNodePairs)/2].Timestamp
	for i := range timestampNodePairs {
		timestampNodePairs[i].OffsetFromMedian = timestampNodePairs[i].Timestamp - donTime
	}
	p.lggr.Debugw("Observed Node Timestamps",
		"timestampNodePairs", timestampNodePairs,
		"median", donTime,
		"collectedDataPoints", len(timestampNodePairs),
		"minOffsetFromMedian", timestampNodePairs[0].OffsetFromMedian,
		"maxOffsetFromMedian", timestampNodePairs[len(timestampNodePairs)-1].OffsetFromMedian,
	)

	prevOutcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, prevOutcome); err != nil {
		p.lggr.Errorf("failed to unmarshal previous outcome in Outcome phase")
	}
	if prevOutcome.ObservedDonTimes == nil {
		prevOutcome.ObservedDonTimes = make(map[string]*pb.ObservedDonTimes)
	}

	// Compare with prior outcome to ensure DON time never goes backward.
	if donTime < prevOutcome.Timestamp+p.minTimeIncrease {
		p.lggr.Infow("DON Time incremented by minimum time increase to ensure time progression", "minTimeIncrease", p.minTimeIncrease)
		donTime = prevOutcome.Timestamp + p.minTimeIncrease
	}

	p.lggr.Infow("New DON Time", "donTime", donTime)

	var outcome *pb.Outcome
	if err := p.sequencedTSEnabled.Check(ctx, config.NewTimestamp(time.UnixMilli(donTime))); err != nil {
		if !errors.Is(err, limits.ErrorBoundLimited[config.Timestamp]{}) {
			p.lggr.Warnw("Failed to check for sequenced timestamp feature flag", "err", err)
		}
		outcome = p.unsequencedOutcome(aos, prevOutcome, donTime)
	} else {
		outcome = p.sequencedOutcome(aos, prevOutcome, donTime)
	}

	var outcomeBatchOverflowCount int64
	if len(outcome.ObservedDonTimes) > p.batchSize {
		ids := slices.Sorted(maps.Keys(outcome.ObservedDonTimes))
		outcomeBatchOverflowCount = int64(len(ids) - p.batchSize)
		for _, id := range ids[p.batchSize:] {
			delete(outcome.ObservedDonTimes, id)
		}
		p.lggr.Warnw("Trimmed outcome observed don times to batch size",
			"batchSize", p.batchSize,
			"removedEntries", outcomeBatchOverflowCount,
		)
	}

	outcomeBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(outcome)
	p.lggr.Infow("Outcome computed",
		"observedDonTimesEntries", len(outcome.ObservedDonTimes),
		"outcomeSizeBytes", len(outcomeBytes),
	)
	p.metrics.donTime.Record(ctx, outcome.Timestamp)
	p.metrics.donTimeEntries.Record(ctx, int64(len(outcome.ObservedDonTimes)))
	p.metrics.outcomeBatchOverflow.Record(ctx, outcomeBatchOverflowCount)
	p.metrics.outcomeSize.Record(ctx, int64(len(outcomeBytes)))
	return outcomeBytes, err
}

// unsequencedOutcome executes the original outcome logic to produce an unsequenced slice of [pb.ObservedDonTimes.Timestamps].
func (p *Plugin) unsequencedOutcome(aos []types.AttributedObservation, prevOutcome *pb.Outcome, donTime int64) *pb.Outcome {
	// req_id->count - how many nodes reported where a new DON timestamp might be needed
	observationCounts := map[string]int64{}
	for _, ao := range aos {
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			p.lggr.Errorf("failed to unmarshal observation in Outcome phase")
			continue
		}

		for id, requestSeqNum := range observation.Requests {
			var currSeqNum int64
			if times, ok := prevOutcome.ObservedDonTimes[id]; ok {
				currSeqNum = int64(len(times.Timestamps))
			}
			// We only count requests for the next sequence number and ignore all other ones.
			if requestSeqNum == currSeqNum {
				observationCounts[id]++
			} else if requestSeqNum > currSeqNum {
				// This should never happen since we don't include out of sequence requests in the Observation phase
				p.lggr.Errorf("request seqNum %d for executionID %s is greater than the number of observed don times %d",
					requestSeqNum, id, currSeqNum)
			}
		}
	}

	outcome := prevOutcome
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

	// Remove expired and empty workflow executions
	for id, observedTimes := range outcome.ObservedDonTimes {
		if observedTimes == nil || len(observedTimes.Timestamps) == 0 {
			delete(outcome.ObservedDonTimes, id)
			p.store.deleteExecutionID(id)
			continue
		}
		if donTime >= observedTimes.Timestamps[0]+p.offChainConfig.ExecutionRemovalTime.AsDuration().Milliseconds() {
			delete(outcome.ObservedDonTimes, id)
			p.store.deleteExecutionID(id)
		}
	}
	return outcome
}

// sequencedOutcome executed the updated outcome logic to produce a sequenced map of [pb.ObservedDonTimes.TimestampsBySequence].
func (p *Plugin) sequencedOutcome(aos []types.AttributedObservation, prevOutcome *pb.Outcome, donTime int64) *pb.Outcome {
	type reqSeq struct {
		reqID  string
		seqNum int64
	}
	// [req_id+seq_num]->count - how many nodes reported where a new DON timestamp might be needed
	observationCounts := map[reqSeq]int64{}

	// At the transition point, we need to convert from the old slice format to maps
	for _, observedTimes := range prevOutcome.ObservedDonTimes {
		if len(observedTimes.Timestamps) > 0 {
			for seqNum, ts := range observedTimes.Timestamps {
				observedTimes.TimestampsBySequence[int64(seqNum)] = ts
			}
			observedTimes.Timestamps = nil
		}
	}

	for _, ao := range aos {
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			p.lggr.Errorf("failed to unmarshal observation in Outcome phase")
			continue
		}

		for id, requestSeqNum := range observation.Requests {
			// We only count requests for future sequence numbers and ignore all other ones.
			if times, ok := prevOutcome.ObservedDonTimes[id]; ok {
				if requestSeqNum <= times.MaxSeqNum() {
					continue
				}
			}
			observationCounts[reqSeq{id, requestSeqNum}]++
		}
	}

	outcome := prevOutcome
	outcome.Timestamp = donTime

	for key, numRequests := range observationCounts {
		if numRequests > int64(p.config.F) {
			observedDonTimes, ok := outcome.ObservedDonTimes[key.reqID]
			if !ok {
				observedDonTimes = &pb.ObservedDonTimes{TimestampsBySequence: make(map[int64]int64)}
			}
			observedDonTimes.TimestampsBySequence[key.seqNum] = donTime
			outcome.ObservedDonTimes[key.reqID] = observedDonTimes
		}
	}

	// Remove expired and empty workflow executions
	for id, observedTimes := range outcome.ObservedDonTimes {
		if observedTimes == nil || len(observedTimes.TimestampsBySequence) == 0 {
			delete(outcome.ObservedDonTimes, id)
			p.store.deleteExecutionID(id)
			continue
		}
		if donTime >= observedTimes.EarliestTS()+p.offChainConfig.ExecutionRemovalTime.AsDuration().Milliseconds() {
			delete(outcome.ObservedDonTimes, id)
			p.store.deleteExecutionID(id)
		}
	}
	return outcome
}

func (p *Plugin) Reports(_ context.Context, _ uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportPlus[[]byte], error) {
	allOraclesTransmitNow := &ocr3types.TransmissionSchedule{
		Transmitters:       make([]commontypes.OracleID, p.config.N),
		TransmissionDelays: make([]time.Duration, p.config.N),
	}

	for i := 0; i < p.config.N; i++ {
		allOraclesTransmitNow.Transmitters[i] = commontypes.OracleID(i)
	}

	info, err := structpb.NewStruct(map[string]any{
		"keyBundleName": "evm",
	})
	if err != nil {
		return nil, err
	}
	infoBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(info)
	if err != nil {
		return nil, err
	}

	return []ocr3types.ReportPlus[[]byte]{
		{
			ReportWithInfo: ocr3types.ReportWithInfo[[]byte]{
				Report: types.Report(outcome),
				Info:   infoBytes,
			},
			TransmissionScheduleOverride: allOraclesTransmitNow,
		},
	}, nil
}

func (p *Plugin) ShouldAcceptAttestedReport(ctx context.Context, seqNr uint64, reportWithInfo ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	return true, nil
}

func (p *Plugin) ShouldTransmitAcceptedReport(ctx context.Context, seqNr uint64, reportWithInfo ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	return true, nil
}

func (p *Plugin) Close() error {
	return nil
}
