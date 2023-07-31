package mercury_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateValidFromTimestamp(t *testing.T) {
	t.Run("succeeds when observationTimestamp is >= validFromTimestamp", func(t *testing.T) {
		err := ValidateValidFromTimestamp(456, 123)
		assert.NoError(t, err)
	})
	t.Run("fails when observationTimestamp is < validFromTimestamp", func(t *testing.T) {
		err := ValidateValidFromTimestamp(111, 112)
		assert.EqualError(t, err, "observationTimestamp (111) must be >= validFromTimestamp (112)")
	})
}
