//go:build !wasip1

package wasm

import "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"

func NewDonRunner() sdk.DonRunner {
	panic("error: NewRunner() is only intended for use with `GOOS=wasip1 GOARCH=wasm`. For testing, use testutils.NewRunner() instead.")
}
