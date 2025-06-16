//go:build wasip1

package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testhelpers/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2"
)

func main() {
	testhelpers.RunIdenticalTriggersWorkflow(wasm.New(func(configBytes []byte) (string, error) {
		return string(configBytes), nil
	}))
}
