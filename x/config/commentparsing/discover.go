package commentparsing

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

var (
	textMarshaler   = reflect.TypeFor[encoding.TextMarshaler]()
	textUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()
)

// Discover collects the documentation of every struct type reachable from the given roots, each of
// which may be a struct, a pointer to one, or a config value of any shape holding them.
//
// Roots are variadic because a package often exposes more than one entry point - a config and a
// secrets file, or two sections a consumer may embed independently - and a type reachable from
// none of them would be left undocumented for whoever embeds it next.
//
// dir anchors the search: the module enclosing it decides which types are local. A type this
// module declares has its source on hand, so its comments are parsed and stay incapable of
// disagreeing with the declaration. A type from a dependency has no reachable source, so its
// generated DocComments method is read, and a dependency that never generated is named rather
// than documented as blank.
//
// The walk follows pointers, slices, arrays, maps and embedded fields, so a caller names one root
// and never enumerates what it contains. Types decoded from a single string - a timestamp, a
// duration - are left out: they are leaf values, not config sections, and a dependency has every
// right to document no fields on one.
//
// Only the types the walk reached are returned, so a package's internal structs are not carried
// into a generator that would have no idea which of them a config file can name.
func Discover(dir string, roots ...any) ([]Package, error) {
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
	for _, structType := range collectStructs(roots) {
		importPath := structType.PkgPath()
		if importPath == "" {
			continue // an anonymous struct has no package to document it
		}

		pkg, ok := byPath[importPath]
		if !ok {
			pkg = &Package{ImportPath: importPath}
			byPath[importPath] = pkg
		}

		rel, local := strings.CutPrefix(importPath, modulePath)
		if !local {
			fields, err := Lookup(structType)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			pkg.Types = append(pkg.Types, Type{Name: structType.Name(), Fields: fields})
			continue
		}

		source, ok := parsed[importPath]
		if !ok {
			source, err = ParseDir(filepath.Join(moduleRoot, filepath.FromSlash(strings.TrimPrefix(rel, "/"))))
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
		pkg.Types = append(pkg.Types, typ)
	}

	return sortedPackages(byPath), errors.Join(errs...)
}

// sortedPackages orders packages and their types by name, so a regenerated file is byte-identical
// to the last one and a CI diff means something.
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
// Exactly once is what the generated code depends on: a type collected twice would become two
// DocComments methods on one receiver, which does not compile. Reaching one type through a field,
// a slice and an embed at the same time is ordinary, and a config that refers back to itself would
// not terminate without the same memory.
func collectStructs(roots []any) []reflect.Type {
	var found []reflect.Type
	visited := make(map[reflect.Type]bool)

	var walk func(reflect.Type)
	walk = func(t reflect.Type) {
		t = derefType(t)
		// reflect.TypeOf(nil) is nil, so an untyped nil root arrives here as one.
		if t == nil || visited[t] {
			return
		}
		visited[t] = true

		switch t.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			walk(t.Elem())
		case reflect.Struct:
			if isScalarStruct(t) {
				return
			}
			found = append(found, t)
			for i := range t.NumField() {
				field := t.Field(i)
				// An unexported embed still contributes its exported fields, so it is followed;
				// an unexported named field cannot be set from a config file.
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

// isScalarStruct reports whether a struct is decoded from a single value, which is how a type opts
// out of being a config section with fields of its own.
func isScalarStruct(t reflect.Type) bool {
	pointer := reflect.PointerTo(t)
	return t.Implements(textMarshaler) || pointer.Implements(textMarshaler) ||
		t.Implements(textUnmarshaler) || pointer.Implements(textUnmarshaler)
}

func derefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// enclosingModule walks up from dir to the nearest go.mod.
//
// This is what tells a locally declared type from a dependency's, and it is read from the
// filesystem rather than asked of the go tool so that discovery stays a library operation - see
// the package README.
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
	for line := range strings.SplitSeq(string(gomod), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(rest), nil
		}
	}
	return "", errors.New("no module directive in go.mod")
}
