package internal

import (
	"github.com/google/uuid"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

var defaultComponentType = component.MustNewType("nop")

// NewReceiverSettings returns a new settings for factory.CreateMetrics function
func NewReceiverSettings(logger *zap.Logger) receiver.Settings {
	return receiver.Settings{
		ID:                component.NewIDWithName(defaultComponentType, uuid.NewString()),
		TelemetrySettings: NewTelemetrySettings(logger),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}
}

func NewExporterSettings(logger *zap.Logger) exporter.Settings {
	return exporter.Settings{
		ID:                component.NewIDWithName(defaultComponentType, uuid.NewString()),
		TelemetrySettings: NewTelemetrySettings(logger),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}
}
