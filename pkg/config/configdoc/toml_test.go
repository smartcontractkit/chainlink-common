package configdoc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

func TestTOMLName(t *testing.T) {
	// Used as the markdown code fence language.
	assert.Equal(t, "toml", TOML{}.Name())
}

func TestTOMLComment(t *testing.T) {
	assert.Equal(t, "# the host", TOML{}.Comment("the host"))
}

func TestTOMLTable(t *testing.T) {
	assert.Equal(t, "[chain]", TOML{}.Table("chain"))
	assert.Equal(t, "[chain.nodes]", TOML{}.Table("chain.nodes"))
}

func TestTOMLField(t *testing.T) {
	f := TOML{}
	assert.Equal(t, "host = 'x' # Default", f.Field("host", "'x'", f.DefaultMarker()))
	assert.Equal(t, "host = 'x' # Example", f.Field("host", "'x'", f.ExampleMarker()))
	// No marker: a plain value line, for the example config rather than the docs.
	assert.Equal(t, "host = 'x'", f.Field("host", "'x'", ""))
}

func TestTOMLMarkers(t *testing.T) {
	assert.Equal(t, FieldDefault, TOML{}.DefaultMarker())
	assert.Equal(t, FieldExample, TOML{}.ExampleMarker())
}

func TestTOMLParseLine(t *testing.T) {
	for _, tt := range []struct {
		name string
		line string
		want Line
	}{
		{"comment", "# the host", Line{Kind: LineComment, Text: "the host"}},
		{"comment without space", "#the host", Line{Kind: LineComment, Text: "the host"}},
		{"empty", "", Line{Kind: LineBlank}},
		{"whitespace only", "   ", Line{Kind: LineBlank}},
		{"table", "[chain]", Line{Kind: LineTable, Text: "chain"}},
		{"nested table", "[chain.nodes]", Line{Kind: LineTable, Text: "chain.nodes"}},
		{"array of tables", "[[EVM]]", Line{Kind: LineArrayOfTables, Text: "EVM"}},
		{"field", "host = 'x' # Default", Line{Kind: LineField, Text: "host"}},
		{"field with no marker", "host = 'x'", Line{Kind: LineField, Text: "host"}},
		{"indented field", "  host = 'x'", Line{Kind: LineField, Text: "host"}},
		// Degenerate, but must not panic - Text is simply the whole line.
		{"field with no space", "host", Line{Kind: LineField, Text: "host"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, TOML{}.ParseLine(tt.line))
		})
	}
}

// ParseLine has to read back whatever the emitting side writes, or a document assembled with
// this format can't be turned into documentation by it.
func TestTOMLEmitAndParseAgree(t *testing.T) {
	f := TOML{}

	assert.Equal(t, Line{Kind: LineComment, Text: "the host"}, f.ParseLine(f.Comment("the host")))
	assert.Equal(t, Line{Kind: LineTable, Text: "chain.nodes"}, f.ParseLine(f.Table("chain.nodes")))
	assert.Equal(t, Line{Kind: LineField, Text: "host"}, f.ParseLine(f.Field("host", "'x'", f.DefaultMarker())))
	assert.Equal(t, Line{Kind: LineField, Text: "host"}, f.ParseLine(f.Field("host", "'x'", "")))
}

func TestTOMLLiteral(t *testing.T) {
	for _, tt := range []struct {
		name string
		val  any
		want string
	}{
		{"string", "example.com", "'example.com'"},
		{"empty string", "", "''"},
		{"bool", true, "true"},
		{"int", 42, "42"},
		{"negative int", -7, "-7"},
		{"sized int", int32(42), "42"},
		{"uint", uint32(42), "42"},
		{"float", 1.5, "1.5"},
		{"nil", nil, "''"},
		{"string slice", []string{"a", "b"}, "['a', 'b']"},
		{"empty slice", []string{}, "[]"},
		{"int slice", []int{1, 2}, "[1, 2]"},
		// A duration is an int64 underneath; it must not render as a raw nanosecond count.
		{"duration", 90 * time.Second, "'1m30s'"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, TOML{}.Literal(tt.val))
		})
	}
}

func TestTOMLLiteralUsesTextMarshaler(t *testing.T) {
	// config.Duration marshals itself to text, so it round-trips through the same
	// UnmarshalText that reads it back rather than being printed as a struct.
	d := config.MustNewDuration(90 * time.Second)

	assert.Equal(t, "'1m30s'", TOML{}.Literal(*d))
	assert.Equal(t, "'1m30s'", TOML{}.Literal(d), "a pointer is dereferenced first")
}

func TestTOMLLiteralNilPointer(t *testing.T) {
	var d *config.Duration
	assert.Equal(t, "''", TOML{}.Literal(d))
}

// The literals this format writes have to be valid input for the document it produces, so a
// generated document round-trips through Generate without tripping its own parser.
func TestTOMLLiteralsRoundTripThroughGenerate(t *testing.T) {
	f := TOML{}

	doc := f.Comment("host the host") + "\n" +
		f.Field("host", f.Literal("example.com"), f.DefaultMarker()) + "\n" +
		f.Comment("retries how many retries") + "\n" +
		f.Field("retries", f.Literal(3), f.DefaultMarker()) + "\n" +
		f.Comment("timeout how long to wait") + "\n" +
		f.Field("timeout", f.Literal(90*time.Second), f.ExampleMarker()) + "\n"

	out, err := GenerateWith(f, doc, "# Docs\n", "", nil)
	require.NoError(t, err)
	assert.Contains(t, out, "host = 'example.com' # Default")
	assert.Contains(t, out, "retries = 3 # Default")
	assert.Contains(t, out, "timeout = '1m30s' # Example")
}

func TestGenerateWithUsesFormatName(t *testing.T) {
	out, err := GenerateWith(TOML{}, "# host the host\nhost = 'x' # Default\n", "# Docs\n", "host = 'x'", nil)
	require.NoError(t, err)

	// The code fences are labelled with the format, exactly once.
	assert.Contains(t, out, "```toml\n")
	assert.NotContains(t, out, "```tomltoml")
}
