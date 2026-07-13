package stellar

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

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

// ConvertGetLedgersRequestToProto converts a domain GetLedgersRequest to its proto representation.
func ConvertGetLedgersRequestToProto(req stellar.GetLedgersRequest) (*GetLedgersRequest, error) {
	return &GetLedgersRequest{
		StartLedger: req.StartLedger,
		Pagination:  ledgerPaginationToProto(req.Pagination),
	}, nil
}

// ConvertGetLedgersRequestFromProto converts a proto GetLedgersRequest to the domain type.
func ConvertGetLedgersRequestFromProto(p *GetLedgersRequest) (stellar.GetLedgersRequest, error) {
	if p == nil {
		return stellar.GetLedgersRequest{}, errors.New("get ledgers request is nil")
	}
	return stellar.GetLedgersRequest{
		StartLedger: p.GetStartLedger(),
		Pagination:  ledgerPaginationFromProto(p.GetPagination()),
	}, nil
}

// ConvertLedgerInfoToProto converts a domain LedgerInfo to its proto representation.
func ConvertLedgerInfoToProto(l stellar.LedgerInfo) (*LedgerInfo, error) {
	hash, err := hex.DecodeString(l.Hash)
	if err != nil {
		return nil, fmt.Errorf("invalid hex hash %q: %w", l.Hash, err)
	}
	headerXDR, err := base64.StdEncoding.DecodeString(l.LedgerHeaderXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid ledger header xdr %q: %w", l.LedgerHeaderXDR, err)
	}
	metaXDR, err := base64.StdEncoding.DecodeString(l.LedgerMetadataXDR)
	if err != nil {
		return nil, fmt.Errorf("invalid ledger metadata xdr %q: %w", l.LedgerMetadataXDR, err)
	}
	return &LedgerInfo{
		Hash:              hash,
		Sequence:          l.Sequence,
		LedgerCloseTime:   l.LedgerCloseTime,
		LedgerHeaderXdr:   headerXDR,
		LedgerMetadataXdr: metaXDR,
	}, nil
}

// ConvertLedgerInfoFromProto converts a proto LedgerInfo to the domain type.
func ConvertLedgerInfoFromProto(p *LedgerInfo) (stellar.LedgerInfo, error) {
	if p == nil {
		return stellar.LedgerInfo{}, errors.New("ledger info is nil")
	}
	return stellar.LedgerInfo{
		Hash:              hex.EncodeToString(p.GetHash()),
		Sequence:          p.GetSequence(),
		LedgerCloseTime:   p.GetLedgerCloseTime(),
		LedgerHeaderXDR:   base64.StdEncoding.EncodeToString(p.GetLedgerHeaderXdr()),
		LedgerMetadataXDR: base64.StdEncoding.EncodeToString(p.GetLedgerMetadataXdr()),
	}, nil
}

// ConvertGetLedgersResponseToProto converts a domain GetLedgersResponse to its proto representation.
func ConvertGetLedgersResponseToProto(resp stellar.GetLedgersResponse) (*GetLedgersResponse, error) {
	ledgers := make([]*LedgerInfo, 0, len(resp.Ledgers))
	for i, l := range resp.Ledgers {
		pl, err := ConvertLedgerInfoToProto(l)
		if err != nil {
			return nil, fmt.Errorf("ledgers[%d]: %w", i, err)
		}
		ledgers = append(ledgers, pl)
	}
	return &GetLedgersResponse{
		Ledgers:               ledgers,
		LatestLedger:          resp.LatestLedger,
		LatestLedgerCloseTime: resp.LatestLedgerCloseTime,
		OldestLedger:          resp.OldestLedger,
		OldestLedgerCloseTime: resp.OldestLedgerCloseTime,
		Cursor:                resp.Cursor,
	}, nil
}

