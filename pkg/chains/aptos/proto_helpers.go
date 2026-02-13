package aptos

import (
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

// ConvertViewPayloadFromProto converts a proto ViewPayload to Go types
func ConvertViewPayloadFromProto(proto *ViewPayload) (*aptos.ViewPayload, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto payload is nil")
	}

	if len(proto.Module.Address) != aptos.AccountAddressLength {
		return nil, fmt.Errorf("invalid address length: expected %d, got %d", aptos.AccountAddressLength, len(proto.Module.Address))
	}

	var address aptos.AccountAddress
	copy(address[:], proto.Module.Address)

	module := aptos.ModuleID{
		Address: address,
		Name:    proto.Module.Name,
	}

	argTypes := make([]aptos.TypeTag, len(proto.ArgTypes))
	for i, protoTypeTag := range proto.ArgTypes {
		typeTag, err := ConvertTypeTagFromProto(protoTypeTag)
		if err != nil {
			return nil, fmt.Errorf("failed to convert arg type %d: %w", i, err)
		}
		argTypes[i] = *typeTag
	}

	args := make([][]byte, len(proto.Args))
	for i, protoArg := range proto.Args {
		// Extract the BCS encoded bytes from Any
		args[i] = protoArg.Value
	}

	return &aptos.ViewPayload{
		Module:   module,
		Function: proto.Function,
		ArgTypes: argTypes,
		Args:     args,
	}, nil
}

// ConvertViewPayloadToProto converts a Go ViewPayload to proto types
func ConvertViewPayloadToProto(payload *aptos.ViewPayload) (*ViewPayload, error) {
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

	protoArgs := make([]*anypb.Any, len(payload.Args))
	for i, arg := range payload.Args {
		// Args are already BCS encoded bytes, wrap them in Any
		anyArg := &anypb.Any{
			Value: arg,
		}
		protoArgs[i] = anyArg
	}

	return &ViewPayload{
		Module:   protoModule,
		Function: payload.Function,
		ArgTypes: protoArgTypes,
		Args:     protoArgs,
	}, nil
}

// ConvertTypeTagFromProto converts a proto TypeTag to Go types
func ConvertTypeTagFromProto(proto *TypeTag) (*aptos.TypeTag, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto type tag is nil")
	}

	var impl aptos.TypeTagImpl

	switch proto.Type {
	case TypeTagType_TYPE_TAG_BOOL:
		impl = aptos.BoolTag{}
	case TypeTagType_TYPE_TAG_U8:
		impl = aptos.U8Tag{}
	case TypeTagType_TYPE_TAG_U16:
		impl = aptos.U16Tag{}
	case TypeTagType_TYPE_TAG_U32:
		impl = aptos.U32Tag{}
	case TypeTagType_TYPE_TAG_U64:
		impl = aptos.U64Tag{}
	case TypeTagType_TYPE_TAG_U128:
		impl = aptos.U128Tag{}
	case TypeTagType_TYPE_TAG_U256:
		impl = aptos.U256Tag{}
	case TypeTagType_TYPE_TAG_ADDRESS:
		impl = aptos.AddressTag{}
	case TypeTagType_TYPE_TAG_SIGNER:
		impl = aptos.SignerTag{}
	case TypeTagType_TYPE_TAG_VECTOR:
		vectorValue := proto.GetVector()
		if vectorValue == nil {
			return nil, fmt.Errorf("vector type tag missing vector value")
		}
		elementType, err := ConvertTypeTagFromProto(vectorValue.ElementType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vector element type: %w", err)
		}
		impl = aptos.VectorTag{
			ElementType: *elementType,
		}
	case TypeTagType_TYPE_TAG_STRUCT:
		structValue := proto.GetStruct()
		if structValue == nil {
			return nil, fmt.Errorf("struct type tag missing struct value")
		}
		if len(structValue.Address) != aptos.AccountAddressLength {
			return nil, fmt.Errorf("invalid struct address length: expected %d, got %d", aptos.AccountAddressLength, len(structValue.Address))
		}
		var address aptos.AccountAddress
		copy(address[:], structValue.Address)

		typeParams := make([]aptos.TypeTag, len(structValue.TypeParams))
		for i, protoParam := range structValue.TypeParams {
			param, err := ConvertTypeTagFromProto(protoParam)
			if err != nil {
				return nil, fmt.Errorf("failed to convert struct type param %d: %w", i, err)
			}
			typeParams[i] = *param
		}
		impl = aptos.StructTag{
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
		impl = aptos.GenericTag{
			Index: uint16(genericValue.Index),
		}
	default:
		return nil, fmt.Errorf("unknown type tag type: %v", proto.Type)
	}

	return &aptos.TypeTag{
		Value: impl,
	}, nil
}

