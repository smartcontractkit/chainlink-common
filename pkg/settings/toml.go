package settings

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/pelletier/go-toml"
)

// CombineTOMLFiles reads a set of TOML config files and combines them in to one file. The expected inputs are:
//	- global.toml
// 	- org/*.toml
// 	- owner/*.toml
// 	- workflow/*.toml
// The directory and file names translate to keys in the TOML structure, while the file extensions are discarded.
// For example: owner/0x1234.toml:Foo.Bar becomes owner.0x1234.Foo.Bar
func CombineTOMLFiles(files fs.FS) ([]byte, error) {
	tree, err := toml.TreeFromMap(map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("failed to ecode TOML: %w", err)
	}
	global, err := readTOMLTree(files, "global.toml")
	if err != nil {
		return nil, err
	}
	tree.Set("global", global)
	orgs, err := readTOMLTrees(files, "org")
	if err != nil {
		return nil, err
	}
	tree.Set("org", orgs)
	owners, err := readTOMLTrees(files, "owner")
	if err != nil {
		return nil, err
	}
	tree.Set("owner", owners)
	workflows, err := readTOMLTrees(files, "workflow")
	if err != nil {
		return nil, err
	}
	tree.Set("workflow", workflows)
	var b bytes.Buffer
	e := toml.NewEncoder(&b).Indentation("")
	err = e.Encode(tree)
	return b.Bytes(), err
}

func readTOMLTree(files fs.FS, name string) (*toml.Tree, error) {
	f, err := files.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", name, err)
	}
	defer f.Close()
	t, err := toml.LoadReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", name, err)
	}

	return t, nil
}

func readTOMLTrees(files fs.FS, dir string) (*toml.Tree, error) {
	trees, err := toml.TreeFromMap(map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("failed to create TOML tree: %w", err)
	}
	if err := fs.WalkDir(files, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // ignore
		}
		t, err := readTOMLTree(files, path)
		if err != nil {
			return fmt.Errorf("failed to read toml file %s: %w", path, err)
		}
		name := strings.TrimSuffix(d.Name(), ".toml")
		trees.Set(name, t)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk %s: %w", dir, err)
	}
	return trees, nil
}

type tomlSettings struct {
	tree *toml.Tree // opt: flat-map of full-qualified keys may be faster
}

func newTOMLSettings(b []byte) (*tomlSettings, error) {
	tree, err := toml.LoadBytes(b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse toml: %w", err)
	}
	return &tomlSettings{tree: tree}, nil
}

func (t *tomlSettings) getFirst(keys ...string) (string, error) {
	for _, k := range keys {
		v := t.tree.Get(k)
		if v == nil {
			continue // next key
		}
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("non-string value: %s: %t(%v)", k, v, v)
		}
		return s, nil
	}
	return "", nil // no values
}

var _ Getter = &tomlGetter{}

type tomlGetter struct {
	settings *tomlSettings
}

// NewTOMLGetter returns a static Getter backed by the given TOML.
//TODO https://smartcontract-it.atlassian.net/browse/CAPPL-775
// NewTOMLRegistry with polling & subscriptions
func NewTOMLGetter(b []byte) (Getter, error) {
	s, err := newTOMLSettings(b)
	if err != nil {
		return nil, err
	}
	return &tomlGetter{settings: s}, nil
}

func (t *tomlGetter) GetScoped(ctx context.Context, scope Scope, key string) (value string, err error) {
	keys, err := scope.rawKeys(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get raw keys: %w", err)
	}
	return t.settings.getFirst(keys...)
}