// ConvertGetLedgersResponseFromProto converts a proto GetLedgersResponse to the domain type.
func ConvertGetLedgersResponseFromProto(p *GetLedgersResponse) (stellar.GetLedgersResponse, error) {
	if p == nil {
		return stellar.GetLedgersResponse{}, errors.New("get ledgers response is nil")
	}
	pLedgers := p.GetLedgers()
	ledgers := make([]stellar.LedgerInfo, 0, len(pLedgers))
	for i, pl := range pLedgers {
		l, err := ConvertLedgerInfoFromProto(pl)
		if err != nil {
			return stellar.GetLedgersResponse{}, fmt.Errorf("ledgers[%d]: %w", i, err)
		}
		ledgers = append(ledgers, l)
	}
	return stellar.GetLedgersResponse{
		Ledgers:               ledgers,
		LatestLedger:          p.GetLatestLedger(),
		LatestLedgerCloseTime: p.GetLatestLedgerCloseTime(),
		OldestLedger:          p.GetOldestLedger(),
		OldestLedgerCloseTime: p.GetOldestLedgerCloseTime(),
		Cursor:                p.GetCursor(),
	}, nil
}

func ledgerPaginationToProto(p *stellar.LedgerPaginationOptions) *LedgerPaginationOptions {
	if p == nil {
		return nil
	}
	return &LedgerPaginationOptions{Cursor: p.Cursor, Limit: p.Limit}
}

func ledgerPaginationFromProto(p *LedgerPaginationOptions) *stellar.LedgerPaginationOptions {
	if p == nil {
		return nil
	}
	return &stellar.LedgerPaginationOptions{Cursor: p.GetCursor(), Limit: p.GetLimit()}
}

// ConvertSimulateTransactionRequestToProto converts a domain SimulateTransactionRequest to proto.
func ConvertSimulateTransactionRequestToProto(req stellar.SimulateTransactionRequest) (*SimulateTransactionRequest, error) {
	if req.ContractID == "" {
		return nil, errors.New("contractID is required")
	}
	if req.Function == "" {
		return nil, errors.New("function is required")
	}
	if err := validateSimulateAuthMode(req.AuthMode); err != nil {
		return nil, err
	}

	args, err := scValsToProto("args", req.Args)
	if err != nil {
		return nil, err
	}

	return &SimulateTransactionRequest{
		ContractId:     req.ContractID,
		Function:       req.Function,
		Args:           args,
		SourceAccount:  req.SourceAccount,
		AuthMode:       string(req.AuthMode),
		ResourceConfig: simulateResourceConfigToProto(req.ResourceConfig),
	}, nil
}

// ConvertSimulateTransactionRequestFromProto converts proto SimulateTransactionRequest to domain.
func ConvertSimulateTransactionRequestFromProto(p *SimulateTransactionRequest) (stellar.SimulateTransactionRequest, error) {
	if p == nil {
		return stellar.SimulateTransactionRequest{}, errors.New("simulateTransaction request is nil")
	}
	if p.GetContractId() == "" {
		return stellar.SimulateTransactionRequest{}, errors.New("contractID is required")
	}
	if p.GetFunction() == "" {
		return stellar.SimulateTransactionRequest{}, errors.New("function is required")
	}

	authMode := stellar.SimulateAuthMode(p.GetAuthMode())
	if err := validateSimulateAuthMode(authMode); err != nil {
		return stellar.SimulateTransactionRequest{}, err
	}

	args, err := scValsFromProto("args", p.GetArgs())
	if err != nil {
		return stellar.SimulateTransactionRequest{}, err
	}

	return stellar.SimulateTransactionRequest{
		ContractID:     p.GetContractId(),
		Function:       p.GetFunction(),
		Args:           args,
		SourceAccount:  p.GetSourceAccount(),
		AuthMode:       authMode,
		ResourceConfig: simulateResourceConfigFromProto(p.GetResourceConfig()),
	}, nil
}

func ConvertSimulateTransactionResponseToProto(resp stellar.SimulateTransactionResponse) *SimulateTransactionResponse {
	protoResp := &SimulateTransactionResponse{
		LedgerSequence:     resp.LedgerSequence,
		Success:            resp.Success,
		Error:              resp.Error,
		ReturnValueXdr:     resp.ReturnValueXDR,
		RequiredAuthXdr:    resp.RequiredAuthXDR,
		EventsXdr:          resp.EventsXDR,
		TransactionDataXdr: resp.TransactionDataXDR,
		MinResourceFee:     resp.MinResourceFee,
	}

	if resp.RestorePreamble != nil {
		protoResp.RestorePreamble = &SimulateRestorePreamble{
			TransactionDataXdr: resp.RestorePreamble.TransactionDataXDR,
			MinResourceFee:     resp.RestorePreamble.MinResourceFee,
		}
	}
	return protoResp
}

