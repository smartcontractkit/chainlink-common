package testutils

import (
	"testing"

	test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainreader/test"
	"github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
)

// This file exposes functions from pkg/loop/internal/test without exposing internal details.
// the duplication of the function is required so that the test of the LOOP servers themselves
// can dog food the same testers without creating a circular dependency.

// WrapChainComponentsTesterForLoop allows you to test a [types.ContractReader] implementation behind a LOOP server
func WrapChainComponentsTesterForLoop(wrapped interfacetests.ChainComponentsInterfaceTester[*testing.T]) interfacetests.ChainComponentsInterfaceTester[*testing.T] {
	return test.WrapChainComponentsTesterForLoop(wrapped)
}

func WrapCodecTesterForLoop(wrapped interfacetests.CodecInterfaceTester) interfacetests.CodecInterfaceTester {
	return test.WrapCodecTesterForLoop(wrapped)
}
