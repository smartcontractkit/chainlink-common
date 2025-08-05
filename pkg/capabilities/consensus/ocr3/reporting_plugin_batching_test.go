package ocr3

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

const (
	defaultBatchSizeMiB = 1024 * 1024 * 5 // 5 MB
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
		_, serialized, err := packToSizeLimit(lggr, QueriesSerializable{all}, defaultBatchSizeMiB)

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

// TestBinarySearchEdgeCase tests the specific edge case that was fixed in the binary search
func TestBinarySearchEdgeCase(t *testing.T) {
	lggr := logger.Test(t)

	// Create a test serializable that has predictable sizes
	testData := &testSerializable{
		items:    []string{"item1", "item2", "item3", "item4", "item5"},
		baseSize: 100, // Each item adds 100 bytes
	}

	// Set a limit that would exactly fit 3 items (300 bytes)
	testLimit := 300

	// The bug would have missed the optimal solution when:
	// - mid=3 fits (300 bytes <= 300 limit) -> low = 4
	// - mid=4 doesn't fit (400 bytes > 300 limit) -> high = 3 (BUG: was mid-1=3)
	// - With the bug: loop exits with low=4, high=3
	// - With the fix: high=4, then mid=4 doesn't fit -> high=4, finds optimal solution

	executionIDs, serialized, err := packToSizeLimit(lggr, testData, testLimit)
	require.NoError(t, err)
	require.NotEmpty(t, serialized)

	// Should find the maximum number of items that fit
	require.Equal(t, 5, len(executionIDs)) // Should pack all items since they all fit
}

// testSerializable is a simple test implementation of Serializable
type testSerializable struct {
	items    []string
	baseSize int
}

func (t *testSerializable) Serialize(lggr logger.Logger) ([]string, []byte, error) {
	size := len(t.items) * t.baseSize
	data := make([]byte, size)
	return t.items, data, nil
}

func (t *testSerializable) Len() int {
	return len(t.items)
}

func (t *testSerializable) Mid(mid int) Serializable {
	if mid < 0 {
		return &testSerializable{items: []string{}, baseSize: t.baseSize}
	}
	if mid >= len(t.items) {
		return t
	}
	return &testSerializable{items: t.items[:mid], baseSize: t.baseSize}
}
