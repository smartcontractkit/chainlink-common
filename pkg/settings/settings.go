package settings

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

// Getter fetches scoped settings by key.
type Getter interface {
	// GetScoped returns the value for key, including checking for any Scope based overrides up to scope.
	// The tenant for each scope is based on the contexts.CRE in the ctx.
	GetScoped(ctx context.Context, scope Scope, key string) (value string, err error)
}

// Registry extends Getter with subscriptions.
type Registry interface {
	Getter
	// SubscribeScoped returns a channel for updates as an efficient alternative to polling.
	// Stop must be closed to terminate the subscription.
	SubscribeScoped(ctx context.Context, scope Scope, key string) (updates <-chan Update[string], stop func())
}

type IsSetting[T any] interface {
	GetSpec() SettingSpec[T]
}

type SettingSpec[T any] interface {
	GetKey() string
	GetScope() Scope
	GetUnit() string
	GetOrDefault(context.Context, Getter) (T, error)
	Subscribe(context.Context, Registry) (<-chan Update[T], func())
}

var _ IsSetting[int] = Setting[int]{}

// Setting holds a key, default value, and parsing function for a particular setting.
// Use Setting.GetOrDefault with a Getter to look up settings.
// Use Setting.Subscribe with a Registry to have updates pushed over a channel.
type Setting[T any] struct {
	Key          string
	DefaultValue T
	Scope        Scope
	Parse        func(string) (T, error) `json:"-" toml:"-"`
	Unit         string
	// FeatureFlag marks this setting as a feature flag (set via the `featureflag:"true"` field tag).
	// Interim marker, superseded by CRE Flags 2.0. Schema metadata only: never serialized or read at runtime.
	FeatureFlag bool `json:"-" toml:"-"`
}

func (s Setting[T]) GetSpec() SettingSpec[T] { return &s }

func (s *Setting[T]) GetKey() string { return s.Key }

func (s *Setting[T]) GetScope() Scope { return s.Scope }

func (s *Setting[T]) GetUnit() string { return s.Unit }

func (s *Setting[T]) info() SettingInfo {
	var zero T
	return SettingInfo{Key: s.Key, Scope: s.Scope, ValueType: reflect.TypeOf(&zero).Elem(), FeatureFlag: s.FeatureFlag}
}

func (s *Setting[T]) setFeatureFlag() { s.FeatureFlag = true }

func (s Setting[T]) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%v", s.DefaultValue), nil
}

func (s *Setting[T]) UnmarshalText(b []byte) (err error) {
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1] // unquote string
	}
	if s.Parse == nil {
		return errors.New("missing Parse func")
	}
	s.DefaultValue, err = s.Parse(string(b))
	if err != nil {
		err = fmt.Errorf("%s: failed to parse %s: %w", s.Key, string(b), err)
	}
	return
}

func (s *Setting[T]) initSetting(key string, scope Scope, unit *string) error {
	s.Key = key
	s.Scope = scope
	if unit != nil {
		s.Unit = *unit
	}
	if s.Parse == nil {
		return fmt.Errorf("setting parser must not be nil: %s", key)
	}
	return nil
}

func NewSetting[T any](defaultValue T, parse func(string) (T, error)) Setting[T] {
	return Setting[T]{DefaultValue: defaultValue, Parse: parse}
}

func MarshaledText[T encoding.TextUnmarshaler](defaultValue T) Setting[T] {
	return NewSetting(defaultValue, func(s string) (t T, err error) {
		err = t.UnmarshalText([]byte(s))
		return
	})
}

func Bool(defaultValue bool) Setting[bool] {
	return NewSetting(defaultValue, strconv.ParseBool)
}

func String(defaultValue string) Setting[string] {
	return NewSetting(defaultValue, func(s string) (string, error) { return s, nil })
}

func Duration(defaultValue time.Duration) Setting[time.Duration] {
	s := NewSetting(defaultValue, time.ParseDuration)
	s.Unit = "s"
	return s
}

func Time(defaultValue time.Time) Setting[config.Timestamp] {
	s := NewSetting(config.Timestamp(defaultValue.Unix()), config.ParseTimestamp)
	s.Unit = "s"
	return s
}

func TimeRange(lower, upper time.Time) Setting[Range[config.Timestamp]] {
	s := NewSetting(Range[config.Timestamp]{
		Lower: config.Timestamp(lower.Unix()),
		Upper: config.Timestamp(upper.Unix()),
	}, ParseRangeFn(config.ParseTimestamp))
	s.Unit = "s"
	return s
}

func URL(defaultValue *url.URL) Setting[*url.URL] {
	return NewSetting(defaultValue, url.Parse)
}

func Float64(defaultValue float64) Setting[float64] {
	return NewSetting(defaultValue, func(s string) (float64, error) {
		f, err := strconv.ParseFloat(s, 64)
		return float64(f), err
	})
}
func Float32(defaultValue float32) Setting[float32] {
	return NewSetting(defaultValue, func(s string) (float32, error) {
		f, err := strconv.ParseFloat(s, 32)
		return float32(f), err
	})
}

func Rate(defaultLimit rate.Limit, defaultBurst int) Setting[config.Rate] {
	s := NewSetting(config.Rate{Limit: defaultLimit, Burst: defaultBurst}, config.ParseRate)
	s.Unit = "rps"
	return s
}

func Size(defaultValue config.Size) Setting[config.Size] {
	s := NewSetting(defaultValue, config.ParseByte)
	s.Unit = "By"
	return s
}

