package aptos

import (
	"fmt"

	typeaptos "github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

// ConvertViewPayloadFromProto converts a proto ViewPayload to Go types
func ConvertViewPayloadFromProto(proto *ViewPayload) (*typeaptos.ViewPayload, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto payload is nil")
	}

	if len(proto.Module.Address) != typeaptos.AccountAddressLength {
		return nil, fmt.Errorf("invalid address length: expected %d, got %d", typeaptos.AccountAddressLength, len(proto.Module.Address))
	}

	var address typeaptos.AccountAddress
	copy(address[:], proto.Module.Address)

	module := typeaptos.ModuleID{
		Address: address,
		Name:    proto.Module.Name,
	}

	argTypes := make([]typeaptos.TypeTag, len(proto.ArgTypes))
	for i, protoTypeTag := range proto.ArgTypes {
		typeTag, err := ConvertTypeTagFromProto(protoTypeTag)
		if err != nil {
			return nil, fmt.Errorf("failed to convert arg type %d: %w", i, err)
		}
		argTypes[i] = *typeTag
	}

	args := make([][]byte, len(proto.Args))
	for i, protoArg := range proto.Args {
		args[i] = protoArg
	}

	return &typeaptos.ViewPayload{
		Module:   module,
		Function: proto.Function,
		ArgTypes: argTypes,
		Args:     args,
	}, nil
}

// ConvertViewPayloadToProto converts a Go ViewPayload to proto types
func ConvertViewPayloadToProto(payload *typeaptos.ViewPayload) (*ViewPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload is nil")
	}

	protoModule := &ModuleID{
		Address: payload.Module.Address[:],
		Name:    payload.Module.Name,
	}

	protoArgTypes := make([]*TypeTag, len(payload.ArgTypes))
	for i, argType := range payload.ArgTypes {
		protoTypeTag, err := ConvertTypeTagToProto(&argType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert arg type %d: %w", i, err)
		}
		protoArgTypes[i] = protoTypeTag
	}

	protoArgs := make([][]byte, len(payload.Args))
	for i, arg := range payload.Args {
		protoArgs[i] = arg
	}

	return &ViewPayload{
		Module:   protoModule,
		Function: payload.Function,
		ArgTypes: protoArgTypes,
		Args:     protoArgs,
	}, nil
}

// ConvertTypeTagFromProto converts a proto TypeTag to Go types
func ConvertTypeTagFromProto(proto *TypeTag) (*typeaptos.TypeTag, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto type tag is nil")
	}

	var impl typeaptos.TypeTagImpl

	switch proto.Type {
	case TypeTagType_TYPE_TAG_BOOL:
		impl = typeaptos.BoolTag{}
	case TypeTagType_TYPE_TAG_U8:
		impl = typeaptos.U8Tag{}
	case TypeTagType_TYPE_TAG_U16:
		impl = typeaptos.U16Tag{}
	case TypeTagType_TYPE_TAG_U32:
		impl = typeaptos.U32Tag{}
	case TypeTagType_TYPE_TAG_U64:
		impl = typeaptos.U64Tag{}
	case TypeTagType_TYPE_TAG_U128:
		impl = typeaptos.U128Tag{}
	case TypeTagType_TYPE_TAG_U256:
		impl = typeaptos.U256Tag{}
	case TypeTagType_TYPE_TAG_ADDRESS:
		impl = typeaptos.AddressTag{}
	case TypeTagType_TYPE_TAG_SIGNER:
		impl = typeaptos.SignerTag{}
	case TypeTagType_TYPE_TAG_VECTOR:
		vectorValue := proto.GetVector()
		if vectorValue == nil {
			return nil, fmt.Errorf("vector type tag missing vector value")
		}
		elementType, err := ConvertTypeTagFromProto(vectorValue.ElementType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vector element type: %w", err)
		}
		impl = typeaptos.VectorTag{
			ElementType: *elementType,
		}
	case TypeTagType_TYPE_TAG_STRUCT:
		structValue := proto.GetStruct()
		if structValue == nil {
			return nil, fmt.Errorf("struct type tag missing struct value")
		}
		if len(structValue.Address) != typeaptos.AccountAddressLength {
			return nil, fmt.Errorf("invalid struct address length: expected %d, got %d", typeaptos.AccountAddressLength, len(structValue.Address))
		}
		var address typeaptos.AccountAddress
		copy(address[:], structValue.Address)

		typeParams := make([]typeaptos.TypeTag, len(structValue.TypeParams))
		for i, protoParam := range structValue.TypeParams {
			param, err := ConvertTypeTagFromProto(protoParam)
			if err != nil {
				return nil, fmt.Errorf("failed to convert struct type param %d: %w", i, err)
			}
			typeParams[i] = *param
		}
		impl = typeaptos.StructTag{
			Address:    address,
			Module:     structValue.Module,
			Name:       structValue.Name,
			TypeParams: typeParams,
		}
	case TypeTagType_TYPE_TAG_GENERIC:
		genericValue := proto.GetGeneric()
		if genericValue == nil {
			return nil, fmt.Errorf("generic type tag missing generic value")
		}
		impl = typeaptos.GenericTag{
			Index: uint16(genericValue.Index),
		}
	default:
		return nil, fmt.Errorf("unknown type tag type: %v", proto.Type)
	}

	return &typeaptos.TypeTag{
		Value: impl,
	}, nil
}

