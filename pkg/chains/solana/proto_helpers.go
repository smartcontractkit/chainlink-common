package solana

import (
	"errors"
	"fmt"
	"time"

	"github.com/mr-tron/base58"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codecpb "github.com/smartcontractkit/chainlink-common/pkg/internal/codec"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	typesolana "github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	solprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/solana"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

func ConvertPublicKeysFromProto(pubKeys [][]byte) ([]typesolana.PublicKey, error) {
	out := make([]typesolana.PublicKey, 0, len(pubKeys))
	var errs []error
	for i, b := range pubKeys {
		pk, err := ConvertPublicKeyFromProto(b)
		if err != nil {
			errs = append(errs, fmt.Errorf("public key[%d]: %w", i, err))
			continue
		}
		out = append(out, pk)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return out, nil
}

func ConvertPublicKeyFromProto(b []byte) (typesolana.PublicKey, error) {
	if err := ValidatePublicKeyBytes(b); err != nil {
		return typesolana.PublicKey{}, err
	}
	return typesolana.PublicKey(b), nil
}

func ValidatePublicKeyBytes(b []byte) error {
	if b == nil {
		return fmt.Errorf("address can't be nil")
	}
	if len(b) != typesolana.PublicKeyLength {
		return fmt.Errorf("invalid public key: got %d bytes, expected %d, value=%s",
			len(b), typesolana.PublicKeyLength, base58.Encode(b))
	}
	return nil
}

func ConvertSignatureFromProto(b []byte) (typesolana.Signature, error) {
	if err := ValidateSignatureBytes(b); err != nil {
		return typesolana.Signature{}, err
	}
	var s typesolana.Signature
	copy(s[:], b[:typesolana.SignatureLength])
	return s, nil
}

func ConvertSignaturesFromProto(arr [][]byte) ([]typesolana.Signature, error) {
	out := make([]typesolana.Signature, 0, len(arr))
	var errs []error
	for i, b := range arr {
		s, err := ConvertSignatureFromProto(b)
		if err != nil {
			errs = append(errs, fmt.Errorf("signature[%d]: %w", i, err))
			continue
		}
		out = append(out, s)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return out, nil
}

func ConvertSignaturesToProto(sigs []typesolana.Signature) [][]byte {
	out := make([][]byte, 0, len(sigs))
	for _, s := range sigs {
		out = append(out, s[:])
	}
	return out
}

func ValidateSignatureBytes(b []byte) error {
	if b == nil {
		return fmt.Errorf("signature can't be nil")
	}
	if len(b) != typesolana.SignatureLength {
		return fmt.Errorf("invalid signature: got %d bytes, expected %d, value=%s",
			len(b), typesolana.SignatureLength, base58.Encode(b))
	}
	return nil
}

func ConvertHashFromProto(b []byte) (typesolana.Hash, error) {
	if b == nil {
		return typesolana.Hash{}, fmt.Errorf("hash can't be nil")
	}

	if len(b) != solana.HashLength {
		return typesolana.Hash{}, fmt.Errorf("invalid hash: got %d bytes, expected %d, value=%s",
			len(b), solana.HashLength, base58.Encode(b))
	}
	return typesolana.Hash(typesolana.PublicKey(b)), nil
}

func ConvertEventSigFromProto(b []byte) (typesolana.EventSignature, error) {
	if b == nil {
		return typesolana.EventSignature{}, fmt.Errorf("hash can't be nil")
	}

	if len(b) != solana.EventSignatureLength {
		return typesolana.EventSignature{}, fmt.Errorf("invalid event signature: got %d bytes, expected %d, value=%x",
			len(b), solana.EventSignatureLength, b)
	}

	return typesolana.EventSignature(b), nil

}

func ConvertEncodingTypeFromProto(e EncodingType) typesolana.EncodingType {
	switch e {
	case EncodingType_ENCODING_TYPE_BASE58:
		return typesolana.EncodingBase58
	case EncodingType_ENCODING_TYPE_BASE64:
		return typesolana.EncodingBase64
	case EncodingType_ENCODING_TYPE_BASE64_ZSTD:
		return typesolana.EncodingBase64Zstd
	case EncodingType_ENCODING_TYPE_JSON:
		return typesolana.EncodingJSON
	case EncodingType_ENCODING_TYPE_JSON_PARSED:
		return typesolana.EncodingJSONParsed
	default:
		return typesolana.EncodingType("")
	}
}

func ConvertEncodingTypeToProto(e typesolana.EncodingType) EncodingType {
	switch e {
	case typesolana.EncodingBase64:
		return EncodingType_ENCODING_TYPE_BASE64
	case typesolana.EncodingBase58:
		return EncodingType_ENCODING_TYPE_BASE58
	case typesolana.EncodingBase64Zstd:
		return EncodingType_ENCODING_TYPE_BASE64_ZSTD
	case typesolana.EncodingJSONParsed:
		return EncodingType_ENCODING_TYPE_JSON_PARSED
	case typesolana.EncodingJSON:
		return EncodingType_ENCODING_TYPE_JSON
	default:
		return EncodingType_ENCODING_TYPE_NONE
	}
}

func ConvertCommitmentFromProto(c CommitmentType) typesolana.CommitmentType {
	switch c {
	case CommitmentType_COMMITMENT_TYPE_CONFIRMED:
		return typesolana.CommitmentConfirmed
	case CommitmentType_COMMITMENT_TYPE_FINALIZED:
		return typesolana.CommitmentFinalized
	case CommitmentType_COMMITMENT_TYPE_PROCESSED:
		return typesolana.CommitmentProcessed
	default:
		return typesolana.CommitmentType("")
	}
}

func ConvertCommitmentToProto(c typesolana.CommitmentType) CommitmentType {
	switch c {
	case typesolana.CommitmentFinalized:
		return CommitmentType_COMMITMENT_TYPE_FINALIZED
	case typesolana.CommitmentConfirmed:
		return CommitmentType_COMMITMENT_TYPE_CONFIRMED
	case typesolana.CommitmentProcessed:
		return CommitmentType_COMMITMENT_TYPE_PROCESSED
	default:
		return CommitmentType_COMMITMENT_TYPE_NONE
	}
}

func ConvertConfirmationStatusToProto(c typesolana.ConfirmationStatusType) ConfirmationStatusType {
	switch c {
	case typesolana.ConfirmationStatusFinalized:
		return ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_FINALIZED
	case typesolana.ConfirmationStatusConfirmed:
		return ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_CONFIRMED
	case typesolana.ConfirmationStatusProcessed:
		return ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_PROCESSED
	default:
		return ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_NONE
	}
}

func ConvertConfirmationStatusFromProto(c ConfirmationStatusType) typesolana.ConfirmationStatusType {
	switch c {
	case ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_CONFIRMED:
		return typesolana.ConfirmationStatusConfirmed
	case ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_FINALIZED:
		return typesolana.ConfirmationStatusFinalized
	case ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_PROCESSED:
		return typesolana.ConfirmationStatusProcessed
	default:
		return typesolana.ConfirmationStatusType("")
	}
}

func ConvertDataSliceFromProto(p *DataSlice) *typesolana.DataSlice {
	if p == nil {
		return nil
	}
	return &typesolana.DataSlice{
		Offset: ptrUint64(p.Offset),
		Length: ptrUint64(p.Length),
	}
}

func ConvertDataSliceToProto(d *typesolana.DataSlice) *DataSlice {
	if d == nil {
		return nil
	}
	var off, ln uint64
	if d.Offset != nil {
		off = *d.Offset
	}
	if d.Length != nil {
		ln = *d.Length
	}
	return &DataSlice{
		Offset: off,
		Length: ln,
	}
}

func ConvertUiTokenAmountFromProto(p *UiTokenAmount) *typesolana.UiTokenAmount {
	if p == nil {
		return nil
	}
	return &typesolana.UiTokenAmount{
		Amount:         p.Amount,
		Decimals:       uint8(p.Decimals),
		UiAmountString: p.UiAmountString,
	}
}

func ConvertUiTokenAmountToProto(u *typesolana.UiTokenAmount) *UiTokenAmount {
	if u == nil {
		return nil
	}
	return &UiTokenAmount{
		Amount:         u.Amount,
		Decimals:       uint32(u.Decimals),
		UiAmountString: u.UiAmountString,
	}
}

func ConvertAccountFromProto(p *Account) (*typesolana.Account, error) {
	if p == nil {
		return nil, nil
	}
	owner, err := ConvertPublicKeyFromProto(p.Owner)
	if err != nil {
		return nil, fmt.Errorf("owner: %w", err)
	}
	data := ConvertDataBytesOrJSONFromProto(p.Data)

	return &typesolana.Account{
		Lamports:   p.Lamports,
		Owner:      owner,
		Data:       data,
		Executable: p.Executable,
		RentEpoch:  valuespb.NewIntFromBigInt(p.RentEpoch),
		Space:      p.Space,
	}, nil
}

func ConvertAccountToProto(a *typesolana.Account) *Account {
	if a == nil {
		return nil
	}
	return &Account{
		Lamports:   a.Lamports,
		Owner:      a.Owner[:],
		Data:       ConvertDataBytesOrJSONToProto(a.Data),
		Executable: a.Executable,
		RentEpoch:  valuespb.NewBigIntFromInt(a.RentEpoch),
		Space:      a.Space,
	}
}

func ConvertDataBytesOrJSONFromProto(p *DataBytesOrJSON) *typesolana.DataBytesOrJSON {
	if p == nil {
		return nil
	}
	switch t := p.GetBody().(type) {
	case *DataBytesOrJSON_Raw:
		return &typesolana.DataBytesOrJSON{
			AsDecodedBinary: t.Raw,
			RawDataEncoding: ConvertEncodingTypeFromProto(p.Encoding),
		}
	case *DataBytesOrJSON_Json:
		return &typesolana.DataBytesOrJSON{
			AsJSON:          t.Json,
			RawDataEncoding: ConvertEncodingTypeFromProto(p.Encoding),
		}
	}

	return nil
}

func ConvertDataBytesOrJSONToProto(d *typesolana.DataBytesOrJSON) *DataBytesOrJSON {
	if d == nil {
		return nil
	}

	ret := &DataBytesOrJSON{
		Encoding: ConvertEncodingTypeToProto(d.RawDataEncoding),
	}
	if d.AsJSON != nil {
		ret.Body = &DataBytesOrJSON_Json{Json: d.AsJSON}
		return ret
	}

	ret.Body = &DataBytesOrJSON_Raw{Raw: d.AsDecodedBinary}

	return ret
}

func ConvertGetAccountInfoOptsFromProto(p *GetAccountInfoOpts) *typesolana.GetAccountInfoOpts {
	if p == nil {
		return nil
	}
	return &typesolana.GetAccountInfoOpts{
		Encoding:       ConvertEncodingTypeFromProto(p.Encoding),
		Commitment:     ConvertCommitmentFromProto(p.Commitment),
		DataSlice:      ConvertDataSliceFromProto(p.DataSlice),
		MinContextSlot: ptrUint64(p.MinContextSlot),
	}
}

func ConvertGetAccountInfoOptsToProto(o *typesolana.GetAccountInfoOpts) *GetAccountInfoOpts {
	if o == nil {
		return nil
	}
	var min uint64
	if o.MinContextSlot != nil {
		min = *o.MinContextSlot
	}
	return &GetAccountInfoOpts{
		Encoding:       ConvertEncodingTypeToProto(o.Encoding),
		Commitment:     ConvertCommitmentToProto(o.Commitment),
		DataSlice:      ConvertDataSliceToProto(o.DataSlice),
		MinContextSlot: min,
	}
}

func ConvertGetMultipleAccountsOptsFromProto(p *GetMultipleAccountsOpts) *typesolana.GetMultipleAccountsOpts {
	if p == nil {
		return nil
	}
	return &typesolana.GetMultipleAccountsOpts{
		Encoding:       ConvertEncodingTypeFromProto(p.Encoding),
		Commitment:     ConvertCommitmentFromProto(p.Commitment),
		DataSlice:      ConvertDataSliceFromProto(p.DataSlice),
		MinContextSlot: ptrUint64(p.MinContextSlot),
	}
}

func ConvertGetMultipleAccountsOptsToProto(o *typesolana.GetMultipleAccountsOpts) *GetMultipleAccountsOpts {
	if o == nil {
		return nil
	}
	var min uint64
	if o.MinContextSlot != nil {
		min = *o.MinContextSlot
	}
	return &GetMultipleAccountsOpts{
		Encoding:       ConvertEncodingTypeToProto(o.Encoding),
		Commitment:     ConvertCommitmentToProto(o.Commitment),
		DataSlice:      ConvertDataSliceToProto(o.DataSlice),
		MinContextSlot: min,
	}
}

func ConvertGetBlockOptsFromProto(p *GetBlockOpts) *typesolana.GetBlockOpts {
	if p == nil {
		return nil
	}
	return &typesolana.GetBlockOpts{
		Commitment: ConvertCommitmentFromProto(p.Commitment),
	}
}

func ConvertGetBlockOptsToProto(o *typesolana.GetBlockOpts) *GetBlockOpts {
	if o == nil {
		return nil
	}
	return &GetBlockOpts{
		Commitment: ConvertCommitmentToProto(o.Commitment),
	}
}

func ConvertMessageHeaderFromProto(p *MessageHeader) typesolana.MessageHeader {
	if p == nil {
		return typesolana.MessageHeader{}
	}
	return typesolana.MessageHeader{
		NumRequiredSignatures:       uint8(p.NumRequiredSignatures),
		NumReadonlySignedAccounts:   uint8(p.NumReadonlySignedAccounts),
		NumReadonlyUnsignedAccounts: uint8(p.NumReadonlyUnsignedAccounts),
	}
}

func ConvertMessageHeaderToProto(h typesolana.MessageHeader) *MessageHeader {
	return &MessageHeader{
		NumRequiredSignatures:       uint32(h.NumRequiredSignatures),
		NumReadonlySignedAccounts:   uint32(h.NumReadonlySignedAccounts),
		NumReadonlyUnsignedAccounts: uint32(h.NumReadonlyUnsignedAccounts),
	}
}

func ConvertCompiledInstructionFromProto(p *CompiledInstruction) typesolana.CompiledInstruction {
	if p == nil {
		return typesolana.CompiledInstruction{}
	}
	accts := make([]uint16, len(p.Accounts))
	for i, a := range p.Accounts {
		accts[i] = uint16(a)
	}

	return typesolana.CompiledInstruction{
		ProgramIDIndex: uint16(p.ProgramIdIndex),
		Accounts:       accts,
		Data:           p.Data,
		StackHeight:    uint16(p.StackHeight),
	}
}

func ConvertCompiledInstructionToProto(ci typesolana.CompiledInstruction) *CompiledInstruction {
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

func ConvertInnerInstructionFromProto(p *InnerInstruction) typesolana.InnerInstruction {
	if p == nil {
		return typesolana.InnerInstruction{}
	}
	out := typesolana.InnerInstruction{
		Index:        uint16(p.Index),
		Instructions: make([]typesolana.CompiledInstruction, 0, len(p.Instructions)),
	}
	for _, in := range p.Instructions {
		out.Instructions = append(out.Instructions, ConvertCompiledInstructionFromProto(in))
	}
	return out
}

func ConvertInnerInstructionToProto(ii typesolana.InnerInstruction) *InnerInstruction {
	out := &InnerInstruction{
		Index:        uint32(ii.Index),
		Instructions: make([]*CompiledInstruction, 0, len(ii.Instructions)),
	}
	for _, in := range ii.Instructions {
		out.Instructions = append(out.Instructions, ConvertCompiledInstructionToProto(in))
	}
	return out
}

func ConvertLoadedAddressesFromProto(p *LoadedAddresses) typesolana.LoadedAddresses {
	if p == nil {
		return typesolana.LoadedAddresses{}
	}
	ro, _ := ConvertPublicKeysFromProto(p.Readonly)
	wr, _ := ConvertPublicKeysFromProto(p.Writable)
	return typesolana.LoadedAddresses{
		ReadOnly: ro,
		Writable: wr,
	}
}

func ConvertLoadedAddressesToProto(l typesolana.LoadedAddresses) *LoadedAddresses {
	return &LoadedAddresses{
		Readonly: ConvertPublicKeysToProto(l.ReadOnly),
		Writable: ConvertPublicKeysToProto(l.Writable),
	}
}

func ConvertPublicKeysToProto(keys []typesolana.PublicKey) [][]byte {
	out := make([][]byte, 0, len(keys))
	for _, k := range keys {
		out = append(out, k[:])
	}
	return out
}

func ConvertParsedMessageFromProto(p *ParsedMessage) (typesolana.Message, error) {
	if p == nil {
		return typesolana.Message{}, nil
	}
	rb, err := ConvertHashFromProto(p.RecentBlockhash)
	if err != nil {
		return typesolana.Message{}, fmt.Errorf("recent blockhash: %w", err)
	}
	keys, err := ConvertPublicKeysFromProto(p.AccountKeys)
	if err != nil {
		return typesolana.Message{}, fmt.Errorf("account keys: %w", err)
	}
	msg := typesolana.Message{
		AccountKeys:     keys,
		Header:          ConvertMessageHeaderFromProto(p.Header),
		RecentBlockhash: rb,
		Instructions:    make([]typesolana.CompiledInstruction, 0, len(p.Instructions)),
	}
	for _, ins := range p.Instructions {
		msg.Instructions = append(msg.Instructions, ConvertCompiledInstructionFromProto(ins))
	}
	return msg, nil
}

func ConvertParsedMessageToProto(m typesolana.Message) *ParsedMessage {
	out := &ParsedMessage{
		RecentBlockhash: m.RecentBlockhash[:],
		AccountKeys:     ConvertPublicKeysToProto(m.AccountKeys),
		Header:          ConvertMessageHeaderToProto(m.Header),
		Instructions:    make([]*CompiledInstruction, 0, len(m.Instructions)),
	}
	for _, ins := range m.Instructions {
		out.Instructions = append(out.Instructions, ConvertCompiledInstructionToProto(ins))
	}
	return out
}

func ConvertParsedTransactionFromProto(p *ParsedTransaction) (typesolana.Transaction, error) {
	if p == nil {
		return typesolana.Transaction{}, nil
	}
	sigs, err := ConvertSignaturesFromProto(p.Signatures)
	if err != nil {
		return typesolana.Transaction{}, fmt.Errorf("signatures: %w", err)
	}
	msg, err := ConvertParsedMessageFromProto(p.Message)
	if err != nil {
		return typesolana.Transaction{}, fmt.Errorf("message: %w", err)
	}
	return typesolana.Transaction{
		Signatures: sigs,
		Message:    msg,
	}, nil
}

func ConvertParsedTransactionToProto(t typesolana.Transaction) *ParsedTransaction {
	return &ParsedTransaction{
		Signatures: ConvertSignaturesToProto(t.Signatures),
		Message:    ConvertParsedMessageToProto(t.Message),
	}
}

func ConvertTokenBalanceFromProto(p *TokenBalance) (*typesolana.TokenBalance, error) {
	if p == nil {
		return nil, nil
	}
	mint, err := ConvertPublicKeyFromProto(p.Mint)
	if err != nil {
		return nil, fmt.Errorf("mint: %w", err)
	}
	var owner *typesolana.PublicKey
	if len(p.Owner) > 0 {
		o, err := ConvertPublicKeyFromProto(p.Owner)
		if err != nil {
			return nil, fmt.Errorf("owner: %w", err)
		}
		owner = &o
	}
	var program *typesolana.PublicKey
	if len(p.ProgramId) > 0 {
		prog, err := ConvertPublicKeyFromProto(p.ProgramId)
		if err != nil {
			return nil, fmt.Errorf("programId: %w", err)
		}
		program = &prog
	}
	return &typesolana.TokenBalance{
		AccountIndex:  uint16(p.AccountIndex),
		Owner:         owner,
		ProgramId:     program,
		Mint:          mint,
		UiTokenAmount: ConvertUiTokenAmountFromProto(p.Ui),
	}, nil
}

func ConvertTokenBalanceToProto(tb *typesolana.TokenBalance) *TokenBalance {
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
		Ui:           ConvertUiTokenAmountToProto(tb.UiTokenAmount),
	}
}

func ConvertReturnDataFromProto(p *ReturnData) (*typesolana.ReturnData, error) {
	if p == nil {
		return nil, nil
	}
	prog, err := ConvertPublicKeyFromProto(p.ProgramId)
	if err != nil {
		return nil, fmt.Errorf("programId: %w", err)
	}
	return &typesolana.ReturnData{
		ProgramId: prog,
		Data: typesolana.Data{
			Content:  p.Data.GetContent(),
			Encoding: ConvertEncodingTypeFromProto(p.Data.GetEncoding()),
		},
	}, nil
}

func ConvertReturnDataToProto(r *typesolana.ReturnData) *ReturnData {
	if r == nil {
		return nil
	}

	return &ReturnData{
		ProgramId: r.ProgramId[:],
		Data: &Data{
			Content:  r.Data.Content,
			Encoding: ConvertEncodingTypeToProto(r.Data.Encoding),
		},
	}
}

func ConvertTransactionMetaFromProto(p *TransactionMeta) (*typesolana.TransactionMeta, error) {
	if p == nil {
		return nil, nil
	}
	preTB := make([]typesolana.TokenBalance, 0, len(p.PreTokenBalances))
	for _, x := range p.PreTokenBalances {
		tb, err := ConvertTokenBalanceFromProto(x)
		if err != nil {
			return nil, fmt.Errorf("pre token balance: %w", err)
		}
		preTB = append(preTB, *tb)
	}
	postTB := make([]typesolana.TokenBalance, 0, len(p.PostTokenBalances))
	for _, x := range p.PostTokenBalances {
		tb, err := ConvertTokenBalanceFromProto(x)
		if err != nil {
			return nil, fmt.Errorf("post token balance: %w", err)
		}
		postTB = append(postTB, *tb)
	}
	inner := make([]typesolana.InnerInstruction, 0, len(p.InnerInstructions))
	for _, in := range p.InnerInstructions {
		inner = append(inner, ConvertInnerInstructionFromProto(in))
	}
	ret, err := ConvertReturnDataFromProto(p.ReturnData)
	if err != nil {
		return nil, fmt.Errorf("return data: %w", err)
	}
	la := ConvertLoadedAddressesFromProto(p.LoadedAddresses)

	meta := &typesolana.TransactionMeta{
		Err:                  p.ErrJson,
		Fee:                  p.Fee,
		PreBalances:          p.PreBalances,
		PostBalances:         p.PostBalances,
		InnerInstructions:    inner,
		PreTokenBalances:     preTB,
		PostTokenBalances:    postTB,
		LogMessages:          p.LogMessages,
		LoadedAddresses:      la,
		ReturnData:           *ret,
		ComputeUnitsConsumed: p.ComputeUnitsConsumed,
	}
	return meta, nil
}

func ConvertTransactionMetaToProto(m *typesolana.TransactionMeta) *TransactionMeta {
	if m == nil {
		return nil
	}
	preTB := make([]*TokenBalance, 0, len(m.PreTokenBalances))
	for i := range m.PreTokenBalances {
		preTB = append(preTB, ConvertTokenBalanceToProto(&m.PreTokenBalances[i]))
	}
	postTB := make([]*TokenBalance, 0, len(m.PostTokenBalances))
	for i := range m.PostTokenBalances {
		postTB = append(postTB, ConvertTokenBalanceToProto(&m.PostTokenBalances[i]))
	}
	inner := make([]*InnerInstruction, 0, len(m.InnerInstructions))
	for _, in := range m.InnerInstructions {
		inner = append(inner, ConvertInnerInstructionToProto(in))
	}
	var cuc uint64
	if m.ComputeUnitsConsumed != nil {
		cuc = *m.ComputeUnitsConsumed
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
		LoadedAddresses:      ConvertLoadedAddressesToProto(m.LoadedAddresses),
		ReturnData:           ConvertReturnDataToProto(&m.ReturnData),
		ComputeUnitsConsumed: &cuc,
	}
}

func ConvertTransactionEnvelopeFromProto(p *TransactionEnvelope) (typesolana.TransactionResultEnvelope, error) {
	if p == nil {
		return typesolana.TransactionResultEnvelope{}, nil
	}
	var out typesolana.TransactionResultEnvelope
	switch t := p.GetTransaction().(type) {
	case *TransactionEnvelope_Raw:
		out.AsDecodedBinary = typesolana.Data{
			Content:  t.Raw,
			Encoding: typesolana.EncodingBase64,
		}
	case *TransactionEnvelope_Parsed:
		ptx, err := ConvertParsedTransactionFromProto(t.Parsed)
		if err != nil {
			return typesolana.TransactionResultEnvelope{}, err
		}
		out.AsParsedTransaction = &ptx
	default:
	}
	return out, nil
}

func ConvertTransactionEnvelopeToProto(e typesolana.TransactionResultEnvelope) *TransactionEnvelope {
	switch {
	case e.AsParsedTransaction != nil:
		return &TransactionEnvelope{
			Transaction: &TransactionEnvelope_Parsed{
				Parsed: ConvertParsedTransactionToProto(*e.AsParsedTransaction),
			},
		}
	case len(e.AsDecodedBinary.Content) > 0:
		return &TransactionEnvelope{
			Transaction: &TransactionEnvelope_Raw{
				Raw: e.AsDecodedBinary.Content,
			},
		}
	default:
		return &TransactionEnvelope{}
	}
}

func ConvertGetTransactionReplyFromProto(p *GetTransactionReply) (*typesolana.GetTransactionReply, error) {
	if p == nil {
		return nil, nil
	}
	env, err := ConvertTransactionEnvelopeFromProto(p.Transaction)
	if err != nil {
		return nil, err
	}
	meta, err := ConvertTransactionMetaFromProto(p.Meta)
	if err != nil {
		return nil, err
	}
	var bt *typesolana.UnixTimeSeconds
	if p.BlockTime != nil {
		bt = ptrUnix(typesolana.UnixTimeSeconds(*p.BlockTime))
	}

	return &typesolana.GetTransactionReply{
		Slot:        p.Slot,
		BlockTime:   bt,
		Transaction: &env,
		Meta:        meta,
	}, nil
}

func ConvertGetTransactionReplyToProto(r *typesolana.GetTransactionReply) *GetTransactionReply {
	if r == nil {
		return nil
	}
	var bt int64
	if r.BlockTime != nil {
		bt = int64(*r.BlockTime)
	}
	var tx *TransactionEnvelope
	if r.Transaction != nil {
		tx = ConvertTransactionEnvelopeToProto(*r.Transaction)
	}
	return &GetTransactionReply{
		Slot:        r.Slot,
		BlockTime:   &bt,
		Transaction: tx,
		Meta:        ConvertTransactionMetaToProto(r.Meta),
	}
}

func ConvertGetTransactionRequestFromProto(p *GetTransactionRequest) (typesolana.GetTransactionRequest, error) {
	sig, err := ConvertSignatureFromProto(p.Signature)
	if err != nil {
		return typesolana.GetTransactionRequest{}, err
	}
	return typesolana.GetTransactionRequest{Signature: sig}, nil
}

func ConvertGetTransactionRequestToProto(r typesolana.GetTransactionRequest) *GetTransactionRequest {
	return &GetTransactionRequest{Signature: r.Signature[:]}
}

func ConvertGetBalanceReplyFromProto(p *GetBalanceReply) *typesolana.GetBalanceReply {
	if p == nil {
		return nil
	}
	return &typesolana.GetBalanceReply{Value: p.Value}
}

func ConvertGetBalanceReplyToProto(r *typesolana.GetBalanceReply) *GetBalanceReply {
	if r == nil {
		return nil
	}
	return &GetBalanceReply{Value: r.Value}
}

func ConvertGetBalanceRequestFromProto(p *GetBalanceRequest) (typesolana.GetBalanceRequest, error) {
	pk, err := ConvertPublicKeyFromProto(p.Addr)
	if err != nil {
		return typesolana.GetBalanceRequest{}, err
	}
	return typesolana.GetBalanceRequest{
		Addr:       pk,
		Commitment: ConvertCommitmentFromProto(p.Commitment),
	}, nil
}

func ConvertGetBalanceRequestToProto(r typesolana.GetBalanceRequest) *GetBalanceRequest {
	return &GetBalanceRequest{
		Addr:       r.Addr[:],
		Commitment: ConvertCommitmentToProto(r.Commitment),
	}
}

func ConvertGetSlotHeightReplyFromProto(p *GetSlotHeightReply) *typesolana.GetSlotHeightReply {
	if p == nil {
		return nil
	}
	return &typesolana.GetSlotHeightReply{Height: p.Height}
}

func ConvertGetSlotHeightReplyToProto(r *typesolana.GetSlotHeightReply) *GetSlotHeightReply {
	if r == nil {
		return nil
	}
	return &GetSlotHeightReply{Height: r.Height}
}

func ConvertGetSlotHeightRequestFromProto(p *GetSlotHeightRequest) typesolana.GetSlotHeightRequest {
	return typesolana.GetSlotHeightRequest{Commitment: ConvertCommitmentFromProto(p.Commitment)}
}

func ConvertGetSlotHeightRequestToProto(r typesolana.GetSlotHeightRequest) *GetSlotHeightRequest {
	return &GetSlotHeightRequest{Commitment: ConvertCommitmentToProto(r.Commitment)}
}

func ConvertGetBlockOptsReplyFromProto(p *GetBlockReply) (*typesolana.GetBlockReply, error) {
	if p == nil {
		return nil, nil
	}
	hash, err := ConvertHashFromProto(p.Blockhash)
	if err != nil {
		return nil, fmt.Errorf("blockhash: %w", err)
	}
	prev, err := ConvertHashFromProto(p.PreviousBlockhash)
	if err != nil {
		return nil, fmt.Errorf("previous blockhash: %w", err)
	}

	var bt *solana.UnixTimeSeconds
	if p.BlockTime != nil {
		bt = ptrUnix(typesolana.UnixTimeSeconds(*p.BlockTime))
	}

	return &typesolana.GetBlockReply{
		Blockhash:         hash,
		PreviousBlockhash: prev,
		ParentSlot:        p.ParentSlot,
		BlockTime:         bt,
		BlockHeight:       ptrUint64(p.BlockHeight),
	}, nil
}

func ConvertGetBlockReplyToProto(r *typesolana.GetBlockReply) *GetBlockReply {
	if r == nil {
		return nil
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
	}
}

func ConvertGetBlockRequestFromProto(p *GetBlockRequest) *typesolana.GetBlockRequest {
	if p == nil {
		return nil
	}
	return &typesolana.GetBlockRequest{
		Slot: p.Slot,
		Opts: ConvertGetBlockOptsFromProto(p.Opts),
	}
}

func ConvertGetBlockRequestToProto(r *typesolana.GetBlockRequest) *GetBlockRequest {
	if r == nil {
		return nil
	}
	return &GetBlockRequest{
		Slot: r.Slot,
		Opts: ConvertGetBlockOptsToProto(r.Opts),
	}
}

func ConvertGetFeeForMessageRequestFromProto(p *GetFeeForMessageRequest) *typesolana.GetFeeForMessageRequest {
	if p == nil {
		return nil
	}
	return &typesolana.GetFeeForMessageRequest{
		Message:    p.Message,
		Commitment: ConvertCommitmentFromProto(p.Commitment),
	}
}

func ConvertGetFeeForMessageRequestToProto(r *typesolana.GetFeeForMessageRequest) *GetFeeForMessageRequest {
	if r == nil {
		return nil
	}
	return &GetFeeForMessageRequest{
		Message:    r.Message,
		Commitment: ConvertCommitmentToProto(r.Commitment),
	}
}

func ConvertGetFeeForMessageReplyFromProto(p *GetFeeForMessageReply) *typesolana.GetFeeForMessageReply {
	if p == nil {
		return nil
	}
	return &typesolana.GetFeeForMessageReply{Fee: p.Fee}
}

func ConvertGetFeeForMessageReplyToProto(r *typesolana.GetFeeForMessageReply) *GetFeeForMessageReply {
	if r == nil {
		return nil
	}
	return &GetFeeForMessageReply{Fee: r.Fee}
}

func ConvertGetMultipleAccountsRequestFromProto(p *GetMultipleAccountsWithOptsRequest) *typesolana.GetMultipleAccountsRequest {
	if p == nil {
		return nil
	}
	accts, _ := ConvertPublicKeysFromProto(p.Accounts)
	return &typesolana.GetMultipleAccountsRequest{
		Accounts: accts,
		Opts:     ConvertGetMultipleAccountsOptsFromProto(p.Opts),
	}
}

func ConvertGetMultipleAccountsRequestToProto(r *typesolana.GetMultipleAccountsRequest) *GetMultipleAccountsWithOptsRequest {
	if r == nil {
		return nil
	}
	return &GetMultipleAccountsWithOptsRequest{
		Accounts: ConvertPublicKeysToProto(r.Accounts),
		Opts:     ConvertGetMultipleAccountsOptsToProto(r.Opts),
	}
}

func ConvertGetMultipleAccountsReplyFromProto(p *GetMultipleAccountsWithOptsReply) (*typesolana.GetMultipleAccountsReply, error) {
	if p == nil {
		return nil, nil
	}
	val := make([]*typesolana.Account, 0, len(p.Value))
	for _, a := range p.Value {
		acc, err := ConvertAccountFromProto(a.Account)
		if err != nil {
			return nil, err
		}
		val = append(val, acc)
	}
	return &typesolana.GetMultipleAccountsReply{
		Value: val,
	}, nil
}

func ConvertGetMultipleAccountsReplyToProto(r *typesolana.GetMultipleAccountsReply) *GetMultipleAccountsWithOptsReply {
	if r == nil {
		return nil
	}
	val := make([]*OptionalAccountWrapper, 0, len(r.Value))
	for _, a := range r.Value {
		val = append(val, &OptionalAccountWrapper{Account: ConvertAccountToProto(a)})
	}

	return &GetMultipleAccountsWithOptsReply{
		Value: val,
	}
}

func ConvertGetSignatureStatusesRequestFromProto(p *GetSignatureStatusesRequest) (*typesolana.GetSignatureStatusesRequest, error) {
	if p == nil {
		return nil, nil
	}
	sigs, err := ConvertSignaturesFromProto(p.Sigs)
	if err != nil {
		return nil, err
	}
	return &typesolana.GetSignatureStatusesRequest{Sigs: sigs}, nil
}

func ConvertGetSignatureStatusesRequestToProto(r *typesolana.GetSignatureStatusesRequest) *GetSignatureStatusesRequest {
	if r == nil {
		return nil
	}
	return &GetSignatureStatusesRequest{Sigs: ConvertSignaturesToProto(r.Sigs)}
}

func ConvertGetSignatureStatusesReplyFromProto(p *GetSignatureStatusesReply) *typesolana.GetSignatureStatusesReply {
	if p == nil {
		return nil
	}
	out := &typesolana.GetSignatureStatusesReply{Results: make([]typesolana.GetSignatureStatusesResult, 0, len(p.Results))}
	for _, r := range p.Results {
		out.Results = append(out.Results, typesolana.GetSignatureStatusesResult{
			Slot:               r.Slot,
			Confirmations:      r.Confirmations,
			Err:                r.Err,
			ConfirmationStatus: ConvertConfirmationStatusFromProto(r.ConfirmationStatus),
		})
	}
	return out
}

func ConvertGetSignatureStatusesReplyToProto(r *typesolana.GetSignatureStatusesReply) *GetSignatureStatusesReply {
	if r == nil {
		return nil
	}
	out := &GetSignatureStatusesReply{Results: make([]*GetSignatureStatusesResult, 0, len(r.Results))}
	for i := range r.Results {
		var conf uint64
		if r.Results[i].Confirmations != nil {
			conf = *r.Results[i].Confirmations
		}
		out.Results = append(out.Results, &GetSignatureStatusesResult{
			Slot:               r.Results[i].Slot,
			Confirmations:      ptrUint64(conf),
			Err:                r.Results[i].Err,
			ConfirmationStatus: ConvertConfirmationStatusToProto(r.Results[i].ConfirmationStatus),
		})
	}
	return out
}

func ConvertSimulateTransactionAccountsOptsFromProto(p *SimulateTransactionAccountsOpts) *typesolana.SimulateTransactionAccountsOpts {
	if p == nil {
		return nil
	}
	addrs, _ := ConvertPublicKeysFromProto(p.Addresses)
	return &typesolana.SimulateTransactionAccountsOpts{
		Encoding:  ConvertEncodingTypeFromProto(p.Encoding),
		Addresses: addrs,
	}
}

func ConvertSimulateTransactionAccountsOptsToProto(o *typesolana.SimulateTransactionAccountsOpts) *SimulateTransactionAccountsOpts {
	if o == nil {
		return nil
	}
	return &SimulateTransactionAccountsOpts{
		Encoding:  ConvertEncodingTypeToProto(o.Encoding),
		Addresses: ConvertPublicKeysToProto(o.Addresses),
	}
}

func ConvertSimulateTXOptsFromProto(p *SimulateTXOpts) *typesolana.SimulateTXOpts {
	if p == nil {
		return nil
	}
	return &typesolana.SimulateTXOpts{
		SigVerify:              p.SigVerify,
		Commitment:             ConvertCommitmentFromProto(p.Commitment),
		ReplaceRecentBlockhash: p.ReplaceRecentBlockhash,
		Accounts:               ConvertSimulateTransactionAccountsOptsFromProto(p.Accounts),
	}
}

func ConvertSimulateTXOptsToProto(o *typesolana.SimulateTXOpts) *SimulateTXOpts {
	if o == nil {
		return nil
	}
	return &SimulateTXOpts{
		SigVerify:              o.SigVerify,
		Commitment:             ConvertCommitmentToProto(o.Commitment),
		ReplaceRecentBlockhash: o.ReplaceRecentBlockhash,
		Accounts:               ConvertSimulateTransactionAccountsOptsToProto(o.Accounts),
	}
}

func ConvertSimulateTXRequestFromProto(p *SimulateTXRequest) (typesolana.SimulateTXRequest, error) {
	recv, err := ConvertPublicKeyFromProto(p.Receiver)
	if err != nil {
		return typesolana.SimulateTXRequest{}, fmt.Errorf("receiver: %w", err)
	}
	return typesolana.SimulateTXRequest{
		Receiver:           recv,
		EncodedTransaction: p.EncodedTransaction,
		Opts:               ConvertSimulateTXOptsFromProto(p.Opts),
	}, nil
}

func ConvertSimulateTXRequestToProto(r typesolana.SimulateTXRequest) *SimulateTXRequest {
	return &SimulateTXRequest{
		Receiver:           r.Receiver[:],
		EncodedTransaction: r.EncodedTransaction,
		Opts:               ConvertSimulateTXOptsToProto(r.Opts),
	}
}

func ConvertSimulateTXReplyFromProto(p *SimulateTXReply) (*typesolana.SimulateTXReply, error) {
	if p == nil {
		return nil, nil
	}
	accs := make([]*typesolana.Account, 0, len(p.Accounts))
	for _, a := range p.Accounts {
		acc, err := ConvertAccountFromProto(a)
		if err != nil {
			return nil, err
		}
		accs = append(accs, acc)
	}
	return &typesolana.SimulateTXReply{
		Err:           p.Err,
		Logs:          p.Logs,
		Accounts:      accs,
		UnitsConsumed: ptrUint64(p.UnitsConsumed),
	}, nil
}

func ConvertSimulateTXReplyToProto(r *typesolana.SimulateTXReply) *SimulateTXReply {
	if r == nil {
		return nil
	}
	var units uint64
	if r.UnitsConsumed != nil {
		units = *r.UnitsConsumed
	}
	out := &SimulateTXReply{
		Err:           r.Err,
		Logs:          r.Logs,
		UnitsConsumed: units,
	}
	for _, a := range r.Accounts {
		out.Accounts = append(out.Accounts, ConvertAccountToProto(a))
	}
	return out
}

func ConvertComputeConfigFromProto(p *ComputeConfig) *typesolana.ComputeConfig {
	if p == nil {
		return nil
	}
	return &typesolana.ComputeConfig{
		ComputeLimit:    &p.ComputeLimit,
		ComputeMaxPrice: &p.ComputeMaxPrice,
	}
}

func ConvertComputeConfigToProto(c *typesolana.ComputeConfig) *ComputeConfig {
	if c == nil {
		return nil
	}

	var cl uint32
	if c.ComputeLimit != nil {
		cl = *c.ComputeLimit
	}
	var cmp uint64
	if c.ComputeMaxPrice != nil {
		cmp = *c.ComputeMaxPrice
	}
	return &ComputeConfig{
		ComputeLimit:    cl,
		ComputeMaxPrice: cmp,
	}
}

func ConvertSubmitTransactionRequestFromProto(p *SubmitTransactionRequest) (typesolana.SubmitTransactionRequest, error) {
	if p == nil {
		return typesolana.SubmitTransactionRequest{}, nil
	}
	rcv, err := ConvertPublicKeyFromProto(p.Receiver)
	if err != nil {
		return typesolana.SubmitTransactionRequest{}, fmt.Errorf("receiver: %w", err)
	}
	return typesolana.SubmitTransactionRequest{
		Cfg:                ConvertComputeConfigFromProto(p.Cfg),
		Receiver:           rcv,
		EncodedTransaction: p.EncodedTransaction,
	}, nil
}

func ConvertSubmitTransactionRequestToProto(r typesolana.SubmitTransactionRequest) *SubmitTransactionRequest {
	return &SubmitTransactionRequest{
		Cfg:                ConvertComputeConfigToProto(r.Cfg),
		Receiver:           r.Receiver[:],
		EncodedTransaction: r.EncodedTransaction,
	}
}

func ConvertSubmitTransactionReplyFromProto(p *SubmitTransactionReply) (*typesolana.SubmitTransactionReply, error) {
	if p == nil {
		return nil, nil
	}
	sig, err := ConvertSignatureFromProto(p.Signature)
	if err != nil {
		return nil, err
	}
	return &typesolana.SubmitTransactionReply{
		Signature:      sig,
		IdempotencyKey: p.IdempotencyKey,
		Status:         typesolana.TransactionStatus(p.Status),
	}, nil
}

func ConvertSubmitTransactionReplyToProto(r *typesolana.SubmitTransactionReply) *SubmitTransactionReply {
	if r == nil {
		return nil
	}
	return &SubmitTransactionReply{
		Signature:      r.Signature[:],
		IdempotencyKey: r.IdempotencyKey,
		Status:         TxStatus(r.Status),
	}
}

func ConvertRPCContextFromProto(p *RPCContext) typesolana.RPCContext {
	if p == nil {
		return typesolana.RPCContext{}
	}
	return typesolana.RPCContext{Slot: p.Slot}
}

func ConvertRPCContextToProto(r typesolana.RPCContext) *RPCContext {
	return &RPCContext{Slot: r.Slot}
}

func ConvertLogFromProto(p *Log) (*typesolana.Log, error) {
	if p == nil {
		return nil, nil
	}
	addr, err := ConvertPublicKeyFromProto(p.Address)
	if err != nil && len(p.Address) > 0 { // address optional
		return nil, fmt.Errorf("address: %w", err)
	}
	ev, _ := ConvertEventSigFromProto(p.EventSig)
	tx, err := ConvertSignatureFromProto(p.TxHash)
	if err != nil && len(p.TxHash) > 0 {
		return nil, fmt.Errorf("txHash: %w", err)
	}
	bh, _ := ConvertHashFromProto(p.BlockHash)
	var lErr *string
	if p.Error != "" {
		lErr = &p.Error
	}

	return &typesolana.Log{
		ChainID:        p.ChainId,
		LogIndex:       p.LogIndex,
		BlockHash:      bh,
		BlockNumber:    p.BlockNumber,
		BlockTimestamp: uint64(p.BlockTimestamp),
		Address:        addr,
		EventSig:       ev,
		TxHash:         tx,
		Data:           p.Data,
		SequenceNum:    p.SequenceNum,
		Error:          lErr,
	}, nil
}

func ConvertLogToProto(l *typesolana.Log) *Log {
	if l == nil {
		return nil
	}
	var err string
	if l.Error != nil {
		err = *l.Error
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
		Error:          err,
	}
}

func ConvertExpressionsToProto(expressions []query.Expression) ([]*Expression, error) {
	protoExpressions := make([]*Expression, 0, len(expressions))
	for _, expr := range expressions {
		protoExpression, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, err
		}
		protoExpressions = append(protoExpressions, protoExpression)
	}
	return protoExpressions, nil
}

func convertExpressionToProto(expression query.Expression) (*Expression, error) {
	pbExpression := &Expression{}
	if expression.IsPrimitive() {
		ep := &Primitive{}
		switch primitive := expression.Primitive.(type) {
		case *solprimitives.Address:
			ep.Primitive = &Primitive_Address{Address: primitive.PubKey[:]}

			putPrimitive(pbExpression, ep)
		case *solprimitives.EventSig:
			ep.Primitive = &Primitive_EventSig{EventSig: primitive.Sig[:]}

			putPrimitive(pbExpression, ep)
		case *solprimitives.EventBySubkey:
			ep.Primitive = &Primitive_EventBySubkey{
				EventBySubkey: &EventBySubkey{
					//nolint: gosec // G115
					SubkeyIndex:    primitive.SubKeyIndex,
					ValueComparers: ConvertValueComparatorsToProto(primitive.ValueComparers),
				},
			}

			putPrimitive(pbExpression, ep)
		default:
			generalPrimitive, err := chaincommonpb.ConvertPrimitiveToProto(primitive, func(value any) (*codecpb.VersionedBytes, error) {
				return nil, fmt.Errorf("unsupported primitive type: %T", value)
			})
			if err != nil {
				return nil, err
			}
			putGeneralPrimitive(pbExpression, generalPrimitive)
		}
		return pbExpression, nil
	}

	pbExpression.Evaluator = &Expression_BooleanExpression{BooleanExpression: &BooleanExpression{}}
	expressions := make([]*Expression, 0)
	for _, expr := range expression.BoolExpression.Expressions {
		pbExpr, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, pbExpr)
	}
	pbExpression.Evaluator = &Expression_BooleanExpression{
		BooleanExpression: &BooleanExpression{
			//nolint: gosec // G115
			BooleanOperator: chaincommonpb.BooleanOperator(expression.BoolExpression.BoolOperator),
			Expression:      expressions,
		}}

	return pbExpression, nil
}

func ConvertValueComparatorsToProto(comparators []solprimitives.IndexedValueComparator) []*IndexedValueComparator {
	if len(comparators) == 0 {
		return nil
	}

	out := make([]*IndexedValueComparator, 0, len(comparators))
	for _, c := range comparators {
		out = append(out, &IndexedValueComparator{
			Value:    c.Value,
			Operator: chaincommonpb.ComparisonOperator(c.Operator),
		})
	}

	return out
}

func ConvertValueCompraratorsFromProto(comparators []*IndexedValueComparator) []solprimitives.IndexedValueComparator {
	if len(comparators) == 0 {
		return nil
	}

	out := make([]solprimitives.IndexedValueComparator, 0, len(comparators))
	for _, c := range comparators {
		out = append(out, solprimitives.IndexedValueComparator{
			Value:    c.Value,
			Operator: primitives.ComparisonOperator(c.Operator),
		})
	}
	return out
}

func putGeneralPrimitive(exp *Expression, p *chaincommonpb.Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: &Primitive_GeneralPrimitive{GeneralPrimitive: p}}}
}