func ConvertSubmitTransactionRequestToProto(req stellar.SubmitTransactionRequest) (*SubmitTransactionRequest, error) {
	if req.ContractID == "" {
		return nil, errors.New("contractId is required")
	}
	if req.Function == "" {
		return nil, errors.New("function is required")
	}

	args, err := scValsToProto("args", req.Args)
	if err != nil {
		return nil, err
	}

	return &SubmitTransactionRequest{
		IdempotencyKey:     req.IdempotencyKey,
		FromAddress:        req.FromAddress,
		ContractId:         req.ContractID,
		Function:           req.Function,
		Args:               args,
		LedgerBoundsOffset: req.LedgerBoundsOffset,
	}, nil
}

func ConvertSubmitTransactionRequestFromProto(p *SubmitTransactionRequest) (stellar.SubmitTransactionRequest, error) {
	if p == nil {
		return stellar.SubmitTransactionRequest{}, errors.New("submit transaction request is nil")
	}
	if p.GetContractId() == "" {
		return stellar.SubmitTransactionRequest{}, errors.New("contractId is required")
	}
	if p.GetFunction() == "" {
		return stellar.SubmitTransactionRequest{}, errors.New("function is required")
	}

	args, err := scValsFromProto("args", p.GetArgs())
	if err != nil {
		return stellar.SubmitTransactionRequest{}, err
	}

	return stellar.SubmitTransactionRequest{
		IdempotencyKey:     p.GetIdempotencyKey(),
		FromAddress:        p.GetFromAddress(),
		ContractID:         p.GetContractId(),
		Function:           p.GetFunction(),
		Args:               args,
		LedgerBoundsOffset: p.GetLedgerBoundsOffset(),
	}, nil
}

// ConvertSubmitTransactionResponseToProto converts a domain SubmitTransactionResponse to proto.
func ConvertSubmitTransactionResponseToProto(reply *stellar.SubmitTransactionResponse) (*SubmitTransactionResponse, error) {
	if reply == nil {
		return nil, errors.New("submit transaction reply is nil")
	}
	var resultXDR, resultMetaXDR []byte
	if reply.ResultXDR != "" {
		var err error
		resultXDR, err = base64.StdEncoding.DecodeString(reply.ResultXDR)
		if err != nil {
			return nil, fmt.Errorf("invalid result xdr %q: %w", reply.ResultXDR, err)
		}
	}
	if reply.ResultMetaXDR != "" {
		var err error
		resultMetaXDR, err = base64.StdEncoding.DecodeString(reply.ResultMetaXDR)
		if err != nil {
			return nil, fmt.Errorf("invalid result meta xdr %q: %w", reply.ResultMetaXDR, err)
		}
	}
	resp := &SubmitTransactionResponse{
		TxStatus:         TxStatus(reply.TxStatus),
		TxHash:           reply.TxHash,
		TxIdempotencyKey: reply.TxIdempotencyKey,
		ResultXdr:        resultXDR,
		ResultMetaXdr:    resultMetaXDR,
		Error:            reply.Error,
	}
	if reply.TransactionFee != nil {
		resp.TransactionFee = reply.TransactionFee
	}
	if reply.BlockTimestamp != nil {
		resp.BlockTimestamp = reply.BlockTimestamp
	}
	return resp, nil
}

// ConvertSubmitTransactionResponseFromProto converts proto SubmitTransactionResponse to domain.
func ConvertSubmitTransactionResponseFromProto(p *SubmitTransactionResponse) (*stellar.SubmitTransactionResponse, error) {
	if p == nil {
		return nil, errors.New("submit transaction reply is nil")
	}
	resp := &stellar.SubmitTransactionResponse{
		TxStatus:         stellar.TransactionStatus(p.GetTxStatus()),
		TxHash:           p.GetTxHash(),
		TxIdempotencyKey: p.GetTxIdempotencyKey(),
		ResultXDR:        base64.StdEncoding.EncodeToString(p.GetResultXdr()),
		ResultMetaXDR:    base64.StdEncoding.EncodeToString(p.GetResultMetaXdr()),
		Error:            p.GetError(),
	}
	if p.TransactionFee != nil {
		fee := p.GetTransactionFee()
		resp.TransactionFee = &fee
	}
	if p.BlockTimestamp != nil {
		ts := p.GetBlockTimestamp()
		resp.BlockTimestamp = &ts
	}
	return resp, nil
}

