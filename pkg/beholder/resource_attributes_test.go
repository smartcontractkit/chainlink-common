package beholder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func TestResourceAttributesToStringMap(t *testing.T) {
	attrs := []attribute.KeyValue{
		attribute.String("chain_id", "1"),
		attribute.Bool("is_bootstrap", true),
		attribute.Int64("node_index", 42),
	}

	got := resourceAttributesToStringMap(attrs)

	assert.Equal(t, map[string]string{
		"chain_id":     "1",
		"is_bootstrap": "true",
		"node_index":   "42",
	}, got)
}

func TestResourceAttributesToStringMap_Empty(t *testing.T) {
	assert.Empty(t, resourceAttributesToStringMap(nil))
}
