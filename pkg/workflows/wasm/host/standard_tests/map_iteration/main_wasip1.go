package main

import (
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	m := map[string]int{
		"alpha":   1,
		"bravo":   2,
		"charlie": 3,
		"delta":   4,
		"echo":    5,
		"foxtrot": 6,
		"golf":    7,
		"hotel":   8,
		"india":   9,
		"juliet":  10,
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	rawsdk.SendResponse(strings.Join(keys, ","))
}
