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
}
