package beholder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestBuildOtelResources(t *testing.T) {
	baseConfig := func() Config {
		return Config{
			AuthPublicKeyHex: "csa-key",
			NodeID:           "node-1",
			ResourceAttributes: []attribute.KeyValue{
				attribute.String("service.name", "chainlink"),
				attribute.String("service.version", "1.2.3"),
				attribute.String("service.sha", "abcdef"),
				attribute.String("service.shortversion", "1.2.3@abcdef"),
				attribute.String("package_name", "beholder"),
				attribute.Int("process.pid", 1234),
				attribute.String("service.instance.id", "instance-1"),
				attribute.String("k8s.pod.owner", "deployment-1"),
				attribute.String("custom.attribute", "keep"),
			},
		}
	}

	cases := []struct {
		name                 string
		reduced, volatile    bool
		wantMetricKeys       []string
		wantAbsentMetricKeys []string
	}{
		{
			name: "default is backwards compatible",
			wantMetricKeys: []string{
				"service.sha", "package_name", "process.pid", attrKeyCSAPublicKey, attrKeyNodeID,
			},
		},
		{
			name:    "reduced removes detector and redundant attributes",
			reduced: true,
			wantMetricKeys: []string{
				"service.name", "service.version", "custom.attribute", "process.pid", "service.instance.id",
				attrKeyCSAPublicKey, attrKeyNodeID,
			},
			wantAbsentMetricKeys: []string{
				"service.sha", "service.shortversion", "package_name",
			},
		},
		{
			name:     "volatile removes process scoped attributes",
			volatile: true,
			wantMetricKeys: []string{
				"service.sha", "package_name", attrKeyCSAPublicKey, attrKeyNodeID,
			},
			wantAbsentMetricKeys: []string{
				"process.pid", "service.instance.id", "k8s.pod.owner",
			},
		},
		{
			name:     "reduced and volatile combine filtering",
			reduced:  true,
			volatile: true,
			wantMetricKeys: []string{
				"service.name", "service.version", "custom.attribute", attrKeyCSAPublicKey, attrKeyNodeID,
			},
			wantAbsentMetricKeys: []string{
				"service.sha", "service.shortversion", "package_name", "process.pid", "service.instance.id", "k8s.pod.owner",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := baseConfig()
			cfg.ReducedMetricResourceAttributesEnabled = tc.reduced
			cfg.ExcludeVolatileResourceAttributesFromMetricsEnabled = tc.volatile

			resources, err := buildOtelResources(cfg)
			require.NoError(t, err)

			fullAttrs := resourceAttributeMap(resources.Full)
			metricAttrs := resourceAttributeMap(resources.Metric)
			assert.Contains(t, fullAttrs, "process.pid")
			assert.Contains(t, fullAttrs, "service.instance.id")
			for _, key := range tc.wantMetricKeys {
				assert.Contains(t, metricAttrs, key)
			}
			for _, key := range tc.wantAbsentMetricKeys {
				assert.NotContains(t, metricAttrs, key)
			}
		})
	}
}

func TestReducedMetricResourcePreservesFullAttributesAndIdentity(t *testing.T) {
	cfg := Config{
		ReducedMetricResourceAttributesEnabled: true,
		AuthPublicKeyHex:                       "authoritative-csa-key",
		NodeID:                                 "authoritative-node-id",
		ResourceAttributes: []attribute.KeyValue{
			attribute.String(attrKeyCSAPublicKey, "caller-csa-key"),
			attribute.String(attrKeyNodeID, "caller-node-id"),
			attribute.String("k8s.pod.extra", "pod-attribute"),
		},
	}
	resources, err := buildOtelResources(cfg)
	require.NoError(t, err)

	fullAttrs := resourceAttributeMap(resources.Full)
	metricAttrs := resourceAttributeMap(resources.Metric)
	assert.Equal(t, "authoritative-csa-key", fullAttrs[attrKeyCSAPublicKey].AsString())
	assert.Equal(t, "authoritative-csa-key", metricAttrs[attrKeyCSAPublicKey].AsString())
	assert.Equal(t, "authoritative-node-id", fullAttrs[attrKeyNodeID].AsString())
	assert.Equal(t, "authoritative-node-id", metricAttrs[attrKeyNodeID].AsString())
	assert.Equal(t, fullAttrs["service.name"], metricAttrs["service.name"])

	cfg.ExcludeVolatileResourceAttributesFromMetricsEnabled = true
	resources, err = buildOtelResources(cfg)
	require.NoError(t, err)
	assert.NotContains(t, resourceAttributeMap(resources.Metric), "k8s.pod.extra")
}

func TestRecordFullResourceAttributesMetric(t *testing.T) {
	cfg := Config{
		ReducedMetricResourceAttributesEnabled: true,
		AuthPublicKeyHex:                       "csa-key",
		NodeID:                                 "node-1",
		ResourceAttributes: []attribute.KeyValue{
			attribute.String("service.name", "chainlink"),
			attribute.Int("process.pid", 1234),
		},
	}
	resources, err := buildOtelResources(cfg)
	require.NoError(t, err)

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(resources.Metric),
	)
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })

	client := Client{
		Meter:        provider.Meter(defaultPackageName),
		fullResource: resources.Full,
	}
	require.NoError(t, client.RecordFullResourceAttributesMetric(context.Background()))

	var metrics metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &metrics))

	for _, scopeMetrics := range metrics.ScopeMetrics {
		for _, metric := range scopeMetrics.Metrics {
			if metric.Name != "beholder.otel.resource_attributes.full" {
				continue
			}
			gauge, ok := metric.Data.(metricdata.Gauge[int64])
			require.True(t, ok)
			require.Len(t, gauge.DataPoints, 1)
			assert.Equal(t, int64(1), gauge.DataPoints[0].Value)
			_, hasPID := gauge.DataPoints[0].Attributes.Value("process.pid")
			assert.True(t, hasPID)
			return
		}
	}
	t.Fatal("full resource attributes metric not found")
}

func resourceAttributeMap(resource interface{ Attributes() []attribute.KeyValue }) map[string]attribute.Value {
	attrs := make(map[string]attribute.Value)
	for _, attr := range resource.Attributes() {
		attrs[string(attr.Key)] = attr.Value
	}
	return attrs
}
