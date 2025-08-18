package exec

import "github.com/smartcontractkit/chainlink-protos/cre/go/values"

type Results interface {
	ResultForStep(string) (*Result, bool)
}

type Result struct {
	Inputs  values.Value
	Outputs values.Value
	Error   error
}
