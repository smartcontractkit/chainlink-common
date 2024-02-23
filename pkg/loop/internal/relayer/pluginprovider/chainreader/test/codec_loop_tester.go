package test

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainreader"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
)

// WrapCodecTesterForLoop allows you to test a [types.Codec] implementation behind a LOOP server
func WrapCodecTesterForLoop[T TestingT[T]](wrapped CodecInterfaceTester[T]) CodecInterfaceTester[T] {
	return &codecReaderLoopTester[T]{CodecInterfaceTester: wrapped}
}

type codecReaderLoopTester[T TestingT[T]] struct {
	CodecInterfaceTester[T]
	lst loopServerTester[T]
}

func (c *codecReaderLoopTester[T]) Setup(t T) {
	c.CodecInterfaceTester.Setup(t)
	codec := c.CodecInterfaceTester.GetCodec(t)
	c.lst.registerHook = func(server *grpc.Server) {
		if codec != nil {
			impl := chainreader.NewCodecServer(codec)
			pb.RegisterCodecServer(server, impl)
		}
	}
	c.lst.Setup(t)
}

func (c *codecReaderLoopTester[T]) Name() string {
	return c.CodecInterfaceTester.Name() + " on loop"
}

func (c *codecReaderLoopTester[T]) GetCodec(t T) types.Codec {
	return chainreader.NewCodecClient(nil, c.lst.GetConn(t))
}

func (c *codecReaderLoopTester[T]) IncludeArrayEncodingSizeEnforcement() bool {
	return false
}
