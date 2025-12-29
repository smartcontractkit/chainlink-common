package ring

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/ring/pb"
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
	timeToSync    time.Duration
}

var _ ocr3types.ReportingPlugin[[]byte] = (*Plugin)(nil)

// ConsensusConfig holds the plugin configuration
type ConsensusConfig struct {
	MinShardCount uint32
	MaxShardCount uint32
	BatchSize     int
	TimeToSync    time.Duration
}

const (
	DefaultMinShardCount = 1
	DefaultMaxShardCount = 100
	DefaultBatchSize     = 100
	DefaultTimeToSync    = 5 * time.Minute
)

// NewPlugin creates a consensus reporting plugin for shard orchestration
func NewPlugin(store *Store, config ocr3types.ReportingPluginConfig, lggr logger.Logger, cfg *ConsensusConfig) (*Plugin, error) {
	if cfg == nil {
		cfg = &ConsensusConfig{
			MinShardCount: DefaultMinShardCount,
			MaxShardCount: DefaultMaxShardCount,
			BatchSize:     DefaultBatchSize,
			TimeToSync:    DefaultTimeToSync,
		}
	}

	if cfg.MaxShardCount == 0 {
		return nil, errors.New("max shard count cannot be 0")
	}
	if cfg.MinShardCount == 0 {
		lggr.Infow("using default minShardCount", "default", DefaultMinShardCount)
		cfg.MinShardCount = DefaultMinShardCount
	}
	if cfg.BatchSize <= 0 {
		lggr.Infow("using default batchSize", "default", DefaultBatchSize)
		cfg.BatchSize = DefaultBatchSize
	}
	if cfg.TimeToSync <= 0 {
		lggr.Infow("using default timeToSync", "default", DefaultTimeToSync)
		cfg.TimeToSync = DefaultTimeToSync
	}

	lggr.Infow("RingPlugin config",
		"minShardCount", cfg.MinShardCount,
		"maxShardCount", cfg.MaxShardCount,
		"batchSize", cfg.BatchSize,
		"timeToSync", cfg.TimeToSync,
	)

	return &Plugin{
		store:         store,
		config:        config,
		lggr:          logger.Named(lggr, "RingPlugin"),
		batchSize:     cfg.BatchSize,
		minShardCount: cfg.MinShardCount,
		maxShardCount: cfg.MaxShardCount,
		timeToSync:    cfg.TimeToSync,
	}, nil
}

func (p *Plugin) Query(_ context.Context, _ ocr3types.OutcomeContext) (types.Query, error) {
	return nil, nil
}

func (p *Plugin) Observation(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query) (types.Observation, error) {
	shardHealth := p.store.GetShardHealth()

	allWorkflowIDs := make([]string, 0)
	for wfID := range p.store.GetAllRoutingState() {
		allWorkflowIDs = append(allWorkflowIDs, wfID)
	}
	slices.Sort(allWorkflowIDs)

	observation := &pb.Observation{
		ShardHealthStatus: shardHealth,
		WorkflowIds:       allWorkflowIDs,
		Now:               timestamppb.Now(),
	}

	return proto.MarshalOptions{Deterministic: true}.Marshal(observation)
}

func (p *Plugin) ValidateObservation(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query, _ types.AttributedObservation) error {
	return nil
}

func (p *Plugin) ObservationQuorum(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (quorumReached bool, err error) {
	return quorumhelper.ObservationCountReachesObservationQuorum(quorumhelper.QuorumTwoFPlusOne, p.config.N, p.config.F, aos), nil
}

func (p *Plugin) collectShardInfo(aos []types.AttributedObservation) (shardHealth map[uint32]int, workflows []string, timestamps []time.Time) {
	shardHealth = make(map[uint32]int)
	for _, ao := range aos {
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			p.lggr.Warnf("failed to unmarshal observation: %v", err)
			continue
		}

		for shardID, healthy := range observation.ShardHealthStatus {
			if healthy {
				shardHealth[shardID]++
			}
		}

		workflows = append(workflows, observation.WorkflowIds...)

		if observation.Now != nil {
			timestamps = append(timestamps, observation.Now.AsTime())
		}
	}
	return shardHealth, workflows, timestamps
}

