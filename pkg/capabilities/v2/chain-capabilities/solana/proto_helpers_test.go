package solana_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	solcap "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/solana"
	typesolana "github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
)

// pk32 builds a 32-byte Solana pubkey with every byte set to b.
func pk32(b byte) []byte {
	out := make([]byte, 32)
	for i := range out {
		out[i] = b
	}
	return out
}

// domainPK builds a typesolana.PublicKey with every byte set to b.
func domainPK(b byte) typesolana.PublicKey {
	var pk typesolana.PublicKey
	for i := range pk {
		pk[i] = b
	}
	return pk
}

// domainPKBytes returns domainPK(b) as a []byte slice.
func domainPKBytes(b byte) []byte {
	pk := domainPK(b)
	return pk[:]
}

// ---- ConvertGetProgramAccountsRequestFromProto ----

func TestConvertGetProgramAccountsRequestFromProto_Nil(t *testing.T) {
	got, err := solcap.ConvertGetProgramAccountsRequestFromProto(nil)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestConvertGetProgramAccountsRequestFromProto_NoOpts(t *testing.T) {
	req := &solcap.GetProgramAccountsRequest{
		Program: pk32(0xAB),
	}
	got, err := solcap.ConvertGetProgramAccountsRequestFromProto(req)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, domainPK(0xAB), got.Program)
	assert.Nil(t, got.Opts)
}

func TestConvertGetProgramAccountsRequestFromProto_WithOpts(t *testing.T) {
	offset := uint64(8)
	length := uint64(16)
	req := &solcap.GetProgramAccountsRequest{
		Program: pk32(0x01),
		Opts: &solcap.GetProgramAccountsOpts{
			Encoding:   solcap.EncodingType_ENCODING_TYPE_BASE64,
			Commitment: solcap.CommitmentType_COMMITMENT_TYPE_CONFIRMED,
			DataSlice:  &solcap.DataSlice{Offset: offset, Length: length},
			Filters: []*solcap.RPCFilter{
				{DataSize: 165},
				{Memcmp: &solcap.RPCFilterMemcmp{Offset: 0, Bytes: pk32(0x02)}},
			},
		},
	}
	got, err := solcap.ConvertGetProgramAccountsRequestFromProto(req)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.Opts)

	assert.Equal(t, typesolana.EncodingBase64, got.Opts.Encoding)
	assert.Equal(t, typesolana.CommitmentConfirmed, got.Opts.Commitment)
	require.NotNil(t, got.Opts.DataSlice)
	assert.Equal(t, &offset, got.Opts.DataSlice.Offset)
	assert.Equal(t, &length, got.Opts.DataSlice.Length)

	require.Len(t, got.Opts.Filters, 2)
	assert.Equal(t, uint64(165), got.Opts.Filters[0].DataSize)
	assert.Nil(t, got.Opts.Filters[0].Memcmp)

	require.NotNil(t, got.Opts.Filters[1].Memcmp)
	assert.Equal(t, uint64(0), got.Opts.Filters[1].Memcmp.Offset)
	assert.Equal(t, pk32(0x02), got.Opts.Filters[1].Memcmp.Bytes)
}

func TestConvertGetProgramAccountsRequestFromProto_NilFilters(t *testing.T) {
	req := &solcap.GetProgramAccountsRequest{
		Program: pk32(0x01),
		Opts:    &solcap.GetProgramAccountsOpts{Filters: nil},
	}
	got, err := solcap.ConvertGetProgramAccountsRequestFromProto(req)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.Opts.Filters)
}

func TestConvertGetProgramAccountsRequestFromProto_NilFilterEntry(t *testing.T) {
	req := &solcap.GetProgramAccountsRequest{
		Program: pk32(0x01),
		Opts: &solcap.GetProgramAccountsOpts{
			Filters: []*solcap.RPCFilter{nil, {DataSize: 64}},
		},
	}
	_, err := solcap.ConvertGetProgramAccountsRequestFromProto(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "filters[0]")
}

func TestConvertGetProgramAccountsRequestFromProto_InvalidProgram(t *testing.T) {
	req := &solcap.GetProgramAccountsRequest{
		Program: []byte{0x01, 0x02}, // wrong length — must be 32 bytes
	}
	_, err := solcap.ConvertGetProgramAccountsRequestFromProto(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "program")
}

