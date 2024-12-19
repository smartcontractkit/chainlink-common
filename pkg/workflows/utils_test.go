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
	got, err := GenerateWorkflowIDFromStrings(owner, []byte("workflow"), []byte("config"), "http://mysecrets.com")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// Without prefix
	owner = "26729408f179371be6433b9585d8427f121bfe82"
	got, err = GenerateWorkflowIDFromStrings(owner, []byte("workflow"), []byte("config"), "http://mysecrets.com")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// Very short; empty but with a prefix
	owner = "0x"
	got, err = GenerateWorkflowIDFromStrings(owner, []byte("workflow"), []byte("config"), "http://mysecrets.com")
	require.NoError(t, err)
	assert.NotNil(t, got)

	owner = "invalid"
	_, err = GenerateWorkflowIDFromStrings(owner, []byte("workflow"), []byte("config"), "http://mysecrets.com")
	assert.ErrorContains(t, err, "encoding/hex")
}
