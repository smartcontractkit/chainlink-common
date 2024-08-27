package chaincomponentstest

import (
	"testing"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainreader"
	"github.com/smartcontractkit/chainlink-common/pkg/types"

	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests" //nolint common practice to import test mods with .
)

func TestAllEncodings(t *testing.T, test func(chainreader.EncodingVersion) func(t *testing.T)) {
	t.Helper()

	encodings := []struct {
		Name    string
		Version chainreader.EncodingVersion
	}{
		{Name: "JSONv1", Version: chainreader.JSONEncodingVersion1},
		{Name: "JSONv2", Version: chainreader.JSONEncodingVersion2},
		{Name: "CBOR", Version: chainreader.CBOREncodingVersion},
	}

	for idx := range encodings {
		encoding := encodings[idx]

		t.Run(encoding.Name, test(encoding.Version))
	}
}

type LoopTesterOpt func(*contractReaderLoopTester)

// WrapChainComponentsTesterForLoop allows you to test a [types.ContractReader] and [types.ChainWriter] implementation behind a LOOP server
func WrapChainComponentsTesterForLoop(wrapped ChainComponentsInterfaceTester[*testing.T], opts ...LoopTesterOpt) ChainComponentsInterfaceTester[*testing.T] {
	tester := &contractReaderLoopTester{
		ChainComponentsInterfaceTester: wrapped,
		encodeWith:                     chainreader.DefaultEncodingVersion,
	}

	for _, opt := range opts {
		opt(tester)
	}

	return tester
}

func WithChainReaderLoopEncoding(version chainreader.EncodingVersion) LoopTesterOpt {
	return func(tester *contractReaderLoopTester) {
		tester.encodeWith = version
	}
}

type contractReaderLoopTester struct {
	ChainComponentsInterfaceTester[*testing.T]
	lst        loopServerTester
	encodeWith chainreader.EncodingVersion
}

func (c *contractReaderLoopTester) Setup(t *testing.T) {
	c.ChainComponentsInterfaceTester.Setup(t)
	chainReader := c.ChainComponentsInterfaceTester.GetChainReader(t)

	c.lst.registerHook = func(server *grpc.Server) {
		if chainReader != nil {
			impl := chainreader.NewServer(chainReader, chainreader.WithServerEncoding(c.encodeWith))
			pb.RegisterChainReaderServer(server, impl)
		}
	}

	c.lst.Setup(t)
}

func (c *contractReaderLoopTester) GetChainReader(t *testing.T) types.ContractReader {
	return chainreader.NewClient(nil, c.lst.GetConn(t), chainreader.WithClientEncoding(c.encodeWith))
}

func (c *contractReaderLoopTester) Name() string {
	return c.ChainComponentsInterfaceTester.Name() + " on loop"
}