// ConvertTypeTagToProto converts a Go TypeTag to proto types
func ConvertTypeTagToProto(tag *typeaptos.TypeTag) (*TypeTag, error) {
	if tag == nil || tag.Value == nil {
		return nil, fmt.Errorf("type tag or value is nil")
	}

	protoTag := &TypeTag{
		Type: TypeTagType(tag.Value.TypeTagType()),
	}

	switch v := tag.Value.(type) {
	case typeaptos.VectorTag:
		elementType, err := ConvertTypeTagToProto(&v.ElementType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vector element type: %w", err)
		}
		protoTag.Value = &TypeTag_Vector{
			Vector: &VectorTag{
				ElementType: elementType,
			},
		}
	case typeaptos.StructTag:
		typeParams := make([]*TypeTag, len(v.TypeParams))
		for i, param := range v.TypeParams {
			protoParam, err := ConvertTypeTagToProto(&param)
			if err != nil {
				return nil, fmt.Errorf("failed to convert struct type param %d: %w", i, err)
			}
			typeParams[i] = protoParam
		}
		protoTag.Value = &TypeTag_Struct{
			Struct: &StructTag{
				Address:    v.Address[:],
				Module:     v.Module,
				Name:       v.Name,
				TypeParams: typeParams,
			},
		}
	case typeaptos.GenericTag:
		protoTag.Value = &TypeTag_Generic{
			Generic: &GenericTag{
				Index: uint32(v.Index),
			},
		}
	default:
		// For simple types (Bool, U8, U16, etc.), only the type field is needed
		// No value field is set for these
	}

	return protoTag, nil
}

// ConvertViewReplyFromProto converts proto reply to Go types
func ConvertViewReplyFromProto(protoReply *ViewReply) (*typeaptos.ViewReply, error) {
	if protoReply == nil {
		return nil, fmt.Errorf("proto reply is nil")
	}
	return &typeaptos.ViewReply{
		Data: protoReply.Data,
	}, nil
}

// ConvertViewReplyToProto converts Go reply to proto types
func ConvertViewReplyToProto(reply *typeaptos.ViewReply) (*ViewReply, error) {
	if reply == nil {
		return nil, fmt.Errorf("reply is nil")
	}
	return &ViewReply{
		Data: reply.Data,
	}, nil
}

// ========== TransactionByHash Conversion ==========

func ConvertTransactionByHashRequestToProto(req typeaptos.TransactionByHashRequest) *TransactionByHashRequest {
	return &TransactionByHashRequest{
		Hash: req.Hash,
	}
}

func ConvertTransactionByHashRequestFromProto(proto *TransactionByHashRequest) typeaptos.TransactionByHashRequest {
	return typeaptos.TransactionByHashRequest{
		Hash: proto.Hash,
	}
}

func ConvertTransactionByHashReplyToProto(reply *typeaptos.TransactionByHashReply) *TransactionByHashReply {
	if reply == nil {
		return nil
	}
	return &TransactionByHashReply{
		Transaction: ConvertTransactionToProto(reply.Transaction),
	}
}

func ConvertTransactionByHashReplyFromProto(proto *TransactionByHashReply) (*typeaptos.TransactionByHashReply, error) {
	if proto == nil {
		return nil, nil
	}

	tx, err := ConvertTransactionFromProto(proto.Transaction)
	if err != nil {
		return nil, err
	}

	return &typeaptos.TransactionByHashReply{
		Transaction: tx,
	}, nil
}

