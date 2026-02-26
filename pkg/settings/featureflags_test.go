package settings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFeatureFlag_Active(t *testing.T) {
	activateAt := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	flag := NewFeatureFlag(Time(activateAt), nil)

	t.Run("zero timestamp is never active", func(t *testing.T) {
		assert.False(t, flag.IsActive(t.Context(), 0))
	})

	t.Run("negative timestamp is never active", func(t *testing.T) {
		assert.False(t, flag.IsActive(t.Context(), -1))
	})

	t.Run("before activation time", func(t *testing.T) {
		before := activateAt.Add(-time.Second).UnixMilli()
		assert.False(t, flag.IsActive(t.Context(), before))
	})

	t.Run("exactly at activation time", func(t *testing.T) {
		assert.True(t, flag.IsActive(t.Context(), activateAt.UnixMilli()))
	})

	t.Run("after activation time", func(t *testing.T) {
		after := activateAt.Add(time.Hour).UnixMilli()
		assert.True(t, flag.IsActive(t.Context(), after))
	})

	t.Run("zero-value FeatureFlag is never active", func(t *testing.T) {
		var zeroFlag FeatureFlag
		assert.False(t, zeroFlag.IsActive(t.Context(), time.Now().UnixMilli()))
	})
}
