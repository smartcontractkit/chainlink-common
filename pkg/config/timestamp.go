package config

import (
	"fmt"
	"strconv"
	"time"
)

// Timestamp represents a Unix timestamp (seconds) that marshals to/from RFC3339 strings.
type Timestamp int64

func (t Timestamp) String() string {
	return time.Unix(int64(t), 0).UTC().Format("2006-01-02 15:04:05 -0700 MST")
}

func (t Timestamp) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *Timestamp) UnmarshalText(b []byte) error {
	v, err := ParseTimestamp(string(b))
	if err != nil {
		return err
	}
	*t = v
	return nil
}

func ParseTimestamp(s string) (Timestamp, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return Timestamp(t.Unix()), nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", s); err == nil {
		return Timestamp(t.Unix()), nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse timestamp %q: %w", s, err)
	}
	return Timestamp(v), nil
}
