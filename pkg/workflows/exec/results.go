package exec

import "github.com/smartcontractkit/chainlink-common/pkg/values"

type Results interface {
	ResultForStep(string) (*Result, bool)
}

type Result struct {
	Inputs  values.Value
	Outputs values.Value
	Error   error
}
