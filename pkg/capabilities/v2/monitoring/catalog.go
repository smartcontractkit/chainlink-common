package monitoring

import (
	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

const (
	MetricPrefix = "capabilities_v2"

	ActionCountMetric    = MetricPrefix + "_action_count"
	ActionDurationMetric = MetricPrefix + "_action_duration"

	InitSuccessMetric = "chain_capability_initialization_success"
	InitFailureMetric = "chain_capability_initialization_failure"

	OutcomeSuccess = "success"
	OutcomeError   = "error"

	LabelOutcome         = "outcome"
	LabelMethod          = "method"
	LabelChainFamilyName = "chain_family_name"
	LabelChainID         = "chain_id"
	LabelNetworkName     = "network_name"
	LabelNetworkNameFull = "network_name_full"
	LabelWorkflowDonID   = "workflow_don_id"
	LabelCapabilityType  = "capability_type"
	LabelCapabilityID    = "capability_id"
)

// ActionLatencyBucketBoundariesMs matches read-action latency buckets.
var ActionLatencyBucketBoundariesMs = []float64{
	0, 5, 10, 25, 50, 75, 100,
	250, 500, 750, 1000,
	2500, 5000, 7500, 10000,
	15000, 30000,
}

type actionInstrumentInfo struct {
	count    beholder.MetricInfo
	duration beholder.MetricInfo
}

type initInstrumentInfo struct {
	success beholder.MetricInfo
	failure beholder.MetricInfo
}

func newActionInstrumentInfo() actionInstrumentInfo {
	return actionInstrumentInfo{
		count: beholder.MetricInfo{
			Name:        ActionCountMetric,
			Description: "The count of v2 capability action lifecycle events",
		},
		duration: beholder.MetricInfo{
			Name:        ActionDurationMetric,
			Unit:        "ms",
			Description: "The duration since capability exec start to action lifecycle emit",
		},
	}
}

func newInitInstrumentInfo() initInstrumentInfo {
	return initInstrumentInfo{
		success: beholder.MetricInfo{
			Name:        InitSuccessMetric,
			Description: "1 if the v2 capability initialised successfully, 0 otherwise",
		},
		failure: beholder.MetricInfo{
			Name:        InitFailureMetric,
			Description: "1 if the v2 capability failed to initialise, 0 otherwise",
		},
	}
}
