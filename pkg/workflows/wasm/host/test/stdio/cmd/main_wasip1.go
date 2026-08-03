package main

import (
	"fmt"
	"os"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func main() {
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_DON))
	request := rawsdk.GetRequest()

	fmt.Println("stdout from guest")
	fmt.Fprintln(os.Stderr, "stderr from guest")

	rawsdk.SendResponse(request.Config)
}
