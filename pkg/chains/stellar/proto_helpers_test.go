package stellar_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	conv "github.com/smartcontractkit/chainlink-common/pkg/chains/stellar"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar/scval"
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
	u32 := uint32(77)
	sym := "hello"
	domain := stellartypes.ReadContractRequest{
		ContractID: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:   "transfer",
		Args: []stellartypes.ScVal{
			{Type: stellartypes.ScValTypeBool, Bool: &boolVal},
			{Type: stellartypes.ScValTypeU32, U32: &u32},
			{Type: stellartypes.ScValTypeSymbol, Symbol: &sym},
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

func TestConvertReadContractRequest_RoundTrip_RichArgs(t *testing.T) {
	// End-to-end coverage of the chains-level pipeline with a mix of ScVal shapes:
	// address (account), nested vec-of-map, and contract-instance-with-storage.
	accountBytes := make([]byte, 32)
	for i := range accountBytes {
		accountBytes[i] = 0x11
	}
	wasmHash := make([]byte, 32)
	for i := range wasmHash {
		wasmHash[i] = 0x22
	}
	addrArg := stellartypes.ScVal{
		Type: stellartypes.ScValTypeAddress,
		Address: &stellartypes.ScAddress{
			Type:      stellartypes.ScAddressTypeAccountID,
			AccountID: accountBytes,
		},
	}
	mapKey := "amount"
	mapVal := uint64(1_000_000)
	innerMap := &stellartypes.ScVal{
		Type: stellartypes.ScValTypeMap,
		Map: &stellartypes.ScMap{Entries: []stellartypes.ScMapEntry{
			{
				Key: &stellartypes.ScVal{Type: stellartypes.ScValTypeSymbol, Symbol: &mapKey},
				Val: &stellartypes.ScVal{Type: stellartypes.ScValTypeU64, U64: &mapVal},
			},
		}},
	}
	vecArg := stellartypes.ScVal{
		Type: stellartypes.ScValTypeVec,
		Vec:  &stellartypes.ScVec{Values: []*stellartypes.ScVal{innerMap}},
	}
	slot := "slot"
	slotVal := uint32(7)
	instanceArg := stellartypes.ScVal{
		Type: stellartypes.ScValTypeContractInstance,
		ContractInstance: &stellartypes.ScContractInstance{
			Executable: &stellartypes.ContractExecutable{
				Type:     stellartypes.ContractExecutableTypeWasmHash,
				WasmHash: wasmHash,
			},
			Storage: []stellartypes.ScMapEntry{
				{
					Key: &stellartypes.ScVal{Type: stellartypes.ScValTypeSymbol, Symbol: &slot},
					Val: &stellartypes.ScVal{Type: stellartypes.ScValTypeU32, U32: &slotVal},
				},
			},
		},
	}

	domain := stellartypes.ReadContractRequest{
		ContractID:     "C_RICH",
		Function:       "do_work",
		Args:           []stellartypes.ScVal{addrArg, vecArg, instanceArg},
		LedgerSequence: 9,
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
		Args:       []stellartypes.ScVal{{Type: stellartypes.ScValTypeBool}}, // Bool is nil
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

func TestConvertSubmitTransactionRequest_RoundTrip(t *testing.T) {
	boolVal := true
	u64 := uint64(42)
	sym := "amount"
	domain := stellartypes.SubmitTransactionRequest{
		IdempotencyKey:     "idem-123",
		FromAddress:        "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWHF",
		ContractID:         "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:           "transfer",
		Args:               []stellartypes.ScVal{
			{Type: stellartypes.ScValTypeBool, Bool: &boolVal},
			{Type: stellartypes.ScValTypeU64, U64: &u64},
			{Type: stellartypes.ScValTypeSymbol, Symbol: &sym},
		},
		LedgerBoundsOffset: 10,
	}

	proto, err := conv.ConvertSubmitTransactionRequestToProto(domain)
	require.NoError(t, err)
	require.Equal(t, "idem-123", proto.GetIdempotencyKey())
	require.Equal(t, "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4", proto.GetContractId())
	require.Equal(t, "transfer", proto.GetFunction())
	require.Len(t, proto.GetArgs(), 3)
	require.Equal(t, uint32(10), proto.GetLedgerBoundsOffset())

	got, err := conv.ConvertSubmitTransactionRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertSubmitTransactionRequest_NoArgs(t *testing.T) {
	domain := stellartypes.SubmitTransactionRequest{
		ContractID: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:   "ping",
	}

	proto, err := conv.ConvertSubmitTransactionRequestToProto(domain)
	require.NoError(t, err)
	require.Empty(t, proto.GetArgs())

	got, err := conv.ConvertSubmitTransactionRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertSubmitTransactionRequestToProto_MissingContractID(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionRequestToProto(stellartypes.SubmitTransactionRequest{Function: "fn"})
	require.EqualError(t, err, "contractId is required")
}

func TestConvertSubmitTransactionRequestToProto_MissingFunction(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionRequestToProto(stellartypes.SubmitTransactionRequest{ContractID: "C_X"})
	require.EqualError(t, err, "function is required")
}

func TestConvertSubmitTransactionRequestToProto_BadArg(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionRequestToProto(stellartypes.SubmitTransactionRequest{
		ContractID: "C_X",
		Function:   "fn",
		Args:       []stellartypes.ScVal{{Type: stellartypes.ScValTypeBool}}, // Bool is nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "args[0]")
}

func TestConvertSubmitTransactionRequestFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionRequestFromProto(nil)
	require.EqualError(t, err, "submit transaction request is nil")
}

func TestConvertSubmitTransactionRequestFromProto_MissingContractID(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionRequestFromProto(&conv.SubmitTransactionRequest{Function: "fn"})
	require.EqualError(t, err, "contractId is required")
}

func TestConvertSubmitTransactionRequestFromProto_MissingFunction(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionRequestFromProto(&conv.SubmitTransactionRequest{ContractId: "C_X"})
	require.EqualError(t, err, "function is required")
}

func TestConvertSubmitTransactionResponse_RoundTrip(t *testing.T) {
	domain := &stellartypes.SubmitTransactionResponse{
		TxStatus:         stellartypes.TxSuccess,
		TxHash:           "abc123hash",
		TxIdempotencyKey: "idem-456",
		ResultXDR:        base64.StdEncoding.EncodeToString([]byte("result")),
		ResultMetaXDR:    base64.StdEncoding.EncodeToString([]byte("meta")),
	}

	proto, err := conv.ConvertSubmitTransactionResponseToProto(domain)
	require.NoError(t, err)
	require.Equal(t, conv.TxStatus_TX_STATUS_SUCCESS, proto.GetTxStatus())

	got, err := conv.ConvertSubmitTransactionResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertSubmitTransactionResponse_RoundTrip_EmptyResultFields(t *testing.T) {
	domain := &stellartypes.SubmitTransactionResponse{
		TxStatus:         stellartypes.TxFatal,
		TxHash:           "",
		TxIdempotencyKey: "idem-789",
	}

	proto, err := conv.ConvertSubmitTransactionResponseToProto(domain)
	require.NoError(t, err)

	got, err := conv.ConvertSubmitTransactionResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertSubmitTransactionResponseToProto_Nil(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionResponseToProto(nil)
	require.EqualError(t, err, "submit transaction reply is nil")
}

func TestConvertSubmitTransactionResponseFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionResponseFromProto(nil)
	require.EqualError(t, err, "submit transaction reply is nil")
}

func TestConvertSubmitTransactionResponseToProto_InvalidResultXDR(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionResponseToProto(&stellartypes.SubmitTransactionResponse{
		ResultXDR: "!!!invalid!!!",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid result xdr")
}

func TestConvertSubmitTransactionResponseToProto_InvalidResultMetaXDR(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionResponseToProto(&stellartypes.SubmitTransactionResponse{
		ResultMetaXDR: "!!!invalid!!!",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid result meta xdr")
}

func TestConvertSubmitTransactionRequestFromProto_BadArg(t *testing.T) {
	_, err := conv.ConvertSubmitTransactionRequestFromProto(&conv.SubmitTransactionRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args:       []*scval.ScVal{{}}, // missing oneof value
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "args[0]")
}

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
