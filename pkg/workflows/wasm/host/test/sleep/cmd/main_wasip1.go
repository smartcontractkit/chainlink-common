package main

import (
	"fmt"
	"time"

	// Registers the env.version_v2 import (and the other v2 host imports) so
	// the host links this guest with the v2/NoDAG path.
	_ "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	time.Sleep(1 * time.Hour)
	fmt.Printf("hello")
}
