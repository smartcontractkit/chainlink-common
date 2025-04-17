package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2"
)

func main() {
	runner := wasm.NewDonRunner()
	pkg.Workflow(runner)
}
