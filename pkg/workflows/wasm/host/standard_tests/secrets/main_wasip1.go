package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	s, err := rawsdk.GetSecret("Foo")
	if err != nil {
		rawsdk.SendError(err)
	} else {
		rawsdk.SendResponse(s)
	}
}
