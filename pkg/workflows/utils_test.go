package workflows

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
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
