package commentparsing

import (
	"errors"
	"fmt"
	"reflect"
)

// FieldDoc is what a field's declaration says about itself that compilation discards.
type FieldDoc struct {
	// Comment is the field's whole doc comment, not just the first sentence.
	Comment string
}

// DocCommenter is implemented by the generated code, returning the documentation of the type it
// was generated for, keyed by Go field name.
type DocCommenter interface {
	DocComments() map[string]FieldDoc
}

// Lookup returns the documentation recorded for t, following pointers to the struct type.
//
// A failure means the generator has not run over the package that declares t.
//
// One gap is not reported: a type that embeds a documented type promotes its method, so a struct
// whose own package never generated answers with the embedded type's fields. Those comments are
// right for the fields the embed promoted, and the embedder's own fields are simply absent -
// undocumented, as any field without a comment is. Only a field that shadows an embedded one by
// name takes the wrong comment.
func Lookup(t reflect.Type) (map[string]FieldDoc, error) {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil {
		return nil, errors.New("cannot look up documentation for a nil type")
	}
	if isInstantiatedGeneric(t) {
		return nil, fmt.Errorf("%s: %s", typeName(t), genericUnsupported)
	}

	// A generated method has a value receiver, so *T carries it too, and only *T can be
	// constructed from a type alone.
	documented, ok := reflect.New(t).Interface().(DocCommenter)
	if !ok {
		return nil, fmt.Errorf("%s has no DocComments method: %s", typeName(t), regenerate)
	}
	return documented.DocComments(), nil
}

const regenerate = "run `go generate` in the package that declares it"

// genericUnsupported is reported instead of regenerate, which would send a caller after a file no
// generator can write.
const genericUnsupported = "generic types are not supported: a generated DocComments method " +
	"would need a receiver type parameter list"

// typeName qualifies with the package path.
func typeName(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.String()
	}
	return t.PkgPath() + "." + t.Name()
}
