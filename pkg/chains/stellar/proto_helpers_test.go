package stellar_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"
	"github.com/stretchr/testify/require"

	conv "github.com/smartcontractkit/chainlink-common/pkg/chains/stellar"
	stellartypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

func TestConvertGetLedgerEntriesRequest_RoundTrip(t *testing.T) {
	key1 := base64.StdEncoding.EncodeToString([]byte("key-one"))
	key2 := base64.StdEncoding.EncodeToString([]byte("key-two"))
	domain := stellartypes.GetLedgerEntriesRequest{Keys: []string{key1, key2}}

	proto, err := conv.ConvertGetLedgerEntriesRequestToProto(domain)
	require.NoError(t, err)
	require.Len(t, proto.GetKeys(), 2)

	got, err := conv.ConvertGetLedgerEntriesRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetLedgerEntriesRequestToProto_InvalidBase64(t *testing.T) {
	domain := stellartypes.GetLedgerEntriesRequest{Keys: []string{"not-valid-base64!!"}}
	_, err := conv.ConvertGetLedgerEntriesRequestToProto(domain)
	require.Error(t, err)
	require.Contains(t, err.Error(), "key[0]")
}

func TestConvertGetLedgerEntriesRequestFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetLedgerEntriesRequestFromProto(nil)
	require.EqualError(t, err, "get ledger entries request is nil")
}

func TestConvertGetLedgerEntriesRequestFromProto_EmptyKeys(t *testing.T) {
	_, err := conv.ConvertGetLedgerEntriesRequestFromProto(&conv.GetLedgerEntriesRequest{})
	require.EqualError(t, err, "ledger entry keys are empty")
}

func TestConvertLedgerEntryResult_RoundTrip(t *testing.T) {
	liveUntil := uint32(9999)
	domain := stellartypes.LedgerEntryResult{
		KeyXDR:             base64.StdEncoding.EncodeToString([]byte("key-xdr")),
		DataXDR:            base64.StdEncoding.EncodeToString([]byte("data-xdr")),
		LastModifiedLedger: 42,
		LiveUntilLedgerSeq: &liveUntil,
		ExtensionXDR:       base64.StdEncoding.EncodeToString([]byte("ext-xdr")),
	}

	proto, err := conv.ConvertLedgerEntryResultToProto(domain)
	require.NoError(t, err)
	require.True(t, proto.GetHasLiveUntilLedgerSeq())
	require.Equal(t, uint32(9999), proto.GetLiveUntilLedgerSeq())

	got, err := conv.ConvertLedgerEntryResultFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertLedgerEntryResult_NoLiveUntil(t *testing.T) {
	domain := stellartypes.LedgerEntryResult{
		KeyXDR:             base64.StdEncoding.EncodeToString([]byte("k")),
		DataXDR:            base64.StdEncoding.EncodeToString([]byte("d")),
		LastModifiedLedger: 1,
		ExtensionXDR:       base64.StdEncoding.EncodeToString([]byte("e")),
	}

	proto, err := conv.ConvertLedgerEntryResultToProto(domain)
	require.NoError(t, err)
	require.False(t, proto.GetHasLiveUntilLedgerSeq())

	got, err := conv.ConvertLedgerEntryResultFromProto(proto)
	require.NoError(t, err)
	require.Nil(t, got.LiveUntilLedgerSeq)
	require.Equal(t, domain, got)
}

func TestConvertLedgerEntryResultFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertLedgerEntryResultFromProto(nil)
	require.EqualError(t, err, "ledger entry result is nil")
}

func TestConvertGetLedgerEntriesResponse_RoundTrip(t *testing.T) {
	liveUntil := uint32(500)
	domain := stellartypes.GetLedgerEntriesResponse{
		Entries: []stellartypes.LedgerEntryResult{
			{KeyXDR: base64.StdEncoding.EncodeToString([]byte("k1")), DataXDR: base64.StdEncoding.EncodeToString([]byte("d1")), LastModifiedLedger: 10, ExtensionXDR: base64.StdEncoding.EncodeToString(nil)},
			{KeyXDR: base64.StdEncoding.EncodeToString([]byte("k2")), DataXDR: base64.StdEncoding.EncodeToString([]byte("d2")), LastModifiedLedger: 20, LiveUntilLedgerSeq: &liveUntil, ExtensionXDR: base64.StdEncoding.EncodeToString(nil)},
		},
		LatestLedger: 999,
	}

	proto, err := conv.ConvertGetLedgerEntriesResponseToProto(domain)
	require.NoError(t, err)
	require.Len(t, proto.GetEntries(), 2)

	got, err := conv.ConvertGetLedgerEntriesResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetLedgerEntriesResponseFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetLedgerEntriesResponseFromProto(nil)
	require.EqualError(t, err, "get ledger entries response is nil")
}

func TestConvertGetLatestLedgerResponse_RoundTrip(t *testing.T) {
	var b [32]byte
	for i := range b {
		b[i] = 0xAB
	}
	hash := b
	domain := stellartypes.GetLatestLedgerResponse{
		Hash:              hex.EncodeToString(hash[:]),
		ProtocolVersion:   21,
		Sequence:          1234567,
		LedgerCloseTime:   1_700_000_000,
		LedgerHeaderXDR:   base64.StdEncoding.EncodeToString([]byte("header-xdr")),
		LedgerMetadataXDR: base64.StdEncoding.EncodeToString([]byte("meta-xdr")),
	}

	proto, err := conv.ConvertGetLatestLedgerResponseToProto(domain)
	require.NoError(t, err)

	got, err := conv.ConvertGetLatestLedgerResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetLatestLedgerResponseFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetLatestLedgerResponseFromProto(nil)
	require.EqualError(t, err, "get latest ledger response is nil")
}

