package commentparsing

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func writePackage(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	return dir
}

func TestParseDir(t *testing.T) {
	t.Run("Comments and validate tags", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

// Config configures the thing.
type Config struct {
	// Endpoint is the upstream URL to dial.
	// It must include a scheme.
	Endpoint string ` + "`toml:\"endpoint\" validate:\"required,url\"`" + `

	Timeout int ` + "`toml:\"timeout\"`" + ` // Timeout bounds a single request.
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Equal(t, "example", pkg.Name)
		require.Len(t, pkg.Types, 1)
		require.Equal(t, "Config", pkg.Types[0].Name)

		require.Equal(t, FieldDoc{
			Comment:  "Endpoint is the upstream URL to dial.\nIt must include a scheme.",
			Validate: "required,url",
		}, pkg.Types[0].Fields["Endpoint"])

		require.Equal(t, FieldDoc{Comment: "Timeout bounds a single request."}, pkg.Types[0].Fields["Timeout"])
	})

	t.Run("Embedded fields keyed by the name reflect reports", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

type Config struct {
	// Local is embedded by value.
	Local
	// Remote is embedded through a pointer to another package's type.
	*other.Remote
	Plain string ` + "`toml:\"plain\"`" + `
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Contains(t, pkg.Types[0].Fields, "Local")
		require.Contains(t, pkg.Types[0].Fields, "Remote")
	})

	t.Run("Generic embeds resolve to their base name", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

type Config struct {
	Single[int]
	Pair[string, int]
	*other.Remote[int]
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Contains(t, pkg.Types[0].Fields, "Single")
		require.Contains(t, pkg.Types[0].Fields, "Pair")
		require.Contains(t, pkg.Types[0].Fields, "Remote")
	})

	t.Run("One comment covers every name a field declares", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

type Config struct {
	// Host and Port address the upstream.
	Host, Port string ` + "`toml:\"-\"`" + `
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Equal(t, pkg.Types[0].Fields["Host"], pkg.Types[0].Fields["Port"])
		require.Equal(t, "Host and Port address the upstream.", pkg.Types[0].Fields["Host"].Comment)
	})

	// An embed alone qualifies, because Render descends through it and would report a struct that
	// was skipped as undocumented.
	t.Run("Structs with nothing to document are skipped, embeds are not", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

type internalState struct {
	mu    int
	cache map[string]string
}

type Wrapper struct {
	Config
}

type Config struct {
	// Endpoint is documented.
	Endpoint string
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		names := []string{}
		for _, typ := range pkg.Types {
			names = append(names, typ.Name)
		}
		require.Equal(t, []string{"Config", "Wrapper"}, names)
	})

	t.Run("Declarations that are not struct types", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

const Version = "1"

// Name is a documented alias, not a struct.
type Name = string

// Handler is a documented interface, not a struct.
type Handler interface{ Handle() }

type Config struct {
	// Endpoint is documented.
	Endpoint string
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Len(t, pkg.Types, 1)
		require.Equal(t, "Config", pkg.Types[0].Name)
	})

	t.Run("Test files are not read", func(t *testing.T) {
		dir := writePackage(t, map[string]string{
			"config.go":      "package example\n\ntype Config struct {\n\t// Endpoint is documented.\n\tEndpoint string\n}\n",
			"config_test.go": "package example\n\ntype TestOnly struct {\n\t// Field is documented.\n\tField string\n}\n",
		})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Len(t, pkg.Types, 1)
		require.Equal(t, "Config", pkg.Types[0].Name)
	})

	// A tag needs no particular shape to compile, so one that is not key:value form reaches the
	// parser intact and simply yields no validate rule.
	t.Run("Tag that is not in key:value form", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

type Config struct {
	// Endpoint is documented.
	Endpoint string ` + "`not a tag at all`" + `
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Equal(t, FieldDoc{Comment: "Endpoint is documented."}, pkg.Types[0].Fields["Endpoint"])
	})

	t.Run("Missing directory", func(t *testing.T) {
		_, err := ParseDir(filepath.Join(t.TempDir(), "absent"))
		require.Error(t, err)
	})

	t.Run("Directory holding no Go package", func(t *testing.T) {
		_, err := ParseDir(writePackage(t, map[string]string{"README.md": "not go"}))
		require.ErrorContains(t, err, "no Go package")
	})

	t.Run("Unparseable source", func(t *testing.T) {
		_, err := ParseDir(writePackage(t, map[string]string{"config.go": "package example\n\ntype Config struct {"}))
		require.Error(t, err)
	})
}

// TestParserInvariantGuards reaches three branches that ParseDir cannot, because go/parser rejects
// every input that would produce them and ParseDir returns on a parse error: an embedded type that
// is not a name, a string literal it would not accept, and a type declaration holding anything but
// a type spec. Each is exercised on a hand-built node so the guard is pinned rather than merely
// believed, and so a future parser that accepts more does not turn one into a panic.
func TestParserInvariantGuards(t *testing.T) {
	t.Run("Embedded type that is not a name", func(t *testing.T) {
		require.Nil(t, fieldNames(&ast.Field{Type: &ast.ArrayType{Elt: ast.NewIdent("int")}}))
	})

	t.Run("Tag literal the parser would have rejected", func(t *testing.T) {
		badLiteral := `"\q"` // an unknown escape: "unknown escape sequence" from the scanner
		require.Empty(t, structTag(&ast.Field{Tag: &ast.BasicLit{Kind: token.STRING, Value: badLiteral}}))
	})

	t.Run("Type declaration holding a non-type spec", func(t *testing.T) {
		file := &ast.File{Name: ast.NewIdent("example"), Decls: []ast.Decl{&ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`}}},
		}}}
		require.Empty(t, typesIn(file))
	})
}

