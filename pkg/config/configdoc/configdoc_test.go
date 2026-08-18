package configdoc

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/gen_exp.md
var exp string

func TestGenerate(t *testing.T) {
	def := `
# Foo is a boolean field.
Foo = false # Default
# Bar is a number.
Bar = 42 # Example
[Baz]
# Test holds a string.
Test = "test" # Example`

	header := `# Example docs
This is the header. It has a list:
- first
- second`

	example := `Bar = 10
Baz.Test = "asdf"`

	s, err := Generate(def, header, example, map[string]string{
		"Baz": "Baz has an extended description",
	})
	require.NoError(t, err)

	require.Equal(t, exp, s)
}
