package chaincomponentstest

import (
	"testing"

	"google.golang.org/grpc"

	codecpb "github.com/smartcontractkit/chainlink-common/pkg/internal/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/types"

	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests" //nolint common practice to import test mods with .
)

func TestAllEncodings(t *testing.T, test func(codecpb.EncodingVersion) func(t *testing.T)) {
	t.Helper()

	encodings := []struct {
		Name    string
		Version codecpb.EncodingVersion
	}{
		{Name: "JSONv1", Version: codecpb.JSONEncodingVersion1},
		{Name: "JSONv2", Version: codecpb.JSONEncodingVersion2},
		{Name: "CBOR", Version: codecpb.CBOREncodingVersion},
	}

	for idx := range encodings {
		encoding := encodings[idx]

		t.Run(encoding.Name, test(encoding.Version))
	}
}

type LoopTesterOpt func(*contractReaderLoopTester)

// WrapContractReaderTesterForLoop allows you to test a [types.ContractReader] and [types.ContractWriter] implementation behind a LOOP server
func WrapContractReaderTesterForLoop(wrapped ChainComponentsInterfaceTester[*testing.T], opts ...LoopTesterOpt) ChainComponentsInterfaceTester[*testing.T] {
	tester := &contractReaderLoopTester{
		ChainComponentsInterfaceTester: wrapped,
		encodeWith:                     codecpb.DefaultEncodingVersion,
	}

	for _, opt := range opts {
		opt(tester)
	}

	return tester
}

func WithContractReaderLoopEncoding(version codecpb.EncodingVersion) LoopTesterOpt {
	return func(tester *contractReaderLoopTester) {
		tester.encodeWith = version
	}
}

type contractReaderLoopTester struct {
	ChainComponentsInterfaceTester[*testing.T]
	lst        loopServerTester
	conn       *grpc.ClientConn
	encodeWith codecpb.EncodingVersion
}

func (c *contractReaderLoopTester) Setup(t *testing.T) {
	c.ChainComponentsInterfaceTester.Setup(t)
	contractReader := c.ChainComponentsInterfaceTester.GetContractReader(t)

	c.lst.registerHook = func(server *grpc.Server) {
		if contractReader != nil {
			impl := contractreader.NewServer(contractReader, contractreader.WithServerEncoding(c.encodeWith))
			pb.RegisterContractReaderServer(server, impl)
		}
	}

	c.lst.Setup(t)
	c.conn = c.lst.GetConn(t)
}

func (c *contractReaderLoopTester) GetContractReader(t *testing.T) types.ContractReader {
	return contractreader.NewClient(nil, pb.NewContractReaderClient(c.conn), contractreader.WithClientEncoding(c.encodeWith))
}

func (c *contractReaderLoopTester) Name() string {
	return c.ChainComponentsInterfaceTester.Name() + " on loop"
}
