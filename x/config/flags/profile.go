package flags

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// RegisterProfile attaches a profile map to a command using a selector field path (e.g.
// "Chain.ID" or "chain.id"): whichever profile the selector's decoded value picks fills in every
// field of the selector's own section that the user did not set themselves.
//
// T must match the type of a struct already registered for this command via
// RegisterCommandFlags/RegisterSubcommandFlags - if more than one entry of type T is registered
// on cmd, or none are, this returns an error.
func RegisterProfile[T any, K comparable](
	cmd *cobra.Command,
	selectorFieldName string,
	profiles map[K]T,
	opts Options,
) error {
	meta := getOrCreateMeta(cmd)

	var zero T
	targetType := reflect.TypeOf(zero)
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	entry, err := findEntryByTargetType(meta, targetType)
	if err != nil {
		return err
	}

	selectorPath, err := verifySelectorType(targetType, selectorFieldName, reflect.TypeOf((*K)(nil)).Elem(), opts)
	if err != nil {
		return err
	}

	entry.profiles = append(entry.profiles, func(cmd *cobra.Command) error {
		applyProfile(cmd, entry, opts, targetType, selectorPath, profiles)
		return nil
	})
	enableProfileHelp(cmd, selectorPath, profiles)
	return nil
}

func applyProfile[T any, K comparable](
	cmd *cobra.Command,
	entry *targetEntry,
	opts Options,
	targetType reflect.Type,
	selectorPath []string,
	profiles map[K]T,
) {
	vTarget := reflect.ValueOf(entry.target)
	if vTarget.Kind() == reflect.Pointer {
		vTarget = vTarget.Elem()
	}

	profile, exists := profiles[navigateFields(vTarget, selectorPath).Interface().(K)]
	if !exists {
		// No profile for this selector value: nothing to default, not an error.
		return
	}

	// Scope the copy to the substruct owning the selector (e.g. "System" for "System.Env") so
	// this profile can't clobber unrelated branches (e.g. "Chain") that just happen to be
	// zero-valued in this profile's map entry.
	scopePath := selectorPath[:len(selectorPath)-1]
	leafName := selectorPath[len(selectorPath)-1]
	targetScope := navigateFields(vTarget, scopePath)
	profileScope := navigateFields(reflect.ValueOf(profile), scopePath)

	// The selector's own field (e.g. Chain.ID) sits inside the scoped substruct too; preserve the
	// value that was actually used to pick this profile, since the profile's map entry usually
	// leaves it zero (the profile is keyed by that value, not describing it).
	leafField := targetScope.FieldByName(leafName)
	selectedValue := reflect.ValueOf(leafField.Interface())

	if profileScope.Kind() == reflect.Pointer {
		profileScope = profileScope.Elem()
	}
	copyDefaults(cmd, entry, opts, scopePrefix(targetType, scopePath, opts), targetScope, profileScope)

	leafField.Set(selectedValue)
}

func findEntryByTargetType(meta *commandMetaData, targetType reflect.Type) (*targetEntry, error) {
	var match *targetEntry
	for _, e := range meta.entries {
		et := reflect.TypeOf(e.target)
		if et.Kind() == reflect.Pointer {
			et = et.Elem()
		}
		if et != targetType {
			continue
		}
		if match != nil {
			return nil, fmt.Errorf("multiple registered targets of type %s on this command; ambiguous profile registration", targetType)
		}
		match = e
	}
	if match == nil {
		return nil, fmt.Errorf("no registered target of type %s on this command; call RegisterCommandFlags/RegisterSubcommandFlags first", targetType)
	}
	return match, nil
}

func verifySelectorType(tType reflect.Type, selectorFieldName string, kType reflect.Type, opts Options) ([]string, error) {
	if tType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target type T must be a struct")
	}

	path, fType, found := findFieldByTagOrName(tType, selectorFieldName, opts)
	if !found {
		return nil, fmt.Errorf("field or tag %q not found in struct %s", selectorFieldName, tType.Name())
	}

	if fType != kType {
		return nil, fmt.Errorf("type mismatch for %q in %s: field is %s, profile key K is %s", selectorFieldName, tType.Name(), fType, kType)
	}

	return path, nil
}

