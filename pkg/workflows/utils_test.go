package workflows

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_EncodeExecutionID(t *testing.T) {
	var (
		workflowID = "workflowID"
		eventID    = "eventID"
		s          = sha256.New()
	)

	_, err := s.Write([]byte(workflowID))
	assert.NoError(t, err)

	_, err = s.Write([]byte(eventID))
	assert.NoError(t, err)

	expected := hex.EncodeToString(s.Sum(nil))
	actual, err := EncodeExecutionID(workflowID, eventID)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	// Test ordering
	s.Reset()
	_, err = s.Write([]byte(eventID))
	assert.NoError(t, err)

	_, err = s.Write([]byte(workflowID))
	assert.NoError(t, err)

	reversed := hex.EncodeToString(s.Sum(nil))
	assert.NotEqual(t, reversed, actual)
}

func Test_GenerateWorkflowIDFromStrings(t *testing.T) {
	// With prefix
	owner := "0x26729408f179371be6433b9585d8427f121bfe82"
	got, err := GenerateWorkflowIDFromStrings(owner, "porporpore", []byte("workflow"), []byte("config"), "http://mysecrets.com")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// Always starts with the version byte
	assert.Equal(t, got[:2], hex.EncodeToString([]byte{versionByte}))

	// Without prefix
	owner = "26729408f179371be6433b9585d8427f121bfe82"
	got, err = GenerateWorkflowIDFromStrings(owner, "porporpore", []byte("workflow"), []byte("config"), "http://mysecrets.com")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// Very short; empty but with a prefix
	owner = "0x"
	got, err = GenerateWorkflowIDFromStrings(owner, "porporpore", []byte("workflow"), []byte("config"), "http://mysecrets.com")
	require.NoError(t, err)
	assert.NotNil(t, got)

	owner = "invalid"
	_, err = GenerateWorkflowIDFromStrings(owner, "porporpore", []byte("workflow"), []byte("config"), "http://mysecrets.com")
	assert.ErrorContains(t, err, "encoding/hex")
}

func TestNormalizeWorkflowName(t *testing.T) {
	tt := []struct {
		input    string
		expected [10]byte
	}{
		{
			input:    "Hello, world!",
			expected: [10]byte{0x31, 0x5f, 0x5b, 0xdb, 0x76, 0xd0, 0x78, 0xc4, 0x3b, 0x8a},
		},
		{
			input:    "My Incredible Workflow Name",
			expected: [10]byte{0x84, 0x00, 0x2e, 0xb9, 0xe2, 0xa0, 0x6b, 0x09, 0x97, 0x7c},
		},
		{
			input:    "You either die a hero, or live long enough to see yourself become the villain.",
			expected: [10]byte{0x6b, 0xa1, 0xf7, 0xa6, 0xa0, 0x91, 0x95, 0x1a, 0x2d, 0xd2},
		},
	}

	for _, tc := range tt {
		t.Run(tc.input, func(t *testing.T) {
			// Call the function with the test input
			result := HashTruncateName(tc.input)

			// Assert that the result is exactly the expected output
			require.Equal(t, tc.expected, result)

			// Assert that the result is 10 bytes long
			require.Len(t, result, 10)
		})
	}
}
