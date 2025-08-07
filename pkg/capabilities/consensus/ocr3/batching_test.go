package ocr3

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbtypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	pbvalues "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

func TestCheckQuerySizeLimit(t *testing.T) {
	// Helper function to create a simple ID with predictable size
	createSimpleId := func(workflowExecutionId string) *pbtypes.Id {
		return &pbtypes.Id{
			WorkflowExecutionId: workflowExecutionId,
			WorkflowId:          "workflow-1",
			WorkflowOwner:       "owner",
			WorkflowName:        "test",
			ReportId:            "report-1",
			KeyId:               "key-1",
		}
	}

	// Helper function to create an ID with all fields populated for larger size
	createLargeId := func(suffix string) *pbtypes.Id {
		return &pbtypes.Id{
			WorkflowExecutionId:      "very-long-workflow-execution-id-" + suffix,
			WorkflowId:               "very-long-workflow-id-" + suffix,
			WorkflowOwner:            "very-long-workflow-owner-" + suffix,
			WorkflowName:             "very-long-workflow-name-" + suffix,
			ReportId:                 "very-long-report-id-" + suffix,
			WorkflowDonId:            12345,
			WorkflowDonConfigVersion: 67890,
			KeyId:                    "very-long-key-id-" + suffix,
		}
	}

	// Helper function to create an empty ID (zero values)
	createEmptyId := func() *pbtypes.Id {
		return &pbtypes.Id{}
	}

	tests := []struct {
		name        string
		existingIds []*pbtypes.Id
		newId       *pbtypes.Id
		sizeLimit   int
		expected    bool
		description string
	}{
		// Zero ID objects tests
		{
			name:        "empty list, empty new ID, small limit",
			existingIds: []*pbtypes.Id{},
			newId:       createEmptyId(),
			sizeLimit:   10,
			expected:    true, // Empty ID has 0 size, so should be within limit
			description: "Adding empty ID to empty list should be within any reasonable limit",
		},
		{
			name:        "empty list, empty new ID, zero limit",
			existingIds: []*pbtypes.Id{},
			newId:       createEmptyId(),
			sizeLimit:   0,
			expected:    true, // Empty ID has 0 size
			description: "Empty ID should fit in zero limit",
		},
		{
			name:        "empty list, simple ID, zero limit",
			existingIds: []*pbtypes.Id{},
			newId:       createSimpleId("exec-1"),
			sizeLimit:   0,
			expected:    false, // Simple ID has size > 0, exceeds zero limit
			description: "Non-empty ID should not fit in zero limit",
		},

		// Within limits tests
		{
			name:        "empty list, simple ID, generous limit",
			existingIds: []*pbtypes.Id{},
			newId:       createSimpleId("exec-1"),
			sizeLimit:   1000,
			expected:    true,
			description: "Simple ID should fit in generous limit",
		},
		{
			name:        "one existing ID, add another simple ID, generous limit",
			existingIds: []*pbtypes.Id{createSimpleId("exec-1")},
			newId:       createSimpleId("exec-2"),
			sizeLimit:   1000,
			expected:    true,
			description: "Two simple IDs should fit in generous limit",
		},
		{
			name: "three existing IDs, add fourth, generous limit",
			existingIds: []*pbtypes.Id{
				createSimpleId("exec-1"),
				createSimpleId("exec-2"),
				createSimpleId("exec-3"),
			},
			newId:       createSimpleId("exec-4"),
			sizeLimit:   1000,
			expected:    true,
			description: "Four simple IDs should fit in generous limit",
		},

		// Above limits tests
		{
			name:        "empty list, simple ID, very small limit",
			existingIds: []*pbtypes.Id{},
			newId:       createSimpleId("exec-1"),
			sizeLimit:   1,
			expected:    false,
			description: "Simple ID should exceed very small limit",
		},
		{
			name:        "one existing ID, add large ID, small limit",
			existingIds: []*pbtypes.Id{createSimpleId("exec-1")},
			newId:       createLargeId("large"),
			sizeLimit:   100,
			expected:    false,
			description: "Large ID should exceed small limit when added to existing",
		},
		{
			name: "multiple existing IDs, add another, tight limit",
			existingIds: []*pbtypes.Id{
				createSimpleId("exec-1"),
				createSimpleId("exec-2"),
				createSimpleId("exec-3"),
			},
			newId:       createSimpleId("exec-4"),
			sizeLimit:   200, // Adjust based on actual size calculations
			expected:    false,
			description: "Multiple IDs should exceed tight limit",
		},

		// Edge cases
		{
			name:        "exactly at limit boundary",
			existingIds: []*pbtypes.Id{},
			newId:       createSimpleId("exec-1"),
			sizeLimit:   0, // Will be set to exact size in the test
			expected:    true,
			description: "ID exactly at limit should fit",
		},
		{
			name:        "one byte over limit",
			existingIds: []*pbtypes.Id{},
			newId:       createSimpleId("exec-1"),
			sizeLimit:   0, // Will be set to exact size - 1 in the test
			expected:    false,
			description: "ID one byte over limit should not fit",
		},
		{
			name:        "large ID alone",
			existingIds: []*pbtypes.Id{},
			newId:       createLargeId("huge"),
			sizeLimit:   50,
			expected:    false,
			description: "Large ID should exceed moderate limit",
		},
		{
			name: "mix of empty and non-empty existing IDs",
			existingIds: []*pbtypes.Id{
				createEmptyId(),
				createSimpleId("exec-1"),
				createEmptyId(),
			},
			newId:       createSimpleId("exec-2"),
			sizeLimit:   1000,
			expected:    true,
			description: "Mix of empty and non-empty IDs should work correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle special edge case tests that need dynamic size calculation
			sizeLimit := tt.sizeLimit
			if tt.name == "exactly at limit boundary" {
				// Calculate exact size needed for the new ID
				newIdSize := calculateIdSize(tt.newId)
				if newIdSize > 0 {
					tagSize := varintSize(uint64(1<<3 | 2))
					lengthSize := varintSize(uint64(newIdSize))
					sizeLimit = tagSize + lengthSize + newIdSize
				} else {
					sizeLimit = 0
				}
			} else if tt.name == "one byte over limit" {
				// Calculate exact size needed for the new ID minus 1
				newIdSize := calculateIdSize(tt.newId)
				if newIdSize > 0 {
					tagSize := varintSize(uint64(1<<3 | 2))
					lengthSize := varintSize(uint64(newIdSize))
					sizeLimit = tagSize + lengthSize + newIdSize - 1
				} else {
					sizeLimit = -1 // This would be impossible, but for test completeness
				}
			}

			currentSize := calculateQuerySize(tt.existingIds)
			result, _ := checkQuerySizeLimit(currentSize, tt.newId, sizeLimit)
			if result != tt.expected {
				// Provide detailed debugging information
				currentSize := calculateQuerySize(tt.existingIds)
				newIdSize := calculateIdSize(tt.newId)
				var totalSizeWithNewId int
				if newIdSize > 0 {
					totalSizeWithNewId = currentSize + varintSize(uint64(1<<3|2)) + varintSize(uint64(newIdSize)) + newIdSize
				} else {
					totalSizeWithNewId = currentSize
				}

				t.Errorf("%s: enough() = %v, expected %v\n"+
					"  Description: %s\n"+
					"  Current size: %d\n"+
					"  New ID size: %d\n"+
					"  Total size with new ID: %d\n"+
					"  Size limit: %d\n"+
					"  Would exceed: %v",
					tt.name, result, tt.expected,
					tt.description,
					currentSize, newIdSize, totalSizeWithNewId, sizeLimit,
					totalSizeWithNewId > sizeLimit)
			}
		})
	}
}

