// Package yamlmarkup is YAML, as sigs.k8s.io/yaml reads it: converted to JSON, then decoded by encoding/json.
//
// So fields are keyed by their json tags, and a type decodes as it does from JSON. A struct written for
// encoding/json reads YAML unchanged. It is a package of its own so that only a program configured in YAML
// links a YAML library.
package yamlmarkup

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"sigs.k8s.io/yaml"

	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

// New returns YAML as sigs.k8s.io/yaml reads it.
func New() markup.Markup { return yamlMarkup{} }

// yamlMarkup implements [markup.Markup].
type yamlMarkup struct{}

// Extension implements [markup.Markup.Extension].
func (yamlMarkup) Extension() string { return "yaml" }

// Unmarshal implements [markup.Markup.Unmarshal].
func (yamlMarkup) Unmarshal(data []byte, v any) error {
	// The library decodes into &v, so it would fill nothing and report no error.
	if rv := reflect.ValueOf(v); rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("decoding can only be performed into a non-nil pointer, not %T", v)
	}

	// Strict, so a typo can't silently leave a default, and a key set twice is an error.
	return yaml.UnmarshalStrict(data, v)
}

// Marshal implements [markup.Markup.Marshal].
func (yamlMarkup) Marshal(v any) ([]byte, error) { return yaml.Marshal(v) }

// Key implements [markup.Markup.Key].
func (yamlMarkup) Key(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("json")
	key, _, _ := strings.Cut(tag, ",")
	if !validKey(key) {
		key = ""
	}

	t := f.Type
	if t.Name() == "" && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch {
	case tag == "-":
		return "", false
	// encoding/json reads an embedded struct even if it's unexported, as it may have exported fields.
	case f.Anonymous && t.Kind() == reflect.Struct:
		return key, true
	case !f.IsExported():
		return "", false
	case key == "":
		return f.Name, true
	default:
		return key, true
	}
}

func validKey(key string) bool {
	return !strings.ContainsFunc(key, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", r)
	})
}

// RenameTag implements [markup.Markup.RenameTag].
func (yamlMarkup) RenameTag(key string, tag reflect.StructTag) reflect.StructTag {
	// ",string" is the one option that changes how encoding/json decodes a value.
	if _, opts, _ := strings.Cut(tag.Get("json"), ","); slices.Contains(strings.Split(opts, ","), "string") {
		key += ",string"
	} else if key == "-" {
		// A bare "-" tag drops the field.
		key = "-,"
	}

	return reflect.StructTag(fmt.Sprintf("json:%q", key))
}

// MatchesKey implements [markup.Markup.MatchesKey].
func (yamlMarkup) MatchesKey(key, name string) bool { return strings.EqualFold(key, name) }

// IsLeaf implements [markup.Markup.IsLeaf].
func (yamlMarkup) IsLeaf(t reflect.Type) bool {
	return implements(t, jsonUnmarshaler) || implements(t, textUnmarshaler)
}

func implements(t, iface reflect.Type) bool {
	return t.Implements(iface) || reflect.PointerTo(t).Implements(iface)
}

var (
	jsonUnmarshaler = reflect.TypeFor[json.Unmarshaler]()
	textUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()
)
