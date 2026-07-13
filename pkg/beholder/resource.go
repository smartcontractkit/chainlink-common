package beholder

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

const (
	attrKeyCSAPublicKey = "csa_public_key"
	attrKeyNodeID       = "node_id"
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
		"k8s.pod.uid":         {},
		"k8s.pod.ip":          {},
		"k8s.pod.name":        {},
	}
)

func buildOtelResources(cfg Config) (ResourcePair, error) {
	full, err := buildFullOtelResource(cfg)
	if err != nil {
		return ResourcePair{}, err
	}

	if !cfg.ReducedMetricResourceAttributesEnabled &&
		!cfg.ExcludeVolatileResourceAttributesFromMetricsEnabled {
		return ResourcePair{Full: full, Metric: full}, nil
	}

	metric, err := buildMetricOtelResource(cfg, full)
	if err != nil {
		return ResourcePair{}, err
	}

	return ResourcePair{Full: full, Metric: metric}, nil
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
	var attrs []attribute.KeyValue

	if cfg.ReducedMetricResourceAttributesEnabled {
		serviceName := serviceNameFrom(cfg)
		attrs = append(attrs, attribute.String("service.name", serviceName))
		attrs = append(attrs, identityResourceAttributes(cfg)...)
		attrs = append(attrs, filterResourceAttributes(cfg.ResourceAttributes, reducedResourceAttributeFilter())...)
	} else {
		attrs = resourceAttributesAsKV(full)
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

	if cfg.NodeID != "" {
		attrs = append(attrs, attribute.String(attrKeyNodeID, cfg.NodeID))
	}

	return attrs
}

func serviceNameFrom(cfg Config) string {
	for _, attr := range cfg.ResourceAttributes {
		if string(attr.Key) == "service.name" {
			return attr.Value.AsString()
		}
	}
	return "chainlink"
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
		_, denied := volatileResourceAttributeDenylist[key]
		return denied
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
