package shardorchestrator

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/shardorchestrator/pb"
)

// ShardAllocator is an interface for determining which shard should handle a workflow
// This is implemented by ring.Store to avoid import cycles
type ShardAllocator interface {
	GetShardForWorkflow(ctx context.Context, workflowID string) (uint32, error)
}

// Server implements the gRPC ShardOrchestratorService
// This runs on shard zero and serves requests from other shards
type Server struct {
	pb.UnimplementedShardOrchestratorServiceServer
	store          *Store
	shardAllocator ShardAllocator
	logger         logger.Logger
}

func NewServer(store *Store, shardAllocator ShardAllocator, lggr logger.Logger) *Server {
	return &Server{
		store:          store,
		shardAllocator: shardAllocator,
		logger:         logger.Named(lggr, "ShardOrchestratorServer"),
	}
}

// RegisterWithGRPCServer registers this service with a gRPC server
func (s *Server) RegisterWithGRPCServer(grpcServer *grpc.Server) {
	pb.RegisterShardOrchestratorServiceServer(grpcServer, s)
	s.logger.Info("Registered ShardOrchestrator gRPC service")
}

// GetWorkflowShardMapping handles batch requests for workflow-to-shard mappings
// This is called by other shards to determine where to route workflow executions
func (s *Server) GetWorkflowShardMapping(ctx context.Context, req *pb.GetWorkflowShardMappingRequest) (*pb.GetWorkflowShardMappingResponse, error) {
	s.logger.Debugw("GetWorkflowShardMapping called", "workflowCount", len(req.WorkflowIds))

	if len(req.WorkflowIds) == 0 {
		return nil, fmt.Errorf("workflow_ids is required and must not be empty")
	}

	// Retrieve batch from store
	mappings, version, err := s.store.GetWorkflowMappingsBatch(ctx, req.WorkflowIds)
	if err != nil {
		s.logger.Errorw("Failed to get workflow mappings", "error", err)
		return nil, fmt.Errorf("failed to get workflow mappings: %w", err)
	}

	// Build simple mappings map (workflow_id -> shard_id)
	simpleMappings := make(map[string]uint32, len(mappings))
	// Build detailed mapping states
	mappingStates := make(map[string]*pb.WorkflowMappingState, len(mappings))

	for workflowID, mapping := range mappings {
		// Simple mapping: just the current shard
		simpleMappings[workflowID] = mapping.NewShardID

		// Detailed state: includes transition information
		mappingStates[workflowID] = &pb.WorkflowMappingState{
			OldShardId:   mapping.OldShardID,
			NewShardId:   mapping.NewShardID,
			InTransition: mapping.TransitionState.InTransition(),
		}
	}

	return &pb.GetWorkflowShardMappingResponse{
		Mappings:       simpleMappings,
		MappingStates:  mappingStates,
		MappingVersion: version,
	}, nil
}

// ReportWorkflowTriggerRegistration handles shard registration reports
// Shards call this to inform shard zero about which workflows they have loaded
func (s *Server) ReportWorkflowTriggerRegistration(ctx context.Context, req *pb.ReportWorkflowTriggerRegistrationRequest) (*pb.ReportWorkflowTriggerRegistrationResponse, error) {
	s.logger.Debugw("ReportWorkflowTriggerRegistration called",
		"shardID", req.SourceShardId,
		"workflowCount", len(req.RegisteredWorkflows),
		"totalActive", req.TotalActiveWorkflows,
	)

	// Extract workflow IDs from the map
	workflowIDs := make([]string, 0, len(req.RegisteredWorkflows))
	for workflowID := range req.RegisteredWorkflows {
		workflowIDs = append(workflowIDs, workflowID)
	}

	err := s.store.ReportShardRegistration(ctx, req.SourceShardId, workflowIDs)
	if err != nil {
		s.logger.Errorw("Failed to update shard registrations",
			"shardID", req.SourceShardId,
			"error", err,
		)
		return &pb.ReportWorkflowTriggerRegistrationResponse{
			Success: false,
		}, nil
	}

	s.logger.Infow("Successfully registered workflows",
		"shardID", req.SourceShardId,
		"workflowCount", len(workflowIDs),
	)

	return &pb.ReportWorkflowTriggerRegistrationResponse{
		Success: true,
	}, nil
}

// ReportWorkflows handles workflow discovery from shard 0
// Shard 0 calls this to inform the orchestrator about workflows seen on-chain
func (s *Server) ReportWorkflows(ctx context.Context, req *pb.ReportWorkflowsRequest) (*pb.ReportWorkflowsResponse, error) {
	s.logger.Debugw("ReportWorkflows called", "workflowCount", len(req.WorkflowIds))

	if len(req.WorkflowIds) == 0 {
		return nil, fmt.Errorf("workflow_ids is required and must not be empty")
	}

	// Process each workflow ID to determine shard assignment
	successCount := 0
	for _, workflowID := range req.WorkflowIds {
		existingMapping, err := s.store.GetWorkflowMapping(ctx, workflowID)
		if err == nil && !existingMapping.TransitionState.InTransition() {
			// Workflow already has a stable mapping; skip reassignment
			s.logger.Debugw("Workflow already has stable shard mapping; skipping",
				"workflowID", workflowID,
				"shardID", existingMapping.NewShardID,
			)
			successCount++
			continue
		}

		// Determine shard assignment using the shard allocator
		shardID, err := s.shardAllocator.GetShardForWorkflow(ctx, workflowID)
		if err != nil {
			s.logger.Warnw("Failed to get shard assignment for workflow",
				"workflowID", workflowID,
				"error", err,
			)
			continue
		}

		s.logger.Debugw("Assigned workflow to shard",
			"workflowID", workflowID,
			"shardID", shardID,
		)
		successCount++
	}

	if successCount == 0 {
		s.logger.Errorw("Failed to assign any workflows to shards",
			"totalWorkflows", len(req.WorkflowIds),
		)
		return &pb.ReportWorkflowsResponse{
			Success: false,
		}, nil
	}

	s.logger.Infow("Successfully processed workflow assignments",
		"totalWorkflows", len(req.WorkflowIds),
		"successCount", successCount,
	)

	return &pb.ReportWorkflowsResponse{
		Success: true,
	}, nil
}
