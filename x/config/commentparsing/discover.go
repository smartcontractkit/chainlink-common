package commentparsing

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

// discover collects the documentation of every struct type reachable from the given roots.
//
// dir anchors the search: the module enclosing it decides which types are local. A type this
// module declares has its source on hand, so its comments are parsed. A type from a dependency has
// no reachable source, so its generated DocComments method is read instead, and a dependency that
// never generated is reported as an error rather than documented as blank.
//
// The walk follows pointers, slices, arrays, maps and embedded fields, and stops at any type the
// language reads whole: its fields are not configured, so not documented, one at a time.
func discover(dir string, lang markup.Markup, roots ...any) ([]Package, error) {
	moduleRoot, modulePath, err := enclosingModule(dir)
	if err != nil {
		return nil, err
	}
	runDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	byPath := make(map[string]*Package)
	parsed := make(map[string]*Package)

	var errs []error
	for _, structType := range collectStructs(roots, lang.IsLeaf) {
		importPath := structType.PkgPath()
		if importPath == "" {
			continue // an anonymous struct has no package to document it
		}

		if isInstantiatedGeneric(structType) {
			errs = append(errs, fmt.Errorf("%s.%s: %s", importPath, structType.Name(), genericUnsupported))
			continue
		}

		if ambiguousErr := ambiguousPromotion(structType); ambiguousErr != nil {
			errs = append(errs, ambiguousErr)
			continue
		}

		pkg, ok := byPath[importPath]
		if !ok {
			pkg = &Package{ImportPath: importPath}
			byPath[importPath] = pkg
		}

		dir, local := localPackageDir(moduleRoot, modulePath, importPath)
		if !local {
			fields, lookupErr := Lookup(structType)
			if lookupErr != nil {
				errs = append(errs, lookupErr)
				continue
			}
			pkg.Types = append(pkg.Types, Type{Name: structType.Name(), Fields: fields})
			continue
		}

		source, ok := parsed[importPath]
		if !ok {
			source, err = ParseDir(dir)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", importPath, err))
				continue
			}
			parsed[importPath] = source

			// A generator's paths are relative to the run directory, so a package's is too.
			relDir, err := filepath.Rel(runDir, source.Dir)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", importPath, err))
				continue
			}
			pkg.Name, pkg.Dir = source.Name, relDir
		}

		typ, ok := source.Type(structType.Name())
		if !ok {
			errs = append(errs, fmt.Errorf("%s.%s: declared in no source file of %s",
				importPath, structType.Name(), source.Dir))
			continue
		}
		if source.declaresDocComments[typ.Name] {
			errs = append(errs, fmt.Errorf("%s.%s: declares a DocComments method of its own, which a "+
				"generated one would redeclare; rename it or drop the type from the roots",
				importPath, typ.Name))
			continue
		}
		if reservedErr := reservedFieldName(importPath, typ.Name, typ.Fields); reservedErr != nil {
			errs = append(errs, reservedErr)
			continue
		}
		pkg.Types = append(pkg.Types, typ)
	}

	return sortedPackages(byPath), errors.Join(errs...)
}

func localPackageDir(moduleRoot, modulePath, importPath string) (dir string, local bool) {
	rel, within := packageWithinModule(importPath, modulePath)
	if !within {
		return "", false
	}
	dir = filepath.Join(moduleRoot, filepath.FromSlash(rel))

	owner, _, err := enclosingModule(dir)
	if err != nil || owner != moduleRoot {
		return "", false
	}
	return dir, true
}

func packageWithinModule(importPath, modulePath string) (rel string, within bool) {
	if importPath == modulePath {
		return "", true
	}
	return strings.CutPrefix(importPath, modulePath+"/")
}

// ambiguousPromotion rejects a struct with a promoted field name that Go cannot resolve, which
// happens when two embedded types provide it at the same depth.
//
// Go leaves such a selector legal to declare and illegal to use, and a config file keying on the
// name has no way to say which of the two it means, so there is nothing for documentation to
// describe. The resolution is [reflect.Type.FieldByName]'s rather than one of this package's own,
// so shadowing and depth follow the same rules the compiler applies: a name the outer type
// declares itself wins, and so does one promoted from a shallower embed.
//
// A name is left alone when every path to it crosses an embedded pointer. That shape is how a
// polymorphic section is declared - one embed per variant, all nil but the one a discriminator
// selects - so the tie is between fields that never exist at once, and an encoder skipping the nil
// embeds emits the selected variant's keys unambiguously. A tie with even one pointer-free path is
// still refused: those fields do coexist, and nothing decides between them.
func ambiguousPromotion(t reflect.Type) error {
	var ambiguous []string
	for name, always := range promotedNames(t, false, map[embedVisit]bool{}) {
		if !always {
			continue
		}
		if _, resolved := t.FieldByName(name); !resolved {
			ambiguous = append(ambiguous, name)
		}
	}
	if len(ambiguous) == 0 {
		return nil
	}

	sort.Strings(ambiguous)
	return fmt.Errorf("%s: %s promoted from more than one embedded type at the same depth, so "+
		"nothing can name it to configure: shadow it on %s, or embed only one of them",
		typeName(t), strings.Join(ambiguous, ", "), t.Name())
}