func (p *Plugin) Outcome(_ context.Context, outctx ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	// Bootstrap with minimum shards on first round; subsequent rounds build on prior outcome
	prior := &pb.Outcome{}
	if outctx.PreviousOutcome == nil {
		prior.Routes = map[string]*pb.WorkflowRoute{}
		prior.State = &pb.RoutingState{Id: outctx.SeqNr, State: &pb.RoutingState_RoutableShards{RoutableShards: p.minShardCount}}
	} else if err := proto.Unmarshal(outctx.PreviousOutcome, prior); err != nil {
		return nil, err
	}

	totalObservations := len(aos)
	currentShardHealth, allWorkflows, nows := p.collectShardInfo(aos)

	// Need at least F+1 timestamps; fewer means >F faulty nodes and we can't trust this round
	if len(nows) < p.config.F+1 {
		return nil, errors.New("insufficient observation timestamps")
	}
	slices.SortFunc(nows, time.Time.Compare)

	// Use the median timestamp to determine the current time
	now := nows[len(nows)/2]

	allWorkflows = uniqueSorted(allWorkflows)

	// Determine desired shard count based on observations
	healthyShardCount := uint32(0)
	for shardID, count := range currentShardHealth {
		if count > p.config.F {
			healthyShardCount++
			// Update store with healthy shard
			p.store.SetShardHealth(shardID, true)
		}
	}

	// Ensure within bounds
	if healthyShardCount < p.minShardCount {
		healthyShardCount = p.minShardCount
	}
	if healthyShardCount > p.maxShardCount {
		healthyShardCount = p.maxShardCount
	}

	// Calculate next state using state machine
	nextState, err := p.calculateNextState(prior.State, healthyShardCount, now)
	if err != nil {
		return nil, err
	}

	// Use deterministic hashing to assign workflows to shards
	routes := make(map[string]*pb.WorkflowRoute)
	for _, wfID := range allWorkflows {
		assignedShard := p.store.GetShardForWorkflow(wfID)
		routes[wfID] = &pb.WorkflowRoute{
			Shard: assignedShard,
		}
	}

	// Update routing state
	outcome := &pb.Outcome{
		State:  nextState,
		Routes: routes,
	}

	p.lggr.Infow("Consensus Outcome", "healthyShards", healthyShardCount, "totalObservations", totalObservations, "workflowCount", len(routes))

	return proto.MarshalOptions{Deterministic: true}.Marshal(outcome)
}

func (p *Plugin) calculateNextState(priorState *pb.RoutingState, wantShards uint32, now time.Time) (*pb.RoutingState, error) {
	switch ps := priorState.State.(type) {
	case *pb.RoutingState_RoutableShards:
		// If already at desired count, stay in stable state
		if ps.RoutableShards == wantShards {
			return priorState, nil
		}

		// Otherwise, initiate transition
		return &pb.RoutingState{
			Id: priorState.Id + 1,
			State: &pb.RoutingState_Transition{
				Transition: &pb.Transition{
					WantShards:       wantShards,
					LastStableCount:  ps.RoutableShards,
					ChangesSafeAfter: timestamppb.New(now.Add(p.timeToSync)),
				},
			},
		}, nil

	case *pb.RoutingState_Transition:
		// If still in safety period, stay in transition
		if now.Before(ps.Transition.ChangesSafeAfter.AsTime()) {
			return priorState, nil
		}

		// Safety period elapsed, transition to stable state
		return &pb.RoutingState{
			Id: priorState.Id + 1,
			State: &pb.RoutingState_RoutableShards{
				RoutableShards: ps.Transition.WantShards,
			},
		}, nil

	default:
		return nil, errors.New("unknown prior state type")
	}
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
		},
	}, nil
}

func (p *Plugin) ShouldAcceptAttestedReport(_ context.Context, _ uint64, _ ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	return true, nil
}

func (p *Plugin) ShouldTransmitAcceptedReport(_ context.Context, _ uint64, _ ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	return true, nil
}

func (p *Plugin) Close() error {
	return nil
}
