package types

import (
	"context"
	"math/big"
)

// ChainSet is a [ChainTransactor] with the ability to get a specific chain by id
type ChainSet[I any, C ChainService] interface {
	Service
	ChainTransactor
	Chain(ctx context.Context, id I) (C, error)
}

// ChainTransactor transacts on a given chains and provides status information
// for chains and nodes
type ChainTransactor interface {
	ChainStatus(ctx context.Context, id string) (ChainStatus, error)
	ChainStatuses(ctx context.Context, offset, limit int) (chains []ChainStatus, count int, err error)

	NodeStatuses(ctx context.Context, offset, limit int, chainIDs ...string) (nodes []NodeStatus, count int, err error)
	SendTx(ctx context.Context, chainID, from, to string, amount *big.Int, balanceCheck bool) error
}

type ChainService interface {
	Service

	SendTx(ctx context.Context, from, to string, amount *big.Int, balanceCheck bool) error
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
