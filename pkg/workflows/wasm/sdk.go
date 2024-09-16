package wasm

import "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"

type Runtime struct{}

var _ sdk.Runtime = (*Runtime)(nil)
