package api

import (
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/stretchr/testify/require"
)

func TestObjectMatchersEqual(t *testing.T) {
	t.Run("returns true if the two object matchers are equal", func(t *testing.T) {
		a := alerting.ObjectMatchers{{"team", "=", "chainlink"}}
		b := alerting.ObjectMatchers{{"team", "=", "chainlink"}}

		result := objectMatchersEqual(a, b)
		require.True(t, result)
	})

	t.Run("returns true if the two object matchers with multiple matches are equal", func(t *testing.T) {
		a := alerting.ObjectMatchers{
			{"team", "=", "chainlink"},
			{"severity", "=", "critical"},
		}
		b := alerting.ObjectMatchers{
			{"severity", "=", "critical"},
			{"team", "=", "chainlink"},
		}

		result := objectMatchersEqual(a, b)
		require.True(t, result)
	})

	t.Run("returns false if the two object matchers with multiple matches are different", func(t *testing.T) {
		a := alerting.ObjectMatchers{
			{"team", "=", "chainlink"},
			{"severity", "=", "critical"},
		}
		b := alerting.ObjectMatchers{
			{"severity", "=", "warning"},
			{"team", "=", "chainlink"},
		}

		result := objectMatchersEqual(a, b)
		require.False(t, result)
	})
}