func TestConvertGetProgramAccountsRequestFromProto_InvalidEncoding(t *testing.T) {
	req := &solcap.GetProgramAccountsRequest{
		Program: pk32(0x01),
		Opts:    &solcap.GetProgramAccountsOpts{Encoding: solcap.EncodingType(999)},
	}
	_, err := solcap.ConvertGetProgramAccountsRequestFromProto(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "encoding")
}

func TestConvertGetProgramAccountsRequestFromProto_InvalidCommitment(t *testing.T) {
	req := &solcap.GetProgramAccountsRequest{
		Program: pk32(0x01),
		Opts:    &solcap.GetProgramAccountsOpts{Commitment: solcap.CommitmentType(999)},
	}
	_, err := solcap.ConvertGetProgramAccountsRequestFromProto(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "commitment")
}

// ---- ConvertGetProgramAccountsReplyToProto ----

func TestConvertGetProgramAccountsReplyToProto_Nil(t *testing.T) {
	got, err := solcap.ConvertGetProgramAccountsReplyToProto(nil)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestConvertGetProgramAccountsReplyToProto_Empty(t *testing.T) {
	got, err := solcap.ConvertGetProgramAccountsReplyToProto(&typesolana.GetProgramAccountsReply{})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Empty(t, got.Value)
}

func TestConvertGetProgramAccountsReplyToProto_NilEntry(t *testing.T) {
	reply := &typesolana.GetProgramAccountsReply{
		Value: []*typesolana.KeyedAccount{nil, {Pubkey: domainPK(0x10)}},
	}
	got, err := solcap.ConvertGetProgramAccountsReplyToProto(reply)
	require.NoError(t, err)
	// nil entry is skipped
	require.Len(t, got.Value, 1)
	assert.Equal(t, domainPKBytes(0x10), got.Value[0].Pubkey)
}

func TestConvertGetProgramAccountsReplyToProto_WithAccounts(t *testing.T) {
	data := &typesolana.DataBytesOrJSON{
		RawDataEncoding: typesolana.EncodingBase64,
		AsDecodedBinary: []byte{0xCA, 0xFE},
	}
	reply := &typesolana.GetProgramAccountsReply{
		Value: []*typesolana.KeyedAccount{
			{
				Pubkey: domainPK(0x11),
				Account: &typesolana.Account{
					Lamports:   1000,
					Owner:      domainPK(0x22),
					Data:       data,
					Executable: true,
					RentEpoch:  big.NewInt(42),
					Space:      128,
				},
			},
			{
				Pubkey:  domainPK(0x33),
				Account: nil, // account may be absent
			},
		},
	}
	got, err := solcap.ConvertGetProgramAccountsReplyToProto(reply)
	require.NoError(t, err)
	require.Len(t, got.Value, 2)

	// first entry — full account
	ka0 := got.Value[0]
	assert.Equal(t, domainPKBytes(0x11), ka0.Pubkey)
	require.NotNil(t, ka0.Account)
	assert.Equal(t, uint64(1000), ka0.Account.Lamports)
	assert.Equal(t, domainPKBytes(0x22), ka0.Account.Owner)
	assert.True(t, ka0.Account.Executable)
	assert.Equal(t, uint64(128), ka0.Account.Space)

	// second entry — nil account is valid
	ka1 := got.Value[1]
	assert.Equal(t, domainPKBytes(0x33), ka1.Pubkey)
	assert.Nil(t, ka1.Account)
}

func TestConvertGetProgramAccountsReplyToProto_NilAccountData(t *testing.T) {
	reply := &typesolana.GetProgramAccountsReply{
		Value: []*typesolana.KeyedAccount{
			{
				Pubkey:  domainPK(0x55),
				Account: &typesolana.Account{Lamports: 500, Data: nil},
			},
		},
	}
	got, err := solcap.ConvertGetProgramAccountsReplyToProto(reply)
	require.NoError(t, err)
	require.Len(t, got.Value, 1)
	assert.Nil(t, got.Value[0].Account.Data)
}

func TestConvertGetProgramAccountsReplyToProto_InvalidAccountEncoding(t *testing.T) {
	reply := &typesolana.GetProgramAccountsReply{
		Value: []*typesolana.KeyedAccount{
			{
				Pubkey: domainPK(0x01),
				Account: &typesolana.Account{
					Data: &typesolana.DataBytesOrJSON{
						RawDataEncoding: typesolana.EncodingType("invalid-encoding"),
					},
				},
			},
		},
	}
	_, err := solcap.ConvertGetProgramAccountsReplyToProto(reply)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value[0]")
}
