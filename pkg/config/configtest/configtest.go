package configtest

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

// AssertFieldsNotNil recursively checks s for nil fields. s must be a struct.
func AssertFieldsNotNil(t *testing.T, s interface{}) {
	t.Helper()
	err := assertValNotNil(t, "", reflect.ValueOf(s))
	_, err = config.MultiErrorList(err)
	assert.NoError(t, err)
}

// assertFieldsNotNil recursively checks the struct s for nil fields.
func assertFieldsNotNil(t *testing.T, prefix string, s reflect.Value) (err error) {
	t.Helper()
	require.Equal(t, reflect.Struct, s.Kind())

	typ := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		key := prefix
		if tf := typ.Field(i); !tf.Anonymous {
			if key != "" {
				key += "."
			}
			key += tf.Name
		}
		err = errors.Join(err, assertValNotNil(t, key, f))
	}
	return
}

// assertValuesNotNil recursively checks the map m for nil values.
func assertValuesNotNil(t *testing.T, prefix string, m reflect.Value) (err error) {
	t.Helper()
	require.Equal(t, reflect.Map, m.Kind())
	if prefix != "" {
		prefix += "."
	}

	mi := m.MapRange()
	for mi.Next() {
		key := prefix + mi.Key().String()
		err = errors.Join(err, assertValNotNil(t, key, mi.Value()))
	}
	return
}

// assertElementsNotNil recursively checks the slice s for nil values.
func assertElementsNotNil(t *testing.T, prefix string, s reflect.Value) (err error) {
	t.Helper()
	require.Equal(t, reflect.Slice, s.Kind())

	for i := 0; i < s.Len(); i++ {
		err = errors.Join(err, assertValNotNil(t, prefix, s.Index(i)))
	}
	return
}

var (
	textUnmarshaler     encoding.TextUnmarshaler
	textUnmarshalerType = reflect.TypeOf(&textUnmarshaler).Elem()
)

// assertValNotNil recursively checks that val is not nil. val must be a struct, map, slice, or point to one.
func assertValNotNil(t *testing.T, key string, val reflect.Value) error {
	t.Helper()
	k := val.Kind()
	switch k { //nolint:exhaustive
	case reflect.Ptr:
		if val.IsNil() {
			return fmt.Errorf("%s: nil", key)
		}
	}
	if k == reflect.Ptr {
		if val.Type().Implements(textUnmarshalerType) {
			return nil // skip values unmarshaled from strings
		}
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Struct:
		if val.Type().Implements(textUnmarshalerType) {
			return nil // skip values unmarshaled from strings
		}
		return assertFieldsNotNil(t, key, val)
	case reflect.Map:
		if val.IsNil() {
			return nil // not actually a problem
		}
		return assertValuesNotNil(t, key, val)
	case reflect.Slice:
		if val.IsNil() {
			return nil // not actually a problem
		}
		return assertElementsNotNil(t, key, val)
	default:
		return nil
	}
}

// AssertDocsTOMLComplete ensures that docsTOML contains every field in C and no extra fields.
func AssertDocsTOMLComplete[C any](t *testing.T, docsTOML string) {
	t.Helper()
	var c C
	err := config.DecodeTOML(strings.NewReader(docsTOML), &c)
	if err != nil && strings.Contains(err.Error(), "undecoded keys: ") {
		t.Errorf("Docs contain extra fields: %v", err)
	} else {
		require.NoError(t, err)
	}
	AssertFieldsNotNil(t, c)
}

// AssertFullMarshal ensures that c encodes to expTOML and contains no nil fields.
func AssertFullMarshal[C any](t *testing.T, c C, expTOML string) {
	t.Helper()
	AssertFieldsNotNil(t, c)

	b, err := toml.Marshal(c)
	require.NoError(t, err)
	s := string(b)
	t.Log(s)
	assert.Equal(t, expTOML, s, diff.Diff(expTOML, s))
}
