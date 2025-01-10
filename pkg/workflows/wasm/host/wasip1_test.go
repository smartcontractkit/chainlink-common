package host

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSlot(t *testing.T) {
	events := []byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
		250, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63,
	}

	t.Run("getSlot works correctly", func(t *testing.T) {
		expectedSlots := [][]byte{
			{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
			{250, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63},
		}

		for i := int32(0); i < 2; i++ {
			slot, err := getSlot(events, i)
			assert.NoError(t, err)
			assert.Equal(t, expectedSlots[i], slot)
		}
	})
	t.Run("check for out of bound slot", func(t *testing.T) {
		_, err := getSlot(events, 400)
		assert.Error(t, err)
	})
}
