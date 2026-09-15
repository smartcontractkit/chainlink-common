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

var docCommenter = reflect.TypeFor[DocCommenter]()

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
	value, err := allocate(t)
	if err != nil {
		return nil, err
	}
	documented, ok := value.(DocCommenter)
	if !ok {
		return nil, fmt.Errorf("%s has no DocComments method: %s", typeName(t), regenerate)
	}
	return documented.DocComments(), nil
}

// embedDepth bounds the walk allocate makes, since a struct may embed a pointer to itself.
const embedDepth = 16

// allocate builds the value the method is called on, with the embedded pointers a promoted method
// may dispatch through allocated: the zero value of an embedded pointer is nil, and a value
// receiver reached through it dereferences nothing.
func allocate(t reflect.Type) (any, error) {
	value := reflect.New(t)
	if err := fillEmbedded(value.Elem(), 0); err != nil {
		return nil, err
	}
	return value.Interface(), nil
}

func fillEmbedded(v reflect.Value, depth int) error {
	t := v.Type()
	if t.Kind() != reflect.Struct || depth == embedDepth {
		return nil
	}

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.Anonymous {
			continue // only an embedded field promotes a method
		}

		embedded := v.Field(i)
		if embedded.Kind() != reflect.Pointer {
			if err := fillEmbedded(embedded, depth+1); err != nil {
				return err
			}
			continue
		}

		// An unexported embedded pointer is not settable, so a type documented only through
		// one is reported rather than called on a nil receiver.
		if !embedded.CanSet() {
			if implementsDocCommenter(field.Type) {
				return fmt.Errorf("%s promotes DocComments through the unexported embedded pointer %s, "+
					"which cannot be allocated to read it: %s", typeName(t), field.Name, regenerate)
			}
			continue
		}

		embedded.Set(reflect.New(field.Type.Elem()))
		if err := fillEmbedded(embedded.Elem(), depth+1); err != nil {
			return err
		}
	}
	return nil
}

func implementsDocCommenter(t reflect.Type) bool {
	return t.Implements(docCommenter) || reflect.PointerTo(t).Implements(docCommenter)
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