func Int(defaultValue int) Setting[int] {
	return NewSetting(defaultValue, strconv.Atoi)
}
func Int64(defaultValue int64) Setting[int64] {
	return NewSetting(defaultValue, func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
}
func Int32(defaultValue int32) Setting[int32] {
	return NewSetting(defaultValue, func(s string) (int32, error) {
		i, err := strconv.ParseInt(s, 10, 32)
		return int32(i), err
	})
}
func Int16(defaultValue int16) Setting[int16] {
	return NewSetting(defaultValue, func(s string) (int16, error) {
		i, err := strconv.ParseInt(s, 10, 16)
		return int16(i), err
	})
}
func Int8(defaultValue int8) Setting[int8] {
	return NewSetting(defaultValue, func(s string) (int8, error) {
		i, err := strconv.ParseInt(s, 10, 8)
		return int8(i), err
	})
}
func Uint64(defaultValue uint64) Setting[uint64] {
	return NewSetting(defaultValue, func(s string) (uint64, error) {
		return strconv.ParseUint(s, 10, 64)
	})
}
func Uint32(defaultValue uint32) Setting[uint32] {
	return NewSetting(defaultValue, func(s string) (uint32, error) {
		i, err := strconv.ParseUint(s, 10, 32)
		return uint32(i), err
	})
}
func Uint16(defaultValue uint16) Setting[uint16] {
	return NewSetting(defaultValue, func(s string) (uint16, error) {
		i, err := strconv.ParseUint(s, 10, 16)
		return uint16(i), err
	})
}
func Uint8(defaultValue uint8) Setting[uint8] {
	return NewSetting(defaultValue, func(s string) (uint8, error) {
		i, err := strconv.ParseUint(s, 10, 8)
		return uint8(i), err
	})
}

// GetOrDefault gets the setting from the Getter for the given Scope, or returns the default value with an error.
func (s *Setting[T]) GetOrDefault(ctx context.Context, g Getter) (value T, err error) {
	if g == nil {
		return s.DefaultValue, nil
	}
	str, err := g.GetScoped(ctx, s.Scope, s.Key)
	if err != nil || str == "" {
		return s.DefaultValue, err
	}
	value, err = s.Parse(str)
	if err != nil {
		return s.DefaultValue, err
	}
	return value, nil
}

type Update[T any] struct {
	Value T
	Err   error
}

func (s *Setting[T]) Subscribe(ctx context.Context, r Registry) (<-chan Update[T], func()) {
	updates, stop := r.SubscribeScoped(ctx, s.Scope, s.Key)
	values := make(chan Update[T])
	stopped := make(chan struct{})
	go func() {
		defer close(values)
		for {
			select {
			case <-stopped:
				return
			case update := <-updates:
				if update.Err != nil {
					values <- Update[T]{Value: s.DefaultValue, Err: update.Err}
				} else if v, err := s.Parse(update.Value); err != nil {
					values <- Update[T]{Value: s.DefaultValue, Err: err}
				} else {
					values <- Update[T]{Value: v}
				}
			}
		}
	}()
	return values, sync.OnceFunc(func() {
		stop()
		close(stopped)
	})
}

// InitConfig accepts a pointer to a config struct and iterates over all the fields, initializing any Setting fields
// with a full-qualified, dot-separated key.
// Every field must either be a Setting or a nested struct following the same rules.
func InitConfig(a any) error {
	return initConfig(a, ScopeGlobal, "")
}

func initConfig(a any, scope Scope, parent string) error {
	v := reflect.ValueOf(a)

	if v.Kind() != reflect.Pointer {
		return fmt.Errorf("must be a pointer to a struct: %s: %v", parent, v.Kind())
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("must be a pointer to a struct: %s: %v", parent, v.Kind())
	}
	for i := range v.NumField() {
		f := v.Field(i)
		key := parent
		if key != "" {
			key += "."
		}
		ft := v.Type().Field(i)
		key += ft.Name
		if s := ft.Tag.Get("scope"); s != "" {
			var err error
			scope, err = ParseScope(s)
			if err != nil {
				return fmt.Errorf("%s invalid scope: %s", key, scope)
			}
		}
		var unit *string
		if s, ok := ft.Tag.Lookup("unit"); ok {
			unit = &s
		}
		if s, ok := f.Addr().Interface().(keySetter); ok {
			if err := s.initSetting(key, scope, unit); err != nil {
				return fmt.Errorf("%s failed to init Setting: %w", key, err)
			}
			if ft.Tag.Get("featureflag") == "true" {
				s.setFeatureFlag()
			}
		} else if f.Type().Kind() == reflect.Struct {
			if err := initConfig(f.Addr().Interface(), scope, key); err != nil {
				return fmt.Errorf("%s failed to set keys: %w", key, err)
			}
		} else {
			return fmt.Errorf("%s unsupported field type %v", key, f.Type())
		}
	}
	return nil
}

var _ keySetter = &Setting[struct{}]{}

type keySetter interface {
	initSetting(key string, scope Scope, unit *string) error
	setFeatureFlag()
	info() SettingInfo
}

// SettingInfo describes one Setting in a config struct initialized with InitConfig.
type SettingInfo struct {
	Key         string
	Scope       Scope
	ValueType   reflect.Type // the T in Setting[T]
	FeatureFlag bool
}

// Infos returns every Setting in a config struct previously initialized with InitConfig, in field order.
func Infos(a any) ([]SettingInfo, error) {
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("must be a pointer to a struct: %v", v.Kind())
	}
	var out []SettingInfo
	var walk func(v reflect.Value) error
	walk = func(v reflect.Value) error {
		for i := range v.NumField() {
			f := v.Field(i)
			if s, ok := f.Addr().Interface().(keySetter); ok {
				out = append(out, s.info())
			} else if f.Kind() == reflect.Struct {
				if err := walk(f); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("%s unsupported field type %v", v.Type().Field(i).Name, f.Type())
			}
		}
		return nil
	}
	return out, walk(v.Elem())
}
