package settings

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

func CombineYAMLFiles(files fs.FS) ([]byte, error) {
	m := make(map[string]any)
	global, err := readYAMLMap(files, "global.yaml")
	if err != nil {
		return nil, err
	}
	if len(global) > 0 {
		m["global"] = global
	}
	orgs, err := readYAMLMaps(files, "org")
	if err != nil {
		return nil, err
	}
	if len(orgs) > 0 {
		m["org"] = orgs
	}
	owners, err := readYAMLMaps(files, "owner")
	if err != nil {
		return nil, err
	}
	if len(owners) > 0 {
		m["owner"] = owners
	}
	workflows, err := readYAMLMaps(files, "workflow")
	if err != nil {
		return nil, err
	}
	if len(workflows) > 0 {
		m["workflow"] = workflows
	}

	var b bytes.Buffer
	b.WriteString(generatedHeader)
	e := yaml.NewEncoder(&b)
	err = e.Encode(m)
	return b.Bytes(), err
}

func readYAMLMap(files fs.FS, name string) (map[string]any, error) {
	f, err := files.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", name, err)
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	var m jsonMap
	err = d.Decode(&m)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", name, err)
	}

	return m, nil
}

func readYAMLMaps(files fs.FS, dir string) (jsonMap, error) {
	ms := make(map[string]any)
	if err := fs.WalkDir(files, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // ignore
		}
		m, err := readYAMLMap(files, path)
		if err != nil {
			return fmt.Errorf("failed to read yaml file %s: %w", path, err)
		}
		name := strings.TrimSuffix(d.Name(), ".yaml")
		ms[name] = m
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk %s: %w", dir, err)
	}
	return ms, nil
}

type yamlSettings struct {
	m map[string]any
}

func newYAMLSettings(b []byte) (*yamlSettings, error) {
	var m map[string]any
	err := yaml.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return &yamlSettings{m: m}, nil
}

func (y *yamlSettings) getFirst(keys ...string) (string, error) {
	for _, k := range keys {
		v, err := y.get(k)
		if err != nil {
			return "", err
		}
		if v != "" {
			return v, nil
		}
	}
	return "", nil // no values
}

func (y *yamlSettings) get(key string) (string, error) {
	m := y.m
	parts := strings.Split(key, ".")
	for i, part := range parts[:len(parts)-1] {
		v := m[part]
		if v == nil {
			return "", nil
		}
		var ok bool
		m, ok = v.(map[string]any)
		if !ok {
			return "", fmt.Errorf("invalid key %s: %s is a field", key, strings.Join(parts[:i+1], "."))
		}
	}

	field := parts[len(parts)-1]
	if val, ok := m[field]; ok {
		switch t := val.(type) {
		case string:
			return t, nil
		case bool:
			return strconv.FormatBool(t), nil
		default:
			return "", fmt.Errorf("non-string value: %s: %t(%v)", key, val, val)
		}
	}
	return "", nil // no value
}

type yamlGetter struct {
	settings *yamlSettings
}

func NewYAMLGetter(b []byte) (Getter, error) {
	s, err := newYAMLSettings(b)
	if err != nil {
		return nil, err
	}
	return &yamlGetter{settings: s}, nil
}

func (y *yamlGetter) GetScoped(ctx context.Context, scope Scope, key string) (value string, err error) {
	keys, err := scope.rawKeys(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get raw keys: %w", err)
	}
	return y.settings.getFirst(keys...)
}
