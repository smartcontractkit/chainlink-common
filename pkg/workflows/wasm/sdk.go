package wasm

import (
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type Runtime struct {
	fetchFn func(req sdk.FetchRequest) (sdk.FetchResponse, error)
	emitFn  func(msg string, labels map[string]any) error
	logger  logger.Logger
}

type RuntimeConfig struct {
	MaxFetchResponseSizeBytes int64
}

const (
	defaultMaxFetchResponseSizeBytes = 5 * 1024
)

func defaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		MaxFetchResponseSizeBytes: defaultMaxFetchResponseSizeBytes,
	}
}

var _ sdk.Runtime = (*Runtime)(nil)

func (r *Runtime) Fetch(req sdk.FetchRequest) (sdk.FetchResponse, error) {
	return r.fetchFn(req)
}

func (r *Runtime) Logger() logger.Logger {
	return r.logger
}

func (r *Runtime) Emit(msg string, labels map[string]any) error {
	return r.emitFn(msg, labels)
}
