package internal

import (
	"context"
	"time"

	uuid3 "github.com/google/uuid"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/guregu/null.v4"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

var _ PipelineRunner = (*pipelineRunnerServiceClient)(nil)

type pipelineRunnerServiceClient struct {
	*brokerExt
	grpc pb.PipelineRunnerServiceClient
}

func (p pipelineRunnerServiceClient) ExecuteRun(ctx context.Context, spec types.Spec, vars types.Vars, l logger.Logger) (*types.Run, types.TaskRunResults, error) {
	varsStruct, err := structpb.NewStruct(vars.Vars)
	if err != nil {
		return nil, nil, err
	}

	rr := pb.RunRequest{
		Id:     spec.ID,
		Source: spec.DotDagSource,
		Vars:   varsStruct,
		Options: &pb.Options{
			PersistErroredRuns: false,
			MaxTaskDuration:    durationpb.New(spec.MaxTaskDuration),
			GasLimit:           *spec.GasLimit,
			ForwardingAllowed:  spec.ForwardingAllowed,
		},
		Job: &pb.Job{
			Id:   spec.JobID,
			Name: spec.JobName,
			Type: spec.JobType,
		},
	}

	executeRunResult, err := p.grpc.ExecuteRun(ctx, &rr)
	if err != nil {
		return nil, nil, err
	}

	trrs := make([]types.TaskRunResult, len(executeRunResult.Results))

	for i, trr := range executeRunResult.Results {
		id, err := uuid.FromString(trr.Id)
		if err != nil {
			return nil, nil, err
		}

		trrs[i] = types.TaskRunResult{
			ID:         uuid3.UUID(id),
			Task:       nil,
			TaskRun:    types.TaskRun{},
			Result:     types.Result{},
			Attempts:   0,
			CreatedAt:  time.Time{},
			FinishedAt: null.Time{},
			RunInfo:    types.RunInfo{},
		}
	}

	return nil, trrs, nil
}

var _ pb.PipelineRunnerServiceServer = (*pipelineRunnerServiceServer)(nil)

type pipelineRunnerServiceServer struct {
	pb.UnimplementedPipelineRunnerServiceServer
	*brokerExt

	impl PipelineRunner
}

func (p *pipelineRunnerServiceServer) ExecuteRun(ctx context.Context, rr *pb.RunRequest) (*pb.RunResponse, error) {

	spec := types.Spec{
		ID:                rr.Id,
		DotDagSource:      rr.Source,
		CreatedAt:         time.Now(),
		MaxTaskDuration:   rr.Options.MaxTaskDuration.AsDuration(),
		GasLimit:          &rr.Options.GasLimit,
		ForwardingAllowed: rr.Options.ForwardingAllowed,
		JobID:             rr.Job.Id,
		JobName:           rr.Job.Name,
		JobType:           rr.Job.Type,
	}
	vars := types.Vars{Vars: rr.Vars.AsMap()}

	_, trrs, err := p.impl.ExecuteRun(ctx, spec, vars, nil) // Need logger
	if err != nil {
		return nil, err
	}

	taskResults := make([]*pb.TaskResult, len(trrs))

	for i, trr := range trrs {
		v, err := structpb.NewValue(trr.Result.Value)
		if err != nil {
			return nil, err
		}

		taskResults[i] = &pb.TaskResult{
			Id:    trr.ID.String(),
			Type:  trr.TaskRun.Type,
			Value: v,
			Index: 0, // TODO: figure out what this is,
		}
	}

	return &pb.RunResponse{
		Results: taskResults,
	}, nil

}
