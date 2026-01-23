package settings

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
)

// PerChainSelector returns a new SettingMap for the given values, which is keyed on
// chain selector from the context.Context.
func PerChainSelector[T any](defaultValue Setting[T], vals map[string]T) SettingMap[T] {
	svals := make(map[string]string, len(vals))
	for k, v := range vals {
		svals[k] = fmt.Sprint(v)
	}
	return SettingMap[T]{
		Default:    defaultValue,
		Values:     svals,
		KeyFromCtx: contexts.ChainSelectorValue,
	}
}

var _ IsSetting[int] = SettingMap[int]{}

type SettingMap[T any] struct {
	Default    Setting[T]
	Values     map[string]string                     // unparsed
	KeyFromCtx func(context.Context) (uint64, error) `json:"-" toml:"-"`
}

func (s SettingMap[T]) GetSpec() SettingSpec[T] { return &s }

func (s *SettingMap[T]) GetKey() string { return s.Default.Key }

func (s *SettingMap[T]) GetScope() Scope { return s.Default.Scope }

func (s *SettingMap[T]) GetUnit() string { return s.Default.Unit }

func (s *SettingMap[T]) initSetting(key string, scope Scope, unit *string) error {
	if s.KeyFromCtx == nil {
		return errors.New("missing KeyFromCtx func")
	}
	return s.Default.initSetting(key, scope, unit)
}

// GetOrDefault gets the setting from the Getter for the given Scope, or returns the default value with an error.
func (s *SettingMap[T]) GetOrDefault(ctx context.Context, g Getter) (value T, err error) {
	if s.KeyFromCtx == nil {
		return s.Default.DefaultValue, errors.New("missing KeyFromCtx func")
	}
	k, err := s.KeyFromCtx(ctx)
	if err != nil {
		return s.Default.DefaultValue, fmt.Errorf("failed to get value from context: %w", err)
	}
	valueOrDefault := func() (T, error) {
		if str, ok := s.Values[strconv.FormatUint(k, 10)]; ok {
			value, err = s.Default.Parse(str)
			if err != nil {
				return s.Default.DefaultValue, err
			}
			return value, nil
		}
		return s.Default.DefaultValue, nil
	}
	if g == nil {
		return valueOrDefault()
	}

	valueKey := s.Default.Key + ".Values." + strconv.FormatUint(k, 10)
	defaultKey := s.Default.Key + ".Default"

	// Values override
	str, err := g.GetScoped(ctx, s.Default.Scope, valueKey)
	if err != nil {
		return s.Default.DefaultValue, err
	} else if str != "" {
		value, err = s.Default.Parse(str)
		if err != nil {
			return valueOrDefault()
		}
		return
	}

	// Default override
	str, err = g.GetScoped(ctx, s.Default.Scope, defaultKey)
	if err != nil || str == "" {
		return valueOrDefault()
	}

	value, err = s.Default.Parse(str)
	if err != nil {
		return valueOrDefault()
	}
	return
}

func (s *SettingMap[T]) Subscribe(ctx context.Context, registry Registry) (<-chan Update[T], func()) {
	//TODO subscribe to Values & Default

	// no-op
	ch := make(chan Update[T])
	return ch, sync.OnceFunc(func() { close(ch) })
}
