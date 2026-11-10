package beholder

import (
	"go.opentelemetry.io/otel"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// NewOtelErrorHandler returns an [otel.ErrorHandler] that logs OpenTelemetry SDK errors, such as failed metric exports,
// as "Telemetry error" at error level. Without one, the SDK falls back to log.Print on stderr.
// Pass a logger that does not stream logs through OTel, or export failures will feed back into the failing pipeline.
func NewOtelErrorHandler(lggr logger.Logger) otel.ErrorHandler {
	return otel.ErrorHandlerFunc(func(err error) {
		lggr.Errorw("Telemetry error", "err", err)
	})
}
