package chainreader_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
)

func FuzzCodec(f *testing.F) {
	interfaceTester := WrapCodecTesterForLoop(&fakeCodecInterfaceTester{impl: &fakeCodec{}})
	interfacetests.RunCodecInterfaceFuzzTests(f, interfaceTester)
}
