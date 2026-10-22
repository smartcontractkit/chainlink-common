package cli

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/huandu/go-clone"
	"github.com/spf13/cobra"
)

// Profile builds the registered struct with the function its key's value picks. The key must be a top-level field of
// T bound as a flag, such as &c.ID; to profile a section, register it on its own with [Binder.RegisterInNamespace], or
// use [ProfileWithSelector]. The function gets T's defaults, as they are at Register, with the key already set, and
// changes what it needs, zeros included. A flag, env var, or config file value still takes priority:
//
//	b.RegisterInNamespace(root, "Chain", &chain, cli.Profile(
//		func(c *ChainConfig) *uint64 { return &c.ID },
//		map[uint64]func(*ChainConfig){
//			137: func(c *ChainConfig) { c.RPC = "https://polygon.example.com"; c.Finality = 0 },
//		}))
func Profile[T any, K comparable](key func(*T) *K, profiles map[K]func(*T)) RegisterOption[T] {
	return ProfileWithSelector(func(t *T) *T { return t }, key, profiles)
}

// ProfileWithSelector is [Profile] for the section of T that section returns. The key must be a top-level field of
// that section. For a pointer section, S is the pointer, so its function gets a pointer to it, such as **ChainConfig;
// the section holds the key, so it is allocated again if the function leaves it nil. A key under a nil section picks
// no profile. Register fails if another profile's section holds this one or is held by it, as each profile rebuilds
// its whole section and two would drop each other's values:
//
//	b.Register(root, &cfg, cli.ProfileWithSelector(
//		func(c *Config) *ChainConfig { return &c.Chain },
//		func(c *ChainConfig) *uint64 { return &c.ID },
//		map[uint64]func(*ChainConfig){
//			137: func(c *ChainConfig) { c.RPC = "https://polygon.example.com"; c.Finality = 0 },
//		}))
func ProfileWithSelector[T, S any, K comparable](
	section func(*T) *S, key func(*S) *K, profiles map[K]func(*S),
) RegisterOption[T] {
	return RegisterOption[T]{setup: func(entry *typedEntry[T]) (func(), error) {
		selector, err := locateProfile(entry, section, key)
		if err != nil {
			return nil, err
		}

		scope := selector.goPath[:len(selector.goPath)-1]
		for _, other := range entry.profileSelectors {
			if scopesNest(other.goPath[:len(other.goPath)-1], scope) {
				return nil, fmt.Errorf("profiles for %s and %s overlap: each rebuilds its whole section, so one would drop the other's values", other.key, selector.key)
			}
		}

		entry.profileSelectors = append(entry.profileSelectors, selector)

		// Snapshot once, like flag defaults, so every Execute starts from the defaults as they were at Register.
		defaults := snapshotSection[S](entry, scope)
		// Copied, so a change to the caller's map can't skip the checks.
		profiles = maps.Clone(profiles)
		if err := checkProfileKeys(profiles); err != nil {
			return nil, err
		}

		return func() {
			entry.profiles = append(entry.profiles, func(supplied map[string]sourceLayer) error {
				return applyProfile(entry, supplied, selector, scope, clone.Slowly(defaults).(S), profiles)
			})
			help := entry.cmd.HelpFunc()
			entry.cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
				help(c, args)
				_, _ = fmt.Fprintf(c.OutOrStdout(), "\nAVAILABLE PROFILES FOR --%s:\n%s", selector.flagName,
					describeProfiles(entry, selector, scope, defaults, profiles))
			})
		}, nil
	}}
}

func checkProfileKeys[K comparable, S any](profiles map[K]func(*S)) error {
	for k, build := range profiles {
		if build == nil {
			return fmt.Errorf("profile %v has no function", k)
		}

		// NaN never equals itself, so no lookup could select its profile.
		if vk := reflect.ValueOf(k); !vk.Equal(vk) {
			return fmt.Errorf("profile %v never equals itself, as NaN doesn't, so it can't be selected", k)
		}
	}

	return nil
}

