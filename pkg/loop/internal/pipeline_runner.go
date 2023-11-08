package internal

import (
	"context"
	"encoding/json"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

var _ types.PipelineRunnerService = (*pipelineRunnerServiceClient)(nil)

type pipelineRunnerServiceClient struct {
	*brokerExt
	grpc pb.PipelineRunnerServiceClient
}

func newPipelineRunnerClient(cc grpc.ClientConnInterface) *pipelineRunnerServiceClient {
	return &pipelineRunnerServiceClient{grpc: pb.NewPipelineRunnerServiceClient(cc)}
}

func (p pipelineRunnerServiceClient) ExecuteRun(ctx context.Context, spec string, vars types.Vars, options types.Options) (types.TaskResults, error) {
	varsStruct, err := structpb.NewStruct(vars.Vars)
	if err != nil {
		return nil, err
	}

	rr := pb.RunRequest{
		Spec: spec,
		Vars: varsStruct,
		Options: &pb.Options{
			MaxTaskDuration: durationpb.New(options.MaxTaskDuration),
		},
	}

	executeRunResult, err := p.grpc.ExecuteRun(ctx, &rr)
	if err != nil {
		return nil, err
	}

	trs := make([]types.TaskResult, len(executeRunResult.Results))
	for i, trr := range executeRunResult.Results {
		var err error
		if trr.HasError {
			err = errors.New(trr.Error)
		}
		var tv types.TaskValue
		if trr.GetPrimitive() != nil {
			tv = types.TaskValue{
				Value:      trr.GetPrimitive().AsInterface(),
				Error:      err,
				IsTerminal: trr.IsTerminal,
			}
		} else {
			c := trr.GetComplex()
			var v interface{}
			switch c.Type {
			case "decimal.Decimal":
				d := decimal.Zero
				err = json.Unmarshal(c.Value, &d)
				if err != nil {
					return nil, err
				}
				v = d
			}
			tv = types.TaskValue{
				Value:      v,
				Error:      err,
				IsTerminal: trr.IsTerminal,
			}
		}
		trs[i] = types.TaskResult{
			ID:        trr.Id,
			Type:      trr.Type,
			TaskValue: tv,
			Index:     int(trr.Index),
		}
	}

	return trs, nil
}

var _ pb.PipelineRunnerServiceServer = (*pipelineRunnerServiceServer)(nil)

type pipelineRunnerServiceServer struct {
	pb.UnimplementedPipelineRunnerServiceServer
	*brokerExt

	impl types.PipelineRunnerService
}

func (p *pipelineRunnerServiceServer) ExecuteRun(ctx context.Context, rr *pb.RunRequest) (*pb.RunResponse, error) {
	vars := types.Vars{
		Vars: rr.Vars.AsMap(),
	}
	options := types.Options{
		MaxTaskDuration: rr.Options.MaxTaskDuration.AsDuration(),
	}
	trs, err := p.impl.ExecuteRun(ctx, rr.Spec, vars, options)
	if err != nil {
		return nil, err
	}

	taskResults := make([]*pb.TaskResult, len(trs))
	for i, trr := range trs {
		hasError := trr.Error != nil
		errs := ""
		if hasError {
			errs = trr.Error.Error()
		}
		tr := &pb.TaskResult{
			Id:         trr.ID,
			Type:       trr.Type,
			Error:      errs,
			HasError:   hasError,
			IsTerminal: trr.IsTerminal,
			Index:      int32(trr.Index),
		}
		switch trr.Value.(type) {
		case decimal.Decimal:
			b, err := json.Marshal(trr.Value)
			if err != nil {
				return nil, err
			}
			tr.Value = &pb.TaskResult_Complex{
				Complex: &pb.Value{
					Type:  "decimal.Decimal",
					Value: b,
				},
			}
		default:
			v, err := structpb.NewValue(trr.Value)
			if err != nil {
				return nil, err
			}
			tr.Value = &pb.TaskResult_Primitive{
				Primitive: v,
			}
		}

		taskResults[i] = tr
	}

	return &pb.RunResponse{
		Results: taskResults,
	}, nil

}
