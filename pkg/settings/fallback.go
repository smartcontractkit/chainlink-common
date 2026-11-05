package settings

import (
	"context"
	"sync"
)

var _ IsSetting[int] = SettingWithFallback[int]{}

// SettingWithFallback resolves a value through several Settings, ordered from
// most specific to most general. The first one explicitly set in the Getter
// wins; if none is set, the default is that of the last (most general) Setting,
// so an operator override of the general setting's default keeps applying
// wherever the specific ones are unset.
//
// A Setting without a Key is skipped, so a zero value can stand for "no
// override".
//
// It implements IsSetting, so it can be used wherever a Setting can, including
// by the limiters in the limits package.
type SettingWithFallback[T any] []Setting[T]

// WithFallback returns a SettingWithFallback over s, ordered from most specific
// to most general.
func WithFallback[T any](s ...Setting[T]) SettingWithFallback[T] { return s }

func (s SettingWithFallback[T]) GetSpec() SettingSpec[T] { return s }

// first returns the most specific Setting that has a Key.
func (s SettingWithFallback[T]) first() Setting[T] {
	for _, st := range s {
		if st.Key != "" {
			return st
		}
	}
	return Setting[T]{}
}

// GetKey returns the key of the most specific Setting.
func (s SettingWithFallback[T]) GetKey() string { return s.first().Key }

// GetScope returns the scope of the most specific Setting.
func (s SettingWithFallback[T]) GetScope() Scope { return s.first().Scope }

// GetUnit returns the unit of the most specific Setting.
func (s SettingWithFallback[T]) GetUnit() string { return s.first().Unit }

// GetOrDefault returns the value of the first Setting set in the Getter, or the
// default of the most general Setting, with an error if the lookup failed.
func (s SettingWithFallback[T]) GetOrDefault(ctx context.Context, g Getter) (value T, err error) {
	if len(s) == 0 {
		return value, nil
	}
	def := s[len(s)-1].DefaultValue
	if g == nil {
		return def, nil
	}
	for _, st := range s {
		if st.Key == "" {
			continue
		}
		str, err := g.GetScoped(ctx, st.Scope, st.Key)
		if err != nil {
			return def, err
		}
		if str == "" {
			continue
		}
		value, err := st.Parse(str)
		if err != nil {
			return def, err
		}
		return value, nil
	}
	return def, nil
}

// Subscribe pushes the resolved value whenever any of the underlying Settings
// changes.
func (s SettingWithFallback[T]) Subscribe(ctx context.Context, r Registry) (<-chan Update[T], func()) {
	changed := make(chan struct{}, 1)
	stopped := make(chan struct{})

	var stops []func()
	for _, st := range s {
		if st.Key == "" {
			continue
		}
		updates, stop := r.SubscribeScoped(ctx, st.Scope, st.Key)
		stops = append(stops, stop)
		go func() {
			for {
				select {
				case <-stopped:
					return
				case _, ok := <-updates:
					if !ok {
						return
					}
					select {
					case changed <- struct{}{}:
					default: // a resolve is already pending
					}
				}
			}
		}()
	}

	values := make(chan Update[T])
	go func() {
		defer close(values)
		for {
			select {
			case <-stopped:
				return
			case <-changed:
				v, err := s.GetOrDefault(ctx, r)
				select {
				case values <- Update[T]{Value: v, Err: err}:
				case <-stopped:
					return
				}
			}
		}
	}()

	return values, sync.OnceFunc(func() {
		for _, stop := range stops {
			stop()
		}
		close(stopped)
	})
}
