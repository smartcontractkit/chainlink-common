package test

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainreader"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
)

// WrapChainReaderTesterForLoop allows you to test a [types.ChainReader] implementation behind a LOOP server
func WrapChainReaderTesterForLoop[T TestingT[T]](wrapped ChainReaderInterfaceTester[T]) ChainReaderInterfaceTester[T] {
	return &chainReaderLoopTester[T]{ChainReaderInterfaceTester: wrapped}
}

type chainReaderLoopTester[T TestingT[T]] struct {
	ChainReaderInterfaceTester[T]
	lst loopServerTester[T]
}

func (c *chainReaderLoopTester[T]) Setup(t T) {
	c.ChainReaderInterfaceTester.Setup(t)
	chainReader := c.ChainReaderInterfaceTester.GetChainReader(t)
	c.lst.registerHook = func(server *grpc.Server) {
		if chainReader != nil {
			impl := chainreader.NewServer(chainReader)
			pb.RegisterChainReaderServer(server, impl)
		}
	}
	c.lst.Setup(t)
}

func (c *chainReaderLoopTester[T]) GetChainReader(t T) types.ChainReader {
	return chainreader.NewClient(nil, c.lst.GetConn(t))
}

func (c *chainReaderLoopTester[T]) Name() string {
	return c.ChainReaderInterfaceTester.Name() + " on loop"
}
