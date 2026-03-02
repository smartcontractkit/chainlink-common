package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTimestamp(t *testing.T) {
	ref := time.Date(2025, 6, 15, 12, 30, 45, 0, time.UTC)

	tests := []struct {
		name  string
		input string
		want  Timestamp
	}{
		{
			name:  "RFC3339",
			input: "2025-06-15T12:30:45Z",
			want:  Timestamp(ref.Unix()),
		},
		{
			name:  "Go default without nanoseconds",
			input: "2025-06-15 12:30:45 +0000 UTC",
			want:  Timestamp(ref.Unix()),
		},
		{
			name:  "Go default - nanoseconds truncated",
			input: "2025-06-15 12:30:45.123456789 +0000 UTC",
			want:  Timestamp(ref.Unix()),
		},
		{
			name:  "Unix integer",
			input: "1749990645",
			want:  Timestamp(ref.Unix()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimestamp(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