func findFieldByTagOrName(t reflect.Type, name string, opts Options) ([]string, reflect.Type, bool) {
	if !strings.Contains(name, ".") {
		return findSingleField(t, name, opts)
	}

	curr := t
	var fullPath []string
	for _, part := range strings.Split(name, ".") {
		if curr.Kind() == reflect.Pointer {
			curr = curr.Elem()
		}
		if curr.Kind() != reflect.Struct {
			return nil, nil, false
		}

		subPath, fType, ok := findSingleField(curr, part, opts)
		if !ok {
			return nil, nil, false
		}
		fullPath = append(fullPath, subPath...)
		curr = fType
	}
	return fullPath, curr, true
}

func findSingleField(t reflect.Type, name string, opts Options) ([]string, reflect.Type, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		keyName, _ := opts.tagKey(field)
		if strings.EqualFold(field.Name, name) || strings.EqualFold(keyName, name) {
			return []string{field.Name}, field.Type, true
		}

		elemType := field.Type
		if elemType.Kind() == reflect.Pointer {
			elemType = elemType.Elem()
		}
		if elemType.Kind() == reflect.Struct {
			if subPath, fType, ok := findSingleField(elemType, name, opts); ok {
				return append([]string{field.Name}, subPath...), fType, true
			}
		}
	}
	return nil, nil, false
}

// navigateFields dereferences a pointer at every step, so a path may cross an allocated
// optional section without the caller checking.
func navigateFields(v reflect.Value, path []string) reflect.Value {
	curr := v
	for _, p := range path {
		if curr.Kind() == reflect.Pointer {
			curr = curr.Elem()
		}
		curr = curr.FieldByName(p)
	}
	return curr
}

// scopePrefix computes the dotted config key prefix (e.g. "system") that corresponds to the given
// Go field-name path (e.g. ["System"]) starting from struct type t.
func scopePrefix(t reflect.Type, path []string, opts Options) string {
	var parts []string
	curr := t
	for _, name := range path {
		if curr.Kind() == reflect.Pointer {
			curr = curr.Elem()
		}
		field, ok := curr.FieldByName(name)
		if !ok {
			break
		}
		key, _ := opts.tagKey(field)
		parts = append(parts, key)
		curr = field.Type
	}
	return strings.Join(parts, ".")
}

func copyDefaults(cmd *cobra.Command, entry *targetEntry, opts Options, prefix string, tVal, pVal reflect.Value) {
	tType := tVal.Type()
	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		key, _ := opts.tagKey(field)
		relKey := key
		if prefix != "" {
			relKey = prefix + "." + key
		}
		viperKey := relKey
		if entry.namespace != "" {
			viperKey = entry.namespace + "." + relKey
		}

		targetField, profileField := tVal.Field(i), pVal.Field(i)

		if targetField.Kind() == reflect.Pointer {
			if profileField.IsNil() {
				continue
			}
			if targetField.IsNil() {
				targetField.Set(reflect.New(targetField.Type().Elem()))
			}
			targetField, profileField = targetField.Elem(), profileField.Elem()
		}

		switch {
		case targetField.Kind() == reflect.Struct:
			copyDefaults(cmd, entry, opts, relKey, targetField, profileField)

		// A list of structs has no flag of its own - its fields are bound column-wise underneath
		// it (see nested.go) - so "did the user set this?" has to ask about every key under the
		// list, not about a flag named after the list itself.
		case isStructList(targetField.Type()):
			if !entry.anyKeySetUnder(cmd, viperKey) {
				targetField.Set(profileField)
			}

		default:
			if !entry.isExplicitlySet(cmd, flagNameFromViperKey(relKey), viperKey) {
				targetField.Set(profileField)
			}
		}
	}
}

func enableProfileHelp[T any, K comparable](cmd *cobra.Command, selectorPath []string, profiles map[K]T) {
	selectorFlagName := strings.ReplaceAll(strings.ToLower(selectorPath[len(selectorPath)-1]), "_", "-")
	existingHelp := cmd.HelpFunc()

	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if existingHelp != nil {
			existingHelp(c, args)
		} else {
			fmt.Printf("Usage of %s:\n\n", c.CommandPath())
			c.Flags().VisitAll(func(f *pflag.Flag) {
				defaultMsg := ""
				if f.DefValue != "" {
					defaultMsg = fmt.Sprintf(" (default %q)", f.DefValue)
				}
				fmt.Printf("  --%-35s %s%s\n", f.Name, f.Usage, defaultMsg)
			})
		}

		fmt.Printf("\nAVAILABLE PROFILES FOR --%s:\n", selectorFlagName)
		for k := range profiles {
			fmt.Printf("  - %v\n", k)
		}
	})
}
