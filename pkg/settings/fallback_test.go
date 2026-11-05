package settings

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type getterFunc func(ctx context.Context, scope Scope, key string) (string, error)

func (f getterFunc) GetScoped(ctx context.Context, scope Scope, key string) (string, error) {
	return f(ctx, scope, key)
}

func TestSettingWithFallback_GetOrDefault(t *testing.T) {
	var s struct {
		Shared   Setting[int]
		Specific Setting[int]
		Narrow   Setting[int]
	}
	s.Shared = Int(1)
	s.Specific = Int(2)
	s.Narrow = Int(3)
	require.NoError(t, InitConfig(&s))

	getter := func(t *testing.T, json string) Getter {
		g, err := NewJSONGetter([]byte(json))
		require.NoError(t, err)
		return g
	}

	for _, tt := range []struct {
		name    string
		setting SettingWithFallback[int]
		getter  func(t *testing.T) Getter
		want    int
	}{
		{
			name:    "specific unset falls back to the shared value",
			setting: WithFallback(s.Specific, s.Shared),
			getter:  func(t *testing.T) Getter { return getter(t, `{"global":{"Shared":"4"}}`) },
			want:    4,
		},
		{
			name:    "specific set wins over the shared value",
			setting: WithFallback(s.Specific, s.Shared),
			getter:  func(t *testing.T) Getter { return getter(t, `{"global":{"Shared":"4","Specific":"7"}}`) },
			want:    7,
		},
		{
			name:    "neither set returns the shared default",
			setting: WithFallback(s.Specific, s.Shared),
			getter:  func(t *testing.T) Getter { return getter(t, `{}`) },
			want:    1,
		},
		{
			name:    "nil getter returns the shared default",
			setting: WithFallback(s.Specific, s.Shared),
			getter:  func(t *testing.T) Getter { return nil },
			want:    1,
		},
		{
			name:    "first set of several wins",
			setting: WithFallback(s.Narrow, s.Specific, s.Shared),
			getter:  func(t *testing.T) Getter { return getter(t, `{"global":{"Shared":"4","Specific":"7"}}`) },
			want:    7,
		},
		{
			name:    "most specific of several wins",
			setting: WithFallback(s.Narrow, s.Specific, s.Shared),
			getter:  func(t *testing.T) Getter { return getter(t, `{"global":{"Shared":"4","Specific":"7","Narrow":"9"}}`) },
			want:    9,
		},
		{
			name:    "zero setting is skipped",
			setting: WithFallback(Setting[int]{}, s.Shared),
			getter:  func(t *testing.T) Getter { return getter(t, `{"global":{"Shared":"4","Specific":"7"}}`) },
			want:    4,
		},
		{
			name:    "all zero settings",
			setting: WithFallback(Setting[int]{}),
			getter:  func(t *testing.T) Getter { return getter(t, `{}`) },
			want:    0,
		},
		{
			name:    "empty",
			setting: SettingWithFallback[int]{},
			getter:  func(t *testing.T) Getter { return getter(t, `{}`) },
			want:    0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.setting.GetOrDefault(t.Context(), tt.getter(t))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("getter error returns the shared default", func(t *testing.T) {
		wantErr := errors.New("lookup failed")
		g := getterFunc(func(context.Context, Scope, string) (string, error) { return "", wantErr })
		got, err := WithFallback(s.Specific, s.Shared).GetOrDefault(t.Context(), g)
		require.ErrorIs(t, err, wantErr)
		assert.Equal(t, 1, got)
	})

	t.Run("parse error returns the shared default", func(t *testing.T) {
		got, err := WithFallback(s.Specific, s.Shared).GetOrDefault(t.Context(), getter(t, `{"global":{"Specific":"nope"}}`))
		require.Error(t, err)
		assert.Equal(t, 1, got)
	})
}

func TestSettingWithFallback_Spec(t *testing.T) {
	var s struct {
		Shared   Setting[time.Duration] `scope:"owner"`
		Specific Setting[time.Duration]
	}
	s.Shared = Duration(time.Second)
	s.Specific = Duration(time.Minute)
	require.NoError(t, InitConfig(&s))

	spec := WithFallback(Setting[time.Duration]{}, s.Specific, s.Shared).GetSpec()
	assert.Equal(t, "Specific", spec.GetKey())
	assert.Equal(t, s.Specific.Scope, spec.GetScope())
	assert.Equal(t, "s", spec.GetUnit())
}

type fakeRegistry struct {
	Getter

	mu   sync.Mutex
	subs map[string][]chan Update[string]
}

func (r *fakeRegistry) SubscribeScoped(_ context.Context, _ Scope, key string) (<-chan Update[string], func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ch := make(chan Update[string], 1)
	if r.subs == nil {
		r.subs = map[string][]chan Update[string]{}
	}
	r.subs[key] = append(r.subs[key], ch)
	return ch, func() {}
}

func (r *fakeRegistry) notify(key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ch := range r.subs[key] {
		ch <- Update[string]{Value: value}
	}
}

func TestSettingWithFallback_Subscribe(t *testing.T) {
	var s struct {
		Shared   Setting[int]
		Specific Setting[int]
	}
	s.Shared = Int(1)
	s.Specific = Int(2)
	require.NoError(t, InitConfig(&s))

	var mu sync.Mutex
	values := map[string]string{}
	set := func(key, value string) {
		mu.Lock()
		defer mu.Unlock()
		values[key] = value
	}
	r := &fakeRegistry{Getter: getterFunc(func(_ context.Context, _ Scope, key string) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		return values[key], nil
	})}

	updates, stop := WithFallback(s.Specific, s.Shared).Subscribe(t.Context(), r)
	t.Cleanup(stop)

	next := func() Update[int] {
		select {
		case u := <-updates:
			return u
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for update")
			return Update[int]{}
		}
	}

	set("Shared", "5")
	r.notify("Shared", "5")
	assert.Equal(t, Update[int]{Value: 5}, next(), "a shared change reaches an unset specific setting")

	set("Specific", "8")
	r.notify("Specific", "8")
	assert.Equal(t, Update[int]{Value: 8}, next(), "a specific change takes precedence")

	set("Shared", "6")
	r.notify("Shared", "6")
	assert.Equal(t, Update[int]{Value: 8}, next(), "a shared change does not move a set specific setting")

	set("Specific", "")
	r.notify("Specific", "")
	assert.Equal(t, Update[int]{Value: 6}, next(), "unsetting the specific setting falls back again")
}
