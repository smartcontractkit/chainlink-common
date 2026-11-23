// Package decode holds the pieces [markup.Decode] and x/config/cli share: the key tree a document is checked against,
// the reshaping that reads every map key as text, text parsing, and tag validation. It is internal so both read the
// same code while markup's public API stays one entry point.
package decode

import "reflect"

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

	// RenameTag is a new struct tag that makes the decoder read a field by key, for a type built to decode into.
	// From tag, the tag of the field it stands in for, if any, it keeps only what changes how a value decodes.
	RenameTag(key string, tag reflect.StructTag) reflect.StructTag

	// MatchesKey reports whether the decoder reads a document's key as the field keyed name.
	MatchesKey(key, name string) bool

	// IsLeaf reports whether the decoder reads a value of type t whole, through a method of t's,
	// rather than field by field.
	IsLeaf(t reflect.Type) bool
}
