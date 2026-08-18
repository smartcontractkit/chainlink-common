package stellar_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	stellarcap "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar/scval"
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

func TestConvertGetLedgersRequest_RoundTrip(t *testing.T) {
	domain := stellartypes.GetLedgersRequest{
		Pagination: &stellartypes.LedgerPaginationOptions{Cursor: "cur-1", Limit: 5},
	}

	proto, err := conv.ConvertGetLedgersRequestToProto(domain)
	require.NoError(t, err)
	require.Equal(t, uint32(0), proto.GetStartLedger())
	require.Equal(t, "cur-1", proto.GetPagination().GetCursor())
	require.Equal(t, uint32(5), proto.GetPagination().GetLimit())

	got, err := conv.ConvertGetLedgersRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetLedgersRequest_RejectsStartLedgerAndCursorTogether(t *testing.T) {
	domain := stellartypes.GetLedgersRequest{
		StartLedger: 987654,
		Pagination:  &stellartypes.LedgerPaginationOptions{Cursor: "cur-1", Limit: 5},
	}

	_, err := conv.ConvertGetLedgersRequestToProto(domain)
	require.Error(t, err)
	require.Contains(t, err.Error(), "mutually exclusive")

	_, err = conv.ConvertGetLedgersRequestFromProto(&conv.GetLedgersRequest{
		StartLedger: 987654,
		Pagination:  &conv.LedgerPaginationOptions{Cursor: "cur-1", Limit: 5},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "mutually exclusive")
}

func TestConvertGetLedgersRequest_RejectsNeitherStartLedgerNorCursor(t *testing.T) {
	_, err := conv.ConvertGetLedgersRequestToProto(stellartypes.GetLedgersRequest{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "startLedger is required")

	_, err = conv.ConvertGetLedgersRequestFromProto(&conv.GetLedgersRequest{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "startLedger is required")
}

func TestConvertGetLedgersRequest_RoundTrip_NoPagination(t *testing.T) {
	domain := stellartypes.GetLedgersRequest{StartLedger: 42}

	proto, err := conv.ConvertGetLedgersRequestToProto(domain)
	require.NoError(t, err)
	require.Nil(t, proto.GetPagination())

	got, err := conv.ConvertGetLedgersRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetLedgersRequestFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetLedgersRequestFromProto(nil)
	require.EqualError(t, err, "get ledgers request is nil")
}

func TestConvertGetLedgersResponse_RoundTrip(t *testing.T) {
	var b [32]byte
	for i := range b {
		b[i] = 0xCD
	}
	hash := b
	domain := stellartypes.GetLedgersResponse{
		Ledgers: []stellartypes.LedgerInfo{
			{
				Hash:              hex.EncodeToString(hash[:]),
				Sequence:          1234567,
				LedgerCloseTime:   1_700_000_000,
				LedgerHeaderXDR:   base64.StdEncoding.EncodeToString([]byte("header-xdr")),
				LedgerMetadataXDR: base64.StdEncoding.EncodeToString([]byte("meta-xdr")),
			},
			{
				Hash:              hex.EncodeToString(hash[:]),
				Sequence:          1234568,
				LedgerCloseTime:   1_700_000_005,
				LedgerHeaderXDR:   base64.StdEncoding.EncodeToString([]byte("header-xdr-2")),
				LedgerMetadataXDR: base64.StdEncoding.EncodeToString([]byte("meta-xdr-2")),
			},
		},
		LatestLedger:          1234570,
		LatestLedgerCloseTime: 1_700_000_010,
		OldestLedger:          1000000,
		OldestLedgerCloseTime: 1_600_000_000,
		Cursor:                "next-cursor",
	}

	proto, err := conv.ConvertGetLedgersResponseToProto(domain)
	require.NoError(t, err)

	got, err := conv.ConvertGetLedgersResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetLedgersResponse_RoundTrip_NoLedgers(t *testing.T) {
	domain := stellartypes.GetLedgersResponse{
		Ledgers:      []stellartypes.LedgerInfo{},
		LatestLedger: 55,
		OldestLedger: 1,
		Cursor:       "cur",
	}

	proto, err := conv.ConvertGetLedgersResponseToProto(domain)
	require.NoError(t, err)

	got, err := conv.ConvertGetLedgersResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetLedgersResponseFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetLedgersResponseFromProto(nil)
	require.EqualError(t, err, "get ledgers response is nil")
}

func TestConvertLedgerInfoFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertLedgerInfoFromProto(nil)
	require.EqualError(t, err, "ledger info is nil")
}

func TestConvertLedgerInfoToProto_InvalidFields(t *testing.T) {
	cases := []struct {
		name    string
		in      stellartypes.LedgerInfo
		wantErr string
	}{
		{"invalid hash", stellartypes.LedgerInfo{Hash: "not-hex!"}, "invalid hex hash"},
		{"invalid header xdr", stellartypes.LedgerInfo{LedgerHeaderXDR: "!!!bad!!!"}, "invalid ledger header xdr"},
		{"invalid metadata xdr", stellartypes.LedgerInfo{LedgerMetadataXDR: "!!!bad!!!"}, "invalid ledger metadata xdr"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := conv.ConvertLedgerInfoToProto(tc.in)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestConvertGetLedgersResponseToProto_BadLedger(t *testing.T) {
	_, err := conv.ConvertGetLedgersResponseToProto(stellartypes.GetLedgersResponse{
		Ledgers: []stellartypes.LedgerInfo{{Hash: "not-hex!"}},
	})
	require.ErrorContains(t, err, "ledgers[0]")
}

func TestConvertSimulateTransactionRequest_RoundTrip(t *testing.T) {
	boolVal := true
	u32 := uint32(77)
	sym := "hello"
	domain := stellartypes.SimulateTransactionRequest{
		ContractID: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:   "report",
		Args: []stellartypes.ScVal{
			{Type: stellartypes.ScValTypeBool, Bool: &boolVal},
			{Type: stellartypes.ScValTypeU32, U32: &u32},
			{Type: stellartypes.ScValTypeSymbol, Symbol: &sym},
		},
		SourceAccount: "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
		AuthMode:      stellartypes.SimulateAuthModeRecord,
		ResourceConfig: &stellartypes.SimulateResourceConfig{
			InstructionLeeway: 10_000,
		},
	}

	proto, err := conv.ConvertSimulateTransactionRequestToProto(domain)
	require.NoError(t, err)
	require.Equal(t, domain.ContractID, proto.GetContractId())
	require.Equal(t, domain.Function, proto.GetFunction())
	require.Len(t, proto.GetArgs(), 3)
	require.Equal(t, string(stellartypes.SimulateAuthModeRecord), proto.GetAuthMode())
	require.NotNil(t, proto.GetResourceConfig())
	require.Equal(t, uint64(10_000), proto.GetResourceConfig().GetInstructionLeeway())

	got, err := conv.ConvertSimulateTransactionRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertSimulateTransactionRequest_RoundTrip_NoArgsNoAuthModeNoResourceConfig(t *testing.T) {
	domain := stellartypes.SimulateTransactionRequest{
		ContractID: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:   "ping",
	}

	proto, err := conv.ConvertSimulateTransactionRequestToProto(domain)
	require.NoError(t, err)
	require.Empty(t, proto.GetArgs())
	require.Empty(t, proto.GetAuthMode())
	require.Nil(t, proto.GetResourceConfig())

	got, err := conv.ConvertSimulateTransactionRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertSimulateTransactionRequest_RoundTrip_RichArgs(t *testing.T) {
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

	domain := stellartypes.SimulateTransactionRequest{
		ContractID: "C_RICH",
		Function:   "do_work",
		Args:       []stellartypes.ScVal{addrArg, vecArg, instanceArg},
		AuthMode:   stellartypes.SimulateAuthModeRecordAllowNonroot,
	}

	proto, err := conv.ConvertSimulateTransactionRequestToProto(domain)
	require.NoError(t, err)
	require.Len(t, proto.GetArgs(), 3)
	require.Equal(t, string(stellartypes.SimulateAuthModeRecordAllowNonroot), proto.GetAuthMode())

	got, err := conv.ConvertSimulateTransactionRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertSimulateTransactionRequestFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertSimulateTransactionRequestFromProto(nil)
	require.EqualError(t, err, "simulateTransaction request is nil")
}

func TestConvertSimulateTransactionRequestFromProto_MissingContractID(t *testing.T) {
	_, err := conv.ConvertSimulateTransactionRequestFromProto(&conv.SimulateTransactionRequest{Function: "fn"})
	require.EqualError(t, err, "contractID is required")
}

func TestConvertSimulateTransactionRequestFromProto_MissingFunction(t *testing.T) {
	_, err := conv.ConvertSimulateTransactionRequestFromProto(&conv.SimulateTransactionRequest{ContractId: "C123"})
	require.EqualError(t, err, "function is required")
}

func TestConvertSimulateTransactionRequestFromProto_UnsupportedAuthMode(t *testing.T) {
	_, err := conv.ConvertSimulateTransactionRequestFromProto(&conv.SimulateTransactionRequest{
		ContractId: "C123",
		Function:   "fn",
		AuthMode:   "unsupported",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported auth mode")
}

func TestConvertSimulateTransactionRequestToProto_MissingContractID(t *testing.T) {
	_, err := conv.ConvertSimulateTransactionRequestToProto(stellartypes.SimulateTransactionRequest{Function: "fn"})
	require.EqualError(t, err, "contractID is required")
}

func TestConvertSimulateTransactionRequestToProto_MissingFunction(t *testing.T) {
	_, err := conv.ConvertSimulateTransactionRequestToProto(stellartypes.SimulateTransactionRequest{ContractID: "C_X"})
	require.EqualError(t, err, "function is required")
}

func TestConvertSimulateTransactionRequestToProto_UnsupportedAuthMode(t *testing.T) {
	_, err := conv.ConvertSimulateTransactionRequestToProto(stellartypes.SimulateTransactionRequest{
		ContractID: "C_X",
		Function:   "fn",
		AuthMode:   "unsupported",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported auth mode")
}

func TestConvertSimulateTransactionRequestToProto_BadArg(t *testing.T) {
	_, err := conv.ConvertSimulateTransactionRequestToProto(stellartypes.SimulateTransactionRequest{
		ContractID: "C_X",
		Function:   "fn",
		Args:       []stellartypes.ScVal{{Type: stellartypes.ScValTypeBool}}, // Bool is nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "args[0]")
}

func TestConvertSimulateTransactionRequestFromProto_BadArg(t *testing.T) {
	_, err := conv.ConvertSimulateTransactionRequestFromProto(&conv.SimulateTransactionRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args:       []*scval.ScVal{{}}, // missing oneof value
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "args[0]")
}

func TestConvertSubmitTransactionResponse_RoundTrip_WithOptionalFeeAndTimestamp(t *testing.T) {
	fee := uint64(123456)
	blockTimestamp := uint64(1_700_000_000_000_000)

	domain := &stellartypes.SubmitTransactionResponse{
		TxStatus:         stellartypes.TxSuccess,
		TxHash:           "hash-with-fee",
		TxIdempotencyKey: "idem-with-fee",
		ResultXDR:        base64.StdEncoding.EncodeToString([]byte("result")),
		ResultMetaXDR:    base64.StdEncoding.EncodeToString([]byte("meta")),
		TransactionFee:   &fee,
		BlockTimestamp:   &blockTimestamp,
	}

	proto, err := conv.ConvertSubmitTransactionResponseToProto(domain)
	require.NoError(t, err)
	require.NotNil(t, proto.TransactionFee)
	require.Equal(t, fee, proto.GetTransactionFee())
	require.NotNil(t, proto.BlockTimestamp)
	require.Equal(t, blockTimestamp, proto.GetBlockTimestamp())

	got, err := conv.ConvertSubmitTransactionResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
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
		IdempotencyKey: "idem-123",
		FromAddress:    "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWHF",
		ContractID:     "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:       "transfer",
		Args: []stellartypes.ScVal{
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
	fee := uint64(12_345)
	blockTimestamp := uint64(1_700_000_000_000_000) // microseconds
	domain := &stellartypes.SubmitTransactionResponse{
		TxStatus:         stellartypes.TxSuccess,
		TxHash:           "abc123hash",
		TxIdempotencyKey: "idem-456",
		ResultXDR:        base64.StdEncoding.EncodeToString([]byte("result")),
		ResultMetaXDR:    base64.StdEncoding.EncodeToString([]byte("meta")),
		Error:            "",
		TransactionFee:   &fee,
		BlockTimestamp:   &blockTimestamp,
	}

	proto, err := conv.ConvertSubmitTransactionResponseToProto(domain)
	require.NoError(t, err)
	require.Equal(t, conv.TxStatus_TX_STATUS_SUCCESS, proto.GetTxStatus())
	require.Empty(t, proto.GetError())
	require.Equal(t, blockTimestamp, proto.GetBlockTimestamp())

	got, err := conv.ConvertSubmitTransactionResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertSubmitTransactionResponse_RoundTrip_WithError(t *testing.T) {
	domain := &stellartypes.SubmitTransactionResponse{
		TxStatus:         stellartypes.TxFailed,
		TxHash:           "failhash",
		TxIdempotencyKey: "idem-fail",
		ResultXDR:        base64.StdEncoding.EncodeToString([]byte("failed-result")),
		Error:            "transaction result: InvokeHostFunctionTrapped",
	}

	proto, err := conv.ConvertSubmitTransactionResponseToProto(domain)
	require.NoError(t, err)
	require.Equal(t, conv.TxStatus_TX_STATUS_FAILED, proto.GetTxStatus())
	require.Equal(t, domain.Error, proto.GetError())

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

func TestConvertGetEventsRequest_RoundTrip(t *testing.T) {
	symTransfer := "transfer"
	account := "from"
	wildcard := "*"
	flexibleWildcard := "**"

	domain := stellartypes.GetEventsRequest{
		StartLedger: 100,
		EndLedger:   200,
		Filters: []stellartypes.EventFilter{
			{
				EventTypes:  []stellartypes.EventType{stellartypes.EventTypeContract},
				ContractIDs: []string{"CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4"},
				Topics: []stellartypes.TopicFilter{
					{
						Segments: []stellartypes.TopicSegment{
							{
								Value: &stellartypes.ScVal{
									Type:   stellartypes.ScValTypeSymbol,
									Symbol: &symTransfer,
								},
							},
							{Wildcard: &wildcard},
						},
					},
					{
						Segments: []stellartypes.TopicSegment{
							{
								Value: &stellartypes.ScVal{
									Type:   stellartypes.ScValTypeSymbol,
									Symbol: &account,
								},
							},
							{Wildcard: &flexibleWildcard},
						},
					},
				},
			},
		},
		Pagination: &stellartypes.PaginationOptions{
			Cursor: "0000000010-0000000001",
			Limit:  50,
		},
	}

	proto, err := conv.ConvertGetEventsRequestToProto(domain)
	require.NoError(t, err)
	require.Equal(t, uint32(100), proto.GetStartLedger())
	require.Equal(t, uint32(200), proto.GetEndLedger())
	require.Len(t, proto.GetFilters(), 1)
	require.Equal(t, "0000000010-0000000001", proto.GetPagination().GetCursor())
	require.Equal(t, uint32(50), proto.GetPagination().GetLimit())

	got, err := conv.ConvertGetEventsRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetEventsRequest_RoundTrip_NoFiltersNoPagination(t *testing.T) {
	domain := stellartypes.GetEventsRequest{
		StartLedger: 100,
		EndLedger:   200,
	}

	proto, err := conv.ConvertGetEventsRequestToProto(domain)
	require.NoError(t, err)
	require.Empty(t, proto.GetFilters())
	require.Nil(t, proto.GetPagination())

	got, err := conv.ConvertGetEventsRequestFromProto(proto)
	require.NoError(t, err)

	require.Equal(t, domain.StartLedger, got.StartLedger)
	require.Equal(t, domain.EndLedger, got.EndLedger)
	require.Empty(t, got.Filters)
	require.Nil(t, got.Pagination)
}
func TestConvertGetEventsRequestFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetEventsRequestFromProto(nil)
	require.EqualError(t, err, "get events request is nil")
}

func TestConvertGetEventsRequestToProto_EmptyTopicFilter(t *testing.T) {
	_, err := conv.ConvertGetEventsRequestToProto(stellartypes.GetEventsRequest{
		StartLedger: 1,
		Filters: []stellartypes.EventFilter{
			{
				Topics: []stellartypes.TopicFilter{
					{},
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "filters[0]")
	require.Contains(t, err.Error(), "topics[0]")
	require.Contains(t, err.Error(), "topic filter must have at least one segment")
}

func TestConvertGetEventsRequestFromProto_EmptyTopicFilter(t *testing.T) {
	_, err := conv.ConvertGetEventsRequestFromProto(&conv.GetEventsRequest{
		StartLedger: 1,
		Filters: []*conv.EventFilter{
			{
				Topics: []*conv.TopicFilter{
					{},
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "filters[0]")
	require.Contains(t, err.Error(), "topics[0]")
	require.Contains(t, err.Error(), "topic filter must have at least one segment")
}

func TestConvertGetEventsRequestToProto_EmptyTopicSegment(t *testing.T) {
	_, err := conv.ConvertGetEventsRequestToProto(stellartypes.GetEventsRequest{
		StartLedger: 1,
		Filters: []stellartypes.EventFilter{
			{
				Topics: []stellartypes.TopicFilter{
					{
						Segments: []stellartypes.TopicSegment{{}},
					},
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "filters[0]")
	require.Contains(t, err.Error(), "topics[0]")
	require.Contains(t, err.Error(), "segments[0]")
	require.Contains(t, err.Error(), "topic segment must set either wildcard or value")
}

func TestConvertGetEventsRequestToProto_TopicSegmentBothWildcardAndValue(t *testing.T) {
	wildcard := "*"
	sym := "transfer"

	_, err := conv.ConvertGetEventsRequestToProto(stellartypes.GetEventsRequest{
		StartLedger: 1,
		Filters: []stellartypes.EventFilter{
			{
				Topics: []stellartypes.TopicFilter{
					{
						Segments: []stellartypes.TopicSegment{
							{
								Wildcard: &wildcard,
								Value: &stellartypes.ScVal{
									Type:   stellartypes.ScValTypeSymbol,
									Symbol: &sym,
								},
							},
						},
					},
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "segments[0]")
	require.Contains(t, err.Error(), "topic segment cannot set both wildcard and value")
}

func TestConvertGetEventsRequestToProto_InvalidWildcard(t *testing.T) {
	wildcard := "bad"

	_, err := conv.ConvertGetEventsRequestToProto(stellartypes.GetEventsRequest{
		StartLedger: 1,
		Filters: []stellartypes.EventFilter{
			{
				Topics: []stellartypes.TopicFilter{
					{
						Segments: []stellartypes.TopicSegment{
							{Wildcard: &wildcard},
						},
					},
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "segments[0]")
	require.Contains(t, err.Error(), "wildcard must be '*' or '**'")
}

func TestConvertGetEventsRequestFromProto_InvalidWildcard(t *testing.T) {
	_, err := conv.ConvertGetEventsRequestFromProto(&conv.GetEventsRequest{
		StartLedger: 1,
		Filters: []*conv.EventFilter{
			{
				Topics: []*conv.TopicFilter{
					{
						Segments: []*conv.TopicSegment{
							{
								Value: &conv.TopicSegment_Wildcard{Wildcard: "bad"},
							},
						},
					},
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "segments[0]")
	require.Contains(t, err.Error(), "wildcard must be '*' or '**'")
}

func TestConvertGetEventsRequestToProto_BadTopicScVal(t *testing.T) {
	_, err := conv.ConvertGetEventsRequestToProto(stellartypes.GetEventsRequest{
		StartLedger: 1,
		Filters: []stellartypes.EventFilter{
			{
				Topics: []stellartypes.TopicFilter{
					{
						Segments: []stellartypes.TopicSegment{
							{
								Value: &stellartypes.ScVal{
									Type: stellartypes.ScValTypeBool, // Bool is nil
								},
							},
						},
					},
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "segments[0]")
}

func TestConvertGetEventsRequestFromProto_BadTopicScVal(t *testing.T) {
	_, err := conv.ConvertGetEventsRequestFromProto(&conv.GetEventsRequest{
		StartLedger: 1,
		Filters: []*conv.EventFilter{
			{
				Topics: []*conv.TopicFilter{
					{
						Segments: []*conv.TopicSegment{
							{
								Value: &conv.TopicSegment_Scval{Scval: &scval.ScVal{}},
							},
						},
					},
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "segments[0]")
}

func TestConvertGetEventsRequestToProto_UnsupportedEventType(t *testing.T) {
	_, err := conv.ConvertGetEventsRequestToProto(stellartypes.GetEventsRequest{
		StartLedger: 1,
		Filters: []stellartypes.EventFilter{
			{
				EventTypes: []stellartypes.EventType{stellartypes.EventType(99)},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "filters[0]")
	require.Contains(t, err.Error(), "eventTypes[0]")
	require.Contains(t, err.Error(), "unsupported event type")
}

func TestConvertGetEventsResponse_RoundTrip(t *testing.T) {
	topicSym := "transfer"
	valueU64 := uint64(12345)

	domain := stellartypes.GetEventsResponse{
		Events: []stellartypes.EventInfo{
			{
				EventType:        stellartypes.EventTypeContract,
				Ledger:           123,
				LedgerClosedAt:   "2025-01-01T00:00:00Z",
				ContractID:       "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
				ID:               "0000000123-0000000001",
				OperationIndex:   1,
				TransactionIndex: 2,
				TransactionHash:  "abc123",
				Topics: []stellartypes.ScVal{
					{
						Type:   stellartypes.ScValTypeSymbol,
						Symbol: &topicSym,
					},
				},
				Value: stellartypes.ScVal{
					Type: stellartypes.ScValTypeU64,
					U64:  &valueU64,
				},
			},
		},
		Cursor:                "0000000123-0000000001",
		LatestLedger:          200,
		OldestLedger:          100,
		LatestLedgerCloseTime: 1_700_000_100,
		OldestLedgerCloseTime: 1_700_000_000,
	}

	proto, err := conv.ConvertGetEventsResponseToProto(domain)
	require.NoError(t, err)
	require.Len(t, proto.GetEvents(), 1)
	require.NotNil(t, proto.GetEvents()[0].GetValue())

	got, err := conv.ConvertGetEventsResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetEventsResponse_RoundTrip_NoEvents(t *testing.T) {
	domain := stellartypes.GetEventsResponse{
		Cursor:                "",
		LatestLedger:          200,
		OldestLedger:          100,
		LatestLedgerCloseTime: 1_700_000_100,
		OldestLedgerCloseTime: 1_700_000_000,
	}

	proto, err := conv.ConvertGetEventsResponseToProto(domain)
	require.NoError(t, err)
	require.Empty(t, proto.GetEvents())

	got, err := conv.ConvertGetEventsResponseFromProto(proto)
	require.NoError(t, err)

	require.Empty(t, got.Events)
	require.Equal(t, domain.Cursor, got.Cursor)
	require.Equal(t, domain.LatestLedger, got.LatestLedger)
	require.Equal(t, domain.OldestLedger, got.OldestLedger)
	require.Equal(t, domain.LatestLedgerCloseTime, got.LatestLedgerCloseTime)
	require.Equal(t, domain.OldestLedgerCloseTime, got.OldestLedgerCloseTime)
}

func TestConvertGetEventsResponseFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetEventsResponseFromProto(nil)
	require.EqualError(t, err, "get events response is nil")
}

func TestConvertGetEventsResponseFromProto_NilEvent(t *testing.T) {
	_, err := conv.ConvertGetEventsResponseFromProto(&conv.GetEventsResponse{
		Events: []*conv.EventInfo{nil},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "events[0]")
	require.Contains(t, err.Error(), "event info is nil")
}

func TestConvertGetEventsResponseFromProto_MissingValue(t *testing.T) {
	_, err := conv.ConvertGetEventsResponseFromProto(&conv.GetEventsResponse{
		Events: []*conv.EventInfo{
			{
				EventType: conv.EventType_EVENT_TYPE_CONTRACT,
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "events[0]")
	require.Contains(t, err.Error(), "value is required")
}

func TestConvertGetEventsResponseToProto_BadValue(t *testing.T) {
	_, err := conv.ConvertGetEventsResponseToProto(stellartypes.GetEventsResponse{
		Events: []stellartypes.EventInfo{
			{
				EventType: stellartypes.EventTypeContract,
				Value: stellartypes.ScVal{
					Type: stellartypes.ScValTypeBool, // Bool is nil
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "events[0]")
	require.Contains(t, err.Error(), "value")
}

func TestConvertGetEventsResponseToProto_BadTopic(t *testing.T) {
	u64 := uint64(1)

	_, err := conv.ConvertGetEventsResponseToProto(stellartypes.GetEventsResponse{
		Events: []stellartypes.EventInfo{
			{
				EventType: stellartypes.EventTypeContract,
				Topics: []stellartypes.ScVal{
					{Type: stellartypes.ScValTypeBool}, // Bool is nil
				},
				Value: stellartypes.ScVal{
					Type: stellartypes.ScValTypeU64,
					U64:  &u64,
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "events[0]")
	require.Contains(t, err.Error(), "topics[0]")
}

func TestConvertGetEventsResponseFromProto_BadTopic(t *testing.T) {
	u64 := uint64(1)
	value, err := stellarcap.ScValToProto(stellartypes.ScVal{
		Type: stellartypes.ScValTypeU64,
		U64:  &u64,
	})
	require.NoError(t, err)

	_, err = conv.ConvertGetEventsResponseFromProto(&conv.GetEventsResponse{
		Events: []*conv.EventInfo{
			{
				EventType: conv.EventType_EVENT_TYPE_CONTRACT,
				Topics: []*scval.ScVal{
					{},
				},
				Value: value,
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "events[0]")
	require.Contains(t, err.Error(), "topics[0]")
}

func TestConvertGetEventsResponseToProto_UnsupportedEventType(t *testing.T) {
	u64 := uint64(1)

	_, err := conv.ConvertGetEventsResponseToProto(stellartypes.GetEventsResponse{
		Events: []stellartypes.EventInfo{
			{
				EventType: stellartypes.EventType(99),
				Value: stellartypes.ScVal{
					Type: stellartypes.ScValTypeU64,
					U64:  &u64,
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "events[0]")
	require.Contains(t, err.Error(), "eventType")
	require.Contains(t, err.Error(), "unsupported event type")
}

func TestConvertGetTransactionRequest_RoundTrip(t *testing.T) {
	domain := stellartypes.GetTransactionRequest{TxHash: "abc123hash"}
	proto := conv.ConvertGetTransactionRequestToProto(domain)
	require.Equal(t, "abc123hash", proto.GetTxHash())

	got, err := conv.ConvertGetTransactionRequestFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetTransactionRequestFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetTransactionRequestFromProto(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil")
}

func TestConvertGetTransactionRequestFromProto_EmptyTxHash(t *testing.T) {
	_, err := conv.ConvertGetTransactionRequestFromProto(&conv.GetTransactionRequest{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "tx hash is required")
}

func TestConvertGetTransactionResponse_RoundTrip(t *testing.T) {
	domain := stellartypes.GetTransactionResponse{
		FeeStroops:      42,
		LedgerSequence:  100,
		LedgerCloseTime: 1_700_000_000,
	}
	proto := conv.ConvertGetTransactionResponseToProto(domain)

	got, err := conv.ConvertGetTransactionResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetTransactionResponseFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetTransactionResponseFromProto(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil")
}

func TestConvertGetSigningAccountResponse_RoundTrip(t *testing.T) {
	domain := stellartypes.GetSigningAccountResponse{AccountAddress: "GAAZI4TCR3TY5OJHCTJC2A4QSY6CJWJH5IAJTGKIN2ER7LBNVKOCCWN7"}
	proto := conv.ConvertGetSigningAccountResponseToProto(domain)
	require.Equal(t, domain.AccountAddress, proto.GetAccountAddress())

	got, err := conv.ConvertGetSigningAccountResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertGetSigningAccountResponseFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertGetSigningAccountResponseFromProto(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil")
}
