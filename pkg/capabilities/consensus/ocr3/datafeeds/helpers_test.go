package datafeeds

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

var LLOStreamPrices = lloStreamPrices

type LLOAggregatorConfig = lloAggregatorConfig

type WrappableUpdate = WrappableStreamUpdate

type FeedConfig = feedConfig

func NewLLOconfig(t *testing.T, m map[uint32]FeedConfig, opts ...lloConfigOpt) values.Map {
	unwrappedConfig := map[string]any{
		"streams": map[string]any{},
	}

	for feedID, cfg := range m {
		unwrappedConfig["streams"].(map[string]any)[strconv.FormatUint(uint64(feedID), 10)] = map[string]any{
			"deviation": cfg.Deviation.String(),
			"heartbeat": cfg.Heartbeat,
			//			"remappedID": cfg.RemappedID,
		}
	}
	for _, opt := range opts {
		opt(t, unwrappedConfig)
	}

	t.Logf("unwrappedConfig: %v", unwrappedConfig)
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return *config
}

type lloConfigOpt = func(t *testing.T, m map[string]any)

func LLOConfigAllowStaleness(staleness float64) lloConfigOpt {
	return func(t *testing.T, m map[string]any) {
		m["allowedPartialStaleness"] = strconv.FormatFloat(staleness, 'f', -1, 64)
	}
}
