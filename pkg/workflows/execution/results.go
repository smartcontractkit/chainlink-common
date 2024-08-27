package execution

import "github.com/smartcontractkit/chainlink-common/pkg/values"

type Results interface {
	GetResultForStep(string) (*Result, bool)
}

type Result struct {
	Inputs  values.Value
	Outputs values.Value
	Error   error
}
