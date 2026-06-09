package stellar_test

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	stellarcap "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar/scval"
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
		}
		got, err := stellarcap.ConvertReadContractRequestFromProto(p)
		require.NoError(t, err)
		require.Equal(t, "C_TESTCONTRACT", got.ContractID)
		require.Equal(t, "transfer", got.Function)
		require.Len(t, got.Args, 2)
		require.Equal(t, stellartypes.ScValTypeU32, got.Args[0].Type)
		require.Equal(t, uint32(42), *got.Args[0].U32)
		require.Equal(t, stellartypes.ScValTypeSymbol, got.Args[1].Type)
		require.Equal(t, "transfer", *got.Args[1].Symbol)
	})
}

func TestScValToProto_NilArmPointers(t *testing.T) {
	tests := []struct {
		name    string
		sv      stellartypes.ScVal
		wantErr string
	}{
		{"bool nil", stellartypes.ScVal{Type: stellartypes.ScValTypeBool}, "scvBool: nil"},
		{"void nil", stellartypes.ScVal{Type: stellartypes.ScValTypeVoid}, "scvVoid: nil"},
		{"error nil", stellartypes.ScVal{Type: stellartypes.ScValTypeError}, "scvError: nil"},
		{"u32 nil", stellartypes.ScVal{Type: stellartypes.ScValTypeU32}, "scvU32: nil"},
		{"i32 nil", stellartypes.ScVal{Type: stellartypes.ScValTypeI32}, "scvI32: nil"},
		{"u64 nil", stellartypes.ScVal{Type: stellartypes.ScValTypeU64}, "scvU64: nil"},
		{"i64 nil", stellartypes.ScVal{Type: stellartypes.ScValTypeI64}, "scvI64: nil"},
		{"timepoint nil", stellartypes.ScVal{Type: stellartypes.ScValTypeTimepoint}, "scvTimepoint: nil"},
		{"duration nil", stellartypes.ScVal{Type: stellartypes.ScValTypeDuration}, "scvDuration: nil"},
		{"u128 nil", stellartypes.ScVal{Type: stellartypes.ScValTypeU128}, "scvU128: nil"},
		{"i128 nil", stellartypes.ScVal{Type: stellartypes.ScValTypeI128}, "scvI128: nil"},
		{"u256 nil", stellartypes.ScVal{Type: stellartypes.ScValTypeU256}, "scvU256: nil"},
		{"i256 nil", stellartypes.ScVal{Type: stellartypes.ScValTypeI256}, "scvI256: nil"},
		{"bytes nil", stellartypes.ScVal{Type: stellartypes.ScValTypeBytes}, "scvBytes: nil"},
		{"string nil", stellartypes.ScVal{Type: stellartypes.ScValTypeString}, "scvString: nil"},
		{"symbol nil", stellartypes.ScVal{Type: stellartypes.ScValTypeSymbol}, "scvSymbol: nil"},
		{"vec nil", stellartypes.ScVal{Type: stellartypes.ScValTypeVec}, "scvVec: nil"},
		{"map nil", stellartypes.ScVal{Type: stellartypes.ScValTypeMap}, "scvMap: nil"},
		{"address nil", stellartypes.ScVal{Type: stellartypes.ScValTypeAddress}, "scvAddress: nil"},
		{"contractInstance nil", stellartypes.ScVal{Type: stellartypes.ScValTypeContractInstance}, "scvContractInstance: nil"},
		{"ledgerKeyContractInstance nil", stellartypes.ScVal{Type: stellartypes.ScValTypeLedgerKeyContractInstance}, "scvLedgerKeyContractInstance: nil"},
		{"nonceKey nil", stellartypes.ScVal{Type: stellartypes.ScValTypeNonceKey}, "scvNonceKey: nil"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := stellarcap.ScValToProto(tc.sv)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestScValToProto_UnsupportedType(t *testing.T) {
	_, err := stellarcap.ScValToProto(stellartypes.ScVal{Type: stellartypes.ScValType(999)})
	require.ErrorContains(t, err, "unsupported ScVal type")
}

func TestScValToProto_InvalidArgs(t *testing.T) {
	tests := []struct {
		name    string
		sv      stellartypes.ScVal
		wantErr string
	}{
		{
			"scError contractCode nil",
			stellartypes.ScVal{Type: stellartypes.ScValTypeError, Error: &stellartypes.ScError{Type: stellartypes.ScErrorTypeContract}},
			"scError.contractCode: nil",
		},
		{
			"scError code nil",
			stellartypes.ScVal{Type: stellartypes.ScValTypeError, Error: &stellartypes.ScError{Type: stellartypes.ScErrorTypeWasmVM}},
			"nil code",
		},
		{
			"scAddress account wrong size",
			stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{Type: stellartypes.ScAddressTypeAccountID}},
			"scAddress.account",
		},
		{
			"scAddress contract wrong size",
			stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{Type: stellartypes.ScAddressTypeContractID}},
			"scAddress.contract: contractId must be 32 bytes",
		},
		{
			"scAddress muxed nil",
			stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{Type: stellartypes.ScAddressTypeMuxedAccount}},
			"scAddress.muxed: nil",
		},
		{
			"scAddress claimableBalance nil",
			stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{Type: stellartypes.ScAddressTypeClaimableBalanceID}},
			"scAddress.claimableBalance: nil",
		},
		{
			"scAddress liquidityPool wrong size",
			stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{Type: stellartypes.ScAddressTypeLiquidityPoolID}},
			"scAddress.liquidityPool: poolId must be 32 bytes",
		},
		{
			"scAddress unsupported type",
			stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{Type: stellartypes.ScAddressType(999)}},
			"unsupported ScAddress type",
		},
		{
			"contractExecutable wasm wrong size",
			stellartypes.ScVal{
				Type: stellartypes.ScValTypeContractInstance,
				ContractInstance: &stellartypes.ScContractInstance{
					Executable: &stellartypes.ContractExecutable{Type: stellartypes.ContractExecutableTypeWasmHash},
				},
			},
			"contractExecutable.wasm: wasmHash must be 32 bytes",
		},
		{
			"contractExecutable unsupported type",
			stellartypes.ScVal{
				Type: stellartypes.ScValTypeContractInstance,
				ContractInstance: &stellartypes.ScContractInstance{
					Executable: &stellartypes.ContractExecutable{Type: stellartypes.ContractExecutableType(999)},
				},
			},
			"unsupported ContractExecutable type",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := stellarcap.ScValToProto(tc.sv)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

// ---- Round-trip helpers ----------------------------------------------------

// scValRoundTrip exercises the cap-level ScValToProto/ProtoToScVal pair directly.
func scValRoundTrip(t *testing.T, sv stellartypes.ScVal) stellartypes.ScVal {
	t.Helper()
	proto, err := stellarcap.ScValToProto(sv)
	require.NoError(t, err)
	got, err := stellarcap.ProtoToScVal(proto)
	require.NoError(t, err)
	return got
}

func TestScVal_Bool(t *testing.T) {
	b := true
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeBool, Bool: &b}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Void(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeVoid, Void: &stellartypes.Void{}}
	got := scValRoundTrip(t, sv)
	require.Equal(t, stellartypes.ScValTypeVoid, got.Type)
}

