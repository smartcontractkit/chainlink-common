package contractwriter

import (
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// TxMetaToProto converts a TxMeta to it's generated protobuf Go type.
func TxMetaToProto(meta *types.TxMeta) *pb.TransactionMeta {
	if meta == nil {
		return nil
	}

	proto := &pb.TransactionMeta{}

	if meta.WorkflowExecutionID != nil {
		proto.WorkflowExecutionId = *meta.WorkflowExecutionID
	}

	if meta.GasLimit != nil {
		proto.GasLimit = pb.NewBigIntFromInt(meta.GasLimit)
	}

	return proto
}

// SimulationOptionsToProto converts Go SimulationOptions to proto.
func SimulationOptionsToProto(opts *types.SimulationOptions) *pb.SimulationOptions {
	if opts == nil {
		return nil
	}
	return &pb.SimulationOptions{
		SimulateTransaction:            opts.SimulateTransaction,
		ExpectedSimulationFailureErrors: failureRulesToProto(opts.ExpectedSimulationFailureErrors),
	}
}

// SimulationOptionsFromProto converts proto SimulationOptions to Go types.
func SimulationOptionsFromProto(proto *pb.SimulationOptions) *types.SimulationOptions {
	if proto == nil {
		return nil
	}
	return &types.SimulationOptions{
		SimulateTransaction:            proto.GetSimulateTransaction(),
		ExpectedSimulationFailureErrors: failureRulesFromProto(proto.GetExpectedSimulationFailureErrors()),
	}
}

func failureRulesToProto(rules []types.ExpectedSimulationFailureError) []*pb.ExpectedSimulationFailureError {
	if len(rules) == 0 {
		return nil
	}
	pbRules := make([]*pb.ExpectedSimulationFailureError, len(rules))
	for i, rule := range rules {
		pbRules[i] = &pb.ExpectedSimulationFailureError{
			ErrorString: rule.ErrorString,
		}
	}
	return pbRules
}

func failureRulesFromProto(pbRules []*pb.ExpectedSimulationFailureError) []types.ExpectedSimulationFailureError {
	if len(pbRules) == 0 {
		return nil
	}
	rules := make([]types.ExpectedSimulationFailureError, len(pbRules))
	for i, pbRule := range pbRules {
		rules[i] = types.ExpectedSimulationFailureError{
			ErrorString: pbRule.GetErrorString(),
		}
	}
	return rules
}

// TxMetaFromProto converts a TxMeta from it's generated protobuf Go type to our internal Go type.
func TxMetaFromProto(proto *pb.TransactionMeta) *types.TxMeta {
	if proto == nil {
		return nil
	}

	meta := &types.TxMeta{}

	if proto.WorkflowExecutionId != "" {
		meta.WorkflowExecutionID = &proto.WorkflowExecutionId
	}

	if proto.GasLimit != nil {
		meta.GasLimit = proto.GasLimit.Int()
	}

	return meta
}
