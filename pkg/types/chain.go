package types

import (
	"context"
	"math/big"
)

type ChainSet[I any, C ChainService] interface {
	Service

	Chain(ctx context.Context, id I) (C, error)

	ChainStatus(ctx context.Context, id string) (ChainStatus, error)
	ChainStatuses(ctx context.Context, offset, limit int) (chains []ChainStatus, count int, err error)

	NodeStatuses(ctx context.Context, offset, limit int, chainIDs ...string) (nodes []NodeStatus, count int, err error)

	SendTx(ctx context.Context, chainID, from, to string, amount *big.Int, balanceCheck bool) error
}

type ChainService interface {
	Service

	GetChainStatus(ctx context.Context) (ChainStatus, error)
	ListNodeStatuses(ctx context.Context, page_size int32, page_token string) (stats []NodeStatus, next_page_token string, err error)
	// choose different name than SendTx to avoid collison during refactor.
	Transact(ctx context.Context, from, to string, amount *big.Int, balanceCheck bool) error
}

type ChainStatus struct {
	ID      string
	Enabled bool
	Config  string // TOML
}

type NodeStatus struct {
	ChainID string
	Name    string
	Config  string // TOML
	State   string
}