func TestScVal_Error_ContractCode(t *testing.T) {
	cc := uint32(42)
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeError, Error: &stellartypes.ScError{
		Type:         stellartypes.ScErrorTypeContract,
		ContractCode: &cc,
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Error_Code(t *testing.T) {
	code := stellartypes.ScErrorCodeArithDomain
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeError, Error: &stellartypes.ScError{
		Type: stellartypes.ScErrorTypeWasmVM,
		Code: &code,
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_U32(t *testing.T) {
	u := uint32(0xDEAD)
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeU32, U32: &u}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_I32(t *testing.T) {
	i := int32(-1234)
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeI32, I32: &i}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_U64(t *testing.T) {
	u := uint64(1 << 40)
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeU64, U64: &u}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_I64(t *testing.T) {
	i := int64(-1 << 40)
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeI64, I64: &i}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Timepoint(t *testing.T) {
	tp := uint64(1_700_000_000)
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeTimepoint, Timepoint: &tp}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Duration(t *testing.T) {
	d := uint64(3600)
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeDuration, Duration: &d}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_U128(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeU128, U128: &stellartypes.UInt128Parts{Hi: 0xAAAA, Lo: 0xBBBB}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_I128(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeI128, I128: &stellartypes.Int128Parts{Hi: -7, Lo: 999}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_U256(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeU256, U256: &stellartypes.UInt256Parts{
		HiHi: 1, HiLo: 2, LoHi: 3, LoLo: 4,
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_I256(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeI256, I256: &stellartypes.Int256Parts{
		HiHi: -1, HiLo: 2, LoHi: 3, LoLo: 4,
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Bytes(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeBytes, Bytes: []byte{0x01, 0x02, 0x03}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_String(t *testing.T) {
	s := "hello world"
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeString, String: &s}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Symbol(t *testing.T) {
	sym := "transfer"
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeSymbol, Symbol: &sym}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Vec(t *testing.T) {
	u := uint32(1)
	inner := &stellartypes.ScVal{Type: stellartypes.ScValTypeU32, U32: &u}
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeVec, Vec: &stellartypes.ScVec{Values: []*stellartypes.ScVal{inner}}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Map(t *testing.T) {
	sym := "key"
	u := uint32(99)
	key := &stellartypes.ScVal{Type: stellartypes.ScValTypeSymbol, Symbol: &sym}
	val := &stellartypes.ScVal{Type: stellartypes.ScValTypeU32, U32: &u}
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeMap, Map: &stellartypes.ScMap{Entries: []stellartypes.ScMapEntry{
		{Key: key, Val: val},
	}}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func bytes32(b byte) []byte {
	out := make([]byte, 32)
	for i := range out {
		out[i] = b
	}
	return out
}

func TestScVal_Address_Account(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{
		Type:      stellartypes.ScAddressTypeAccountID,
		AccountID: bytes32(0x01),
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Address_Contract(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{
		Type:       stellartypes.ScAddressTypeContractID,
		ContractID: bytes32(0x02),
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Address_MuxedAccount(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{
		Type: stellartypes.ScAddressTypeMuxedAccount,
		MuxedAccount: &stellartypes.MuxedEd25519Account{
			ID:      777,
			Ed25519: bytes32(0x03),
		},
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Address_ClaimableBalance(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{
		Type: stellartypes.ScAddressTypeClaimableBalanceID,
		ClaimableBalance: &stellartypes.ClaimableBalanceID{
			V0: bytes32(0x04),
		},
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_Address_LiquidityPool(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeAddress, Address: &stellartypes.ScAddress{
		Type:            stellartypes.ScAddressTypeLiquidityPoolID,
		LiquidityPoolID: bytes32(0x05),
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_ContractInstance_Wasm(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeContractInstance, ContractInstance: &stellartypes.ScContractInstance{
		Executable: &stellartypes.ContractExecutable{
			Type:     stellartypes.ContractExecutableTypeWasmHash,
			WasmHash: bytes32(0x06),
		},
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_ContractInstance_StellarAsset(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeContractInstance, ContractInstance: &stellartypes.ScContractInstance{
		Executable: &stellartypes.ContractExecutable{
			Type:         stellartypes.ContractExecutableTypeStellarAsset,
			StellarAsset: true,
		},
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_ContractInstance_WithStorage(t *testing.T) {
	sym := "slot"
	u := uint32(1)
	key := &stellartypes.ScVal{Type: stellartypes.ScValTypeSymbol, Symbol: &sym}
	val := &stellartypes.ScVal{Type: stellartypes.ScValTypeU32, U32: &u}
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeContractInstance, ContractInstance: &stellartypes.ScContractInstance{
		Executable: &stellartypes.ContractExecutable{
			Type:     stellartypes.ContractExecutableTypeWasmHash,
			WasmHash: bytes32(0x07),
		},
		Storage: []stellartypes.ScMapEntry{
			{Key: key, Val: val},
		},
	}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_LedgerKeyContractInstance(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeLedgerKeyContractInstance, LedgerKeyContractInstance: &stellartypes.Void{}}
	got := scValRoundTrip(t, sv)
	require.Equal(t, stellartypes.ScValTypeLedgerKeyContractInstance, got.Type)
}

func TestScVal_LedgerKeyNonce(t *testing.T) {
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeNonceKey, NonceKey: &stellartypes.ScNonceKey{Nonce: 12345}}
	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_NestedVecMap(t *testing.T) {
	// Vec containing a Map: [{sym:"x" -> u32:1}]
	sym := "x"
	u := uint32(1)
	mapVal := &stellartypes.ScVal{Type: stellartypes.ScValTypeMap, Map: &stellartypes.ScMap{Entries: []stellartypes.ScMapEntry{
		{
			Key: &stellartypes.ScVal{Type: stellartypes.ScValTypeSymbol, Symbol: &sym},
			Val: &stellartypes.ScVal{Type: stellartypes.ScValTypeU32, U32: &u},
		},
	}}}
	sv := stellartypes.ScVal{Type: stellartypes.ScValTypeVec, Vec: &stellartypes.ScVec{Values: []*stellartypes.ScVal{mapVal}}}

	require.Equal(t, sv, scValRoundTrip(t, sv))
}

func TestScVal_ExceedsMaxDepth(t *testing.T) {
	// Build a domain ScVal nested 66 levels deep and confirm ScValToProto rejects it.
	u := uint32(0)
	cur := &stellartypes.ScVal{Type: stellartypes.ScValTypeU32, U32: &u}
	for i := 0; i < 66; i++ {
		cur = &stellartypes.ScVal{Type: stellartypes.ScValTypeVec, Vec: &stellartypes.ScVec{Values: []*stellartypes.ScVal{cur}}}
	}
	_, err := stellarcap.ScValToProto(*cur)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nesting exceeds maximum depth")
}

func TestScVal_ExceedsMaxDepth_EmptyStorageContractInstance(t *testing.T) {
	leaf := &stellartypes.ScVal{
		Type: stellartypes.ScValTypeContractInstance,
		ContractInstance: &stellartypes.ScContractInstance{
			Executable: &stellartypes.ContractExecutable{
				Type:         stellartypes.ContractExecutableTypeStellarAsset,
				StellarAsset: true,
			},
		},
	}
	cur := leaf
	for i := 0; i < 64; i++ {
		cur = &stellartypes.ScVal{Type: stellartypes.ScValTypeVec, Vec: &stellartypes.ScVec{Values: []*stellartypes.ScVal{cur}}}
	}
	_, err := stellarcap.ScValToProto(*cur)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nesting exceeds maximum depth")
}

func TestProtoScVal_Nil(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(nil)
	require.ErrorContains(t, err, "proto ScVal is nil")
}

func TestProtoScVal_AccountId_WrongLength(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
		Address: &scval.ScAddress_AccountId{AccountId: []byte{0x01, 0x02}},
	}}})
	require.ErrorContains(t, err, "accountId must be 32 bytes")
}

func TestProtoScVal_ContractId_WrongLength(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
		Address: &scval.ScAddress_ContractId{ContractId: []byte("short")},
	}}})
	require.ErrorContains(t, err, "contractId must be 32 bytes")
}

func TestProtoScVal_WasmHash_WrongLength(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: &scval.ScContractInstance{
		Executable: &scval.ContractExecutable{
			Type: &scval.ContractExecutable_WasmHash{WasmHash: []byte("tooshort")},
		},
	}}})
	require.ErrorContains(t, err, "wasmHash must be 32 bytes")
}

func TestProtoScVal_ExceedsMaxDepth(t *testing.T) {
	// Build a proto ScVal nested 66 levels deep and confirm ProtoToScVal rejects it.
	u := uint32(0)
	cur := &scval.ScVal{Value: &scval.ScVal_U32{U32: u}}
	for i := 0; i < 66; i++ {
		cur = &scval.ScVal{Value: &scval.ScVal_Vec{Vec: &scval.ScVec{Values: []*scval.ScVal{cur}}}}
	}
	_, err := stellarcap.ProtoToScVal(cur)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "nesting exceeds maximum depth"), "unexpected error: %v", err)
}

func TestProtoScVal_ExceedsMaxDepth_EmptyStorageContractInstance(t *testing.T) {
	leaf := &scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: &scval.ScContractInstance{
		Executable: &scval.ContractExecutable{Type: &scval.ContractExecutable_StellarAsset{StellarAsset: true}},
	}}}
	cur := leaf
	for i := 0; i < 64; i++ {
		cur = &scval.ScVal{Value: &scval.ScVal_Vec{Vec: &scval.ScVec{Values: []*scval.ScVal{cur}}}}
	}
	_, err := stellarcap.ProtoToScVal(cur)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nesting exceeds maximum depth")
}

func TestProtoScVal_NilInnerFields(t *testing.T) {
	tests := []struct {
		name    string
		in      *scval.ScVal
		wantErr string
	}{
		{"u128 nil", &scval.ScVal{Value: &scval.ScVal_U128{U128: nil}}, "scvU128: nil"},
		{"i128 nil", &scval.ScVal{Value: &scval.ScVal_I128{I128: nil}}, "scvI128: nil"},
		{"u256 nil", &scval.ScVal{Value: &scval.ScVal_U256{U256: nil}}, "scvU256: nil"},
		{"i256 nil", &scval.ScVal{Value: &scval.ScVal_I256{I256: nil}}, "scvI256: nil"},
		{"vec nil", &scval.ScVal{Value: &scval.ScVal_Vec{Vec: nil}}, "scvVec: nil"},
		{"map nil", &scval.ScVal{Value: &scval.ScVal_Map{Map: nil}}, "scvMap: nil"},
		{"nonceKey nil", &scval.ScVal{Value: &scval.ScVal_NonceKey{NonceKey: nil}}, "scvNonceKey: nil"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := stellarcap.ProtoToScVal(tc.in)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestProtoScVal_UnsupportedOneof(t *testing.T) {
	// A zero-value ScVal (nil Value oneof) hits the default case.
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{})
	require.ErrorContains(t, err, "unsupported proto ScVal type")
}

func TestProtoScAddress_Nil(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Address{Address: nil}})
	require.ErrorContains(t, err, "proto ScAddress is nil")
}

func TestProtoScAddress_MuxedAccount_Nil(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
		Address: &scval.ScAddress_MuxedAccount{MuxedAccount: nil},
	}}})
	require.ErrorContains(t, err, "muxedAccount: nil")
}

func TestProtoScAddress_MuxedAccount_WrongEd25519Size(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
		Address: &scval.ScAddress_MuxedAccount{MuxedAccount: &scval.MuxedEd25519Account{
			Id:      1,
			Ed25519: []byte{0x01, 0x02}, // not 32 bytes
		}},
	}}})
	require.ErrorContains(t, err, "muxedAccount.ed25519 must be 32 bytes")
}

func TestProtoScAddress_ClaimableBalanceId_Nil(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
		Address: &scval.ScAddress_ClaimableBalanceId{ClaimableBalanceId: nil},
	}}})
	require.ErrorContains(t, err, "claimableBalanceId: nil")
}

