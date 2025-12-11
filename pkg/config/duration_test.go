package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuration_MarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input Duration
		want  string
	}{
		{"zero", *MustNewDuration(0), `"0s"`},
		{"one second", *MustNewDuration(time.Second), `"1s"`},
		{"one minute", *MustNewDuration(time.Minute), `"1m0s"`},
		{"one hour", *MustNewDuration(time.Hour), `"1h0m0s"`},
		{"one hour thirty minutes", *MustNewDuration(time.Hour + 30*time.Minute), `"1h30m0s"`},
		{"1 day", *MustNewDuration(24 * time.Hour), `"1d"`},
		{"2 days", *MustNewDuration(48 * time.Hour), `"2d"`},
		{"1 day 12 hours", *MustNewDuration(36 * time.Hour), `"1d12h0m0s"`},
		{"3 days 6 hours 30 minutes", *MustNewDuration(78*time.Hour + 30*time.Minute), `"3d6h30m0s"`},
		{"4 days 1 hour 0 minutes 10 seconds", *MustNewDuration(97*time.Hour + 10*time.Second), `"4d1h0m10s"`},
		{"4 days 0 hours 0 minutes 10 seconds", *MustNewDuration(4*24*time.Hour + 10*time.Second), `"4d0h0m10s"`},
		{"4 days 0 hours 22 minutes 10 seconds", *MustNewDuration(4*24*time.Hour + 22*time.Minute + 10*time.Second), `"4d0h22m10s"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, err := json.Marshal(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.want, string(b))
		})
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Duration
		wantErr bool
	}{
		{"zero", `"0s"`, *MustNewDuration(0), false},
		{"one second", `"1s"`, *MustNewDuration(time.Second), false},
		{"one minute", `"1m0s"`, *MustNewDuration(time.Minute), false},
		{"one hour", `"1h0m0s"`, *MustNewDuration(time.Hour), false},
		{"one hour thirty minutes", `"1h30m0s"`, *MustNewDuration(time.Hour + 30*time.Minute), false},
		{"1 day", `"1d"`, *MustNewDuration(24 * time.Hour), false},
		{"2 days", `"2d"`, *MustNewDuration(48 * time.Hour), false},
		{"1 day 12 hours", `"1d12h0m0s"`, *MustNewDuration(36 * time.Hour), false},
		{"3 days 6 hours 30 minutes", `"3d6h30m0s"`, *MustNewDuration(78*time.Hour + 30*time.Minute), false},
		{"4 days 1 hour 0 minutes 10 seconds", `"4d1h0m10s"`, *MustNewDuration(97*time.Hour + 10*time.Second), false},
		{"4 days 0 hours 0 minutes 10 seconds", `"4d0h0m10s"`, *MustNewDuration(4*24*time.Hour + 10*time.Second), false},
		{"4 days 0 hours 22 minutes 10 seconds", `"4d0h22m10s"`, *MustNewDuration(4*24*time.Hour + 22*time.Minute + 10*time.Second), false},
		{"invalid", `"invalid"`, Duration{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := json.Unmarshal([]byte(tt.input), &d)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, d)
			}
		})
	}
}

func TestDuration_Scan_Value(t *testing.T) {
	t.Parallel()

	d := MustNewDuration(100)
	require.NotNil(t, d)

	val, err := d.Value()
	require.NoError(t, err)

	dNew := MustNewDuration(0)
	err = dNew.Scan(val)
	require.NoError(t, err)

	require.Equal(t, d, dNew)
}

func TestDuration_MarshalJSON_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	d := MustNewDuration(100)
	require.NotNil(t, d)

	json, err := d.MarshalJSON()
	require.NoError(t, err)

	dNew := MustNewDuration(0)
	err = dNew.UnmarshalJSON(json)
	require.NoError(t, err)

	require.Equal(t, d, dNew)
}

func TestDuration_MakeDurationFromString(t *testing.T) {
	t.Parallel()

	d, err := ParseDuration("1s")
	require.NoError(t, err)
	require.Equal(t, 1*time.Second, d.Duration())

	_, err = ParseDuration("xyz")
	require.Error(t, err)
}