// ConvertTypeTagToProto converts a Go TypeTag to proto types
func ConvertTypeTagToProto(tag *aptos.TypeTag) (*TypeTag, error) {
	if tag == nil || tag.Value == nil {
		return nil, fmt.Errorf("type tag or value is nil")
	}

	protoTag := &TypeTag{
		Type: TypeTagType(tag.Value.TypeTagType()),
	}

	switch v := tag.Value.(type) {
	case aptos.VectorTag:
		elementType, err := ConvertTypeTagToProto(&v.ElementType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vector element type: %w", err)
		}
		protoTag.Value = &TypeTag_Vector{
			Vector: &VectorTag{
				ElementType: elementType,
			},
		}
	case aptos.StructTag:
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
	case aptos.GenericTag:
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

// ConvertViewResultFromProto converts proto result to Go types
func ConvertViewResultFromProto(protoResult []*anypb.Any) ([]*aptos.ViewResultValue, error) {
	result := make([]*aptos.ViewResultValue, len(protoResult))
	for i, protoVal := range protoResult {
		// The Any.Value contains the raw bytes, which could be BCS or JSON
		// For now, we store it as raw bytes and let the caller decide how to decode
		result[i] = &aptos.ViewResultValue{
			AsDecodedBinary: protoVal.Value,
		}
	}
	return result, nil
}

// ConvertViewResultToProto converts Go result to proto types
func ConvertViewResultToProto(result []*aptos.ViewResultValue) ([]*anypb.Any, error) {
	protoResult := make([]*anypb.Any, len(result))
	for i, val := range result {
		if val == nil {
			return nil, fmt.Errorf("view result value %d is nil", i)
		}
		// Prefer JSON if available, otherwise use raw bytes
		var data []byte
		if len(val.AsJSON) > 0 {
			data = val.AsJSON
		} else {
			data = val.AsDecodedBinary
		}
		protoResult[i] = &anypb.Any{
			Value: data,
		}
	}
	return protoResult, nil
}

// ConvertEventsByHandleRequestToProto converts Go request to proto
func ConvertEventsByHandleRequestToProto(req *aptos.EventsByHandleRequest) (*EventsByHandleRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	protoReq := &EventsByHandleRequest{
		Account:     req.Account[:],
		EventHandle: req.EventHandle,
		FieldName:   req.FieldName,
	}
	if req.Start != nil {
		protoReq.Start = req.Start
	}
	if req.Limit != nil {
		protoReq.Limit = req.Limit
	}

	return protoReq, nil
}

// ConvertEventsByHandleRequestFromProto converts proto request to Go
func ConvertEventsByHandleRequestFromProto(proto *EventsByHandleRequest) (*aptos.EventsByHandleRequest, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto request is nil")
	}

	if len(proto.Account) != aptos.AccountAddressLength {
		return nil, fmt.Errorf("invalid account address length: expected %d, got %d", aptos.AccountAddressLength, len(proto.Account))
	}

	var account aptos.AccountAddress
	copy(account[:], proto.Account)

	req := &aptos.EventsByHandleRequest{
		Account:     account,
		EventHandle: proto.EventHandle,
		FieldName:   proto.FieldName,
	}
	if proto.Start != nil {
		req.Start = proto.Start
	}
	if proto.Limit != nil {
		req.Limit = proto.Limit
	}

	return req, nil
}

// ConvertEventsByHandleReplyToProto converts Go reply to proto
func ConvertEventsByHandleReplyToProto(reply *aptos.EventsByHandleReply) (*EventsByHandleReply, error) {
	if reply == nil {
		return nil, fmt.Errorf("reply is nil")
	}

	protoEvents := make([]*Event, len(reply.Events))
	for i, event := range reply.Events {
		protoEvent, err := ConvertEventToProto(event)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event %d: %w", i, err)
		}
		protoEvents[i] = protoEvent
	}

	return &EventsByHandleReply{
		Events: protoEvents,
	}, nil
}

