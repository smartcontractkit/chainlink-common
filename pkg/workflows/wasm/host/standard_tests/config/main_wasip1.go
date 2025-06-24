package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	request := rawsdk.GetRequest()
	rawsdk.SendResponse(request.Config)
}
