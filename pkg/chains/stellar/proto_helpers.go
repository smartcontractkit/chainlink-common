package stellar

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/stellar/go-stellar-sdk/xdr"

	v1alpha "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar/scval"

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
	var errs []error
	for i, e := range resp.Entries {
		protoEntry, err := ConvertLedgerEntryResultToProto(e)
		if err != nil {
			errs = append(errs, fmt.Errorf("entry[%d]: %w", i, err))
			continue
		}
		entries = append(entries, protoEntry)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
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
	entries := make([]stellar.LedgerEntryResult, 0, len(p.GetEntries()))
	var errs []error
	for i, pe := range p.GetEntries() {
		e, err := ConvertLedgerEntryResultFromProto(pe)
		if err != nil {
			errs = append(errs, fmt.Errorf("entry[%d]: %w", i, err))
			continue
		}
		entries = append(entries, e)
	}
	if len(errs) > 0 {
		return stellar.GetLedgerEntriesResponse{}, errors.Join(errs...)
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

// ============================================================
// ReadContract converters
// ============================================================

// xdrScValToProto XDR-encodes an xdr.ScVal to binary, then wraps the bytes in a
// proto ScVal with the bytes_val field (SCV_BYTES transport container).
// Both the LOOP client and server unwrap using protoScValToXDR.
func xdrScValToProto(sv xdr.ScVal) (*v1alpha.ScVal, error) {
	var buf bytes.Buffer
	if _, err := xdr.Marshal(&buf, sv); err != nil {
		return nil, fmt.Errorf("failed to XDR-marshal ScVal: %w", err)
	}
	return &v1alpha.ScVal{Value: &v1alpha.ScVal_BytesVal{BytesVal: buf.Bytes()}}, nil
}

// protoScValToXDR unwraps a proto ScVal that was produced by xdrScValToProto and
// XDR-decodes the inner bytes back to an xdr.ScVal.
func protoScValToXDR(sv *v1alpha.ScVal) (xdr.ScVal, error) {
	if sv == nil {
		return xdr.ScVal{}, fmt.Errorf("proto ScVal is nil")
	}
	bv, ok := sv.Value.(*v1alpha.ScVal_BytesVal)
	if !ok {
		return xdr.ScVal{}, fmt.Errorf("proto ScVal does not carry bytes_val (got %T)", sv.Value)
	}
	var out xdr.ScVal
	if _, err := xdr.Unmarshal(bytes.NewReader(bv.BytesVal), &out); err != nil {
		return xdr.ScVal{}, fmt.Errorf("failed to XDR-unmarshal ScVal: %w", err)
	}
	return out, nil
}

// ConvertReadContractRequestToProto converts a domain ReadContractRequest to the
// proto representation.  Each xdr.ScVal arg is XDR-encoded and stored as
// bytes_val inside a proto ScVal for transport.
func ConvertReadContractRequestToProto(req stellar.ReadContractRequest) (*ReadContractRequest, error) {
	if req.ContractID == "" {
		return nil, fmt.Errorf("contract_id is required")
	}
	if req.Function == "" {
		return nil, fmt.Errorf("function is required")
	}
	args := make([]*v1alpha.ScVal, 0, len(req.Args))
	for i, sv := range req.Args {
		psv, err := xdrScValToProto(sv)
		if err != nil {
			return nil, fmt.Errorf("args[%d]: %w", i, err)
		}
		args = append(args, psv)
	}
	return &ReadContractRequest{
		ContractId:     req.ContractID,
		Function:       req.Function,
		Args:           args,
		LedgerSequence: req.LedgerSequence,
	}, nil
}

// ConvertReadContractRequestFromProto converts a proto ReadContractRequest to the
// domain type.  Each bytes_val proto ScVal is XDR-decoded back to xdr.ScVal.
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
	args := make([]xdr.ScVal, 0, len(p.GetArgs()))
	for i, psv := range p.GetArgs() {
		sv, err := protoScValToXDR(psv)
		if err != nil {
			return stellar.ReadContractRequest{}, fmt.Errorf("args[%d]: %w", i, err)
		}
		args = append(args, sv)
	}
	return stellar.ReadContractRequest{
		ContractID:     p.GetContractId(),
		Function:       p.GetFunction(),
		Args:           args,
		LedgerSequence: p.GetLedgerSequence(),
	}, nil
}

// ConvertReadContractResponseToProto converts a domain ReadContractResponse to its
// proto representation.  If resp.Result is non-nil it is XDR-encoded and stored as
// bytes_val inside a proto ScVal.
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
// domain type.  The bytes_val proto ScVal result is XDR-decoded back to xdr.ScVal.
func ConvertReadContractResponseFromProto(p *ReadContractResponse) (stellar.ReadContractResponse, error) {
	if p == nil {
		return stellar.ReadContractResponse{}, fmt.Errorf("ReadContractResponse is nil")
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

// ============================================================
// ScVal builder helpers
//
// These helpers produce proto-marshalled ScVal bytes ([]byte)
// suitable for ReadContractRequest.Args / ReadContractResponse.Result.
// ============================================================

// MarshalScVal proto-marshals the given ScVal and returns the bytes, or an error.
func MarshalScVal(sv *v1alpha.ScVal) ([]byte, error) {
	if sv == nil {
		return nil, fmt.Errorf("ScVal is nil")
	}
	return proto.Marshal(sv)
}

// UnmarshalScVal unmarshals a proto-encoded ScVal from raw bytes.
func UnmarshalScVal(raw []byte) (*v1alpha.ScVal, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("raw ScVal bytes are empty")
	}
	sv := &v1alpha.ScVal{}
	if err := proto.Unmarshal(raw, sv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ScVal: %w", err)
	}
	return sv, nil
}

// ScValBool creates a proto-marshalled SCV_BOOL ScVal.
func ScValBool(v bool) ([]byte, error) {
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_B{B: v}})
}

