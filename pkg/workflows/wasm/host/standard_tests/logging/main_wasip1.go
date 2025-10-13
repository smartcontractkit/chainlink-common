package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func main() {
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_DON))
	request := rawsdk.GetRequest()
	msg := []byte("log from wasm!")
	rawsdk.Log(rawsdk.BufferToPointerLen(msg))
	rawsdk.SendResponse(request.Config)
}
