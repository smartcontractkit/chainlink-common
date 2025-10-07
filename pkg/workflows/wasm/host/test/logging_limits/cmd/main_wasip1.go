package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func main() {
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_DON))
	request := rawsdk.GetRequest()
	msg := []byte("short log 1")
	rawsdk.Log(rawsdk.BufferToPointerLen(msg))
	msg = []byte("super duper excessively long log 2") // exceeding 20 byte limit set in the test
	rawsdk.Log(rawsdk.BufferToPointerLen(msg))
	msg = []byte("short log 3")
	rawsdk.Log(rawsdk.BufferToPointerLen(msg))
	msg = []byte("short log 4")
	rawsdk.Log(rawsdk.BufferToPointerLen(msg))
	rawsdk.SendResponse(request.Config)
}
