package wasm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncer(t *testing.T) {
	logs = make([][]byte, 0)
	w := &writer{}

	n, err := w.Write([]byte("Hello, World!"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)

	n, err = w.Write([]byte("Again"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)

	assert.Len(t, logs, 2)
	assert.Equal(t, "Hello, World!", string(logs[0]))
	assert.Equal(t, "Again", string(logs[1]))
}
