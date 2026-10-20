package custmsg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/beholdertest"
)

// tests CustomMessageAgent does not share state across new instances created by `With`
func Test_CustomMessageAgent(t *testing.T) {
	cma := NewLabeler()
	cma1 := cma.With("key1", "value1")
	cma2 := cma1.With("key2", "value2")

	assert.NotEqual(t, cma1.Labels(), cma2.Labels())
}

func Test_CustomMessageAgent_With(t *testing.T) {
	cma := NewLabeler().With("key1", "value1")
	assert.Equal(t, map[string]string{"key1": "value1"}, cma.Labels())
}

func Test_CustomMessageAgent_WithMapLabels(t *testing.T) {
	cma := NewLabeler().WithMapLabels(map[string]string{"key1": "value1"})
	assert.Equal(t, map[string]string{"key1": "value1"}, cma.Labels())
}

func TestCustomLabelerWithEmitter(t *testing.T) {
	t.Parallel()

	rec := beholdertest.NewRecorder()
	l := NewLabelerWithEmitter(rec).
		With("key1", "value1").
		WithMapLabels(map[string]string{"key2": "value2"})

	require.NoError(t, l.Emit(t.Context(), "hello"))

	found, err := rec.BaseMessagesForLabels(t, map[string]string{"key1": "value1", "key2": "value2"})
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.Equal(t, "hello", found[0].Msg)

	require.Len(t, rec.Messages(t), 1)
}

func TestCustomLabelerUsesGlobalEmitter(t *testing.T) {
	o := beholdertest.NewObserver(t)

	require.NoError(t, NewLabeler().Emit(t.Context(), "hello"))

	found, err := o.BaseMessagesForLabels(t, nil)
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.Equal(t, "hello", found[0].Msg)
}
