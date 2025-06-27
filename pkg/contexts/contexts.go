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

func WithCRE(ctx context.Context, cre CRE) context.Context {
	return context.WithValue(ctx, creCtxKey, &cre)
}

func CREValue(ctx context.Context) CRE {
	v := Value[*CRE](ctx, creCtxKey)
	if v == nil {
		return CRE{}
	}
	return *v // copy
}

type CRE struct {
	Org, Owner, Workflow string
}

type key string

const creCtxKey key = "creCtx"
