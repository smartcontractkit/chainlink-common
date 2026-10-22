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
// that section. For a pointer section, S is the pointer, so its function gets a **T it can allocate or leave nil; a
// key under a nil section picks no profile:
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
	return RegisterOption[T]{setup: func(entry *targetEntry) (func(), error) {
		selector, err := locateProfile(entry, section, key)
		if err != nil {
			return nil, err
		}

		scope := selector.goPath[:len(selector.goPath)-1]
		return func() {
			// Snapshot once, like flag defaults, so every Execute starts from the defaults as they were at Register.
			defaults := snapshotSection[S](entry, scope)
			entry.profiles = append(entry.profiles, func(cc commandConfig) {
				applyProfile(entry, cc, selector, scope, clone.Clone(defaults).(S), profiles)
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

// A required field the profile leaves at its default is omitted from help display
func describeProfiles[K comparable, S any](
	entry *targetEntry, selector leafKey, scope []string, defaults S, profiles map[K]func(*S),
) string {
	byText := func(a, b K) int { return strings.Compare(fmt.Sprint(a), fmt.Sprint(b)) }
	vDefaults := reflect.ValueOf(&defaults).Elem()
	var out strings.Builder
	for _, k := range slices.SortedFunc(maps.Keys(profiles), byText) {
		_, _ = fmt.Fprintf(&out, "  %v:\n", k)
		// A fresh copy each time, so one profile's changes don't carry into the next.
		built := clone.Clone(defaults).(S)
		vBuilt := reflect.ValueOf(&built).Elem()
		allocateField(vBuilt, selector.goPath[len(scope):]).Set(reflect.ValueOf(k))
		profiles[k](&built)

		for _, leaf := range entry.keys {
			if !inScope(leaf, scope) || slices.Equal(leaf.goPath, selector.goPath) {
				continue
			}

			path := leaf.goPath[len(scope):]
			if leaf.required && !changed(vDefaults, vBuilt, path) {
				continue
			}

			_, _ = fmt.Fprintf(&out, "    %s = %s\n", leaf.key, textOf(lookupField(vBuilt, path)))
		}
	}

	return out.String()
}

func locateProfile[T, S any, K comparable](
	entry *targetEntry, section func(*T) *S, key func(*S) *K,
) (selector leafKey, err error) {
	// Allocated, so the functions can reach through pointers.
	scratch := new(T)
	v := reflect.ValueOf(scratch).Elem()
	for _, k := range entry.keys {
		allocateField(v, k.goPath)
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("profile selector panicked: %v", r)
		}
	}()

	s := section(scratch)
	wantKey := reflect.ValueOf(key(s)).Pointer()
	wantSection := reflect.ValueOf(s).Pointer()
	for _, k := range entry.keys {
		parent := lookupField(v, k.goPath[:len(k.goPath)-1])
		// Types are checked as a struct shares its first field's address.
		if parent.Type() != reflect.TypeFor[S]() || parent.Addr().Pointer() != wantSection {
			continue
		}

		if k.goType == reflect.TypeFor[K]() &&
			lookupField(parent, k.goPath[len(k.goPath)-1:]).Addr().Pointer() == wantKey {
			return k, nil
		}
	}

	return leafKey{}, errors.New("profile key must return a top-level field of its section bound as a flag, such as &c.ID")
}

func inScope(k leafKey, scope []string) bool {
	return len(k.goPath) > len(scope) && slices.Equal(k.goPath[:len(scope)], scope)
}

// Decoding overwrites the section, so its defaults are copied first, deeply. Each use copies them again, so that a
// profile changing a map or a pointer in its copy leaves them alone.
func snapshotSection[S any](entry *targetEntry, scope []string) S {
	v := lookupField(reflect.ValueOf(entry.target).Elem(), scope)
	if !v.IsValid() {
		var zero S
		return zero
	}

	return clone.Clone(v.Interface()).(S)
}

func applyProfile[K comparable, S any](
	entry *targetEntry, cc commandConfig, selector leafKey, scope []string, base S, profiles map[K]func(*S),
) {
	vTarget := reflect.ValueOf(entry.target).Elem()
	selected := lookupField(vTarget, selector.goPath)
	if !selected.IsValid() {
		// The key's section is nil: no profile requested.
		return
	}

	// A key with no profile, such as a chain that isn't built in, leaves its section to the sources.
	build, exists := profiles[selected.Interface().(K)]
	if !exists {
		return
	}

	vBase := reflect.ValueOf(&base).Elem()
	allocateField(vBase, selector.goPath[len(scope):]).Set(selected)
	defaults := reflect.ValueOf(clone.Clone(base))
	build(&base)

	var supplied [][]string
	for _, k := range entry.keys {
		if !inScope(k, scope) {
			continue
		}

		// decodeEntry already reported any error.
		vals, _ := entry.sources(k, cc)
		switch {
		case len(vals) > 0:
			supplied = append(supplied, k.goPath)
		case slices.Equal(k.goPath, selector.goPath):
		case changed(defaults, vBase, k.goPath[len(scope):]):
			// So a profile can satisfy a `set` rule its default can't.
			supplied = append(supplied, k.goPath)
			continue
		default:
			continue
		}

		// Sources and the key beat the profile, so it can't switch what was asked for.
		if src := lookupField(vTarget, k.goPath); src.IsValid() {
			if dst := allocateField(vBase, k.goPath[len(scope):]); dst.IsValid() {
				dst.Set(src)
			}
		}
	}

	allocateField(vTarget, scope).Set(vBase)

	// Replacing a pointer moves the fields under it, so every supplied field is recorded at its new address.
	for _, path := range supplied {
		f := lookupField(vTarget, path)
		for f.IsValid() && f.Kind() == reflect.Pointer && !f.IsNil() {
			f = f.Elem()
		}

		if f.IsValid() && f.Kind() != reflect.Pointer {
			entry.suppliedFields[f.Addr().Pointer()] = true
		}
	}
}

func changed(from, to reflect.Value, path []string) bool {
	a, b := lookupField(from, path), lookupField(to, path)
	if !a.IsValid() || !b.IsValid() {
		return a.IsValid() != b.IsValid()
	}

	return !reflect.DeepEqual(a.Interface(), b.Interface())
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
