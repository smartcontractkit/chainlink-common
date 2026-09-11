package main

import (
	// Registers the env.version_v2 import (and the other v2 host imports) so
	// the host links this guest with the v2/NoDAG path.
	_ "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	// allocate more bytes than the binary should be able to access
	_ = make([]byte, int64(512e6))
}