func TestCheckQuerySizeLimitWithRealSizes(t *testing.T) {
	// Test with realistic size calculations to verify our understanding
	simpleId := &pbtypes.Id{
		WorkflowExecutionId: "exec-123",
		WorkflowId:          "workflow-1",
		WorkflowOwner:       "owner",
		WorkflowName:        "test",
		ReportId:            "report-1",
		KeyId:               "key-1",
	}

	t.Run("verify size calculations", func(t *testing.T) {
		// Test empty list with simple ID
		size := calculateQuerySize([]*pbtypes.Id{})
		if size != 0 {
			t.Errorf("Empty list should have size 0, got %d", size)
		}

		// Test single ID
		singleIdSize := calculateQuerySize([]*pbtypes.Id{simpleId})
		if singleIdSize <= 0 {
			t.Errorf("Single ID should have positive size, got %d", singleIdSize)
		}

		t.Logf("Single ID size: %d bytes", singleIdSize)

		// Test that enough function works correctly with these sizes
		result, _ := checkQuerySizeLimit(0, simpleId, singleIdSize)
		if !result {
			t.Errorf("Should be able to add ID when limit equals exact size")
		}

		result, _ = checkQuerySizeLimit(0, simpleId, singleIdSize-1)
		if result {
			t.Errorf("Should not be able to add ID when limit is one byte less than size")
		}
	})

	t.Run("verify behavior with empty IDs", func(t *testing.T) {
		emptyId := &pbtypes.Id{}
		ids := []*pbtypes.Id{simpleId}
		currentSize := calculateQuerySize(ids)
		result, _ := checkQuerySizeLimit(currentSize, emptyId, 10000)
		if !result {
			t.Errorf("Should be able to add empty ID - it doesn't increase size")
		}
	})
}

func TestCheckQuerySizeLimitCaching(t *testing.T) {
	// Test that the caching mechanism works correctly
	id1 := &pbtypes.Id{WorkflowExecutionId: "exec-1", WorkflowId: "wf-1"}
	id2 := &pbtypes.Id{WorkflowExecutionId: "exec-2", WorkflowId: "wf-2"}
	id3 := &pbtypes.Id{WorkflowExecutionId: "exec-3", WorkflowId: "wf-3"}

	t.Run("incremental size calculation matches full recalculation", func(t *testing.T) {
		// Build up incrementally using caching
		cachedSize := 0
		ids := []*pbtypes.Id{}

		// Add first ID
		canAdd, newSize := checkQuerySizeLimit(cachedSize, id1, 10000)
		if !canAdd {
			t.Fatal("Should be able to add first ID")
		}
		ids = append(ids, id1)
		cachedSize = newSize

		// Verify cached size matches full calculation
		fullSize := calculateQuerySize(ids)
		if cachedSize != fullSize {
			t.Errorf("After adding id1: cached size %d != full calculation %d", cachedSize, fullSize)
		}

		// Add second ID
		canAdd, newSize = checkQuerySizeLimit(cachedSize, id2, 10000)
		if !canAdd {
			t.Fatal("Should be able to add second ID")
		}
		ids = append(ids, id2)
		cachedSize = newSize

		// Verify cached size matches full calculation
		fullSize = calculateQuerySize(ids)
		if cachedSize != fullSize {
			t.Errorf("After adding id2: cached size %d != full calculation %d", cachedSize, fullSize)
		}

		// Add third ID
		canAdd, newSize = checkQuerySizeLimit(cachedSize, id3, 10000)
		if !canAdd {
			t.Fatal("Should be able to add third ID")
		}
		ids = append(ids, id3)
		cachedSize = newSize

		// Verify final cached size matches full calculation
		fullSize = calculateQuerySize(ids)
		if cachedSize != fullSize {
			t.Errorf("After adding id3: cached size %d != full calculation %d", cachedSize, fullSize)
		}
	})

	t.Run("size limit enforcement with caching", func(t *testing.T) {
		// Calculate size of first two IDs
		twoIds := []*pbtypes.Id{id1, id2}
		twoIdsSize := calculateQuerySize(twoIds)

		// Set limit to exactly fit two IDs
		limit := twoIdsSize

		// Build incrementally
		cachedSize := 0

		// Add first ID
		canAdd, newSize := checkQuerySizeLimit(cachedSize, id1, limit)
		if !canAdd {
			t.Fatal("Should be able to add first ID within limit")
		}
		cachedSize = newSize

		// Add second ID
		canAdd, newSize = checkQuerySizeLimit(cachedSize, id2, limit)
		if !canAdd {
			t.Fatal("Should be able to add second ID within limit")
		}
		cachedSize = newSize

		// Try to add third ID - should fail
		canAdd, unchangedSize := checkQuerySizeLimit(cachedSize, id3, limit)
		if canAdd {
			t.Error("Should not be able to add third ID - would exceed limit")
		}
		if unchangedSize != cachedSize {
			t.Errorf("Size should remain unchanged when limit exceeded: got %d, expected %d", unchangedSize, cachedSize)
		}
	})

	t.Run("empty ID handling with caching", func(t *testing.T) {
		emptyId := &pbtypes.Id{}
		cachedSize := 0

		// Add empty ID - should not change size
		canAdd, newSize := checkQuerySizeLimit(cachedSize, emptyId, 1000)
		if !canAdd {
			t.Error("Should be able to add empty ID")
		}
		if newSize != cachedSize {
			t.Errorf("Empty ID should not change size: got %d, expected %d", newSize, cachedSize)
		}

		// Add real ID first
		canAdd, newSize = checkQuerySizeLimit(cachedSize, id1, 1000)
		if !canAdd {
			t.Fatal("Should be able to add real ID")
		}
		cachedSize = newSize

		// Add empty ID after real ID - should not change size
		canAdd, newSize = checkQuerySizeLimit(cachedSize, emptyId, 1000)
		if !canAdd {
			t.Error("Should be able to add empty ID after real ID")
		}
		if newSize != cachedSize {
			t.Errorf("Empty ID should not change size after real ID: got %d, expected %d", newSize, cachedSize)
		}
	})
}

func TestCheckQuerySizeLimitPerformance(t *testing.T) {
	// Performance test with many IDs
	ids := make([]*pbtypes.Id, 100)
	for i := 0; i < 100; i++ {
		ids[i] = &pbtypes.Id{
			WorkflowExecutionId: "exec-" + string(rune('A'+i%26)),
			WorkflowId:          "workflow-1",
			ReportId:            "report-1",
		}
	}

	newId := &pbtypes.Id{
		WorkflowExecutionId: "new-exec",
		WorkflowId:          "workflow-1",
		ReportId:            "report-1",
	}

	t.Run("performance with many IDs", func(t *testing.T) {
		currentSize := calculateQuerySize(ids)
		result, _ := checkQuerySizeLimit(currentSize, newId, 10000)
		// Just ensure it completes without error
		_ = result
	})
}

