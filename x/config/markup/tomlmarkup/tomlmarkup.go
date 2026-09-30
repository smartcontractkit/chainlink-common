// Package tomlmarkup is TOML, as github.com/pelletier/go-toml/v2 reads it.
//
// It is a package of its own so that only a program configured in TOML links a TOML library.
package tomlmarkup

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

// New returns TOML as go-toml reads it.
func New() markup.Markup { return tomlMarkup{} }

// tomlMarkup implements [markup.Markup].
type tomlMarkup struct{}

// Extension implements [markup.Markup.Extension].
func (tomlMarkup) Extension() string { return "toml" }

// Unmarshal implements [markup.Markup.Unmarshal].
func (tomlMarkup) Unmarshal(data []byte, v any) error {
	// Strict, so a typo can't silently leave a default.
	err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(v)
	if strict, ok := errors.AsType[*toml.StrictMissingError](err); ok {
		keys := make([]string, len(strict.Errors))
		for i, e := range strict.Errors {
			line, _ := e.Position()
			keys[i] = fmt.Sprintf("line %d: %s", line, strings.Join(e.Key(), "."))
		}
		return fmt.Errorf("unknown configuration key(s): %s", strings.Join(keys, ", "))
	}
	decode, ok := errors.AsType[*toml.DecodeError](err)
	if !ok {
		return err
	}
	msg := strings.TrimPrefix(decode.Error(), "toml: ")
	// Name the field's type, not the struct's. A caller may build that struct only to decode into.
	if i, j := strings.Index(msg, " into struct field "), strings.LastIndex(msg, " of type "); i >= 0 && j > i {
		msg = msg[:i] + " into " + msg[j+len(" of type "):]
	}
	line, _ := decode.Position()
	return fmt.Errorf("line %d: %s", line, msg)
}

// Marshal implements [markup.Markup.Marshal].
func (tomlMarkup) Marshal(v any) ([]byte, error) { return toml.Marshal(v) }

// Key implements [markup.Markup.Key].
func (tomlMarkup) Key(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("toml")
	key, _, _ := strings.Cut(tag, ",")
	switch t := f.Type; {
	case tag == "-":
		return "", false
	case f.Anonymous && key == "":
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		return "", t.Kind() == reflect.Struct
	case f.Anonymous:
		return key, true
	case !f.IsExported():
		return "", false
	case key == "":
		return f.Name, true
	default:
		return key, true
	}
}

// RenameTag implements [markup.Markup.RenameTag].
func (tomlMarkup) RenameTag(key string) reflect.StructTag {
	// A bare "-" tag drops the field.
	if key == "-" {
		key = "-,"
	}
	return reflect.StructTag(fmt.Sprintf("toml:%q", key))
}

// MatchesKey implements [markup.Markup.MatchesKey].
func (tomlMarkup) MatchesKey(key, name string) bool {
	return key == name || strings.ToLower(key) == strings.ToLower(name) //nolint:staticcheck // SA6005: go-toml lowercases; EqualFold differs on some runes
}

// IsLeaf implements [markup.Markup.IsLeaf].
func (tomlMarkup) IsLeaf(t reflect.Type) bool {
	return t.Implements(textUnmarshaler) || reflect.PointerTo(t).Implements(textUnmarshaler)
}

var textUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()