func ConvertExpressionsFromProto(protoExpressions []*Expression) ([]query.Expression, error) {
	expressions := make([]query.Expression, 0, len(protoExpressions))
	if len(protoExpressions) == 0 {
		return nil, nil
	}
	for idx, protoExpression := range protoExpressions {
		expr, err := convertExpressionFromProto(protoExpression)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "err to convert expr idx %d err: %s", idx, err.Error())
		}

		expressions = append(expressions, expr)
	}
	return expressions, nil
}

func convertExpressionFromProto(protoExpression *Expression) (query.Expression, error) {
	if protoExpression == nil {
		return query.Expression{}, errors.New("expression can not be nil")
	}

	switch protoEvaluatedExpr := protoExpression.GetEvaluator().(type) {
	case *Expression_BooleanExpression:
		var expressions []query.Expression
		for idx, expression := range protoEvaluatedExpr.BooleanExpression.GetExpression() {
			convertedExpression, err := convertExpressionFromProto(expression)
			if err != nil {
				return query.Expression{}, fmt.Errorf("failed to convert sub-expression %d: %w", idx, err)
			}
			expressions = append(expressions, convertedExpression)
		}
		if protoEvaluatedExpr.BooleanExpression.GetBooleanOperator() == chaincommonpb.BooleanOperator_AND {
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil

	case *Expression_Primitive:
		switch primitive := protoEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *Primitive_GeneralPrimitive:
			return chaincommonpb.ConvertPrimitiveFromProto(primitive.GeneralPrimitive, func(_ string, _ bool) (any, error) {
				return nil, fmt.Errorf("unsupported primitive type: %T", primitive)
			})
		default:
			return convertSolPrimitiveFromProto(protoEvaluatedExpr.Primitive)
		}
	default:
		return query.Expression{}, fmt.Errorf("unknown expression type: %T", protoExpression.GetEvaluator())
	}
}

