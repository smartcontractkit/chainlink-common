package stellar

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar/scval"

	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

// ConvertGetLedgerEntriesRequestToProto converts a domain GetLedgerEntriesRequest to its proto representation.
func ConvertGetLedgerEntriesRequestToProto(req stellar.GetLedgerEntriesRequest) (*GetLedgerEntriesRequest, error) {
	keys := make([][]byte, len(req.Keys))
	var errs []error
	for i, k := range req.Keys {
		b, err := base64.StdEncoding.DecodeString(k)
		if err != nil {
			errs = append(errs, fmt.Errorf("key[%d]: invalid base64 XDR %q: %w", i, k, err))
			continue
		}
		keys[i] = b
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return &GetLedgerEntriesRequest{Keys: keys}, nil
}

// ConvertGetLedgerEntriesRequestFromProto converts a proto GetLedgerEntriesRequest to the domain type.
func ConvertGetLedgerEntriesRequestFromProto(p *GetLedgerEntriesRequest) (stellar.GetLedgerEntriesRequest, error) {
	if p == nil {
		return stellar.GetLedgerEntriesRequest{}, fmt.Errorf("get ledger entries request is nil")
	}
	if len(p.GetKeys()) == 0 {
		return stellar.GetLedgerEntriesRequest{}, fmt.Errorf("ledger entry keys are empty")
	}
	rawKeys := p.GetKeys()
	keys := make([]string, len(rawKeys))
	for i, k := range rawKeys {
		keys[i] = base64.StdEncoding.EncodeToString(k)
	}
	return stellar.GetLedgerEntriesRequest{Keys: keys}, nil
}

// ConvertLedgerEntryResultToProto converts a domain LedgerEntryResult to its proto representation.
func ConvertLedgerEntryResultToProto(r stellar.LedgerEntryResult) (*LedgerEntryResult, error) {
	keyXDR, err := base64.StdEncoding.DecodeString(r.KeyXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid key xdr %q: %w", r.KeyXDR, err)
	}
	dataXDR, err := base64.StdEncoding.DecodeString(r.DataXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid data xdr %q: %w", r.DataXDR, err)
	}
	extXDR, err := base64.StdEncoding.DecodeString(r.ExtensionXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid extension xdr %q: %w", r.ExtensionXDR, err)
	}
	pr := &LedgerEntryResult{
		KeyXdr:             keyXDR,
		DataXdr:            dataXDR,
		LastModifiedLedger: r.LastModifiedLedger,
		ExtensionXdr:       extXDR,
	}
	if r.LiveUntilLedgerSeq != nil {
		pr.HasLiveUntilLedgerSeq = true
		pr.LiveUntilLedgerSeq = *r.LiveUntilLedgerSeq
	}
	return pr, nil
}

// ConvertLedgerEntryResultFromProto converts a proto LedgerEntryResult to the domain type.
func ConvertLedgerEntryResultFromProto(p *LedgerEntryResult) (stellar.LedgerEntryResult, error) {
	if p == nil {
		return stellar.LedgerEntryResult{}, fmt.Errorf("ledger entry result is nil")
	}
	r := stellar.LedgerEntryResult{
		KeyXDR:             base64.StdEncoding.EncodeToString(p.GetKeyXdr()),
		DataXDR:            base64.StdEncoding.EncodeToString(p.GetDataXdr()),
		LastModifiedLedger: p.GetLastModifiedLedger(),
		ExtensionXDR:       base64.StdEncoding.EncodeToString(p.GetExtensionXdr()),
	}
	if p.GetHasLiveUntilLedgerSeq() {
		v := p.GetLiveUntilLedgerSeq()
		r.LiveUntilLedgerSeq = &v
	}
	return r, nil
}

// ConvertGetLedgerEntriesResponseToProto converts a domain GetLedgerEntriesResponse to its proto representation.
func ConvertGetLedgerEntriesResponseToProto(resp stellar.GetLedgerEntriesResponse) (*GetLedgerEntriesResponse, error) {
	entries := make([]*LedgerEntryResult, 0, len(resp.Entries))
	for i, e := range resp.Entries {
		protoEntry, err := ConvertLedgerEntryResultToProto(e)
		if err != nil {
			return nil, fmt.Errorf("entry[%d]: %w", i, err)
		}
		entries = append(entries, protoEntry)
	}
	return &GetLedgerEntriesResponse{
		Entries:      entries,
		LatestLedger: resp.LatestLedger,
	}, nil
}

// ConvertGetLedgerEntriesResponseFromProto converts a proto GetLedgerEntriesResponse to the domain type.
func ConvertGetLedgerEntriesResponseFromProto(p *GetLedgerEntriesResponse) (stellar.GetLedgerEntriesResponse, error) {
	if p == nil {
		return stellar.GetLedgerEntriesResponse{}, fmt.Errorf("get ledger entries response is nil")
	}
	pEntries := p.GetEntries()
	entries := make([]stellar.LedgerEntryResult, 0, len(pEntries))
	for i, pe := range pEntries {
		e, err := ConvertLedgerEntryResultFromProto(pe)
		if err != nil {
			return stellar.GetLedgerEntriesResponse{}, fmt.Errorf("entry[%d]: %w", i, err)
		}
		entries = append(entries, e)
	}
	return stellar.GetLedgerEntriesResponse{
		Entries:      entries,
		LatestLedger: p.GetLatestLedger(),
	}, nil
}

// ConvertGetLatestLedgerResponseToProto converts a domain GetLatestLedgerResponse to its proto representation.
func ConvertGetLatestLedgerResponseToProto(resp stellar.GetLatestLedgerResponse) (*GetLatestLedgerResponse, error) {
	hash, err := hex.DecodeString(resp.Hash)
	if err != nil {
		return nil, fmt.Errorf("invalid hex hash %q: %w", resp.Hash, err)
	}

	headerXDR, err := base64.StdEncoding.DecodeString(resp.LedgerHeaderXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid ledger header xdr %q: %w", resp.LedgerHeaderXDR, err)
	}
	metaXDR, err := base64.StdEncoding.DecodeString(resp.LedgerMetadataXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid ledger metadata xdr %q: %w", resp.LedgerMetadataXDR, err)
	}
	return &GetLatestLedgerResponse{
		Hash:              hash,
		ProtocolVersion:   resp.ProtocolVersion,
		Sequence:          resp.Sequence,
		LedgerCloseTime:   resp.LedgerCloseTime,
		LedgerHeaderXdr:   headerXDR,
		LedgerMetadataXdr: metaXDR,
	}, nil
}

// ConvertGetLatestLedgerResponseFromProto converts a proto GetLatestLedgerResponse to the domain type.
func ConvertGetLatestLedgerResponseFromProto(p *GetLatestLedgerResponse) (stellar.GetLatestLedgerResponse, error) {
	if p == nil {
		return stellar.GetLatestLedgerResponse{}, fmt.Errorf("get latest ledger response is nil")
	}
	return stellar.GetLatestLedgerResponse{
		Hash:              hex.EncodeToString(p.GetHash()),
		ProtocolVersion:   p.GetProtocolVersion(),
		Sequence:          p.GetSequence(),
		LedgerCloseTime:   p.GetLedgerCloseTime(),
		LedgerHeaderXDR:   base64.StdEncoding.EncodeToString(p.GetLedgerHeaderXdr()),
		LedgerMetadataXDR: base64.StdEncoding.EncodeToString(p.GetLedgerMetadataXdr()),
	}, nil
}

func unwrapScVec(v **xdr.ScVec) (xdr.ScVec, error) {
	if v == nil || *v == nil {
		return nil, fmt.Errorf("vec is nil")
	}
	return **v, nil
}

func unwrapScMap(v **xdr.ScMap) (xdr.ScMap, error) {
	if v == nil || *v == nil {
		return nil, fmt.Errorf("map is nil")
	}
	return **v, nil
}

// xdrScValToProto converts an XDR ScVal to its proto representation.
// Returns an error if nesting depth exceeds 64 levels.
func xdrScValToProto(sv xdr.ScVal) (*scval.ScVal, error) {
	return xdrScValToProtoAt(sv, 0)
}

func xdrScValToProtoAt(sv xdr.ScVal, depth int) (*scval.ScVal, error) {
	if depth > 64 {
		return nil, fmt.Errorf("ScVal nesting exceeds maximum depth of 64")
	}
	switch sv.Type {
	case xdr.ScValTypeScvBool:
		if sv.B == nil {
			return nil, fmt.Errorf("scvBool: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_B{B: *sv.B}}, nil
	case xdr.ScValTypeScvVoid:
		return &scval.ScVal{Value: &scval.ScVal_VoidVal{VoidVal: &scval.Void{}}}, nil
	case xdr.ScValTypeScvError:
		if sv.Error == nil {
			return nil, fmt.Errorf("scvError: nil")
		}
		pe, err := xdrScErrorToProto(sv.Error)
		if err != nil {
			return nil, err
		}
		return &scval.ScVal{Value: &scval.ScVal_Error{Error: pe}}, nil
	case xdr.ScValTypeScvU32:
		if sv.U32 == nil {
			return nil, fmt.Errorf("scvU32: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_U32{U32: uint32(*sv.U32)}}, nil
	case xdr.ScValTypeScvI32:
		if sv.I32 == nil {
			return nil, fmt.Errorf("scvI32: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_I32{I32: int32(*sv.I32)}}, nil
	case xdr.ScValTypeScvU64:
		if sv.U64 == nil {
			return nil, fmt.Errorf("scvU64: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_U64{U64: uint64(*sv.U64)}}, nil
	case xdr.ScValTypeScvI64:
		if sv.I64 == nil {
			return nil, fmt.Errorf("scvI64: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_I64{I64: int64(*sv.I64)}}, nil
	case xdr.ScValTypeScvTimepoint:
		if sv.Timepoint == nil {
			return nil, fmt.Errorf("scvTimepoint: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_Timepoint{Timepoint: uint64(*sv.Timepoint)}}, nil
	case xdr.ScValTypeScvDuration:
		if sv.Duration == nil {
			return nil, fmt.Errorf("scvDuration: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_Duration{Duration: uint64(*sv.Duration)}}, nil
	case xdr.ScValTypeScvU128:
		if sv.U128 == nil {
			return nil, fmt.Errorf("scvU128: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_U128{U128: &scval.UInt128Parts{
			Hi: uint64(sv.U128.Hi),
			Lo: uint64(sv.U128.Lo),
		}}}, nil
	case xdr.ScValTypeScvI128:
		if sv.I128 == nil {
			return nil, fmt.Errorf("scvI128: nil")
		}
		// Per XDR spec for signed integer types: Hi is signed, Lo is unsigned.
		return &scval.ScVal{Value: &scval.ScVal_I128{I128: &scval.Int128Parts{
			Hi: int64(sv.I128.Hi),
			Lo: uint64(sv.I128.Lo),
		}}}, nil
	case xdr.ScValTypeScvU256:
		if sv.U256 == nil {
			return nil, fmt.Errorf("scvU256: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_U256{U256: &scval.UInt256Parts{
			HiHi: uint64(sv.U256.HiHi),
			HiLo: uint64(sv.U256.HiLo),
			LoHi: uint64(sv.U256.LoHi),
			LoLo: uint64(sv.U256.LoLo),
		}}}, nil
	case xdr.ScValTypeScvI256:
		if sv.I256 == nil {
			return nil, fmt.Errorf("scvI256: nil")
		}
		// Per XDR spec for signed integer types: HiHi is signed, remaining parts are unsigned.
		return &scval.ScVal{Value: &scval.ScVal_I256{I256: &scval.Int256Parts{
			HiHi: int64(sv.I256.HiHi),
			HiLo: uint64(sv.I256.HiLo),
			LoHi: uint64(sv.I256.LoHi),
			LoLo: uint64(sv.I256.LoLo),
		}}}, nil
	case xdr.ScValTypeScvBytes:
		if sv.Bytes == nil {
			return nil, fmt.Errorf("scvBytes: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_BytesVal{BytesVal: *sv.Bytes}}, nil
	case xdr.ScValTypeScvString:
		if sv.Str == nil {
			return nil, fmt.Errorf("scvString: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_Str{Str: string(*sv.Str)}}, nil
	case xdr.ScValTypeScvSymbol:
		if sv.Sym == nil {
			return nil, fmt.Errorf("scvSymbol: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_Sym{Sym: string(*sv.Sym)}}, nil
	case xdr.ScValTypeScvVec:
		xVec, err := unwrapScVec(sv.Vec)
		if err != nil {
			return nil, fmt.Errorf("scvVec: %w", err)
		}
		pVals := make([]*scval.ScVal, len(xVec))
		for i, elem := range xVec {
			pv, err := xdrScValToProtoAt(elem, depth+1)
			if err != nil {
				return nil, fmt.Errorf("vec[%d]: %w", i, err)
			}
			pVals[i] = pv
		}
		return &scval.ScVal{Value: &scval.ScVal_Vec{Vec: &scval.ScVec{Values: pVals}}}, nil
	case xdr.ScValTypeScvMap:
		xMap, err := unwrapScMap(sv.Map)
		if err != nil {
			return nil, fmt.Errorf("scvMap: %w", err)
		}
		entries := make([]*scval.ScMapEntry, len(xMap))
		for i, entry := range xMap {
			pk, err := xdrScValToProtoAt(entry.Key, depth+1)
			if err != nil {
				return nil, fmt.Errorf("map[%d].key: %w", i, err)
			}
			pv, err := xdrScValToProtoAt(entry.Val, depth+1)
			if err != nil {
				return nil, fmt.Errorf("map[%d].val: %w", i, err)
			}
			entries[i] = &scval.ScMapEntry{Key: pk, Val: pv}
		}
		return &scval.ScVal{Value: &scval.ScVal_Map{Map: &scval.ScMap{Entries: entries}}}, nil
	case xdr.ScValTypeScvAddress:
		if sv.Address == nil {
			return nil, fmt.Errorf("scvAddress: nil")
		}
		pa, err := xdrScAddressToProto(sv.Address)
		if err != nil {
			return nil, err
		}
		return &scval.ScVal{Value: &scval.ScVal_Address{Address: pa}}, nil
	case xdr.ScValTypeScvContractInstance:
		if sv.Instance == nil {
			return nil, fmt.Errorf("scvContractInstance: nil")
		}
		pi, err := xdrScContractInstanceToProtoAt(sv.Instance, depth+1)
		if err != nil {
			return nil, err
		}
		return &scval.ScVal{Value: &scval.ScVal_ContractInstance{ContractInstance: pi}}, nil
	case xdr.ScValTypeScvLedgerKeyContractInstance:
		return &scval.ScVal{Value: &scval.ScVal_LedgerKeyContractInstance{LedgerKeyContractInstance: &scval.Void{}}}, nil
	case xdr.ScValTypeScvLedgerKeyNonce:
		if sv.NonceKey == nil {
			return nil, fmt.Errorf("scvLedgerKeyNonce: nil")
		}
		return &scval.ScVal{Value: &scval.ScVal_NonceKey{NonceKey: &scval.ScNonceKey{Nonce: int64(sv.NonceKey.Nonce)}}}, nil
	default:
		return nil, fmt.Errorf("unsupported ScVal type: %d", sv.Type)
	}
}

func xdrScErrorToProto(e *xdr.ScError) (*scval.ScError, error) {
	pe := &scval.ScError{Type: scval.ScError_Type(e.Type)}
	if e.Type == xdr.ScErrorTypeSceContract {
		if e.ContractCode == nil {
			return nil, fmt.Errorf("scError.contractCode: nil")
		}
		pe.CodeOrContract = &scval.ScError_ContractCode{ContractCode: uint32(*e.ContractCode)}
	} else {
		if e.Code == nil {
			return nil, fmt.Errorf("scError type %d: nil code", e.Type)
		}
		pe.CodeOrContract = &scval.ScError_Code_{Code: scval.ScError_Code(*e.Code)}
	}
	return pe, nil
}

func xdrScAddressToProto(a *xdr.ScAddress) (*scval.ScAddress, error) {
	switch a.Type {
	case xdr.ScAddressTypeScAddressTypeAccount:
		if a.AccountId == nil || a.AccountId.Ed25519 == nil {
			return nil, fmt.Errorf("scAddress.account: nil accountId or ed25519")
		}
		return &scval.ScAddress{Address: &scval.ScAddress_AccountId{AccountId: (*a.AccountId.Ed25519)[:]}}, nil
	case xdr.ScAddressTypeScAddressTypeContract:
		if a.ContractId == nil {
			return nil, fmt.Errorf("scAddress.contract: nil contractId")
		}
		return &scval.ScAddress{Address: &scval.ScAddress_ContractId{ContractId: (*a.ContractId)[:]}}, nil
	case xdr.ScAddressTypeScAddressTypeMuxedAccount:
		if a.MuxedAccount == nil {
			return nil, fmt.Errorf("scAddress.muxed: nil")
		}
		return &scval.ScAddress{Address: &scval.ScAddress_MuxedAccount{MuxedAccount: &scval.MuxedEd25519Account{
			Id:      uint64(a.MuxedAccount.Id),
			Ed25519: a.MuxedAccount.Ed25519[:],
		}}}, nil
	case xdr.ScAddressTypeScAddressTypeClaimableBalance:
		if a.ClaimableBalanceId == nil || a.ClaimableBalanceId.V0 == nil {
			return nil, fmt.Errorf("scAddress.claimableBalance: nil")
		}
		return &scval.ScAddress{Address: &scval.ScAddress_ClaimableBalanceId{ClaimableBalanceId: &scval.ClaimableBalanceId{
			V0: (*a.ClaimableBalanceId.V0)[:],
		}}}, nil
	case xdr.ScAddressTypeScAddressTypeLiquidityPool:
		if a.LiquidityPoolId == nil {
			return nil, fmt.Errorf("scAddress.liquidityPool: nil poolId")
		}
		return &scval.ScAddress{Address: &scval.ScAddress_LiquidityPoolId{LiquidityPoolId: (*a.LiquidityPoolId)[:]}}, nil
	default:
		return nil, fmt.Errorf("unsupported ScAddress type: %d", a.Type)
	}
}

func xdrContractExecutableToProto(exec xdr.ContractExecutable) (*scval.ContractExecutable, error) {
	switch exec.Type {
	case xdr.ContractExecutableTypeContractExecutableWasm:
		if exec.WasmHash == nil {
			return nil, fmt.Errorf("contractExecutable.wasm: nil wasmHash")
		}
		return &scval.ContractExecutable{Type: &scval.ContractExecutable_WasmHash{WasmHash: (*exec.WasmHash)[:]}}, nil
	case xdr.ContractExecutableTypeContractExecutableStellarAsset:
		return &scval.ContractExecutable{Type: &scval.ContractExecutable_StellarAsset{StellarAsset: true}}, nil
	default:
		return nil, fmt.Errorf("unsupported ContractExecutable type: %d", exec.Type)
	}
}

func xdrScContractInstanceToProtoAt(inst *xdr.ScContractInstance, depth int) (*scval.ScContractInstance, error) {
	pExec, err := xdrContractExecutableToProto(inst.Executable)
	if err != nil {
		return nil, err
	}
	pi := &scval.ScContractInstance{Executable: pExec}
	if inst.Storage != nil {
		xMap := *inst.Storage
		entries := make([]*scval.ScMapEntry, len(xMap))
		for i, entry := range xMap {
			pk, err := xdrScValToProtoAt(entry.Key, depth+1)
			if err != nil {
				return nil, fmt.Errorf("instance.storage[%d].key: %w", i, err)
			}
			pv, err := xdrScValToProtoAt(entry.Val, depth+1)
			if err != nil {
				return nil, fmt.Errorf("instance.storage[%d].val: %w", i, err)
			}
			entries[i] = &scval.ScMapEntry{Key: pk, Val: pv}
		}
		pi.Storage = entries
	}
	return pi, nil
}

// protoScValToXDR converts a proto ScVal to its XDR representation.
// Returns an error if nesting depth exceeds 64 levels.
func protoScValToXDR(sv *scval.ScVal) (xdr.ScVal, error) {
	return protoScValToXDRAt(sv, 0)
}

func protoScValToXDRAt(sv *scval.ScVal, depth int) (xdr.ScVal, error) {
	if depth > 64 {
		return xdr.ScVal{}, fmt.Errorf("scVal nesting exceeds maximum depth of 64")
	}
	if sv == nil {
		return xdr.ScVal{}, fmt.Errorf("proto ScVal is nil")
	}
	switch v := sv.Value.(type) {
	case *scval.ScVal_B:
		b := v.B
		return xdr.ScVal{Type: xdr.ScValTypeScvBool, B: &b}, nil
	case *scval.ScVal_VoidVal:
		return xdr.ScVal{Type: xdr.ScValTypeScvVoid}, nil
	case *scval.ScVal_Error:
		xe, err := protoScErrorToXDR(v.Error)
		if err != nil {
			return xdr.ScVal{}, err
		}
		return xdr.ScVal{Type: xdr.ScValTypeScvError, Error: &xe}, nil
	case *scval.ScVal_U32:
		u := xdr.Uint32(v.U32)
		return xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &u}, nil
	case *scval.ScVal_I32:
		i := xdr.Int32(v.I32)
		return xdr.ScVal{Type: xdr.ScValTypeScvI32, I32: &i}, nil
	case *scval.ScVal_U64:
		u := xdr.Uint64(v.U64)
		return xdr.ScVal{Type: xdr.ScValTypeScvU64, U64: &u}, nil
	case *scval.ScVal_I64:
		i := xdr.Int64(v.I64)
		return xdr.ScVal{Type: xdr.ScValTypeScvI64, I64: &i}, nil
	case *scval.ScVal_Timepoint:
		t := xdr.TimePoint(v.Timepoint)
		return xdr.ScVal{Type: xdr.ScValTypeScvTimepoint, Timepoint: &t}, nil
	case *scval.ScVal_Duration:
		d := xdr.Duration(v.Duration)
		return xdr.ScVal{Type: xdr.ScValTypeScvDuration, Duration: &d}, nil
	case *scval.ScVal_U128:
		if v.U128 == nil {
			return xdr.ScVal{}, fmt.Errorf("scvU128: nil")
		}
		u128 := xdr.UInt128Parts{Hi: xdr.Uint64(v.U128.Hi), Lo: xdr.Uint64(v.U128.Lo)}
		return xdr.ScVal{Type: xdr.ScValTypeScvU128, U128: &u128}, nil
	case *scval.ScVal_I128:
		if v.I128 == nil {
			return xdr.ScVal{}, fmt.Errorf("scvI128: nil")
		}
		// Per XDR spec for signed integer types: Hi is signed, Lo is unsigned.
		i128 := xdr.Int128Parts{Hi: xdr.Int64(v.I128.Hi), Lo: xdr.Uint64(v.I128.Lo)}
		return xdr.ScVal{Type: xdr.ScValTypeScvI128, I128: &i128}, nil
	case *scval.ScVal_U256:
		if v.U256 == nil {
			return xdr.ScVal{}, fmt.Errorf("scvU256: nil")
		}
		u256 := xdr.UInt256Parts{
			HiHi: xdr.Uint64(v.U256.HiHi),
			HiLo: xdr.Uint64(v.U256.HiLo),
			LoHi: xdr.Uint64(v.U256.LoHi),
			LoLo: xdr.Uint64(v.U256.LoLo),
		}
		return xdr.ScVal{Type: xdr.ScValTypeScvU256, U256: &u256}, nil
	case *scval.ScVal_I256:
		if v.I256 == nil {
			return xdr.ScVal{}, fmt.Errorf("scvI256: nil")
		}
		// Per XDR spec for signed integer types: HiHi is signed, remaining parts are unsigned.
		i256 := xdr.Int256Parts{
			HiHi: xdr.Int64(v.I256.HiHi),
			HiLo: xdr.Uint64(v.I256.HiLo),
			LoHi: xdr.Uint64(v.I256.LoHi),
			LoLo: xdr.Uint64(v.I256.LoLo),
		}
		return xdr.ScVal{Type: xdr.ScValTypeScvI256, I256: &i256}, nil
	case *scval.ScVal_BytesVal:
		xb := xdr.ScBytes(v.BytesVal)
		return xdr.ScVal{Type: xdr.ScValTypeScvBytes, Bytes: &xb}, nil
	case *scval.ScVal_Str:
		xs := xdr.ScString(v.Str)
		return xdr.ScVal{Type: xdr.ScValTypeScvString, Str: &xs}, nil
	case *scval.ScVal_Sym:
		xs := xdr.ScSymbol(v.Sym)
		return xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &xs}, nil
	case *scval.ScVal_Vec:
		if v.Vec == nil {
			return xdr.ScVal{}, fmt.Errorf("scvVec: nil")
		}
		xVec := make(xdr.ScVec, len(v.Vec.Values))
		for i, pv := range v.Vec.Values {
			xv, err := protoScValToXDRAt(pv, depth+1)
			if err != nil {
				return xdr.ScVal{}, fmt.Errorf("vec[%d]: %w", i, err)
			}
			xVec[i] = xv
		}
		xVecP := &xVec
		return xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &xVecP}, nil
	case *scval.ScVal_Map:
		if v.Map == nil {
			return xdr.ScVal{}, fmt.Errorf("scvMap: nil")
		}
		xMap := make(xdr.ScMap, len(v.Map.Entries))
		for i, pe := range v.Map.Entries {
			xk, err := protoScValToXDRAt(pe.Key, depth+1)
			if err != nil {
				return xdr.ScVal{}, fmt.Errorf("map[%d].key: %w", i, err)
			}
			xv, err := protoScValToXDRAt(pe.Val, depth+1)
			if err != nil {
				return xdr.ScVal{}, fmt.Errorf("map[%d].val: %w", i, err)
			}
			xMap[i] = xdr.ScMapEntry{Key: xk, Val: xv}
		}
		xMapP := &xMap
		return xdr.ScVal{Type: xdr.ScValTypeScvMap, Map: &xMapP}, nil
	case *scval.ScVal_Address:
		xa, err := protoScAddressToXDR(v.Address)
		if err != nil {
			return xdr.ScVal{}, err
		}
		return xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &xa}, nil
	case *scval.ScVal_ContractInstance:
		xi, err := protoScContractInstanceToXDRAt(v.ContractInstance, depth+1)
		if err != nil {
			return xdr.ScVal{}, err
		}
		return xdr.ScVal{Type: xdr.ScValTypeScvContractInstance, Instance: &xi}, nil
	case *scval.ScVal_LedgerKeyContractInstance:
		return xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyContractInstance}, nil
	case *scval.ScVal_NonceKey:
		if v.NonceKey == nil {
			return xdr.ScVal{}, fmt.Errorf("scvLedgerKeyNonce: nil")
		}
		return xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyNonce, NonceKey: &xdr.ScNonceKey{Nonce: xdr.Int64(v.NonceKey.Nonce)}}, nil
	default:
		return xdr.ScVal{}, fmt.Errorf("unsupported proto ScVal type: %T", sv.Value)
	}
}

func protoScErrorToXDR(e *scval.ScError) (xdr.ScError, error) {
	if e == nil {
		return xdr.ScError{}, fmt.Errorf("proto ScError is nil")
	}
	xe := xdr.ScError{Type: xdr.ScErrorType(e.Type)}
	switch v := e.CodeOrContract.(type) {
	case *scval.ScError_ContractCode:
		cc := xdr.Uint32(v.ContractCode)
		xe.ContractCode = &cc
	case *scval.ScError_Code_:
		code := xdr.ScErrorCode(v.Code)
		xe.Code = &code
	default:
		return xdr.ScError{}, fmt.Errorf("unsupported ScError oneof: %T", e.CodeOrContract)
	}
	return xe, nil
}

func protoScAddressToXDR(a *scval.ScAddress) (xdr.ScAddress, error) {
	if a == nil {
		return xdr.ScAddress{}, fmt.Errorf("proto ScAddress is nil")
	}
	switch v := a.Address.(type) {
	case *scval.ScAddress_AccountId:
		if len(v.AccountId) != 32 {
			return xdr.ScAddress{}, fmt.Errorf("accountId must be 32 bytes, got %d", len(v.AccountId))
		}
		var ed25519 xdr.Uint256
		copy(ed25519[:], v.AccountId)
		aid := xdr.AccountId(xdr.PublicKey{Type: xdr.PublicKeyTypePublicKeyTypeEd25519, Ed25519: &ed25519})
		return xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeAccount, AccountId: &aid}, nil
	case *scval.ScAddress_ContractId:
		if len(v.ContractId) != 32 {
			return xdr.ScAddress{}, fmt.Errorf("contractId must be 32 bytes, got %d", len(v.ContractId))
		}
		var cid xdr.ContractId
		copy(cid[:], v.ContractId)
		return xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeContract, ContractId: &cid}, nil
	case *scval.ScAddress_MuxedAccount:
		if v.MuxedAccount == nil {
			return xdr.ScAddress{}, fmt.Errorf("muxedAccount: nil")
		}
		if len(v.MuxedAccount.Ed25519) != 32 {
			return xdr.ScAddress{}, fmt.Errorf("muxedAccount.ed25519 must be 32 bytes, got %d", len(v.MuxedAccount.Ed25519))
		}
		var ed25519 xdr.Uint256
		copy(ed25519[:], v.MuxedAccount.Ed25519)
		return xdr.ScAddress{
			Type: xdr.ScAddressTypeScAddressTypeMuxedAccount,
			MuxedAccount: &xdr.MuxedEd25519Account{
				Id:      xdr.Uint64(v.MuxedAccount.Id),
				Ed25519: ed25519,
			},
		}, nil
	case *scval.ScAddress_ClaimableBalanceId:
		if v.ClaimableBalanceId == nil {
			return xdr.ScAddress{}, fmt.Errorf("claimableBalanceId: nil")
		}
		if len(v.ClaimableBalanceId.V0) != 32 {
			return xdr.ScAddress{}, fmt.Errorf("claimableBalanceId.v0 must be 32 bytes, got %d", len(v.ClaimableBalanceId.V0))
		}
		var h xdr.Hash
		copy(h[:], v.ClaimableBalanceId.V0)
		return xdr.ScAddress{
			Type: xdr.ScAddressTypeScAddressTypeClaimableBalance,
			ClaimableBalanceId: &xdr.ClaimableBalanceId{
				Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
				V0:   &h,
			},
		}, nil
	case *scval.ScAddress_LiquidityPoolId:
		if len(v.LiquidityPoolId) != 32 {
			return xdr.ScAddress{}, fmt.Errorf("liquidityPoolId must be 32 bytes, got %d", len(v.LiquidityPoolId))
		}
		var poolId xdr.PoolId
		copy(poolId[:], v.LiquidityPoolId)
		return xdr.ScAddress{Type: xdr.ScAddressTypeScAddressTypeLiquidityPool, LiquidityPoolId: &poolId}, nil
	default:
		return xdr.ScAddress{}, fmt.Errorf("unsupported proto ScAddress type: %T", a.Address)
	}
}

func protoContractExecutableToXDR(exec *scval.ContractExecutable) (xdr.ContractExecutable, error) {
	if exec == nil {
		return xdr.ContractExecutable{}, fmt.Errorf("proto ContractExecutable is nil")
	}
	switch v := exec.Type.(type) {
	case *scval.ContractExecutable_WasmHash:
		if len(v.WasmHash) != 32 {
			return xdr.ContractExecutable{}, fmt.Errorf("wasmHash must be 32 bytes, got %d", len(v.WasmHash))
		}
		var h xdr.Hash
		copy(h[:], v.WasmHash)
		return xdr.ContractExecutable{Type: xdr.ContractExecutableTypeContractExecutableWasm, WasmHash: &h}, nil
	case *scval.ContractExecutable_StellarAsset:
		return xdr.ContractExecutable{Type: xdr.ContractExecutableTypeContractExecutableStellarAsset}, nil
	default:
		return xdr.ContractExecutable{}, fmt.Errorf("unsupported proto ContractExecutable type: %T", exec.Type)
	}
}

func protoScContractInstanceToXDRAt(inst *scval.ScContractInstance, depth int) (xdr.ScContractInstance, error) {
	if inst == nil {
		return xdr.ScContractInstance{}, fmt.Errorf("proto ScContractInstance is nil")
	}
	xExec, err := protoContractExecutableToXDR(inst.Executable)
	if err != nil {
		return xdr.ScContractInstance{}, err
	}
	xi := xdr.ScContractInstance{Executable: xExec}
	if len(inst.Storage) > 0 {
		xMap := make(xdr.ScMap, len(inst.Storage))
		for i, pe := range inst.Storage {
			xk, err := protoScValToXDRAt(pe.Key, depth+1)
			if err != nil {
				return xdr.ScContractInstance{}, fmt.Errorf("instance.storage[%d].key: %w", i, err)
			}
			xv, err := protoScValToXDRAt(pe.Val, depth+1)
			if err != nil {
				return xdr.ScContractInstance{}, fmt.Errorf("instance.storage[%d].val: %w", i, err)
			}
			xMap[i] = xdr.ScMapEntry{Key: xk, Val: xv}
		}
		xi.Storage = &xMap
	}
	return xi, nil
}

// ConvertReadContractRequestToProto converts a domain ReadContractRequest to its
// proto representation.
func ConvertReadContractRequestToProto(req stellar.ReadContractRequest) (*ReadContractRequest, error) {
	if req.ContractID == "" {
		return nil, fmt.Errorf("contractID is required")
	}
	if req.Function == "" {
		return nil, fmt.Errorf("function is required")
	}
	args := make([]*scval.ScVal, len(req.Args))
	for i, sv := range req.Args {
		psv, err := xdrScValToProto(sv)
		if err != nil {
			return nil, fmt.Errorf("args[%d]: %w", i, err)
		}
		args[i] = psv
	}
	return &ReadContractRequest{
		ContractId:     req.ContractID,
		Function:       req.Function,
		Args:           args,
		LedgerSequence: req.LedgerSequence,
	}, nil
}

// ConvertReadContractRequestFromProto converts a proto ReadContractRequest to the
// domain type.
func ConvertReadContractRequestFromProto(p *ReadContractRequest) (stellar.ReadContractRequest, error) {
	if p == nil {
		return stellar.ReadContractRequest{}, fmt.Errorf("ReadContractRequest is nil")
	}
	if p.GetContractId() == "" {
		return stellar.ReadContractRequest{}, fmt.Errorf("contract_id is required")
	}
	if p.GetFunction() == "" {
		return stellar.ReadContractRequest{}, fmt.Errorf("function is required")
	}
	pArgs := p.GetArgs()
	args := make([]xdr.ScVal, len(pArgs))
	for i, psv := range pArgs {
		sv, err := protoScValToXDR(psv)
		if err != nil {
			return stellar.ReadContractRequest{}, fmt.Errorf("args[%d]: %w", i, err)
		}
		args[i] = sv
	}
	return stellar.ReadContractRequest{
		ContractID:     p.GetContractId(),
		Function:       p.GetFunction(),
		Args:           args,
		LedgerSequence: p.GetLedgerSequence(),
	}, nil
}

// ConvertReadContractResponseToProto converts a domain ReadContractResponse to its
// proto representation.
func ConvertReadContractResponseToProto(resp stellar.ReadContractResponse) (*ReadContractResponse, error) {
	pr := &ReadContractResponse{
		LedgerSequence: resp.LedgerSequence,
		Error:          resp.Error,
	}
	if resp.Result != nil {
		psv, err := xdrScValToProto(*resp.Result)
		if err != nil {
			return nil, fmt.Errorf("result: %w", err)
		}
		pr.Result = psv
	}
	return pr, nil
}

// ConvertReadContractResponseFromProto converts a proto ReadContractResponse to the
// domain type.
func ConvertReadContractResponseFromProto(p *ReadContractResponse) (stellar.ReadContractResponse, error) {
	if p == nil {
		return stellar.ReadContractResponse{}, fmt.Errorf("readContractResponse is nil")
	}
	resp := stellar.ReadContractResponse{
		LedgerSequence: p.GetLedgerSequence(),
		Error:          p.GetError(),
	}
	if p.GetResult() != nil {
		sv, err := protoScValToXDR(p.GetResult())
		if err != nil {
			return stellar.ReadContractResponse{}, fmt.Errorf("result: %w", err)
		}
		resp.Result = &sv
	}
	return resp, nil
}