func TestGenerate(t *testing.T) {
	pkg := &Package{
		Name: "example",
		Types: []Type{{
			Name: "Config",
			Fields: map[string]FieldDoc{
				"Endpoint": {Comment: "Endpoint is the upstream URL.\nIt must include a scheme.", Validate: "required,url"},
				"Timeout":  {Comment: "Timeout bounds a single request."},
				"Bare":     {},
			},
		}},
	}

	t.Run("Renders a method per type", func(t *testing.T) {
		source, err := Generate(pkg)
		require.NoError(t, err)
		got := string(source)

		require.Contains(t, got, "// Code generated by "+selfImportPath+". DO NOT EDIT.")
		require.Contains(t, got, "package example")
		require.Contains(t, got, "import \""+selfImportPath+"\"")
		require.Contains(t, got, "func (Config) DocComments() (string, map[string]commentparsing.FieldDoc) {")
		require.Contains(t, got, `Validate: "required,url"`)
		require.Contains(t, got, `"Endpoint is the upstream URL.\nIt must include a scheme."`)
	})

	// A multi-line comment or a quote inside one would break the output if it were not escaped,
	// and a broken file is only discovered at the consumer's next build.
	t.Run("Output is valid Go", func(t *testing.T) {
		source, err := Generate(pkg)
		require.NoError(t, err)
		_, err = parser.ParseFile(token.NewFileSet(), GeneratedFileName, source, parser.SkipObjectResolution)
		require.NoError(t, err)
	})

	t.Run("Same input renders the same bytes", func(t *testing.T) {
		first, err := Generate(pkg)
		require.NoError(t, err)
		second, err := Generate(pkg)
		require.NoError(t, err)
		require.Equal(t, first, second)
	})

	t.Run("This package documents itself unqualified", func(t *testing.T) {
		source, err := Generate(&Package{Name: "commentparsing", Types: pkg.Types})
		require.NoError(t, err)
		require.NotContains(t, string(source), "import")
		require.Contains(t, string(source), "map[string]FieldDoc")
	})

	t.Run("Nothing to document", func(t *testing.T) {
		_, err := Generate(&Package{Name: "example"})
		require.ErrorIs(t, err, ErrNoTypes)
	})
}