// ConvertEventsByHandleReplyFromProto converts proto reply to Go
func ConvertEventsByHandleReplyFromProto(proto *EventsByHandleReply) (*aptos.EventsByHandleReply, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto reply is nil")
	}

	events := make([]*aptos.Event, len(proto.Events))
	for i, protoEvent := range proto.Events {
		event, err := ConvertEventFromProto(protoEvent)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event %d: %w", i, err)
		}
		events[i] = event
	}

	return &aptos.EventsByHandleReply{
		Events: events,
	}, nil
}

// ConvertEventToProto converts Go Event to proto
func ConvertEventToProto(event *aptos.Event) (*Event, error) {
	if event == nil {
		return nil, fmt.Errorf("event is nil")
	}

	protoEvent := &Event{
		Version:        event.Version,
		Type:           event.Type,
		SequenceNumber: event.SequenceNumber,
		Data:           event.Data,
	}

	if event.Guid != nil {
		protoGuid, err := ConvertGUIDToProto(event.Guid)
		if err != nil {
			return nil, fmt.Errorf("failed to convert GUID: %w", err)
		}
		protoEvent.Guid = protoGuid
	}

	return protoEvent, nil
}

// ConvertEventFromProto converts proto Event to Go
func ConvertEventFromProto(proto *Event) (*aptos.Event, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto event is nil")
	}

	event := &aptos.Event{
		Version:        proto.Version,
		Type:           proto.Type,
		SequenceNumber: proto.SequenceNumber,
		Data:           proto.Data,
	}

	if proto.Guid != nil {
		guid, err := ConvertGUIDFromProto(proto.Guid)
		if err != nil {
			return nil, fmt.Errorf("failed to convert GUID: %w", err)
		}
		event.Guid = guid
	}

	return event, nil
}

// ConvertGUIDToProto converts Go GUID to proto
func ConvertGUIDToProto(guid *aptos.GUID) (*GUID, error) {
	if guid == nil {
		return nil, fmt.Errorf("guid is nil")
	}

	return &GUID{
		CreationNumber: guid.CreationNumber,
		AccountAddress: guid.AccountAddress[:],
	}, nil
}

// ConvertGUIDFromProto converts proto GUID to Go
func ConvertGUIDFromProto(proto *GUID) (*aptos.GUID, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto guid is nil")
	}

	if len(proto.AccountAddress) != aptos.AccountAddressLength {
		return nil, fmt.Errorf("invalid account address length: expected %d, got %d", aptos.AccountAddressLength, len(proto.AccountAddress))
	}

	var address aptos.AccountAddress
	copy(address[:], proto.AccountAddress)

	return &aptos.GUID{
		CreationNumber: proto.CreationNumber,
		AccountAddress: address,
	}, nil
}

// ========== TransactionByHash Conversion ==========

func ConvertTransactionByHashRequestToProto(req aptos.TransactionByHashRequest) *TransactionByHashRequest {
	return &TransactionByHashRequest{
		Hash: req.Hash,
	}
}

