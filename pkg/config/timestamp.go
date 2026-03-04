package config

import (
	"fmt"
	"strconv"
	"time"
)

// Timestamp represents a Unix timestamp (seconds) that marshals to/from a string.
// Supported input formats:
// - RFC3339 (e.g. "2025-06-15T12:30:45Z")
// - Go default, nanoseconds truncated (e.g. "2006-01-02 15:04:05 -0700 MST" or "2006-01-02 15:04:05.123456789 -0700 MST")
// - Integer Unix timestamp (seconds)
type Timestamp int64

const defaultFormat = "2006-01-02 15:04:05 -0700 MST"

func (t Timestamp) String() string {
	return time.Unix(int64(t), 0).UTC().Format(defaultFormat)
}

func (t Timestamp) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *Timestamp) UnmarshalText(b []byte) error {
	s := string(b)
	if parsed, err := time.Parse(defaultFormat, s); err == nil {
		*t = Timestamp(parsed.Unix())
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339, s); err == nil {
		*t = Timestamp(parsed.Unix())
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp %q: %w", s, err)
	}
	*t = Timestamp(v)
	return nil
}

func ParseTimestamp(s string) (Timestamp, error) {
	var t Timestamp
	err := t.UnmarshalText([]byte(s))
	return t, err
}
