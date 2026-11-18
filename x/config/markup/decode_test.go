package markup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/x/config/markup"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/yamlmarkup"
)

type decodeConfig struct {
	LogLevel string `validate:"required"`
	Port     uint16 `toml:"listen-port" json:"listen-port"`
	Chains   map[uint32]string
	Optional *struct{ Name string }
}

type requiredConfig struct {
	Set  string
	Port uint16 `toml:"listen-port" json:"listen-port" validate:"required"`
}

func TestDecode(t *testing.T) {
	for _, tc := range []struct {
		name string
		lang markup.Markup
		doc  string
	}{
		{"toml", tomlmarkup.New(), "LogLevel = \"info\"\n[Chains]\n16 = \"a\"\n"},
		{"yaml", yamlmarkup.New(), "LogLevel: info\nChains:\n  16: a\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &decodeConfig{LogLevel: "debug", Port: 9, Chains: map[uint32]string{2: "b"}}
			require.NoError(t, markup.Decode(tc.lang, []byte(tc.doc), c))
			assert.Equal(t, "info", c.LogLevel)
			assert.Equal(t, uint16(9), c.Port)
			assert.Equal(t, map[uint32]string{2: "b", 16: "a"}, c.Chains)
			assert.Nil(t, c.Optional)
		})
	}
}

func TestDecodeErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		lang markup.Markup
		doc  string
	}{
		{"toml", tomlmarkup.New(), "Typo = 1\n"},
		{"yaml", yamlmarkup.New(), "Typo: 1\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := markup.Decode(tc.lang, []byte(tc.doc), &decodeConfig{})
			require.ErrorContains(t, err, "Typo")
		})
	}

	err := markup.Decode(tomlmarkup.New(), []byte("[Chains]\n16 = \"a\"\n0x10 = \"b\"\n"), &decodeConfig{LogLevel: "info"})
	require.EqualError(t, err, `Chains: keys "0x10" and "16" are both 16`)

	t.Run("required", func(t *testing.T) {
		for _, tc := range []struct {
			lang markup.Markup
			doc  string
		}{
			{tomlmarkup.New(), "Set = \"x\"\n"},
			{yamlmarkup.New(), "Set: x\n"},
		} {
			var c requiredConfig

			err := markup.Decode(tc.lang, []byte(tc.doc), &c)
			require.EqualError(t, err, "listen-port failed on the 'required' tag")
		}
	})

	t.Run("target", func(t *testing.T) {
		lang := tomlmarkup.New()
		require.EqualError(t, markup.Decode(lang, nil, decodeConfig{}), "target must be a non-nil pointer to struct")
		assert.ErrorIs(t, markup.Decode(nil, nil, &decodeConfig{}), markup.Err)
	})
}
