package ocr3

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

func TestReportPlugin_packQueriesToSizeLimit(t *testing.T) {
	lggr := logger.Test(t)

	testByCount := func(t *testing.T, countInStorage int) {
		s := requests.NewTestStore(t, countInStorage)
		mcap := &mockCapability{
			aggregator: &aggregator{},
			encoder:    &enc{},
			registeredWorkflows: map[string]bool{
				workflowTestID:  true,
				workflowTestID2: true,
			},
		}
		rp, err := NewReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{F: 1}, defaultOutcomePruningThreshold, lggr)
		require.NoError(t, err)

		all, err := rp.s.All()
		require.NoError(t, err)
		_, serialized, err := packToSizeLimit(lggr, QueriesSerializable{all})

		if countInStorage == 0 {
			require.Error(t, err)
			require.Empty(t, serialized)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, serialized)
		}
	}

	testCases := []int{0, 1, 3, 4, 5, 11, 50, 1000, 1_000_000}

	for _, count := range testCases {
		t.Run(fmt.Sprintf("test optimization (%d)", count), func(t *testing.T) {
			testByCount(t, count)
		})
	}
}