func TestCheckObservationSizeLimit(t *testing.T) {
	// Helper function to create a simple observation
	createSimpleObservation := func(workflowExecutionId string) *pbtypes.Observation {
		return &pbtypes.Observation{
			Id: &pbtypes.Id{
				WorkflowExecutionId: workflowExecutionId,
				WorkflowId:          "workflow-1",
				WorkflowOwner:       "owner",
				WorkflowName:        "test",
				ReportId:            "report-1",
				KeyId:               "key-1",
			},
			OverriddenEncoderName: "encoder-1",
		}
	}

	// Helper function to create a large observation
	createLargeObservation := func(suffix string) *pbtypes.Observation {
		return &pbtypes.Observation{
			Id: &pbtypes.Id{
				WorkflowExecutionId:      "very-long-workflow-execution-id-" + suffix,
				WorkflowId:               "very-long-workflow-id-" + suffix,
				WorkflowOwner:            "very-long-workflow-owner-" + suffix,
				WorkflowName:             "very-long-workflow-name-" + suffix,
				ReportId:                 "very-long-report-id-" + suffix,
				WorkflowDonId:            12345,
				WorkflowDonConfigVersion: 67890,
				KeyId:                    "very-long-key-id-" + suffix,
			},
			OverriddenEncoderName: "very-long-encoder-name-" + suffix,
			Observations: &pbvalues.List{
				Fields: []*pbvalues.Value{
					{Value: &pbvalues.Value_StringValue{StringValue: "observation-data-" + suffix}},
					{Value: &pbvalues.Value_Int64Value{Int64Value: 12345}},
				},
			},
			OverriddenEncoderConfig: &pbvalues.Map{
				Fields: map[string]*pbvalues.Value{
					"config-key-" + suffix: {Value: &pbvalues.Value_StringValue{StringValue: "config-value-" + suffix}},
				},
			},
		}
	}

	// Helper function to create an empty observation
	createEmptyObservation := func() *pbtypes.Observation {
		return &pbtypes.Observation{}
	}

	tests := []struct {
		name                 string
		existingObservations []*pbtypes.Observation
		newObservation       *pbtypes.Observation
		sizeLimit            int
		expected             bool
		description          string
	}{
		// Zero observation objects tests
		{
			name:                 "empty list, empty new observation, small limit",
			existingObservations: []*pbtypes.Observation{},
			newObservation:       createEmptyObservation(),
			sizeLimit:            10,
			expected:             true,
			description:          "Adding empty observation to empty list should be within any reasonable limit",
		},
		{
			name:                 "empty list, empty new observation, zero limit",
			existingObservations: []*pbtypes.Observation{},
			newObservation:       createEmptyObservation(),
			sizeLimit:            0,
			expected:             true,
			description:          "Empty observation should fit in zero limit",
		},
		{
			name:                 "empty list, simple observation, zero limit",
			existingObservations: []*pbtypes.Observation{},
			newObservation:       createSimpleObservation("exec-1"),
			sizeLimit:            0,
			expected:             false,
			description:          "Non-empty observation should not fit in zero limit",
		},

		// Within limits tests
		{
			name:                 "empty list, simple observation, generous limit",
			existingObservations: []*pbtypes.Observation{},
			newObservation:       createSimpleObservation("exec-1"),
			sizeLimit:            1000,
			expected:             true,
			description:          "Simple observation should fit in generous limit",
		},
		{
			name:                 "one existing observation, add another simple observation, generous limit",
			existingObservations: []*pbtypes.Observation{createSimpleObservation("exec-1")},
			newObservation:       createSimpleObservation("exec-2"),
			sizeLimit:            1000,
			expected:             true,
			description:          "Two simple observations should fit in generous limit",
		},

		// Above limits tests
		{
			name:                 "empty list, simple observation, very small limit",
			existingObservations: []*pbtypes.Observation{},
			newObservation:       createSimpleObservation("exec-1"),
			sizeLimit:            1,
			expected:             false,
			description:          "Simple observation should exceed very small limit",
		},
		{
			name:                 "one existing observation, add large observation, small limit",
			existingObservations: []*pbtypes.Observation{createSimpleObservation("exec-1")},
			newObservation:       createLargeObservation("large"),
			sizeLimit:            100,
			expected:             false,
			description:          "Large observation should exceed small limit when added to existing",
		},

		// Edge cases with complex values
		{
			name:                 "large observation alone",
			existingObservations: []*pbtypes.Observation{},
			newObservation:       createLargeObservation("huge"),
			sizeLimit:            50,
			expected:             false,
			description:          "Large observation with complex values should exceed moderate limit",
		},
		{
			name: "mix of empty and non-empty existing observations",
			existingObservations: []*pbtypes.Observation{
				createEmptyObservation(),
				createSimpleObservation("exec-1"),
				createEmptyObservation(),
			},
			newObservation: createSimpleObservation("exec-2"),
			sizeLimit:      1000,
			expected:       true,
			description:    "Mix of empty and non-empty observations should work correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentSize := calculateObservationsSize(tt.existingObservations)
			result, _ := checkObservationSizeLimit(currentSize, tt.newObservation, tt.sizeLimit)
			if result != tt.expected {
				// Provide detailed debugging information
				currentSize := calculateObservationsSize(tt.existingObservations)
				newObsSize := calculateObservationSize(tt.newObservation)
				var totalSizeWithNewObs int
				if newObsSize > 0 {
					totalSizeWithNewObs = currentSize + varintSize(uint64(1<<3|2)) + varintSize(uint64(newObsSize)) + newObsSize
				} else {
					totalSizeWithNewObs = currentSize
				}

				t.Errorf("%s: enoughObservation() = %v, expected %v\n"+
					"  Description: %s\n"+
					"  Current size: %d\n"+
					"  New observation size: %d\n"+
					"  Total size with new observation: %d\n"+
					"  Size limit: %d\n"+
					"  Would exceed: %v",
					tt.name, result, tt.expected,
					tt.description,
					currentSize, newObsSize, totalSizeWithNewObs, tt.sizeLimit,
					totalSizeWithNewObs > tt.sizeLimit)
			}
		})
	}
}

func TestCheckObservationSizeLimitWithRealSizes(t *testing.T) {
	// Test with realistic size calculations
	simpleObs := &pbtypes.Observation{
		Id: &pbtypes.Id{
			WorkflowExecutionId: "exec-123",
			WorkflowId:          "workflow-1",
			ReportId:            "report-1",
		},
		OverriddenEncoderName: "encoder-1",
	}

	t.Run("verify observation size calculations", func(t *testing.T) {
		// Test empty list
		size := calculateObservationsSize([]*pbtypes.Observation{})
		if size != 0 {
			t.Errorf("Empty observation list should have size 0, got %d", size)
		}

		// Test single observation
		singleObsSize := calculateObservationsSize([]*pbtypes.Observation{simpleObs})
		if singleObsSize <= 0 {
			t.Errorf("Single observation should have positive size, got %d", singleObsSize)
		}

		t.Logf("Single observation size: %d bytes", singleObsSize)

		// Test that enoughObservation function works correctly with these sizes
		result, _ := checkObservationSizeLimit(0, simpleObs, singleObsSize)
		if !result {
			t.Errorf("Should be able to add observation when limit equals exact size")
		}

		result, _ = checkObservationSizeLimit(0, simpleObs, singleObsSize-1)
		if result {
			t.Errorf("Should not be able to add observation when limit is one byte less than size")
		}
	})
}

