package testutils

import (
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type NoopRuntime struct{}

var _ sdk.Runtime = &NoopRuntime{}

func (nr *NoopRuntime) Fetch(sdk.FetchRequest) (sdk.FetchResponse, error) {
	return sdk.FetchResponse{}, nil
}

func (nr *NoopRuntime) Logger() logger.Logger {
	l, _ := logger.New()
	return l
}

func (nr *NoopRuntime) Emitter() sdk.MessageEmitter {
	return nil
}
