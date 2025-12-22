package shardorchestrator

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/shardorchestrator/pb"
	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/smartcontractkit/libocr/quorumhelper"
)

type Plugin struct {
	mu sync.RWMutex

	store  *Store
	config ocr3types.ReportingPluginConfig
	lggr   logger.Logger

	batchSize     int
	minShardCount uint32
	maxShardCount uint32
}

var _ ocr3types.ReportingPlugin[[]byte] = (*Plugin)(nil)

// ConsensusConfig holds the plugin configuration
type ConsensusConfig struct {
	MinShardCount uint32
	MaxShardCount uint32
	BatchSize     int
}

// NewPlugin creates a consensus reporting plugin for shard orchestration
func NewPlugin(store *Store, config ocr3types.ReportingPluginConfig, lggr logger.Logger, cfg *ConsensusConfig) (*Plugin, error) {
	if cfg == nil {
		cfg = &ConsensusConfig{
			MinShardCount: 1,
			MaxShardCount: 10,
			BatchSize:     100,
		}
	}

	if cfg.MaxShardCount == 0 {
		return nil, errors.New("max shard count cannot be 0")
	}
	if cfg.MinShardCount == 0 {
		cfg.MinShardCount = 1
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}

	return &Plugin{
		store:         store,
		config:        config,
		lggr:          logger.Named(lggr, "ShardOrchestratorPlugin"),
		batchSize:     cfg.BatchSize,
		minShardCount: cfg.MinShardCount,
		maxShardCount: cfg.MaxShardCount,
	}, nil
}

func (p *Plugin) Query(_ context.Context, _ ocr3types.OutcomeContext) (types.Query, error) {
	return nil, nil
}

func (p *Plugin) Observation(_ context.Context, outctx ocr3types.OutcomeContext, _ types.Query) (types.Observation, error) {
	shardHealth := p.store.GetShardHealth()

	hashes := []string{}
	for wfID := range p.store.GetAllRoutingState() {
		hashes = append(hashes, wfID)
	}
	slices.Sort(hashes)

	observation := &pb.Observation{
		Status: shardHealth,
		Hashes: hashes,
		Now:    nil,
	}

	return proto.MarshalOptions{Deterministic: true}.Marshal(observation)
}

func (p *Plugin) ValidateObservation(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query, _ types.AttributedObservation) error {
	return nil
}

func (p *Plugin) ObservationQuorum(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (quorumReached bool, err error) {
	return quorumhelper.ObservationCountReachesObservationQuorum(quorumhelper.QuorumTwoFPlusOne, p.config.N, p.config.F, aos), nil
}

func (p *Plugin) Outcome(_ context.Context, outctx ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	prevOutcome := &pb.Outcome{}
	if err := proto.Unmarshal(outctx.PreviousOutcome, prevOutcome); err != nil {
		p.lggr.Warnf("failed to unmarshal previous outcome: %v", err)
		prevOutcome = &pb.Outcome{
			State: &pb.RoutingState{
				Id: 0,
				State: &pb.RoutingState_RoutableShards{
					RoutableShards: p.minShardCount,
				},
			},
			Routes: make(map[string]*pb.WorkflowRoute),
		}
	}

	currentShardHealth := make(map[uint32]int)
	totalObservations := len(aos)

	// Collect shard health observations
	for _, ao := range aos {
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			p.lggr.Warnf("failed to unmarshal observation: %v", err)
			continue
		}

		for shardID, healthy := range observation.Status {
			if healthy {
				currentShardHealth[shardID]++
			}
		}
	}

	// Determine desired shard count based on observations
	healthyShardCount := uint32(0)
	for _, count := range currentShardHealth {
		if count > int(p.config.F) {
			healthyShardCount++
		}
	}

	// Ensure within bounds
	if healthyShardCount < p.minShardCount {
		healthyShardCount = p.minShardCount
	}
	if healthyShardCount > p.maxShardCount {
		healthyShardCount = p.maxShardCount
	}

	// Update routing state
	outcome := &pb.Outcome{
		State: &pb.RoutingState{
			Id: prevOutcome.State.Id + 1,
			State: &pb.RoutingState_RoutableShards{
				RoutableShards: healthyShardCount,
			},
		},
		Routes: make(map[string]*pb.WorkflowRoute),
	}

	p.lggr.Infow("Consensus Outcome", "healthyShards", healthyShardCount, "totalObservations", totalObservations)

	// Copy previous routes
	if prevOutcome.Routes != nil {
		for wfID, route := range prevOutcome.Routes {
			outcome.Routes[wfID] = route
		}
	}

	return proto.MarshalOptions{Deterministic: true}.Marshal(outcome)
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