func ConvertGetEventsRequestToProto(req stellar.GetEventsRequest) (*GetEventsRequest, error) {
	filters := make([]*EventFilter, len(req.Filters))
	for i, f := range req.Filters {
		pf, err := convertEventFilterToProto(f)
		if err != nil {
			return nil, fmt.Errorf("filters[%d]: %w", i, err)
		}
		filters[i] = pf
	}

	return &GetEventsRequest{
		StartLedger: req.StartLedger,
		EndLedger:   req.EndLedger,
		Filters:     filters,
		Pagination:  paginationToProto(req.Pagination),
	}, nil
}

func ConvertGetEventsRequestFromProto(p *GetEventsRequest) (stellar.GetEventsRequest, error) {
	if p == nil {
		return stellar.GetEventsRequest{}, errors.New("get events request is nil")
	}

	pFilters := p.GetFilters()
	filters := make([]stellar.EventFilter, len(pFilters))
	for i, pf := range pFilters {
		f, err := convertEventFilterFromProto(pf)
		if err != nil {
			return stellar.GetEventsRequest{}, fmt.Errorf("filters[%d]: %w", i, err)
		}
		filters[i] = f
	}

	return stellar.GetEventsRequest{
		StartLedger: p.GetStartLedger(),
		EndLedger:   p.GetEndLedger(),
		Filters:     filters,
		Pagination:  paginationFromProto(p.GetPagination()),
	}, nil
}

func convertEventFilterToProto(f stellar.EventFilter) (*EventFilter, error) {
	eventTypes, err := eventTypesToProto(f.EventTypes)
	if err != nil {
		return nil, err
	}

	topics := make([]*TopicFilter, len(f.Topics))
	for i, t := range f.Topics {
		pt, err := convertTopicFilterToProto(t)
		if err != nil {
			return nil, fmt.Errorf("topics[%d]: %w", i, err)
		}
		topics[i] = pt
	}

	return &EventFilter{
		EventTypes:  eventTypes,
		ContractIds: append([]string(nil), f.ContractIDs...),
		Topics:      topics,
	}, nil
}

func convertEventFilterFromProto(p *EventFilter) (stellar.EventFilter, error) {
	if p == nil {
		return stellar.EventFilter{}, errors.New("event filter is nil")
	}

	eventTypes, err := eventTypesFromProto(p.GetEventTypes())
	if err != nil {
		return stellar.EventFilter{}, err
	}

	pTopics := p.GetTopics()
	topics := make([]stellar.TopicFilter, len(pTopics))
	for i, pt := range pTopics {
		t, err := convertTopicFilterFromProto(pt)
		if err != nil {
			return stellar.EventFilter{}, fmt.Errorf("topics[%d]: %w", i, err)
		}
		topics[i] = t
	}

	return stellar.EventFilter{
		EventTypes:  eventTypes,
		ContractIDs: append([]string(nil), p.GetContractIds()...),
		Topics:      topics,
	}, nil
}

func convertTopicFilterToProto(t stellar.TopicFilter) (*TopicFilter, error) {
	if len(t.Segments) == 0 {
		return nil, errors.New("topic filter must have at least one segment")
	}

	segments := make([]*TopicSegment, len(t.Segments))
	for i, s := range t.Segments {
		ps, err := convertTopicSegmentToProto(s)
		if err != nil {
			return nil, fmt.Errorf("segments[%d]: %w", i, err)
		}
		segments[i] = ps
	}

	return &TopicFilter{Segments: segments}, nil
}