func convertSolPrimitiveFromProto(protoPrimitive *Primitive) (query.Expression, error) {
	switch primitive := protoPrimitive.GetPrimitive().(type) {
	case *Primitive_Address:
		publicKey, err := ConvertPublicKeyFromProto(primitive.Address)
		if err != nil {
			return query.Expression{}, fmt.Errorf("convert expr err: %w", err)
		}
		return solprimitives.NewAddressFilter(publicKey), nil
	case *Primitive_EventSig:
		sig, err := ConvertEventSigFromProto(primitive.EventSig)
		if err != nil {
			return query.Expression{}, fmt.Errorf("failed to convert event sig: %w", err)
		}
		return solprimitives.NewEventSigFilter(sig), nil
	case *Primitive_EventBySubkey:
		return solprimitives.NewEventBySubkeyFilter(primitive.EventBySubkey.SubkeyIndex, ConvertValueCompraratorsFromProto(primitive.EventBySubkey.ValueComparers)), nil
	default:
		return query.Expression{}, fmt.Errorf("unknown primitive type: %T", primitive)
	}
}

func ConvertLPFilterQueryFromProto(p *LPFilterQuery) (*typesolana.LPFilterQuery, error) {
	if p == nil {
		return nil, nil
	}
	var addr typesolana.PublicKey
	var err error
	if len(p.Address) > 0 {
		addr, err = ConvertPublicKeyFromProto(p.Address)
		if err != nil {
			return nil, fmt.Errorf("filter.address: %w", err)
		}
	}
	var err2 error
	var eventSig typesolana.EventSignature
	if len(p.EventSig) > 0 {
		eventSig, err2 = ConvertEventSigFromProto(p.EventSig)
		if err != nil {
			return nil, fmt.Errorf("filter.event_sig: %w", err2)
		}
	}

	return &typesolana.LPFilterQuery{
		Name:            p.Name,
		Address:         addr,
		EventName:       p.EventName,
		EventSig:        eventSig,
		StartingBlock:   p.StartingBlock,
		ContractIdlJSON: p.ContractIdlJson,
		SubkeyPaths:     ConvertSubkeyPathsFromProto(p.SubkeyPaths),
		Retention:       time.Duration(p.Retention),
		MaxLogsKept:     p.MaxLogsKept,
		IncludeReverted: p.IncludeReverted,
	}, nil
}

