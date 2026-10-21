package cli

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/spf13/pflag"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

// Stand for a map key in an entry's flag name, such as chains.<key>.rpc, and its env vars'.
const cliKeyPlaceholder = "<key>"
const envKeyPlaceholder = "<KEY>"

// entryLeaf sets one value inside a map's entries, as a struct's leaf does inside the struct: chains.<key>.rpc for
// Chains map[string]Chain, or labels.<key> for Labels map[string]string.
type entryLeaf struct {
	// key holds cliKeyPlaceholder for each map key.
	key string
	// keyTypes[i] parses map key i, and paths[i] leads from its entry to the next map or the value.
	keyTypes []reflect.Type
	paths    [][]string
	goType   reflect.Type
	// each sets one entry of the map this leaf holds, so a whole map can be split into its entries.
	each *entryLeaf
}

type entryScope struct {
	owner    *leafKey
	keyTypes []reflect.Type
	// paths[i] leads from an entry of map i to map i+1. The path inside the innermost entry is the walk's goPath, as
	// it is still growing.
	paths [][]string
}

// bindEntries adds a flag for each value inside t's entries and returns the one that sets a whole entry. A map read
// whole, or keyed by something with no text form, has none.
func (e *typedEntry[T]) bindEntries(owner *leafKey, t reflect.Type, scope walkScope, doc string) (*entryLeaf, error) {
	t = commentparsing.DerefType(t)
	lang := e.b.opts.Markup
	if !mergesByKey(t, lang) || !canText(t.Key()) || holdsIdentityKind(t.Key()) || slices.Contains(scope.ancestors, t) {
		return nil, nil
	}

	in := &entryScope{owner: owner, keyTypes: []reflect.Type{t.Key()}}
	if scope.entry != nil {
		in.keyTypes = append(slices.Clone(scope.entry.keyTypes), t.Key())
		in.paths = append(slices.Clone(scope.entry.paths), scope.goPath)
	}

	next := scope
	next.key += "." + cliKeyPlaceholder
	next.goPath = nil
	next.entry = in
	next.ancestors = append(slices.Clone(scope.ancestors), t)

	elem := commentparsing.DerefType(t.Elem())
	if elem.Kind() == reflect.Struct && !lang.IsLeaf(t.Elem()) && !lang.IsLeaf(elem) && !slices.Contains(next.ancestors, elem) {
		next.ancestors = append(next.ancestors, elem)
		return nil, e.walk(reflect.Zero(elem), next)
	}

	return e.bindEntryLeaf(next, t.Elem(), doc)
}

// The flag added here only shows in --help: pflag parses one with the keys filled in, which entryFlag adds.
func (e *typedEntry[T]) bindEntryLeaf(scope walkScope, t reflect.Type, doc string) (*entryLeaf, error) {
	leaf := &entryLeaf{
		key:      scope.key,
		keyTypes: scope.entry.keyTypes,
		paths:    append(slices.Clone(scope.entry.paths), scope.goPath),
		goType:   t,
	}

	usage := doc
	if envs := e.envVars(leafKey{key: leaf.key}); len(envs) > 0 {
		usage = strings.TrimSpace(usage + " [env " + strings.Join(envs, ", ") + "]")
	}

	if f, _ := newFlag(strings.ReplaceAll(leaf.key, "_", "-"), t, reflect.Zero(commentparsing.DerefType(t)), usage); f != nil {
		flags := e.cmd.PersistentFlags()
		if flags.Lookup(f.Name) != nil {
			return nil, fmt.Errorf("%s: flag --%s is already defined on %s", leaf.key, f.Name, e.cmd.Name())
		}

		f.Value = templateValue{f.Value}
		flags.AddFlag(f)
		e.b.entryTemplates[f] = leaf
		scope.entry.owner.entries = append(scope.entry.owner.entries, leaf)
		if doc == "" {
			e.addUndocumented(leaf.key)
		}
	}

	var err error
	leaf.each, err = e.bindEntries(scope.entry.owner, t, scope, doc)
	return leaf, err
}

type templateValue struct{ pflag.Value }

func (templateValue) Set(string) error {
	return fmt.Errorf("replace %s with a map key", cliKeyPlaceholder)
}

type errValue struct {
	pflag.Value
	err error
}

func (v errValue) Set(string) error { return v.err }

type entryFlag struct {
	leaf  *entryLeaf
	keys  []reflect.Value
	value func() reflect.Value
}

