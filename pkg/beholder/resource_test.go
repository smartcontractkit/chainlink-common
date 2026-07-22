package beholder

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestNewOtelResourceCustomAttributesOverrideDefaults(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ResourceAttributes = append(
		cfg.ResourceAttributes,
		attribute.String("service.name", "custom-service"),
	)

	resource, err := newOtelResource(cfg)
	require.NoError(t, err)

	serviceName, ok := resource.Set().Value(attribute.Key("service.name"))
	require.True(t, ok)
	require.Equal(t, "custom-service", serviceName.AsString())
}
