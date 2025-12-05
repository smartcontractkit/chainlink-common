package settings

import (
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
)

func TestMap_MarshalText(t *testing.T) {
	m := PerChainSelector(Bool(false), map[string]bool{
		"1": true,
		"2": true,
	})
	b, err := toml.Marshal(m)
	require.NoError(t, err)
	require.Equal(t, `Default = "false"

[Values]
  1 = "true"
  2 = "true"
`, string(b))

	m2 := PerChainSelector(Bool(false), nil)
	err = toml.Unmarshal(b, &m2)
	require.NoError(t, err)
	require.Equal(t, m.Values, m2.Values)
}
