package tests

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/beholdertest"
)

// Deprecated: use beholdertest.Observer
type BeholderTester = beholdertest.Observer

// Deprecated: use beholdertest.NewObserver
func Beholder(t *testing.T) BeholderTester {
	t.Helper()
	return beholdertest.NewObserver(t)
}
