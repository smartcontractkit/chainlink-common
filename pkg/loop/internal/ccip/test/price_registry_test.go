package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestStaticPriceRegistry(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	// static test implementation is self consistent
	assert.NoError(t, PriceRegistryReader.Evaluate(ctx, PriceRegistryReader))

	// error when the test implementation is evaluates something that differs from the static implementation
	botched := PriceRegistryReader
	botched.addressResponse = "not what we expect"
	err := PriceRegistryReader.Evaluate(ctx, botched)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not what we expect")

}
