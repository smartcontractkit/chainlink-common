package settings

import (
	"context"
	"errors"
	"fmt"
)

type SettingMap[T any] struct {
	Default Setting[T]
	Values  map[string]string // unparsed
	//TODO better name?
	KeyFromCtx func(context.Context) (string, error) `json:"-" toml:"-"`
}

//TODO doc
func PerChainSelector[T any](defaultValue Setting[T], vals map[string]T) SettingMap[T] {
	svals := make(map[string]string, len(vals))
	for k, v := range vals {
		svals[k] = fmt.Sprint(v)
	}
	return SettingMap[T]{
		Default: defaultValue,
		Values:  svals,
		//TODO move this to contexts.ChainSelectorValue
		KeyFromCtx: func(ctx context.Context) (string, error) {
			val := ctx.Value("TODO key")
			u, ok := val.(string)
			if !ok {
				return "", fmt.Errorf("expected string but got %T", val)
			}
			return u, nil
		},
	}
}

func (s *SettingMap[T]) initSetting(key string, scope Scope, unit *string) error {
	if s.KeyFromCtx == nil {
		return errors.New("missing KeyFromCtx func")
	}
	return s.Default.initSetting(key, scope, unit)
}