func TestWriteFile(t *testing.T) {
	source := "package example\n\ntype Config struct {\n\t// Endpoint is documented.\n\tEndpoint string `toml:\"endpoint\"`\n}\n"

	t.Run("Writes beside the source it read", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": source})
		require.NoError(t, WriteFile(dir))

		written, err := os.ReadFile(filepath.Join(dir, GeneratedFileName))
		require.NoError(t, err)
		require.Contains(t, string(written), "func (Config) DocComments()")
	})

	// Regenerating has to be idempotent or a CI diff would flag a package nobody touched.
	t.Run("Regenerating is a no-op", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": source})
		require.NoError(t, WriteFile(dir))
		first, err := os.ReadFile(filepath.Join(dir, GeneratedFileName))
		require.NoError(t, err)

		require.NoError(t, WriteFile(dir))
		second, err := os.ReadFile(filepath.Join(dir, GeneratedFileName))
		require.NoError(t, err)
		require.Equal(t, first, second)
	})

	// A file left behind would answer for types that no longer exist.
	t.Run("A package that stops documenting loses its file", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": source})
		require.NoError(t, WriteFile(dir))

		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.go"), []byte("package example\n"), 0o600))
		require.NoError(t, WriteFile(dir))
		_, err := os.Stat(filepath.Join(dir, GeneratedFileName))
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("A hand-written file under that name is left alone", func(t *testing.T) {
		dir := writePackage(t, map[string]string{
			"config.go":       "package example\n",
			GeneratedFileName: "package example\n\n// mine, not generated\n",
		})
		require.ErrorContains(t, WriteFile(dir), "remove it by hand")
	})

	t.Run("A package that documents nothing and never had a file", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": "package example\n"})
		require.NoError(t, WriteFile(dir))
		_, err := os.Stat(filepath.Join(dir, GeneratedFileName))
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("Output path that cannot be read", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": "package example\n"})
		require.NoError(t, os.Mkdir(filepath.Join(dir, GeneratedFileName), 0o750))
		require.Error(t, WriteFile(dir))
	})

	t.Run("Nothing to parse", func(t *testing.T) {
		require.Error(t, WriteFile(filepath.Join(t.TempDir(), "absent")))
	})
}

type lookupTarget struct {
	Endpoint string
}

func (lookupTarget) DocComments() (string, map[string]FieldDoc) {
	return "lookupTarget", map[string]FieldDoc{"Endpoint": {Comment: "Endpoint is the upstream URL."}}
}

type lookupUndocumented struct {
	Endpoint string
}

// lookupPromoted has no DocComments of its own and reaches lookupTarget's by promotion, which is
// the silent failure [DocCommenter] returns a type name to expose.
type lookupPromoted struct {
	lookupTarget
}

func TestLookup(t *testing.T) {
	t.Run("Documented type", func(t *testing.T) {
		fields, err := Lookup(reflect.TypeFor[lookupTarget]())
		require.NoError(t, err)
		require.Equal(t, "Endpoint is the upstream URL.", fields["Endpoint"].Comment)
	})

	t.Run("Pointer to a documented type", func(t *testing.T) {
		fields, err := Lookup(reflect.TypeFor[**lookupTarget]())
		require.NoError(t, err)
		require.Contains(t, fields, "Endpoint")
	})

	t.Run("Undocumented type names the package to regenerate", func(t *testing.T) {
		_, err := Lookup(reflect.TypeFor[lookupUndocumented]())
		require.ErrorContains(t, err, "has no DocComments method")
		require.ErrorContains(t, err, "commentparsing.lookupUndocumented")
		require.ErrorContains(t, err, "go generate")
	})

	t.Run("Promoted method is reported rather than trusted", func(t *testing.T) {
		_, err := Lookup(reflect.TypeFor[lookupPromoted]())
		require.ErrorContains(t, err, "promotes DocComments from embedded lookupTarget")
	})

	t.Run("Type with no package path", func(t *testing.T) {
		_, err := Lookup(reflect.TypeFor[struct{ Endpoint string }]())
		require.ErrorContains(t, err, "has no DocComments method")
	})
}