func TestCheckObservationSizeLimitCaching(t *testing.T) {
	// Test that the caching mechanism works correctly for observations
	obs1 := &pbtypes.Observation{
		Id:                    &pbtypes.Id{WorkflowExecutionId: "exec-1", WorkflowId: "wf-1"},
		OverriddenEncoderName: "encoder-1",
	}
	obs2 := &pbtypes.Observation{
		Id:                    &pbtypes.Id{WorkflowExecutionId: "exec-2", WorkflowId: "wf-2"},
		OverriddenEncoderName: "encoder-2",
	}
	obs3 := &pbtypes.Observation{
		Id:                    &pbtypes.Id{WorkflowExecutionId: "exec-3", WorkflowId: "wf-3"},
		OverriddenEncoderName: "encoder-3",
	}

	t.Run("incremental size calculation matches full recalculation", func(t *testing.T) {
		// Build up incrementally using caching
		cachedSize := 0
		observations := []*pbtypes.Observation{}

		// Add first observation
		canAdd, newSize := checkObservationSizeLimit(cachedSize, obs1, 10000)
		if !canAdd {
			t.Fatal("Should be able to add first observation")
		}
		observations = append(observations, obs1)
		cachedSize = newSize

		// Verify cached size matches full calculation
		fullSize := calculateObservationsSize(observations)
		if cachedSize != fullSize {
			t.Errorf("After adding obs1: cached size %d != full calculation %d", cachedSize, fullSize)
		}

		// Add second observation
		canAdd, newSize = checkObservationSizeLimit(cachedSize, obs2, 10000)
		if !canAdd {
			t.Fatal("Should be able to add second observation")
		}
		observations = append(observations, obs2)
		cachedSize = newSize

		// Verify cached size matches full calculation
		fullSize = calculateObservationsSize(observations)
		if cachedSize != fullSize {
			t.Errorf("After adding obs2: cached size %d != full calculation %d", cachedSize, fullSize)
		}

		// Add third observation
		canAdd, newSize = checkObservationSizeLimit(cachedSize, obs3, 10000)
		if !canAdd {
			t.Fatal("Should be able to add third observation")
		}
		observations = append(observations, obs3)
		cachedSize = newSize

		// Verify final cached size matches full calculation
		fullSize = calculateObservationsSize(observations)
		if cachedSize != fullSize {
			t.Errorf("After adding obs3: cached size %d != full calculation %d", cachedSize, fullSize)
		}
	})

	t.Run("size limit enforcement with caching", func(t *testing.T) {
		// Calculate size of first two observations
		twoObs := []*pbtypes.Observation{obs1, obs2}
		twoObsSize := calculateObservationsSize(twoObs)

		// Set limit to exactly fit two observations
		limit := twoObsSize

		// Build incrementally
		cachedSize := 0

		// Add first observation
		canAdd, newSize := checkObservationSizeLimit(cachedSize, obs1, limit)
		if !canAdd {
			t.Fatal("Should be able to add first observation within limit")
		}
		cachedSize = newSize

		// Add second observation
		canAdd, newSize = checkObservationSizeLimit(cachedSize, obs2, limit)
		if !canAdd {
			t.Fatal("Should be able to add second observation within limit")
		}
		cachedSize = newSize

		// Try to add third observation - should fail
		canAdd, unchangedSize := checkObservationSizeLimit(cachedSize, obs3, limit)
		if canAdd {
			t.Error("Should not be able to add third observation - would exceed limit")
		}
		if unchangedSize != cachedSize {
			t.Errorf("Size should remain unchanged when limit exceeded: got %d, expected %d", unchangedSize, cachedSize)
		}
	})

	t.Run("empty observation handling with caching", func(t *testing.T) {
		emptyObs := &pbtypes.Observation{}
		cachedSize := 0

		// Add empty observation - should not change size
		canAdd, newSize := checkObservationSizeLimit(cachedSize, emptyObs, 1000)
		if !canAdd {
			t.Error("Should be able to add empty observation")
		}
		if newSize != cachedSize {
			t.Errorf("Empty observation should not change size: got %d, expected %d", newSize, cachedSize)
		}

		// Add real observation first
		canAdd, newSize = checkObservationSizeLimit(cachedSize, obs1, 1000)
		if !canAdd {
			t.Fatal("Should be able to add real observation")
		}
		cachedSize = newSize

		// Add empty observation after real observation - should not change size
		canAdd, newSize = checkObservationSizeLimit(cachedSize, emptyObs, 1000)
		if !canAdd {
			t.Error("Should be able to add empty observation after real observation")
		}
		if newSize != cachedSize {
			t.Errorf("Empty observation should not change size after real observation: got %d, expected %d", newSize, cachedSize)
		}
	})
}

