package tomlmarkup

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtension(t *testing.T) {
	assert.Equal(t, "toml", New().Extension())
}

func TestUnmarshal(t *testing.T) {
	var doc struct{ Port int }
	require.EqualError(t, New().Unmarshal([]byte("a = 1\nb = 2\n"), &doc), "unknown configuration key(s): line 1: a, line 2: b")
	require.EqualError(t, New().Unmarshal([]byte("Port = 'x'\n"), &doc), "line 1: cannot decode TOML string into int")
	require.EqualError(t, New().Unmarshal(nil, doc), "toml: decoding can only be performed into a pointer, not struct")
}

func TestMarshal(t *testing.T) {
	data, err := New().Marshal(struct{ Port int }{Port: 1})
	require.NoError(t, err)
	assert.Equal(t, "Port = 1\n", string(data))
}

func TestMatchesKey(t *testing.T) {
	assert.True(t, New().MatchesKey("chainid", "ChainID"))
	assert.False(t, New().MatchesKey("chain_id", "ChainID"))
}

func TestKey(t *testing.T) {
	str := reflect.TypeFor[string]()
	for _, tc := range []struct {
		field reflect.StructField
		key   string
		ok    bool
	}{
		{reflect.StructField{Name: "V", Type: str, Tag: `toml:"-"`}, "", false},
		{reflect.StructField{Name: "S", Type: reflect.TypeFor[*struct{}](), Anonymous: true}, "", true},
		{reflect.StructField{Name: "S", Type: str, Anonymous: true, Tag: `toml:"k"`}, "k", true},
		{reflect.StructField{Name: "v", PkgPath: "p", Type: str}, "", false},
		{reflect.StructField{Name: "V", Type: str}, "V", true},
		{reflect.StructField{Name: "V", Type: str, Tag: `toml:"k"`}, "k", true},
	} {
		key, ok := New().Key(tc.field)
		assert.Equal(t, tc.key, key, tc.field)
		assert.Equal(t, tc.ok, ok, tc.field)
	}
}

func TestRenameTag(t *testing.T) {
	assert.Equal(t, reflect.StructTag(`toml:"k"`), New().RenameTag("k"))
	assert.Equal(t, reflect.StructTag(`toml:"-,"`), New().RenameTag("-"))
}

func TestIsLeaf(t *testing.T) {
	assert.True(t, New().IsLeaf(reflect.TypeFor[readsTextByPointer]()))
	assert.False(t, New().IsLeaf(reflect.TypeFor[struct{}]()))
}

type readsTextByPointer struct{}

func (*readsTextByPointer) UnmarshalText([]byte) error { return nil }
