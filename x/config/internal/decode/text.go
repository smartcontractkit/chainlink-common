package decode

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// durationType tells time.Duration apart from int64, which shares its Kind.
var durationType = reflect.TypeFor[time.Duration]()

func DerefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

// A flag or env var carries only text, as does a document's map key. An interface can't be parsed into, since there's
// no concrete type to make.
func ReadsText(t reflect.Type) bool {
	return t.Kind() != reflect.Interface &&
		(t.Implements(textUnmarshaler) || reflect.PointerTo(t).Implements(textUnmarshaler))
}

var textUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()

// Parsed as a cli flag would be: any Go literal base, and sized, so 256 overflows a uint8. []byte is text (JSON, PEM),
// not a list of numbers.
func SetText(dst reflect.Value, s string) error {
	t := dst.Type()
	switch {
	case ReadsText(t):
		return dst.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
	case t == durationType:
		d, err := time.ParseDuration(s)
		dst.SetInt(int64(d))
		return err
	}

	var err error
	switch t.Kind() {
	case reflect.String:
		dst.SetString(s)
	case reflect.Slice:
		dst.SetBytes([]byte(s))
	case reflect.Bool:
		var b bool
		b, err = strconv.ParseBool(s)
		dst.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var n int64
		n, err = strconv.ParseInt(s, 0, t.Bits())
		dst.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var n uint64
		n, err = strconv.ParseUint(s, 0, t.Bits())
		dst.SetUint(n)
	case reflect.Float32, reflect.Float64:
		var n float64
		n, err = strconv.ParseFloat(s, t.Bits())
		dst.SetFloat(n)
	default:
		err = fmt.Errorf("%s has no text form", t)
	}

	return err
}
