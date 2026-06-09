package stellar

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar/scval"
	stellarservicetypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

// ConvertGetLatestLedgerResponseFromProto converts a proto GetLatestLedgerResponse to the
// domain type. Hash is returned as lowercase hex; XDR fields are returned as standard base64.
func ConvertGetLatestLedgerResponseFromProto(p *GetLatestLedgerResponse) (stellarservicetypes.GetLatestLedgerResponse, error) {
	if p == nil {
		return stellarservicetypes.GetLatestLedgerResponse{}, fmt.Errorf("getLatestLedgerResponse is nil")
	}

	return stellarservicetypes.GetLatestLedgerResponse{
		Hash:              hex.EncodeToString(p.GetHash()),
		ProtocolVersion:   p.GetProtocolVersion(),
		Sequence:          p.GetSequence(),
		LedgerCloseTime:   p.GetLedgerCloseTime(),
		LedgerHeaderXDR:   base64.StdEncoding.EncodeToString(p.GetLedgerHeaderXdr()),
		LedgerMetadataXDR: base64.StdEncoding.EncodeToString(p.GetLedgerMetadataXdr()),
	}, nil
}

// ConvertGetLatestLedgerResponseToProto converts the domain GetLatestLedgerResponse to a proto.
// Hash must be a valid hex string; XDR fields must be valid standard base64.
func ConvertGetLatestLedgerResponseToProto(r stellarservicetypes.GetLatestLedgerResponse) (*GetLatestLedgerResponse, error) {
	hash, err := hex.DecodeString(r.Hash)
	if err != nil {
		return nil, fmt.Errorf("invalid hex hash %q: %w", r.Hash, err)
	}

	headerXDR, err := base64.StdEncoding.DecodeString(r.LedgerHeaderXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 ledgerHeaderXDR: %w", err)
	}

	metadataXDR, err := base64.StdEncoding.DecodeString(r.LedgerMetadataXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 ledgerMetadataXDR: %w", err)
	}

	return &GetLatestLedgerResponse{
		Hash:              hash,
		ProtocolVersion:   r.ProtocolVersion,
		Sequence:          r.Sequence,
		LedgerCloseTime:   r.LedgerCloseTime,
		LedgerHeaderXdr:   headerXDR,
		LedgerMetadataXdr: metadataXDR,
	}, nil
}

// ConvertReadContractRequestFromProto converts a proto ReadContractRequest to the
// domain type.
func ConvertReadContractRequestFromProto(p *ReadContractRequest) (stellarservicetypes.ReadContractRequest, error) {
	if p == nil {
		return stellarservicetypes.ReadContractRequest{}, fmt.Errorf("readContractRequest is nil")
	}
	if p.GetContractId() == "" {
		return stellarservicetypes.ReadContractRequest{}, fmt.Errorf("contractID is required")
	}
	if p.GetFunction() == "" {
		return stellarservicetypes.ReadContractRequest{}, fmt.Errorf("function is required")
	}

	pArgs := p.GetArgs()
	args := make([]stellarservicetypes.ScVal, len(pArgs))
	for i, psv := range pArgs {
		sv, err := ProtoToScVal(psv)
		if err != nil {
			return stellarservicetypes.ReadContractRequest{}, fmt.Errorf("args[%d]: %w", i, err)
		}
		args[i] = sv
	}
	return stellarservicetypes.ReadContractRequest{
		ContractID:     p.GetContractId(),
		Function:       p.GetFunction(),
		Args:           args,
		LedgerSequence: p.GetLedgerSequence(),
	}, nil
}

// ScValToProto converts a domain ScVal to its proto representation.
// Returns an error if nesting depth exceeds 64 levels.
func ScValToProto(sv stellarservicetypes.ScVal) (*scval.ScVal, error) {
	return scValToProtoAt(sv, 0)
}

