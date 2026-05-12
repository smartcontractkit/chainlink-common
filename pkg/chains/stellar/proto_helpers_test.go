package stellar_test

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"
	"github.com/stretchr/testify/require"

	v1alpha "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar/scval"
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
	require.EqualError(t, err, "ReadContractRequest is nil")
}

func TestConvertReadContractRequestFromProto_MissingContractID(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(&conv.ReadContractRequest{Function: "fn"})
	require.EqualError(t, err, "contract_id is required")
}

func TestConvertReadContractRequestFromProto_MissingFunction(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(&conv.ReadContractRequest{ContractId: "C123"})
	require.EqualError(t, err, "function is required")
}

func TestConvertReadContractResponse_RoundTrip_WithResult(t *testing.T) {
	i64 := xdr.Int64(-99)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvI64, I64: &i64}
	domain := stellartypes.ReadContractResponse{
		Result:         &sv,
		LedgerSequence: 55,
		Error:          "",
	}

	proto, err := conv.ConvertReadContractResponseToProto(domain)
	require.NoError(t, err)
	require.NotNil(t, proto.GetResult())

	got, err := conv.ConvertReadContractResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertReadContractResponse_RoundTrip_WithError(t *testing.T) {
	domain := stellartypes.ReadContractResponse{
		Result:         nil,
		LedgerSequence: 10,
		Error:          "contract panicked",
	}

	proto, err := conv.ConvertReadContractResponseToProto(domain)
	require.NoError(t, err)
	require.Nil(t, proto.GetResult())

	got, err := conv.ConvertReadContractResponseFromProto(proto)
	require.NoError(t, err)
	require.Equal(t, domain, got)
}

func TestConvertReadContractResponseFromProto_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractResponseFromProto(nil)
	require.EqualError(t, err, "readContractResponse is nil")
}

func scValRoundTrip(t *testing.T, sv xdr.ScVal) xdr.ScVal {
	t.Helper()
	req := stellartypes.ReadContractRequest{
		ContractID: "C_TESTCONTRACT",
		Function:   "fn",
		Args:       []xdr.ScVal{sv},
	}
	proto, err := conv.ConvertReadContractRequestToProto(req)
	require.NoError(t, err)
	got, err := conv.ConvertReadContractRequestFromProto(proto)
	require.NoError(t, err)
	require.Len(t, got.Args, 1)
	return got.Args[0]
}

func TestScVal_Bool(t *testing.T) {
	b := true
	sv := xdr.ScVal{Type: xdr.ScValTypeScvBool, B: &b}
	got := scValRoundTrip(t, sv)
	require.Equal(t, sv, got)
}

func TestScVal_Void(t *testing.T) {
	sv := xdr.ScVal{Type: xdr.ScValTypeScvVoid}
	got := scValRoundTrip(t, sv)
	require.Equal(t, xdr.ScValTypeScvVoid, got.Type)
}

func TestScVal_Error_ContractCode(t *testing.T) {
	cc := xdr.Uint32(42)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvError, Error: &xdr.ScError{
		Type:         xdr.ScErrorTypeSceContract,
		ContractCode: &cc,
	}}
	got := scValRoundTrip(t, sv)
	require.Equal(t, sv, got)
}

func TestScVal_Error_Code(t *testing.T) {
	code := xdr.ScErrorCodeScecArithDomain
	sv := xdr.ScVal{Type: xdr.ScValTypeScvError, Error: &xdr.ScError{
		Type: xdr.ScErrorTypeSceWasmVm,
		Code: &code,
	}}
	got := scValRoundTrip(t, sv)
	require.Equal(t, sv, got)
}

