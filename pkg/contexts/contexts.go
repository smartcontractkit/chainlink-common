package contexts

import (
	"context"
	"strings"
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
// The values will be normalized via CRE.Normalized.
func WithCRE(ctx context.Context, cre CRE) context.Context {
	return context.WithValue(ctx, creCtxKey, cre.Normalized())
}

// CREValue returns the [CRE] key/val, which may be empty.
// If it is not empty, the values will be normalized.
func CREValue(ctx context.Context) CRE {
	return Value[CRE](ctx, creCtxKey)
}

// CRE holds contextual Chainlink Runtime Environment metadata.
// This can include organization, owner, and workflow information.
// When a value is present, the higher scoped values are normally also available - except for Org, which may not be set.
// Typically injected via [context.Context]. See WithCRE & CREValue.
type CRE struct {
	Org             string // may be missing even if others are present.
	Owner, Workflow string
}

// Normalized returns a possibly modified CRE with normalized values.
func (c CRE) Normalized() CRE {
	c.Org = strings.TrimPrefix(c.Org, "org_")
	// not hex like the others, so don't look for 0x or change case

	c.Owner = strings.TrimPrefix(c.Owner, "owner_")
	c.Owner = strings.TrimPrefix(c.Owner, "0x")
	c.Owner = strings.ToLower(c.Owner)

	c.Workflow = strings.TrimPrefix(c.Workflow, "0x")
	c.Workflow = strings.ToLower(c.Workflow)
	return c
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
