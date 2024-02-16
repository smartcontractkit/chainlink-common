package ccip

import (
	"context"
	"math/big"
	"time"
)

type PriceRegistryReader interface {
	// GetTokenPriceUpdatesCreatedAfter returns all the token price updates that happened after the provided timestamp.
	// The returned updates are sorted by timestamp in ascending order.
	GetTokenPriceUpdatesCreatedAfter(ctx context.Context, ts time.Time, confirmations int) ([]TokenPriceUpdateWithTxMeta, error)

	// GetGasPriceUpdatesCreatedAfter returns all the gas price updates that happened after the provided timestamp.
	// The returned updates are sorted by timestamp in ascending order.
	GetGasPriceUpdatesCreatedAfter(ctx context.Context, chainSelector uint64, ts time.Time, confirmations int) ([]GasPriceUpdateWithTxMeta, error)

	// Address returns the address of the price registry.
	Address() Address

	GetFeeTokens(ctx context.Context) ([]Address, error)

	// GetTokenPrices returns the latest price and time of quote of the given tokens.
	GetTokenPrices(ctx context.Context, wantedTokens []Address) ([]TokenPriceUpdate, error)

	GetTokensDecimals(ctx context.Context, tokenAddresses []Address) ([]uint8, error)

	Close() error
}

type TokenPriceUpdateWithTxMeta struct {
	TxMeta
	TokenPriceUpdate
}

// TokenPriceUpdate represents a token price at the last it was quoted.
type TokenPriceUpdate struct {
	TokenPrice
	TimestampUnixSec *big.Int
}

// GasPriceUpdateWithTxMeta represents a gas price update with transaction metadata.
type GasPriceUpdateWithTxMeta struct {
	TxMeta
	GasPriceUpdate
}

// GasPriceUpdate represents a gas price at the last it was quoted.
type GasPriceUpdate struct {
	GasPrice
	TimestampUnixSec *big.Int
}
