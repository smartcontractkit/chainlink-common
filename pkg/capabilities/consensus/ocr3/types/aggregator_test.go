package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetadata_EncodeDecode(t *testing.T) {
	metadata := Metadata{
		Version:          1,
		ExecutionID:      "1234567890123456789012345678901234567890123456789012345678901234",
		Timestamp:        1620000000,
		DONID:            1,
		DONConfigVersion: 1,
		WorkflowID:       "1234567890123456789012345678901234567890123456789012345678901234",
		WorkflowName:     "12",
		WorkflowOwner:    "1234567890123456789012345678901234567890",
		ReportID:         "1234",
	}

	metadata.padWorkflowName()

	encoded, err := metadata.Encode()
	require.NoError(t, err)

	require.Len(t, encoded, 109)

	// append tail to encoded
	tail := []byte("tail")
	encoded = append(encoded, tail...)
	decoded, remaining, err := Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, metadata.Version, decoded.Version)
	require.Equal(t, metadata.ExecutionID, decoded.ExecutionID)
	require.Equal(t, metadata.Timestamp, decoded.Timestamp)
	require.Equal(t, metadata.DONID, decoded.DONID)
	require.Equal(t, metadata.DONConfigVersion, decoded.DONConfigVersion)
	require.Equal(t, metadata.WorkflowID, decoded.WorkflowID)
	require.Equal(t, metadata.WorkflowName, decoded.WorkflowName)
	require.Equal(t, metadata.WorkflowOwner, decoded.WorkflowOwner)
	require.Equal(t, metadata.ReportID, decoded.ReportID)
	require.Equal(t, tail, remaining)
}

func TestMetadata_Length(t *testing.T) {
	var m Metadata
	require.Equal(t, MetadataLen, m.Length())
}

func TestPadWorkflowName_NoPadWhenExactLength(t *testing.T) {
	// 20 hex characters = 10 bytes, exact length
	original := "abcdef0123456789abcd"
	m := &Metadata{WorkflowName: original}
	m.padWorkflowName()
	require.Equal(t, original, m.WorkflowName)
}

func TestPadWorkflowName_TooLong(t *testing.T) {
	// 22 hex characters = 11 bytes, should not be truncated by pad
	original := "abcdef0123456789abcd01"
	m := &Metadata{WorkflowName: original}
	m.padWorkflowName()
	require.Equal(t, original, m.WorkflowName)
}

func TestEncode_InvalidHexFields(t *testing.T) {
	m := Metadata{
		Version:          1,
		ExecutionID:      "zzzz", // invalid hex
		Timestamp:        0,
		DONID:            0,
		DONConfigVersion: 0,
		WorkflowID:       strings.Repeat("00", 32),
		WorkflowName:     "00",
		WorkflowOwner:    strings.Repeat("00", 20),
		ReportID:         "0000",
	}
	_, err := m.Encode()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid hex")
}

func TestEncode_WrongLengthFields(t *testing.T) {
	tests := []struct {
		name string
		m    Metadata
	}{
		{
			name: "short ExecutionID",
			m: Metadata{
				Version:          1,
				ExecutionID:      "00", // too short
				Timestamp:        0,
				DONID:            0,
				DONConfigVersion: 0,
				WorkflowID:       strings.Repeat("00", 32),
				WorkflowName:     "00",
				WorkflowOwner:    strings.Repeat("00", 20),
				ReportID:         "0000",
			},
		},
		{
			name: "short WorkflowID",
			m: Metadata{
				Version:          1,
				ExecutionID:      strings.Repeat("00", 32),
				Timestamp:        0,
				DONID:            0,
				DONConfigVersion: 0,
				WorkflowID:       "00", // too short
				WorkflowName:     "00",
				WorkflowOwner:    strings.Repeat("00", 20),
				ReportID:         "0000",
			},
		},
		{
			name: "long WorkflowName",
			m: Metadata{
				Version:          1,
				ExecutionID:      strings.Repeat("00", 32),
				Timestamp:        0,
				DONID:            0,
				DONConfigVersion: 0,
				WorkflowID:       strings.Repeat("00", 32),
				WorkflowName:     strings.Repeat("01", 11), // 22 chars, >20
				WorkflowOwner:    strings.Repeat("00", 20),
				ReportID:         "0000",
			},
		},
		{
			name: "short WorkflowOwner",
			m: Metadata{
				Version:          1,
				ExecutionID:      strings.Repeat("00", 32),
				Timestamp:        0,
				DONID:            0,
				DONConfigVersion: 0,
				WorkflowID:       strings.Repeat("00", 32),
				WorkflowName:     "00",
				WorkflowOwner:    "00", // too short
				ReportID:         "0000",
			},
		},
		{
			name: "short ReportID",
			m: Metadata{
				Version:          1,
				ExecutionID:      strings.Repeat("00", 32),
				Timestamp:        0,
				DONID:            0,
				DONConfigVersion: 0,
				WorkflowID:       strings.Repeat("00", 32),
				WorkflowName:     "00",
				WorkflowOwner:    strings.Repeat("00", 20),
				ReportID:         "00", // too short
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.m.Encode()
			require.Error(t, err)
			require.Contains(t, err.Error(), "wrong length")
		})
	}
}

func TestDecode_RawTooShort(t *testing.T) {
	_, _, err := Decode([]byte{0x01, 0x02})
	require.Error(t, err)
	require.Contains(t, err.Error(), "raw too short")
}

func TestDecode_RemainingData(t *testing.T) {
	m := Metadata{
		Version:          1,
		ExecutionID:      strings.Repeat("11", 32),
		Timestamp:        2,
		DONID:            3,
		DONConfigVersion: 4,
		WorkflowID:       strings.Repeat("22", 32),
		WorkflowName:     "33",
		WorkflowOwner:    strings.Repeat("44", 20),
		ReportID:         "5555",
	}
	m.padWorkflowName()

	encoded, err := m.Encode()
	require.NoError(t, err)
	// add extra bytes to simulate payload
	extra := []byte("extra")
	data := append(encoded, extra...)

	decoded, remaining, err := Decode(data)
	require.NoError(t, err)
	require.Equal(t, extra, remaining)
	require.Equal(t, m, decoded)
}

func TestMetadata_padWorkflowName(t *testing.T) {
	type fields struct {
		WorkflowName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "padWorkflowName hex with 9 bytes",
			fields: fields{
				WorkflowName: "ABCD1234EF567890AB",
			},
			want: "ABCD1234EF567890AB00",
		},
		{
			name: "padWorkflowName hex with 5 bytes",
			fields: fields{
				WorkflowName: "1234ABCD56",
			},
			want: "1234ABCD560000000000",
		},
		{
			name: "padWorkflowName empty",
			fields: fields{
				WorkflowName: "",
			},
			want: "00000000000000000000",
		},
		{
			name: "padWorkflowName non-hex string",
			fields: fields{
				WorkflowName: "not-hex",
			},
			want: "not-hex0000000000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metadata{
				WorkflowName: tt.fields.WorkflowName,
			}
			m.padWorkflowName()
			require.Equal(t, tt.want, m.WorkflowName, tt.name)
		})
	}
}
