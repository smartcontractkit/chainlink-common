package configdoc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

const (
	FieldDefault = "# Default"
	FieldExample = "# Example"

	TokenAdvanced = "**ADVANCED**"
	TokenMap      = "**MAP**"
)

// Generate returns MarkDown documentation generated from the TOML string.
//   - Each field but include a trailing comment of either FieldDefault or FieldExample.
//   - If a description begins with TokenAdvanced, then a warning will be included.
//   - If a table description begins with TokenMap, then its keys are map entries, and
//     they inherit the table description rather than requiring one of their own.
//   - The markdown wil begin with the header, followed by the example
//   - Extended descriptions can be applied to top level tables
func Generate(toml, header, example string, extendedDescriptions map[string]string) (string, error) {
	items, err := parseTOMLDocs(toml, extendedDescriptions)
	var sb strings.Builder

	sb.WriteString(header)
	sb.WriteString(`
## Example

`)
	sb.WriteString("```toml\n")
	sb.WriteString(example)
	sb.WriteString("\n```\n\n")

	for _, item := range items {
		sb.WriteString(item.String())
		sb.WriteString("\n\n")
	}

	return sb.String(), err
}

func advancedWarning(msg string) string {
	return fmt.Sprintf(":warning: **_ADVANCED_**: _%s_\n", msg)
}

func isGeneratedPreamble(desc lines) bool {
	return len(desc) > 0 && strings.Contains(strings.ToLower(desc[0]), "generated")
}

func tableName(line string) string {
	if i := indexOutsideQuotes(line, '#'); i > -1 {
		line = line[:i]
	}
	return strings.Trim(strings.TrimSpace(line), "[]")
}

// lines holds a set of contiguous lines
type lines []string

func (d lines) String() string {
	return strings.Join(d, "\n")
}

type table struct {
	name     string
	codes    lines
	adv      bool
	isMap    bool
	desc     lines
	extended string
}

func newTable(line string, desc lines, extendedDescriptions map[string]string) *table {
	t := &table{
		name:  tableName(line),
		codes: []string{line},
		desc:  desc,
	}
	if extended, ok := extendedDescriptions[t.name]; ok {
		t.extended = extended
	}
	t.parseTokens()
	return t
}

func newArrayOfTables(line string, desc lines, extendedDescriptions map[string]string) *table {
	t := &table{
		name:  tableName(line),
		codes: []string{line},
		desc:  desc,
	}
	if extended, ok := extendedDescriptions[t.name]; ok {
		t.extended = extended
	}
	t.parseTokens()
	return t
}

func (t *table) parseTokens() {
	for len(t.desc) > 0 {
		switch first := strings.TrimSpace(t.desc[0]); {
		case strings.HasPrefix(first, TokenAdvanced):
			t.adv = true
		case strings.HasPrefix(first, TokenMap):
			t.isMap = true
		default:
			return
		}
		t.desc = t.desc[1:]
	}
}

func (t table) advanced() string {
	if t.adv {
		return advancedWarning("Do not change these settings unless you know what you are doing.")
	}
	return ""
}

func (t table) code() string {
	if t.extended == "" {
		return fmt.Sprint("```toml\n", t.codes, "\n```\n")
	}
	return ""
}

// String prints a table as an H2, followed by a code block and description.
func (t *table) String() string {
	return fmt.Sprint("## ", t.name, "\n",
		t.advanced(),
		t.code(),
		t.desc,
		t.extended)
}

type keyval struct {
	name string
	code string
	adv  bool
	desc lines
}

// indexOutsideQuotes returns the index of the first b which is not inside a quoted
// TOML key or string, or -1. Quoted keys may contain '=' and '#', which are
// otherwise a key/value separator and the start of a comment.
func indexOutsideQuotes(line string, b byte) int {
	var inBasic, inLiteral bool
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case inBasic:
			switch c {
			case '\\':
				i++
			case '"':
				inBasic = false
			}
		case inLiteral:
			if c == '\'' {
				inLiteral = false
			}
		case c == '"':
			inBasic = true
		case c == '\'':
			inLiteral = true
		case c == b:
			return i
		}
	}
	return -1
}