func TestScVal_U32(t *testing.T) {
	u := xdr.Uint32(0xDEAD)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &u}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_I32(t *testing.T) {
	i := xdr.Int32(-1234)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvI32, I32: &i}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_U64(t *testing.T) {
	u := xdr.Uint64(1 << 40)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvU64, U64: &u}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_I64(t *testing.T) {
	i := xdr.Int64(-1 << 40)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvI64, I64: &i}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Timepoint(t *testing.T) {
	tp := xdr.TimePoint(1_700_000_000)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvTimepoint, Timepoint: &tp}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Duration(t *testing.T) {
	d := xdr.Duration(3600)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvDuration, Duration: &d}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_U128(t *testing.T) {
	u := xdr.UInt128Parts{Hi: xdr.Uint64(0xAAAA), Lo: xdr.Uint64(0xBBBB)}
	sv := xdr.ScVal{Type: xdr.ScValTypeScvU128, U128: &u}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_I128(t *testing.T) {
	i := xdr.Int128Parts{Hi: xdr.Int64(-7), Lo: xdr.Uint64(999)}
	sv := xdr.ScVal{Type: xdr.ScValTypeScvI128, I128: &i}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_U256(t *testing.T) {
	u := xdr.UInt256Parts{
		HiHi: xdr.Uint64(1), HiLo: xdr.Uint64(2),
		LoHi: xdr.Uint64(3), LoLo: xdr.Uint64(4),
	}
	sv := xdr.ScVal{Type: xdr.ScValTypeScvU256, U256: &u}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_I256(t *testing.T) {
	i := xdr.Int256Parts{
		HiHi: xdr.Int64(-1), HiLo: xdr.Uint64(2),
		LoHi: xdr.Uint64(3), LoLo: xdr.Uint64(4),
	}
	sv := xdr.ScVal{Type: xdr.ScValTypeScvI256, I256: &i}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Bytes(t *testing.T) {
	xb := xdr.ScBytes([]byte{0x01, 0x02, 0x03})
	sv := xdr.ScVal{Type: xdr.ScValTypeScvBytes, Bytes: &xb}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_String(t *testing.T) {
	s := xdr.ScString("hello world")
	sv := xdr.ScVal{Type: xdr.ScValTypeScvString, Str: &s}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Symbol(t *testing.T) {
	sym := xdr.ScSymbol("transfer")
	sv := xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &sym}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Vec(t *testing.T) {
	u := xdr.Uint32(1)
	inner := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &u}
	vec := xdr.ScVec{inner}
	vecp := &vec
	sv := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &vecp}
	got := scValRoundTrip(t, sv)
	require.Equal(t, sv, got)
}

func TestScVal_Map(t *testing.T) {
	sym := xdr.ScSymbol("key")
	u := xdr.Uint32(99)
	xmap := xdr.ScMap{
		{Key: xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &sym}, Val: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &u}},
	}
	xmapp := &xmap
	sv := xdr.ScVal{Type: xdr.ScValTypeScvMap, Map: &xmapp}
	got := scValRoundTrip(t, sv)
	require.Equal(t, sv, got)
}

func TestScVal_Address_Account(t *testing.T) {
	var b [32]byte
	for i := range b {
		b[i] = 0x01
	}
	ed := b
	ed256 := xdr.Uint256(ed)
	aid := xdr.AccountId(xdr.PublicKey{
		Type:    xdr.PublicKeyTypePublicKeyTypeEd25519,
		Ed25519: &ed256,
	})
	sv := xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{
		Type:      xdr.ScAddressTypeScAddressTypeAccount,
		AccountId: &aid,
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Address_Contract(t *testing.T) {
	var b [32]byte
	for i := range b {
		b[i] = 0x02
	}
	cid := xdr.ContractId(b)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{
		Type:       xdr.ScAddressTypeScAddressTypeContract,
		ContractId: &cid,
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Address_MuxedAccount(t *testing.T) {
	var b [32]byte
	for i := range b {
		b[i] = 0x03
	}
	ed := xdr.Uint256(b)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{
		Type: xdr.ScAddressTypeScAddressTypeMuxedAccount,
		MuxedAccount: &xdr.MuxedEd25519Account{
			Id:      xdr.Uint64(777),
			Ed25519: ed,
		},
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Address_ClaimableBalance(t *testing.T) {
	var b [32]byte
	for i := range b {
		b[i] = 0x04
	}
	h := xdr.Hash(b)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{
		Type: xdr.ScAddressTypeScAddressTypeClaimableBalance,
		ClaimableBalanceId: &xdr.ClaimableBalanceId{
			Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
			V0:   &h,
		},
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Address_LiquidityPool(t *testing.T) {
	var b [32]byte
	for i := range b {
		b[i] = 0x05
	}
	pid := xdr.PoolId(b)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{
		Type:            xdr.ScAddressTypeScAddressTypeLiquidityPool,
		LiquidityPoolId: &pid,
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_ContractInstance_Wasm(t *testing.T) {
	var b [32]byte
	for i := range b {
		b[i] = 0x06
	}
	wh := xdr.Hash(b)
	sv := xdr.ScVal{Type: xdr.ScValTypeScvContractInstance, Instance: &xdr.ScContractInstance{
		Executable: xdr.ContractExecutable{
			Type:     xdr.ContractExecutableTypeContractExecutableWasm,
			WasmHash: &wh,
		},
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_ContractInstance_StellarAsset(t *testing.T) {
	sv := xdr.ScVal{Type: xdr.ScValTypeScvContractInstance, Instance: &xdr.ScContractInstance{
		Executable: xdr.ContractExecutable{Type: xdr.ContractExecutableTypeContractExecutableStellarAsset},
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_ContractInstance_WithStorage(t *testing.T) {
	var b [32]byte
	for i := range b {
		b[i] = 0x07
	}
	wh := xdr.Hash(b)
	sym := xdr.ScSymbol("slot")
	u := xdr.Uint32(1)
	storage := xdr.ScMap{
		{Key: xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &sym}, Val: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &u}},
	}
	sv := xdr.ScVal{Type: xdr.ScValTypeScvContractInstance, Instance: &xdr.ScContractInstance{
		Executable: xdr.ContractExecutable{
			Type:     xdr.ContractExecutableTypeContractExecutableWasm,
			WasmHash: &wh,
		},
		Storage: &storage,
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_LedgerKeyContractInstance(t *testing.T) {
	sv := xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyContractInstance}
	got := scValRoundTrip(t, sv)
	require.Equal(t, xdr.ScValTypeScvLedgerKeyContractInstance, got.Type)
}

func TestScVal_LedgerKeyNonce(t *testing.T) {
	sv := xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyNonce, NonceKey: &xdr.ScNonceKey{Nonce: xdr.Int64(12345)}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_NestedVecMap(t *testing.T) {
	// Vec containing a Map: [{sym:"x" -> u32:1}]
	sym := xdr.ScSymbol("x")
	u := xdr.Uint32(1)
	innerMap := xdr.ScMap{{
		Key: xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &sym},
		Val: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &u},
	}}
	innerMapP := &innerMap
	mapVal := xdr.ScVal{Type: xdr.ScValTypeScvMap, Map: &innerMapP}
	vec := xdr.ScVec{mapVal}
	vecp := &vec
	sv := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &vecp}

	got := scValRoundTrip(t, sv)
	require.Equal(t, sv, got)
}

func TestScVal_ExceedsMaxDepth(t *testing.T) {
	// Build a ScVal nested 66 levels deep via ReadContract args conversion.
	// We construct the nesting bottom-up.
	u := xdr.Uint32(0)
	leaf := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &u}
	cur := leaf
	for i := 0; i < 66; i++ {
		vec := xdr.ScVec{cur}
		vecp := &vec
		cur = xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &vecp}
	}

	req := stellartypes.ReadContractRequest{
		ContractID: "C_DEEP",
		Function:   "fn",
		Args:       []xdr.ScVal{cur},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nesting exceeds maximum depth")
}

func TestProtoScVal_NilValue(t *testing.T) {
	// A proto ReadContractRequest carrying a nil ScVal should fail gracefully.
	p := &conv.ReadContractRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args:       []*v1alpha.ScVal{nil},
	}
	_, err := conv.ConvertReadContractRequestFromProto(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "args[0]")
}

func TestProtoScVal_AccountId_WrongLength(t *testing.T) {
	p := &conv.ReadContractRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args: []*v1alpha.ScVal{
			{Value: &v1alpha.ScVal_Address{Address: &v1alpha.ScAddress{
				Address: &v1alpha.ScAddress_AccountId{AccountId: []byte{0x01, 0x02}},
			}}},
		},
	}
	_, err := conv.ConvertReadContractRequestFromProto(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "accountId must be 32 bytes")
}

func TestProtoScVal_ContractId_WrongLength(t *testing.T) {
	p := &conv.ReadContractRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args: []*v1alpha.ScVal{
			{Value: &v1alpha.ScVal_Address{Address: &v1alpha.ScAddress{
				Address: &v1alpha.ScAddress_ContractId{ContractId: []byte("short")},
			}}},
		},
	}
	_, err := conv.ConvertReadContractRequestFromProto(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "contractId must be 32 bytes")
}

func TestProtoScVal_WasmHash_WrongLength(t *testing.T) {
	p := &conv.ReadContractRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args: []*v1alpha.ScVal{
			{Value: &v1alpha.ScVal_ContractInstance{ContractInstance: &v1alpha.ScContractInstance{
				Executable: &v1alpha.ContractExecutable{
					Type: &v1alpha.ContractExecutable_WasmHash{WasmHash: []byte("tooshort")},
				},
			}}},
		},
	}
	_, err := conv.ConvertReadContractRequestFromProto(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "wasmHash must be 32 bytes")
}

func TestProtoScVal_ExceedsMaxDepth(t *testing.T) {
	// Build proto ScVal nested 66 levels deep.
	u := uint32(0)
	leaf := &v1alpha.ScVal{Value: &v1alpha.ScVal_U32{U32: u}}
	cur := leaf
	for i := 0; i < 66; i++ {
		cur = &v1alpha.ScVal{Value: &v1alpha.ScVal_Vec{Vec: &v1alpha.ScVec{Values: []*v1alpha.ScVal{cur}}}}
	}
	p := &conv.ReadContractRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args:       []*v1alpha.ScVal{cur},
	}
	_, err := conv.ConvertReadContractRequestFromProto(p)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "nesting exceeds maximum depth"), "unexpected error: %v", err)
}

// ---- ConvertReadContractRequestToProto validation ---------------------------

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

func TestConvertGetLatestLedgerResponseToProto_InvalidHash(t *testing.T) {
	_, err := conv.ConvertGetLatestLedgerResponseToProto(stellartypes.GetLatestLedgerResponse{
		Hash: "not-hex!",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid hex hash")
}

func TestConvertGetLatestLedgerResponseToProto_InvalidHeaderXDR(t *testing.T) {
	_, err := conv.ConvertGetLatestLedgerResponseToProto(stellartypes.GetLatestLedgerResponse{
		LedgerHeaderXDR: "!!!bad!!!",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid ledger header xdr")
}

func TestConvertGetLatestLedgerResponseToProto_InvalidMetadataXDR(t *testing.T) {
	_, err := conv.ConvertGetLatestLedgerResponseToProto(stellartypes.GetLatestLedgerResponse{
		LedgerMetadataXDR: "!!!bad!!!",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid ledger metadata xdr")
}

// ---- ConvertReadContractResponse result error propagation -------------------

func TestConvertReadContractResponseToProto_BadResult(t *testing.T) {
	sv := xdr.ScVal{Type: xdr.ScValTypeScvBool} // B is nil → error
	_, err := conv.ConvertReadContractResponseToProto(stellartypes.ReadContractResponse{Result: &sv})
	require.Error(t, err)
	require.Contains(t, err.Error(), "result")
}

func TestConvertReadContractResponseFromProto_BadResult(t *testing.T) {
	_, err := conv.ConvertReadContractResponseFromProto(&conv.ReadContractResponse{
		Result: &v1alpha.ScVal{Value: &v1alpha.ScVal_U128{U128: nil}},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "result")
}

// ---- XDR→proto nil arm pointer cases (table-driven) ------------------------

func TestXdrScValToProto_NilArmPointers(t *testing.T) {
	tests := []struct {
		name    string
		sv      xdr.ScVal
		wantErr string
	}{
		{"bool nil", xdr.ScVal{Type: xdr.ScValTypeScvBool}, "scvBool: nil"},
		{"error nil", xdr.ScVal{Type: xdr.ScValTypeScvError}, "scvError: nil"},
		{"u32 nil", xdr.ScVal{Type: xdr.ScValTypeScvU32}, "scvU32: nil"},
		{"i32 nil", xdr.ScVal{Type: xdr.ScValTypeScvI32}, "scvI32: nil"},
		{"u64 nil", xdr.ScVal{Type: xdr.ScValTypeScvU64}, "scvU64: nil"},
		{"i64 nil", xdr.ScVal{Type: xdr.ScValTypeScvI64}, "scvI64: nil"},
		{"timepoint nil", xdr.ScVal{Type: xdr.ScValTypeScvTimepoint}, "scvTimepoint: nil"},
		{"duration nil", xdr.ScVal{Type: xdr.ScValTypeScvDuration}, "scvDuration: nil"},
		{"u128 nil", xdr.ScVal{Type: xdr.ScValTypeScvU128}, "scvU128: nil"},
		{"i128 nil", xdr.ScVal{Type: xdr.ScValTypeScvI128}, "scvI128: nil"},
		{"u256 nil", xdr.ScVal{Type: xdr.ScValTypeScvU256}, "scvU256: nil"},
		{"i256 nil", xdr.ScVal{Type: xdr.ScValTypeScvI256}, "scvI256: nil"},
		{"bytes nil", xdr.ScVal{Type: xdr.ScValTypeScvBytes}, "scvBytes: nil"},
		{"string nil", xdr.ScVal{Type: xdr.ScValTypeScvString}, "scvString: nil"},
		{"symbol nil", xdr.ScVal{Type: xdr.ScValTypeScvSymbol}, "scvSymbol: nil"},
		{"address nil", xdr.ScVal{Type: xdr.ScValTypeScvAddress}, "scvAddress: nil"},
		{"contractInstance nil", xdr.ScVal{Type: xdr.ScValTypeScvContractInstance}, "scvContractInstance: nil"},
		{"nonceKey nil", xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyNonce}, "scvLedgerKeyNonce: nil"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := stellartypes.ReadContractRequest{
				ContractID: "C_X", Function: "fn", Args: []xdr.ScVal{tc.sv},
			}
			_, err := conv.ConvertReadContractRequestToProto(req)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestXdrScValToProto_UnsupportedType(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{Type: xdr.ScValType(999)}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "unsupported ScVal type")
}

func TestXdrScError_NilContractCode(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type:  xdr.ScValTypeScvError,
			Error: &xdr.ScError{Type: xdr.ScErrorTypeSceContract}, // ContractCode nil
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "scError.contractCode: nil")
}

func TestXdrScError_NilCode(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type:  xdr.ScValTypeScvError,
			Error: &xdr.ScError{Type: xdr.ScErrorTypeSceWasmVm}, // Code nil
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "nil code")
}

func TestXdrScAddress_NilAccountId(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type:    xdr.ScValTypeScvAddress,
			Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeAccount}, // AccountId nil
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "scAddress.account")
}

func TestXdrScAddress_NilContractId(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type:    xdr.ScValTypeScvAddress,
			Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeContract}, // ContractId nil
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "scAddress.contract: nil contractId")
}

func TestXdrScAddress_NilMuxedAccount(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type:    xdr.ScValTypeScvAddress,
			Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeMuxedAccount}, // MuxedAccount nil
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "scAddress.muxed: nil")
}

func TestXdrScAddress_NilClaimableBalance(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type:    xdr.ScValTypeScvAddress,
			Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeClaimableBalance}, // ClaimableBalanceId nil
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "scAddress.claimableBalance: nil")
}

func TestXdrScAddress_NilLiquidityPool(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type:    xdr.ScValTypeScvAddress,
			Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeLiquidityPool}, // LiquidityPoolId nil
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "scAddress.liquidityPool: nil poolId")
}

func TestXdrScAddress_UnsupportedType(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type:    xdr.ScValTypeScvAddress,
			Address: &xdr.ScAddress{Type: xdr.ScAddressType(999)},
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "unsupported ScAddress type")
}

func TestXdrContractExecutable_NilWasmHash(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type: xdr.ScValTypeScvContractInstance,
			Instance: &xdr.ScContractInstance{
				Executable: xdr.ContractExecutable{Type: xdr.ContractExecutableTypeContractExecutableWasm}, // WasmHash nil
			},
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "contractExecutable.wasm: nil wasmHash")
}

func TestXdrContractExecutable_UnsupportedType(t *testing.T) {
	req := stellartypes.ReadContractRequest{
		ContractID: "C_X", Function: "fn",
		Args: []xdr.ScVal{{
			Type: xdr.ScValTypeScvContractInstance,
			Instance: &xdr.ScContractInstance{
				Executable: xdr.ContractExecutable{Type: xdr.ContractExecutableType(999)},
			},
		}},
	}
	_, err := conv.ConvertReadContractRequestToProto(req)
	require.ErrorContains(t, err, "unsupported ContractExecutable type")
}

// ---- Proto→XDR nil/invalid cases (table-driven) ----------------------------

func protoScValArg(val *v1alpha.ScVal) *conv.ReadContractRequest {
	return &conv.ReadContractRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args:       []*v1alpha.ScVal{val},
	}
}

func TestProtoScVal_NilInnerFields(t *testing.T) {
	tests := []struct {
		name    string
		req     *conv.ReadContractRequest
		wantErr string
	}{
		{
			"u128 nil",
			protoScValArg(&v1alpha.ScVal{Value: &v1alpha.ScVal_U128{U128: nil}}),
			"scvU128: nil",
		},
		{
			"i128 nil",
			protoScValArg(&v1alpha.ScVal{Value: &v1alpha.ScVal_I128{I128: nil}}),
			"scvI128: nil",
		},
		{
			"u256 nil",
			protoScValArg(&v1alpha.ScVal{Value: &v1alpha.ScVal_U256{U256: nil}}),
			"scvU256: nil",
		},
		{
			"i256 nil",
			protoScValArg(&v1alpha.ScVal{Value: &v1alpha.ScVal_I256{I256: nil}}),
			"scvI256: nil",
		},
		{
			"vec nil",
			protoScValArg(&v1alpha.ScVal{Value: &v1alpha.ScVal_Vec{Vec: nil}}),
			"scvVec: nil",
		},
		{
			"map nil",
			protoScValArg(&v1alpha.ScVal{Value: &v1alpha.ScVal_Map{Map: nil}}),
			"scvMap: nil",
		},
		{
			"nonceKey nil",
			protoScValArg(&v1alpha.ScVal{Value: &v1alpha.ScVal_NonceKey{NonceKey: nil}}),
			"scvLedgerKeyNonce: nil",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := conv.ConvertReadContractRequestFromProto(tc.req)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestProtoScVal_UnsupportedOneof(t *testing.T) {
	// A zero-value ScVal (nil Value oneof) hits the default case.
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(&v1alpha.ScVal{}))
	require.ErrorContains(t, err, "unsupported proto ScVal type")
}

func TestProtoScAddress_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_Address{Address: nil}},
	))
	require.ErrorContains(t, err, "proto ScAddress is nil")
}

