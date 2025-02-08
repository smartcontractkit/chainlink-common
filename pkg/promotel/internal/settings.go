package internal

import (
	"github.com/google/uuid"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/receiver"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var defaultComponentType = component.MustNewType("nop")

// NewReceiverSettings returns a new settings for factory.CreateMetrics function
func NewReceiverSettings(lggr logger.Logger) receiver.Settings {
	return receiver.Settings{
		ID:                component.NewIDWithName(defaultComponentType, uuid.NewString()),
		TelemetrySettings: NewTelemetrySettings(lggr),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}
}

func NewExporterSettings(lggr logger.Logger) exporter.Settings {
	return exporter.Settings{
		ID:                component.NewIDWithName(defaultComponentType, uuid.NewString()),
		TelemetrySettings: NewTelemetrySettings(lggr),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}
}