func TestGenerateDocCommentFiles(t *testing.T) {
	source := "package example\n\ntype Config struct {\n\t// Endpoint is documented.\n\tEndpoint string `toml:\"endpoint\"`\n}\n"

	t.Run("Writes into every discovered directory", func(t *testing.T) {
		first := writePackage(t, map[string]string{"config.go": source})
		second := writePackage(t, map[string]string{"config.go": source})
		docs := &Docs{dirs: map[string]string{"example.com/first": first, "example.com/second": second}}

		require.NoError(t, docs.GenerateDocCommentFiles())
		for _, dir := range []string{first, second} {
			written, err := os.ReadFile(filepath.Join(dir, GeneratedFileName))
			require.NoError(t, err)
			require.Contains(t, string(written), "func (Config) DocComments()")
		}
	})

	// Every directory is reported, not just the first to fail, so one broken package does not hide
	// the rest.
	t.Run("Failures name the package", func(t *testing.T) {
		docs := &Docs{dirs: map[string]string{"example.com/absent": filepath.Join(t.TempDir(), "absent")}}
		require.ErrorContains(t, docs.GenerateDocCommentFiles(), "example.com/absent")
	})
}

func TestDiscoverDependencyWithoutGeneratedMethods(t *testing.T) {
	// reflect.StructField is a struct from another module with no DocComments, which is the gap
	// generation exists to close, so Discover names it instead of documenting it as blank.
	type root struct {
		// Foreign is a type this module does not declare.
		Foreign reflect.StructField `toml:"foreign"`
	}

	_, err := Discover(&root{})
	require.ErrorContains(t, err, "has no DocComments method")
	require.ErrorContains(t, err, "go generate")
}

func TestDiscoverEdges(t *testing.T) {
	// A nil root reaches the walk as a nil reflect.Type, which used to panic there.
	t.Run("Nil root", func(t *testing.T) {
		docs, err := Discover(nil)
		require.NoError(t, err)
		require.Empty(t, docs.Packages())
	})

	// An anonymous struct belongs to no package, so there is nothing to parse or generate for it.
	t.Run("Anonymous struct root", func(t *testing.T) {
		docs, err := Discover(&struct{ Plain string }{})
		require.NoError(t, err)
		require.Empty(t, docs.Packages())
	})

	t.Run("Non-struct root", func(t *testing.T) {
		docs, err := Discover("not a config")
		require.NoError(t, err)
		require.Empty(t, docs.Packages())
	})
}

func TestDocsInternals(t *testing.T) {
	t.Run("A dependency with generated methods is recorded", func(t *testing.T) {
		docs := &Docs{fields: map[string]map[string]FieldDoc{}}
		require.NoError(t, docs.addGenerated(reflect.TypeFor[lookupTarget]()))
		require.Equal(t, "Endpoint is the upstream URL.", docs.Field(reflect.TypeFor[lookupTarget](), "Endpoint").Comment)
	})

	t.Run("A dependency without them is reported", func(t *testing.T) {
		docs := &Docs{fields: map[string]map[string]FieldDoc{}}
		require.Error(t, docs.addGenerated(reflect.TypeFor[lookupUndocumented]()))
	})

	t.Run("An unparseable directory names the package", func(t *testing.T) {
		docs := &Docs{fields: map[string]map[string]FieldDoc{}}
		err := docs.addParsed("example.com/absent", filepath.Join(t.TempDir(), "absent"))
		require.ErrorContains(t, err, "example.com/absent")
	})
}

func TestModulePath(t *testing.T) {
	t.Run("Module directive", func(t *testing.T) {
		path, err := modulePath([]byte("module example.com/x\n\ngo 1.26\n"))
		require.NoError(t, err)
		require.Equal(t, "example.com/x", path)
	})

	t.Run("No module directive", func(t *testing.T) {
		_, err := modulePath([]byte("go 1.26\n"))
		require.ErrorContains(t, err, "no module directive")
	})
}

// An interface field hands a walk a nil type, which must be a reported miss rather than a panic.
func TestDocsNilType(t *testing.T) {
	docs := &Docs{fields: map[string]map[string]FieldDoc{}}
	_, err := docs.Type(nil)
	require.ErrorContains(t, err, "nil type")
	require.Zero(t, docs.Field(nil, "Anything"))
}
