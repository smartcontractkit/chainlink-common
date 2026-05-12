package stellar

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

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
		return stellar.GetLedgerEntriesRequest{}, errors.New("get ledger entries request is nil")
	}
	if len(p.GetKeys()) == 0 {
		return stellar.GetLedgerEntriesRequest{}, errors.New("ledger entry keys are empty")
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
		return stellar.LedgerEntryResult{}, errors.New("ledger entry result is nil")
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
		return stellar.GetLedgerEntriesResponse{}, errors.New("get ledger entries response is nil")
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
		return stellar.GetLatestLedgerResponse{}, errors.New("get latest ledger response is nil")
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
