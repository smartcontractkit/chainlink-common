package yamlmarkup

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/x/config/markup/markuptests"
)

func TestMarkup(t *testing.T) {
	markuptests.Run(t, New())
}

func TestExtension(t *testing.T) {
	assert.Equal(t, "yaml", New().Extension())
}

func TestUnmarshal(t *testing.T) {
	var doc struct{ Port int }
	require.EqualError(t, New().Unmarshal([]byte("a: 1\n"), &doc), `error unmarshaling JSON: while decoding JSON: json: unknown field "a"`)
	require.EqualError(t, New().Unmarshal(nil, doc), "decoding can only be performed into a non-nil pointer, not struct { Port int }")
}

func TestMarshal(t *testing.T) {
	data, err := New().Marshal(struct{ Port int }{Port: 1})
	require.NoError(t, err)
	assert.Equal(t, "Port: 1\n", string(data))
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
		{reflect.StructField{Name: "V", Type: str, Tag: `json:"-"`}, "", false},
		{reflect.StructField{Name: "S", Type: reflect.TypeFor[*struct{}](), Anonymous: true}, "", true},
		{reflect.StructField{Name: "s", PkgPath: "p", Type: reflect.TypeFor[struct{}](), Anonymous: true, Tag: `json:"k"`}, "k", true},
		{reflect.StructField{Name: "v", PkgPath: "p", Type: str}, "", false},
		{reflect.StructField{Name: "V", Type: str}, "V", true},
		{reflect.StructField{Name: "V", Type: str, Tag: `json:"k,omitempty"`}, "k", true},
		{reflect.StructField{Name: "V", Type: str, Tag: `json:"a\"b"`}, "V", true},
	} {
		key, ok := New().Key(tc.field)
		assert.Equal(t, tc.key, key, tc.field)
		assert.Equal(t, tc.ok, ok, tc.field)
	}
}

func TestRenameTag(t *testing.T) {
	assert.Equal(t, reflect.StructTag(`json:"k"`), New().RenameTag("k", `json:"x,omitempty"`))
	assert.Equal(t, reflect.StructTag(`json:"k,string"`), New().RenameTag("k", `json:"x,omitempty,string"`))
	assert.Equal(t, reflect.StructTag(`json:"-,"`), New().RenameTag("-", ""))
}

func TestIsLeaf(t *testing.T) {
	assert.True(t, New().IsLeaf(reflect.TypeFor[readsTextByPointer]()))
	assert.True(t, New().IsLeaf(reflect.TypeFor[readsJSON]()))
	assert.False(t, New().IsLeaf(reflect.TypeFor[struct{}]()))
}

type readsTextByPointer struct{}

func (*readsTextByPointer) UnmarshalText([]byte) error { return nil }

type readsJSON struct{}

func (readsJSON) UnmarshalJSON([]byte) error { return nil }