func TestCheckObservationsSizeLimit(t *testing.T) {
	// Helper function to create a simple observations message
	createSimpleObservations := func(observationsList []*pbtypes.Observation, workflowIds []string) *pbtypes.Observations {
		return &pbtypes.Observations{
			Observations:          observationsList,
			RegisteredWorkflowIds: workflowIds,
			Timestamp:             timestamppb.New(time.Unix(1640995200, 0)), // Fixed timestamp for consistent testing
		}
	}

	// Helper function to create a simple observation
	createSimpleObservation := func(workflowExecutionId string) *pbtypes.Observation {
		return &pbtypes.Observation{
			Id: &pbtypes.Id{
				WorkflowExecutionId: workflowExecutionId,
				WorkflowId:          "workflow-1",
				WorkflowOwner:       "owner",
				WorkflowName:        "test",
				ReportId:            "report-1",
				KeyId:               "key-1",
			},
			OverriddenEncoderName: "encoder-1",
		}
	}

	// Helper function to create a large observation
	createLargeObservation := func(suffix string) *pbtypes.Observation {
		return &pbtypes.Observation{
			Id: &pbtypes.Id{
				WorkflowExecutionId:      "very-long-workflow-execution-id-" + suffix,
				WorkflowId:               "very-long-workflow-id-" + suffix,
				WorkflowOwner:            "very-long-workflow-owner-" + suffix,
				WorkflowName:             "very-long-workflow-name-" + suffix,
				ReportId:                 "very-long-report-id-" + suffix,
				WorkflowDonId:            12345,
				WorkflowDonConfigVersion: 67890,
				KeyId:                    "very-long-key-id-" + suffix,
			},
			OverriddenEncoderName: "very-long-encoder-name-" + suffix,
			Observations: &pbvalues.List{
				Fields: []*pbvalues.Value{
					{Value: &pbvalues.Value_StringValue{StringValue: "observation-data-" + suffix}},
					{Value: &pbvalues.Value_Int64Value{Int64Value: 12345}},
				},
			},
			OverriddenEncoderConfig: &pbvalues.Map{
				Fields: map[string]*pbvalues.Value{
					"config-key-" + suffix: {Value: &pbvalues.Value_StringValue{StringValue: "config-value-" + suffix}},
				},
			},
		}
	}

	// Helper function to create an empty observation
	createEmptyObservation := func() *pbtypes.Observation {
		return &pbtypes.Observation{}
	}

	tests := []struct {
		name                 string
		existingObservations *pbtypes.Observations
		newObservation       *pbtypes.Observation
		sizeLimit            int
		expected             bool
		description          string
	}{
		// Zero observation objects tests
		{
			name:                 "empty observations message, empty new observation, small limit",
			existingObservations: createSimpleObservations([]*pbtypes.Observation{}, []string{}),
			newObservation:       createEmptyObservation(),
			sizeLimit:            100,
			expected:             true,
			description:          "Adding empty observation to empty observations message should be within reasonable limit",
		},
		{
			name:                 "nil observations message, empty new observation, small limit",
			existingObservations: nil,
			newObservation:       createEmptyObservation(),
			sizeLimit:            10,
			expected:             true,
			description:          "Adding empty observation to nil observations should be within any limit",
		},
		{
			name:                 "empty observations message, simple observation, zero limit",
			existingObservations: createSimpleObservations([]*pbtypes.Observation{}, []string{}),
			newObservation:       createSimpleObservation("exec-1"),
			sizeLimit:            0,
			expected:             false,
			description:          "Non-empty observation should not fit in zero limit",
		},

		// Within limits tests
		{
			name:                 "empty observations message, simple observation, generous limit",
			existingObservations: createSimpleObservations([]*pbtypes.Observation{}, []string{"workflow-1"}),
			newObservation:       createSimpleObservation("exec-1"),
			sizeLimit:            1000,
			expected:             true,
			description:          "Simple observation should fit in generous limit",
		},
		{
			name:                 "observations with one existing observation, add another simple observation, generous limit",
			existingObservations: createSimpleObservations([]*pbtypes.Observation{createSimpleObservation("exec-1")}, []string{"workflow-1", "workflow-2"}),
			newObservation:       createSimpleObservation("exec-2"),
			sizeLimit:            1000,
			expected:             true,
			description:          "Two simple observations should fit in generous limit",
		},

		// Above limits tests
		{
			name:                 "empty observations message, simple observation, very small limit",
			existingObservations: createSimpleObservations([]*pbtypes.Observation{}, []string{}),
			newObservation:       createSimpleObservation("exec-1"),
			sizeLimit:            1,
			expected:             false,
			description:          "Simple observation should exceed very small limit",
		},
		{
			name:                 "observations with existing observation, add large observation, small limit",
			existingObservations: createSimpleObservations([]*pbtypes.Observation{createSimpleObservation("exec-1")}, []string{"workflow-1"}),
			newObservation:       createLargeObservation("large"),
			sizeLimit:            100,
			expected:             false,
			description:          "Large observation should exceed small limit when added to existing observations",
		},

		// Edge cases with complex observations messages
		{
			name:                 "large observation alone with many registered workflow IDs",
			existingObservations: createSimpleObservations([]*pbtypes.Observation{}, []string{"workflow-1", "workflow-2", "workflow-3", "very-long-workflow-name-for-testing"}),
			newObservation:       createLargeObservation("huge"),
			sizeLimit:            100,
			expected:             false,
			description:          "Large observation with many registered workflow IDs should exceed moderate limit",
		},
		{
			name: "mix of empty and non-empty existing observations in observations message",
			existingObservations: createSimpleObservations([]*pbtypes.Observation{
				createEmptyObservation(),
				createSimpleObservation("exec-1"),
				createEmptyObservation(),
			}, []string{"workflow-1"}),
			newObservation: createSimpleObservation("exec-2"),
			sizeLimit:      1000,
			expected:       true,
			description:    "Mix of empty and non-empty observations in observations message should work correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentSize := calculateObservationsMessageSize(tt.existingObservations)
			result, _ := checkObservationsSizeLimit(currentSize, tt.newObservation, tt.sizeLimit)
			if result != tt.expected {
				// Provide detailed debugging information
				currentSize := calculateObservationsMessageSize(tt.existingObservations)
				newObsSize := calculateObservationSize(tt.newObservation)
				var totalSizeWithNewObs int
				if newObsSize > 0 {
					totalSizeWithNewObs = currentSize + varintSize(uint64(1<<3|2)) + varintSize(uint64(newObsSize)) + newObsSize
				} else {
					totalSizeWithNewObs = currentSize
				}

				t.Errorf("%s: enoughObservations() = %v, expected %v\n"+
					"  Description: %s\n"+
					"  Current size: %d\n"+
					"  New observation size: %d\n"+
					"  Total size with new observation: %d\n"+
					"  Size limit: %d\n"+
					"  Would exceed: %v",
					tt.name, result, tt.expected,
					tt.description,
					currentSize, newObsSize, totalSizeWithNewObs, tt.sizeLimit,
					totalSizeWithNewObs > tt.sizeLimit)
			}
		})
	}
}

func TestCheckObservationsSizeLimitWithRealSizes(t *testing.T) {
	// Test with realistic size calculations
	simpleObs := &pbtypes.Observation{
		Id: &pbtypes.Id{
			WorkflowExecutionId: "exec-123",
			WorkflowId:          "workflow-1",
			ReportId:            "report-1",
		},
		OverriddenEncoderName: "encoder-1",
	}

	observationsMsg := &pbtypes.Observations{
		Observations:          []*pbtypes.Observation{},
		RegisteredWorkflowIds: []string{"workflow-1", "workflow-2"},
		Timestamp:             timestamppb.New(time.Unix(1640995200, 0)),
	}

	t.Run("verify observations message size calculations", func(t *testing.T) {
		// Test empty observations message
		size := calculateObservationsMessageSize(observationsMsg)
		if size <= 0 {
			t.Errorf("Observations message with workflow IDs and timestamp should have positive size, got %d", size)
		}

		t.Logf("Empty observations message size: %d bytes", size)

		// Test adding observation
		currentSize := calculateObservationsMessageSize(observationsMsg)
		result, _ := checkObservationsSizeLimit(currentSize, simpleObs, size+100)
		if !result {
			t.Errorf("Should be able to add observation when limit has buffer")
		}

		// Calculate actual size with observation
		observationsWithObs := &pbtypes.Observations{
			Observations:          []*pbtypes.Observation{simpleObs},
			RegisteredWorkflowIds: []string{"workflow-1", "workflow-2"},
			Timestamp:             timestamppb.New(time.Unix(1640995200, 0)),
		}
		sizeWithObs := calculateObservationsMessageSize(observationsWithObs)

		t.Logf("Observations message with one observation size: %d bytes", sizeWithObs)

		currentSize = calculateObservationsMessageSize(observationsMsg)
		result, _ = checkObservationsSizeLimit(currentSize, simpleObs, sizeWithObs-1)
		if result {
			t.Errorf("Should not be able to add observation when limit is one byte less than required")
		}
	})
}

