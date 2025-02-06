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
func NewReceiverSettings() (receiver.Settings, error) {
	l, err := zap.NewProduction()
	return receiver.Settings{
		ID:                component.NewIDWithName(defaultComponentType, uuid.NewString()),
		TelemetrySettings: NewTelemetrySettings(l),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}, err
}

func NewExporterSettings() (exporter.Settings, error) {
	l, err := zap.NewProduction()
	return exporter.Settings{
		ID:                component.NewIDWithName(defaultComponentType, uuid.NewString()),
		TelemetrySettings: NewTelemetrySettings(l),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}, err
}
