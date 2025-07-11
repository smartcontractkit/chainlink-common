package contexts

import (
	"context"
)

// Value gets a value from the context and casts it to T.
func Value[T any](ctx context.Context, key any) T {
	v := ctx.Value(key)
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

// WithCRE returns a derived context with a cre key/val.
func WithCRE(ctx context.Context, cre CRE) context.Context {
	return context.WithValue(ctx, creCtxKey, &cre)
}

// CREValue returns the [CRE] key/val, which may be empty.
func CREValue(ctx context.Context) CRE {
	v := Value[*CRE](ctx, creCtxKey)
	if v == nil {
		return CRE{}
	}
	return *v // copy
}

// CRE holds contextual Chainlink Runtime Environment metadata.
// This can include organization, owner, and workflow information.
// When a value is present, the higher scoped values are normally also available - except for Org, which may not be set.
// Typically injected via [context.Context]. See WithCRE & CREValue.
type CRE struct {
	Org             string // may be missing even if others are present.
	Owner, Workflow string
}

func (c CRE) LoggerKVs() []any {
	return []any{
		"org", c.Org,
		"owner", c.Owner,
		"workflow", c.Workflow,
	}
}

type key string

const creCtxKey key = "creCtx"
