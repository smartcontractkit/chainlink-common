package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2"
)

func main() {
	runner := wasm.NewRunner(sdk.ParseJson[pkg.Config])
	runner.Run(pkg.InitWorkflow)
}