func convertTopicFilterFromProto(p *TopicFilter) (stellar.TopicFilter, error) {
	if p == nil {
		return stellar.TopicFilter{}, errors.New("topic filter is nil")
	}
	if len(p.GetSegments()) == 0 {
		return stellar.TopicFilter{}, errors.New("topic filter must have at least one segment")
	}

	pSegments := p.GetSegments()
	segments := make([]stellar.TopicSegment, len(pSegments))
	for i, ps := range pSegments {
		s, err := convertTopicSegmentFromProto(ps)
		if err != nil {
			return stellar.TopicFilter{}, fmt.Errorf("segments[%d]: %w", i, err)
		}
		segments[i] = s
	}

	return stellar.TopicFilter{Segments: segments}, nil
}

func convertTopicSegmentToProto(s stellar.TopicSegment) (*TopicSegment, error) {
	switch {
	case s.Wildcard != nil && s.Value != nil:
		return nil, errors.New("topic segment cannot set both wildcard and value")
	case s.Wildcard != nil:
		if err := validateTopicWildcard(*s.Wildcard); err != nil {
			return nil, err
		}
		return &TopicSegment{
			Value: &TopicSegment_Wildcard{Wildcard: *s.Wildcard},
		}, nil
	case s.Value != nil:
		psv, err := stellarcap.ScValToProto(*s.Value)
		if err != nil {
			return nil, err
		}
		return &TopicSegment{
			Value: &TopicSegment_Scval{Scval: psv},
		}, nil
	default:
		return nil, errors.New("topic segment must set either wildcard or value")
	}
}

func convertTopicSegmentFromProto(p *TopicSegment) (stellar.TopicSegment, error) {
	if p == nil {
		return stellar.TopicSegment{}, errors.New("topic segment is nil")
	}

	switch v := p.GetValue().(type) {
	case *TopicSegment_Wildcard:
		if err := validateTopicWildcard(v.Wildcard); err != nil {
			return stellar.TopicSegment{}, err
		}
		w := v.Wildcard
		return stellar.TopicSegment{Wildcard: &w}, nil
	case *TopicSegment_Scval:
		sv, err := stellarcap.ProtoToScVal(v.Scval)
		if err != nil {
			return stellar.TopicSegment{}, err
		}
		return stellar.TopicSegment{Value: &sv}, nil
	default:
		return stellar.TopicSegment{}, fmt.Errorf("unsupported topic segment type: %T", p.GetValue())
	}
}

// ConvertGetEventsResponseToProto converts a domain GetEventsResponse to proto.
func ConvertGetEventsResponseToProto(resp stellar.GetEventsResponse) (*GetEventsResponse, error) {
	events := make([]*EventInfo, len(resp.Events))
	for i, e := range resp.Events {
		pe, err := convertEventInfoToProto(e)
		if err != nil {
			return nil, fmt.Errorf("events[%d]: %w", i, err)
		}
		events[i] = pe
	}

	return &GetEventsResponse{
		Events:                events,
		Cursor:                resp.Cursor,
		LatestLedger:          resp.LatestLedger,
		OldestLedger:          resp.OldestLedger,
		LatestLedgerCloseTime: resp.LatestLedgerCloseTime,
		OldestLedgerCloseTime: resp.OldestLedgerCloseTime,
	}, nil
}

// ConvertGetEventsResponseFromProto converts proto GetEventsResponse to domain.
func ConvertGetEventsResponseFromProto(p *GetEventsResponse) (stellar.GetEventsResponse, error) {
	if p == nil {
		return stellar.GetEventsResponse{}, errors.New("get events response is nil")
	}

	pEvents := p.GetEvents()
	events := make([]stellar.EventInfo, len(pEvents))
	for i, pe := range pEvents {
		e, err := convertEventInfoFromProto(pe)
		if err != nil {
			return stellar.GetEventsResponse{}, fmt.Errorf("events[%d]: %w", i, err)
		}
		events[i] = e
	}

	return stellar.GetEventsResponse{
		Events:                events,
		Cursor:                p.GetCursor(),
		LatestLedger:          p.GetLatestLedger(),
		OldestLedger:          p.GetOldestLedger(),
		LatestLedgerCloseTime: p.GetLatestLedgerCloseTime(),
		OldestLedgerCloseTime: p.GetOldestLedgerCloseTime(),
	}, nil
}

