package wasm

import (
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type Runtime struct {
	Logger logger.Logger
}

var _ sdk.Runtime = (*Runtime)(nil)