// keyName returns the key of a TOML key/value line. Whitespace around the '='
// separator is optional.
func keyName(line string) string {
	line = strings.TrimSpace(line)
	if i := indexOutsideQuotes(line, '='); i > -1 {
		return strings.TrimSpace(line[:i])
	}
	return line
}

func newKeyval(line string, desc lines) keyval {
	line = strings.TrimSpace(line)
	kv := keyval{
		name: keyName(line),
		code: line,
		desc: desc,
	}
	if len(desc) > 0 && strings.HasPrefix(strings.TrimSpace(desc[0]), TokenAdvanced) {
		kv.adv = true
		kv.desc = kv.desc[1:]
	}
	return kv
}

func (k keyval) advanced() string {
	if k.adv {
		return advancedWarning("Do not change this setting unless you know what you are doing.")
	}
	return ""
}

// String prints a keyval as an H3, followed by a code block and description.
func (k keyval) String() string {
	name := k.name
	if i := strings.LastIndex(name, "."); i > -1 {
		name = name[i+1:]
	}
	return fmt.Sprint("### ", name, "\n",
		k.advanced(),
		"```toml\n",
		k.code,
		"\n```\n",
		k.desc)
}

func parseTOMLDocs(s string, extendedDescriptions map[string]string) (items []fmt.Stringer, err error) {
	defer func() { _, err = config.MultiErrorList(err) }()
	globalTable := table{name: "Global"}
	currentTable := &globalTable
	items = append(items, currentTable)
	var desc lines
	for line := range strings.SplitSeq(s, "\n") {
		// Indentation is insignificant in TOML, and nested tables and their keys
		// are conventionally indented.
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "#"):
			// comment
			desc = append(desc, strings.TrimSpace(trimmed[1:]))
		case trimmed == "":
			// empty
			if len(desc) > 0 {
				if !isGeneratedPreamble(desc) {
					items = append(items, desc)
				}
				desc = nil
			}
		case strings.HasPrefix(trimmed, "[["):
			currentTable = newArrayOfTables(trimmed, desc, extendedDescriptions)
			items = append(items, currentTable)
			desc = nil
		case strings.HasPrefix(trimmed, "["):
			currentTable = newTable(trimmed, desc, extendedDescriptions)
			items = append(items, currentTable)
			desc = nil
		default:
			kv := newKeyval(trimmed, desc)
			shortName := kv.name
			if currentTable != &globalTable {
				// update to full name
				kv.name = currentTable.name + "." + kv.name
			}
			switch {
			case len(kv.desc) == 0 && currentTable.isMap && len(currentTable.desc) > 0:
				// A map entry is keyed by data - a chain selector, a telemetry label -
				// so it has no declaring struct field to carry a doc comment.
				kv.desc = currentTable.desc
			case len(kv.desc) == 0:
				err = errors.Join(err, fmt.Errorf("%s: missing description", kv.name))
			case !strings.HasPrefix(kv.desc[0], shortName):
				err = errors.Join(err, fmt.Errorf("%s: description does not begin with %q", kv.name, shortName))
			}
			if !strings.HasSuffix(trimmed, FieldDefault) && !strings.HasSuffix(trimmed, FieldExample) {
				err = errors.Join(err, fmt.Errorf(`%s: is not one of %v`, kv.name, []string{FieldDefault, FieldExample}))
			}

			items = append(items, kv)
			currentTable.codes = append(currentTable.codes, kv.code)
			desc = nil
		}
	}
	if len(globalTable.codes) == 0 {
		// drop it
		items = items[1:]
	}
	if len(desc) > 0 && !isGeneratedPreamble(desc) {
		items = append(items, desc)
	}
	return
}