func ConvertGetTransactionRequestToProto(req stellar.GetTransactionRequest) *GetTransactionRequest {
	return &GetTransactionRequest{TxHash: req.TxHash}
}

func ConvertGetTransactionRequestFromProto(p *GetTransactionRequest) (stellar.GetTransactionRequest, error) {
	if p == nil {
		return stellar.GetTransactionRequest{}, errors.New("get transaction request is nil")
	}
	if p.GetTxHash() == "" {
		return stellar.GetTransactionRequest{}, errors.New("tx hash is required")
	}
	return stellar.GetTransactionRequest{TxHash: p.GetTxHash()}, nil
}

func ConvertGetTransactionResponseToProto(resp stellar.GetTransactionResponse) *GetTransactionResponse {
	return &GetTransactionResponse{
		FeeStroops:      resp.FeeStroops,
		LedgerSequence:  resp.LedgerSequence,
		LedgerCloseTime: resp.LedgerCloseTime,
	}
}

func ConvertGetTransactionResponseFromProto(p *GetTransactionResponse) (stellar.GetTransactionResponse, error) {
	if p == nil {
		return stellar.GetTransactionResponse{}, errors.New("get transaction response is nil")
	}
	return stellar.GetTransactionResponse{
		FeeStroops:      p.GetFeeStroops(),
		LedgerSequence:  p.GetLedgerSequence(),
		LedgerCloseTime: p.GetLedgerCloseTime(),
	}, nil
}

func ConvertGetSigningAccountResponseToProto(resp stellar.GetSigningAccountResponse) *GetSigningAccountResponse {
	return &GetSigningAccountResponse{AccountAddress: resp.AccountAddress}
}

func ConvertGetSigningAccountResponseFromProto(p *GetSigningAccountResponse) (stellar.GetSigningAccountResponse, error) {
	if p == nil {
		return stellar.GetSigningAccountResponse{}, errors.New("get signing account response is nil")
	}
	return stellar.GetSigningAccountResponse{AccountAddress: p.GetAccountAddress()}, nil
}

func convertEventInfoToProto(e stellar.EventInfo) (*EventInfo, error) {
	eventType, err := convertEventTypeToProto(e.EventType)
	if err != nil {
		return nil, fmt.Errorf("eventType: %w", err)
	}

	topics, err := scValsToProto("topics", e.Topics)
	if err != nil {
		return nil, err
	}

	value, err := stellarcap.ScValToProto(e.Value)
	if err != nil {
		return nil, fmt.Errorf("value: %w", err)
	}

	return &EventInfo{
		EventType:        eventType,
		Ledger:           e.Ledger,
		LedgerClosedAt:   e.LedgerClosedAt,
		ContractId:       e.ContractID,
		Id:               e.ID,
		OperationIndex:   e.OperationIndex,
		TransactionIndex: e.TransactionIndex,
		TransactionHash:  e.TransactionHash,
		Topics:           topics,
		Value:            value,
	}, nil
}

func convertEventInfoFromProto(p *EventInfo) (stellar.EventInfo, error) {
	if p == nil {
		return stellar.EventInfo{}, errors.New("event info is nil")
	}

	eventType, err := convertEventTypeFromProto(p.GetEventType())
	if err != nil {
		return stellar.EventInfo{}, fmt.Errorf("eventType: %w", err)
	}

	topics, err := scValsFromProto("topics", p.GetTopics())
	if err != nil {
		return stellar.EventInfo{}, err
	}

	value, err := scValRequiredFromProto("value", p.GetValue())
	if err != nil {
		return stellar.EventInfo{}, err
	}

	return stellar.EventInfo{
		EventType:        eventType,
		Ledger:           p.GetLedger(),
		LedgerClosedAt:   p.GetLedgerClosedAt(),
		ContractID:       p.GetContractId(),
		ID:               p.GetId(),
		OperationIndex:   p.GetOperationIndex(),
		TransactionIndex: p.GetTransactionIndex(),
		TransactionHash:  p.GetTransactionHash(),
		Topics:           topics,
		Value:            value,
	}, nil
}

