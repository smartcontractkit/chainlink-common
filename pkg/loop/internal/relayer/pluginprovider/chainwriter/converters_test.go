package chainwriter_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestTxMetaFromProto(t *testing.T) {
	t.Run("with nil meta", func(t *testing.T) {
		meta := chainwriter.TxMetaFromProto(nil)
		require.Nil(t, meta)
	})

	t.Run("with nil workflow id", func(t *testing.T) {
		meta := chainwriter.TxMetaFromProto(&pb.TransactionMeta{})
		require.NotNil(t, meta)
		require.Nil(t, meta.WorkflowExecutionID)
	})

	t.Run("with workflow id", func(t *testing.T) {
		meta := chainwriter.TxMetaFromProto(&pb.TransactionMeta{WorkflowExecutionId: "workflow-id"})
		require.NotNil(t, meta)
		require.Equal(t, "workflow-id", *meta.WorkflowExecutionID)
	})
}

func TestTxMetaToProto(t *testing.T) {
	t.Run("with nil meta", func(t *testing.T) {
		proto := chainwriter.TxMetaToProto(nil)
		require.Nil(t, proto)
	})

	t.Run("with empty workflow id", func(t *testing.T) {
		proto := chainwriter.TxMetaToProto(&types.TxMeta{})
		require.NotNil(t, proto)
		require.Empty(t, proto.WorkflowExecutionId)
	})

	t.Run("with workflow id", func(t *testing.T) {
		workflowID := "workflow-id"
		proto := chainwriter.TxMetaToProto(&types.TxMeta{WorkflowExecutionID: &workflowID})
		require.NotNil(t, proto)
		require.Equal(t, workflowID, proto.WorkflowExecutionId)
	})
}