// walkEntries calls visit with each value leaf sets inside m, named as its flag would be, such as
// section.chains.mainnet.rpc. Keys are sorted, so the order is deterministic.
func walkEntries(leaf *entryLeaf, name string, m reflect.Value, depth int, visit func(string, reflect.Value)) {
	parts := strings.Split(leaf.key, "."+cliKeyPlaceholder)
	byText := func(a, b reflect.Value) int { return strings.Compare(textOf(a), textOf(b)) }
	for _, key := range slices.SortedFunc(slices.Values(m.MapKeys()), byText) {
		inner := name + "." + textOf(key) + parts[depth+1]
		v := readField(m.MapIndex(key), leaf.paths[depth])
		if depth == len(leaf.keyTypes)-1 {
			visit(inner, v)
		} else {
			walkEntries(leaf, inner, v, depth+1, visit)
		}
	}
}

// A required field left zero is omitted from help, as a source must set it.
func describeProfiles[T any, K comparable, S any](
	entry *typedEntry[T], selector leafKey, scope []string, defaults S, profiles map[K]func(*S),
) string {
	byText := func(a, b K) int { return strings.Compare(fmt.Sprint(a), fmt.Sprint(b)) }
	var out strings.Builder
	for _, k := range slices.SortedFunc(maps.Keys(profiles), byText) {
		_, _ = fmt.Fprintf(&out, "  %v:\n", k)
		// A fresh copy each time, so one profile's changes don't carry into the next.
		built := clone.Slowly(defaults).(S)
		vBuilt := reflect.ValueOf(&built).Elem()
		allocateField(vBuilt, selector.goPath[len(scope):]).Set(reflect.ValueOf(k))
		profiles[k](&built)

		for _, leaf := range entry.leaves {
			if !inScope(leaf, scope) || slices.Equal(leaf.goPath, selector.goPath) {
				continue
			}

			v := lookupField(vBuilt, leaf.goPath[len(scope):])
			if !v.IsValid() || (leaf.required && v.IsZero()) {
				continue
			}

			if text, ok := displayText(v); ok || len(leaf.entries) == 0 {
				_, _ = fmt.Fprintf(&out, "    %s = %s\n", leaf.key, text)
				continue
			}

			// A map of structs has no text of its own, so each value inside it is shown as its flag would set it.
			for _, el := range leaf.entries {
				if el.each != nil {
					continue
				}

				walkEntries(el, leaf.key, derefOrZero(v), 0, func(name string, v reflect.Value) {
					if text, ok := displayText(v); ok {
						_, _ = fmt.Fprintf(&out, "    %s = %s\n", name, text)
					}
				})
			}
		}
	}

	return out.String()
}

func locateProfile[T, S any, K comparable](
	entry *typedEntry[T], section func(*T) *S, key func(*S) *K,
) (selector leafKey, err error) {
	// Allocated, so the functions can reach through pointers.
	scratch := new(T)
	v := reflect.ValueOf(scratch).Elem()
	for _, k := range entry.leaves {
		allocateField(v, k.goPath)
	}

	sType, kType := reflect.TypeFor[S](), reflect.TypeFor[K]()
	if holdsIdentityKind(kType) {
		return leafKey{}, fmt.Errorf("profile key type %s compares by identity, so no value a source sets would match a profile", kType)
	}

	wantSection, wantKey, err := callSelector(scratch, section, key)
	if err != nil {
		return leafKey{}, err
	}

	for _, k := range entry.leaves {
		parent := lookupField(v, k.goPath[:len(k.goPath)-1])
		// Types are checked as a struct shares its first field's address.
		if parent.Type() != sType || parent.Addr().Pointer() != wantSection {
			continue
		}

		if k.flag != nil && k.goType == kType &&
			lookupField(parent, k.goPath[len(k.goPath)-1:]).Addr().Pointer() == wantKey {
			return k, nil
		}
	}

	return leafKey{}, errors.New("profile key must return a top-level field of its section bound as a flag, such as &c.ID")
}

// The section and key functions are the caller's, and can panic on anything, such as an index out of range.
func callSelector[T, S any, K comparable](scratch *T, section func(*T) *S, key func(*S) *K) (wantSection, wantKey uintptr, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("profile selector panicked: %v", r)
		}
	}()

	s := section(scratch)
	return reflect.ValueOf(s).Pointer(), reflect.ValueOf(key(s)).Pointer(), nil
}

