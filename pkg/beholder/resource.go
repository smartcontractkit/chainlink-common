package beholder

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

const (
	attrKeyCSAPublicKey = "csa_public_key"
)

// ResourcePair holds full and metric-scoped OTel resources.
type ResourcePair struct {
	Full   *sdkresource.Resource
	Metric *sdkresource.Resource
}

var (
	reducedMetricResourceExactDenylist = map[string]struct{}{
		"service.sha":          {},
		"service.shortversion": {},
		"package_name":         {},
	}

	reducedMetricResourcePrefixDenylist = []string{
		"os.",
		"host.",
		"container.",
	}

	volatileResourceAttributeDenylist = map[string]struct{}{
		"process.pid":         {},
		"service.instance.id": {},
		"container.id":        {},
		"host.id":             {},
		"host.name":           {},
	}

	volatileResourceAttributePrefixDenylist = []string{"k8s.pod."}
)

func buildOtelResources(cfg Config) (ResourcePair, error) {
	full, err := buildFullOtelResource(cfg)
	if err != nil {
		return ResourcePair{}, err
	}

	if !cfg.metricResourceFilteringEnabled() {
		return ResourcePair{Full: full, Metric: full}, nil
	}

	metric, err := buildMetricOtelResource(cfg, full)
	if err != nil {
		return ResourcePair{}, err
	}

	return ResourcePair{Full: full, Metric: metric}, nil
}

func (cfg Config) metricResourceFilteringEnabled() bool {
	return cfg.ReducedMetricResourceAttributesEnabled ||
		cfg.ExcludeVolatileResourceAttributesFromMetricsEnabled
}

func buildFullOtelResource(cfg Config) (*sdkresource.Resource, error) {
	extraResources, err := sdkresource.New(
		context.Background(),
		sdkresource.WithOS(),
		sdkresource.WithContainer(),
		sdkresource.WithHost(),
	)
	if err != nil {
		return nil, err
	}

	resource, err := sdkresource.Merge(
		sdkresource.Default(),
		extraResources,
	)
	if err != nil {
		return nil, err
	}

	resource, err = sdkresource.Merge(
		sdkresource.NewSchemaless(identityResourceAttributes(cfg)...),
		resource,
	)
	if err != nil {
		return nil, err
	}

	resource, err = sdkresource.Merge(
		sdkresource.NewSchemaless(cfg.ResourceAttributes...),
		resource,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func buildMetricOtelResource(cfg Config, full *sdkresource.Resource) (*sdkresource.Resource, error) {
	attrs := resourceAttributesAsKV(full)
	if cfg.ReducedMetricResourceAttributesEnabled {
		attrs = filterResourceAttributes(attrs, reducedResourceAttributeFilter())
	}

	if cfg.ExcludeVolatileResourceAttributesFromMetricsEnabled {
		attrs = filterResourceAttributes(attrs, volatileResourceAttributeFilter())
	}

	return sdkresource.NewSchemaless(attrs...), nil
}

func identityResourceAttributes(cfg Config) []attribute.KeyValue {
	csaPublicKeyHex := "not-configured"
	if len(cfg.AuthPublicKeyHex) > 0 {
		csaPublicKeyHex = cfg.AuthPublicKeyHex
	}

	attrs := []attribute.KeyValue{
		attribute.String(attrKeyCSAPublicKey, csaPublicKeyHex),
	}

	return attrs
}

type resourceAttributeFilter func(key string) bool

func reducedResourceAttributeFilter() resourceAttributeFilter {
	return func(key string) bool {
		if _, denied := reducedMetricResourceExactDenylist[key]; denied {
			return true
		}
		for _, prefix := range reducedMetricResourcePrefixDenylist {
			if strings.HasPrefix(key, prefix) {
				return true
			}
		}
		return false
	}
}

func volatileResourceAttributeFilter() resourceAttributeFilter {
	return func(key string) bool {
		if _, denied := volatileResourceAttributeDenylist[key]; denied {
			return true
		}
		for _, prefix := range volatileResourceAttributePrefixDenylist {
			if strings.HasPrefix(key, prefix) {
				return true
			}
		}
		return false
	}
}

func filterResourceAttributes(
	attrs []attribute.KeyValue,
	deny resourceAttributeFilter,
) []attribute.KeyValue {
	if deny == nil {
		return attrs
	}

	filtered := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		if deny(string(attr.Key)) {
			continue
		}
		filtered = append(filtered, attr)
	}
	return filtered
}

func resourceAttributesAsKV(res *sdkresource.Resource) []attribute.KeyValue {
	if res == nil {
		return nil
	}

	var attrs []attribute.KeyValue
	iter := res.Iter()
	for iter.Next() {
		attrs = append(attrs, iter.Attribute())
	}
	return attrs
}