func scValToProtoAt(sv stellarservicetypes.ScVal, depth int) (*scval.ScVal, error) {
	if depth > 64 {
		return nil, fmt.Errorf("scVal nesting exceeds maximum depth of 64")
	}
	switch sv.Type {
	case stellarservicetypes.ScValTypeBool:
		if sv.Bool == nil {
			return nil, fmt.Errorf("scvBool: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_B{B: *sv.Bool}}, nil
	case stellarservicetypes.ScValTypeVoid:
		if sv.Void == nil {
			return nil, fmt.Errorf("scvVoid: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_VoidVal{VoidVal: &scval.Void{}}}, nil
	case stellarservicetypes.ScValTypeError:
		if sv.Error == nil {
			return nil, fmt.Errorf("scvError: nil")
		}
		pe, err := scErrorToProto(sv.Error)
		if err != nil {
			return nil, err
		}
		return &scval.ScVal{Value: &scval.ScVal_Error{Error: pe}}, nil
	case stellarservicetypes.ScValTypeU32:
		if sv.U32 == nil {
			return nil, fmt.Errorf("scvU32: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_U32{U32: *sv.U32}}, nil
	case stellarservicetypes.ScValTypeI32:
		if sv.I32 == nil {
			return nil, fmt.Errorf("scvI32: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_I32{I32: *sv.I32}}, nil
	case stellarservicetypes.ScValTypeU64:
		if sv.U64 == nil {
			return nil, fmt.Errorf("scvU64: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_U64{U64: *sv.U64}}, nil
	case stellarservicetypes.ScValTypeI64:
		if sv.I64 == nil {
			return nil, fmt.Errorf("scvI64: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_I64{I64: *sv.I64}}, nil
	case stellarservicetypes.ScValTypeTimepoint:
		if sv.Timepoint == nil {
			return nil, fmt.Errorf("scvTimepoint: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_Timepoint{Timepoint: *sv.Timepoint}}, nil
	case stellarservicetypes.ScValTypeDuration:
		if sv.Duration == nil {
			return nil, fmt.Errorf("scvDuration: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_Duration{Duration: *sv.Duration}}, nil
	case stellarservicetypes.ScValTypeU128:
		if sv.U128 == nil {
			return nil, fmt.Errorf("scvU128: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_U128{U128: &scval.UInt128Parts{Hi: sv.U128.Hi, Lo: sv.U128.Lo}}}, nil
	case stellarservicetypes.ScValTypeI128:
		if sv.I128 == nil {
			return nil, fmt.Errorf("scvI128: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_I128{I128: &scval.Int128Parts{Hi: sv.I128.Hi, Lo: sv.I128.Lo}}}, nil
	case stellarservicetypes.ScValTypeU256:
		if sv.U256 == nil {
			return nil, fmt.Errorf("scvU256: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_U256{U256: &scval.UInt256Parts{
			HiHi: sv.U256.HiHi, HiLo: sv.U256.HiLo, LoHi: sv.U256.LoHi, LoLo: sv.U256.LoLo,
		}}}, nil
	case stellarservicetypes.ScValTypeI256:
		if sv.I256 == nil {
			return nil, fmt.Errorf("scvI256: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_I256{I256: &scval.Int256Parts{
			HiHi: sv.I256.HiHi, HiLo: sv.I256.HiLo, LoHi: sv.I256.LoHi, LoLo: sv.I256.LoLo,
		}}}, nil
	case stellarservicetypes.ScValTypeBytes:
		return &scval.ScVal{Value: &scval.ScVal_BytesVal{BytesVal: sv.Bytes}}, nil
	case stellarservicetypes.ScValTypeString:
		if sv.String == nil {
			return nil, fmt.Errorf("scvString: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_Str{Str: *sv.String}}, nil
	case stellarservicetypes.ScValTypeSymbol:
		if sv.Symbol == nil {
			return nil, fmt.Errorf("scvSymbol: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_Sym{Sym: *sv.Symbol}}, nil
	case stellarservicetypes.ScValTypeVec:
		if sv.Vec == nil {
			return nil, fmt.Errorf("scvVec: nil")
		}
		pVals := make([]*scval.ScVal, len(sv.Vec.Values))
		for i, elem := range sv.Vec.Values {
			if elem == nil {
				return nil, fmt.Errorf("vec[%d]: nil element", i)
			}
			pv, err := scValToProtoAt(*elem, depth+1)
			if err != nil {
				return nil, fmt.Errorf("vec[%d]: %w", i, err)
			}
			pVals[i] = pv
		}
		return &scval.ScVal{Value: &scval.ScVal_Vec{Vec: &scval.ScVec{Values: pVals}}}, nil
	case stellarservicetypes.ScValTypeMap:
		if sv.Map == nil {
			return nil, fmt.Errorf("scvMap: nil")
		}
		entries := make([]*scval.ScMapEntry, len(sv.Map.Entries))
		for i, entry := range sv.Map.Entries {
			if entry.Key == nil {
				return nil, fmt.Errorf("map[%d].key: nil", i)
			}
			if entry.Val == nil {
				return nil, fmt.Errorf("map[%d].val: nil", i)
			}
			pk, err := scValToProtoAt(*entry.Key, depth+1)
			if err != nil {
				return nil, fmt.Errorf("map[%d].key: %w", i, err)
			}
			pv, err := scValToProtoAt(*entry.Val, depth+1)
			if err != nil {
				return nil, fmt.Errorf("map[%d].val: %w", i, err)
			}
			entries[i] = &scval.ScMapEntry{Key: pk, Val: pv}
		}
		return &scval.ScVal{Value: &scval.ScVal_Map{Map: &scval.ScMap{Entries: entries}}}, nil
	case stellarservicetypes.ScValTypeAddress:
		if sv.Address == nil {
			return nil, fmt.Errorf("scvAddress: nil")
		}
		pa, err := scAddressToProto(sv.Address)
		if err != nil {
			return nil, err
		}
		return &scval.ScVal{Value: &scval.ScVal_Address{Address: pa}}, nil
	case stellarservicetypes.ScValTypeContractInstance:
		if sv.ContractInstance == nil {
			return nil, fmt.Errorf("scvContractInstance: nil")
		}
		pi, err := scContractInstanceToProtoAt(sv.ContractInstance, depth+1)
		if err != nil {
			return nil, err
		}
		return &scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: pi}}, nil
	case stellarservicetypes.ScValTypeLedgerKeyContractInstance:
		if sv.LedgerKeyContractInstance == nil {
			return nil, fmt.Errorf("scvLedgerKeyContractInstance: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_LedgerKeyContractInstance{LedgerKeyContractInstance: &scval.Void{}}}, nil
	case stellarservicetypes.ScValTypeNonceKey:
		if sv.NonceKey == nil {
			return nil, fmt.Errorf("scvNonceKey: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_NonceKey{NonceKey: &scval.ScNonceKey{Nonce: sv.NonceKey.Nonce}}}, nil
	default:
		return nil, fmt.Errorf("unsupported ScVal type: %d", sv.Type)
	}
}

func scErrorToProto(e *stellarservicetypes.ScError) (*scval.ScError, error) {
	pe := &scval.ScError{Type: scval.ScError_Type(e.Type)}
	if e.Type == stellarservicetypes.ScErrorTypeContract {
		if e.ContractCode == nil {
			return nil, fmt.Errorf("scError.contractCode: nil")
		}
		pe.CodeOrContract = &scval.ScError_ContractCode{ContractCode: *e.ContractCode}
	} else {
		if e.Code == nil {
			return nil, fmt.Errorf("scError type %d: nil code", e.Type)
		}
		pe.CodeOrContract = &scval.ScError_Code_{Code: scval.ScError_Code(*e.Code)}
	}
	return pe, nil
}

func scAddressToProto(a *stellarservicetypes.ScAddress) (*scval.ScAddress, error) {
	switch a.Type {
	case stellarservicetypes.ScAddressTypeAccountID:
		if len(a.AccountID) != 32 {
			return nil, fmt.Errorf("scAddress.account: accountId must be 32 bytes, got %d", len(a.AccountID))
		}
		return &scval.ScAddress{Address: &scval.ScAddress_AccountId{AccountId: a.AccountID}}, nil
	case stellarservicetypes.ScAddressTypeContractID:
		if len(a.ContractID) != 32 {
			return nil, fmt.Errorf("scAddress.contract: contractId must be 32 bytes, got %d", len(a.ContractID))
		}
		return &scval.ScAddress{Address: &scval.ScAddress_ContractId{ContractId: a.ContractID}}, nil
	case stellarservicetypes.ScAddressTypeMuxedAccount:
		if a.MuxedAccount == nil {
			return nil, fmt.Errorf("scAddress.muxed: nil")
		}
		if len(a.MuxedAccount.Ed25519) != 32 {
			return nil, fmt.Errorf("scAddress.muxed: ed25519 must be 32 bytes, got %d", len(a.MuxedAccount.Ed25519))
		}
		return &scval.ScAddress{Address: &scval.ScAddress_MuxedAccount{MuxedAccount: &scval.MuxedEd25519Account{
			Id:      a.MuxedAccount.ID,
			Ed25519: a.MuxedAccount.Ed25519,
		}}}, nil
	case stellarservicetypes.ScAddressTypeClaimableBalanceID:
		if a.ClaimableBalance == nil {
			return nil, fmt.Errorf("scAddress.claimableBalance: nil")
		}
		if len(a.ClaimableBalance.V0) != 32 {
			return nil, fmt.Errorf("scAddress.claimableBalance: v0 must be 32 bytes, got %d", len(a.ClaimableBalance.V0))
		}
		return &scval.ScAddress{Address: &scval.ScAddress_ClaimableBalanceId{ClaimableBalanceId: &scval.ClaimableBalanceId{
			V0: a.ClaimableBalance.V0,
		}}}, nil
	case stellarservicetypes.ScAddressTypeLiquidityPoolID:
		if len(a.LiquidityPoolID) != 32 {
			return nil, fmt.Errorf("scAddress.liquidityPool: poolId must be 32 bytes, got %d", len(a.LiquidityPoolID))
		}
		return &scval.ScAddress{Address: &scval.ScAddress_LiquidityPoolId{LiquidityPoolId: a.LiquidityPoolID}}, nil
	default:
		return nil, fmt.Errorf("unsupported ScAddress type: %d", a.Type)
	}
}

func contractExecutableToProto(exec *stellarservicetypes.ContractExecutable) (*scval.ContractExecutable, error) {
	if exec == nil {
		return nil, fmt.Errorf("contractExecutable: nil")
	}
	switch exec.Type {
	case stellarservicetypes.ContractExecutableTypeWasmHash:
		if len(exec.WasmHash) != 32 {
			return nil, fmt.Errorf("contractExecutable.wasm: wasmHash must be 32 bytes, got %d", len(exec.WasmHash))
		}
		return &scval.ContractExecutable{Type: &scval.ContractExecutable_WasmHash{WasmHash: exec.WasmHash}}, nil
	case stellarservicetypes.ContractExecutableTypeStellarAsset:
		return &scval.ContractExecutable{Type: &scval.ContractExecutable_StellarAsset{StellarAsset: true}}, nil
	default:
		return nil, fmt.Errorf("unsupported ContractExecutable type: %d", exec.Type)
	}
}

func scContractInstanceToProtoAt(inst *stellarservicetypes.ScContractInstance, depth int) (*scval.ScContractInstance, error) {
	if depth > 64 {
		return nil, fmt.Errorf("scVal nesting exceeds maximum depth of 64")
	}
	pExec, err := contractExecutableToProto(inst.Executable)
	if err != nil {
		return nil, err
	}
	pi := &scval.ScContractInstance{Executable: pExec}
	if len(inst.Storage) > 0 {
		entries := make([]*scval.ScMapEntry, len(inst.Storage))
		for i, entry := range inst.Storage {
			if entry.Key == nil {
				return nil, fmt.Errorf("instance.storage[%d].key: nil", i)
			}
			if entry.Val == nil {
				return nil, fmt.Errorf("instance.storage[%d].val: nil", i)
			}
			pk, err := scValToProtoAt(*entry.Key, depth+1)
			if err != nil {
				return nil, fmt.Errorf("instance.storage[%d].key: %w", i, err)
			}
			pv, err := scValToProtoAt(*entry.Val, depth+1)
			if err != nil {
				return nil, fmt.Errorf("instance.storage[%d].val: %w", i, err)
			}
			entries[i] = &scval.ScMapEntry{Key: pk, Val: pv}
		}
		pi.Storage = entries
	}
	return pi, nil
}

// ProtoToScVal converts a proto ScVal to its domain representation.
// Returns an error if nesting depth exceeds 64 levels.
func ProtoToScVal(sv *scval.ScVal) (stellarservicetypes.ScVal, error) {
	return protoToScValAt(sv, 0)
}

func protoToScValAt(sv *scval.ScVal, depth int) (stellarservicetypes.ScVal, error) {
	if depth > 64 {
		return stellarservicetypes.ScVal{}, fmt.Errorf("scVal nesting exceeds maximum depth of 64")
	}
	if sv == nil {
		return stellarservicetypes.ScVal{}, fmt.Errorf("proto ScVal is nil")
	}
	switch v := sv.Value.(type) {
	case *scval.ScVal_B:
		b := v.B
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeBool, Bool: &b}, nil
	case *scval.ScVal_VoidVal:
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeVoid, Void: &stellarservicetypes.Void{}}, nil
	case *scval.ScVal_Error:
		e, err := protoToScError(v.Error)
		if err != nil {
			return stellarservicetypes.ScVal{}, err
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeError, Error: e}, nil
	case *scval.ScVal_U32:
		u := v.U32
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeU32, U32: &u}, nil
	case *scval.ScVal_I32:
		i := v.I32
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeI32, I32: &i}, nil
	case *scval.ScVal_U64:
		u := v.U64
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeU64, U64: &u}, nil
	case *scval.ScVal_I64:
		i := v.I64
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeI64, I64: &i}, nil
	case *scval.ScVal_Timepoint:
		t := v.Timepoint
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeTimepoint, Timepoint: &t}, nil
	case *scval.ScVal_Duration:
		d := v.Duration
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeDuration, Duration: &d}, nil
	case *scval.ScVal_U128:
		if v.U128 == nil {
			return stellarservicetypes.ScVal{}, fmt.Errorf("scvU128: nil")
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeU128, U128: &stellarservicetypes.UInt128Parts{Hi: v.U128.Hi, Lo: v.U128.Lo}}, nil
	case *scval.ScVal_I128:
		if v.I128 == nil {
			return stellarservicetypes.ScVal{}, fmt.Errorf("scvI128: nil")
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeI128, I128: &stellarservicetypes.Int128Parts{Hi: v.I128.Hi, Lo: v.I128.Lo}}, nil
	case *scval.ScVal_U256:
		if v.U256 == nil {
			return stellarservicetypes.ScVal{}, fmt.Errorf("scvU256: nil")
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeU256, U256: &stellarservicetypes.UInt256Parts{
			HiHi: v.U256.HiHi, HiLo: v.U256.HiLo, LoHi: v.U256.LoHi, LoLo: v.U256.LoLo,
		}}, nil
	case *scval.ScVal_I256:
		if v.I256 == nil {
			return stellarservicetypes.ScVal{}, fmt.Errorf("scvI256: nil")
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeI256, I256: &stellarservicetypes.Int256Parts{
			HiHi: v.I256.HiHi, HiLo: v.I256.HiLo, LoHi: v.I256.LoHi, LoLo: v.I256.LoLo,
		}}, nil
	case *scval.ScVal_BytesVal:
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeBytes, Bytes: v.BytesVal}, nil
	case *scval.ScVal_Str:
		s := v.Str
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeString, String: &s}, nil
	case *scval.ScVal_Sym:
		s := v.Sym
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeSymbol, Symbol: &s}, nil
	case *scval.ScVal_Vec:
		if v.Vec == nil {
			return stellarservicetypes.ScVal{}, fmt.Errorf("scvVec: nil")
		}
		values := make([]*stellarservicetypes.ScVal, len(v.Vec.Values))
		for i, pv := range v.Vec.Values {
			dv, err := protoToScValAt(pv, depth+1)
			if err != nil {
				return stellarservicetypes.ScVal{}, fmt.Errorf("vec[%d]: %w", i, err)
			}
			elem := dv
			values[i] = &elem
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeVec, Vec: &stellarservicetypes.ScVec{Values: values}}, nil
	case *scval.ScVal_Map:
		if v.Map == nil {
			return stellarservicetypes.ScVal{}, fmt.Errorf("scvMap: nil")
		}
		entries := make([]stellarservicetypes.ScMapEntry, len(v.Map.Entries))
		for i, pe := range v.Map.Entries {
			if pe == nil {
				return stellarservicetypes.ScVal{}, fmt.Errorf("map[%d]: nil entry", i)
			}
			dk, err := protoToScValAt(pe.Key, depth+1)
			if err != nil {
				return stellarservicetypes.ScVal{}, fmt.Errorf("map[%d].key: %w", i, err)
			}
			dv, err := protoToScValAt(pe.Val, depth+1)
			if err != nil {
				return stellarservicetypes.ScVal{}, fmt.Errorf("map[%d].val: %w", i, err)
			}
			key, val := dk, dv
			entries[i] = stellarservicetypes.ScMapEntry{Key: &key, Val: &val}
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeMap, Map: &stellarservicetypes.ScMap{Entries: entries}}, nil
	case *scval.ScVal_Address:
		a, err := protoToScAddress(v.Address)
		if err != nil {
			return stellarservicetypes.ScVal{}, err
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeAddress, Address: a}, nil
	case *scval.ScVal_ContractInstance:
		ci, err := protoToScContractInstanceAt(v.ContractInstance, depth+1)
		if err != nil {
			return stellarservicetypes.ScVal{}, err
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeContractInstance, ContractInstance: ci}, nil
	case *scval.ScVal_LedgerKeyContractInstance:
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeLedgerKeyContractInstance, LedgerKeyContractInstance: &stellarservicetypes.Void{}}, nil
	case *scval.ScVal_NonceKey:
		if v.NonceKey == nil {
			return stellarservicetypes.ScVal{}, fmt.Errorf("scvNonceKey: nil")
		}
		return stellarservicetypes.ScVal{Type: stellarservicetypes.ScValTypeNonceKey, NonceKey: &stellarservicetypes.ScNonceKey{Nonce: v.NonceKey.Nonce}}, nil
	default:
		return stellarservicetypes.ScVal{}, fmt.Errorf("unsupported proto ScVal type: %T", sv.Value)
	}
}

func protoToScError(e *scval.ScError) (*stellarservicetypes.ScError, error) {
	if e == nil {
		return nil, fmt.Errorf("proto ScError is nil")
	}
	de := &stellarservicetypes.ScError{Type: stellarservicetypes.ScErrorType(e.Type)}
	switch v := e.CodeOrContract.(type) {
	case *scval.ScError_ContractCode:
		cc := v.ContractCode
		de.ContractCode = &cc
	case *scval.ScError_Code_:
		code := stellarservicetypes.ScErrorCode(v.Code)
		de.Code = &code
	default:
		return nil, fmt.Errorf("unsupported ScError oneof: %T", e.CodeOrContract)
	}
	return de, nil
}

func protoToScAddress(a *scval.ScAddress) (*stellarservicetypes.ScAddress, error) {
	if a == nil {
		return nil, fmt.Errorf("proto ScAddress is nil")
	}
	switch v := a.Address.(type) {
	case *scval.ScAddress_AccountId:
		if len(v.AccountId) != 32 {
			return nil, fmt.Errorf("accountId must be 32 bytes, got %d", len(v.AccountId))
		}
		return &stellarservicetypes.ScAddress{Type: stellarservicetypes.ScAddressTypeAccountID, AccountID: v.AccountId}, nil
	case *scval.ScAddress_ContractId:
		if len(v.ContractId) != 32 {
			return nil, fmt.Errorf("contractId must be 32 bytes, got %d", len(v.ContractId))
		}
		return &stellarservicetypes.ScAddress{Type: stellarservicetypes.ScAddressTypeContractID, ContractID: v.ContractId}, nil
	case *scval.ScAddress_MuxedAccount:
		if v.MuxedAccount == nil {
			return nil, fmt.Errorf("muxedAccount: nil")
		}
		if len(v.MuxedAccount.Ed25519) != 32 {
			return nil, fmt.Errorf("muxedAccount.ed25519 must be 32 bytes, got %d", len(v.MuxedAccount.Ed25519))
		}
		return &stellarservicetypes.ScAddress{
			Type: stellarservicetypes.ScAddressTypeMuxedAccount,
			MuxedAccount: &stellarservicetypes.MuxedEd25519Account{
				ID:      v.MuxedAccount.Id,
				Ed25519: v.MuxedAccount.Ed25519,
			},
		}, nil
	case *scval.ScAddress_ClaimableBalanceId:
		if v.ClaimableBalanceId == nil {
			return nil, fmt.Errorf("claimableBalanceId: nil")
		}
		if len(v.ClaimableBalanceId.V0) != 32 {
			return nil, fmt.Errorf("claimableBalanceId.v0 must be 32 bytes, got %d", len(v.ClaimableBalanceId.V0))
		}
		return &stellarservicetypes.ScAddress{
			Type:             stellarservicetypes.ScAddressTypeClaimableBalanceID,
			ClaimableBalance: &stellarservicetypes.ClaimableBalanceID{V0: v.ClaimableBalanceId.V0},
		}, nil
	case *scval.ScAddress_LiquidityPoolId:
		if len(v.LiquidityPoolId) != 32 {
			return nil, fmt.Errorf("liquidityPoolId must be 32 bytes, got %d", len(v.LiquidityPoolId))
		}
		return &stellarservicetypes.ScAddress{Type: stellarservicetypes.ScAddressTypeLiquidityPoolID, LiquidityPoolID: v.LiquidityPoolId}, nil
	default:
		return nil, fmt.Errorf("unsupported proto ScAddress type: %T", a.Address)
	}
}

func protoToContractExecutable(exec *scval.ContractExecutable) (*stellarservicetypes.ContractExecutable, error) {
	if exec == nil {
		return nil, fmt.Errorf("proto ContractExecutable is nil")
	}
	switch v := exec.Type.(type) {
	case *scval.ContractExecutable_WasmHash:
		if len(v.WasmHash) != 32 {
			return nil, fmt.Errorf("wasmHash must be 32 bytes, got %d", len(v.WasmHash))
		}
		return &stellarservicetypes.ContractExecutable{Type: stellarservicetypes.ContractExecutableTypeWasmHash, WasmHash: v.WasmHash}, nil
	case *scval.ContractExecutable_StellarAsset:
		return &stellarservicetypes.ContractExecutable{Type: stellarservicetypes.ContractExecutableTypeStellarAsset, StellarAsset: true}, nil
	default:
		return nil, fmt.Errorf("unsupported proto ContractExecutable type: %T", exec.Type)
	}
}

func protoToScContractInstanceAt(inst *scval.ScContractInstance, depth int) (*stellarservicetypes.ScContractInstance, error) {
	if depth > 64 {
		return nil, fmt.Errorf("scVal nesting exceeds maximum depth of 64")
	}
	if inst == nil {
		return nil, fmt.Errorf("proto ScContractInstance is nil")
	}
	dExec, err := protoToContractExecutable(inst.Executable)
	if err != nil {
		return nil, err
	}
	di := &stellarservicetypes.ScContractInstance{Executable: dExec}
	if len(inst.Storage) > 0 {
		entries := make([]stellarservicetypes.ScMapEntry, len(inst.Storage))
		for i, pe := range inst.Storage {
			if pe == nil {
				return nil, fmt.Errorf("instance.storage[%d]: nil entry", i)
			}
			dk, err := protoToScValAt(pe.Key, depth+1)
			if err != nil {
				return nil, fmt.Errorf("instance.storage[%d].key: %w", i, err)
			}
			dv, err := protoToScValAt(pe.Val, depth+1)
			if err != nil {
				return nil, fmt.Errorf("instance.storage[%d].val: %w", i, err)
			}
			key, val := dk, dv
			entries[i] = stellarservicetypes.ScMapEntry{Key: &key, Val: &val}
		}
		di.Storage = entries
	}
	return di, nil
}