// A scope is the Go path to a profile's section, such as [Chain] for &c.Chain.ID.
func inScope(k leafKey, scope []string) bool {
	return len(k.goPath) > len(scope) && hasPrefix(k.goPath, scope)
}

func scopesNest(a, b []string) bool {
	return hasPrefix(a, b) || hasPrefix(b, a)
}

func hasPrefix(path, prefix []string) bool {
	return len(path) >= len(prefix) && slices.Equal(path[:len(prefix)], prefix)
}

// Decoding overwrites the section, so its defaults are copied first, deeply. Each use copies them again, so that a
// profile changing a map or a pointer in its copy leaves them alone.
func snapshotSection[S, T any](entry *typedEntry[T], scope []string) S {
	v := lookupField(reflect.ValueOf(entry.dst).Elem(), scope)
	if !v.IsValid() {
		var zero S
		return zero
	}

	return clone.Slowly(v.Interface()).(S)
}

func applyProfile[T any, K comparable, S any](
	entry *typedEntry[T], supplied map[string]sourceLayer, selector leafKey, scope []string, base S, profiles map[K]func(*S),
) error {
	vTarget := reflect.ValueOf(entry.dst).Elem()
	// A section under a nil pointer that no source reached stays nil.
	if !lookupField(vTarget, scope).IsValid() &&
		!slices.ContainsFunc(entry.leaves, func(k leafKey) bool { _, set := supplied[k.key]; return set && inScope(k, scope) }) {
		return nil
	}

	vBase := reflect.ValueOf(&base).Elem()
	keyPath := selector.goPath[len(scope):]
	// Prevent an earlier Execute from impacting the target.
	selected := lookupField(vBase, keyPath)
	if _, set := supplied[selector.key]; set {
		selected = lookupField(vTarget, selector.goPath)
	}

	// A key under a nil section picks no profile, but the sources still apply.
	var key *K
	if selected.IsValid() {
		k := selected.Interface().(K)
		key = &k
		allocateField(vBase, keyPath).Set(reflect.ValueOf(k))
		if build, exists := profiles[k]; exists {
			build(&base)
		}
	}

	for _, k := range entry.leaves {
		layer, set := supplied[k.key]
		if !inScope(k, scope) || !set {
			continue
		}

		src := lookupField(vTarget, k.goPath)
		if !src.IsValid() {
			continue
		}

		dst := allocateField(vBase, k.goPath[len(scope):])
		if !dst.IsValid() {
			continue
		}

		if layer == nil {
			dst.Set(src)
			continue
		}

		for dst.Kind() == reflect.Pointer {
			if dst.IsNil() {
				dst.Set(reflect.New(dst.Type().Elem()))
			}

			dst = dst.Elem()
		}

		v, err := layer(dst)
		if err != nil {
			return fmt.Errorf("%s: %w", k.key, err)
		}

		dst.Set(v)
	}

	if key != nil {
		allocateField(vBase, keyPath).Set(reflect.ValueOf(*key))
	}

	allocateField(vTarget, scope).Set(vBase)
	return nil
}

// In the form the field's flag takes.
func displayText(v reflect.Value) (string, bool) {
	v = derefOrZero(v)
	switch t := v.Type(); {
	case canText(t):
		return textOf(v), true
	case t.Kind() == reflect.Slice && t.Elem().Kind() != reflect.Pointer && canText(t.Elem()):
		return writeCSV(textListOf(v)), true
	case t.Kind() == reflect.Map && t.Elem().Kind() != reflect.Pointer && canText(t.Key()) && canText(t.Elem()):
		return writeCSV(textMapOf(v)), true
	}

	return cannotDisplay, false
}

// Unlike allocateField, it never allocates, so reading a key or a section leaves nil sections nil.
func lookupField(v reflect.Value, path []string) reflect.Value {
	for _, p := range path {
		for v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}
			}

			v = v.Elem()
		}

		v = v.FieldByName(p)
	}

	return v
}
