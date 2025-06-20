package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/standard_tests/internal/rawsdk"
)

func main() {
	request := rawsdk.GetRequest()
	rawsdk.SendResponse(request.Config)
}
