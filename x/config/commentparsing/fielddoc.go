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
// name takes the wrong comment; two embeds promoting one name are refused at generation instead.
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
	documented, ok := allocate(t).(DocCommenter)
	if !ok {
		return nil, fmt.Errorf("%s has no DocComments method: %s", typeName(t), regenerate)
	}
	return call(documented, t)
}

// allocate builds the value the method is called on, filling in the embedded pointers a promoted
// method dispatches through: the zero value of one is nil, and a value receiver reached through it
// dereferences nothing.
func allocate(t reflect.Type) any {
	value := reflect.New(t)
	fillEmbedded(value.Elem(), map[reflect.Type]bool{})
	return value.Interface()
}

// fillEmbedded recurses through embedded fields, which are the only ones that promote a method.
//
// A type is filled once along a path, since a struct may embed a pointer to itself and allocating
// that pointer produces another struct with the same field. Only a cycle is left unfilled, and no
// promotion runs through one: Go resolves a promoted method at the shallowest depth it appears, so
// every dispatch path is finite.
func fillEmbedded(v reflect.Value, filling map[reflect.Type]bool) {
	t := v.Type()
	if t.Kind() != reflect.Struct || filling[t] {
		return
	}
	filling[t] = true
	defer delete(filling, t)

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.Anonymous {
			continue
		}

		embedded := v.Field(i)
		switch {
		case embedded.Kind() != reflect.Pointer:
			fillEmbedded(embedded, filling)
		// An unexported embedded pointer cannot be set, and neither can an embedded interface
		// be given an implementation. Whether either is on the path a promotion dispatches
		// through is what call reports.
		case embedded.CanSet():
			embedded.Set(reflect.New(field.Type.Elem()))
			fillEmbedded(embedded.Elem(), filling)
		}
	}
}

// call invokes the method, reporting the receiver a promotion could not be built for rather than
// letting the wrapper's nil dereference reach the caller. An embedded interface, and an unexported
// embedded pointer, are the two [allocate] cannot fill.
func call(documented DocCommenter, t reflect.Type) (fields map[string]FieldDoc, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%s promotes DocComments through an embedded field that cannot be "+
				"constructed from the type alone (%v): %s", typeName(t), recovered, regenerate)
		}
	}()
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
