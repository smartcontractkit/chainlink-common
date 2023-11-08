package internal

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
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
		var taskErr error
		if trr.HasError {
			taskErr = errors.New(trr.Error)
		}

		var v any
		switch {
		case trr.Value.MessageIs(&pb.Decimal{}):
			d := &pb.Decimal{}
			err := trr.Value.UnmarshalTo(d)
			if err != nil {
				return nil, err
			}
			dec, err := decimal.NewFromString(d.Decimal)
			if err != nil {
				return nil, err
			}
			v = dec
		case trr.Value.MessageIs(&structpb.Value{}):
			val := &structpb.Value{}
			err := trr.Value.UnmarshalTo(val)
			if err != nil {
				return nil, err
			}
			v = val.AsInterface()
		default:
			return nil, fmt.Errorf("could not unmarshal value: %+v", trr.Value)
		}

		tv := types.TaskValue{
			Value:      v,
			Error:      taskErr,
			IsTerminal: trr.IsTerminal,
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

		var msg proto.Message
		switch tv := trr.Value.(type) {
		case decimal.Decimal:
			msg = &pb.Decimal{Decimal: tv.String()}
		default:
			v, err := structpb.NewValue(trr.Value)
			if err != nil {
				return nil, err
			}
			msg = v
		}

		anyproto, err := anypb.New(msg)
		if err != nil {
			return nil, err
		}

		taskResults[i] = &pb.TaskResult{
			Id:         trr.ID,
			Type:       trr.Type,
			Error:      errs,
			HasError:   hasError,
			IsTerminal: trr.IsTerminal,
			Index:      int32(trr.Index),
			Value:      anyproto,
		}
	}

	return &pb.RunResponse{
		Results: taskResults,
	}, nil
}
