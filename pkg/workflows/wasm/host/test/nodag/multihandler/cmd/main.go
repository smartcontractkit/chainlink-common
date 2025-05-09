//go:build wasip1

package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/testhelpers"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2"
)

func main() {
	testhelpers.RunIdenticalTriggersWorkflow(wasm.NewDonRunner())
}
