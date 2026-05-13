package stellar

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	stellartypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

// ConvertGetLatestLedgerResponseFromProto converts a proto GetLatestLedgerResponse to the
// domain type. Hash is returned as lowercase hex; XDR fields are returned as standard base64.
func ConvertGetLatestLedgerResponseFromProto(p *GetLatestLedgerResponse) (stellartypes.GetLatestLedgerResponse, error) {
	if p == nil {
		return stellartypes.GetLatestLedgerResponse{}, fmt.Errorf("getLatestLedgerResponse is nil")
	}

	return stellartypes.GetLatestLedgerResponse{
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
func ConvertGetLatestLedgerResponseToProto(r stellartypes.GetLatestLedgerResponse) (*GetLatestLedgerResponse, error) {
	hash, err := hex.DecodeString(r.Hash)
	if err != nil {
		return nil, fmt.Errorf("invalid hex hash %q: %w", r.Hash, err)
	}

	headerXDR, err := base64.StdEncoding.DecodeString(r.LedgerHeaderXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 ledger_header_xdr: %w", err)
	}

	metadataXDR, err := base64.StdEncoding.DecodeString(r.LedgerMetadataXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 ledger_metadata_xdr: %w", err)
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

// ValidateReadContractRequest checks that required fields are present.
func ValidateReadContractRequest(req *ReadContractRequest) error {
	if req == nil {
		return fmt.Errorf("readContractRequest is nil")
	}
	if req.ContractId == "" {
		return fmt.Errorf("contract_id is required")
	}
	if req.Function == "" {
		return fmt.Errorf("function is required")
	}
	return nil
}

// ValidateWriteReportRequest checks that required fields are present.
func ValidateWriteReportRequest(req *WriteReportRequest) error {
	if req == nil {
		return fmt.Errorf("writeReportRequest is nil")
	}
	if req.ContractId == "" {
		return fmt.Errorf("contract_id is required")
	}
	if req.Report == nil {
		return fmt.Errorf("report is required")
	}
	return nil
}
