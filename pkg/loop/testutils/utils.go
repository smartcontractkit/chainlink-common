package testutils

import (
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainreader/test"
	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
)

// This file exposes functions from pkg/loop/internal/test without exposing internal details.
// the duplication of the function is required so that the test of the LOOP servers themselves
// can dog food the same testers without creating a circular dependency.

// WrapChainReaderTesterForLoop allows you to test a [types.ChainReader] implementation behind a LOOP server
func WrapChainReaderTesterForLoop[T TestingT[T]](wrapped ChainReaderInterfaceTester[T]) ChainReaderInterfaceTester[T] {
	return test.WrapChainReaderTesterForLoop(wrapped)
}

func WrapCodecTesterForLoop[T TestingT[T]](wrapped CodecInterfaceTester[T]) CodecInterfaceTester[T] {
	return test.WrapCodecTesterForLoop(wrapped)
}