func TestProtoScAddress_ClaimableBalanceId_WrongV0Size(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
		Address: &scval.ScAddress_ClaimableBalanceId{ClaimableBalanceId: &scval.ClaimableBalanceId{
			V0: []byte{0x01, 0x02}, // not 32 bytes
		}},
	}}})
	require.ErrorContains(t, err, "claimableBalanceId.v0 must be 32 bytes")
}

func TestProtoScAddress_LiquidityPoolId_WrongSize(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{
		Address: &scval.ScAddress_LiquidityPoolId{LiquidityPoolId: []byte("short")},
	}}})
	require.ErrorContains(t, err, "liquidityPoolId must be 32 bytes")
}

func TestProtoScAddress_UnsupportedOneof(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Address{Address: &scval.ScAddress{Address: nil}}})
	require.ErrorContains(t, err, "unsupported proto ScAddress type")
}

func TestProtoContractExecutable_Nil(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: &scval.ScContractInstance{
		Executable: nil,
	}}})
	require.ErrorContains(t, err, "proto ContractExecutable is nil")
}

func TestProtoContractExecutable_UnsupportedOneof(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: &scval.ScContractInstance{
		Executable: &scval.ContractExecutable{Type: nil},
	}}})
	require.ErrorContains(t, err, "unsupported proto ContractExecutable type")
}

func TestProtoScContractInstance_Nil(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: nil}})
	require.ErrorContains(t, err, "proto ScContractInstance is nil")
}

func TestProtoScError_Nil(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Error{Error: nil}})
	require.ErrorContains(t, err, "proto ScError is nil")
}

func TestProtoScError_UnsupportedOneof(t *testing.T) {
	_, err := stellarcap.ProtoToScVal(&scval.ScVal{Value: &scval.ScVal_Error{Error: &scval.ScError{CodeOrContract: nil}}})
	require.ErrorContains(t, err, "unsupported ScError oneof")
}
