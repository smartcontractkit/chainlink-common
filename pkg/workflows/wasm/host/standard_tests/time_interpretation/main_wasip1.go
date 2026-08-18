package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func main() {
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_DON))

	t := rawsdk.Now()

	isoString := t.UTC().Format("2006-01-02T15:04:05Z07:00")

	rawsdk.SendResponse(isoString)
}
