package pkg

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/usercode/pkg2"
)

// A second file is used to make sure that all files in the package are collapsed into one correctly.

type MyType2 struct {
	Nested MyNestedType
	I      int
	S      string
	T      time.Time
	O      pkg2.OtherPackage
}