// addsEntryFlags returns a normalization func with which pflag, looking up a flag it doesn't have while parsing, finds
// one for the map entry its name holds, such as --chains.mainnet.rpc. cobra parses flags before any hook runs, so this
// is the only chance.
func (b *Binder) addsEntryFlags(existing func(*pflag.FlagSet, string) pflag.NormalizedName) func(*pflag.FlagSet, string) pflag.NormalizedName {
	return func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		n := pflag.NormalizedName(name)
		if existing != nil {
			n = existing(f, name)
		}

		// Looking up and adding a flag normalize its name too.
		if b.addingEntryFlag || !f.Parsed() || strings.Contains(string(n), cliKeyPlaceholder) {
			return n
		}

		b.addingEntryFlag = true
		defer func() { b.addingEntryFlag = false }()
		if f.Lookup(string(n)) == nil {
			if flag := b.entryFlag(f, string(n)); flag != nil {
				f.AddFlag(flag)
			}
		}

		return n
	}
}

// A key that doesn't parse still gets a flag, which fails when set. For Chains map[uint32]string, --chains.x fails with
// key "x": strconv.ParseUint: parsing "x": invalid syntax, rather than unknown flag: --chains.x.
func (b *Binder) entryFlag(f *pflag.FlagSet, name string) *pflag.Flag {
	var leaf *entryLeaf
	var usage string
	var texts []string
	f.VisitAll(func(template *pflag.Flag) {
		if l, ok := b.entryTemplates[template]; ok && leaf == nil {
			if keys, ok := matchKeys(template.Name, cliKeyPlaceholder, ".", name); ok {
				leaf, usage, texts = l, template.Usage, keys
			}
		}
	})

	if leaf == nil {
		return nil
	}

	flag, get := newFlag(name, leaf.goType, reflect.Zero(commentparsing.DerefType(leaf.goType)), usage)
	keys := make([]reflect.Value, len(texts))
	for i, text := range texts {
		key, err := mapKey(leaf.keyTypes[i], text, map[any]string{})
		if err != nil {
			flag.Value = errValue{flag.Value, err}
			return flag
		}

		keys[i] = key
	}

	b.entryFlags[flag] = entryFlag{leaf, keys, get}
	return flag
}

// A flag added some other way, such as cmd.Flags().String("map.a", ...) beside Map map[string]string, would be parsed
// in place of the entry flag for map.a, so it's an error.
func (b *Binder) shadowedEntries(flags *pflag.FlagSet) error {
	var templates []*pflag.Flag
	flags.VisitAll(func(f *pflag.Flag) {
		if _, ok := b.entryTemplates[f]; ok {
			templates = append(templates, f)
		}
	})

	var err error
	flags.VisitAll(func(f *pflag.Flag) {
		_, isTemplate := b.entryTemplates[f]
		if _, isEntry := b.entryFlags[f]; isEntry || isTemplate || err != nil {
			return
		}

		for _, template := range templates {
			if _, ok := matchKeys(template.Name, cliKeyPlaceholder, ".", f.Name); ok {
				err = fmt.Errorf("flag --%s would set --%s; rename it", f.Name, template.Name)
				return
			}
		}
	})

	return err
}

// matchKeys returns the keys name puts in template's placeholders, such as [mainnet] for chains.mainnet.rpc and
// chains.<key>.rpc. A key is one segment, so a name fills a template one way or not at all.
func matchKeys(template, placeholder, sep, name string) ([]string, bool) {
	want, got := strings.Split(template, sep), strings.Split(name, sep)
	if len(want) != len(got) {
		return nil, false
	}

	var keys []string
	for i, segment := range want {
		switch {
		case segment == placeholder && got[i] != "":
			keys = append(keys, got[i])
		case segment != got[i]:
			return nil, false
		}
	}

	return keys, len(keys) > 0
}

// overlap returns a name both templates match, if there is one: a placeholder matches any segment.
func overlap(a, b, placeholder, sep string) (string, bool) {
	as, bs := strings.Split(a, sep), strings.Split(b, sep)
	if len(as) != len(bs) {
		return "", false
	}

	name := make([]string, len(as))
	for i := range as {
		switch {
		case as[i] == placeholder && bs[i] == placeholder:
			name[i] = strings.Trim(placeholder, "<>")
		case as[i] == placeholder:
			name[i] = bs[i]
		case bs[i] == placeholder || as[i] == bs[i]:
			name[i] = as[i]
		default:
			return "", false
		}
	}

	return strings.Join(name, sep), true
}

// entrySet is one value a source sets inside a map, such as rpc in entry mainnet.
type entrySet struct {
	leaf *entryLeaf
	keys []reflect.Value
	// texts holds the leading keys still to resolve, named by env vars; keys holds the rest, already parsed.
	texts  []string
	value  reflect.Value
	source string
	// rank is the env var prefix's index, as the first prefix set wins.
	rank int
}

