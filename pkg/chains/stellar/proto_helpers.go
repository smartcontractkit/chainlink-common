package stellar

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

// xdrToBytes base64-decodes a domain XDR string to raw binary.
// Invalid base64 returns nil silently; malformed XDR should not reach ToProto.
func xdrToBytes(x stellar.XDR) []byte {
	b, _ := base64.StdEncoding.DecodeString(string(x))
	return b
}

// bytesToXDR base64-encodes raw binary XDR to the domain type.
func bytesToXDR(b []byte) stellar.XDR {
	return stellar.XDR(base64.StdEncoding.EncodeToString(b))
}

// hashToBytes hex-decodes a domain hash string to raw bytes.
// Returns nil on invalid hex; should not happen for well-formed hashes.
func hashToBytes(h string) []byte {
	b, _ := hex.DecodeString(h)
	return b
}

// bytesToHash hex-encodes raw hash bytes to a string.
func bytesToHash(b []byte) string {
	return hex.EncodeToString(b)
}

// ---- GetLedgerEntries ----

// ConvertGetLedgerEntriesRequestToProto converts a domain GetLedgerEntriesRequest to its proto representation.
func ConvertGetLedgerEntriesRequestToProto(req stellar.GetLedgerEntriesRequest) *GetLedgerEntriesRequest {
	keys := make([][]byte, len(req.Keys))
	for i, k := range req.Keys {
		keys[i] = xdrToBytes(k)
	}
	return &GetLedgerEntriesRequest{Keys: keys}
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
func ConvertLedgerEntryResultToProto(r stellar.LedgerEntryResult) *LedgerEntryResult {
	pr := &LedgerEntryResult{
		KeyXdr:             xdrToBytes(r.KeyXDR),
		DataXdr:            xdrToBytes(r.DataXDR),
		LastModifiedLedger: r.LastModifiedLedger,
		ExtensionXdr:       xdrToBytes(r.ExtensionXDR),
	}
	if r.LiveUntilLedgerSeq != nil {
		pr.HasLiveUntilLedgerSeq = true
		pr.LiveUntilLedgerSeq = *r.LiveUntilLedgerSeq
	}
	return pr
}

// ConvertLedgerEntryResultFromProto converts a proto LedgerEntryResult to the domain type.
func ConvertLedgerEntryResultFromProto(p *LedgerEntryResult) stellar.LedgerEntryResult {
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
	return r
}

// ConvertGetLedgerEntriesResponseToProto converts a domain GetLedgerEntriesResponse to its proto representation.
func ConvertGetLedgerEntriesResponseToProto(resp stellar.GetLedgerEntriesResponse) *GetLedgerEntriesResponse {
	entries := make([]*LedgerEntryResult, 0, len(resp.Entries))
	for _, e := range resp.Entries {
		entries = append(entries, ConvertLedgerEntryResultToProto(e))
	}
	return &GetLedgerEntriesResponse{
		Entries:      entries,
		LatestLedger: resp.LatestLedger,
	}
}

// ConvertGetLedgerEntriesResponseFromProto converts a proto GetLedgerEntriesResponse to the domain type.
func ConvertGetLedgerEntriesResponseFromProto(p *GetLedgerEntriesResponse) stellar.GetLedgerEntriesResponse {
	entries := make([]stellar.LedgerEntryResult, 0, len(p.GetEntries()))
	for _, pe := range p.GetEntries() {
		entries = append(entries, ConvertLedgerEntryResultFromProto(pe))
	}
	return stellar.GetLedgerEntriesResponse{
		Entries:      entries,
		LatestLedger: p.GetLatestLedger(),
	}
}

// ---- GetLatestLedger ----

// ConvertGetLatestLedgerResponseToProto converts a domain GetLatestLedgerResponse to its proto representation.
func ConvertGetLatestLedgerResponseToProto(resp stellar.GetLatestLedgerResponse) *GetLatestLedgerResponse {
	return &GetLatestLedgerResponse{
		Hash:              hashToBytes(string(resp.Hash)),
		ProtocolVersion:   resp.ProtocolVersion,
		Sequence:          resp.Sequence,
		LedgerCloseTime:   resp.LedgerCloseTime,
		LedgerHeaderXdr:   xdrToBytes(resp.LedgerHeaderXDR),
		LedgerMetadataXdr: xdrToBytes(resp.LedgerMetadataXDR),
	}
}

// ConvertGetLatestLedgerResponseFromProto converts a proto GetLatestLedgerResponse to the domain type.
func ConvertGetLatestLedgerResponseFromProto(p *GetLatestLedgerResponse) stellar.GetLatestLedgerResponse {
	return stellar.GetLatestLedgerResponse{
		Hash:              stellar.LedgerHash(bytesToHash(p.GetHash())),
		ProtocolVersion:   p.GetProtocolVersion(),
		Sequence:          p.GetSequence(),
		LedgerCloseTime:   p.GetLedgerCloseTime(),
		LedgerHeaderXDR:   bytesToXDR(p.GetLedgerHeaderXdr()),
		LedgerMetadataXDR: bytesToXDR(p.GetLedgerMetadataXdr()),
	}
}
