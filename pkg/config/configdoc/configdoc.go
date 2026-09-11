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
)

// Generate returns MarkDown documentation generated from the TOML string.
// - Each field but include a trailing comment of either FieldDefault or FieldExample.
// - If a description begins with TokenAdvanced, then a warning will be included.
// - The markdown wil begin with the header, followed by the example
// - Extended descriptions can be applied to top level tables
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
	for _, marker := range []string{FieldDefault, FieldExample} {
		if rest, ok := strings.CutSuffix(line, marker); ok {
			line = strings.TrimSpace(rest)
			break
		}
	}
	return strings.Trim(line, "[]")
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
	if len(desc) > 0 {
		if strings.HasPrefix(strings.TrimSpace(desc[0]), TokenAdvanced) {
			t.adv = true
			t.desc = t.desc[1:]
		}
	}
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
	if len(desc) > 0 {
		if strings.HasPrefix(strings.TrimSpace(desc[0]), TokenAdvanced) {
			t.adv = true
			t.desc = t.desc[1:]
		}
	}
	return t
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

func newKeyval(line string, desc lines) keyval {
	line = strings.TrimSpace(line)
	// A malformed line with no space after the key would otherwise panic.
	name := line
	if i := strings.Index(line, " "); i > -1 {
		name = line[:i]
	}
	kv := keyval{
		name: name,
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
			// A dynamic key — a map entry keyed by data, such as a chain selector
			// or a telemetry label — has no declaring struct field, so it cannot
			// carry a doc comment of its own.
			inherited := false
			if len(desc) == 0 && currentTable != &globalTable && len(currentTable.desc) > 0 {
				desc = currentTable.desc
				inherited = true
			}
			kv := newKeyval(trimmed, desc)
			shortName := kv.name
			if currentTable != &globalTable {
				// update to full name
				kv.name = currentTable.name + "." + kv.name
			}
			switch {
			case inherited:
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
