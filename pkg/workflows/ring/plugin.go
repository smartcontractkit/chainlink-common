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

	batchSize  int
	timeToSync time.Duration
}

var _ ocr3types.ReportingPlugin[[]byte] = (*Plugin)(nil)

type ConsensusConfig struct {
	BatchSize  int
	TimeToSync time.Duration
}

const (
	DefaultBatchSize  = 100
	DefaultTimeToSync = 5 * time.Minute
)

// NewPlugin creates a consensus reporting plugin for shard orchestration
func NewPlugin(store *Store, config ocr3types.ReportingPluginConfig, lggr logger.Logger, cfg *ConsensusConfig) (*Plugin, error) {
	if cfg == nil {
		cfg = &ConsensusConfig{
			BatchSize:  DefaultBatchSize,
			TimeToSync: DefaultTimeToSync,
		}
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
		"batchSize", cfg.BatchSize,
		"timeToSync", cfg.TimeToSync,
	)

	return &Plugin{
		store:      store,
		config:     config,
		lggr:       logger.Named(lggr, "RingPlugin"),
		batchSize:  cfg.BatchSize,
		timeToSync: cfg.TimeToSync,
	}, nil
}

//coverage:ignore
func (p *Plugin) Query(_ context.Context, _ ocr3types.OutcomeContext) (types.Query, error) {
	return nil, nil
}

func (p *Plugin) Observation(_ context.Context, _ ocr3types.OutcomeContext, _ types.Query) (types.Observation, error) {
	shardHealth := p.store.GetShardHealth()

	allWorkflowIDs := make([]string, 0)
	for wfID := range p.store.GetAllRoutingState() {
		allWorkflowIDs = append(allWorkflowIDs, wfID)
	}

	pendingAllocs := p.store.GetPendingAllocations()
	allWorkflowIDs = append(allWorkflowIDs, pendingAllocs...)

	allWorkflowIDs = uniqueSorted(allWorkflowIDs)

	observation := &pb.Observation{
		ShardHealthStatus: shardHealth,
		WorkflowIds:       allWorkflowIDs,
		Now:               timestamppb.Now(),
	}

	return proto.MarshalOptions{Deterministic: true}.Marshal(observation)
}

//coverage:ignore
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

func (p *Plugin) getHealthyShards(shardHealth map[uint32]int) []uint32 {
	var healthyShards []uint32
	for shardID, votes := range shardHealth {
		if votes > p.config.F {
			healthyShards = append(healthyShards, shardID)
			p.store.SetShardHealth(shardID, true)
		}
	}
	slices.Sort(healthyShards)

	return healthyShards
}

func (p *Plugin) Outcome(_ context.Context, outctx ocr3types.OutcomeContext, _ types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	// Bootstrap with 1 shard on first round; subsequent rounds build on prior outcome
	prior := &pb.Outcome{}
	if outctx.PreviousOutcome == nil {
		prior.Routes = map[string]*pb.WorkflowRoute{}
		prior.State = &pb.RoutingState{Id: outctx.SeqNr, State: &pb.RoutingState_RoutableShards{RoutableShards: 1}}
	} else if err := proto.Unmarshal(outctx.PreviousOutcome, prior); err != nil {
		return nil, err
	}

	currentShardHealth, allWorkflows, nows := p.collectShardInfo(aos)

	// Need at least F+1 timestamps; fewer means >F faulty nodes and we can't trust this round
	if len(nows) < p.config.F+1 {
		return nil, errors.New("insufficient observation timestamps")
	}
	slices.SortFunc(nows, time.Time.Compare)

	// Use the median timestamp to determine the current time
	now := nows[len(nows)/2]

	allWorkflows = uniqueSorted(allWorkflows)

	healthyShards := p.getHealthyShards(currentShardHealth)

	nextState, err := p.calculateNextState(prior.State, uint32(len(healthyShards)), now)
	if err != nil {
		return nil, err
	}

	// Deterministic hashing ensures all nodes agree on workflow-to-shard assignments
	// without coordination, preventing protocol failures from inconsistent routing
	routes := make(map[string]*pb.WorkflowRoute)
	for _, wfID := range allWorkflows {
		assignedShard := getShardForWorkflow(wfID, healthyShards)
		routes[wfID] = &pb.WorkflowRoute{
			Shard: assignedShard,
		}
	}

	outcome := &pb.Outcome{
		State:  nextState,
		Routes: routes,
	}

	p.lggr.Infow("Consensus Outcome", "healthyShards", len(healthyShards), "totalObservations", len(aos), "workflowCount", len(routes))

	return proto.MarshalOptions{Deterministic: true}.Marshal(outcome)
}

func (p *Plugin) calculateNextState(priorState *pb.RoutingState, wantShards uint32, now time.Time) (*pb.RoutingState, error) {
	switch ps := priorState.State.(type) {
	case *pb.RoutingState_RoutableShards:
		// No transition needed; avoid unnecessary workflow redistribution
		if ps.RoutableShards == wantShards {
			return priorState, nil
		}

		// Shard count changed; start transition with safety period for workflow redistribution
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
		// Wait for all nodes to sync before committing to new shard assignments
		if now.Before(ps.Transition.ChangesSafeAfter.AsTime()) {
			return priorState, nil
		}

		// All nodes have synced; commit to new routing configuration
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

//coverage:ignore
func (p *Plugin) ShouldAcceptAttestedReport(_ context.Context, _ uint64, _ ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	return true, nil
}

//coverage:ignore
func (p *Plugin) ShouldTransmitAcceptedReport(_ context.Context, _ uint64, _ ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	return true, nil
}

//coverage:ignore
func (p *Plugin) Close() error {
	return nil
}
