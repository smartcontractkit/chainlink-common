package ccip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash_String(t *testing.T) {
	tests := []struct {
		name string
		h    Hash
		want string
	}{
		{
			name: "empty",
			h:    Hash{},
			want: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name: "1..",
			h:    Hash{1},
			want: "0x0100000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name: "1..000..1",
			h:    [32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			want: "0x0100000000000000000000000000000000000000000000000000000000000001",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TxMetaFinality(t *testing.T) {
	tests := []struct {
		name                    string
		finalizedAt             uint64
		meta                    TxMeta
		expectedIsFinalized     bool
		expectedFinalizedStatus FinalizedStatus
	}{
		{
			name: "unknown when not specified",
			meta: TxMeta{
				BlockNumber: 1,
			},
			expectedIsFinalized:     false,
			expectedFinalizedStatus: FinalizedStatusUnknown,
		},
		{
			name:        "not finalized",
			finalizedAt: 1,
			meta: TxMeta{
				BlockNumber: 2,
			},
			expectedIsFinalized:     false,
			expectedFinalizedStatus: FinalizedStatusNotFinalized,
		},
		{
			name:        "the same block finalized",
			finalizedAt: 1,
			meta: TxMeta{
				BlockNumber: 1,
			},
			expectedIsFinalized:     true,
			expectedFinalizedStatus: FinalizedStatusFinalized,
		},
		{
			name:        "later block finalized",
			finalizedAt: 10,
			meta: TxMeta{
				BlockNumber: 5,
			},
			expectedIsFinalized:     true,
			expectedFinalizedStatus: FinalizedStatusFinalized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMeta := tt.meta.WithFinalityStatus(tt.finalizedAt)
			assert.Equal(t, tt.expectedIsFinalized, txMeta.IsFinalized())
		})
	}
}
