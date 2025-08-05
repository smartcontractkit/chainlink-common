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

			result := CheckQuerySizeLimit(tt.existingIds, tt.newId, sizeLimit)
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
		result := CheckQuerySizeLimit([]*pbtypes.Id{}, simpleId, singleIdSize)
		if !result {
			t.Errorf("Should be able to add ID when limit equals exact size")
		}

		result = CheckQuerySizeLimit([]*pbtypes.Id{}, simpleId, singleIdSize-1)
		if result {
			t.Errorf("Should not be able to add ID when limit is one byte less than size")
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
		result := CheckQuerySizeLimit(ids, newId, 10000)
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
			result := CheckObservationSizeLimit(tt.existingObservations, tt.newObservation, tt.sizeLimit)
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
		result := CheckObservationSizeLimit([]*pbtypes.Observation{}, simpleObs, singleObsSize)
		if !result {
			t.Errorf("Should be able to add observation when limit equals exact size")
		}

		result = CheckObservationSizeLimit([]*pbtypes.Observation{}, simpleObs, singleObsSize-1)
		if result {
			t.Errorf("Should not be able to add observation when limit is one byte less than size")
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
			result := CheckObservationsSizeLimit(tt.existingObservations, tt.newObservation, tt.sizeLimit)
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
		result := CheckObservationsSizeLimit(observationsMsg, simpleObs, size+100)
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

		result = CheckObservationsSizeLimit(observationsMsg, simpleObs, sizeWithObs-1)
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
