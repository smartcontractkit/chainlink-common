package utils_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
)

func TestErrorBuffer(t *testing.T) {
	t.Parallel()

	err1 := errors.New("err1")
	err2 := errors.New("err2")
	err3 := errors.New("err3")

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		buff := utils.ErrorBuffer{}
		buff.Append(err1)
		buff.Append(err2)
		combined := buff.Flush()
		errs := utils.UnwrapError(combined)
		assert.Equal(t, 2, len(errs))
		assert.Equal(t, err1.Error(), errs[0].Error())
		assert.Equal(t, err2.Error(), errs[1].Error())
	})

	t.Run("ovewrite oldest error when cap exceeded", func(t *testing.T) {
		t.Parallel()
		buff := utils.ErrorBuffer{}
		buff.SetCap(2)
		buff.Append(err1)
		buff.Append(err2)
		buff.Append(err3)
		combined := buff.Flush()
		errs := utils.UnwrapError(combined)
		assert.Equal(t, 2, len(errs))
		assert.Equal(t, err2.Error(), errs[0].Error())
		assert.Equal(t, err3.Error(), errs[1].Error())
	})

	t.Run("does not overwrite the buffer if cap == 0", func(t *testing.T) {
		t.Parallel()
		buff := utils.ErrorBuffer{}
		for i := 1; i <= 20; i++ {
			buff.Append(errors.Errorf("err#%d", i))
		}

		combined := buff.Flush()
		errs := utils.UnwrapError(combined)
		assert.Equal(t, 20, len(errs))
		assert.Equal(t, "err#20", errs[19].Error())
	})

	t.Run("UnwrapError returns the a single element err array if passed err is not a joinedError", func(t *testing.T) {
		t.Parallel()
		errs := utils.UnwrapError(err1)
		assert.Equal(t, 1, len(errs))
		assert.Equal(t, err1.Error(), errs[0].Error())
	})

	t.Run("flushing an empty err buffer is a nil error", func(t *testing.T) {
		t.Parallel()
		buff := utils.ErrorBuffer{}

		combined := buff.Flush()
		require.Nil(t, combined)
	})

}
