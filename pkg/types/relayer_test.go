package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const networkEVM = "evm"

func TestIdentifier_UnmarshalString(t *testing.T) {
	type fields struct {
		Network string
		ChainID string
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		want    fields
		args    args
		wantErr bool
	}{
		{name: "evm",
			args:    args{s: "evm.1"},
			wantErr: false,
			want:    fields{Network: networkEVM, ChainID: "1"},
		},
		{name: "bad pattern",
			args:    args{s: "evm_1"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &RelayID{}
			err := i.UnmarshalString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Identifier.UnmarshalString() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want.Network, i.Network)
			assert.Equal(t, tt.want.ChainID, i.ChainID)
		})
	}
}

func TestNewID(t *testing.T) {
	rid := NewRelayID(networkEVM, "chain id")
	assert.Equal(t, networkEVM, rid.Network)
	assert.Equal(t, "chain id", rid.ChainID)
}