func ConvertTransactionToProto(tx *typeaptos.Transaction) *Transaction {
	if tx == nil {
		return nil
	}

	protoTx := &Transaction{
		Type: ConvertTransactionVariantToProto(tx.Type),
		Hash: tx.Hash,
		Data: tx.Data,
	}

	if tx.Version != nil {
		protoTx.Version = tx.Version
	}

	if tx.Success != nil {
		protoTx.Success = tx.Success
	}

	return protoTx
}

func ConvertTransactionFromProto(proto *Transaction) (*typeaptos.Transaction, error) {
	if proto == nil {
		return nil, nil
	}

	tx := &typeaptos.Transaction{
		Type: ConvertTransactionVariantFromProto(proto.Type),
		Hash: proto.Hash,
		Data: proto.Data,
	}

	if proto.Version != nil {
		tx.Version = proto.Version
	}

	if proto.Success != nil {
		tx.Success = proto.Success
	}

	return tx, nil
}

func ConvertTransactionVariantToProto(variant typeaptos.TransactionVariant) TransactionVariant {
	switch variant {
	case typeaptos.TransactionVariantPending:
		return TransactionVariant_TRANSACTION_VARIANT_PENDING
	case typeaptos.TransactionVariantUser:
		return TransactionVariant_TRANSACTION_VARIANT_USER
	case typeaptos.TransactionVariantGenesis:
		return TransactionVariant_TRANSACTION_VARIANT_GENESIS
	case typeaptos.TransactionVariantBlockMetadata:
		return TransactionVariant_TRANSACTION_VARIANT_BLOCK_METADATA
	case typeaptos.TransactionVariantBlockEpilogue:
		return TransactionVariant_TRANSACTION_VARIANT_BLOCK_EPILOGUE
	case typeaptos.TransactionVariantStateCheckpoint:
		return TransactionVariant_TRANSACTION_VARIANT_STATE_CHECKPOINT
	case typeaptos.TransactionVariantValidator:
		return TransactionVariant_TRANSACTION_VARIANT_VALIDATOR
	case typeaptos.TransactionVariantUnknown:
		return TransactionVariant_TRANSACTION_VARIANT_UNKNOWN
	default:
		return TransactionVariant_TRANSACTION_VARIANT_UNKNOWN
	}
}

func ConvertTransactionVariantFromProto(proto TransactionVariant) typeaptos.TransactionVariant {
	switch proto {
	case TransactionVariant_TRANSACTION_VARIANT_PENDING:
		return typeaptos.TransactionVariantPending
	case TransactionVariant_TRANSACTION_VARIANT_USER:
		return typeaptos.TransactionVariantUser
	case TransactionVariant_TRANSACTION_VARIANT_GENESIS:
		return typeaptos.TransactionVariantGenesis
	case TransactionVariant_TRANSACTION_VARIANT_BLOCK_METADATA:
		return typeaptos.TransactionVariantBlockMetadata
	case TransactionVariant_TRANSACTION_VARIANT_BLOCK_EPILOGUE:
		return typeaptos.TransactionVariantBlockEpilogue
	case TransactionVariant_TRANSACTION_VARIANT_STATE_CHECKPOINT:
		return typeaptos.TransactionVariantStateCheckpoint
	case TransactionVariant_TRANSACTION_VARIANT_VALIDATOR:
		return typeaptos.TransactionVariantValidator
	case TransactionVariant_TRANSACTION_VARIANT_UNKNOWN:
		return typeaptos.TransactionVariantUnknown
	default:
		return typeaptos.TransactionVariantUnknown
	}
}

// ========== AccountTransactions Conversion ==========

func ConvertAccountTransactionsRequestToProto(req typeaptos.AccountTransactionsRequest) *AccountTransactionsRequest {
	return &AccountTransactionsRequest{
		Address: req.Address[:],
		Start:   req.Start,
		Limit:   req.Limit,
	}
}