// Keys named by an env var resolve against the map the entry layers over, which holds every source below env vars.
func (s entrySet) resolve(m reflect.Value) ([]reflect.Value, error) {
	if s.texts == nil {
		return s.keys, nil
	}

	keys := make([]reflect.Value, 0, len(s.texts)+len(s.keys))
	for i, text := range s.texts {
		keyText, err := envKeyText(entryMap(m, s.leaf, keys), text)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", s.source, err)
		}

		key, err := mapKey(s.leaf.keyTypes[i], keyText, map[any]string{})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", s.source, err)
		}

		keys = append(keys, key)
	}

	return append(keys, s.keys...), nil
}

// decodeEntries layers a map's sources lowest first, each setting whole entries or values inside them, so
// --chains.mainnet.rpc changes only rpc in the file's mainnet. It returns how the sources layer over a map, nil if
// none set it.
func (e *typedEntry[T]) decodeEntries(k leafKey, cc commandConfig, dst reflect.Value) (sourceLayer, error) {
	t := commentparsing.DerefType(k.goType)
	var file reflect.Value
	if raw, ok := cc.keys.lookup(cc.fileValues, k.fileKey); ok {
		var err error
		if file, err = fromFile(t, raw); err != nil {
			return nil, err
		}
	}

	envSets, err := e.envEntrySets(k, cc.fieldEnvVars)
	if err != nil {
		return nil, err
	}

	layers := [][]entrySet{envSets, e.flagEntrySets(k, cc.flags)}
	if !file.IsValid() && len(layers[0]) == 0 && len(layers[1]) == 0 {
		return nil, nil
	}

	layer := func(under reflect.Value) (reflect.Value, error) {
		m := mergeMaps(t, under)
		if file.IsValid() {
			m = mergeMaps(t, m, file)
		}

		for _, sets := range layers {
			// Resolved against m, before any set entry can add a key another set's name would then match.
			resolved := make([]entrySet, 0, len(sets))
			for _, s := range sets {
				keys, kerr := s.resolve(m)
				if kerr != nil {
					return reflect.Value{}, kerr
				}

				s.keys, s.texts = keys, nil
				resolved = append(resolved, s)
			}

			unique, uerr := oneValueEach(resolved)
			if uerr != nil {
				return reflect.Value{}, uerr
			}

			for _, s := range unique {
				setEntry(m, s.leaf, s.keys, 0, s.value)
			}
		}

		return m, nil
	}

	f := allocateField(dst, k.goPath)
	for f.Kind() == reflect.Pointer {
		if f.IsNil() {
			f.Set(reflect.New(f.Type().Elem()))
		}

		f = f.Elem()
	}

	// A profile replaces its section's maps, so the keys an env var names are resolved against the profile's map
	// when it applies this layer, not the one here.
	if slices.ContainsFunc(e.profileSelectors, func(s leafKey) bool { return inScope(k, s.goPath[:len(s.goPath)-1]) }) {
		return layer, nil
	}

	v, err := layer(readField(dst, k.goPath))
	if err != nil {
		return nil, err
	}

	f.Set(v)
	return layer, nil
}

// An env var a field uses sets that field, even if its name also reads as a map entry: with TelemetryAuthHeaders and
// TelemetryAuthHeadersTTL, APP_TELEMETRY_AUTH_HEADERS_TTL is the latter.
func (e *typedEntry[T]) envEntrySets(k leafKey, fieldEnvVars map[string]string) ([]entrySet, error) {
	var sets []entrySet
	if k.flag != nil {
		for rank, name := range e.envVars(k) {
			if s := os.Getenv(name); s != "" {
				val, err := parseText(k.goType, s)
				if err != nil {
					return nil, fmt.Errorf("invalid value %q for %s: %w", s, name, err)
				}

				sets = splitEntries(k.each, nil, nil, val, name, rank)
				break
			}
		}
	}

	environ := os.Environ()
	slices.Sort(environ)
	for _, kv := range environ {
		name, s, _ := strings.Cut(kv, "=")
		if _, isField := fieldEnvVars[name]; s == "" || isField {
			continue
		}

		leaf, rank, texts := e.envEntry(k, name)
		if leaf == nil {
			continue
		}

		val, err := parseText(leaf.goType, s)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for %s: %w", s, name, err)
		}

		sets = append(sets, entrySets(leaf, nil, texts, val, name, rank)...)
	}

	return sets, nil
}

func (e *typedEntry[T]) envEntry(k leafKey, name string) (*entryLeaf, int, []string) {
	for _, leaf := range k.entries {
		for rank, template := range e.envVars(leafKey{key: leaf.key}) {
			if keys, ok := matchKeys(template, envKeyPlaceholder, "_", name); ok {
				return leaf, rank, keys
			}
		}
	}

	return nil, 0, nil
}

