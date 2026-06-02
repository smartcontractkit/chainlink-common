package solana

import (
	"fmt"
	"math"

	chainsolana "github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	typesolana "github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

// ConvertComparisonOperatorFromProto converts a proto ComparisonOperator to primitives.ComparisonOperator
func ConvertComparisonOperatorFromProto(op ComparisonOperator) (primitives.ComparisonOperator, error) {
	switch op {
	case ComparisonOperator_COMPARISON_OPERATOR_EQ:
		return primitives.Eq, nil
	case ComparisonOperator_COMPARISON_OPERATOR_NEQ:
		return primitives.Neq, nil
	case ComparisonOperator_COMPARISON_OPERATOR_GT:
		return primitives.Gt, nil
	case ComparisonOperator_COMPARISON_OPERATOR_LT:
		return primitives.Lt, nil
	case ComparisonOperator_COMPARISON_OPERATOR_GTE:
		return primitives.Gte, nil
	case ComparisonOperator_COMPARISON_OPERATOR_LTE:
		return primitives.Lte, nil
	default:
		return 0, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// ConvertComparisonOperatorToProto converts a primitives.ComparisonOperator to proto ComparisonOperator
func ConvertComparisonOperatorToProto(op primitives.ComparisonOperator) (ComparisonOperator, error) {
	switch op {
	case primitives.Eq:
		return ComparisonOperator_COMPARISON_OPERATOR_EQ, nil
	case primitives.Neq:
		return ComparisonOperator_COMPARISON_OPERATOR_NEQ, nil
	case primitives.Gt:
		return ComparisonOperator_COMPARISON_OPERATOR_GT, nil
	case primitives.Lt:
		return ComparisonOperator_COMPARISON_OPERATOR_LT, nil
	case primitives.Gte:
		return ComparisonOperator_COMPARISON_OPERATOR_GTE, nil
	case primitives.Lte:
		return ComparisonOperator_COMPARISON_OPERATOR_LTE, nil
	default:
		return 0, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// ConvertValueComparatorsFromProto converts proto ValueComparator slice to primitives.ValueComparator slice
func ConvertValueComparatorsFromProto(comparers []*ValueComparator) ([]primitives.ValueComparator, error) {
	if comparers == nil {
		return nil, nil
	}
	result := make([]primitives.ValueComparator, len(comparers))
	for i, c := range comparers {
		if c != nil {
			operator, err := ConvertComparisonOperatorFromProto(c.Operator)
			if err != nil {
				return nil, fmt.Errorf("failed to convert comparison operator: %w", err)
			}
			result[i] = primitives.ValueComparator{
				Value:    c.Value, // []byte is compatible with any
				Operator: operator,
			}
		}
	}
	return result, nil
}

// ConvertValueComparatorsToProto converts primitives.ValueComparator slice to proto ValueComparator slice
func ConvertValueComparatorsToProto(comparers []primitives.ValueComparator) ([]*ValueComparator, error) {
	if comparers == nil {
		return nil, nil
	}
	result := make([]*ValueComparator, len(comparers))
	for i, c := range comparers {
		// Handle the Value field which could be any type, convert to []byte if possible
		var valueBytes []byte
		if b, ok := c.Value.([]byte); ok {
			valueBytes = b
		} else {
			return nil, fmt.Errorf("value is not a []byte: %T", c.Value)
		}
		operator, err := ConvertComparisonOperatorToProto(c.Operator)
		if err != nil {
			return nil, fmt.Errorf("failed to convert comparison operator: %w", err)
		}
		result[i] = &ValueComparator{
			Value:    valueBytes,
			Operator: operator,
		}
	}
	return result, nil
}

// ConvertLogFromProto converts a proto Log to typesolana.Log
func ConvertLogFromProto(p *Log) (*typesolana.Log, error) {
	if p == nil {
		return nil, nil
	}

	blockHash, err := chainsolana.ConvertHashFromProto(p.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to convert block hash: %w", err)
	}

	address, err := chainsolana.ConvertPublicKeyFromProto(p.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to convert address: %w", err)
	}

	eventSig, err := chainsolana.ConvertEventSigFromProto(p.EventSig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert event sig: %w", err)
	}

	txHash, err := chainsolana.ConvertSignatureFromProto(p.TxHash)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tx hash: %w", err)
	}

	return &typesolana.Log{
		ChainID:        p.ChainId,
		LogIndex:       p.LogIndex,
		BlockHash:      blockHash,
		BlockNumber:    p.BlockNumber,
		BlockTimestamp: p.BlockTimestamp,
		Address:        address,
		EventSig:       eventSig,
		TxHash:         txHash,
		Data:           p.Data,
		SequenceNum:    p.SequenceNum,
		Error:          p.Error,
	}, nil
}

// ConvertLogToProto converts a typesolana.Log to proto Log
func ConvertLogToProto(l *typesolana.Log) *Log {
	if l == nil {
		return nil
	}
	return &Log{
		ChainId:        l.ChainID,
		LogIndex:       l.LogIndex,
		BlockHash:      l.BlockHash[:],
		BlockNumber:    l.BlockNumber,
		BlockTimestamp: l.BlockTimestamp,
		Address:        l.Address[:],
		EventSig:       l.EventSig[:],
		TxHash:         l.TxHash[:],
		Data:           l.Data,
		SequenceNum:    l.SequenceNum,
		Error:          l.Error,
	}
}

// convertEncodingTypeFromProto maps capability proto encoding to domain types.
// EncodingType_ENCODING_TYPE_NONE maps to EncodingBase64.
func convertEncodingTypeFromProto(e EncodingType) (typesolana.EncodingType, error) {
	switch e {
	case EncodingType_ENCODING_TYPE_NONE:
		return typesolana.EncodingBase64, nil
	case EncodingType_ENCODING_TYPE_BASE58:
		return typesolana.EncodingBase58, nil
	case EncodingType_ENCODING_TYPE_BASE64:
		return typesolana.EncodingBase64, nil
	case EncodingType_ENCODING_TYPE_BASE64_ZSTD:
		return typesolana.EncodingBase64Zstd, nil
	case EncodingType_ENCODING_TYPE_JSON_PARSED:
		return typesolana.EncodingJSONParsed, nil
	case EncodingType_ENCODING_TYPE_JSON:
		return typesolana.EncodingJSON, nil
	default:
		return "", fmt.Errorf("unknown encoding type: %s", e)
	}
}

// convertEncodingTypeToProto maps domain encoding to capability proto.
func convertEncodingTypeToProto(e typesolana.EncodingType) (EncodingType, error) {
	switch e {
	case typesolana.EncodingBase64:
		return EncodingType_ENCODING_TYPE_BASE64, nil
	case typesolana.EncodingBase58:
		return EncodingType_ENCODING_TYPE_BASE58, nil
	case typesolana.EncodingBase64Zstd:
		return EncodingType_ENCODING_TYPE_BASE64_ZSTD, nil
	case typesolana.EncodingJSONParsed:
		return EncodingType_ENCODING_TYPE_JSON_PARSED, nil
	case typesolana.EncodingJSON:
		return EncodingType_ENCODING_TYPE_JSON, nil
	case "":
		return EncodingType_ENCODING_TYPE_NONE, nil // RPC may return empty string in a response, since we do not know which encoding is default for that RPC client return none.
	default:
		return 0, fmt.Errorf("unknown encoding type: %q", e)
	}
}

// convertCommitmentTypeFromProto maps capability proto commitment to domain types.
// CommitmentType_COMMITMENT_TYPE_NONE maps to CommitmentFinalized.
func convertCommitmentTypeFromProto(c CommitmentType) (typesolana.CommitmentType, error) {
	switch c {
	case CommitmentType_COMMITMENT_TYPE_NONE:
		return typesolana.CommitmentFinalized, nil
	case CommitmentType_COMMITMENT_TYPE_FINALIZED:
		return typesolana.CommitmentFinalized, nil
	case CommitmentType_COMMITMENT_TYPE_CONFIRMED:
		return typesolana.CommitmentConfirmed, nil
	case CommitmentType_COMMITMENT_TYPE_PROCESSED:
		return typesolana.CommitmentProcessed, nil
	default:
		return "", fmt.Errorf("unknown commitment type: %s", c)
	}
}

// convertCommitmentTypeToProto maps domain commitment to capability proto.
func convertCommitmentTypeToProto(c typesolana.CommitmentType) (CommitmentType, error) {
	switch c {
	case typesolana.CommitmentFinalized:
		return CommitmentType_COMMITMENT_TYPE_FINALIZED, nil
	case typesolana.CommitmentConfirmed:
		return CommitmentType_COMMITMENT_TYPE_CONFIRMED, nil
	case typesolana.CommitmentProcessed:
		return CommitmentType_COMMITMENT_TYPE_PROCESSED, nil
	default:
		return 0, fmt.Errorf("unknown commitment type: %q", c)
	}
}

// convertConfirmationStatusTypeToProto maps domain confirmation status to capability proto.
func convertConfirmationStatusTypeToProto(c typesolana.ConfirmationStatusType) (ConfirmationStatusType, error) {
	switch c {
	case typesolana.ConfirmationStatusFinalized:
		return ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_FINALIZED, nil
	case typesolana.ConfirmationStatusConfirmed:
		return ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_CONFIRMED, nil
	case typesolana.ConfirmationStatusProcessed:
		return ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_PROCESSED, nil
	default:
		return 0, fmt.Errorf("unknown confirmation status type: %q", c)
	}
}

func convertDataSliceFromProto(p *DataSlice) *typesolana.DataSlice {
	if p == nil {
		return nil
	}
	// Offset and length are required, so do not treat 0 as nil.
	// If nil values are need user may pass nil DataSlice
	return &typesolana.DataSlice{
		Offset: &p.Offset,
		Length: &p.Length,
	}
}

func convertUiTokenAmountToProto(u *typesolana.UiTokenAmount) *UiTokenAmount {
	if u == nil {
		return nil
	}
	return &UiTokenAmount{
		Amount:         u.Amount,
		Decimals:       uint32(u.Decimals),
		UiAmountString: u.UiAmountString,
	}
}

func convertDataBytesOrJSONToProto(d *typesolana.DataBytesOrJSON) (*DataBytesOrJSON, error) {
	if d == nil {
		return nil, nil
	}
	enc, err := convertEncodingTypeToProto(d.RawDataEncoding)
	if err != nil {
		return nil, fmt.Errorf("encoding: %w", err)
	}
	ret := &DataBytesOrJSON{Encoding: enc}
	if d.AsJSON != nil {
		ret.Body = &DataBytesOrJSON_Json{Json: d.AsJSON}
		return ret, nil
	}
	ret.Body = &DataBytesOrJSON_Raw{Raw: d.AsDecodedBinary}
	return ret, nil
}

func convertAccountToProto(a *typesolana.Account) (*Account, error) {
	if a == nil {
		return nil, nil
	}
	data, err := convertDataBytesOrJSONToProto(a.Data)
	if err != nil {
		return nil, fmt.Errorf("data: %w", err)
	}
	return &Account{
		Lamports:   a.Lamports,
		Owner:      a.Owner[:],
		Data:       data,
		Executable: a.Executable,
		RentEpoch:  valuespb.NewBigIntFromInt(a.RentEpoch),
		Space:      a.Space,
	}, nil
}

func convertGetAccountInfoOptsFromProto(p *GetAccountInfoOpts) (*typesolana.GetAccountInfoOpts, error) {
	if p == nil {
		return nil, nil
	}
	enc, err := convertEncodingTypeFromProto(p.Encoding)
	if err != nil {
		return nil, fmt.Errorf("encoding: %w", err)
	}
	commit, err := convertCommitmentTypeFromProto(p.Commitment)
	if err != nil {
		return nil, fmt.Errorf("commitment: %w", err)
	}
	return &typesolana.GetAccountInfoOpts{
		Encoding:       enc,
		Commitment:     commit,
		DataSlice:      convertDataSliceFromProto(p.DataSlice),
		MinContextSlot: ptrUint64(p.MinContextSlot),
	}, nil
}

func convertGetMultipleAccountsOptsFromProto(p *GetMultipleAccountsOpts) (*typesolana.GetMultipleAccountsOpts, error) {
	if p == nil {
		return nil, nil
	}
	enc, err := convertEncodingTypeFromProto(p.Encoding)
	if err != nil {
		return nil, fmt.Errorf("encoding: %w", err)
	}
	commit, err := convertCommitmentTypeFromProto(p.Commitment)
	if err != nil {
		return nil, fmt.Errorf("commitment: %w", err)
	}
	return &typesolana.GetMultipleAccountsOpts{
		Encoding:       enc,
		Commitment:     commit,
		DataSlice:      convertDataSliceFromProto(p.DataSlice),
		MinContextSlot: ptrUint64(p.MinContextSlot),
	}, nil
}

func convertGetBlockOptsFromProto(p *GetBlockOpts) (*typesolana.GetBlockOpts, error) {
	if p == nil {
		return nil, nil
	}
	commit, err := convertCommitmentTypeFromProto(p.Commitment)
	if err != nil {
		return nil, fmt.Errorf("commitment: %w", err)
	}
	return &typesolana.GetBlockOpts{Commitment: commit}, nil
}

func convertMessageHeaderFromProto(p *MessageHeader) (typesolana.MessageHeader, error) {
	if p == nil {
		return typesolana.MessageHeader{}, nil
	}
	if p.NumRequiredSignatures > math.MaxUint8 {
		return typesolana.MessageHeader{}, fmt.Errorf("num_required_signatures %d exceeds max uint8", p.NumRequiredSignatures)
	}
	if p.NumReadonlySignedAccounts > math.MaxUint8 {
		return typesolana.MessageHeader{}, fmt.Errorf("num_readonly_signed_accounts %d exceeds max uint8", p.NumReadonlySignedAccounts)
	}
	if p.NumReadonlyUnsignedAccounts > math.MaxUint8 {
		return typesolana.MessageHeader{}, fmt.Errorf("num_readonly_unsigned_accounts %d exceeds max uint8", p.NumReadonlyUnsignedAccounts)
	}
	return typesolana.MessageHeader{
		NumRequiredSignatures:       uint8(p.NumRequiredSignatures),
		NumReadonlySignedAccounts:   uint8(p.NumReadonlySignedAccounts),
		NumReadonlyUnsignedAccounts: uint8(p.NumReadonlyUnsignedAccounts),
	}, nil
}

func convertMessageHeaderToProto(h typesolana.MessageHeader) *MessageHeader {
	return &MessageHeader{
		NumRequiredSignatures:       uint32(h.NumRequiredSignatures),
		NumReadonlySignedAccounts:   uint32(h.NumReadonlySignedAccounts),
		NumReadonlyUnsignedAccounts: uint32(h.NumReadonlyUnsignedAccounts),
	}
}

func convertCompiledInstructionFromProto(p *CompiledInstruction) (typesolana.CompiledInstruction, error) {
	if p == nil {
		return typesolana.CompiledInstruction{}, nil
	}
	if p.ProgramIdIndex > math.MaxUint16 {
		return typesolana.CompiledInstruction{}, fmt.Errorf("program_id_index %d exceeds max uint16", p.ProgramIdIndex)
	}
	if p.StackHeight > math.MaxUint16 {
		return typesolana.CompiledInstruction{}, fmt.Errorf("stack_height %d exceeds max uint16", p.StackHeight)
	}
	accts := make([]uint16, len(p.Accounts))
	for i, a := range p.Accounts {
		if a > math.MaxUint16 {
			return typesolana.CompiledInstruction{}, fmt.Errorf("account index %d value %d exceeds max uint16", i, a)
		}
		accts[i] = uint16(a)
	}
	return typesolana.CompiledInstruction{
		ProgramIDIndex: uint16(p.ProgramIdIndex),
		Accounts:       accts,
		Data:           p.Data,
		StackHeight:    uint16(p.StackHeight),
	}, nil
}

func convertCompiledInstructionToProto(ci typesolana.CompiledInstruction) *CompiledInstruction {
	accts := make([]uint32, len(ci.Accounts))
	for i, a := range ci.Accounts {
		accts[i] = uint32(a)
	}
	return &CompiledInstruction{
		ProgramIdIndex: uint32(ci.ProgramIDIndex),
		Accounts:       accts,
		Data:           ci.Data,
		StackHeight:    uint32(ci.StackHeight),
	}
}

func convertInnerInstructionToProto(ii typesolana.InnerInstruction) *InnerInstruction {
	out := &InnerInstruction{
		Index:        uint32(ii.Index),
		Instructions: make([]*CompiledInstruction, 0, len(ii.Instructions)),
	}
	for _, in := range ii.Instructions {
		out.Instructions = append(out.Instructions, convertCompiledInstructionToProto(in))
	}
	return out
}

func convertLoadedAddressesToProto(l typesolana.LoadedAddresses) *LoadedAddresses {
	return &LoadedAddresses{
		Readonly: chainsolana.ConvertPublicKeysToProto(l.ReadOnly),
		Writable: chainsolana.ConvertPublicKeysToProto(l.Writable),
	}
}

func convertTokenBalanceToProto(tb *typesolana.TokenBalance) *TokenBalance {
	if tb == nil {
		return nil
	}
	var owner, program []byte
	if tb.Owner != nil {
		tmp := *tb.Owner
		owner = tmp[:]
	}
	if tb.ProgramId != nil {
		tmp := *tb.ProgramId
		program = tmp[:]
	}
	return &TokenBalance{
		AccountIndex: uint32(tb.AccountIndex),
		Owner:        owner,
		ProgramId:    program,
		Mint:         tb.Mint[:],
		Ui:           convertUiTokenAmountToProto(tb.UiTokenAmount),
	}
}

func convertReturnDataToProto(r *typesolana.ReturnData) (*ReturnData, error) {
	if r == nil {
		return nil, nil
	}
	enc, err := convertEncodingTypeToProto(r.Data.Encoding)
	if err != nil {
		return nil, fmt.Errorf("encoding: %w", err)
	}
	return &ReturnData{
		ProgramId: r.ProgramId[:],
		Data: &Data{
			Content:  r.Data.Content,
			Encoding: enc,
		},
	}, nil
}

func convertTransactionMetaToProto(m *typesolana.TransactionMeta) (*TransactionMeta, error) {
	if m == nil {
		return nil, nil
	}
	preTB := make([]*TokenBalance, 0, len(m.PreTokenBalances))
	for i := range m.PreTokenBalances {
		preTB = append(preTB, convertTokenBalanceToProto(&m.PreTokenBalances[i]))
	}
	postTB := make([]*TokenBalance, 0, len(m.PostTokenBalances))
	for i := range m.PostTokenBalances {
		postTB = append(postTB, convertTokenBalanceToProto(&m.PostTokenBalances[i]))
	}
	inner := make([]*InnerInstruction, 0, len(m.InnerInstructions))
	for _, in := range m.InnerInstructions {
		inner = append(inner, convertInnerInstructionToProto(in))
	}
	rd, err := convertReturnDataToProto(&m.ReturnData)
	if err != nil {
		return nil, fmt.Errorf("return_data: %w", err)
	}
	return &TransactionMeta{
		ErrJson:              m.Err,
		Fee:                  m.Fee,
		PreBalances:          m.PreBalances,
		PostBalances:         m.PostBalances,
		LogMessages:          m.LogMessages,
		PreTokenBalances:     preTB,
		PostTokenBalances:    postTB,
		InnerInstructions:    inner,
		LoadedAddresses:      convertLoadedAddressesToProto(m.LoadedAddresses),
		ReturnData:           rd,
		ComputeUnitsConsumed: m.ComputeUnitsConsumed,
	}, nil
}

func convertParsedMessageFromProto(p *ParsedMessage) (typesolana.Message, error) {
	if p == nil {
		return typesolana.Message{}, nil
	}
	rb, err := chainsolana.ConvertHashFromProto(p.RecentBlockhash)
	if err != nil {
		return typesolana.Message{}, fmt.Errorf("recent_blockhash: %w", err)
	}
	keys, err := chainsolana.ConvertPublicKeysFromProto(p.AccountKeys)
	if err != nil {
		return typesolana.Message{}, fmt.Errorf("account_keys: %w", err)
	}
	hdr, err := convertMessageHeaderFromProto(p.Header)
	if err != nil {
		return typesolana.Message{}, fmt.Errorf("header: %w", err)
	}
	msg := typesolana.Message{
		AccountKeys:     keys,
		Header:          hdr,
		RecentBlockhash: rb,
		Instructions:    make([]typesolana.CompiledInstruction, 0, len(p.Instructions)),
	}
	for i, ins := range p.Instructions {
		ci, err := convertCompiledInstructionFromProto(ins)
		if err != nil {
			return typesolana.Message{}, fmt.Errorf("instruction[%d]: %w", i, err)
		}
		msg.Instructions = append(msg.Instructions, ci)
	}
	return msg, nil
}

func convertParsedMessageToProto(m typesolana.Message) *ParsedMessage {
	out := &ParsedMessage{
		RecentBlockhash: m.RecentBlockhash[:],
		AccountKeys:     chainsolana.ConvertPublicKeysToProto(m.AccountKeys),
		Header:          convertMessageHeaderToProto(m.Header),
		Instructions:    make([]*CompiledInstruction, 0, len(m.Instructions)),
	}
	for _, ins := range m.Instructions {
		out.Instructions = append(out.Instructions, convertCompiledInstructionToProto(ins))
	}
	return out
}

func convertParsedTransactionToProto(t typesolana.Transaction) *ParsedTransaction {
	return &ParsedTransaction{
		Signatures: chainsolana.ConvertSignaturesToProto(t.Signatures),
		Message:    convertParsedMessageToProto(t.Message),
	}
}

func convertTransactionEnvelopeToProto(e typesolana.TransactionResultEnvelope) (*TransactionEnvelope, error) {
	switch {
	case e.AsParsedTransaction != nil:
		return &TransactionEnvelope{
			Transaction: &TransactionEnvelope_Parsed{
				Parsed: convertParsedTransactionToProto(*e.AsParsedTransaction),
			},
		}, nil
	case len(e.AsDecodedBinary.Content) > 0:
		return &TransactionEnvelope{
			Transaction: &TransactionEnvelope_Raw{Raw: e.AsDecodedBinary.Content},
		}, nil
	default:
		return nil, fmt.Errorf("transaction envelope has no content")
	}
}

func convertSimulateTransactionAccountsOptsFromProto(p *SimulateTransactionAccountsOpts) (*typesolana.SimulateTransactionAccountsOpts, error) {
	if p == nil {
		return nil, nil
	}
	enc, err := convertEncodingTypeFromProto(p.Encoding)
	if err != nil {
		return nil, fmt.Errorf("encoding: %w", err)
	}
	addrs, err := chainsolana.ConvertPublicKeysFromProto(p.Addresses)
	if err != nil {
		return nil, fmt.Errorf("addresses: %w", err)
	}
	return &typesolana.SimulateTransactionAccountsOpts{
		Encoding:  enc,
		Addresses: addrs,
	}, nil
}

func convertSimulateTXOptsFromProto(p *SimulateTXOpts) (*typesolana.SimulateTXOpts, error) {
	if p == nil {
		return nil, nil
	}
	commit, err := convertCommitmentTypeFromProto(p.Commitment)
	if err != nil {
		return nil, fmt.Errorf("commitment: %w", err)
	}
	accts, err := convertSimulateTransactionAccountsOptsFromProto(p.Accounts)
	if err != nil {
		return nil, fmt.Errorf("accounts: %w", err)
	}
	return &typesolana.SimulateTXOpts{
		SigVerify:              p.SigVerify,
		Commitment:             commit,
		ReplaceRecentBlockhash: p.ReplaceRecentBlockhash,
		Accounts:               accts,
	}, nil
}

// --- Reply types (domain -> proto) ---

// ConvertGetAccountInfoReplyToProto converts GetAccountInfoReply to GetAccountInfoWithOptsReply.
func ConvertGetAccountInfoReplyToProto(r *typesolana.GetAccountInfoReply) (*GetAccountInfoWithOptsReply, error) {
	if r == nil {
		return nil, nil
	}
	val, err := convertAccountToProto(r.Value)
	if err != nil {
		return nil, fmt.Errorf("value: %w", err)
	}
	return &GetAccountInfoWithOptsReply{
		Value: val,
	}, nil
}

// ConvertGetMultipleAccountsReplyToProto converts GetMultipleAccountsReply to GetMultipleAccountsWithOptsReply.
func ConvertGetMultipleAccountsReplyToProto(r *typesolana.GetMultipleAccountsReply) (*GetMultipleAccountsWithOptsReply, error) {
	if r == nil {
		return nil, nil
	}
	val := make([]*OptionalAccountWrapper, 0, len(r.Value))
	for i, a := range r.Value {
		acc, err := convertAccountToProto(a)
		if err != nil {
			return nil, fmt.Errorf("value[%d]: %w", i, err)
		}
		val = append(val, &OptionalAccountWrapper{Account: acc})
	}
	return &GetMultipleAccountsWithOptsReply{
		Value: val,
	}, nil
}

// ConvertGetBalanceReplyToProto converts GetBalanceReply to proto.
func ConvertGetBalanceReplyToProto(r *typesolana.GetBalanceReply) (*GetBalanceReply, error) {
	if r == nil {
		return nil, nil
	}
	return &GetBalanceReply{Value: r.Value}, nil
}

// ConvertGetBlockReplyToProto converts GetBlockReply to proto.
func ConvertGetBlockReplyToProto(r *typesolana.GetBlockReply) (*GetBlockReply, error) {
	if r == nil {
		return nil, nil
	}
	var bt *int64
	if r.BlockTime != nil {
		t := int64(*r.BlockTime)
		bt = &t
	}
	var bh uint64
	if r.BlockHeight != nil {
		bh = *r.BlockHeight
	}
	return &GetBlockReply{
		Blockhash:         r.Blockhash[:],
		PreviousBlockhash: r.PreviousBlockhash[:],
		ParentSlot:        r.ParentSlot,
		BlockTime:         bt,
		BlockHeight:       bh,
	}, nil
}

// ConvertGetSlotHeightReplyToProto converts GetSlotHeightReply to proto.
func ConvertGetSlotHeightReplyToProto(r *typesolana.GetSlotHeightReply) (*GetSlotHeightReply, error) {
	if r == nil {
		return nil, nil
	}
	return &GetSlotHeightReply{Height: r.Height}, nil
}

// ConvertGetTransactionReplyToProto converts GetTransactionReply to proto.
func ConvertGetTransactionReplyToProto(r *typesolana.GetTransactionReply) (*GetTransactionReply, error) {
	if r == nil {
		return nil, nil
	}
	var bt *int64
	if r.BlockTime != nil {
		t := int64(*r.BlockTime)
		bt = &t
	}
	var tx *TransactionEnvelope
	if r.Transaction != nil {
		var err error
		tx, err = convertTransactionEnvelopeToProto(*r.Transaction)
		if err != nil {
			return nil, fmt.Errorf("failed to convert transaction: %w", err)
		}
	}
	meta, err := convertTransactionMetaToProto(r.Meta)
	if err != nil {
		return nil, fmt.Errorf("meta: %w", err)
	}
	return &GetTransactionReply{
		Slot:        r.Slot,
		BlockTime:   bt,
		Transaction: tx,
		Meta:        meta,
	}, nil
}

// ConvertGetFeeForMessageReplyToProto converts GetFeeForMessageReply to proto.
func ConvertGetFeeForMessageReplyToProto(r *typesolana.GetFeeForMessageReply) (*GetFeeForMessageReply, error) {
	if r == nil {
		return nil, nil
	}
	return &GetFeeForMessageReply{Fee: r.Fee}, nil
}

// ConvertSimulateTXReplyToProto converts SimulateTXReply to proto.
func ConvertSimulateTXReplyToProto(r *typesolana.SimulateTXReply) (*SimulateTXReply, error) {
	if r == nil {
		return nil, nil
	}
	accs := make([]*Account, 0, len(r.Accounts))
	for i, a := range r.Accounts {
		acc, err := convertAccountToProto(a)
		if err != nil {
			return nil, fmt.Errorf("accounts[%d]: %w", i, err)
		}
		accs = append(accs, acc)
	}
	var units uint64
	if r.UnitsConsumed != nil {
		units = *r.UnitsConsumed
	}
	return &SimulateTXReply{
		Err:           r.Err,
		Logs:          r.Logs,
		Accounts:      accs,
		UnitsConsumed: units,
	}, nil
}

// ConvertGetSignatureStatusesReplyToProto converts GetSignatureStatusesReply to proto.
func ConvertGetSignatureStatusesReplyToProto(r *typesolana.GetSignatureStatusesReply) (*GetSignatureStatusesReply, error) {
	if r == nil {
		return nil, nil
	}
	out := &GetSignatureStatusesReply{Results: make([]*GetSignatureStatusesResult, 0, len(r.Results))}
	for i := range r.Results {
		status, err := convertConfirmationStatusTypeToProto(r.Results[i].ConfirmationStatus)
		if err != nil {
			return nil, fmt.Errorf("results[%d].confirmation_status: %w", i, err)
		}
		res := &GetSignatureStatusesResult{
			Slot:               r.Results[i].Slot,
			Err:                r.Results[i].Err,
			ConfirmationStatus: status,
		}
		if r.Results[i].Confirmations != nil {
			res.Confirmations = r.Results[i].Confirmations
		}
		out.Results = append(out.Results, res)
	}
	return out, nil
}

// --- Request types (proto -> domain) ---

// ConvertGetAccountInfoRequestFromProto converts GetAccountInfoWithOptsRequest to GetAccountInfoRequest.
func ConvertGetAccountInfoRequestFromProto(p *GetAccountInfoWithOptsRequest) (typesolana.GetAccountInfoRequest, error) {
	if p == nil {
		return typesolana.GetAccountInfoRequest{}, nil
	}
	account, err := chainsolana.ConvertPublicKeyFromProto(p.Account)
	if err != nil {
		return typesolana.GetAccountInfoRequest{}, fmt.Errorf("account: %w", err)
	}
	opts, err := convertGetAccountInfoOptsFromProto(p.Opts)
	if err != nil {
		return typesolana.GetAccountInfoRequest{}, fmt.Errorf("opts: %w", err)
	}
	return typesolana.GetAccountInfoRequest{Account: account, Opts: opts}, nil
}

// ConvertGetMultipleAccountsRequestFromProto converts GetMultipleAccountsWithOptsRequest to GetMultipleAccountsRequest.
func ConvertGetMultipleAccountsRequestFromProto(p *GetMultipleAccountsWithOptsRequest) (*typesolana.GetMultipleAccountsRequest, error) {
	if p == nil {
		return nil, nil
	}
	accts, err := chainsolana.ConvertPublicKeysFromProto(p.Accounts)
	if err != nil {
		return nil, fmt.Errorf("accounts: %w", err)
	}
	opts, err := convertGetMultipleAccountsOptsFromProto(p.Opts)
	if err != nil {
		return nil, fmt.Errorf("opts: %w", err)
	}
	return &typesolana.GetMultipleAccountsRequest{Accounts: accts, Opts: opts}, nil
}

// ConvertGetBalanceRequestFromProto converts GetBalanceRequest to domain type.
func ConvertGetBalanceRequestFromProto(p *GetBalanceRequest) (typesolana.GetBalanceRequest, error) {
	if p == nil {
		return typesolana.GetBalanceRequest{}, nil
	}
	addr, err := chainsolana.ConvertPublicKeyFromProto(p.Addr)
	if err != nil {
		return typesolana.GetBalanceRequest{}, fmt.Errorf("addr: %w", err)
	}
	commit, err := convertCommitmentTypeFromProto(p.Commitment)
	if err != nil {
		return typesolana.GetBalanceRequest{}, fmt.Errorf("commitment: %w", err)
	}
	return typesolana.GetBalanceRequest{Addr: addr, Commitment: commit}, nil
}

// ConvertGetBlockRequestFromProto converts GetBlockRequest to domain type.
func ConvertGetBlockRequestFromProto(p *GetBlockRequest) (*typesolana.GetBlockRequest, error) {
	if p == nil {
		return nil, nil
	}
	opts, err := convertGetBlockOptsFromProto(p.Opts)
	if err != nil {
		return nil, fmt.Errorf("opts: %w", err)
	}
	return &typesolana.GetBlockRequest{Slot: p.Slot, Opts: opts}, nil
}

// ConvertGetSlotHeightRequestFromProto converts GetSlotHeightRequest to domain type.
func ConvertGetSlotHeightRequestFromProto(p *GetSlotHeightRequest) (typesolana.GetSlotHeightRequest, error) {
	if p == nil {
		return typesolana.GetSlotHeightRequest{}, nil
	}
	commit, err := convertCommitmentTypeFromProto(p.Commitment)
	if err != nil {
		return typesolana.GetSlotHeightRequest{}, fmt.Errorf("commitment: %w", err)
	}
	return typesolana.GetSlotHeightRequest{Commitment: commit}, nil
}

// ConvertGetTransactionRequestFromProto converts GetTransactionRequest to domain type.
func ConvertGetTransactionRequestFromProto(p *GetTransactionRequest) (typesolana.GetTransactionRequest, error) {
	if p == nil {
		return typesolana.GetTransactionRequest{}, nil
	}
	sig, err := chainsolana.ConvertSignatureFromProto(p.Signature)
	if err != nil {
		return typesolana.GetTransactionRequest{}, fmt.Errorf("signature: %w", err)
	}
	return typesolana.GetTransactionRequest{Signature: sig}, nil
}

// ConvertGetFeeForMessageRequestFromProto converts GetFeeForMessageRequest to domain type.
func ConvertGetFeeForMessageRequestFromProto(p *GetFeeForMessageRequest) (*typesolana.GetFeeForMessageRequest, error) {
	if p == nil {
		return nil, nil
	}
	commit, err := convertCommitmentTypeFromProto(p.Commitment)
	if err != nil {
		return nil, fmt.Errorf("commitment: %w", err)
	}
	return &typesolana.GetFeeForMessageRequest{
		Message:    p.Message,
		Commitment: commit,
	}, nil
}

// ConvertSimulateTXRequestFromProto converts SimulateTXRequest to domain type.
func ConvertSimulateTXRequestFromProto(p *SimulateTXRequest) (typesolana.SimulateTXRequest, error) {
	if p == nil {
		return typesolana.SimulateTXRequest{}, nil
	}
	recv, err := chainsolana.ConvertPublicKeyFromProto(p.Receiver)
	if err != nil {
		return typesolana.SimulateTXRequest{}, fmt.Errorf("receiver: %w", err)
	}
	opts, err := convertSimulateTXOptsFromProto(p.Opts)
	if err != nil {
		return typesolana.SimulateTXRequest{}, fmt.Errorf("opts: %w", err)
	}
	return typesolana.SimulateTXRequest{
		Receiver:           recv,
		EncodedTransaction: p.EncodedTransaction,
		Opts:               opts,
	}, nil
}

// ConvertGetSignatureStatusesRequestFromProto converts GetSignatureStatusesRequest to domain type.
func ConvertGetSignatureStatusesRequestFromProto(p *GetSignatureStatusesRequest) (*typesolana.GetSignatureStatusesRequest, error) {
	if p == nil {
		return nil, nil
	}
	sigs, err := chainsolana.ConvertSignaturesFromProto(p.Sigs)
	if err != nil {
		return nil, fmt.Errorf("sigs: %w", err)
	}
	return &typesolana.GetSignatureStatusesRequest{Sigs: sigs}, nil
}

func ptrUint64(v uint64) *uint64 {
	if v == 0 {
		return nil
	}

	return &v
}