func ConvertTransactionByHashRequestFromProto(proto *TransactionByHashRequest) aptos.TransactionByHashRequest {
	return aptos.TransactionByHashRequest{
		Hash: proto.Hash,
	}
}

func ConvertTransactionByHashReplyToProto(reply *aptos.TransactionByHashReply) *TransactionByHashReply {
	if reply == nil {
		return nil
	}
	return &TransactionByHashReply{
		Transaction: ConvertTransactionToProto(reply.Transaction),
	}
}

func ConvertTransactionByHashReplyFromProto(proto *TransactionByHashReply) (*aptos.TransactionByHashReply, error) {
	if proto == nil {
		return nil, nil
	}

	tx, err := ConvertTransactionFromProto(proto.Transaction)
	if err != nil {
		return nil, err
	}

	return &aptos.TransactionByHashReply{
		Transaction: tx,
	}, nil
}

func ConvertTransactionToProto(tx *aptos.Transaction) *Transaction {
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

func ConvertTransactionFromProto(proto *Transaction) (*aptos.Transaction, error) {
	if proto == nil {
		return nil, nil
	}

	tx := &aptos.Transaction{
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

func ConvertTransactionVariantToProto(variant aptos.TransactionVariant) TransactionVariant {
	switch variant {
	case aptos.TransactionVariantPending:
		return TransactionVariant_TRANSACTION_VARIANT_PENDING
	case aptos.TransactionVariantUser:
		return TransactionVariant_TRANSACTION_VARIANT_USER
	case aptos.TransactionVariantGenesis:
		return TransactionVariant_TRANSACTION_VARIANT_GENESIS
	case aptos.TransactionVariantBlockMetadata:
		return TransactionVariant_TRANSACTION_VARIANT_BLOCK_METADATA
	case aptos.TransactionVariantBlockEpilogue:
		return TransactionVariant_TRANSACTION_VARIANT_BLOCK_EPILOGUE
	case aptos.TransactionVariantStateCheckpoint:
		return TransactionVariant_TRANSACTION_VARIANT_STATE_CHECKPOINT
	case aptos.TransactionVariantValidator:
		return TransactionVariant_TRANSACTION_VARIANT_VALIDATOR
	case aptos.TransactionVariantUnknown:
		return TransactionVariant_TRANSACTION_VARIANT_UNKNOWN
	default:
		return TransactionVariant_TRANSACTION_VARIANT_UNKNOWN
	}
}

func ConvertTransactionVariantFromProto(proto TransactionVariant) aptos.TransactionVariant {
	switch proto {
	case TransactionVariant_TRANSACTION_VARIANT_PENDING:
		return aptos.TransactionVariantPending
	case TransactionVariant_TRANSACTION_VARIANT_USER:
		return aptos.TransactionVariantUser
	case TransactionVariant_TRANSACTION_VARIANT_GENESIS:
		return aptos.TransactionVariantGenesis
	case TransactionVariant_TRANSACTION_VARIANT_BLOCK_METADATA:
		return aptos.TransactionVariantBlockMetadata
	case TransactionVariant_TRANSACTION_VARIANT_BLOCK_EPILOGUE:
		return aptos.TransactionVariantBlockEpilogue
	case TransactionVariant_TRANSACTION_VARIANT_STATE_CHECKPOINT:
		return aptos.TransactionVariantStateCheckpoint
	case TransactionVariant_TRANSACTION_VARIANT_VALIDATOR:
		return aptos.TransactionVariantValidator
	case TransactionVariant_TRANSACTION_VARIANT_UNKNOWN:
		return aptos.TransactionVariantUnknown
	default:
		return aptos.TransactionVariantUnknown
	}
}

// ========== SubmitTransaction Conversion ==========

func ConvertSubmitTransactionRequestToProto(req aptos.SubmitTransactionRequest) (*SubmitTransactionRequest, error) {
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

func ConvertSubmitTransactionRequestFromProto(proto *SubmitTransactionRequest) (*aptos.SubmitTransactionRequest, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto request is nil")
	}

	if proto.ReceiverModuleId == nil {
		return nil, fmt.Errorf("receiver module id is nil")
	}

	if len(proto.ReceiverModuleId.Address) != aptos.AccountAddressLength {
		return nil, fmt.Errorf("invalid address length: expected %d, got %d", aptos.AccountAddressLength, len(proto.ReceiverModuleId.Address))
	}

	var address aptos.AccountAddress
	copy(address[:], proto.ReceiverModuleId.Address)

	req := &aptos.SubmitTransactionRequest{
		ReceiverModuleID: aptos.ModuleID{
			Address: address,
			Name:    proto.ReceiverModuleId.Name,
		},
		EncodedPayload: proto.EncodedPayload,
	}

	if proto.GasConfig != nil {
		req.GasConfig = &aptos.GasConfig{
			MaxGasAmount: proto.GasConfig.MaxGasAmount,
			GasUnitPrice: proto.GasConfig.GasUnitPrice,
		}
	}

	return req, nil
}

func ConvertSubmitTransactionReplyToProto(reply *aptos.SubmitTransactionReply) (*SubmitTransactionReply, error) {
	if reply == nil {
		return nil, fmt.Errorf("reply is nil")
	}

	protoReply := &SubmitTransactionReply{}

	if reply.PendingTransaction != nil {
		protoPending, err := ConvertPendingTransactionToProto(reply.PendingTransaction)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pending transaction: %w", err)
		}
		protoReply.PendingTransaction = protoPending
	}

	return protoReply, nil
}

func ConvertSubmitTransactionReplyFromProto(proto *SubmitTransactionReply) (*aptos.SubmitTransactionReply, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto reply is nil")
	}

	reply := &aptos.SubmitTransactionReply{}

	if proto.PendingTransaction != nil {
		pending, err := ConvertPendingTransactionFromProto(proto.PendingTransaction)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pending transaction: %w", err)
		}
		reply.PendingTransaction = pending
	}

	return reply, nil
}

