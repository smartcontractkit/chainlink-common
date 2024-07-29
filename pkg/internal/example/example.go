// Package example contains helpers for implementing testable examples.
package example

import (
	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// Logger returns a logger.Logger which outputs simplified, plaintext logs to std out, without timestamps or caller info.
func Logger() (logger.Logger, error) {
	return logger.NewWith(func(config *zap.Config) {
		config.OutputPaths = []string{"stdout"}
		config.Encoding = "console"
		config.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		config.EncoderConfig.TimeKey = ""
		config.EncoderConfig.CallerKey = ""
	})
}
