package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
)

func main() {
	runner := wasm.DonRunner()
	var r sdk.DonRunner = runner
}
