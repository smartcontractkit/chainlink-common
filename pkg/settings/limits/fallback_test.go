package limits

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func TestMakeUpperBoundLimiter_WithFallback(t *testing.T) {
	t.Parallel()

	var s struct {
		Shared   settings.Setting[int]
		Specific settings.Setting[int]
	}
	s.Shared = settings.Int(10)
	s.Specific = settings.Int(20)
	require.NoError(t, settings.InitConfig(&s))

	for _, tt := range []struct {
		name    string
		json    string
		allowed int
		denied  int
	}{
		{name: "defaults to the shared default", json: `{}`, allowed: 10, denied: 11},
		{name: "unset specific uses the shared override", json: `{"global":{"Shared":"5"}}`, allowed: 5, denied: 6},
		{name: "specific override wins", json: `{"global":{"Shared":"5","Specific":"8"}}`, allowed: 8, denied: 9},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g, err := settings.NewJSONGetter([]byte(tt.json))
			require.NoError(t, err)

			bl, err := MakeUpperBoundLimiter(Factory{Settings: g}, settings.WithFallback(s.Specific, s.Shared))
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, bl.Close()) })

			assert.NoError(t, bl.Check(t.Context(), tt.allowed))
			assert.Error(t, bl.Check(t.Context(), tt.denied))
		})
	}
}
