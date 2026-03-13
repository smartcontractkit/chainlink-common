package contractwriter_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func TestTxMetaFromProto(t *testing.T) {
	t.Run("with nil meta", func(t *testing.T) {
		meta := contractwriter.TxMetaFromProto(nil)
		require.Nil(t, meta)
	})

	t.Run("with nil workflow id", func(t *testing.T) {
		meta := contractwriter.TxMetaFromProto(&pb.TransactionMeta{})
		require.NotNil(t, meta)
		require.Nil(t, meta.WorkflowExecutionID)
	})

	t.Run("with workflow id", func(t *testing.T) {
		meta := contractwriter.TxMetaFromProto(&pb.TransactionMeta{WorkflowExecutionId: "workflow-id"})
		require.NotNil(t, meta)
		require.Equal(t, "workflow-id", *meta.WorkflowExecutionID)
	})

	t.Run("without gas limit", func(t *testing.T) {
		meta := contractwriter.TxMetaFromProto(&pb.TransactionMeta{})
		require.NotNil(t, meta)
		require.Nil(t, meta.GasLimit)
	})

	t.Run("with gas limit", func(t *testing.T) {
		meta := contractwriter.TxMetaFromProto(&pb.TransactionMeta{GasLimit: pb.NewBigIntFromInt(big.NewInt(10))})
		require.NotNil(t, meta)
		require.Equal(t, big.NewInt(10), meta.GasLimit)
	})
}

func TestTxMetaToProto(t *testing.T) {
	t.Run("with nil meta", func(t *testing.T) {
		proto := contractwriter.TxMetaToProto(nil)
		require.Nil(t, proto)
	})

	t.Run("with empty workflow id", func(t *testing.T) {
		proto := contractwriter.TxMetaToProto(&types.TxMeta{})
		require.NotNil(t, proto)
		require.Empty(t, proto.WorkflowExecutionId)
	})

	t.Run("with workflow id", func(t *testing.T) {
		workflowID := "workflow-id"
		proto := contractwriter.TxMetaToProto(&types.TxMeta{WorkflowExecutionID: &workflowID})
		require.NotNil(t, proto)
		require.Equal(t, workflowID, proto.WorkflowExecutionId)
	})

	t.Run("without gas limit", func(t *testing.T) {
		proto := contractwriter.TxMetaToProto(&types.TxMeta{})
		require.NotNil(t, proto)
		require.Empty(t, proto.GasLimit)
	})

	t.Run("with gas limit", func(t *testing.T) {
		proto := contractwriter.TxMetaToProto(&types.TxMeta{GasLimit: big.NewInt(10)})
		require.NotNil(t, proto)
		require.Equal(t, big.NewInt(10), proto.GasLimit.Int())
	})
}

func TestSimulationOptionsToProto(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := contractwriter.SimulationOptionsToProto(nil)
		require.Nil(t, result)
	})

	t.Run("simulate with no rules", func(t *testing.T) {
		opts := &types.SimulationOptions{SimulateTransaction: true}
		result := contractwriter.SimulationOptionsToProto(opts)
		require.NotNil(t, result)
		require.True(t, result.SimulateTransaction)
		require.Empty(t, result.ExpectedSimulationFailureErrors)
	})

	t.Run("simulate with rules", func(t *testing.T) {
		opts := &types.SimulationOptions{
			SimulateTransaction: true,
			ExpectedSimulationFailureErrors: []types.ExpectedSimulationFailureError{
				{ErrorString: "execution reverted"},
				{ErrorString: "insufficient funds"},
			},
		}
		result := contractwriter.SimulationOptionsToProto(opts)
		require.NotNil(t, result)
		require.True(t, result.SimulateTransaction)
		require.Len(t, result.ExpectedSimulationFailureErrors, 2)
		require.Equal(t, "execution reverted", result.ExpectedSimulationFailureErrors[0].ErrorString)
		require.Equal(t, "insufficient funds", result.ExpectedSimulationFailureErrors[1].ErrorString)
	})

	t.Run("no simulate", func(t *testing.T) {
		opts := &types.SimulationOptions{SimulateTransaction: false}
		result := contractwriter.SimulationOptionsToProto(opts)
		require.NotNil(t, result)
		require.False(t, result.SimulateTransaction)
	})
}

func TestSimulationOptionsFromProto(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := contractwriter.SimulationOptionsFromProto(nil)
		require.Nil(t, result)
	})

	t.Run("simulate with no rules", func(t *testing.T) {
		proto := &pb.SimulationOptions{SimulateTransaction: true}
		result := contractwriter.SimulationOptionsFromProto(proto)
		require.NotNil(t, result)
		require.True(t, result.SimulateTransaction)
		require.Empty(t, result.ExpectedSimulationFailureErrors)
	})

	t.Run("simulate with rules", func(t *testing.T) {
		proto := &pb.SimulationOptions{
			SimulateTransaction: true,
			ExpectedSimulationFailureErrors: []*pb.ExpectedSimulationFailureError{
				{ErrorString: "execution reverted"},
				{ErrorString: "insufficient funds"},
			},
		}
		result := contractwriter.SimulationOptionsFromProto(proto)
		require.NotNil(t, result)
		require.True(t, result.SimulateTransaction)
		require.Len(t, result.ExpectedSimulationFailureErrors, 2)
		require.Equal(t, "execution reverted", result.ExpectedSimulationFailureErrors[0].ErrorString)
		require.Equal(t, "insufficient funds", result.ExpectedSimulationFailureErrors[1].ErrorString)
	})
}