func TestProtoScAddress_MuxedAccount_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_Address{Address: &v1alpha.ScAddress{
			Address: &v1alpha.ScAddress_MuxedAccount{MuxedAccount: nil},
		}}},
	))
	require.ErrorContains(t, err, "muxedAccount: nil")
}

func TestProtoScAddress_MuxedAccount_WrongEd25519Size(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_Address{Address: &v1alpha.ScAddress{
			Address: &v1alpha.ScAddress_MuxedAccount{MuxedAccount: &v1alpha.MuxedEd25519Account{
				Id:      1,
				Ed25519: []byte{0x01, 0x02}, // not 32 bytes
			}},
		}}},
	))
	require.ErrorContains(t, err, "muxedAccount.ed25519 must be 32 bytes")
}

func TestProtoScAddress_ClaimableBalanceId_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_Address{Address: &v1alpha.ScAddress{
			Address: &v1alpha.ScAddress_ClaimableBalanceId{ClaimableBalanceId: nil},
		}}},
	))
	require.ErrorContains(t, err, "claimableBalanceId: nil")
}

func TestProtoScAddress_ClaimableBalanceId_WrongV0Size(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_Address{Address: &v1alpha.ScAddress{
			Address: &v1alpha.ScAddress_ClaimableBalanceId{ClaimableBalanceId: &v1alpha.ClaimableBalanceId{
				V0: []byte{0x01, 0x02}, // not 32 bytes
			}},
		}}},
	))
	require.ErrorContains(t, err, "claimableBalanceId.v0 must be 32 bytes")
}

