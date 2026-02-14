package settings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// CombineJSONFiles reads a set of JSON config files and combines them in to one file. The expected inputs are:
//   - global.json
//   - org/*.json
//   - owner/*.json
//   - workflow/*.json
//
// The directory and file names translate to keys in the JSON structure, while the file extensions are discarded.
// For example: owner/0x1234.json:Foo.Bar becomes owner.0x1234.Foo.Bar
func CombineJSONFiles(files fs.FS) ([]byte, error) {
	c := struct {
		Global   jsonMap `json:"global"`
		Org      jsonMap `json:"org"`
		Owner    jsonMap `json:"owner"`
		Workflow jsonMap `json:"workflow"`
	}{}
	global, err := readJSONMap(files, "global.json")
	if err != nil {
		return nil, err
	}
	c.Global = global
	orgs, err := readJSONMaps(files, "org")
	if err != nil {
		return nil, err
	}
	c.Org = orgs
	owners, err := readJSONMaps(files, "owner")
	if err != nil {
		return nil, err
	}
	c.Owner = owners
	workflows, err := readJSONMaps(files, "workflow")
	if err != nil {
		return nil, err
	}
	c.Workflow = workflows
	return json.MarshalIndent(c, "", "  ")
}

type jsonMap map[string]any

func (m jsonMap) MarshalJSON() ([]byte, error) {
	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, key := range keys {
		fmt.Fprintf(&buf, `"%s": `, key)
		b, err := json.Marshal(m[key])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value for %s: %w", key, err)
		}
		buf.Write(b)
		if i < len(keys)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}

func readJSONMap(files fs.FS, name string) (jsonMap, error) {
	f, err := files.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", name, err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	d.UseNumber()
	var m jsonMap
	err = d.Decode(&m)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", name, err)
	}

	return m, nil
}

func readJSONMaps(files fs.FS, dir string) (jsonMap, error) {
	ms := make(jsonMap)
	if err := fs.WalkDir(files, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // ignore
		}
		m, err := readJSONMap(files, path)
		if err != nil {
			return fmt.Errorf("failed to read json file %s: %w", path, err)
		}
		name := strings.TrimSuffix(d.Name(), ".json")
		ms[name] = m
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk %s: %w", dir, err)
	}
	return ms, nil
}

type jsonSettings struct {
	m map[string]any // opt: flat-map of full-qualified keys may be faster
}

func newJSONSettings(b []byte) (*jsonSettings, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var s jsonSettings
	if err := d.Decode(&s.m); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *jsonSettings) getFirst(keys ...string) (string, error) {
	for _, k := range keys {
		v, err := s.get(k)
		if err != nil {
			return "", err
		}
		if v != "" {
			return v, nil
		}
	}
	return "", nil // no values
}

func (s *jsonSettings) get(key string) (string, error) {
	m := s.m
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
		case json.Number:
			return t.String(), nil
		case bool:
			return strconv.FormatBool(t), nil
		default:
			return "", fmt.Errorf("non-string value: %s: %t(%v)", key, val, val)
		}
	}
	return "", nil // no value
}

type jsonGetter struct {
	settings *jsonSettings
}

// NewJSONGetter returns a static Getter backed by the given JSON.
// TODO https://smartcontract-it.atlassian.net/browse/CAPPL-775
// NewJSONRegistry with polling & subscriptions
func NewJSONGetter(b []byte) (Getter, error) {
	s, err := newJSONSettings(b)
	if err != nil {
		return nil, err
	}
	return &jsonGetter{settings: s}, nil
}

func (j *jsonGetter) GetScoped(ctx context.Context, scope Scope, key string) (value string, err error) {
	keys, err := scope.rawKeys(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get raw keys: %w", err)
	}
	return j.settings.getFirst(keys...)
}