func TestConvertReadContractRequest_RoundTrip(t *testing.T) {
	boolVal := true
	u32 := xdr.Uint32(77)
	sym := xdr.ScSymbol("hello")
	domain := stellartypes.ReadContractRequest{
		ContractID: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:   "transfer",
		Args: []xdr.ScVal{
			{Type: xdr.ScValTypeScvBool, B: &boolVal},
			{Type: xdr.ScValTypeScvU32, U32: &u32},
			{Type: xdr.ScValTypeScvSymbol, Sym: &sym},
		},
		LedgerSequence: 42,
	}

	proto, err := conv.ConvertReadContractRequestToProto(domain)
	require.NoError(t, err)
	require.Len(t, proto.GetArgs(), 3)

	got, err := conv.ConvertReadContractRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertReadContractRequestFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(nil)
	require.EqualError(t, err, "readContractRequest is nil")
}

func TestConvertReadContractRequestFromProto_MissingContractID(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(&conv.ReadContractRequest{Function: "fn"})
	require.EqualError(t, err, "contractID is required")
}

func TestConvertReadContractRequestFromProto_MissingFunction(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(&conv.ReadContractRequest{ContractId: "C123"})
	require.EqualError(t, err, "function is required")
}

func TestConvertReadContractRequestToProto_MissingContractID(t *testing.T) {
	_, err := conv.ConvertReadContractRequestToProto(stellartypes.ReadContractRequest{Function: "fn"})
	require.EqualError(t, err, "contractID is required")
}

func TestConvertReadContractRequestToProto_MissingFunction(t *testing.T) {
	_, err := conv.ConvertReadContractRequestToProto(stellartypes.ReadContractRequest{ContractID: "C_X"})
	require.EqualError(t, err, "function is required")
}

func TestConvertReadContractRequestToProto_BadArg(t *testing.T) {
	// A ScVal with a nil arm pointer triggers an error that must be wrapped as args[0].
	_, err := conv.ConvertReadContractRequestToProto(stellartypes.ReadContractRequest{
		ContractID: "C_X",
		Function:   "fn",
		Args:       []xdr.ScVal{{Type: xdr.ScValTypeScvBool}}, // B is nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "args[0]")
}

// ---- ConvertLedgerEntryResultToProto invalid XDR fields ---------------------

func TestConvertLedgerEntryResultToProto_InvalidDataXDR(t *testing.T) {
	_, err := conv.ConvertLedgerEntryResultToProto(stellartypes.LedgerEntryResult{
		KeyXDR:       base64.StdEncoding.EncodeToString([]byte("k")),
		DataXDR:      "!!!invalid!!!",
		ExtensionXDR: base64.StdEncoding.EncodeToString([]byte("e")),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid data xdr")
}

func TestConvertLedgerEntryResultToProto_InvalidExtensionXDR(t *testing.T) {
	_, err := conv.ConvertLedgerEntryResultToProto(stellartypes.LedgerEntryResult{
		KeyXDR:       base64.StdEncoding.EncodeToString([]byte("k")),
		DataXDR:      base64.StdEncoding.EncodeToString([]byte("d")),
		ExtensionXDR: "!!!invalid!!!",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid extension xdr")
}

// ---- GetLedgerEntriesResponse entry error propagation -----------------------

func TestConvertGetLedgerEntriesResponseToProto_BadEntry(t *testing.T) {
	_, err := conv.ConvertGetLedgerEntriesResponseToProto(stellartypes.GetLedgerEntriesResponse{
		Entries: []stellartypes.LedgerEntryResult{
			{KeyXDR: "!!!bad!!!", DataXDR: "", ExtensionXDR: ""},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "entry[0]")
}

func TestConvertGetLedgerEntriesResponseFromProto_NilEntry(t *testing.T) {
	_, err := conv.ConvertGetLedgerEntriesResponseFromProto(&conv.GetLedgerEntriesResponse{
		Entries: []*conv.LedgerEntryResult{nil},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "entry[0]")
}

// ---- ConvertGetLatestLedgerResponseToProto error cases ----------------------

func TestConvertGetLatestLedgerResponseToProto_InvalidFields(t *testing.T) {
	tests := []struct {
		name    string
		in      stellartypes.GetLatestLedgerResponse
		wantErr string
	}{
		{"invalid hash", stellartypes.GetLatestLedgerResponse{Hash: "not-hex!"}, "invalid hex hash"},
		{"invalid header xdr", stellartypes.GetLatestLedgerResponse{LedgerHeaderXDR: "!!!bad!!!"}, "invalid ledger header xdr"},
		{"invalid metadata xdr", stellartypes.GetLatestLedgerResponse{LedgerMetadataXDR: "!!!bad!!!"}, "invalid ledger metadata xdr"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := conv.ConvertGetLatestLedgerResponseToProto(tc.in)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