func ConvertAccountTransactionsRequestFromProto(proto *AccountTransactionsRequest) (typeaptos.AccountTransactionsRequest, error) {
	if proto == nil {
		return typeaptos.AccountTransactionsRequest{}, fmt.Errorf("proto request is nil")
	}
	if len(proto.Address) != typeaptos.AccountAddressLength {
		return typeaptos.AccountTransactionsRequest{}, fmt.Errorf("invalid address length: expected %d, got %d", typeaptos.AccountAddressLength, len(proto.Address))
	}
	var address typeaptos.AccountAddress
	copy(address[:], proto.Address)
	return typeaptos.AccountTransactionsRequest{
		Address: address,
		Start:   proto.Start,
		Limit:   proto.Limit,
	}, nil
}

func ConvertAccountTransactionsReplyToProto(reply *typeaptos.AccountTransactionsReply) *AccountTransactionsReply {
	if reply == nil {
		return nil
	}
	protoTxs := make([]*Transaction, len(reply.Transactions))
	for i, tx := range reply.Transactions {
		protoTxs[i] = ConvertTransactionToProto(tx)
	}
	return &AccountTransactionsReply{
		Transactions: protoTxs,
	}
}

func ConvertAccountTransactionsReplyFromProto(proto *AccountTransactionsReply) (*typeaptos.AccountTransactionsReply, error) {
	if proto == nil {
		return nil, nil
	}
	txs := make([]*typeaptos.Transaction, len(proto.Transactions))
	for i, protoTx := range proto.Transactions {
		tx, err := ConvertTransactionFromProto(protoTx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert transaction %d: %w", i, err)
		}
		txs[i] = tx
	}
	return &typeaptos.AccountTransactionsReply{
		Transactions: txs,
	}, nil
}

// ========== SubmitTransaction Conversion ==========

func ConvertSubmitTransactionRequestToProto(req typeaptos.SubmitTransactionRequest) (*SubmitTransactionRequest, error) {
	protoReq := &SubmitTransactionRequest{
		ReceiverModuleId: &ModuleID{
			Address: req.ReceiverModuleID.Address[:],
			Name:    req.ReceiverModuleID.Name,
		},
		EncodedPayload: req.EncodedPayload,
	}

	if req.GasConfig != nil {
		protoReq.GasConfig = &GasConfig{
			MaxGasAmount: req.GasConfig.MaxGasAmount,
			GasUnitPrice: req.GasConfig.GasUnitPrice,
		}
	}

	return protoReq, nil
}

func ConvertSubmitTransactionRequestFromProto(proto *SubmitTransactionRequest) (*typeaptos.SubmitTransactionRequest, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto request is nil")
	}

	if proto.ReceiverModuleId == nil {
		return nil, fmt.Errorf("receiver module id is nil")
	}

	if len(proto.ReceiverModuleId.Address) != typeaptos.AccountAddressLength {
		return nil, fmt.Errorf("invalid address length: expected %d, got %d", typeaptos.AccountAddressLength, len(proto.ReceiverModuleId.Address))
	}

	var address typeaptos.AccountAddress
	copy(address[:], proto.ReceiverModuleId.Address)

	req := &typeaptos.SubmitTransactionRequest{
		ReceiverModuleID: typeaptos.ModuleID{
			Address: address,
			Name:    proto.ReceiverModuleId.Name,
		},
		EncodedPayload: proto.EncodedPayload,
	}

	if proto.GasConfig != nil {
		req.GasConfig = &typeaptos.GasConfig{
			MaxGasAmount: proto.GasConfig.MaxGasAmount,
			GasUnitPrice: proto.GasConfig.GasUnitPrice,
		}
	}

	return req, nil
}

func ConvertSubmitTransactionReplyToProto(reply *typeaptos.SubmitTransactionReply) (*SubmitTransactionReply, error) {
	if reply == nil {
		return nil, fmt.Errorf("reply is nil")
	}

	return &SubmitTransactionReply{
		TxStatus:         TxStatus(reply.TxStatus),
		TxHash:           reply.TxHash,
		TxIdempotencyKey: reply.TxIdempotencyKey,
	}, nil
}

func ConvertSubmitTransactionReplyFromProto(proto *SubmitTransactionReply) (*typeaptos.SubmitTransactionReply, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto reply is nil")
	}

	return &typeaptos.SubmitTransactionReply{
		TxStatus:         typeaptos.TransactionStatus(proto.TxStatus),
		TxHash:           proto.TxHash,
		TxIdempotencyKey: proto.TxIdempotencyKey,
	}, nil
}