func ConvertSubkeyPathsFromProto(skeys []*Subkeys) [][]string {
	if len(skeys) == 0 {
		return nil
	}
	out := make([][]string, 0, len(skeys))
	for _, k := range skeys {
		out = append(out, k.Subkeys)
	}

	return out
}

func ConvertSubkeyPathsToProto(keys [][]string) []*Subkeys {
	if len(keys) == 0 {
		return nil
	}
	out := make([]*Subkeys, 0, len(keys))
	for _, k := range keys {
		out = append(out, &Subkeys{
			Subkeys: k,
		})
	}

	return out
}

func ConvertLPFilterQueryToProto(f *typesolana.LPFilterQuery) *LPFilterQuery {
	if f == nil {
		return nil
	}

	return &LPFilterQuery{
		Name:            f.Name,
		Address:         f.Address[:],
		EventName:       f.EventName,
		EventSig:        f.EventSig[:],
		StartingBlock:   f.StartingBlock,
		ContractIdlJson: f.ContractIdlJSON,
		SubkeyPaths:     ConvertSubkeyPathsToProto(f.SubkeyPaths),
		Retention:       int64(f.Retention),
		MaxLogsKept:     f.MaxLogsKept,
		IncludeReverted: f.IncludeReverted,
	}
}

func putPrimitive(exp *Expression, p *Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: p.Primitive}}
}

func ptrUint64(v uint64) *uint64 {
	if v == 0 {
		return nil
	}
	return &v

}
func ptrBool(v bool) *bool                                             { return &v }
func ptrUnix(v typesolana.UnixTimeSeconds) *typesolana.UnixTimeSeconds { return &v }
