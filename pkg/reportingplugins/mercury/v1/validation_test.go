package mercury_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateValidFromTimestamp(t *testing.T) {
	t.Run("succeeds when observationTimestamp is > validFromTimestamp", func(t *testing.T) {
		paos := NewValidParsedAttributedObservations()
		err := ValidateValidFromTimestamp(paos)
		assert.NoError(t, err)
	})
	t.Run("fails when observationTimestamp is <= validFromTimestamp", func(t *testing.T) {
		paos := NewInvalidParsedAttributedObservations()
		err := ValidateValidFromTimestamp(paos)
		assert.EqualError(t, err, "observationTimestamp (2) must be > validFromTimestamp (1679648456)")
	})
}
