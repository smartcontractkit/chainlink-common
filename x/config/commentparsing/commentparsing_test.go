package commentparsing

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	t.Run("Doc comments, above and trailing", func(t *testing.T) {
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
		require.Equal(t, dir, pkg.Dir)

		typ, ok := pkg.Type("Config")
		require.True(t, ok)
		require.Equal(t, "Endpoint is the upstream URL to dial.\nIt must include a scheme.", typ.Fields["Endpoint"].Comment)
		require.Equal(t, "Timeout bounds a single request.", typ.Fields["Timeout"].Comment)
	})

	// Struct tags survive compilation, so a consumer reads them off the type and they are not
	// carried here.
	t.Run("Tags are not collected", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

type Config struct {
	// Endpoint is documented.
	Endpoint string ` + "`toml:\"endpoint\" validate:\"required,url\"`" + `
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		typ, _ := pkg.Type("Config")
		require.Equal(t, FieldDoc{Comment: "Endpoint is documented."}, typ.Fields["Endpoint"])
	})

	// The caller walking a real config value knows which types it reached, so the parser makes no
	// judgement about which take part.
	t.Run("Every struct is returned, documented or not", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

type internalState struct {
	mu int
}

type Config struct {
	// Endpoint is documented.
	Endpoint string
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Len(t, pkg.Types, 2)
		require.Equal(t, "Config", pkg.Types[0].Name)
		require.Equal(t, "internalState", pkg.Types[1].Name)
	})

	t.Run("Embedded fields keyed by the name reflect reports", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

type Config struct {
	// Local is embedded by value.
	Local
	// Remote is embedded through a pointer to another package's type.
	*other.Remote
	Single[int]
	Pair[string, int]
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		typ, _ := pkg.Type("Config")
		for _, name := range []string{"Local", "Remote", "Single", "Pair"} {
			require.Contains(t, typ.Fields, name)
		}
	})

	t.Run("One comment covers every name a field declares", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

type Config struct {
	// Host and Port address the upstream.
	Host, Port string
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		typ, _ := pkg.Type("Config")
		require.Equal(t, typ.Fields["Host"], typ.Fields["Port"])
	})

	t.Run("Declarations that are not struct types", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"config.go": `package example

const Version = "1"

type Name = string

type Handler interface{ Handle() }

type Config struct {
	// Endpoint is documented.
	Endpoint string
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Len(t, pkg.Types, 1)
	})

	t.Run("Test files and prior output are not read", func(t *testing.T) {
		dir := writePackage(t, map[string]string{
			"config.go":       "package example\n\ntype Config struct{ Endpoint string }\n",
			"config_test.go":  "package example\n\ntype TestOnly struct{ Field string }\n",
			GeneratedFileName: "package example\n\ntype Leftover struct{ Field string }\n",
		})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Len(t, pkg.Types, 1)
		require.Equal(t, "Config", pkg.Types[0].Name)
	})

	t.Run("A type the package does not declare", func(t *testing.T) {
		pkg, err := ParseDir(writePackage(t, map[string]string{"config.go": "package example\n"}))
		require.NoError(t, err)
		_, ok := pkg.Type("Absent")
		require.False(t, ok)
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

// TestParserInvariantGuards reaches two branches ParseDir cannot, because go/parser rejects every
// input that would produce them and ParseDir returns on a parse error: an embedded type that is
// not a name, and a type declaration holding anything but a type spec. Each is exercised on a
// hand-built node so the guard is pinned rather than merely believed.
func TestParserInvariantGuards(t *testing.T) {
	t.Run("Embedded type that is not a name", func(t *testing.T) {
		require.Nil(t, fieldNames(&ast.Field{Type: &ast.ArrayType{Elt: ast.NewIdent("int")}}))
	})

	t.Run("Type declaration holding a non-type spec", func(t *testing.T) {
		file := &ast.File{Name: ast.NewIdent("example"), Decls: []ast.Decl{&ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`}}},
		}}}
		require.Empty(t, typesIn(file))
	})
}

func TestDocCommentsFile(t *testing.T) {
	pkg := Package{
		ImportPath: "example.com/app/config",
		Name:       "config",
		Dir:        "config",
		Types: []Type{{
			Name: "Config",
			Fields: map[string]FieldDoc{
				"Endpoint": {Comment: "Endpoint is the upstream URL.\nIt must include a scheme."},
				"Timeout":  {Comment: "Timeout bounds a single request."},
				"Bare":     {},
			},
		}},
	}

	t.Run("Renders a method per type", func(t *testing.T) {
		got := docCommentsFile(pkg)
		require.Contains(t, got, "package config")
		require.Contains(t, got, "import \""+selfImportPath+"\"")
		require.Contains(t, got, "func (Config) DocComments() (string, map[string]commentparsing.FieldDoc) {")
		require.Contains(t, got, `return "Config", map[string]commentparsing.FieldDoc{`)
		require.Contains(t, got, `"Endpoint is the upstream URL.\nIt must include a scheme."`)
	})

	// codegen adds the header and formats every file, so a second one here would be duplicated.
	t.Run("No generated-by header of its own", func(t *testing.T) {
		require.NotContains(t, docCommentsFile(pkg), "Code generated by")
	})

	t.Run("A field whose comment compilation kept nothing from is left out", func(t *testing.T) {
		require.NotContains(t, docCommentsFile(pkg), `"Bare"`)
	})

	t.Run("Output is valid Go", func(t *testing.T) {
		_, err := parser.ParseFile(token.NewFileSet(), GeneratedFileName, docCommentsFile(pkg), parser.SkipObjectResolution)
		require.NoError(t, err)
	})

	t.Run("Same input renders the same bytes", func(t *testing.T) {
		require.Equal(t, docCommentsFile(pkg), docCommentsFile(pkg))
	})

	t.Run("This package documents itself unqualified", func(t *testing.T) {
		self := pkg
		self.ImportPath, self.Name = selfImportPath, "commentparsing"
		got := docCommentsFile(self)
		require.NotContains(t, got, "import")
		require.Contains(t, got, "map[string]FieldDoc")
	})
}

func TestDocCommentsGenerator(t *testing.T) {
	t.Run("One file per package, in its own directory", func(t *testing.T) {
		files, err := docComments([]Package{
			{ImportPath: "example.com/app", Name: "app", Dir: ".", Types: []Type{{Name: "Config"}}},
			{ImportPath: "example.com/app/sub", Name: "sub", Dir: "sub", Types: []Type{{Name: "Sub"}}},
		})
		require.NoError(t, err)
		require.Len(t, files, 2)
		require.Contains(t, files, GeneratedFileName)
		require.Contains(t, files, filepath.Join("sub", GeneratedFileName))
	})

	// A dependency has no directory this module may write to, and must already have generated for
	// Discover to have resolved it.
	t.Run("A package with no local directory is skipped", func(t *testing.T) {
		files, err := docComments([]Package{
			{ImportPath: "example.com/dependency", Types: []Type{{Name: "Settings"}}},
		})
		require.NoError(t, err)
		require.Empty(t, files)
	})

	t.Run("A package with no reached types is skipped", func(t *testing.T) {
		files, err := docComments([]Package{{ImportPath: "example.com/app", Dir: "."}})
		require.NoError(t, err)
		require.Empty(t, files)
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

	// An interface field hands a walk a nil type, which must be reported rather than panic.
	t.Run("Nil type", func(t *testing.T) {
		_, err := Lookup(nil)
		require.ErrorContains(t, err, "nil type")
	})
}

func TestDiscoverEdges(t *testing.T) {
	dir := writePackage(t, map[string]string{"go.mod": "module example.com/x\n\ngo 1.26\n"})

	// A nil root reaches the walk as a nil reflect.Type, which used to panic there.
	t.Run("Nil root", func(t *testing.T) {
		pkgs, err := Discover(dir, nil)
		require.NoError(t, err)
		require.Empty(t, pkgs)
	})

	t.Run("No roots at all", func(t *testing.T) {
		pkgs, err := Discover(dir)
		require.NoError(t, err)
		require.Empty(t, pkgs)
	})

	// An anonymous struct belongs to no package, so there is nothing to parse or generate for it.
	t.Run("Anonymous struct root", func(t *testing.T) {
		pkgs, err := Discover(dir, &struct{ Plain string }{})
		require.NoError(t, err)
		require.Empty(t, pkgs)
	})

	t.Run("Non-struct root", func(t *testing.T) {
		pkgs, err := Discover(dir, "not a config")
		require.NoError(t, err)
		require.Empty(t, pkgs)
	})

	// reflect.StructField is a struct from another module with no DocComments, which is the gap
	// generation exists to close, so Discover names it instead of documenting it as blank.
	t.Run("Dependency that never generated", func(t *testing.T) {
		_, err := Discover(dir, &struct{ Foreign reflect.StructField }{})
		require.ErrorContains(t, err, "has no DocComments method")
		require.ErrorContains(t, err, "go generate")
	})

	t.Run("A dependency that did generate is recorded", func(t *testing.T) {
		pkgs, err := Discover(dir, &lookupTarget{})
		require.NoError(t, err)
		require.Len(t, pkgs, 1)
		require.Equal(t, selfImportPath, pkgs[0].ImportPath)
		require.Empty(t, pkgs[0].Dir, "a dependency has no directory this module may write to")
		require.Equal(t, "Endpoint is the upstream URL.", pkgs[0].Types[0].Fields["Endpoint"].Comment)
	})

	t.Run("No enclosing module", func(t *testing.T) {
		_, err := Discover(filepath.Join(t.TempDir(), "nowhere"))
		require.Error(t, err)
	})
}

func TestRun(t *testing.T) {
	module := map[string]string{"go.mod": "module example.com/x\n\ngo 1.26\n"}

	t.Run("Writes what a generator returns, formatted and headed", func(t *testing.T) {
		dir := writePackage(t, module)
		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"},
			Generator(func([]Package) (map[string]string, error) {
				return map[string]string{"out/thing.go": "package out\nfunc  Thing ()  {}\n"}, nil
			})))

		written, err := os.ReadFile(filepath.Join(dir, "out", "thing.go"))
		require.NoError(t, err)
		require.Contains(t, string(written), "// Code generated by example.com/x/gen, DO NOT EDIT.")
		require.Contains(t, string(written), "func Thing() {}", "codegen gofmts every file it writes")
	})

	t.Run("Every generator's files are written", func(t *testing.T) {
		dir := writePackage(t, module)
		file := func(name string) Generator {
			return Generator(func([]Package) (map[string]string, error) {
				return map[string]string{name: "content\n"}, nil
			})
		}
		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "gen"}, file("first.md"), file("second.md")))
		for _, name := range []string{"first.md", "second.md"} {
			_, err := os.Stat(filepath.Join(dir, name))
			require.NoError(t, err)
		}
	})

	// Whichever generator ran last would silently win, and only reading the output would show it.
	t.Run("Two generators claiming one path", func(t *testing.T) {
		dir := writePackage(t, module)
		same := func() Generator {
			return Generator(func([]Package) (map[string]string, error) {
				return map[string]string{"clash.md": "content\n"}, nil
			})
		}
		err := Run(RunArgs{Dir: dir, Tool: "gen"}, same(), same())
		require.ErrorContains(t, err, "written by more than one generator")
	})

	t.Run("A generator's failure stops the write", func(t *testing.T) {
		dir := writePackage(t, module)
		err := Run(RunArgs{Dir: dir, Tool: "gen"}, Generator(func([]Package) (map[string]string, error) {
			return nil, os.ErrInvalid
		}))
		require.ErrorIs(t, err, os.ErrInvalid)

		entries, readErr := os.ReadDir(dir)
		require.NoError(t, readErr)
		require.Len(t, entries, 1, "only the go.mod it started with")
	})

	t.Run("Discovery failure is reported", func(t *testing.T) {
		err := Run(RunArgs{Dir: filepath.Join(t.TempDir(), "nowhere")})
		require.ErrorContains(t, err, "go.mod")
	})

	t.Run("An empty directory means the working directory", func(t *testing.T) {
		// This package's own directory encloses a module, so discovery resolves and writes nothing.
		require.NoError(t, Run(RunArgs{Tool: "gen"}, Generator(func(pkgs []Package) (map[string]string, error) {
			require.Empty(t, pkgs)
			return nil, nil
		})))
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

	t.Run("Unparseable go.mod", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"go.mod": "go 1.26\n"})
		_, _, err := enclosingModule(dir)
		require.ErrorContains(t, err, "parsing")
	})
}

func TestIsScalarStruct(t *testing.T) {
	require.True(t, isScalarStruct(reflect.TypeFor[scalarByValue]()))
	require.True(t, isScalarStruct(reflect.TypeFor[scalarByPointer]()))
	require.False(t, isScalarStruct(reflect.TypeFor[lookupTarget]()))
}

type scalarByValue struct{}

func (scalarByValue) MarshalText() ([]byte, error) { return nil, nil }

type scalarByPointer struct{}

func (*scalarByPointer) UnmarshalText([]byte) error { return nil }

// A type declared only in a test file is not part of the package a consumer imports, so parsing
// finds the package but never the type. Saying so beats reporting it as having no documentation.
func TestDiscoverLocalTypeMissingFromSource(t *testing.T) {
	_, err := Discover(".", &lookupTarget{})
	require.ErrorContains(t, err, "declared in no source file")
}

type repeatedLeaf struct {
	Value string
}

type repeatedEmbed struct {
	Flat string
}

// repeatedRoot reaches repeatedLeaf six ways, and itself through Self, so a walk that did not
// remember what it had seen would emit repeatedLeaf six times and never finish the cycle.
type repeatedRoot struct {
	repeatedEmbed
	Direct   repeatedLeaf
	Pointer  *repeatedLeaf
	Slice    []repeatedLeaf
	Array    [2]repeatedLeaf
	Keyed    map[string]repeatedLeaf
	Pointers []*repeatedLeaf
	Self     *repeatedRoot
}

type repeatedOther struct {
	Shared repeatedLeaf
}

func TestCollectStructsDeduplicates(t *testing.T) {
	names := func(types []reflect.Type) []string {
		var out []string
		for _, typ := range types {
			out = append(out, typ.Name())
		}
		return out
	}

	t.Run("One root reaching a type many ways", func(t *testing.T) {
		require.ElementsMatch(t,
			[]string{"repeatedRoot", "repeatedEmbed", "repeatedLeaf"},
			names(collectStructs([]any{&repeatedRoot{}})))
	})

	// Two roots sharing a type must share the walk's memory, or the second would emit it again.
	t.Run("Several roots sharing a type", func(t *testing.T) {
		require.ElementsMatch(t,
			[]string{"repeatedRoot", "repeatedEmbed", "repeatedLeaf", "repeatedOther"},
			names(collectStructs([]any{&repeatedRoot{}, &repeatedOther{}, &repeatedLeaf{}})))
	})

	t.Run("The same root given twice", func(t *testing.T) {
		require.Equal(t,
			names(collectStructs([]any{&repeatedRoot{}})),
			names(collectStructs([]any{&repeatedRoot{}, &repeatedRoot{}})))
	})
}

// A duplicate would reach the generated file as two methods on one receiver, which does not
// compile - so the guarantee is checked where it is consumed, not only in the walk.
func TestDiscoverEmitsEachTypeOnce(t *testing.T) {
	// Package, Type and FieldDoc are declared in this package's own source, and Type is reachable
	// from both roots as well as through Package's own slice of them.
	pkgs, err := Discover(".", &Package{}, &Type{})
	require.NoError(t, err)
	require.Len(t, pkgs, 1)

	seen := map[string]int{}
	for _, typ := range pkgs[0].Types {
		seen[typ.Name]++
	}
	for name, count := range seen {
		require.Equal(t, 1, count, "%s discovered %d times", name, count)
	}
	require.Contains(t, seen, "Type")
	require.Contains(t, seen, "FieldDoc")

	generated := docCommentsFile(pkgs[0])
	require.Equal(t, 1, strings.Count(generated, "func (Type) DocComments()"))
	_, parseErr := parser.ParseFile(token.NewFileSet(), GeneratedFileName, generated, parser.SkipObjectResolution)
	require.NoError(t, parseErr)
}
