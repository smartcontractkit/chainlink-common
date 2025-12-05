package contexts

import (
	"context"
	"errors"
)

const chainSelectorCtxKey key = "chainSelectorCtx"

// WithChainSelector returns a new context that includes the chain selector.
// Use ChainSelectorValue to get the value.
func WithChainSelector(ctx context.Context, cs uint64) context.Context {
	return context.WithValue(ctx, chainSelectorCtxKey, cs)
}

// ChainSelectorValue returns the chain selector, if one was set via WithChainSelector.
func ChainSelectorValue(ctx context.Context) (uint64, error) {
	val := Value[uint64](ctx, chainSelectorCtxKey)
	if val == 0 {
		return 0, errors.New("context missing chain selector")
	}
	return val, nil
}
