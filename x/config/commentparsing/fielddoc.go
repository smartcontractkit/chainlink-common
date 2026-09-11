package commentparsing

import (
	"fmt"
	"reflect"
)

// FieldDoc is what a field's declaration says about itself beyond its type and name.
//
// It is a struct rather than a bare comment string so a later addition - whether a value is a
// default or an example, a deprecation note - extends it without disturbing the DocComments
// signature every generated file implements.
type FieldDoc struct {
	// Comment is the field's doc comment, whole and unabridged rather than the first sentence,
	// because a config reference is exactly where the caveats in the remaining sentences matter.
	Comment string

	// Validate is the field's `validate` tag, verbatim. Reflection can read this from the
	// compiled struct, but a consumer rendering documentation from the method alone cannot, and
	// it is often the only statement that a field is required or mutually exclusive with another.
	Validate string
}

// DocCommenter is implemented by the generated code, returning the name of the type it was
// generated for alongside its fields keyed by Go field name.
//
// The name is returned because an embedded type promotes its methods: a struct that embeds a
// documented type answers DocComments with the embedded type's fields and no indication that the
// outer type's own fields are missing. [Lookup] compares the name it asked about against the name
// it got back, turning that silent wrong answer into an error.
type DocCommenter interface {
	DocComments() (string, map[string]FieldDoc)
}

// Lookup returns the documentation recorded for t, following pointers to the struct type.
//
// Both failures it reports mean the same thing - the generator has not run over the package that
// declares t - so both say so, since a caller hitting this has no other way to know that a
// generation step exists.
func Lookup(t reflect.Type) (map[string]FieldDoc, error) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// A generated method has a value receiver, so *T carries it too, and only *T can be
	// constructed from a type alone.
	documented, ok := reflect.New(t).Interface().(DocCommenter)
	if !ok {
		return nil, fmt.Errorf("%s has no DocComments method: %s", typeName(t), regenerate)
	}

	name, fields := documented.DocComments()
	if name != t.Name() {
		return nil, fmt.Errorf("%s promotes DocComments from embedded %s, so its own fields are undocumented: %s",
			typeName(t), name, regenerate)
	}
	return fields, nil
}

const regenerate = "run `go generate` in the package that declares it"

// typeName qualifies with the package path so the caller can find the package to regenerate,
// which is the only action the errors here ask for.
func typeName(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.String()
	}
	return t.PkgPath() + "." + t.Name()
}
