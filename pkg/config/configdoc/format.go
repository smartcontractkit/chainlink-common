package configdoc

// LineKind classifies a line of a configuration document.
type LineKind int

const (
	// LineBlank is an empty or whitespace-only line, which terminates a comment block.
	LineBlank LineKind = iota
	// LineComment is a description line.
	LineComment
	// LineTable opens a table (section) of fields.
	LineTable
	// LineArrayOfTables opens a repeated table.
	LineArrayOfTables
	// LineField is a key/value pair.
	LineField
)

// Line is a parsed line: its kind, plus the piece of it that carries meaning - the comment
// text with its marker stripped, the table's name, or the field's key.
type Line struct {
	Kind LineKind
	Text string
}

// Format is a configuration file syntax. It both writes the pieces of a document (so callers
// can assemble one from Go structs) and recognizes them when reading one back (so a
// hand-written document can be turned into documentation). Implement it to document a format
// other than TOML; see TOML for the reference implementation.
type Format interface {
	// Comment renders text as a description line.
	Comment(text string) string
	// Table renders the opening of a section, given its dotted path (e.g. "chain.nodes").
	Table(path string) string
	// Field renders a key/value line. marker is DefaultMarker, ExampleMarker or
	// DocsOnlyMarker, or empty for a plain value line with no documentation annotation.
	Field(key, value, marker string) string
	// Literal renders a Go value as a value literal of this format.
	Literal(v any) string

	// DefaultMarker annotates a field whose value is its real default.
	DefaultMarker() string
	// ExampleMarker annotates a field with no usable default, whose value is a placeholder.
	ExampleMarker() string
	// DocsOnlyMarker annotates a field that is documented but shown in no example.
	DocsOnlyMarker() string

	// ParseLine classifies one line of a document.
	ParseLine(line string) Line

	// Name returns the markdown name for code blocks
	Name() string
}
