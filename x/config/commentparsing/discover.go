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

// Docs is the documentation of a whole config tree: every struct type reachable from the root,
// wherever it was declared.
//
// It is what a consumer holds instead of a list of packages. Nothing about a config tree is the
// caller's to restate - which types take part, which packages they live in, which of those are
// local - so [Discover] works all of it out from the root value and this carries the answer.
type Docs struct {
	// fields is keyed by package path and type name together, because two packages of one config
	// tree may well declare a Config apiece.
	fields map[string]map[string]FieldDoc

	// dirs are the discovered packages whose source this module owns, keyed by package path.
	// Only these can be generated into: another module's source is in the read-only module cache,
	// which is the whole reason it has to ship its documentation compiled.
	dirs map[string]string
}

// Discover collects the documentation of every struct type reachable from root, which may be a
// struct, a pointer to one, or a config value of any shape holding them.
//
// Each type is resolved the only way it can be. A type this module declares has its source on
// hand, so its comments are parsed directly and stay incapable of disagreeing with the
// declaration. A type from a dependency has no reachable source, so its generated DocComments
// method is used, and a dependency that never generated is named rather than silently documented
// as blank.
//
// The walk follows pointers, slices, arrays, maps and embedded fields, so a caller names one root
// and never enumerates what it contains. Types decoded from a single string - a timestamp, a
// duration - are left out: they are leaf values, not config sections, and have no fields to
// describe.
func Discover(root any) (*Docs, error) {
	moduleRoot, modulePath, err := enclosingModule()
	if err != nil {
		return nil, err
	}

	docs := &Docs{
		fields: make(map[string]map[string]FieldDoc),
		dirs:   make(map[string]string),
	}
	parsed := make(map[string]bool)

	var errs []error
	for _, structType := range collectStructs(reflect.TypeOf(root)) {
		pkgPath := structType.PkgPath()
		if pkgPath == "" {
			continue // an anonymous struct has no package to document it
		}

		rel, inModule := strings.CutPrefix(pkgPath, modulePath)
		if !inModule {
			if err := docs.addGenerated(structType); err != nil {
				errs = append(errs, err)
			}
			continue
		}

		dir := filepath.Join(moduleRoot, filepath.FromSlash(strings.TrimPrefix(rel, "/")))
		docs.dirs[pkgPath] = dir
		if parsed[pkgPath] {
			continue
		}
		parsed[pkgPath] = true
		if err := docs.addParsed(pkgPath, dir); err != nil {
			errs = append(errs, err)
		}
	}
	return docs, errors.Join(errs...)
}

// addParsed records a package this module declares, read from its source.
func (d *Docs) addParsed(pkgPath, dir string) error {
	pkg, err := ParseDir(dir)
	if err != nil {
		return fmt.Errorf("%s: %w", pkgPath, err)
	}
	for _, typ := range pkg.Types {
		d.fields[pkgPath+"."+typ.Name] = typ.Fields
	}
	return nil
}

// addGenerated records a dependency's type from the method compiled into it.
func (d *Docs) addGenerated(structType reflect.Type) error {
	fields, err := Lookup(structType)
	if err != nil {
		return err
	}
	d.fields[structType.PkgPath()+"."+structType.Name()] = fields
	return nil
}

// Type returns the documentation of structType's fields, keyed by Go field name.
//
// A type the walk never reached is an error rather than an empty result, because the alternative
// is a reference that silently describes nothing and reads as though the fields were never
// documented.
func (d *Docs) Type(structType reflect.Type) (map[string]FieldDoc, error) {
	key, ok := typeKey(structType)
	if !ok {
		return nil, errors.New("no documentation discovered for a nil type")
	}
	fields, ok := d.fields[key]
	if !ok {
		return nil, fmt.Errorf("no documentation discovered for %s: it is not reachable from the root value", key)
	}
	return fields, nil
}

// Field returns one field's documentation, zero if it has none. It is the lookup a renderer wants
// while walking a struct, where a field with no comment is a fact to report against that field
// rather than a reason to abandon the walk.
func (d *Docs) Field(structType reflect.Type, fieldName string) FieldDoc {
	key, _ := typeKey(structType)
	return d.fields[key][fieldName]
}

// typeKey names a struct type the way the index is keyed, reporting false for the nil type a
// caller can reach through an interface field.
func typeKey(structType reflect.Type) (string, bool) {
	structType = derefType(structType)
	if structType == nil {
		return "", false
	}
	return structType.PkgPath() + "." + structType.Name(), true
}

// Packages returns the package paths the walk reached, sorted.
func (d *Docs) Packages() []string {
	seen := make(map[string]bool)
	for key := range d.fields {
		seen[key[:strings.LastIndex(key, ".")]] = true
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// GenerateDocCommentFiles writes a DocComments file into every discovered directory this module
// owns, so the types documented here stay readable from a module that cannot see this source.
//
// A dependency's package is not written to: its source is in the read-only module cache, and its
// documentation is its own repository's to generate.
func (d *Docs) GenerateDocCommentFiles() error {
	var errs []error
	for _, pkgPath := range sortedDirKeys(d.dirs) {
		if err := WriteFile(d.dirs[pkgPath]); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", pkgPath, err))
		}
	}
	return errors.Join(errs...)
}

// collectStructs returns every struct type reachable through t, the root first and each type once.
func collectStructs(t reflect.Type) []reflect.Type {
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
			return
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

	walk(t)
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

func sortedDirKeys(dirs map[string]string) []string {
	keys := make([]string, 0, len(dirs))
	for key := range dirs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// enclosingModule locates the module the caller is running in, by walking up from the working
// directory to the nearest go.mod.
//
// This is what tells a locally declared type from a dependency's, and it is read from the
// filesystem rather than asked of the go tool so that discovery stays a library operation - see
// the package README.
func enclosingModule() (root, path string, err error) {
	dir, err := os.Getwd()
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
