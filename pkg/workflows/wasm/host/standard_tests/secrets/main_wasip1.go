package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/standard_tests/internal/rawsdk"
)

func main() {
	s, err := rawsdk.GetSecrets("Foo")
	if err != nil {
		rawsdk.SendResponse(err.Error())
	} else {
		rawsdk.SendResponse(s)
	}
}
