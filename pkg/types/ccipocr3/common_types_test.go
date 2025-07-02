package ccipocr3

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBytes32FromString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected Bytes32
		expErr   bool
	}{
		{
			name:     "valid input",
			input:    "0x200000000000000000000000",
			expected: Bytes32{0x20, 0},
			expErr:   false,
		},
		{
			name:     "invalid hex characters",
			input:    "lrfv",
			expected: Bytes32{},
			expErr:   true,
		},
		{
			name:     "invalid input, no 0x prefix",
			input:    "200000000000000000000000",
			expected: Bytes32{},
			expErr:   true,
		},
		{
			name:     "invalid input, odd len",
			input:    "0x2",
			expected: Bytes32{},
			expErr:   true,
		},
		{
			name:     "valid input, not enough hex chars",
			input:    "0x22",
			expected: Bytes32{0x22},
			expErr:   false,
		},
		{
			name:  "valid input exact length",
			input: "0x" + strings.Repeat("12", 32),
			expected: Bytes32{
				0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12,
				0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12,
				0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12,
				0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12,
			},
			expErr: false,
		},
		{
			name:     "invalid input, tou much hex chars",
			input:    "0x" + strings.Repeat("12", 33),
			expected: Bytes32{},
			expErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewBytes32FromString(tc.input)
			if tc.expErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestBytes32_IsEmpty(t *testing.T) {
	testCases := []struct {
		name     string
		input    Bytes32
		expected bool
	}{
		{
			name:     "empty",
			input:    Bytes32{},
			expected: true,
		},
		{
			name:     "not empty",
			input:    Bytes32{0x20, 0},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.input.IsEmpty())
		})
	}
}

func TestNewBytesFromString(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    Bytes
		wantErr bool
	}{
		{
			"valid input",
			"0x20",
			Bytes{0x20},
			false,
		},
		{
			"valid long input",
			"0x2010201020",
			Bytes{0x20, 0x10, 0x20, 0x10, 0x20},
			false,
		},
		{
			"invalid input",
			"0",
			nil,
			true,
		},
		{
			"invalid input, not enough hex chars",
			"0x2",
			nil,
			true,
		},
		{
			"invalid input, no 0x prefix",
			"20",
			nil,
			true,
		},
		{
			"invalid hex characters",
			"0x2g",
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBytesFromString(tt.arg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})

		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUnknownAddressFromHex(tt.arg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, UnknownAddress(tt.want), got)
			}
		})
	}
}

func TestBytes_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    Bytes
		wantErr bool
	}{
		{
			name:    "valid hex",
			input:   []byte(`"0x20"`),
			want:    Bytes{0x20},
			wantErr: false,
		},
		{
			name:    "valid long hex",
			input:   []byte(`"0x201020"`),
			want:    Bytes{0x20, 0x10, 0x20},
			wantErr: false,
		},
		{
			name:    "empty hex",
			input:   []byte(`"0x"`),
			want:    Bytes{},
			wantErr: false,
		},
		{
			name:    "missing quotes",
			input:   []byte(`0x20`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "no 0x prefix",
			input:   []byte(`"20"`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid hex chars",
			input:   []byte(`"0x2g"`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "too short",
			input:   []byte(`"0"`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   []byte(`""`),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b Bytes
			err := b.UnmarshalJSON(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !assert.ObjectsAreEqual(tc.want, b) {
					t.Errorf("expected %v, got %v", tc.want, b)
				}
			}
		})
	}
}

func TestUnknownAddress_IsZeroOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    UnknownAddress
		expected bool
	}{
		{
			name:     "empty slice",
			input:    UnknownAddress{},
			expected: true,
		},
		{
			name:     "all zero bytes",
			input:    UnknownAddress{0, 0, 0},
			expected: true,
		},
		{
			name:     "non-zero byte at start",
			input:    UnknownAddress{1, 0, 0},
			expected: false,
		},
		{
			name:     "non-zero byte in middle",
			input:    UnknownAddress{0, 2, 0},
			expected: false,
		},
		{
			name:     "non-zero byte at end",
			input:    UnknownAddress{0, 0, 3},
			expected: false,
		},
		{
			name:     "single non-zero byte",
			input:    UnknownAddress{4},
			expected: false,
		},
		{
			name:     "single zero byte",
			input:    UnknownAddress{0},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.input.IsZeroOrEmpty()
			if result != tc.expected {
				t.Errorf("IsZeroOrEmpty() = %v, want %v", result, tc.expected)
			}
		})
	}
}
