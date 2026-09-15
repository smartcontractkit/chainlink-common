package commentparsing

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
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

	t.Run("Generic declaration", func(t *testing.T) {
		dir := writePackage(t, map[string]string{"types.go": `package example

type Plain struct {
	// Value is documented.
	Value string
}

type Generic[T any] struct {
	// Value is documented.
	Value T
}
`})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		_, ok := pkg.Type("Plain")
		require.True(t, ok)
		_, ok = pkg.Type("Generic")
		require.False(t, ok, "no receiver can be written for it")
	})

	// A file the build excludes declares types this build does not compile. Collecting them
	// yields a receiver for a type that is not there, or a second receiver for one that is.
	t.Run("Only the files the build selects", func(t *testing.T) {
		dir := writePackage(t, map[string]string{
			"always.go": "package example\n\ntype Always struct {\n\t// Value is documented.\n\tValue string\n}\n",
			"tagged.go": "//go:build ignore\n\npackage example\n\ntype Tagged struct {\n\t// Value is documented.\n\tValue string\n}\n",
			// A GOOS suffix excludes a file without any constraint in it. js is never the
			// platform a test runs on.
			"config_js.go": "package example\n\ntype OtherPlatform struct {\n\t// Value is documented.\n\tValue string\n}\n",
		})

		pkg, err := ParseDir(dir)
		require.NoError(t, err)
		require.Equal(t, []string{"Always"}, typeNames(pkg))
	})
}

// An embedded type that is not a name is a branch ParseDir cannot reach, since go/parser rejects
// every input that would produce it, so it runs on a hand-built node.
func TestFieldNames(t *testing.T) {
	require.Nil(t, fieldNames(&ast.Field{Type: &ast.ArrayType{Elt: ast.NewIdent("int")}}))
}