// embedVisit keys the walk by the type reached and whether a pointer embed was crossed to reach
// it, so a type embedded both ways is walked once for each. Keying by type alone would let the
// first arrival decide, and a name is exempt only when no arrival was pointer-free.
type embedVisit struct {
	t       reflect.Type
	crossed bool
}

// promotedNames maps the exported field names t's embedded types provide, at every depth, to
// whether the name has a path to it that crosses no embedded pointer - that is, whether the field
// is there whatever the pointers hold. An unexported name is left out: no config file can name it
// either way.
//
// crossed reports whether the caller already went through a pointer to reach t; a field found
// below one is optional however many value embeds sit under it.
func promotedNames(t reflect.Type, crossed bool, walked map[embedVisit]bool) map[string]bool {
	visit := embedVisit{t: t, crossed: crossed}
	if t.Kind() != reflect.Struct || walked[visit] {
		return nil
	}
	walked[visit] = true

	names := make(map[string]bool)
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.Anonymous {
			continue
		}
		embedded := DerefType(field.Type)
		if embedded == nil || embedded.Kind() != reflect.Struct {
			continue
		}
		through := crossed || field.Type.Kind() == reflect.Pointer

		for j := range embedded.NumField() {
			if promoted := embedded.Field(j); promoted.IsExported() {
				names[promoted.Name] = names[promoted.Name] || !through
			}
		}
		for name, always := range promotedNames(embedded, through, walked) {
			names[name] = names[name] || always
		}
	}
	return names
}

func reservedFieldName(importPath, typeName string, fields map[string]FieldDoc) error {
	if _, taken := fields[docCommentsMethod]; !taken {
		return nil
	}
	return fmt.Errorf("%s.%s: has a field named %s, which the generated method of that name cannot "+
		"coexist with; rename the field", importPath, typeName, docCommentsMethod)
}

func isInstantiatedGeneric(t reflect.Type) bool {
	return strings.Contains(t.Name(), "[")
}

// sortedPackages orders packages and their types by name, so a regenerated file is byte-identical
// to the last one.
func sortedPackages(byPath map[string]*Package) []Package {
	packages := make([]Package, 0, len(byPath))
	for _, pkg := range byPath {
		sort.Slice(pkg.Types, func(i, j int) bool { return pkg.Types[i].Name < pkg.Types[j].Name })
		packages = append(packages, *pkg)
	}
	sort.Slice(packages, func(i, j int) bool { return packages[i].ImportPath < packages[j].ImportPath })
	return packages
}

// collectStructs returns every struct type reachable through the roots, each exactly once.
//
// A type collected twice would become two DocComments methods on one receiver, and a config that
// refers back to itself would not terminate.
func collectStructs(roots []any, isLeaf func(reflect.Type) bool) []reflect.Type {
	var found []reflect.Type
	visited := make(map[reflect.Type]bool)

	var walk func(reflect.Type)
	walk = func(t reflect.Type) {
		t = DerefType(t)
		// reflect.TypeOf(nil) is nil, so an untyped nil root arrives here as one.
		if t == nil || visited[t] {
			return
		}
		visited[t] = true

		switch t.Kind() {
		default:
			// A leaf value, with no fields of its own to document.
		case reflect.Slice, reflect.Array, reflect.Map:
			walk(t.Elem())
		case reflect.Struct:
			if isLeaf(t) {
				return
			}
			found = append(found, t)
			for i := range t.NumField() {
				field := t.Field(i)
				// An unexported embed still contributes its exported fields; an unexported
				// named field cannot be set from a config file.
				if field.IsExported() || field.Anonymous {
					walk(field.Type)
				}
			}
		}
	}

	for _, root := range roots {
		walk(reflect.TypeOf(root))
	}
	return found
}

// DerefType strips every level of pointer from t. A nil t stays nil.
func DerefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// enclosingModule walks up from dir to the nearest go.mod.
func enclosingModule(dir string) (root, path string, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod")) //nolint:gosec // G304: ancestor-walked module root, not caller input
		if err == nil {
			path, err := modulePath(data)
			if err != nil {
				return "", "", fmt.Errorf("parsing %s/go.mod: %w", dir, err)
			}
			return dir, path, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", fmt.Errorf("no go.mod at or above %s", dir)
		}
		dir = parent
	}
}

func modulePath(gomod []byte) (string, error) {
	if path := modfile.ModulePath(gomod); path != "" {
		return path, nil
	}
	return "", errors.New("no module directive in go.mod")
}
