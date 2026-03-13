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
	protoOpts := &pb.SimulationOptions{
		SimulateTransaction: opts.SimulateTransaction,
	}
	for _, rule := range opts.ExpectedSimulationFailureErrors {
		protoOpts.ExpectedSimulationFailureErrors = append(protoOpts.ExpectedSimulationFailureErrors, &pb.ExpectedSimulationFailureError{
			ErrorString: rule.ErrorString,
		})
	}
	return protoOpts
}

// SimulationOptionsFromProto converts proto SimulationOptions to Go types.
func SimulationOptionsFromProto(proto *pb.SimulationOptions) *types.SimulationOptions {
	if proto == nil {
		return nil
	}
	opts := &types.SimulationOptions{
		SimulateTransaction: proto.GetSimulateTransaction(),
	}
	for _, rule := range proto.GetExpectedSimulationFailureErrors() {
		opts.ExpectedSimulationFailureErrors = append(opts.ExpectedSimulationFailureErrors, types.ExpectedSimulationFailureError{
			ErrorString: rule.GetErrorString(),
		})
	}
	return opts
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
