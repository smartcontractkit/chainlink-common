package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytes_MarshalText_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    Size
		expected string
	}{
		{Size(0), "0b"},
		{Size(1), "1b"},
		{MByte, "1mb"},
		{MByte + 100*KByte, "1.1mb"},
		{KByte, "1kb"},
		{999 * KByte, "999kb"},
		{GByte, "1gb"},
		{TByte, "1tb"},
		{5 * GByte, "5gb"},
		{500 * MByte, "500mb"},
		{999 * MByte, "999mb"},
		{999*MByte + 999*KByte, "999.999mb"},
		{999*MByte + 999*KByte + 999, "999.999999mb"},
	}

	for _, test := range tests {
		test := test

		t.Run(test.expected, func(t *testing.T) {
			t.Parallel()

			bstr, err := test.input.MarshalText()
			assert.NoError(t, err)
			assert.Equal(t, test.expected, string(bstr))
			assert.Equal(t, test.expected, test.input.String())
		})
	}
}

func TestBytes_UnmarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected Size
		valid    bool
	}{
		// valid
		{"0", Size(0), true},
		{"0.0", Size(0), true},
		{"123", Size(123), true},
		{"123", Size(123), true},
		{"123b", Size(123), true},
		{"123B", Size(123), true},
		{"123kb", 123 * KByte, true},
		{"123KB", 123 * KByte, true},
		{"123mb", 123 * MByte, true},
		{"123gb", 123 * GByte, true},
		{"123tb", 123 * TByte, true},
		{"5.5mb", 5500 * KByte, true},
		{"0.5mb", 500 * KByte, true},
		// invalid
		{"1.12345", Size(1), false},
		{"", Size(0), false},
		{"xyz", Size(0), false},
		{"-1g", Size(0), false},
		{"+1g", Size(0), false},
		{"1g", Size(0), false},
		{"1t", Size(0), false},
		{"1a", Size(0), false},
		{"1tbtb", Size(0), false},
		{"1tb1tb", Size(0), false},
	}

	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			t.Parallel()

			var fs Size
			err := fs.UnmarshalText([]byte(test.input))
			if test.valid {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, fs)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
