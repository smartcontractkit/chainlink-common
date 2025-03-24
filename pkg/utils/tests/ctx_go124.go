//go:build go1.24

package tests

import (
	"context"
)

// Deprecated: use [*testing.T.Context]
func Context(tb TestingT) context.Context {
	if hasCtx, ok := tb.(interface {
		Context() context.Context
	}); ok {
		return hasCtx.Context()
	}
	return getContext(tb)
}