// As with [TestFieldNames], a type declaration holding anything but a type spec is unreachable
// through ParseDir.
func TestTypesIn(t *testing.T) {
	file := &ast.File{Name: ast.NewIdent("example"), Decls: []ast.Decl{&ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`}}},
	}}}
	require.Empty(t, typesIn(file))
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
		require.Contains(t, got, "func (Config) DocComments() map[string]commentparsing.FieldDoc {")
		require.Contains(t, got, "return map[string]commentparsing.FieldDoc{")
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
		first := docCommentsFile(pkg)
		require.Equal(t, first, docCommentsFile(pkg))
	})

	t.Run("This package documents itself unqualified", func(t *testing.T) {
		self := pkg
		self.ImportPath, self.Name = selfImportPath, "commentparsing"
		got := docCommentsFile(self)
		require.NotContains(t, got, "import")
		require.Contains(t, got, "map[string]FieldDoc")
	})
}

func TestDocComments(t *testing.T) {
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

	// A dependency has no directory this module may write to, and must already have generated
	// for Discover to have resolved it.
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

func (lookupTarget) DocComments() map[string]FieldDoc {
	return map[string]FieldDoc{"Endpoint": {Comment: "Endpoint is the upstream URL."}}
}

type lookupUndocumented struct {
	Endpoint string
}

// LookupEmbeddable is exported, so reflection can allocate an embedded pointer to it.
type LookupEmbeddable struct {
	Endpoint string
}

func (LookupEmbeddable) DocComments() map[string]FieldDoc {
	return map[string]FieldDoc{"Endpoint": {Comment: "Endpoint is the upstream URL."}}
}

// lookupEmbedsPointer promotes its documentation through a pointer, which is nil in the value a
// lookup constructs.
type lookupEmbedsPointer struct {
	*LookupEmbeddable
	Extra string
}

// lookupEmbedsUnexportedPointer promotes it through a pointer reflection may not set.
type lookupEmbedsUnexportedPointer struct {
	*lookupTarget
	Extra string
}

// lookupEmbedsInterface promotes its documentation through an interface, which has no
// implementation to construct.
type lookupEmbedsInterface struct {
	DocCommenter
	Extra string
}

// LookupOwnMethod embeds a pointer reflection may not set, and documents itself anyway.
type LookupOwnMethod struct {
	*lookupTarget
	Extra string
}

func (LookupOwnMethod) DocComments() map[string]FieldDoc {
	return map[string]FieldDoc{"Extra": {Comment: "Extra is this type's own field."}}
}

// LookupCycle embeds a pointer to itself, which a walk that allocates them has to stop for.
type LookupCycle struct {
	*LookupCycle
	Endpoint string
}

func (LookupCycle) DocComments() map[string]FieldDoc {
	return map[string]FieldDoc{"Endpoint": {Comment: "Endpoint is the upstream URL."}}
}

// lookupPromoted never generated for itself and reaches lookupTarget's method by promotion.
type lookupPromoted struct {
	lookupTarget
	Extra string
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

	// A type that never generated for itself answers with what its embed promoted: right for
	// the promoted fields, silent about its own, which is how any undocumented field reads.
	t.Run("A promoted method documents what the embed brought with it", func(t *testing.T) {
		fields, err := Lookup(reflect.TypeFor[lookupPromoted]())
		require.NoError(t, err)
		require.Equal(t, "Endpoint is the upstream URL.", fields["Endpoint"].Comment)
		require.NotContains(t, fields, "Extra", "the embedder's own field is undocumented")
	})

	t.Run("Type with no package path", func(t *testing.T) {
		_, err := Lookup(reflect.TypeFor[struct{ Endpoint string }]())
		require.ErrorContains(t, err, "has no DocComments method")
	})

	t.Run("Documentation promoted through an embedded pointer", func(t *testing.T) {
		fields, err := Lookup(reflect.TypeFor[lookupEmbedsPointer]())
		require.NoError(t, err)
		require.Equal(t, "Endpoint is the upstream URL.", fields["Endpoint"].Comment)
	})

	// Neither an unexported embedded pointer nor an embedded interface can be filled in from
	// the type alone, so the promotion is reported rather than dispatched on nothing.
	t.Run("Documentation promoted through a receiver that cannot be built", func(t *testing.T) {
		for _, typ := range []reflect.Type{
			reflect.TypeFor[lookupEmbedsUnexportedPointer](),
			reflect.TypeFor[lookupEmbedsInterface](),
		} {
			_, err := Lookup(typ)
			require.ErrorContains(t, err, "promotes DocComments through an embedded field that cannot be constructed")
			require.ErrorContains(t, err, typ.Name())
		}
	})

	// The outer type's own method shadows what its embed promotes, so a receiver the promotion
	// would have needed is beside the point.
	t.Run("An unexported embedded pointer under a type documented in its own right", func(t *testing.T) {
		fields, err := Lookup(reflect.TypeFor[LookupOwnMethod]())
		require.NoError(t, err)
		require.Equal(t, "Extra is this type's own field.", fields["Extra"].Comment)
	})

	t.Run("A type embedding a pointer to itself", func(t *testing.T) {
		fields, err := Lookup(reflect.TypeFor[LookupCycle]())
		require.NoError(t, err)
		require.Contains(t, fields, "Endpoint")
	})

	// A generic type has no method either, but the advice every other miss carries -
	// regenerate - would send a caller after a file no generator can write.
	t.Run("Generic type says why rather than pointing at go generate", func(t *testing.T) {
		_, err := Lookup(reflect.TypeFor[genericConfig[string]]())
		require.ErrorContains(t, err, "genericConfig[string]")
		require.ErrorContains(t, err, "generic types are not supported")
		require.NotContains(t, err.Error(), "go generate")
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

	// A prefix-only match would send a dependency's types down the local path, to be parsed
	// from whichever directory the string manipulation lands on rather than read through Lookup.
	t.Run("A module sharing a prefix but not a path component", func(t *testing.T) {
		outside := writePackage(t, map[string]string{
			"go.mod": "module " + selfImportPath + "-extra\n\ngo 1.26\n",
		})

		pkgs, err := Discover(outside, &lookupTarget{})
		require.NoError(t, err)
		require.Len(t, pkgs, 1)
		require.Equal(t, selfImportPath, pkgs[0].ImportPath)
		require.Empty(t, pkgs[0].Dir, "resolved through Lookup, so there is nothing local to write")
	})

	// A nested module's package shares the parent's path prefix but belongs to its own run,
	// which is the one that applies the header and staleness handling to its files.
	t.Run("A nested module's package is a dependency", func(t *testing.T) {
		parent := writePackage(t, map[string]string{
			"go.mod": "module " + path.Dir(selfImportPath) + "\n\ngo 1.26\n",
		})
		nested := filepath.Join(parent, path.Base(selfImportPath))
		require.NoError(t, os.MkdirAll(nested, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(nested, "go.mod"),
			[]byte("module "+selfImportPath+"\n\ngo 1.26\n"), 0o600))

		pkgs, err := Discover(parent, &lookupTarget{})
		require.NoError(t, err)
		require.Len(t, pkgs, 1)
		require.Empty(t, pkgs[0].Dir, "resolved through Lookup, so there is nothing local to write")
	})

	// A receiver for a generic declaration needs its own type parameter list, which the name
	// reflection reports cannot supply, so the type is named rather than turned into a file
	// that does not compile.
	t.Run("Generic root", func(t *testing.T) {
		_, err := Discover(dir, &genericConfig[string]{})
		require.ErrorContains(t, err, "genericConfig[string]")
		require.ErrorContains(t, err, "generic types are not supported")
	})

	t.Run("No enclosing module", func(t *testing.T) {
		_, err := Discover(filepath.Join(t.TempDir(), "nowhere"))
		require.Error(t, err)
	})
}

// TestFiles covers the orchestration without touching disk: what a generator returns, how several
// combine, and how a failure is reported are all properties of the map Files hands back.
func TestFiles(t *testing.T) {
	module := map[string]string{"go.mod": "module example.com/x\n\ngo 1.26\n"}
	returning := func(files map[string]string) Generator {
		return func([]Package) (map[string]string, error) { return files, nil }
	}

	t.Run("Returns what a write would put on disk", func(t *testing.T) {
		dir := writePackage(t, module)
		files, err := Files(RunArgs{Dir: dir, Tool: "example.com/x/gen"},
			returning(map[string]string{"out/thing.go": "package out\nfunc  Thing ()  {}\n"}))
		require.NoError(t, err)
		require.Equal(t, map[string]string{
			"out/thing.go": "// Code generated by example.com/x/gen, DO NOT EDIT.\n\npackage out\n\nfunc Thing() {}\n",
		}, files)
	})

	t.Run("A file that is not Go source is untouched", func(t *testing.T) {
		dir := writePackage(t, module)
		files, err := Files(RunArgs{Dir: dir, Tool: "example.com/x/gen"},
			returning(map[string]string{"notes.md": "left  alone\n"}))
		require.NoError(t, err)
		require.Equal(t, map[string]string{"notes.md": "left  alone\n"}, files)
	})

	t.Run("A run with no tool name is refused", func(t *testing.T) {
		dir := writePackage(t, module)
		_, err := Files(RunArgs{Dir: dir}, returning(map[string]string{"out/thing.go": "package out\n"}))
		require.ErrorContains(t, err, "tool name")

		require.ErrorContains(t, Run(RunArgs{Dir: dir}), "tool name")
		entries, readErr := os.ReadDir(dir)
		require.NoError(t, readErr)
		require.Len(t, entries, 1, "only the go.mod it started with")
	})

	t.Run("Every generator's files are collected", func(t *testing.T) {
		dir := writePackage(t, module)
		files, err := Files(RunArgs{Dir: dir, Tool: "gen"},
			returning(map[string]string{"first.md": "content\n"}),
			returning(map[string]string{"second.md": "content\n"}))
		require.NoError(t, err)
		require.Len(t, files, 2)
		require.Contains(t, files, "first.md")
		require.Contains(t, files, "second.md")
	})

	// Whichever generator ran last would silently win, and only reading the output would show it.
	t.Run("Two generators claiming one path", func(t *testing.T) {
		dir := writePackage(t, module)
		clash := map[string]string{"clash.md": "content\n"}
		_, err := Files(RunArgs{Dir: dir, Tool: "gen"}, returning(clash), returning(clash))
		require.ErrorContains(t, err, "written by more than one generator")
	})

	t.Run("A generator's failure yields no files", func(t *testing.T) {
		dir := writePackage(t, module)
		files, err := Files(RunArgs{Dir: dir, Tool: "gen"}, func([]Package) (map[string]string, error) {
			return nil, os.ErrInvalid
		})
		require.ErrorIs(t, err, os.ErrInvalid)
		require.Nil(t, files)
	})

	t.Run("A path spelled two ways is one file", func(t *testing.T) {
		dir := writePackage(t, module)
		files, err := Files(RunArgs{Dir: dir, Tool: "gen"},
			returning(map[string]string{filepath.Join(".", "out", "..", "out", "thing.md"): "content\n"}))
		require.NoError(t, err)
		require.Equal(t, map[string]string{filepath.Join("out", "thing.md"): "content\n"}, files)
	})

	t.Run("Two generators claiming one path spelled differently", func(t *testing.T) {
		dir := writePackage(t, module)
		_, err := Files(RunArgs{Dir: dir, Tool: "gen"},
			returning(map[string]string{"clash.md": "content\n"}),
			returning(map[string]string{filepath.Join(".", "clash.md"): "content\n"}))
		require.ErrorContains(t, err, "written by more than one generator")
	})

	t.Run("A path outside the run directory", func(t *testing.T) {
		dir := writePackage(t, module)
		_, err := Files(RunArgs{Dir: dir, Tool: "gen"},
			returning(map[string]string{filepath.Join("..", "escaped.md"): "content\n"}))
		require.ErrorContains(t, err, "lies outside the run directory")

		require.ErrorContains(t, Run(RunArgs{Dir: dir, Tool: "gen"},
			returning(map[string]string{filepath.Join("..", "escaped.md"): "content\n"})),
			"lies outside the run directory")
		require.NoFileExists(t, filepath.Join(filepath.Dir(dir), "escaped.md"))
	})

	t.Run("An absolute path", func(t *testing.T) {
		dir := writePackage(t, module)
		_, err := Files(RunArgs{Dir: dir, Tool: "gen"},
			returning(map[string]string{filepath.Join(dir, "absolute.md"): "content\n"}))
		require.ErrorContains(t, err, "relative to the run directory")
	})

	// A root declared above the run directory puts the package's own generated file above it
	// too, where nothing scans for it afterwards.
	t.Run("A discovered package above the run directory", func(t *testing.T) {
		_, err := Files(RunArgs{
			Dir:   filepath.Join("examples", "simple"),
			Roots: []any{&Package{}},
			Tool:  "gen",
		})
		require.ErrorContains(t, err, "lies outside the run directory")
	})

	t.Run("Discovery failure is reported", func(t *testing.T) {
		_, err := Files(RunArgs{Dir: filepath.Join(t.TempDir(), "nowhere"), Tool: "gen"})
		require.ErrorContains(t, err, "go.mod")
	})

	t.Run("An empty directory means the working directory", func(t *testing.T) {
		// This package's directory encloses a module, so discovery resolves and, with no
		// roots, finds nothing to document.
		files, err := Files(RunArgs{Tool: "gen"}, func(pkgs []Package) (map[string]string, error) {
			require.Empty(t, pkgs)
			return nil, nil
		})
		require.NoError(t, err)
		require.Empty(t, files)
	})
}

// TestRun covers what Files cannot: that the result reaches disk, formatted and headed the same
// way whichever generator produced it, and that a failure writes nothing.
func TestRun(t *testing.T) {
	const header = "// Code generated by example.com/x/gen, DO NOT EDIT.\n\n"
	module := map[string]string{"go.mod": "module example.com/x\n\ngo 1.26\n"}

	write := func(t *testing.T, dir, rel, content string) string {
		t.Helper()
		path := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		return path
	}

	t.Run("Writes every generator's files, formatted and headed", func(t *testing.T) {
		dir := writePackage(t, module)
		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"},
			func([]Package) (map[string]string, error) {
				return map[string]string{
					"out/thing.go": "package out\nfunc  Thing ()  {}\n",
					"notes.md":     "left alone\n",
				}, nil
			}))

		written, err := os.ReadFile(filepath.Join(dir, "out", "thing.go"))
		require.NoError(t, err)
		require.Contains(t, string(written), "// Code generated by example.com/x/gen, DO NOT EDIT.")
		require.Contains(t, string(written), "func Thing() {}", "codegen gofmts every file it writes")

		// Only Go source is formatted and headed, so a generator emitting anything else gets it
		// through untouched.
		notes, err := os.ReadFile(filepath.Join(dir, "notes.md"))
		require.NoError(t, err)
		require.Equal(t, "left alone\n", string(notes))
	})

	t.Run("A generator's failure stops the write", func(t *testing.T) {
		dir := writePackage(t, module)
		err := Run(RunArgs{Dir: dir, Tool: "gen"}, func([]Package) (map[string]string, error) {
			return nil, os.ErrInvalid
		})
		require.ErrorIs(t, err, os.ErrInvalid)

		entries, readErr := os.ReadDir(dir)
		require.NoError(t, readErr)
		require.Len(t, entries, 1, "only the go.mod it started with")
	})

	// A package that stops documenting anything would otherwise keep a file whose methods
	// name types it no longer has.
	t.Run("A generated file this run did not produce is removed", func(t *testing.T) {
		dir := writePackage(t, module)
		stale := write(t, dir, GeneratedFileName, header+"package x\n")
		nested := write(t, dir, filepath.Join("sub", GeneratedFileName), header+"package sub\n")

		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"}))
		require.NoFileExists(t, stale)
		require.NoFileExists(t, nested)
	})

	t.Run("A regenerated file stays", func(t *testing.T) {
		dir := writePackage(t, module)
		path := write(t, dir, GeneratedFileName, header+"package x\n")

		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"},
			func([]Package) (map[string]string, error) {
				return map[string]string{GeneratedFileName: "package x\n\nfunc Now() {}\n"}, nil
			}))

		written, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Contains(t, string(written), "func Now() {}")
	})

	// A nested module generates through its own run, and this one has no directory there to
	// write to, so it has no business deleting what it finds.
	t.Run("A nested module's generated file is left alone", func(t *testing.T) {
		dir := writePackage(t, module)
		write(t, dir, filepath.Join("inner", "go.mod"), "module example.com/inner\n\ngo 1.26\n")
		inner := write(t, dir, filepath.Join("inner", GeneratedFileName), header+"package inner\n")

		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"}))
		require.FileExists(t, inner)
	})

	t.Run("Another tool's file is not written over", func(t *testing.T) {
		dir := writePackage(t, module)
		theirs := "// Code generated by example.com/x/other, DO NOT EDIT.\n\npackage x\n"
		path := write(t, dir, GeneratedFileName, theirs)

		err := Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"},
			func([]Package) (map[string]string, error) {
				return map[string]string{GeneratedFileName: "package x\n\nfunc Generated() {}\n"}, nil
			})
		require.ErrorContains(t, err, "already generated by example.com/x/other")

		kept, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		require.Equal(t, theirs, string(kept))
	})

	t.Run("A hand-written file under the reserved name is not overwritten", func(t *testing.T) {
		dir := writePackage(t, module)
		path := write(t, dir, GeneratedFileName, "package x\n\n// Written by hand.\n")

		err := Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"},
			func([]Package) (map[string]string, error) {
				return map[string]string{GeneratedFileName: "package x\n\nfunc Generated() {}\n"}, nil
			})
		require.ErrorContains(t, err, "reserved for generated files")
		require.ErrorContains(t, err, GeneratedFileName)

		kept, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		require.Equal(t, "package x\n\n// Written by hand.\n", string(kept))
	})

	t.Run("A hand-written file nothing generated over is not removed", func(t *testing.T) {
		dir := writePackage(t, module)
		path := write(t, dir, GeneratedFileName, "package x\n\n// Written by hand.\n")

		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"}))
		require.FileExists(t, path)
	})

	// codegen decides whether a path is already rooted at the run directory from a string
	// prefix, which this path passes without being under it.
	t.Run("A generated path beginning with the run directory's name", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "config")
		require.NoError(t, os.MkdirAll(dir, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"),
			[]byte("module example.com/x\n\ngo 1.26\n"), 0o600))

		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"},
			func([]Package) (map[string]string, error) {
				return map[string]string{filepath.Join("config_types", "out.go"): "package out\n"}, nil
			}))

		require.FileExists(t, filepath.Join(dir, "config_types", "out.go"))
	})

	// Two runs over one module each own what their own tool wrote, or the second would delete
	// the first's output for not being in its own file set.
	t.Run("A file another tool generated is left alone", func(t *testing.T) {
		dir := writePackage(t, module)
		other := write(t, dir, GeneratedFileName,
			"// Code generated by example.com/x/other, DO NOT EDIT.\n\npackage x\n")

		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"}))
		require.FileExists(t, other)
	})

	// gofmt hoists a build constraint above the header, so the whole comment block is read.
	t.Run("A header below a build constraint still marks the file generated", func(t *testing.T) {
		dir := writePackage(t, module)
		path := write(t, dir, GeneratedFileName, "//go:build linux\n\n"+header+"package x\n")

		require.NoError(t, Run(RunArgs{Dir: dir, Tool: "example.com/x/gen"}))
		require.NoFileExists(t, path)
	})
}

func TestModulePath(t *testing.T) {
	t.Run("Module directive", func(t *testing.T) {
		path, err := modulePath([]byte("module example.com/x\n\ngo 1.26\n"))
		require.NoError(t, err)
		require.Equal(t, "example.com/x", path)
	})

	// Forms the go tool accepts, which a line-by-line reader gets wrong: the path comes back
	// with the comment attached, or the directive is not recognised at all. Either one makes
	// every local type look like a dependency's.
	t.Run("Module directive with an inline comment", func(t *testing.T) {
		path, err := modulePath([]byte("module example.com/x // deliberately\n\ngo 1.26\n"))
		require.NoError(t, err)
		require.Equal(t, "example.com/x", path)
	})

	t.Run("Module directive separated by a tab", func(t *testing.T) {
		path, err := modulePath([]byte("module\texample.com/x\n\ngo 1.26\n"))
		require.NoError(t, err)
		require.Equal(t, "example.com/x", path)
	})

	t.Run("Quoted module path", func(t *testing.T) {
		path, err := modulePath([]byte("module \"example.com/x\"\n\ngo 1.26\n"))
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
// finds the package but never the type - an error, not blank documentation.
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

// repeatedRoot reaches repeatedLeaf six ways and itself through Self, so a walk without memory
// would emit repeatedLeaf six times and never finish the cycle.
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

func TestCollectStructs(t *testing.T) {
	names := func(types []reflect.Type) []string {
		out := make([]string, 0, len(types))
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
// compile, so the guarantee is checked where it is consumed as well as in the walk.
func TestDiscoverEmitsEachTypeOnce(t *testing.T) {
	// Type is reachable from both roots as well as through Package's slice of them.
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

// Go allows a type no field and method of one name, so a type carrying the field compiles only
// until its documentation is generated. The parsed source is where that is caught, since the
// reflected type would already have to have been generated for to be seen here.
// Go leaves a selector ambiguous between two embedded types legal to declare and illegal to use,
// so there is no name a config file could set.
func TestAmbiguousPromotion(t *testing.T) {
	require.NoError(t, ambiguousPromotion(reflect.TypeFor[repeatedRoot]()))

	err := ambiguousPromotion(reflect.TypeFor[ambiguousOuter]())
	require.ErrorContains(t, err, "Region from ambiguousLeft and ambiguousRight")

	// The outer type's own field wins at the shallower depth, so it is not ambiguous.
	require.NoError(t, ambiguousPromotion(reflect.TypeFor[ambiguousShadowed]()))
}

type ambiguousLeft struct {
	Region string
}

type ambiguousRight struct {
	Region string
	Tenant string
}

type ambiguousOuter struct {
	ambiguousLeft
	ambiguousRight
}

type ambiguousShadowed struct {
	ambiguousLeft
	Region string
}

func TestReservedFieldName(t *testing.T) {
	err := reservedFieldName("example.com/app", "Config", map[string]FieldDoc{
		"DocComments": {Comment: "DocComments is a field, oddly."},
	})
	require.ErrorContains(t, err, "example.com/app.Config")
	require.ErrorContains(t, err, "has a field named DocComments")

	require.NoError(t, reservedFieldName("example.com/app", "Config", map[string]FieldDoc{
		"Endpoint": {Comment: "Endpoint is the upstream URL."},
	}))
}

func TestDocCommentsReceivers(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "example.go", `package example

type Named struct{}

func (n Named) DocComments() map[string]FieldDoc { return nil }

type Unnamed struct{}

func (Unnamed) DocComments() map[string]FieldDoc { return nil }

type Pointer struct{}

func (p *Pointer) DocComments() map[string]FieldDoc { return nil }

type Other struct{}

func (Other) Something() {}

func Free() {}
`, parser.SkipObjectResolution)
	require.NoError(t, err)

	// The receiver's own name is not the type's, so a named receiver must resolve to the type.
	require.ElementsMatch(t, []string{"Named", "Unnamed", "Pointer"}, docCommentsReceivers(file))
}

// A generated method would redeclare the one the package wrote by hand.
func TestParseDirRecordsAHandWrittenDocCommentsMethod(t *testing.T) {
	dir := writePackage(t, map[string]string{"types.go": `package example

type Own struct {
	// Value is documented.
	Value string
}

func (o Own) DocComments() map[string]FieldDoc { return nil }

type Plain struct {
	// Value is documented.
	Value string
}
`})

	pkg, err := ParseDir(dir)
	require.NoError(t, err)
	require.True(t, pkg.declaresDocComments["Own"])
	require.False(t, pkg.declaresDocComments["Plain"])
}

func TestPackageWithinModule(t *testing.T) {
	const module = "example.com/app"

	t.Run("The module itself", func(t *testing.T) {
		rel, within := packageWithinModule(module, module)
		require.True(t, within)
		require.Empty(t, rel)
	})

	t.Run("A package inside the module", func(t *testing.T) {
		rel, within := packageWithinModule(module+"/config/sub", module)
		require.True(t, within)
		require.Equal(t, "config/sub", rel)
	})

	// A module whose path only shares a prefix would otherwise be resolved against this
	// module's source, reading a directory that documents something else entirely.
	t.Run("A module sharing a prefix but not a path component", func(t *testing.T) {
		_, within := packageWithinModule(module+"-extra/config", module)
		require.False(t, within)
	})

	t.Run("An unrelated module", func(t *testing.T) {
		_, within := packageWithinModule("example.org/other", module)
		require.False(t, within)
	})
}

type genericConfig[T any] struct {
	Value T
}

func typeNames(pkg *Package) []string {
	names := make([]string, 0, len(pkg.Types))
	for _, typ := range pkg.Types {
		names = append(names, typ.Name)
	}
	return names
}
