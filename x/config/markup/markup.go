// Package markup describes the languages a configuration can be written in.
//
// Nothing outside a language's own package names its library, so a config in another language
// is a value passed in rather than a second code path - and a program linking one language does
// not carry the others. This package depends on nothing but the standard library; a language
// lives beside it, as x/config/markup/tomlmarkup does.
package markup

import (
	"errors"
	"reflect"
)

// Markup is a configuration language, described as its own decoder reads a document.
type Markup interface {
	// Extension is the suffix of a file written in the language, without the dot.
	Extension() string

	// Unmarshal decodes a whole document into v, as the decoder does, and rejects a key v has no field for, naming it.
	Unmarshal(data []byte, v any) error

	// Marshal encodes v as a document the language's own encoder would write.
	Marshal(v any) ([]byte, error)

	// Key is the key a document names f by, or "" if f's own fields are read into its parent's
	// table instead. ok is false if the decoder never reads f.
	Key(f reflect.StructField) (string, bool)

	// RenameTag is a new struct tag that makes the decoder read a field by key,
	// for a type built to decode into. It is built from key alone, not copied from any field.
	RenameTag(key string) reflect.StructTag

	// MatchesKey reports whether the decoder reads a document's key as the field keyed name.
	MatchesKey(key, name string) bool

	// IsLeaf reports whether the decoder reads a value of type t whole, through a method of t's,
	// rather than field by field.
	IsLeaf(t reflect.Type) bool
}

// Err is what a caller reports without a Markup: no language is chosen silently on its behalf.
var Err = errors.New("no markup language set: see x/config/markup/tomlmarkup for an example")
