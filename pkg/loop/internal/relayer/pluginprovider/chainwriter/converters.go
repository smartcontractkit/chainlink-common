package chainwriter

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