func (e *typedEntry[T]) flagEntrySets(k leafKey, flags *pflag.FlagSet) []entrySet {
	var sets []entrySet
	if k.flag != nil && k.flag.Changed {
		sets = splitEntries(k.each, nil, nil, k.value(), "--"+k.flagName, 0)
	}

	flags.Visit(func(f *pflag.Flag) {
		if ef, ok := e.b.entryFlags[f]; ok && slices.Contains(k.entries, ef.leaf) {
			sets = append(sets, entrySets(ef.leaf, ef.keys, nil, ef.value(), "--"+f.Name, 0)...)
		}
	})

	return sets
}

// An env var name is upper case, so its key is the existing key that matches ignoring case, or else lower case.
func envKeyText(existing reflect.Value, text string) (string, error) {
	var found []string
	if existing.IsValid() {
		for key := range existing.Seq() {
			if keyText := textOf(key); strings.EqualFold(keyText, text) {
				found = append(found, keyText)
			}
		}
	}

	switch len(found) {
	case 0:
		return strings.ToLower(text), nil
	case 1:
		return found[0], nil
	default:
		slices.Sort(found)
		return "", fmt.Errorf("%s could be any of the keys %q", text, found)
	}
}

// The map that keys[len(keys)] would index, read without allocating.
func entryMap(m reflect.Value, leaf *entryLeaf, keys []reflect.Value) reflect.Value {
	for i, key := range keys {
		if m = m.MapIndex(key); !m.IsValid() {
			return m
		}

		m = readField(m, leaf.paths[i])
	}

	return m
}

// Reads without allocating, so an optional section left out stays nil.
func readField(v reflect.Value, path []string) reflect.Value {
	for _, p := range path {
		v = derefOrZero(v).FieldByName(p)
	}

	return derefOrZero(v)
}

// Splits a whole map into its entries: --labels a=1,b=2 becomes the same as --labels.a 1 --labels.b 2, so giving
// --labels a=1 --labels.a 3 is caught as setting it twice.
func splitEntries(each *entryLeaf, keys []reflect.Value, texts []string, m reflect.Value, source string, rank int) []entrySet {
	var sets []entrySet
	for key, val := range m.Seq2() {
		entrySource := fmt.Sprintf("%s (key %q)", source, textOf(key))
		sets = append(sets, entrySets(each, append(slices.Clone(keys), key), texts, val, entrySource, rank)...)
	}

	return sets
}

func entrySets(leaf *entryLeaf, keys []reflect.Value, texts []string, val reflect.Value, source string, rank int) []entrySet {
	if leaf.each != nil && val.Kind() == reflect.Map {
		return splitEntries(leaf.each, keys, texts, val, source, rank)
	}

	return []entrySet{{leaf: leaf, keys: keys, texts: texts, value: val, source: source, rank: rank}}
}

// Within one source, two values for one entry are an error, as two keys for one are. Across env var prefixes the
// first prefix wins, as for any env var.
func oneValueEach(sets []entrySet) ([]entrySet, error) {
	var out []entrySet
next:
	for _, s := range sets {
		for i, o := range out {
			if o.leaf != s.leaf || !slices.EqualFunc(o.keys, s.keys, reflect.Value.Equal) {
				continue
			}

			switch {
			case o.rank == s.rank:
				return nil, fmt.Errorf("%s and %s set the same value", o.source, s.source)
			case s.rank < o.rank:
				out[i] = s
			}

			continue next
		}

		out = append(out, s)
	}

	return out, nil
}

// Copies entries, map and pointer on the way down, so one a lower source or the default holds isn't changed in place.
func setEntry(m reflect.Value, leaf *entryLeaf, keys []reflect.Value, depth int, val reflect.Value) {
	entry := reflect.New(m.Type().Elem()).Elem()
	if old := m.MapIndex(keys[depth]); old.IsValid() {
		entry.Set(old)
	}

	target := ownedField(entry, leaf.paths[depth])
	if depth == len(keys)-1 {
		target.Set(val)
	} else {
		inner := mergeMaps(target.Type(), target)
		target.Set(inner)
		setEntry(inner, leaf, keys, depth+1, val)
	}

	m.SetMapIndex(keys[depth], entry)
}

func ownedField(v reflect.Value, path []string) reflect.Value {
	v = ownPointers(v)
	for _, p := range path {
		v = ownPointers(v.FieldByName(p))
	}

	return v
}

func ownPointers(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		owned := reflect.New(v.Type().Elem())
		if !v.IsNil() {
			owned.Elem().Set(v.Elem())
		}

		v.Set(owned)
		v = owned.Elem()
	}

	return v
}
