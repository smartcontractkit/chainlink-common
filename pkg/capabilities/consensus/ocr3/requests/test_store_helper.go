package requests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestStore(t *testing.T, count int) *Store {
	s := NewStore()
	n := time.Now()

	for i := 0; i < count; i++ {
		err := s.Add(&Request{WorkflowExecutionID: uuid.New().String(), ExpiresAt: n.Add(1 * time.Hour)})
		require.NoError(t, err)
	}

	return s
}
