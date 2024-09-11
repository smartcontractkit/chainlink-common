package contractreader_test

import (
	"testing"

	chaincomponentstest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader/test"
	"github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
)

func FuzzCodec(f *testing.F) {
	interfaceTester := chaincomponentstest.WrapCodecTesterForLoop(&fakeCodecInterfaceTester{impl: &fakeCodec{}})
	interfacetests.RunCodecInterfaceFuzzTests(f, interfaceTester)
}
