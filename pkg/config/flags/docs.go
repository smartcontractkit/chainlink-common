package flags

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config/configdoc"
)

// docsOutputPath is where the "docs" subcommand writes generated documentation, mirroring
// chainlink core's docs/CONFIG.md convention.
const docsOutputPath = "docs/CONFIG.md"

// addDocsCommand registers a "docs" subcommand on root (if one isn't already present) that
// writes generated Markdown documentation - covering root's registered config struct and
// every descendant subcommand's registered config struct - to docsOutputPath.
//
// opts.GenerateDoc is recorded on the command so the whole document renders with the generator
// the caller asked for. The first registration wins, since one document can only be rendered
// one way however many targets contribute to it.
func addDocsCommand(root *cobra.Command, opts Options) {
	meta := getOrCreateMeta(root)
	if meta.docFormat == nil {
		meta.docFormat = opts.format()
	}

	for _, c := range root.Commands() {
		if c.Name() == "docs" {
			return
		}
	}

	root.AddCommand(&cobra.Command{
		Use:   "docs",
		Short: "Write generated configuration documentation to " + docsOutputPath,
		// Documenting the config must not require a valid config. Defining any
		// PersistentPreRunE here shadows the root's (cobra runs only the nearest one),
		// which is what skips the decode + `validate:"required"` step for this command.
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := GenerateDocs(cmd.Root())
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(docsOutputPath), 0o755); err != nil {
				return fmt.Errorf("failed to create %s: %w", filepath.Dir(docsOutputPath), err)
			}
			if err := os.WriteFile(docsOutputPath, []byte(doc), 0o600); err != nil {
				return fmt.Errorf("failed to write %s: %w", docsOutputPath, err)
			}
			fmt.Println("Wrote", docsOutputPath)
			return nil
		},
	})
}

