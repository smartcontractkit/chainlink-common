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

// NewPlugin creates a consensus reporting plugin for shard orchestration
func NewPlugin(store *Store, config ocr3types.ReportingPluginConfig, lggr logger.Logger, cfg *ConsensusConfig) (*Plugin, error) {
	if cfg == nil {
		cfg = &ConsensusConfig{
			MinShardCount: 1,
			MaxShardCount: 100,
			BatchSize:     1000,
			TimeToSync:    5 * time.Minute,
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
	if cfg.TimeToSync <= 0 {
		cfg.TimeToSync = 5 * time.Minute
	}

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

func (p *Plugin) Outcome(_ context.Context, outctx ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	// Load prior state
	prior := &pb.Outcome{}
	if outctx.PreviousOutcome == nil {
		prior.Routes = map[string]*pb.WorkflowRoute{}
		prior.State = &pb.RoutingState{Id: outctx.SeqNr, State: &pb.RoutingState_RoutableShards{RoutableShards: p.minShardCount}}
	} else if err := proto.Unmarshal(outctx.PreviousOutcome, prior); err != nil {
		return nil, err
	}

	currentShardHealth := make(map[uint32]int)
	totalObservations := len(aos)
	allWorkflows := make([]string, 0)
	nows := make([]time.Time, 0)

	// Collect shard health observations and workflows
	for _, ao := range aos {
		observation := &pb.Observation{}
		if err := proto.Unmarshal(ao.Observation, observation); err != nil {
			p.lggr.Warnf("failed to unmarshal observation: %v", err)
			continue
		}

		for shardID, healthy := range observation.ShardHealthStatus {
			if healthy {
				currentShardHealth[shardID]++
			}
		}

		// Collect workflow IDs
		allWorkflows = append(allWorkflows, observation.WorkflowIds...)

		// Collect timestamps
		if observation.Now != nil {
			nows = append(nows, observation.Now.AsTime())
		}
	}

	// Calculate median time
	now := time.Now()
	if len(nows) > 0 {
		slices.SortFunc(nows, time.Time.Compare)
		now = nows[len(nows)/2]
	}

	// Deduplicate workflows
	workflowMap := make(map[string]bool)
	for _, wf := range allWorkflows {
		workflowMap[wf] = true
	}
	allWorkflows = make([]string, 0, len(workflowMap))
	for wf := range workflowMap {
		allWorkflows = append(allWorkflows, wf)
	}
	slices.Sort(allWorkflows) // Ensure deterministic order

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
