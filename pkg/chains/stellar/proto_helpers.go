package stellar

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/stellar/go-stellar-sdk/xdr"

	stellarcap "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar"
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
		return stellar.GetLedgerEntriesResponse{}, errors.New("get ledger entries response is nil")
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
		psv, err := stellarcap.XdrScValToProto(sv)
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
		return stellar.ReadContractRequest{}, fmt.Errorf("readContractRequest is nil")
	}
	if p.GetContractId() == "" {
		return stellar.ReadContractRequest{}, fmt.Errorf("contractID is required")
	}
	if p.GetFunction() == "" {
		return stellar.ReadContractRequest{}, fmt.Errorf("function is required")
	}
	pArgs := p.GetArgs()
	args := make([]xdr.ScVal, len(pArgs))
	for i, psv := range pArgs {
		sv, err := stellarcap.ProtoScValToXDR(psv)
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

// ConvertSubmitTransactionRequestToProto converts a domain SubmitTransactionRequest to its proto representation.
func ConvertSubmitTransactionRequestToProto(req stellar.SubmitTransactionRequest) (*SubmitTransactionRequest, error) {
	return &SubmitTransactionRequest{
		Id:                 req.ID,
		FromAddress:        req.FromAddress,
		OperationsXdr:      req.OperationsXDR,
		LedgerBoundsOffset: req.LedgerBoundsOffset,
	}, nil
}

// ConvertSubmitTransactionRequestFromProto converts a proto SubmitTransactionRequest to the domain type.
func ConvertSubmitTransactionRequestFromProto(p *SubmitTransactionRequest) (stellar.SubmitTransactionRequest, error) {
	if p == nil {
		return stellar.SubmitTransactionRequest{}, fmt.Errorf("submit transaction request is nil")
	}
	return stellar.SubmitTransactionRequest{
		ID:                 p.GetId(),
		FromAddress:        p.GetFromAddress(),
		OperationsXDR:      p.GetOperationsXdr(),
		LedgerBoundsOffset: p.GetLedgerBoundsOffset(),
	}, nil
}

// ConvertSubmitTransactionResponseToProto converts a domain SubmitTransactionResponse to its proto representation.
func ConvertSubmitTransactionResponseToProto(resp *stellar.SubmitTransactionResponse) (*SubmitTransactionResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("submit transaction response is nil")
	}
	return &SubmitTransactionResponse{
		TxHash: resp.TxHash,
	}, nil
}

// ConvertSubmitTransactionResponseFromProto converts a proto SubmitTransactionResponse to the domain type.
func ConvertSubmitTransactionResponseFromProto(p *SubmitTransactionResponse) (*stellar.SubmitTransactionResponse, error) {
	if p == nil {
		return nil, fmt.Errorf("submit transaction response is nil")
	}
	return &stellar.SubmitTransactionResponse{
		TxHash: p.GetTxHash(),
	}, nil
}

// ConvertSimulateTransactionRequestToProto converts a domain SimulateTransactionRequest to its proto representation.
func ConvertSimulateTransactionRequestToProto(req stellar.SimulateTransactionRequest) (*SimulateTransactionRequest, error) {
	return &SimulateTransactionRequest{
		FromAddress:        req.FromAddress,
		OperationsXdr:      req.OperationsXDR,
		LedgerBoundsOffset: req.LedgerBoundsOffset,
	}, nil
}

// ConvertSimulateTransactionRequestFromProto converts a proto SimulateTransactionRequest to the domain type.
func ConvertSimulateTransactionRequestFromProto(p *SimulateTransactionRequest) (stellar.SimulateTransactionRequest, error) {
	if p == nil {
		return stellar.SimulateTransactionRequest{}, fmt.Errorf("simulate transaction request is nil")
	}
	return stellar.SimulateTransactionRequest{
		FromAddress:        p.GetFromAddress(),
		OperationsXDR:      p.GetOperationsXdr(),
		LedgerBoundsOffset: p.GetLedgerBoundsOffset(),
	}, nil
}

// ConvertSimulateTransactionResponseToProto converts a domain SimulateTransactionResponse to its proto representation.
func ConvertSimulateTransactionResponseToProto(resp stellar.SimulateTransactionResponse) (*SimulateTransactionResponse, error) {
	return &SimulateTransactionResponse{
		Error:          resp.Error,
		MinResourceFee: resp.MinResourceFee,
		Results:        resp.Results,
	}, nil
}

// ConvertSimulateTransactionResponseFromProto converts a proto SimulateTransactionResponse to the domain type.
func ConvertSimulateTransactionResponseFromProto(p *SimulateTransactionResponse) (stellar.SimulateTransactionResponse, error) {
	if p == nil {
		return stellar.SimulateTransactionResponse{}, fmt.Errorf("simulate transaction response is nil")
	}
	return stellar.SimulateTransactionResponse{
		Error:          p.GetError(),
		MinResourceFee: p.GetMinResourceFee(),
		Results:        p.GetResults(),
	}, nil
}

// ConvertTxResultToProto converts a domain TxResult to its proto representation.
func ConvertTxResultToProto(res stellar.TxResult) (*GetTransactionResultResponse, error) {
	return &GetTransactionResultResponse{
		Id:            res.ID,
		Hash:          res.Hash,
		Status:        res.Status,
		Fee:           res.Fee,
		ResultXdr:     res.ResultXDR,
		ResultMetaXdr: res.ResultMetaXDR,
		Error:         res.Error,
	}, nil
}

// ConvertTxResultFromProto converts a proto GetTransactionResultResponse to the domain type.
func ConvertTxResultFromProto(p *GetTransactionResultResponse) (stellar.TxResult, error) {
	if p == nil {
		return stellar.TxResult{}, fmt.Errorf("transaction result response is nil")
	}
	return stellar.TxResult{
		ID:            p.GetId(),
		Hash:          p.GetHash(),
		Status:        p.GetStatus(),
		Fee:           p.GetFee(),
		ResultXDR:     p.GetResultXdr(),
		ResultMetaXDR: p.GetResultMetaXdr(),
		Error:         p.GetError(),
	}, nil
}
