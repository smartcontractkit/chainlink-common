package settings

import (
	"context"
	"fmt"
	"strconv"
)

// FeatureFlag represents a single feature flag with an activation threshold and optional metadata.
type FeatureFlag struct {
	Name       string            `json:"Name"`
	ActivateAt int64             `json:"ActivateAt"`
	Metadata   map[string]string `json:"Metadata,omitempty"`
}

// FeatureFlags holds a collection of feature flags at a given scope.
// It integrates with InitConfig and follows the same scoped precedence as Setting[T].
//
// Flags defined in the schema serve as defaults. The Getter provides scoped overrides
// using dot-separated keys: <FeatureFlags key>.<flagName>.ActivateAt
//
// Use With(getter) to bind a Getter, so callers of IsActive/GetFlag/GetMetadata
// don't need to pass one explicitly.
type FeatureFlags struct {
	Flags []FeatureFlag `json:"Flags,omitempty"`

	// key is the settings path prefix assigned by InitConfig (e.g. "PerWorkflow.FeatureFlags").
	// It positions this collection in the settings tree for scoped getter lookups.
	key    string
	scope  Scope
	getter Getter
}

var _ keySetter = &FeatureFlags{}

func (f *FeatureFlags) initSetting(key string, scope Scope, _ *string) error {
	f.key = key
	f.scope = scope
	return nil
}

func (f *FeatureFlags) GetKey() string  { return f.key }
func (f *FeatureFlags) GetScope() Scope { return f.scope }

// With returns a copy of FeatureFlags bound to the given Getter.
// The returned value can be used without passing a Getter to each method call.
func (f FeatureFlags) With(g Getter) FeatureFlags {
	f.getter = g
	return f
}

func (f *FeatureFlags) getDefault(name string) *FeatureFlag {
	for i := range f.Flags {
		if f.Flags[i].Name == name {
			return &f.Flags[i]
		}
	}
	return nil
}

// GetFlag looks up a feature flag by name, checking scoped overrides via the bound Getter first,
// then falling back to the default flags defined in the schema.
// The returned FeatureFlag is a copy; mutating it has no effect on the stored flags.
func (f *FeatureFlags) GetFlag(ctx context.Context, name string) (*FeatureFlag, error) {
	if f.getter != nil {
		activateAtKey := f.key + "." + name + ".ActivateAt"
		str, err := f.getter.GetScoped(ctx, f.scope, activateAtKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get feature flag %s: %w", name, err)
		}
		if str != "" {
			activateAt, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid ActivateAt for flag %s: %w", name, err)
			}
			flag := &FeatureFlag{
				Name:       name,
				ActivateAt: activateAt,
			}
			if def := f.getDefault(name); def != nil && def.Metadata != nil {
				flag.Metadata = make(map[string]string, len(def.Metadata))
				for k, v := range def.Metadata {
					flag.Metadata[k] = v
				}
			}
			return flag, nil
		}
	}
	if def := f.getDefault(name); def != nil {
		cp := *def
		return &cp, nil
	}
	return nil, nil
}

// IsActive returns true if the named flag exists and executionTimestamp >= ActivateAt.
func (f *FeatureFlags) IsActive(ctx context.Context, name string, executionTimestamp int64) (bool, error) {
	flag, err := f.GetFlag(ctx, name)
	if err != nil {
		return false, err
	}
	if flag == nil {
		return false, nil
	}
	return executionTimestamp >= flag.ActivateAt, nil
}

// GetMetadata returns a single metadata value for a flag, checking scoped overrides via
// the bound Getter first (at key <FeatureFlags key>.<flagName>.Metadata.<metaKey>),
// then falling back to the default flag's metadata.
func (f *FeatureFlags) GetMetadata(ctx context.Context, flagName, metaKey string) (string, error) {
	if f.getter != nil {
		key := f.key + "." + flagName + ".Metadata." + metaKey
		str, err := f.getter.GetScoped(ctx, f.scope, key)
		if err != nil {
			return "", err
		}
		if str != "" {
			return str, nil
		}
	}
	if def := f.getDefault(flagName); def != nil && def.Metadata != nil {
		return def.Metadata[metaKey], nil
	}
	return "", nil
}
