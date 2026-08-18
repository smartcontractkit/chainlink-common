package promutil

import (
	"slices"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type prefixGatherer struct {
	gatherer prometheus.Gatherer
	prefixes []string
}

// NewPrefixGatherer  returns a prometheus gatherer that will only produce metrics matching one of the prefixes.
func NewPrefixGatherer(gatherer prometheus.Gatherer, prefixes []string) prometheus.Gatherer {
	return &prefixGatherer{gatherer, slices.DeleteFunc(prefixes, func(s string) bool {
		return s == "" // ignore empty, which would match everything
	})}
}

func (g *prefixGatherer) Gather() ([]*dto.MetricFamily, error) {
	if len(g.prefixes) == 0 {
		return g.gatherer.Gather()
	}
	var ret []*dto.MetricFamily
	all, err := g.gatherer.Gather()
	if err != nil {
		return nil, err
	}
	for _, m := range all {
		if m.Name == nil {
			continue
		}
		for _, prefix := range g.prefixes {
			if strings.HasPrefix(*m.Name, prefix) {
				ret = append(ret, m)
				break
			}
		}
	}
	return ret, nil
}