// ScValVoid creates a proto-marshalled SCV_VOID ScVal.
func ScValVoid() ([]byte, error) {
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_VoidVal{VoidVal: &v1alpha.Void{}}})
}

// ScValU32 creates a proto-marshalled SCV_U32 ScVal.
func ScValU32(v uint32) ([]byte, error) {
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_U32{U32: v}})
}

// ScValI32 creates a proto-marshalled SCV_I32 ScVal.
func ScValI32(v int32) ([]byte, error) {
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_I32{I32: v}})
}

// ScValU64 creates a proto-marshalled SCV_U64 ScVal.
func ScValU64(v uint64) ([]byte, error) {
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_U64{U64: v}})
}

// ScValI64 creates a proto-marshalled SCV_I64 ScVal.
func ScValI64(v int64) ([]byte, error) {
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_I64{I64: v}})
}

// ScValBytes creates a proto-marshalled SCV_BYTES ScVal.
func ScValBytes(v []byte) ([]byte, error) {
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_BytesVal{BytesVal: v}})
}

// ScValString creates a proto-marshalled SCV_STRING ScVal.
func ScValString(v string) ([]byte, error) {
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_Str{Str: v}})
}

// ScValSymbol creates a proto-marshalled SCV_SYMBOL ScVal.
// Stellar symbols are limited to 32 characters.
func ScValSymbol(v string) ([]byte, error) {
	if len(v) > 32 {
		return nil, fmt.Errorf("symbol %q exceeds 32-character limit (%d chars)", v, len(v))
	}
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_Sym{Sym: v}})
}

// ScValAddress creates a proto-marshalled SCV_ADDRESS ScVal.
func ScValAddress(addr *v1alpha.ScAddress) ([]byte, error) {
	if addr == nil {
		return nil, fmt.Errorf("ScAddress is nil")
	}
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_Address{Address: addr}})
}

// ScValAccountAddress creates a proto-marshalled SCV_ADDRESS ScVal from a 32-byte
// Ed25519 public key (account address).
func ScValAccountAddress(pubKey []byte) ([]byte, error) {
	if len(pubKey) != 32 {
		return nil, fmt.Errorf("account public key must be 32 bytes, got %d", len(pubKey))
	}
	return ScValAddress(&v1alpha.ScAddress{
		Address: &v1alpha.ScAddress_AccountId{AccountId: pubKey},
	})
}

// ScValContractAddress creates a proto-marshalled SCV_ADDRESS ScVal from a 32-byte
// contract hash.
func ScValContractAddress(contractHash []byte) ([]byte, error) {
	if len(contractHash) != 32 {
		return nil, fmt.Errorf("contract hash must be 32 bytes, got %d", len(contractHash))
	}
	return ScValAddress(&v1alpha.ScAddress{
		Address: &v1alpha.ScAddress_ContractId{ContractId: contractHash},
	})
}

// ScValVec creates a proto-marshalled SCV_VEC ScVal from a slice of already
// proto-marshalled ScVal elements (as returned by the other helpers).
func ScValVec(elements [][]byte) ([]byte, error) {
	vals := make([]*v1alpha.ScVal, 0, len(elements))
	for i, raw := range elements {
		sv, err := UnmarshalScVal(raw)
		if err != nil {
			return nil, fmt.Errorf("element[%d]: %w", i, err)
		}
		vals = append(vals, sv)
	}
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_Vec{Vec: &v1alpha.ScVec{Values: vals}}})
}

// ScValMap creates a proto-marshalled SCV_MAP ScVal.  keys and values must be
// slices of the same length, each element being a proto-marshalled ScVal.
func ScValMap(keys, values [][]byte) ([]byte, error) {
	if len(keys) != len(values) {
		return nil, fmt.Errorf("keys (%d) and values (%d) must have the same length", len(keys), len(values))
	}
	entries := make([]*v1alpha.ScMapEntry, 0, len(keys))
	for i := range keys {
		k, err := UnmarshalScVal(keys[i])
		if err != nil {
			return nil, fmt.Errorf("key[%d]: %w", i, err)
		}
		v, err := UnmarshalScVal(values[i])
		if err != nil {
			return nil, fmt.Errorf("value[%d]: %w", i, err)
		}
		entries = append(entries, &v1alpha.ScMapEntry{Key: k, Val: v})
	}
	return MarshalScVal(&v1alpha.ScVal{Value: &v1alpha.ScVal_Map{Map: &v1alpha.ScMap{Entries: entries}}})
}