func TestQuerySizeCalculationMatchesRealMarshalling(t *testing.T) {
	// Helper function to create a simple ID (reused from existing tests)
	createSimpleId := func(workflowExecutionId string) *pbtypes.Id {
		return &pbtypes.Id{
			WorkflowExecutionId: workflowExecutionId,
			WorkflowId:          "workflow-1",
			WorkflowOwner:       "owner",
			WorkflowName:        "test",
			ReportId:            "report-1",
			KeyId:               "key-1",
		}
	}

	// Create test data using existing helper
	ids := []*pbtypes.Id{
		createSimpleId("exec-1"),
		createSimpleId("exec-2"),
		createSimpleId("exec-3"),
	}

	// Calculate size using our function
	calculatedSize := calculateQuerySize(ids)

	// Create actual Query message and marshal it
	query := &pbtypes.Query{Ids: ids}
	marshalled, err := proto.MarshalOptions{Deterministic: true}.Marshal(query)
	if err != nil {
		t.Fatalf("Failed to marshal query: %v", err)
	}
	actualSize := len(marshalled)

	// Verify they match
	if calculatedSize != actualSize {
		t.Errorf("Query size calculation mismatch: calculated=%d, actual=%d", calculatedSize, actualSize)
	}

	t.Logf("Query size calculation matches: %d bytes", actualSize)
}

func TestObservationsSizeCalculationMatchesRealMarshalling(t *testing.T) {
	// Helper function to create a simple observation (reused from existing tests)
	createSimpleObservation := func(workflowExecutionId string) *pbtypes.Observation {
		return &pbtypes.Observation{
			Id: &pbtypes.Id{
				WorkflowExecutionId: workflowExecutionId,
				WorkflowId:          "workflow-1",
				WorkflowOwner:       "owner",
				WorkflowName:        "test",
				ReportId:            "report-1",
				KeyId:               "key-1",
			},
			OverriddenEncoderName: "encoder-1",
		}
	}

	// Create test data using existing helpers
	observations := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			createSimpleObservation("exec-1"),
			createSimpleObservation("exec-2"),
		},
		RegisteredWorkflowIds: []string{"workflow-1", "workflow-2"},
		Timestamp:             timestamppb.New(time.Unix(1640995200, 0)),
	}

	// Calculate size using our function
	calculatedSize := calculateObservationsMessageSize(observations)

	// Marshal the actual message
	marshalled, err := proto.MarshalOptions{Deterministic: true}.Marshal(observations)
	if err != nil {
		t.Fatalf("Failed to marshal observations: %v", err)
	}
	actualSize := len(marshalled)

	// Verify they match
	if calculatedSize != actualSize {
		t.Errorf("Observations size calculation mismatch: calculated=%d, actual=%d", calculatedSize, actualSize)
	}

	t.Logf("Observations size calculation matches: %d bytes", actualSize)
}

func TestReportSizeCalculationMatchesRealMarshalling(t *testing.T) {
	// Helper function to create a simple report with realistic data
	createSimpleReport := func(workflowExecutionId string) *pbtypes.Report {
		return &pbtypes.Report{
			Id: &pbtypes.Id{
				WorkflowExecutionId: workflowExecutionId,
				WorkflowId:          "workflow-1",
				WorkflowOwner:       "owner",
				WorkflowName:        "test",
				ReportId:            "report-1",
				KeyId:               "key-1",
			},
			Outcome: &pbtypes.AggregationOutcome{
				EncodableOutcome: &pbvalues.Map{
					Fields: map[string]*pbvalues.Value{
						"result": {
							Value: &pbvalues.Value_StringValue{StringValue: "success"},
						},
					},
				},
				Metadata:     []byte("test-metadata"),
				ShouldReport: true,
				LastSeenAt:   12345,
				Timestamp:    timestamppb.New(time.Unix(1640995200, 0)),
				EncoderName:  "test-encoder",
			},
		}
	}

	// Create test report
	report := createSimpleReport("exec-1")

	// Calculate size using our function
	calculatedSize := calculateReportSize(report)

	// Marshal the actual message
	marshalled, err := proto.MarshalOptions{Deterministic: true}.Marshal(report)
	if err != nil {
		t.Fatalf("Failed to marshal report: %v", err)
	}
	actualSize := len(marshalled)

	// Verify they match
	if calculatedSize != actualSize {
		t.Errorf("Report size calculation mismatch: calculated=%d, actual=%d", calculatedSize, actualSize)
	}

	t.Logf("Report size calculation matches: %d bytes", actualSize)
}

