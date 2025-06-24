package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	request := rawsdk.GetRequest()
	msg := []byte("log from wasm!")
	rawsdk.Log(rawsdk.BufferToPointerLen(msg))
	rawsdk.SendResponse(request.Config)
}
