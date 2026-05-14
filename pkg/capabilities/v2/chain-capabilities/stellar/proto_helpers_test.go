package stellar_test

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"
	"github.com/stretchr/testify/require"

	stellarcap "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar/scval"
	conv "github.com/smartcontractkit/chainlink-common/pkg/chains/stellar"
	stellartypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

func TestConvertGetLatestLedgerResponseRoundtrip(t *testing.T) {
	t.Parallel()

	rawHash := []byte{0xde, 0xad, 0xbe, 0xef, 0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13,
		0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b}
	rawHeader := []byte{0xaa, 0xbb, 0xcc}
	rawMeta := []byte{0x11, 0x22, 0x33, 0x44}

	protoResp := &stellarcap.GetLatestLedgerResponse{
		Hash:              rawHash,
		ProtocolVersion:   22,
		Sequence:          987654,
		LedgerCloseTime:   1_700_000_000,
		LedgerHeaderXdr:   rawHeader,
		LedgerMetadataXdr: rawMeta,
	}

	domain, err := stellarcap.ConvertGetLatestLedgerResponseFromProto(protoResp)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(rawHash), domain.Hash)
	require.Equal(t, uint32(22), domain.ProtocolVersion)
	require.Equal(t, uint32(987654), domain.Sequence)
	require.Equal(t, int64(1_700_000_000), domain.LedgerCloseTime)
	require.Equal(t, base64.StdEncoding.EncodeToString(rawHeader), domain.LedgerHeaderXDR)
	require.Equal(t, base64.StdEncoding.EncodeToString(rawMeta), domain.LedgerMetadataXDR)

	// Round-trip back to proto.
	roundTripped, err := stellarcap.ConvertGetLatestLedgerResponseToProto(domain)
	require.NoError(t, err)
	require.Equal(t, rawHash, roundTripped.Hash)
	require.Equal(t, uint32(22), roundTripped.ProtocolVersion)
	require.Equal(t, uint32(987654), roundTripped.Sequence)
	require.Equal(t, int64(1_700_000_000), roundTripped.LedgerCloseTime)
	require.Equal(t, rawHeader, roundTripped.LedgerHeaderXdr)
	require.Equal(t, rawMeta, roundTripped.LedgerMetadataXdr)
}

func TestConvertGetLatestLedgerResponseFromProto_RejectsNil(t *testing.T) {
	t.Parallel()

	_, err := stellarcap.ConvertGetLatestLedgerResponseFromProto(nil)
	require.ErrorContains(t, err, "getLatestLedgerResponse is nil")
}

func TestConvertGetLatestLedgerResponseFromProto_EmptyFieldsReturnZeroValues(t *testing.T) {
	t.Parallel()

	domain, err := stellarcap.ConvertGetLatestLedgerResponseFromProto(&stellarcap.GetLatestLedgerResponse{})
	require.NoError(t, err)
	require.Equal(t, "", domain.Hash) // hex.EncodeToString(nil) == ""
	require.Equal(t, base64.StdEncoding.EncodeToString(nil), domain.LedgerHeaderXDR)
}

