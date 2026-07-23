package custmsg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tests CustomMessageAgent does not share state across new instances created by `With`
func Test_CustomMessageAgent(t *testing.T) {
	cma := NewLabeler()
	cma1 := cma.With("key1", "value1").WithType("TestType")
	cma2 := cma1.With("key2", "value2")

	assert.NotEqual(t, cma1.Labels(), cma2.Labels())
}

func Test_CustomMessageAgent_With(t *testing.T) {
	cma := NewLabeler().WithType("TestType").With("key1", "value1")
	assert.Equal(t, map[string]string{"key1": "value1", LabelKeyType: "TestType"}, cma.Labels())
}

func Test_CustomMessageAgent_WithMapLabels(t *testing.T) {
	cma := NewLabeler().WithType("TestType").WithMapLabels(map[string]string{"key1": "value1"})
	assert.Equal(t, map[string]string{"key1": "value1", LabelKeyType: "TestType"}, cma.Labels())
}

func Test_CustomMessageAgent_WithType(t *testing.T) {
	cma := NewLabeler().WithType("NodeConfig")
	assert.Equal(t, map[string]string{LabelKeyType: "NodeConfig"}, cma.Labels())
}

func Test_CustomMessageAgent_WithLabelsAndType(t *testing.T) {
	cma := NewLabeler().WithLabelsAndType(map[string]string{
		"system": "Application",
		"type":   "ignored",
	}, "NodeConfig")
	assert.Equal(t, map[string]string{
		"system":       "Application",
		LabelKeyType:   "NodeConfig",
	}, cma.Labels())
}

func Test_CustomMessageAgent_EmitRequiresType(t *testing.T) {
	err := NewLabeler().Emit(context.Background(), "msg")
	require.ErrorContains(t, err, `missing required label "type"`)
}
