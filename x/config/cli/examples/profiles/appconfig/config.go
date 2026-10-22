package appconfig

//go:generate go run ./gen

import "github.com/smartcontractkit/chainlink-common/pkg/config"

type ChainConfig struct {
	// ID is the chain to connect to. A built-in chain fills in the fields below.
	ID uint64 `validate:"required"`

	// RPC is the node to dial.
	RPC string `validate:"required"`

	// BlockTime is how often the chain produces a block.
	BlockTime config.Duration

	// Finality is how many blocks until a block is final.
	Finality uint32
}
