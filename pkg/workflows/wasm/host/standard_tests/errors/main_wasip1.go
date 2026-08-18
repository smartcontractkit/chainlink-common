package main

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	rawsdk.SendError(errors.New("workflow execution failure"))
}
