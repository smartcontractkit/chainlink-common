package cli

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTextRejectsATypeWithNoTextForm(t *testing.T) {
	_, err := parseText(reflect.TypeFor[struct{}](), "x")
	require.ErrorContains(t, err, "has no text form")
}