func TestCheckReportSizeLimit(t *testing.T) {
	// Helper function to create a simple report with predictable size
	createSimpleReport := func(workflowExecutionId string) *pbtypes.Report {
		return &pbtypes.Report{
			Id: &pbtypes.Id{
				WorkflowExecutionId: workflowExecutionId,
				WorkflowId:          "workflow-1",
				WorkflowOwner:       "owner",
				WorkflowName:        "test",
				ReportId:            "report-1",
				KeyId:               "key-1",
			},
			Outcome: &pbtypes.AggregationOutcome{
				EncodableOutcome: &pbvalues.Map{
					Fields: map[string]*pbvalues.Value{
						"result": {
							Value: &pbvalues.Value_StringValue{StringValue: "success"},
						},
					},
				},
				Metadata:     []byte("metadata"),
				ShouldReport: true,
				LastSeenAt:   12345,
				EncoderName:  "encoder",
			},
		}
	}

	// Helper function to create a report with all fields populated for larger size
	createLargeReport := func(suffix string) *pbtypes.Report {
		return &pbtypes.Report{
			Id: &pbtypes.Id{
				WorkflowExecutionId:      "very-long-workflow-execution-id-" + suffix,
				WorkflowId:               "very-long-workflow-id-" + suffix,
				WorkflowOwner:            "very-long-workflow-owner-" + suffix,
				WorkflowName:             "very-long-workflow-name-" + suffix,
				ReportId:                 "very-long-report-id-" + suffix,
				WorkflowDonId:            12345,
				WorkflowDonConfigVersion: 67890,
				KeyId:                    "very-long-key-id-" + suffix,
			},
			Outcome: &pbtypes.AggregationOutcome{
				EncodableOutcome: &pbvalues.Map{
					Fields: map[string]*pbvalues.Value{
						"very-long-result-key-" + suffix: {
							Value: &pbvalues.Value_StringValue{StringValue: "very-long-result-value-" + suffix},
						},
						"another-long-key-" + suffix: {
							Value: &pbvalues.Value_StringValue{StringValue: "another-long-value-" + suffix},
						},
					},
				},
				Metadata:     []byte("very-long-metadata-content-for-testing-" + suffix),
				ShouldReport: true,
				LastSeenAt:   123456789,
				Timestamp:    timestamppb.New(time.Unix(1640995200, 0)),
				EncoderName:  "very-long-encoder-name-" + suffix,
				EncoderConfig: &pbvalues.Map{
					Fields: map[string]*pbvalues.Value{
						"config-key-" + suffix: {
							Value: &pbvalues.Value_StringValue{StringValue: "config-value-" + suffix},
						},
					},
				},
			},
		}
	}

	// Helper function to create an empty report (zero values)
	createEmptyReport := func() *pbtypes.Report {
		return &pbtypes.Report{}
	}

	tests := []struct {
		name            string
		existingReports []*pbtypes.Report
		newReport       *pbtypes.Report
		sizeLimit       int
		expected        bool
		description     string
	}{
		// Zero report objects tests
		{
			name:            "empty list, empty new report, small limit",
			existingReports: []*pbtypes.Report{},
			newReport:       createEmptyReport(),
			sizeLimit:       10,
			expected:        true, // Empty report has 0 size, so should be within limit
			description:     "Adding empty report to empty list should be within any reasonable limit",
		},
		{
			name:            "empty list, empty new report, zero limit",
			existingReports: []*pbtypes.Report{},
			newReport:       createEmptyReport(),
			sizeLimit:       0,
			expected:        true, // Empty report has 0 size
			description:     "Empty report should fit in zero limit",
		},
		{
			name:            "empty list, simple report, zero limit",
			existingReports: []*pbtypes.Report{},
			newReport:       createSimpleReport("exec-1"),
			sizeLimit:       0,
			expected:        false, // Simple report has size > 0, exceeds zero limit
			description:     "Non-empty report should not fit in zero limit",
		},

		// Within limits tests
		{
			name:            "empty list, simple report, generous limit",
			existingReports: []*pbtypes.Report{},
			newReport:       createSimpleReport("exec-1"),
			sizeLimit:       2000,
			expected:        true,
			description:     "Simple report should fit in generous limit",
		},
		{
			name:            "one existing report, add another simple report, generous limit",
			existingReports: []*pbtypes.Report{createSimpleReport("exec-1")},
			newReport:       createSimpleReport("exec-2"),
			sizeLimit:       2000,
			expected:        true,
			description:     "Two simple reports should fit in generous limit",
		},
		{
			name: "three existing reports, add fourth, generous limit",
			existingReports: []*pbtypes.Report{
				createSimpleReport("exec-1"),
				createSimpleReport("exec-2"),
				createSimpleReport("exec-3"),
			},
			newReport:   createSimpleReport("exec-4"),
			sizeLimit:   2000,
			expected:    true,
			description: "Four simple reports should fit in generous limit",
		},

		// Above limits tests
		{
			name:            "empty list, simple report, very small limit",
			existingReports: []*pbtypes.Report{},
			newReport:       createSimpleReport("exec-1"),
			sizeLimit:       1,
			expected:        false,
			description:     "Simple report should exceed very small limit",
		},
		{
			name:            "one existing report, add large report, small limit",
			existingReports: []*pbtypes.Report{createSimpleReport("exec-1")},
			newReport:       createLargeReport("large"),
			sizeLimit:       200,
			expected:        false,
			description:     "Large report should exceed small limit when added to existing",
		},
		{
			name: "multiple existing reports, add another, tight limit",
			existingReports: []*pbtypes.Report{
				createSimpleReport("exec-1"),
				createSimpleReport("exec-2"),
				createSimpleReport("exec-3"),
			},
			newReport:   createSimpleReport("exec-4"),
			sizeLimit:   400, // Adjust based on actual size calculations
			expected:    false,
			description: "Multiple reports should exceed tight limit",
		},

		// Edge cases
		{
			name:            "exactly at limit boundary",
			existingReports: []*pbtypes.Report{},
			newReport:       createSimpleReport("exec-1"),
			sizeLimit:       0, // Will be set to exact size in the test
			expected:        true,
			description:     "Report exactly at limit should fit",
		},
		{
			name:            "one byte over limit",
			existingReports: []*pbtypes.Report{},
			newReport:       createSimpleReport("exec-1"),
			sizeLimit:       0, // Will be set to exact size - 1 in the test
			expected:        false,
			description:     "Report one byte over limit should not fit",
		},
		{
			name:            "large report alone",
			existingReports: []*pbtypes.Report{},
			newReport:       createLargeReport("huge"),
			sizeLimit:       100,
			expected:        false,
			description:     "Large report should exceed moderate limit",
		},
		{
			name: "mix of empty and non-empty existing reports",
			existingReports: []*pbtypes.Report{
				createEmptyReport(),
				createSimpleReport("exec-1"),
				createEmptyReport(),
			},
			newReport:   createSimpleReport("exec-2"),
			sizeLimit:   2000,
			expected:    true,
			description: "Mix of empty and non-empty reports should work correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle special edge case tests that need dynamic size calculation
			sizeLimit := tt.sizeLimit
			if tt.name == "exactly at limit boundary" {
				// Calculate exact size needed for the new report
				newReportSize := calculateReportSize(tt.newReport)
				if newReportSize > 0 {
					tagSize := varintSize(uint64(2<<3 | 2))
					lengthSize := varintSize(uint64(newReportSize))
					sizeLimit = tagSize + lengthSize + newReportSize
				} else {
					sizeLimit = 0
				}
			} else if tt.name == "one byte over limit" {
				// Calculate exact size needed for the new report minus 1
				newReportSize := calculateReportSize(tt.newReport)
				if newReportSize > 0 {
					tagSize := varintSize(uint64(2<<3 | 2))
					lengthSize := varintSize(uint64(newReportSize))
					sizeLimit = tagSize + lengthSize + newReportSize - 1
				} else {
					sizeLimit = -1 // This would be impossible, but for test completeness
				}
			}

			currentSize := calculateReportsSize(tt.existingReports)
			result, _ := checkReportSizeLimit(currentSize, tt.newReport, sizeLimit)
			if result != tt.expected {
				// Provide detailed debugging information
				currentSize := calculateReportsSize(tt.existingReports)
				newReportSize := calculateReportSize(tt.newReport)
				var totalSizeWithNewReport int
				if newReportSize > 0 {
					totalSizeWithNewReport = currentSize + varintSize(uint64(2<<3|2)) + varintSize(uint64(newReportSize)) + newReportSize
				} else {
					totalSizeWithNewReport = currentSize
				}

				t.Errorf("%s: CheckReportSizeLimit() = %v, expected %v\n"+
					"  Description: %s\n"+
					"  Current size: %d\n"+
					"  New report size: %d\n"+
					"  Total size with new report: %d\n"+
					"  Size limit: %d\n"+
					"  Would exceed: %v",
					tt.name, result, tt.expected,
					tt.description,
					currentSize, newReportSize, totalSizeWithNewReport, sizeLimit,
					totalSizeWithNewReport > sizeLimit)
			}
		})
	}
}