// GenerateDocs generates Markdown documentation covering root's registered config struct and
// every descendant subcommand's registered config struct (each under its own namespaced TOML
// table), by deriving TOML doc lines directly from struct tags:
//   - `toml`/`mapstructure` (or the lowercased field name) supplies the key, same convention
//     used for CLI/env/viper binding.
//   - `usage` supplies the field's (or, for a nested struct field, the table's) description.
//   - a field tagged `validate:"required"` (the same tag checked at runtime, see
//     decodeAndApplyProfiles) is documented as configdoc.FieldExample instead of
//     configdoc.FieldDefault, since it has no real default; its `example` tag supplies the
//     placeholder value shown.
//   - a `flagdocs` tag adjusts how the field appears in the example config; see flagDocsTag.
//
// Arrays/slices of structs ([[array-of-tables]] in hand-written docs) are not supported.
//
// Each command's block is preceded by a divider heading - "Global Configuration" for root,
// "Command: <path>" for every subcommand - naming exactly which `myapp ...` invocation the
// fields below it apply to, since root and subcommand fields would otherwise render as
// visually identical tables with no indication of which command binds them.
func GenerateDocs(root *cobra.Command) (string, error) {
	var toml, example strings.Builder

	// The rendering step belongs to the whole document, so it uses the format recorded on the
	// root command; per-entry options still drive how each entry's own fields are read.
	format := Options{}.format()
	if rootMeta := getMeta(root); rootMeta != nil && rootMeta.docFormat != nil {
		format = rootMeta.docFormat
	}

	var walkCmd func(cmd *cobra.Command) error
	walkCmd = func(cmd *cobra.Command) error {
		if meta := getMeta(cmd); meta != nil && len(meta.entries) > 0 {
			label := "Global Configuration"
			if cmd != root {
				label = "Command: " + cmd.CommandPath()
			}

			var cmdToml, cmdExample strings.Builder
			for _, entry := range meta.entries {
				if entry.target == nil {
					continue
				}

				section, err := structToDocs(entry.target, entry.namespace, entry.opts)
				if err != nil {
					return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
				}
				cmdToml.WriteString(section)

				exampleSection, err := exampleDoc(entry.target, entry.namespace, entry.opts)
				if err != nil {
					return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
				}
				cmdExample.WriteString(exampleSection)
			}

			// A comment whose text itself starts with "#" renders as a standalone H1 divider
			// (the format strips its own comment marker, leaving the markdown one), set apart
			// from the H2 table headings structToDocs emits. The blank line after it keeps it
			// from being read as the next item's description.
			toml.WriteString(format.Comment("# "+label) + "\n\n" + cmdToml.String())
			example.WriteString(format.Comment("----- "+label+" -----") + "\n" + cmdExample.String() + "\n")
		}
		for _, sub := range cmd.Commands() {
			if err := walkCmd(sub); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walkCmd(root); err != nil {
		return "", err
	}

	header := fmt.Sprintf("# %s Configuration\n", root.Name())
	return configdoc.GenerateWith(format, toml.String(), header, example.String(), nil)
}

// exampleDoc renders target's fields as plain, directly-usable TOML - real default values,
// and for `validate:"required"` fields (which have none), the placeholder from their
// `example` tag - namespaced the same way structToDocs is, so it can double as a working
// example config file for a namespaced subcommand.
func exampleDoc(target any, namespace string, opts Options) (string, error) {
	var sb strings.Builder
	f := opts.format()
	chosen := newExclusiveChoice()
	currentSection := ""

	if namespace != "" {
		sb.WriteString(f.Table(namespace) + "\n")
		currentSection = namespace
	}

	err := walkStruct(target, opts, structVisitor{
		branch: func(m fieldMeta) (bool, error) {
			if omitFromExample(m.field) || chosen.excluded(m, opts) {
				return true, nil
			}
			chosen.keep(m)

			// A squashed struct contributes no key of its own, so it gets no table: its fields
			// belong to the enclosing one. Checked before the namespace is prefixed, or a
			// squashed struct under a namespace would open an empty "[namespace.]" table.
			section := m.key()
			if section == "" {
				return false, nil
			}
			if namespace != "" {
				section = namespace + "." + section
			}
			if section == currentSection {
				return false, nil
			}
			currentSection = section
			sb.WriteString("\n" + f.Table(section) + "\n")
			return false, nil
		},
		leaf: func(m fieldMeta) error {
			if omitFromExample(m.field) || chosen.excluded(m, opts) {
				return nil
			}
			chosen.keep(m)

			key := m.keyPath[len(m.keyPath)-1]
			value := f.Literal(m.elem.Interface())
			if hasNoRealDefault(m.field) {
				if ex := m.field.Tag.Get("example"); ex != "" {
					value = ex
				}
			}
			if override, ok := exampleOverride(m.field); ok {
				value = override
			}
			sb.WriteString(f.Field(key, value, "") + "\n")
			return nil
		},
	})
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// structToDocs renders target's fields as annotated config-document lines in opts' format. If namespace is
// non-empty, all of target's keys are nested under a "[namespace]" table (and further nested
// structs under "[namespace.sub]", etc) instead of the top-level "Global" bucket - this is how
// a subcommand's settings (bound under that namespace by RegisterSubcommandFlags) are kept
// from colliding with the root command's own tables.
func structToDocs(target any, namespace string, opts Options) (string, error) {
	var sb strings.Builder
	f := opts.format()
	currentSection := ""

	targetType := reflect.TypeOf(target)
	for targetType != nil && targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	if namespace != "" {
		// No description of its own: a namespace is a grouping, and the command's Short
		// describes the command rather than this slice of its configuration. Squashed structs
		// underneath still contribute theirs.
		writeComments(&sb, f, squashedUsages(targetType, opts))
		sb.WriteString(f.Table(namespace) + "\n")
		currentSection = namespace
	} else if usages := squashedUsages(targetType, opts); len(usages) > 0 {
		// No table header to hang them on, so they become standalone prose. A comment block
		// followed by a blank line is a free-standing item to configdoc; without the blank
		// line it would instead be read as the next field's description.
		writeComments(&sb, f, usages)
		sb.WriteString("\n")
	}

	err := walkStruct(target, opts, structVisitor{
		branch: func(m fieldMeta) (bool, error) {
			// A squashed struct contributes no section of its own, so its description would
			// have nowhere to go here; it is folded into the enclosing table's header instead
			// (see squashedUsages), alongside that table's own description. Checked before the
			// namespace is prefixed, or a squashed struct under a namespace would open an empty
			// "[namespace.]" table.
			section := m.key()
			if section == "" {
				return false, nil
			}
			if namespace != "" {
				section = namespace + "." + section
			}
			if section == currentSection {
				return false, nil
			}
			currentSection = section

			writeComments(&sb, f, append(nonEmpty(m.field.Tag.Get("usage")), squashedUsages(m.elemType, opts)...))
			sb.WriteString(f.Table(section) + "\n")
			return false, nil
		},
		leaf: func(m fieldMeta) error {
			key := m.keyPath[len(m.keyPath)-1]

			desc := m.field.Tag.Get("usage")
			switch {
			case desc == "":
				desc = key
			case !strings.HasPrefix(desc, key):
				desc = key + " " + desc
			}
			sb.WriteString(f.Comment(desc+constraintNote(m, opts)) + "\n")

			marker := f.DefaultMarker()
			value := f.Literal(m.elem.Interface())
			if hasNoRealDefault(m.field) {
				marker = f.ExampleMarker()
				if ex := m.field.Tag.Get("example"); ex != "" {
					value = ex
				}
			}
			if omitFromExample(m.field) {
				// Documented, but absent from the example config and from its own table's
				// code block, so every example in the document describes the same setup.
				marker = f.DocsOnlyMarker()
			}

			sb.WriteString(f.Field(key, value, marker) + "\n\n")
			return nil
		},
	})
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// isRequired reports whether field carries a `required` validate rule - either the plain
// `required` or one of validator's conditional forms (`required_without=Other`,
// `required_if=...`, etc). Conditionally-required fields have no usable default either, so
// they're documented as FieldExample the same way.
// squashedUsages returns the `usage` descriptions of t's squashed struct fields, recursively.
// Squashed structs are flattened into their parent and get no table of their own, so their
// descriptions are reported as part of the enclosing table's header - otherwise they'd be lost,
// and emitting them next to their fields would make configdoc read them as the following
// field's description.
func squashedUsages(t reflect.Type, opts Options) []string {
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}

	var usages []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		elemType := field.Type
		if elemType.Kind() == reflect.Pointer {
			elemType = elemType.Elem()
		}
		if elemType.Kind() != reflect.Struct || implementsTextUnmarshaler(elemType) {
			continue
		}
		if _, squash := opts.tagKey(field); !squash {
			continue
		}

		usages = append(usages, nonEmpty(field.Tag.Get("usage"))...)
		usages = append(usages, squashedUsages(elemType, opts)...)
	}
	return usages
}

// exclusiveChoice tracks which of a set of mutually exclusive fields the example already
// shows. An example config has to be one that actually works, and "exactly one of these"
// cannot be illustrated by listing all of them - so the first one declared wins and the rest
// are left out. The docs still describe every option; only the example has to choose.
//
// Fields are recorded by their full config key rather than by the Go struct and field name
// declaring them, so the rules resolve the same way validator does at runtime: a squashed
// struct's fields are promoted into its parent's table, and a rule's parameter may be a dotted
// path reaching down into a nested struct's own table.
type exclusiveChoice map[string]bool

func newExclusiveChoice() exclusiveChoice { return exclusiveChoice{} }

// keep records that the example shows this field.
func (c exclusiveChoice) keep(m fieldMeta) {
	c[strings.Join(m.keyPath, ".")] = true
}

// excluded reports whether a field the example already shows rules this one out.
func (c exclusiveChoice) excluded(m fieldMeta, opts Options) bool {
	if len(c) == 0 {
		return false
	}
	for _, rule := range strings.Split(m.field.Tag.Get("validate"), ",") {
		name, args, hasArgs := strings.Cut(rule, "=")
		if !hasArgs || !strings.HasPrefix(name, "excluded_with") {
			continue
		}
		for _, goPath := range strings.Fields(args) {
			if key, ok := ruleTargetKey(m, goPath, opts); ok && c[key] {
				return true
			}
		}
	}
	return false
}

// ruleTargetKey resolves a cross-field rule's parameter to the full config key of the field it
// names, in the same terms keep records: validator resolves the parameter from the struct
// declaring the rule, so the key is that struct's own key path followed by the resolved path.
func ruleTargetKey(m fieldMeta, goPath string, opts Options) (string, bool) {
	keys, ok := siblingKeyPath(m.parent, goPath, opts)
	if !ok {
		return "", false
	}

	// The declaring struct's key path is m's without its own segment - unless m is itself
	// squashed and so contributed none.
	prefix := m.keyPath
	if _, squash := opts.tagKey(m.field); !squash && len(prefix) > 0 {
		prefix = prefix[:len(prefix)-1]
	}

	full := make([]string, 0, len(prefix)+len(keys))
	full = append(full, prefix...)
	full = append(full, keys...)
	return strings.Join(full, "."), true
}

// writeComments emits each line as a comment in f's syntax.
func writeComments(sb *strings.Builder, f configdoc.Format, lines []string) {
	for _, l := range lines {
		sb.WriteString(f.Comment(l) + "\n")
	}
}

// nonEmpty returns s as a one-element slice, or nothing if s is empty.
func nonEmpty(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

// The `flagdocs` tag adjusts how a field appears in the generated example config, without
// changing the field itself or how it is documented. It holds one directive:
//
//	flagdocs:"noexample"        leave the field out of every example (see configdoc.FieldDocsOnly)
//	flagdocs:"example=<value>"  use this literal in the example instead of the field's default
//
// Use noexample for a setting the example must not carry - one that only applies in a mode the
// example isn't showing - where including it would describe a setup nobody should copy. Use
// example= where the field's own default would make the example wrong or unhelpful.
//
// The directive is taken whole, so an example= literal may itself contain commas.
const flagDocsTag = "flagdocs"

// omitFromExample reports whether a field is kept out of the example config.
func omitFromExample(field reflect.StructField) bool {
	return field.Tag.Get(flagDocsTag) == "noexample"
}

// exampleOverride returns the literal to show for a field in the example config, if the field
// asked for one.
func exampleOverride(field reflect.StructField) (string, bool) {
	const prefix = "example="
	tag := field.Tag.Get(flagDocsTag)
	if !strings.HasPrefix(tag, prefix) {
		return "", false
	}
	return strings.TrimPrefix(tag, prefix), true
}

// hasNoRealDefault reports whether a field's zero value is a placeholder rather than a usable
// default, so the docs should show it as configdoc.FieldExample. That's true when the field is
// required (its value must be supplied), and also whenever it carries an `example` tag - which
// is how a field says "there is nothing sensible to default to" even when the rule making it
// required lives outside this struct and so can't be a `validate` tag.
func hasNoRealDefault(field reflect.StructField) bool {
	return isRequired(field) || field.Tag.Get("example") != ""
}

func isRequired(field reflect.StructField) bool {
	for _, rule := range strings.Split(field.Tag.Get("validate"), ",") {
		if name, _, _ := strings.Cut(rule, "="); name == "required" || strings.HasPrefix(name, "required_") {
			return true
		}
	}
	return false
}

// crossFieldRules are the validator rules whose outcome depends on a sibling field, and which
// therefore need a field to be distinguishably "absent" rather than merely zero.
var crossFieldRules = map[string]bool{
	"required_with": true, "required_with_all": true,
	"required_without": true, "required_without_all": true,
	"excluded_with": true, "excluded_with_all": true,
	"excluded_without": true, "excluded_without_all": true,
}

// crossFieldRuleNames returns the cross-field rule names present on field, in tag order.
func crossFieldRuleNames(field reflect.StructField) []string {
	var names []string
	for _, rule := range strings.Split(field.Tag.Get("validate"), ",") {
		if name, _, hasArgs := strings.Cut(rule, "="); hasArgs && crossFieldRules[name] {
			names = append(names, name)
		}
	}
	return names
}

// constraintNote renders a field's cross-field `validate` rules as a human-readable sentence
// (e.g. "must not be set unless database-url is set"), so the generated docs explain when a
// conditionally required or mutually exclusive setting actually applies. Sibling fields named
// by the rules are reported by their config key rather than their Go field name.
func constraintNote(m fieldMeta, opts Options) string {
	var notes []string
	for _, rule := range strings.Split(m.field.Tag.Get("validate"), ",") {
		name, args, hasArgs := strings.Cut(rule, "=")
		if !hasArgs {
			continue
		}

		var keys []string
		for _, goName := range strings.Fields(args) {
			keys = append(keys, siblingKey(m.parent, goName, opts))
		}
		list := strings.Join(keys, ", ")

		switch name {
		case "required_with":
			notes = append(notes, "required when "+list+" is set")
		case "required_with_all":
			notes = append(notes, "required when all of "+list+" are set")
		case "required_without":
			notes = append(notes, "required unless "+list+" is set")
		case "required_without_all":
			notes = append(notes, "required unless all of "+list+" are set")
		case "excluded_with":
			notes = append(notes, "must not be set when "+list+" is set")
		case "excluded_with_all":
			notes = append(notes, "must not be set when all of "+list+" are set")
		case "excluded_without":
			notes = append(notes, "must not be set unless "+list+" is set")
		case "excluded_without_all":
			notes = append(notes, "must not be set unless all of "+list+" are set")
		}
	}
	if len(notes) == 0 {
		return ""
	}
	return " (" + strings.Join(notes, "; ") + ")"
}

// siblingKey maps a Go field path within parent to its dotted config key, falling back to the Go
// path if it doesn't resolve.
func siblingKey(parent reflect.Type, goPath string, opts Options) string {
	if keys, ok := siblingKeyPath(parent, goPath, opts); ok {
		return strings.Join(keys, ".")
	}
	return goPath
}

// siblingKeyPath resolves a cross-field rule's parameter - a Go field name, or a dotted path into
// a nested struct ("Mid.Deep.Bar"), the way go-playground/validator resolves it from the struct
// carrying the rule - into the config keys of the fields along it. A squashed struct contributes
// no key of its own, matching how its fields are bound.
func siblingKeyPath(parent reflect.Type, goPath string, opts Options) ([]string, bool) {
	current := parent
	var keys []string
	for _, name := range strings.Split(goPath, ".") {
		if current == nil || current.Kind() != reflect.Struct {
			return nil, false
		}
		field, ok := current.FieldByName(name)
		if !ok {
			return nil, false
		}

		if key, squash := opts.tagKey(field); !squash {
			keys = append(keys, key)
		}

		current = field.Type
		for current.Kind() == reflect.Pointer {
			current = current.Elem()
		}
	}
	if len(keys) == 0 {
		return nil, false
	}
	return keys, true
}