func ConvertPendingTransactionToProto(tx *aptos.PendingTransaction) (*PendingTransaction, error) {
	if tx == nil {
		return nil, fmt.Errorf("pending transaction is nil")
	}

	protoTx := &PendingTransaction{
		Hash:                    tx.Hash,
		Sender:                  tx.Sender[:],
		SequenceNumber:          tx.SequenceNumber,
		MaxGasAmount:            tx.MaxGasAmount,
		GasUnitPrice:            tx.GasUnitPrice,
		ExpirationTimestampSecs: tx.ExpirationTimestampSecs,
		Payload:                 tx.Payload,
		Signature:               tx.Signature,
	}

	if tx.ReplayProtectionNonce != nil {
		protoTx.ReplayProtectionNonce = tx.ReplayProtectionNonce
	}

	return protoTx, nil
}

func ConvertPendingTransactionFromProto(proto *PendingTransaction) (*aptos.PendingTransaction, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto pending transaction is nil")
	}

	if len(proto.Sender) != aptos.AccountAddressLength {
		return nil, fmt.Errorf("invalid sender address length: expected %d, got %d", aptos.AccountAddressLength, len(proto.Sender))
	}

	var sender aptos.AccountAddress
	copy(sender[:], proto.Sender)

	tx := &aptos.PendingTransaction{
		Hash:                    proto.Hash,
		Sender:                  sender,
		SequenceNumber:          proto.SequenceNumber,
		MaxGasAmount:            proto.MaxGasAmount,
		GasUnitPrice:            proto.GasUnitPrice,
		ExpirationTimestampSecs: proto.ExpirationTimestampSecs,
		Payload:                 proto.Payload,
		Signature:               proto.Signature,
	}

	if proto.ReplayProtectionNonce != nil {
		tx.ReplayProtectionNonce = proto.ReplayProtectionNonce
	}

	return tx, nil
}
