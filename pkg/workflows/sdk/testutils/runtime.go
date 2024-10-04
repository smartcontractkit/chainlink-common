package testutils

import (
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type NopRuntime struct{}

var _ sdk.Runtime = &NopRuntime{}

func (nr *NopRuntime) Fetch(sdk.FetchRequest) (sdk.FetchResponse, error) {
	return sdk.FetchResponse{}, nil
}

func (nr *NopRuntime) Logger() logger.Logger {
	l, _ := logger.New()
	return l
}
