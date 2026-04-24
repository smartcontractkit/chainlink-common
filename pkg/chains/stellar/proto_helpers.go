package stellar

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

// xdrToBytes base64-decodes a domain XDR string to raw binary.
func xdrToBytes(x stellar.XDR) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(string(x))
	if err != nil {
		return nil, fmt.Errorf("invalid base64 XDR %q: %w", x, err)
	}
	return b, nil
}

// bytesToXDR base64-encodes raw binary XDR to the domain type.
func bytesToXDR(b []byte) stellar.XDR {
	return stellar.XDR(base64.StdEncoding.EncodeToString(b))
}

// hashToBytes hex-decodes a domain hash string to raw bytes.
func hashToBytes(h string) ([]byte, error) {
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil, fmt.Errorf("invalid hex hash %q: %w", h, err)
	}
	return b, nil
}

// bytesToHash hex-encodes raw hash bytes to a string.
func bytesToHash(b []byte) string {
	return hex.EncodeToString(b)
}

// ---- GetLedgerEntries ----

// ConvertGetLedgerEntriesRequestToProto converts a domain GetLedgerEntriesRequest to its proto representation.
func ConvertGetLedgerEntriesRequestToProto(req stellar.GetLedgerEntriesRequest) (*GetLedgerEntriesRequest, error) {
	keys := make([][]byte, len(req.Keys))
	var errs []error
	for i, k := range req.Keys {
		b, err := xdrToBytes(k)
		if err != nil {
			errs = append(errs, fmt.Errorf("key[%d]: %w", i, err))
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
	keys := make([]stellar.XDR, len(rawKeys))
	for i, k := range rawKeys {
		keys[i] = bytesToXDR(k)
	}
	return stellar.GetLedgerEntriesRequest{Keys: keys}, nil
}

// ConvertLedgerEntryResultToProto converts a domain LedgerEntryResult to its proto representation.
func ConvertLedgerEntryResultToProto(r stellar.LedgerEntryResult) (*LedgerEntryResult, error) {
	keyXDR, err := xdrToBytes(r.KeyXDR)
	if err != nil {
		return nil, fmt.Errorf("key_xdr: %w", err)
	}
	dataXDR, err := xdrToBytes(r.DataXDR)
	if err != nil {
		return nil, fmt.Errorf("data_xdr: %w", err)
	}
	extXDR, err := xdrToBytes(r.ExtensionXDR)
	if err != nil {
		return nil, fmt.Errorf("extension_xdr: %w", err)
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
		KeyXDR:             bytesToXDR(p.GetKeyXdr()),
		DataXDR:            bytesToXDR(p.GetDataXdr()),
		LastModifiedLedger: p.GetLastModifiedLedger(),
		ExtensionXDR:       bytesToXDR(p.GetExtensionXdr()),
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

// ---- GetLatestLedger ----

// ConvertGetLatestLedgerResponseToProto converts a domain GetLatestLedgerResponse to its proto representation.
func ConvertGetLatestLedgerResponseToProto(resp stellar.GetLatestLedgerResponse) (*GetLatestLedgerResponse, error) {
	hash, err := hashToBytes(string(resp.Hash))
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}
	headerXDR, err := xdrToBytes(resp.LedgerHeaderXDR)
	if err != nil {
		return nil, fmt.Errorf("ledger_header_xdr: %w", err)
	}
	metaXDR, err := xdrToBytes(resp.LedgerMetadataXDR)
	if err != nil {
		return nil, fmt.Errorf("ledger_metadata_xdr: %w", err)
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
		Hash:              stellar.LedgerHash(bytesToHash(p.GetHash())),
		ProtocolVersion:   p.GetProtocolVersion(),
		Sequence:          p.GetSequence(),
		LedgerCloseTime:   p.GetLedgerCloseTime(),
		LedgerHeaderXDR:   bytesToXDR(p.GetLedgerHeaderXdr()),
		LedgerMetadataXDR: bytesToXDR(p.GetLedgerMetadataXdr()),
	}, nil
}