func TestProtoScAddress_LiquidityPoolId_WrongSize(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_Address{Address: &v1alpha.ScAddress{
			Address: &v1alpha.ScAddress_LiquidityPoolId{LiquidityPoolId: []byte("short")},
		}}},
	))
	require.ErrorContains(t, err, "liquidityPoolId must be 32 bytes")
}

func TestProtoScAddress_UnsupportedOneof(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_Address{Address: &v1alpha.ScAddress{Address: nil}}},
	))
	require.ErrorContains(t, err, "unsupported proto ScAddress type")
}

func TestProtoContractExecutable_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_ContractInstance{ContractInstance: &v1alpha.ScContractInstance{
			Executable: nil,
		}}},
	))
	require.ErrorContains(t, err, "proto ContractExecutable is nil")
}

func TestProtoContractExecutable_UnsupportedOneof(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_ContractInstance{ContractInstance: &v1alpha.ScContractInstance{
			Executable: &v1alpha.ContractExecutable{Type: nil},
		}}},
	))
	require.ErrorContains(t, err, "unsupported proto ContractExecutable type")
}

func TestProtoScContractInstance_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_ContractInstance{ContractInstance: nil}},
	))
	require.ErrorContains(t, err, "proto ScContractInstance is nil")
}

func TestProtoScError_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_Error{Error: nil}},
	))
	require.ErrorContains(t, err, "proto ScError is nil")
}

func TestProtoScError_UnsupportedOneof(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&v1alpha.ScVal{Value: &v1alpha.ScVal_Error{Error: &v1alpha.ScError{CodeOrContract: nil}}},
	))
	require.ErrorContains(t, err, "unsupported ScError oneof")
}
