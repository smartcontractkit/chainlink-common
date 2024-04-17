package v2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type typedCapability struct {
	Capability[AnInput, AnOutput, AConfig]
	inp  AnInput
	conf AConfig
}

type AnInput struct {
	Foo string `json:"foo"`
}

type AConfig struct {
	Bar string `json:"bar"`
}

type AnOutput struct {
	Baz string `json:"baz"`
}

func (t *typedCapability) Execute(ctx context.Context, callback chan<- CapabilityResponse[AnOutput], req CapabilityRequest[AnInput, AConfig]) error {
	t.inp = req.Inputs
	t.conf = req.Config
	return nil
}

func TestCapabilityV2_Execute(t *testing.T) {
	c := &typedCapability{}
	cap := NewCapability(c)

	cb := make(chan capabilities.CapabilityResponse)

	conf, err := values.NewMap(map[string]any{
		"bar": "config-string",
	})
	require.NoError(t, err)

	inp, err := values.NewMap(map[string]any{
		"foo": "input-string",
	})
	require.NoError(t, err)

	req := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{},
		Config:   conf,
		Inputs:   inp,
	}
	err = cap.Execute(context.Background(), cb, req)
	require.NoError(t, err)

	assert.Equal(t, AnInput{Foo: "input-string"}, c.inp)
	assert.Equal(t, AConfig{Bar: "config-string"}, c.conf)
}