func TestCheckReportSizeLimitWithRealSizes(t *testing.T) {
	// Test with realistic size calculations to verify our understanding
	simpleReport := &pbtypes.Report{
		Id: &pbtypes.Id{
			WorkflowExecutionId: "exec-123",
			WorkflowId:          "workflow-1",
			WorkflowOwner:       "owner",
			WorkflowName:        "test",
			ReportId:            "report-1",
			KeyId:               "key-1",
		},
		Outcome: &pbtypes.AggregationOutcome{
			EncodableOutcome: &pbvalues.Map{
				Fields: map[string]*pbvalues.Value{
					"result": {
						Value: &pbvalues.Value_StringValue{StringValue: "success"},
					},
				},
			},
			Metadata:     []byte("metadata"),
			ShouldReport: true,
			LastSeenAt:   12345,
			EncoderName:  "encoder",
		},
	}

	// Helper function to create an outcome with reports
	createOutcomeWithReports := func(reports []*pbtypes.Report) *pbtypes.Outcome {
		return &pbtypes.Outcome{
			Outcomes:       map[string]*pbtypes.AggregationOutcome{},
			CurrentReports: reports,
		}
	}

	t.Run("verify size calculations", func(t *testing.T) {
		// Test empty list
		emptyOutcome := createOutcomeWithReports([]*pbtypes.Report{})
		size := calculateReportsSize(emptyOutcome.CurrentReports)
		if size != 0 {
			t.Errorf("Empty list should have size 0, got %d", size)
		}

		// Test single report
		singleReportOutcome := createOutcomeWithReports([]*pbtypes.Report{simpleReport})
		singleReportSize := calculateReportsSize(singleReportOutcome.CurrentReports)
		if singleReportSize <= 0 {
			t.Errorf("Single report should have positive size, got %d", singleReportSize)
		}

		t.Logf("Single report size: %d bytes", singleReportSize)

		// Test that size limit function works correctly with these sizes
		result, _ := checkReportSizeLimit(0, simpleReport, singleReportSize)
		if !result {
			t.Errorf("Should be able to add report when limit equals exact size")
		}

		result, _ = checkReportSizeLimit(0, simpleReport, singleReportSize-1)
		if result {
			t.Errorf("Should not be able to add report when limit is one byte less than size")
		}
	})
}

func TestCheckReportSizeLimitCaching(t *testing.T) {
	// Test that the caching mechanism works correctly for reports
	report1 := &pbtypes.Report{
		Id: &pbtypes.Id{WorkflowExecutionId: "exec-1", WorkflowId: "wf-1"},
		Outcome: &pbtypes.AggregationOutcome{
			EncodableOutcome: &pbvalues.Map{
				Fields: map[string]*pbvalues.Value{
					"result": {Value: &pbvalues.Value_StringValue{StringValue: "result1"}},
				},
			},
		},
	}
	report2 := &pbtypes.Report{
		Id: &pbtypes.Id{WorkflowExecutionId: "exec-2", WorkflowId: "wf-2"},
		Outcome: &pbtypes.AggregationOutcome{
			EncodableOutcome: &pbvalues.Map{
				Fields: map[string]*pbvalues.Value{
					"result": {Value: &pbvalues.Value_StringValue{StringValue: "result2"}},
				},
			},
		},
	}
	report3 := &pbtypes.Report{
		Id: &pbtypes.Id{WorkflowExecutionId: "exec-3", WorkflowId: "wf-3"},
		Outcome: &pbtypes.AggregationOutcome{
			EncodableOutcome: &pbvalues.Map{
				Fields: map[string]*pbvalues.Value{
					"result": {Value: &pbvalues.Value_StringValue{StringValue: "result3"}},
				},
			},
		},
	}

	t.Run("incremental size calculation matches full recalculation", func(t *testing.T) {
		// Build up incrementally using caching
		cachedSize := 0
		reports := []*pbtypes.Report{}

		// Add first report
		canAdd, newSize := checkReportSizeLimit(cachedSize, report1, 10000)
		if !canAdd {
			t.Fatal("Should be able to add first report")
		}
		reports = append(reports, report1)
		cachedSize = newSize

		// Verify cached size matches full calculation
		fullSize := calculateReportsSize(reports)
		if cachedSize != fullSize {
			t.Errorf("After adding report1: cached size %d != full calculation %d", cachedSize, fullSize)
		}

		// Add second report
		canAdd, newSize = checkReportSizeLimit(cachedSize, report2, 10000)
		if !canAdd {
			t.Fatal("Should be able to add second report")
		}
		reports = append(reports, report2)
		cachedSize = newSize

		// Verify cached size matches full calculation
		fullSize = calculateReportsSize(reports)
		if cachedSize != fullSize {
			t.Errorf("After adding report2: cached size %d != full calculation %d", cachedSize, fullSize)
		}

		// Add third report
		canAdd, newSize = checkReportSizeLimit(cachedSize, report3, 10000)
		if !canAdd {
			t.Fatal("Should be able to add third report")
		}
		reports = append(reports, report3)
		cachedSize = newSize

		// Verify final cached size matches full calculation
		fullSize = calculateReportsSize(reports)
		if cachedSize != fullSize {
			t.Errorf("After adding report3: cached size %d != full calculation %d", cachedSize, fullSize)
		}
	})

	t.Run("size limit enforcement with caching", func(t *testing.T) {
		// Calculate size of first two reports
		twoReports := []*pbtypes.Report{report1, report2}
		twoReportsSize := calculateReportsSize(twoReports)

		// Set limit to exactly fit two reports
		limit := twoReportsSize

		// Build incrementally
		cachedSize := 0

		// Add first report
		canAdd, newSize := checkReportSizeLimit(cachedSize, report1, limit)
		if !canAdd {
			t.Fatal("Should be able to add first report within limit")
		}
		cachedSize = newSize

		// Add second report
		canAdd, newSize = checkReportSizeLimit(cachedSize, report2, limit)
		if !canAdd {
			t.Fatal("Should be able to add second report within limit")
		}
		cachedSize = newSize

		// Try to add third report - should fail
		canAdd, unchangedSize := checkReportSizeLimit(cachedSize, report3, limit)
		if canAdd {
			t.Error("Should not be able to add third report - would exceed limit")
		}
		if unchangedSize != cachedSize {
			t.Errorf("Size should remain unchanged when limit exceeded: got %d, expected %d", unchangedSize, cachedSize)
		}
	})

	t.Run("nil report handling with caching", func(t *testing.T) {
		cachedSize := 100 // Some initial size

		// Add nil report - should not change size and should return true
		canAdd, newSize := checkReportSizeLimit(cachedSize, nil, 1000)
		if !canAdd {
			t.Error("Should be able to add nil report")
		}
		if newSize != cachedSize {
			t.Errorf("Nil report should not change size: got %d, expected %d", newSize, cachedSize)
		}
	})

	t.Run("empty report handling with caching", func(t *testing.T) {
		emptyReport := &pbtypes.Report{}
		cachedSize := 0

		// Add empty report - should not change size
		canAdd, newSize := checkReportSizeLimit(cachedSize, emptyReport, 1000)
		if !canAdd {
			t.Error("Should be able to add empty report")
		}
		if newSize != cachedSize {
			t.Errorf("Empty report should not change size: got %d, expected %d", newSize, cachedSize)
		}

		// Add real report first
		canAdd, newSize = checkReportSizeLimit(cachedSize, report1, 1000)
		if !canAdd {
			t.Fatal("Should be able to add real report")
		}
		cachedSize = newSize

		// Add empty report after real report - should not change size
		canAdd, newSize = checkReportSizeLimit(cachedSize, emptyReport, 1000)
		if !canAdd {
			t.Error("Should be able to add empty report after real report")
		}
		if newSize != cachedSize {
			t.Errorf("Empty report should not change size after real report: got %d, expected %d", newSize, cachedSize)
		}
	})
}
