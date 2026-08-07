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
	// FieldDocsOnly marks a field that is documented but left out of every example - both the
	// document's example config and the code block of the table it belongs to. Use it for a
	// setting that only applies in a mode the examples do not show, so that a reader can copy
	// any example verbatim and get a configuration that works.
	FieldDocsOnly = "# Docs only"

	TokenAdvanced = "**ADVANCED**"
)

// Generate returns MarkDown documentation generated from the TOML string.
// - Each field but include a trailing comment of FieldDefault, FieldExample or FieldDocsOnly.
// - If a description begins with TokenAdvanced, then a warning will be included.
// - The markdown wil begin with the header, followed by the example
// - Extended descriptions can be applied to top level tables
func Generate(toml, header, example string, extendedDescriptions map[string]string) (string, error) {
	return GenerateWith(TOML{}, toml, header, example, extendedDescriptions)
}

// GenerateWith is Generate for a document written in the syntax described by format. Generate
// is GenerateWith(TOML{}, ...).
func GenerateWith(format Format, doc, header, example string, extendedDescriptions map[string]string) (string, error) {
	items, err := parseDocs(format, doc, extendedDescriptions)
	var sb strings.Builder

	sb.WriteString(header)
	sb.WriteString(`
## Example

`)
	sb.WriteString("```")
	sb.WriteString(format.Name())
	sb.WriteString("\n")
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

// lines holds a set of contiguous lines
type lines []string

func (d lines) String() string {
	return strings.Join(d, "\n")
}

type table struct {
	lang     string
	name     string
	codes    lines
	adv      bool
	desc     lines
	extended string
}

func newTable(lang, line, name string, desc lines, extendedDescriptions map[string]string) *table {
	t := &table{
		lang:  lang,
		name:  name,
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

func newArrayOfTables(lang, line, name string, desc lines, extendedDescriptions map[string]string) *table {
	t := &table{
		lang:  lang,
		name:  name,
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
		return fmt.Sprint("```", t.lang, "\n", t.codes, "\n```\n")
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
	lang string
	name string
	code string
	adv  bool
	desc lines
}

func newKeyval(lang, line, name string, desc lines) keyval {
	line = strings.TrimSpace(line)
	kv := keyval{
		lang: lang,
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
		"```", k.lang, "\n",
		k.code,
		"\n```\n",
		k.desc)
}

func parseDocs(format Format, s string, extendedDescriptions map[string]string) (items []fmt.Stringer, err error) {
	defer func() { _, err = config.MultiErrorList(err) }()
	globalTable := table{lang: format.Name(), name: "Global"}
	currentTable := &globalTable
	items = append(items, currentTable)
	var desc lines
	defaultMarker, exampleMarker, docsOnlyMarker := format.DefaultMarker(), format.ExampleMarker(), format.DocsOnlyMarker()
	for line := range strings.SplitSeq(s, "\n") {
		parsed := format.ParseLine(line)
		switch parsed.Kind {
		case LineComment:
			desc = append(desc, parsed.Text)
		case LineBlank:
			if len(desc) > 0 {
				items = append(items, desc)
				desc = nil
			}
		case LineArrayOfTables:
			currentTable = newArrayOfTables(format.Name(), line, parsed.Text, desc, extendedDescriptions)
			items = append(items, currentTable)
			desc = nil
		case LineTable:
			currentTable = newTable(format.Name(), line, parsed.Text, desc, extendedDescriptions)
			items = append(items, currentTable)
			desc = nil
		default:
			kv := newKeyval(format.Name(), line, parsed.Text, desc)
			shortName := kv.name
			if currentTable != &globalTable {
				// update to full name
				kv.name = currentTable.name + "." + kv.name
			}
			if len(kv.desc) == 0 {
				err = errors.Join(err, fmt.Errorf("%s: missing description", kv.name))
			} else if !strings.HasPrefix(kv.desc[0], shortName) {
				err = errors.Join(err, fmt.Errorf("%s: description does not begin with %q", kv.name, shortName))
			}
			docsOnly := strings.HasSuffix(line, docsOnlyMarker)
			if !docsOnly && !strings.HasSuffix(line, defaultMarker) && !strings.HasSuffix(line, exampleMarker) {
				err = errors.Join(err, fmt.Errorf(`%s: is not one of %v`, kv.name, []string{defaultMarker, exampleMarker, docsOnlyMarker}))
			}

			items = append(items, kv)
			// A docs-only field still gets its own entry, but is kept out of the table's code
			// block so every example in the document agrees on what a working config contains.
			if !docsOnly {
				currentTable.codes = append(currentTable.codes, kv.code)
			}
			desc = nil
		}
	}
	if len(globalTable.codes) == 0 {
		// drop it
		items = items[1:]
	}
	if len(desc) > 0 {
		items = append(items, desc)
	}
	return
}
