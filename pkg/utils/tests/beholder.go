package tests

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/beholdertest"
)

// Deprecated: use beholdertest.Observer
//
//go:fix inline
type BeholderTester = beholdertest.Observer

// Deprecated: use beholdertest.NewObserver
//
//go:fix inline
func Beholder(t *testing.T) beholdertest.Observer {
	return beholdertest.NewObserver(t)
}
