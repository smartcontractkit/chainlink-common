package core

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type PluginCCIPCommit interface {
	// NewCCIPCommitFactory returns a new ReportingPluginFactory.
	// If provider implements GRPCClientConn, it can be forwarded efficiently via proxy.
	NewCCIPCommitFactory(
		ctx context.Context,
		contractReaders map[types.RelayID]types.ContractReader,
	) (types.ReportingPluginFactory, error)
}
