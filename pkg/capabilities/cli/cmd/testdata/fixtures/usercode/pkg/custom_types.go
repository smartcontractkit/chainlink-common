package pkg

import (
	"time"
	
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/usercode/pkg2"
)

type MyType struct {
	Nested MyNestedType
	I      int
	S      string
	T      time.Time
	O      pkg2.OtherPackage
}

type MyNestedType struct {
	II int
	SS string
}
