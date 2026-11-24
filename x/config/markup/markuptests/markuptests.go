package markuptests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

func Run(t *testing.T, lang markup.Markup) {
	t.Helper()

	t.Run("roundtrip", func(t *testing.T) {
		data, err := lang.Marshal(struct {
			Size, Duration, Rate, URL, SecretURL, SecretString, Timestamp string
		}{
			Size:         "1.5kb",
			Duration:     "1m30s",
			Rate:         "every2s:3",
			URL:          "https://user:pass@example.com/path?q=1",
			SecretURL:    "https://user:pass@example.com/secret",
			SecretString: "hunter2",
			Timestamp:    "2025-06-15T12:30:45Z",
		})
		require.NoError(t, err)

		var got struct {
			Size         config.Size
			Duration     config.Duration
			Rate         config.Rate
			URL          config.URL
			SecretURL    config.SecretURL
			SecretString config.SecretString
			Timestamp    config.Timestamp
		}
		require.NoError(t, lang.Unmarshal(data, &got), string(data))

		require.Equal(t, 1500*config.Byte, got.Size)
		require.Equal(t, 90*time.Second, got.Duration.Duration())
		require.Equal(t, config.Rate{Limit: rate.Every(2 * time.Second), Burst: 3}, got.Rate)
		require.Equal(t, "https://user:pass@example.com/path?q=1", got.URL.URL().String())
		require.Equal(t, "https://user:pass@example.com/secret", got.SecretURL.URL().String())
		require.Equal(t, config.SecretString("hunter2"), got.SecretString)
		require.Equal(t, config.NewTimestamp(time.Date(2025, 6, 15, 12, 30, 45, 0, time.UTC)), got.Timestamp)
	})

	t.Run("discard", func(t *testing.T) {
		type configWithDiscard struct {
			Deprecated markup.Discard
			Keep       string
		}

		// All Deprecated field values contain the word "old" to ensure it's not serialized
		for _, tc := range []struct {
			name  string
			value any
		}{
			{"leaf", "old"},
			{"array", []any{map[string]any{"Name": "old"}, map[string]any{"Name": "b"}}},
			{"object", map[string]any{"URL": "https://user:pass@example.com/old/secret", "Secret": "old"}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				data, err := lang.Marshal(map[string]any{"Deprecated": tc.value, "Keep": "kept"})
				require.NoError(t, err)

				got := configWithDiscard{Keep: "default"}
				require.NoError(t, markup.Decode(lang, data, &got), string(data))
				require.Equal(t, configWithDiscard{Keep: "kept"}, got)
				r, err := lang.Marshal(got)
				require.NoError(t, err)
				require.NotContains(t, string(r), "old")
			})
		}
	})
}
