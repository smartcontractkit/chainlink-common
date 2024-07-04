package ccipocr3

import (
	"context"
	"math/big"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type TokenPricesReader interface {
	// GetTokenPricesUSD returns the prices of the provided tokens in USD.
	// The order of the returned prices corresponds to the order of the provided tokens.
	GetTokenPricesUSD(ctx context.Context, tokens []types.Account) ([]*big.Int, error)
}

type CommitPluginCodec interface {
	Encode(context.Context, CommitPluginReport) ([]byte, error)
	Decode(context.Context, []byte) (CommitPluginReport, error)
}

type ExecutePluginCodec interface {
	Encode(context.Context, ExecutePluginReport) ([]byte, error)
	Decode(context.Context, []byte) (ExecutePluginReport, error)
}

type MessageHasher interface {
	Hash(context.Context, CCIPMsg) (Bytes32, error)
}