func convertEventTypeToProto(t stellar.EventType) (EventType, error) {
	switch t {
	case stellar.EventTypeSystem:
		return EventType_EVENT_TYPE_SYSTEM, nil
	case stellar.EventTypeContract:
		return EventType_EVENT_TYPE_CONTRACT, nil
	default:
		return 0, fmt.Errorf("unsupported event type: %d", t)
	}
}

func convertEventTypeFromProto(t EventType) (stellar.EventType, error) {
	switch t {
	case EventType_EVENT_TYPE_SYSTEM:
		return stellar.EventTypeSystem, nil
	case EventType_EVENT_TYPE_CONTRACT:
		return stellar.EventTypeContract, nil
	default:
		return 0, fmt.Errorf("unsupported proto event type: %d", t)
	}
}

func scValsToProto(field string, vals []stellar.ScVal) ([]*scval.ScVal, error) {
	if len(vals) == 0 {
		return nil, nil
	}

	out := make([]*scval.ScVal, len(vals))
	for i, sv := range vals {
		psv, err := stellarcap.ScValToProto(sv)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", field, i, err)
		}
		out[i] = psv
	}

	return out, nil
}

func scValsFromProto(field string, vals []*scval.ScVal) ([]stellar.ScVal, error) {
	if len(vals) == 0 {
		return nil, nil
	}

	out := make([]stellar.ScVal, len(vals))
	for i, psv := range vals {
		sv, err := stellarcap.ProtoToScVal(psv)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", field, i, err)
		}
		out[i] = sv
	}

	return out, nil
}

func scValRequiredFromProto(field string, val *scval.ScVal) (stellar.ScVal, error) {
	if val == nil {
		return stellar.ScVal{}, fmt.Errorf("%s is required", field)
	}

	out, err := stellarcap.ProtoToScVal(val)
	if err != nil {
		return stellar.ScVal{}, fmt.Errorf("%s: %w", field, err)
	}

	return out, nil
}

func paginationToProto(p *stellar.PaginationOptions) *PaginationOptions {
	if p == nil {
		return nil
	}

	return &PaginationOptions{
		Cursor: p.Cursor,
		Limit:  p.Limit,
	}
}

func paginationFromProto(p *PaginationOptions) *stellar.PaginationOptions {
	if p == nil {
		return nil
	}

	return &stellar.PaginationOptions{
		Cursor: p.GetCursor(),
		Limit:  p.GetLimit(),
	}
}

func eventTypesToProto(types []stellar.EventType) ([]EventType, error) {
	out := make([]EventType, len(types))
	for i, t := range types {
		pt, err := convertEventTypeToProto(t)
		if err != nil {
			return nil, fmt.Errorf("eventTypes[%d]: %w", i, err)
		}
		out[i] = pt
	}

	return out, nil
}

func eventTypesFromProto(types []EventType) ([]stellar.EventType, error) {
	out := make([]stellar.EventType, len(types))
	for i, pt := range types {
		t, err := convertEventTypeFromProto(pt)
		if err != nil {
			return nil, fmt.Errorf("eventTypes[%d]: %w", i, err)
		}
		out[i] = t
	}

	return out, nil
}

func validateTopicWildcard(w string) error {
	switch w {
	case "*", "**":
		return nil
	default:
		return fmt.Errorf("wildcard must be '*' or '**', got %q", w)
	}
}

func simulateResourceConfigToProto(c *stellar.SimulateResourceConfig) *SimulateResourceConfig {
	if c == nil {
		return nil
	}

	return &SimulateResourceConfig{
		InstructionLeeway: c.InstructionLeeway,
	}
}

func simulateResourceConfigFromProto(c *SimulateResourceConfig) *stellar.SimulateResourceConfig {
	if c == nil {
		return nil
	}

	return &stellar.SimulateResourceConfig{
		InstructionLeeway: c.GetInstructionLeeway(),
	}
}

func validateSimulateAuthMode(mode stellar.SimulateAuthMode) error {
	switch mode {
	case "",
		stellar.SimulateAuthModeRecord,
		stellar.SimulateAuthModeEnforce,
		stellar.SimulateAuthModeRecordAllowNonroot:
		return nil
	default:
		return fmt.Errorf("unsupported auth mode %q", mode)
	}
}