func TestConvertGetLatestLedgerResponseToProto_RejectsInvalidFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      stellartypes.GetLatestLedgerResponse
		wantErr string
	}{
		{"invalid hex hash", stellartypes.GetLatestLedgerResponse{Hash: "not-hex!"}, "invalid hex hash"},
		{"invalid base64 header", stellartypes.GetLatestLedgerResponse{LedgerHeaderXDR: "not-valid-base64!!!"}, "invalid base64 ledgerHeaderXDR"},
		{"invalid base64 metadata", stellartypes.GetLatestLedgerResponse{LedgerMetadataXDR: "not-valid-base64!!!"}, "invalid base64 ledgerMetadataXDR"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := stellarcap.ConvertGetLatestLedgerResponseToProto(tc.in)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestConvertReadContractRequestFromProto_Cap(t *testing.T) {
	t.Parallel()

	t.Run("nil request", func(t *testing.T) {
		_, err := stellarcap.ConvertReadContractRequestFromProto(nil)
		require.EqualError(t, err, "readContractRequest is nil")
	})

	t.Run("missing contract_id", func(t *testing.T) {
		_, err := stellarcap.ConvertReadContractRequestFromProto(&stellarcap.ReadContractRequest{Function: "fn"})
		require.EqualError(t, err, "contractID is required")
	})

	t.Run("missing function", func(t *testing.T) {
		_, err := stellarcap.ConvertReadContractRequestFromProto(&stellarcap.ReadContractRequest{ContractId: "C_X"})
		require.EqualError(t, err, "function is required")
	})

	t.Run("bad arg propagates index", func(t *testing.T) {
		_, err := stellarcap.ConvertReadContractRequestFromProto(&stellarcap.ReadContractRequest{
			ContractId: "C_X",
			Function:   "fn",
			Args:       []*scval.ScVal{nil},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "args[0]")
	})

	t.Run("round-trip", func(t *testing.T) {
		u := uint32(42)
		sym := "transfer"
		p := &stellarcap.ReadContractRequest{
			ContractId: "C_TESTCONTRACT",
			Function:   "transfer",
			Args: []*scval.ScVal{
				{Value: &scval.ScVal_U32{U32: u}},
				{Value: &scval.ScVal_Sym{Sym: sym}},
			},
			LedgerSequence: 7,
		}
		got, err := stellarcap.ConvertReadContractRequestFromProto(p)
		require.NoError(t, err)
		require.Equal(t, "C_TESTCONTRACT", got.ContractID)
		require.Equal(t, "transfer", got.Function)
		require.Equal(t, uint32(7), got.LedgerSequence)
		require.Len(t, got.Args, 2)
		require.Equal(t, xdr.ScValTypeScvU32, got.Args[0].Type)
		require.Equal(t, xdr.Uint32(42), *got.Args[0].U32)
		require.Equal(t, xdr.ScValTypeScvSymbol, got.Args[1].Type)
		require.Equal(t, xdr.ScSymbol("transfer"), *got.Args[1].Sym)
	})
}

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

func TestXdrScValToProto_InvalidArgs(t *testing.T) {
	tests := []struct {
		name    string
		sv      xdr.ScVal
		wantErr string
	}{
		{
			"scError contractCode nil",
			xdr.ScVal{Type: xdr.ScValTypeScvError, Error: &xdr.ScError{Type: xdr.ScErrorTypeSceContract}},
			"scError.contractCode: nil",
		},
		{
			"scError code nil",
			xdr.ScVal{Type: xdr.ScValTypeScvError, Error: &xdr.ScError{Type: xdr.ScErrorTypeSceWasmVm}},
			"nil code",
		},
		{
			"scAddress account nil accountId",
			xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeAccount}},
			"scAddress.account",
		},
		{
			"scAddress contract nil contractId",
			xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeContract}},
			"scAddress.contract: nil contractId",
		},
		{
			"scAddress muxed nil",
			xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeMuxedAccount}},
			"scAddress.muxed: nil",
		},
		{
			"scAddress claimableBalance nil",
			xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeClaimableBalance}},
			"scAddress.claimableBalance: nil",
		},
		{
			"scAddress liquidityPool nil poolId",
			xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeLiquidityPool}},
			"scAddress.liquidityPool: nil poolId",
		},
		{
			"scAddress unsupported type",
			xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xdr.ScAddress{Type: xdr.ScAddressType(999)}},
			"unsupported ScAddress type",
		},
		{
			"contractExecutable wasm nil wasmHash",
			xdr.ScVal{
				Type: xdr.ScValTypeScvContractInstance,
				Instance: &xdr.ScContractInstance{
					Executable: xdr.ContractExecutable{Type: xdr.ContractExecutableTypeContractExecutableWasm},
				},
			},
			"contractExecutable.wasm: nil wasmHash",
		},
		{
			"contractExecutable unsupported type",
			xdr.ScVal{
				Type: xdr.ScValTypeScvContractInstance,
				Instance: &xdr.ScContractInstance{
					Executable: xdr.ContractExecutable{Type: xdr.ContractExecutableType(999)},
				},
			},
			"unsupported ContractExecutable type",
		},
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

// ---- Proto→XDR nil/invalid cases (table-driven) ----------------------------

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
		Args:       []*scval.ScVal{nil},
	}
	_, err := conv.ConvertReadContractRequestFromProto(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "args[0]")
}

func TestProtoScVal_AccountId_WrongLength(t *testing.T) {
	p := &conv.ReadContractRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args: []*scval.ScVal{
			{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
				Address: &scval.ScAddress_AccountId{AccountId: []byte{0x01, 0x02}},
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
		Args: []*scval.ScVal{
			{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
				Address: &scval.ScAddress_ContractId{ContractId: []byte("short")},
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
		Args: []*scval.ScVal{
			{Value: &scval.ScVal_ContractInstance{ContractInstance: &scval.ScContractInstance{
				Executable: &scval.ContractExecutable{
					Type: &scval.ContractExecutable_WasmHash{WasmHash: []byte("tooshort")},
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
	leaf := &scval.ScVal{Value: &scval.ScVal_U32{U32: u}}
	cur := leaf
	for i := 0; i < 66; i++ {
		cur = &scval.ScVal{Value: &scval.ScVal_Vec{Vec: &scval.ScVec{Values: []*scval.ScVal{cur}}}}
	}
	p := &conv.ReadContractRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args:       []*scval.ScVal{cur},
	}
	_, err := conv.ConvertReadContractRequestFromProto(p)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "nesting exceeds maximum depth"), "unexpected error: %v", err)
}

func protoScValArg(val *scval.ScVal) *conv.ReadContractRequest {
	return &conv.ReadContractRequest{
		ContractId: "C_X",
		Function:   "fn",
		Args:       []*scval.ScVal{val},
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
			protoScValArg(&scval.ScVal{Value: &scval.ScVal_U128{U128: nil}}),
			"scvU128: nil",
		},
		{
			"i128 nil",
			protoScValArg(&scval.ScVal{Value: &scval.ScVal_I128{I128: nil}}),
			"scvI128: nil",
		},
		{
			"u256 nil",
			protoScValArg(&scval.ScVal{Value: &scval.ScVal_U256{U256: nil}}),
			"scvU256: nil",
		},
		{
			"i256 nil",
			protoScValArg(&scval.ScVal{Value: &scval.ScVal_I256{I256: nil}}),
			"scvI256: nil",
		},
		{
			"vec nil",
			protoScValArg(&scval.ScVal{Value: &scval.ScVal_Vec{Vec: nil}}),
			"scvVec: nil",
		},
		{
			"map nil",
			protoScValArg(&scval.ScVal{Value: &scval.ScVal_Map{Map: nil}}),
			"scvMap: nil",
		},
		{
			"nonceKey nil",
			protoScValArg(&scval.ScVal{Value: &scval.ScVal_NonceKey{NonceKey: nil}}),
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
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(&scval.ScVal{}))
	require.ErrorContains(t, err, "unsupported proto ScVal type")
}

func TestProtoScAddress_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_Address{Address: nil}},
	))
	require.ErrorContains(t, err, "proto ScAddress is nil")
}

func TestProtoScAddress_MuxedAccount_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
			Address: &scval.ScAddress_MuxedAccount{MuxedAccount: nil},
		}}},
	))
	require.ErrorContains(t, err, "muxedAccount: nil")
}

func TestProtoScAddress_MuxedAccount_WrongEd25519Size(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
			Address: &scval.ScAddress_MuxedAccount{MuxedAccount: &scval.MuxedEd25519Account{
				Id:      1,
				Ed25519: []byte{0x01, 0x02}, // not 32 bytes
			}},
		}}},
	))
	require.ErrorContains(t, err, "muxedAccount.ed25519 must be 32 bytes")
}

func TestProtoScAddress_ClaimableBalanceId_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
			Address: &scval.ScAddress_ClaimableBalanceId{ClaimableBalanceId: nil},
		}}},
	))
	require.ErrorContains(t, err, "claimableBalanceId: nil")
}

func TestProtoScAddress_ClaimableBalanceId_WrongV0Size(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
			Address: &scval.ScAddress_ClaimableBalanceId{ClaimableBalanceId: &scval.ClaimableBalanceId{
				V0: []byte{0x01, 0x02}, // not 32 bytes
			}},
		}}},
	))
	require.ErrorContains(t, err, "claimableBalanceId.v0 must be 32 bytes")
}

func TestProtoScAddress_LiquidityPoolId_WrongSize(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
			Address: &scval.ScAddress_LiquidityPoolId{LiquidityPoolId: []byte("short")},
		}}},
	))
	require.ErrorContains(t, err, "liquidityPoolId must be 32 bytes")
}

func TestProtoScAddress_UnsupportedOneof(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{Address: nil}}},
	))
	require.ErrorContains(t, err, "unsupported proto ScAddress type")
}

func TestProtoContractExecutable_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: &scval.ScContractInstance{
			Executable: nil,
		}}},
	))
	require.ErrorContains(t, err, "proto ContractExecutable is nil")
}

func TestProtoContractExecutable_UnsupportedOneof(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: &scval.ScContractInstance{
			Executable: &scval.ContractExecutable{Type: nil},
		}}},
	))
	require.ErrorContains(t, err, "unsupported proto ContractExecutable type")
}

func TestProtoScContractInstance_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: nil}},
	))
	require.ErrorContains(t, err, "proto ScContractInstance is nil")
}

func TestProtoScError_Nil(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_Error{Error: nil}},
	))
	require.ErrorContains(t, err, "proto ScError is nil")
}

func TestProtoScError_UnsupportedOneof(t *testing.T) {
	_, err := conv.ConvertReadContractRequestFromProto(protoScValArg(
		&scval.ScVal{Value: &scval.ScVal_Error{Error: &scval.ScError{CodeOrContract: nil}}},
	))
	require.ErrorContains(t, err, "unsupported ScError oneof")
}
