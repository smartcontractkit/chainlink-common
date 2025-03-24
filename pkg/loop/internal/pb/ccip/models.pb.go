// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: models.proto

package ccippb

import (
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type FinalityStatus int32

const (
	FinalityStatus_Unknown      FinalityStatus = 0
	FinalityStatus_Finalized    FinalityStatus = 1
	FinalityStatus_NotFinalized FinalityStatus = 2
)

// Enum value maps for FinalityStatus.
var (
	FinalityStatus_name = map[int32]string{
		0: "Unknown",
		1: "Finalized",
		2: "NotFinalized",
	}
	FinalityStatus_value = map[string]int32{
		"Unknown":      0,
		"Finalized":    1,
		"NotFinalized": 2,
	}
)

func (x FinalityStatus) Enum() *FinalityStatus {
	p := new(FinalityStatus)
	*p = x
	return p
}

func (x FinalityStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FinalityStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_models_proto_enumTypes[0].Descriptor()
}

func (FinalityStatus) Type() protoreflect.EnumType {
	return &file_models_proto_enumTypes[0]
}

func (x FinalityStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FinalityStatus.Descriptor instead.
func (FinalityStatus) EnumDescriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{0}
}

// TxMeta is a message that contains the metadata of a transaction. It is a gRPC adapter to
// [github.com/smartcontractkit/chainlink-common/pkg/types/ccip.TxMeta]
type TxMeta struct {
	state                   protoimpl.MessageState `protogen:"open.v1"`
	BlockTimestampUnixMilli int64                  `protobuf:"varint,1,opt,name=block_timestamp_unix_milli,json=blockTimestampUnixMilli,proto3" json:"block_timestamp_unix_milli,omitempty"`
	BlockNumber             uint64                 `protobuf:"varint,2,opt,name=block_number,json=blockNumber,proto3" json:"block_number,omitempty"`
	TxHash                  string                 `protobuf:"bytes,3,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
	LogIndex                uint64                 `protobuf:"varint,4,opt,name=log_index,json=logIndex,proto3" json:"log_index,omitempty"`
	Finalized               FinalityStatus         `protobuf:"varint,5,opt,name=finalized,proto3,enum=loop.internal.pb.ccip.FinalityStatus" json:"finalized,omitempty"`
	unknownFields           protoimpl.UnknownFields
	sizeCache               protoimpl.SizeCache
}

func (x *TxMeta) Reset() {
	*x = TxMeta{}
	mi := &file_models_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TxMeta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TxMeta) ProtoMessage() {}

func (x *TxMeta) ProtoReflect() protoreflect.Message {
	mi := &file_models_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TxMeta.ProtoReflect.Descriptor instead.
func (*TxMeta) Descriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{0}
}

func (x *TxMeta) GetBlockTimestampUnixMilli() int64 {
	if x != nil {
		return x.BlockTimestampUnixMilli
	}
	return 0
}

func (x *TxMeta) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *TxMeta) GetTxHash() string {
	if x != nil {
		return x.TxHash
	}
	return ""
}

func (x *TxMeta) GetLogIndex() uint64 {
	if x != nil {
		return x.LogIndex
	}
	return 0
}

func (x *TxMeta) GetFinalized() FinalityStatus {
	if x != nil {
		return x.Finalized
	}
	return FinalityStatus_Unknown
}

// EVM2EVMMesssage is a gRPC adapter to [github.com/smartcontractkit/chainlink-common/pkg/types/ccip.EVM2EVMMessage]
type EVM2EVMMessage struct {
	state               protoimpl.MessageState `protogen:"open.v1"`
	SequenceNumber      uint64                 `protobuf:"varint,1,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number,omitempty"`
	GasLimit            *pb.BigInt             `protobuf:"bytes,2,opt,name=gas_limit,json=gasLimit,proto3" json:"gas_limit,omitempty"`
	Nonce               uint64                 `protobuf:"varint,3,opt,name=nonce,proto3" json:"nonce,omitempty"`
	GasPrice            uint64                 `protobuf:"varint,4,opt,name=gas_price,json=gasPrice,proto3" json:"gas_price,omitempty"`
	MessageId           []byte                 `protobuf:"bytes,5,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"` // Hash [32]byte
	SourceChainSelector uint64                 `protobuf:"varint,6,opt,name=source_chain_selector,json=sourceChainSelector,proto3" json:"source_chain_selector,omitempty"`
	Sender              string                 `protobuf:"bytes,7,opt,name=sender,proto3" json:"sender,omitempty"`     // Address
	Receiver            string                 `protobuf:"bytes,8,opt,name=receiver,proto3" json:"receiver,omitempty"` // Address
	Strict              bool                   `protobuf:"varint,9,opt,name=strict,proto3" json:"strict,omitempty"`
	FeeToken            string                 `protobuf:"bytes,10,opt,name=fee_token,json=feeToken,proto3" json:"fee_token,omitempty"` // Address
	FeeTokenAmount      *pb.BigInt             `protobuf:"bytes,11,opt,name=fee_token_amount,json=feeTokenAmount,proto3" json:"fee_token_amount,omitempty"`
	Data                []byte                 `protobuf:"bytes,12,opt,name=data,proto3" json:"data,omitempty"`
	TokenAmounts        []*TokenAmount         `protobuf:"bytes,13,rep,name=token_amounts,json=tokenAmounts,proto3" json:"token_amounts,omitempty"`
	SourceTokenData     [][]byte               `protobuf:"bytes,14,rep,name=source_token_data,json=sourceTokenData,proto3" json:"source_token_data,omitempty"` // Note: we don't bother with the Hash field here in the gRPC because it's derived from the golang struct
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *EVM2EVMMessage) Reset() {
	*x = EVM2EVMMessage{}
	mi := &file_models_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EVM2EVMMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVM2EVMMessage) ProtoMessage() {}

func (x *EVM2EVMMessage) ProtoReflect() protoreflect.Message {
	mi := &file_models_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EVM2EVMMessage.ProtoReflect.Descriptor instead.
func (*EVM2EVMMessage) Descriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{1}
}

func (x *EVM2EVMMessage) GetSequenceNumber() uint64 {
	if x != nil {
		return x.SequenceNumber
	}
	return 0
}

func (x *EVM2EVMMessage) GetGasLimit() *pb.BigInt {
	if x != nil {
		return x.GasLimit
	}
	return nil
}

func (x *EVM2EVMMessage) GetNonce() uint64 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

func (x *EVM2EVMMessage) GetGasPrice() uint64 {
	if x != nil {
		return x.GasPrice
	}
	return 0
}

func (x *EVM2EVMMessage) GetMessageId() []byte {
	if x != nil {
		return x.MessageId
	}
	return nil
}

func (x *EVM2EVMMessage) GetSourceChainSelector() uint64 {
	if x != nil {
		return x.SourceChainSelector
	}
	return 0
}

func (x *EVM2EVMMessage) GetSender() string {
	if x != nil {
		return x.Sender
	}
	return ""
}

func (x *EVM2EVMMessage) GetReceiver() string {
	if x != nil {
		return x.Receiver
	}
	return ""
}

func (x *EVM2EVMMessage) GetStrict() bool {
	if x != nil {
		return x.Strict
	}
	return false
}

func (x *EVM2EVMMessage) GetFeeToken() string {
	if x != nil {
		return x.FeeToken
	}
	return ""
}

func (x *EVM2EVMMessage) GetFeeTokenAmount() *pb.BigInt {
	if x != nil {
		return x.FeeTokenAmount
	}
	return nil
}

func (x *EVM2EVMMessage) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *EVM2EVMMessage) GetTokenAmounts() []*TokenAmount {
	if x != nil {
		return x.TokenAmounts
	}
	return nil
}

func (x *EVM2EVMMessage) GetSourceTokenData() [][]byte {
	if x != nil {
		return x.SourceTokenData
	}
	return nil
}

// EVM2EVMOnRampCCIPSendRequestedWithMeta is a gRPC adapter to [github.com/smartcontractkit/chainlink-common/pkg/types/ccip.EVM2EVMOnRampCCIPSendRequestedWithMeta]
type EVM2EVMOnRampCCIPSendRequestedWithMeta struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	EvmToEvmMsg    *EVM2EVMMessage        `protobuf:"bytes,1,opt,name=evm_to_evm_msg,json=evmToEvmMsg,proto3" json:"evm_to_evm_msg,omitempty"`
	BlockTimestamp *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=block_timestamp,json=blockTimestamp,proto3" json:"block_timestamp,omitempty"`
	Executed       bool                   `protobuf:"varint,3,opt,name=executed,proto3" json:"executed,omitempty"`
	Finalized      bool                   `protobuf:"varint,4,opt,name=finalized,proto3" json:"finalized,omitempty"`
	LogIndex       uint64                 `protobuf:"varint,5,opt,name=log_index,json=logIndex,proto3" json:"log_index,omitempty"`
	TxHash         string                 `protobuf:"bytes,6,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *EVM2EVMOnRampCCIPSendRequestedWithMeta) Reset() {
	*x = EVM2EVMOnRampCCIPSendRequestedWithMeta{}
	mi := &file_models_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EVM2EVMOnRampCCIPSendRequestedWithMeta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVM2EVMOnRampCCIPSendRequestedWithMeta) ProtoMessage() {}

func (x *EVM2EVMOnRampCCIPSendRequestedWithMeta) ProtoReflect() protoreflect.Message {
	mi := &file_models_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EVM2EVMOnRampCCIPSendRequestedWithMeta.ProtoReflect.Descriptor instead.
func (*EVM2EVMOnRampCCIPSendRequestedWithMeta) Descriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{2}
}

func (x *EVM2EVMOnRampCCIPSendRequestedWithMeta) GetEvmToEvmMsg() *EVM2EVMMessage {
	if x != nil {
		return x.EvmToEvmMsg
	}
	return nil
}

func (x *EVM2EVMOnRampCCIPSendRequestedWithMeta) GetBlockTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.BlockTimestamp
	}
	return nil
}

func (x *EVM2EVMOnRampCCIPSendRequestedWithMeta) GetExecuted() bool {
	if x != nil {
		return x.Executed
	}
	return false
}

func (x *EVM2EVMOnRampCCIPSendRequestedWithMeta) GetFinalized() bool {
	if x != nil {
		return x.Finalized
	}
	return false
}

func (x *EVM2EVMOnRampCCIPSendRequestedWithMeta) GetLogIndex() uint64 {
	if x != nil {
		return x.LogIndex
	}
	return 0
}

func (x *EVM2EVMOnRampCCIPSendRequestedWithMeta) GetTxHash() string {
	if x != nil {
		return x.TxHash
	}
	return ""
}

// TokenPoolRateLimit is a gRPC adapter for the struct
// [github.com/smartcontractkit/chainlink-common/pkg/types/ccip/TokenPoolRateLimit]
type TokenPoolRateLimit struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tokens        *pb.BigInt             `protobuf:"bytes,1,opt,name=tokens,proto3" json:"tokens,omitempty"`
	LastUpdated   uint32                 `protobuf:"varint,2,opt,name=last_updated,json=lastUpdated,proto3" json:"last_updated,omitempty"`
	IsEnabled     bool                   `protobuf:"varint,3,opt,name=is_enabled,json=isEnabled,proto3" json:"is_enabled,omitempty"`
	Capacity      *pb.BigInt             `protobuf:"bytes,4,opt,name=capacity,proto3" json:"capacity,omitempty"`
	Rate          *pb.BigInt             `protobuf:"bytes,5,opt,name=rate,proto3" json:"rate,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TokenPoolRateLimit) Reset() {
	*x = TokenPoolRateLimit{}
	mi := &file_models_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TokenPoolRateLimit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenPoolRateLimit) ProtoMessage() {}

func (x *TokenPoolRateLimit) ProtoReflect() protoreflect.Message {
	mi := &file_models_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenPoolRateLimit.ProtoReflect.Descriptor instead.
func (*TokenPoolRateLimit) Descriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{3}
}

func (x *TokenPoolRateLimit) GetTokens() *pb.BigInt {
	if x != nil {
		return x.Tokens
	}
	return nil
}

func (x *TokenPoolRateLimit) GetLastUpdated() uint32 {
	if x != nil {
		return x.LastUpdated
	}
	return 0
}

func (x *TokenPoolRateLimit) GetIsEnabled() bool {
	if x != nil {
		return x.IsEnabled
	}
	return false
}

func (x *TokenPoolRateLimit) GetCapacity() *pb.BigInt {
	if x != nil {
		return x.Capacity
	}
	return nil
}

func (x *TokenPoolRateLimit) GetRate() *pb.BigInt {
	if x != nil {
		return x.Rate
	}
	return nil
}

// TokenAmount is a gRPC adapter to [github.com/smartcontractkit/chainlink-common/pkg/types/ccip.TokenAmount]
type TokenAmount struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Token         string                 `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"` // Address
	Amount        *pb.BigInt             `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TokenAmount) Reset() {
	*x = TokenAmount{}
	mi := &file_models_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TokenAmount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenAmount) ProtoMessage() {}

func (x *TokenAmount) ProtoReflect() protoreflect.Message {
	mi := &file_models_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenAmount.ProtoReflect.Descriptor instead.
func (*TokenAmount) Descriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{4}
}

func (x *TokenAmount) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *TokenAmount) GetAmount() *pb.BigInt {
	if x != nil {
		return x.Amount
	}
	return nil
}

// ExecutionReport is a gRPC adapter to [github.com/smartcontractkit/chainlink-common/pkg/types/ccip.ExecutionReport]
type ExecutionReport struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	EvmToEvmMessages  []*EVM2EVMMessage      `protobuf:"bytes,1,rep,name=evm_to_evm_messages,json=evmToEvmMessages,proto3" json:"evm_to_evm_messages,omitempty"`
	OffchainTokenData []*TokenData           `protobuf:"bytes,2,rep,name=offchain_token_data,json=offchainTokenData,proto3" json:"offchain_token_data,omitempty"`
	Proofs            [][]byte               `protobuf:"bytes,3,rep,name=proofs,proto3" json:"proofs,omitempty"` // [][32]byte
	ProofFlagBits     *pb.BigInt             `protobuf:"bytes,4,opt,name=proof_flag_bits,json=proofFlagBits,proto3" json:"proof_flag_bits,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *ExecutionReport) Reset() {
	*x = ExecutionReport{}
	mi := &file_models_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExecutionReport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecutionReport) ProtoMessage() {}

func (x *ExecutionReport) ProtoReflect() protoreflect.Message {
	mi := &file_models_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecutionReport.ProtoReflect.Descriptor instead.
func (*ExecutionReport) Descriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{5}
}

func (x *ExecutionReport) GetEvmToEvmMessages() []*EVM2EVMMessage {
	if x != nil {
		return x.EvmToEvmMessages
	}
	return nil
}

func (x *ExecutionReport) GetOffchainTokenData() []*TokenData {
	if x != nil {
		return x.OffchainTokenData
	}
	return nil
}

func (x *ExecutionReport) GetProofs() [][]byte {
	if x != nil {
		return x.Proofs
	}
	return nil
}

func (x *ExecutionReport) GetProofFlagBits() *pb.BigInt {
	if x != nil {
		return x.ProofFlagBits
	}
	return nil
}

type TokenData struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Data          [][]byte               `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TokenData) Reset() {
	*x = TokenData{}
	mi := &file_models_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TokenData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenData) ProtoMessage() {}

func (x *TokenData) ProtoReflect() protoreflect.Message {
	mi := &file_models_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenData.ProtoReflect.Descriptor instead.
func (*TokenData) Descriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{6}
}

func (x *TokenData) GetData() [][]byte {
	if x != nil {
		return x.Data
	}
	return nil
}

// TokenPrice is the price of the stated token. It is a gRPC adapter to [github.com/smartcontractkit/chainlink-common/pkg/types/ccip.TokenPrice]
type TokenPrice struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Token         string                 `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"` // Address
	Value         *pb.BigInt             `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TokenPrice) Reset() {
	*x = TokenPrice{}
	mi := &file_models_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TokenPrice) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenPrice) ProtoMessage() {}

func (x *TokenPrice) ProtoReflect() protoreflect.Message {
	mi := &file_models_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenPrice.ProtoReflect.Descriptor instead.
func (*TokenPrice) Descriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{7}
}

func (x *TokenPrice) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *TokenPrice) GetValue() *pb.BigInt {
	if x != nil {
		return x.Value
	}
	return nil
}

// GasPrice is a gRPC adapter to [github.com/smartcontractkit/chainlink-common/pkg/types/ccip.GasPrice]
type GasPrice struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	DestChainSelector uint64                 `protobuf:"varint,1,opt,name=dest_chain_selector,json=destChainSelector,proto3" json:"dest_chain_selector,omitempty"`
	Value             *pb.BigInt             `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *GasPrice) Reset() {
	*x = GasPrice{}
	mi := &file_models_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GasPrice) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GasPrice) ProtoMessage() {}

func (x *GasPrice) ProtoReflect() protoreflect.Message {
	mi := &file_models_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GasPrice.ProtoReflect.Descriptor instead.
func (*GasPrice) Descriptor() ([]byte, []int) {
	return file_models_proto_rawDescGZIP(), []int{8}
}

func (x *GasPrice) GetDestChainSelector() uint64 {
	if x != nil {
		return x.DestChainSelector
	}
	return 0
}

func (x *GasPrice) GetValue() *pb.BigInt {
	if x != nil {
		return x.Value
	}
	return nil
}

var File_models_proto protoreflect.FileDescriptor

const file_models_proto_rawDesc = "" +
	"\n" +
	"\fmodels.proto\x12\x15loop.internal.pb.ccip\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\rrelayer.proto\"\xe3\x01\n" +
	"\x06TxMeta\x12;\n" +
	"\x1ablock_timestamp_unix_milli\x18\x01 \x01(\x03R\x17blockTimestampUnixMilli\x12!\n" +
	"\fblock_number\x18\x02 \x01(\x04R\vblockNumber\x12\x17\n" +
	"\atx_hash\x18\x03 \x01(\tR\x06txHash\x12\x1b\n" +
	"\tlog_index\x18\x04 \x01(\x04R\blogIndex\x12C\n" +
	"\tfinalized\x18\x05 \x01(\x0e2%.loop.internal.pb.ccip.FinalityStatusR\tfinalized\"\x94\x04\n" +
	"\x0eEVM2EVMMessage\x12'\n" +
	"\x0fsequence_number\x18\x01 \x01(\x04R\x0esequenceNumber\x12)\n" +
	"\tgas_limit\x18\x02 \x01(\v2\f.loop.BigIntR\bgasLimit\x12\x14\n" +
	"\x05nonce\x18\x03 \x01(\x04R\x05nonce\x12\x1b\n" +
	"\tgas_price\x18\x04 \x01(\x04R\bgasPrice\x12\x1d\n" +
	"\n" +
	"message_id\x18\x05 \x01(\fR\tmessageId\x122\n" +
	"\x15source_chain_selector\x18\x06 \x01(\x04R\x13sourceChainSelector\x12\x16\n" +
	"\x06sender\x18\a \x01(\tR\x06sender\x12\x1a\n" +
	"\breceiver\x18\b \x01(\tR\breceiver\x12\x16\n" +
	"\x06strict\x18\t \x01(\bR\x06strict\x12\x1b\n" +
	"\tfee_token\x18\n" +
	" \x01(\tR\bfeeToken\x126\n" +
	"\x10fee_token_amount\x18\v \x01(\v2\f.loop.BigIntR\x0efeeTokenAmount\x12\x12\n" +
	"\x04data\x18\f \x01(\fR\x04data\x12G\n" +
	"\rtoken_amounts\x18\r \x03(\v2\".loop.internal.pb.ccip.TokenAmountR\ftokenAmounts\x12*\n" +
	"\x11source_token_data\x18\x0e \x03(\fR\x0fsourceTokenData\"\xa9\x02\n" +
	"&EVM2EVMOnRampCCIPSendRequestedWithMeta\x12J\n" +
	"\x0eevm_to_evm_msg\x18\x01 \x01(\v2%.loop.internal.pb.ccip.EVM2EVMMessageR\vevmToEvmMsg\x12C\n" +
	"\x0fblock_timestamp\x18\x02 \x01(\v2\x1a.google.protobuf.TimestampR\x0eblockTimestamp\x12\x1a\n" +
	"\bexecuted\x18\x03 \x01(\bR\bexecuted\x12\x1c\n" +
	"\tfinalized\x18\x04 \x01(\bR\tfinalized\x12\x1b\n" +
	"\tlog_index\x18\x05 \x01(\x04R\blogIndex\x12\x17\n" +
	"\atx_hash\x18\x06 \x01(\tR\x06txHash\"\xc8\x01\n" +
	"\x12TokenPoolRateLimit\x12$\n" +
	"\x06tokens\x18\x01 \x01(\v2\f.loop.BigIntR\x06tokens\x12!\n" +
	"\flast_updated\x18\x02 \x01(\rR\vlastUpdated\x12\x1d\n" +
	"\n" +
	"is_enabled\x18\x03 \x01(\bR\tisEnabled\x12(\n" +
	"\bcapacity\x18\x04 \x01(\v2\f.loop.BigIntR\bcapacity\x12 \n" +
	"\x04rate\x18\x05 \x01(\v2\f.loop.BigIntR\x04rate\"I\n" +
	"\vTokenAmount\x12\x14\n" +
	"\x05token\x18\x01 \x01(\tR\x05token\x12$\n" +
	"\x06amount\x18\x02 \x01(\v2\f.loop.BigIntR\x06amount\"\x87\x02\n" +
	"\x0fExecutionReport\x12T\n" +
	"\x13evm_to_evm_messages\x18\x01 \x03(\v2%.loop.internal.pb.ccip.EVM2EVMMessageR\x10evmToEvmMessages\x12P\n" +
	"\x13offchain_token_data\x18\x02 \x03(\v2 .loop.internal.pb.ccip.TokenDataR\x11offchainTokenData\x12\x16\n" +
	"\x06proofs\x18\x03 \x03(\fR\x06proofs\x124\n" +
	"\x0fproof_flag_bits\x18\x04 \x01(\v2\f.loop.BigIntR\rproofFlagBits\"\x1f\n" +
	"\tTokenData\x12\x12\n" +
	"\x04data\x18\x01 \x03(\fR\x04data\"F\n" +
	"\n" +
	"TokenPrice\x12\x14\n" +
	"\x05token\x18\x01 \x01(\tR\x05token\x12\"\n" +
	"\x05value\x18\x02 \x01(\v2\f.loop.BigIntR\x05value\"^\n" +
	"\bGasPrice\x12.\n" +
	"\x13dest_chain_selector\x18\x01 \x01(\x04R\x11destChainSelector\x12\"\n" +
	"\x05value\x18\x02 \x01(\v2\f.loop.BigIntR\x05value*>\n" +
	"\x0eFinalityStatus\x12\v\n" +
	"\aUnknown\x10\x00\x12\r\n" +
	"\tFinalized\x10\x01\x12\x10\n" +
	"\fNotFinalized\x10\x02BOZMgithub.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip;ccippbb\x06proto3"

var (
	file_models_proto_rawDescOnce sync.Once
	file_models_proto_rawDescData []byte
)

func file_models_proto_rawDescGZIP() []byte {
	file_models_proto_rawDescOnce.Do(func() {
		file_models_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_models_proto_rawDesc), len(file_models_proto_rawDesc)))
	})
	return file_models_proto_rawDescData
}

var file_models_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_models_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_models_proto_goTypes = []any{
	(FinalityStatus)(0),    // 0: loop.internal.pb.ccip.FinalityStatus
	(*TxMeta)(nil),         // 1: loop.internal.pb.ccip.TxMeta
	(*EVM2EVMMessage)(nil), // 2: loop.internal.pb.ccip.EVM2EVMMessage
	(*EVM2EVMOnRampCCIPSendRequestedWithMeta)(nil), // 3: loop.internal.pb.ccip.EVM2EVMOnRampCCIPSendRequestedWithMeta
	(*TokenPoolRateLimit)(nil),                     // 4: loop.internal.pb.ccip.TokenPoolRateLimit
	(*TokenAmount)(nil),                            // 5: loop.internal.pb.ccip.TokenAmount
	(*ExecutionReport)(nil),                        // 6: loop.internal.pb.ccip.ExecutionReport
	(*TokenData)(nil),                              // 7: loop.internal.pb.ccip.TokenData
	(*TokenPrice)(nil),                             // 8: loop.internal.pb.ccip.TokenPrice
	(*GasPrice)(nil),                               // 9: loop.internal.pb.ccip.GasPrice
	(*pb.BigInt)(nil),                              // 10: loop.BigInt
	(*timestamppb.Timestamp)(nil),                  // 11: google.protobuf.Timestamp
}
var file_models_proto_depIdxs = []int32{
	0,  // 0: loop.internal.pb.ccip.TxMeta.finalized:type_name -> loop.internal.pb.ccip.FinalityStatus
	10, // 1: loop.internal.pb.ccip.EVM2EVMMessage.gas_limit:type_name -> loop.BigInt
	10, // 2: loop.internal.pb.ccip.EVM2EVMMessage.fee_token_amount:type_name -> loop.BigInt
	5,  // 3: loop.internal.pb.ccip.EVM2EVMMessage.token_amounts:type_name -> loop.internal.pb.ccip.TokenAmount
	2,  // 4: loop.internal.pb.ccip.EVM2EVMOnRampCCIPSendRequestedWithMeta.evm_to_evm_msg:type_name -> loop.internal.pb.ccip.EVM2EVMMessage
	11, // 5: loop.internal.pb.ccip.EVM2EVMOnRampCCIPSendRequestedWithMeta.block_timestamp:type_name -> google.protobuf.Timestamp
	10, // 6: loop.internal.pb.ccip.TokenPoolRateLimit.tokens:type_name -> loop.BigInt
	10, // 7: loop.internal.pb.ccip.TokenPoolRateLimit.capacity:type_name -> loop.BigInt
	10, // 8: loop.internal.pb.ccip.TokenPoolRateLimit.rate:type_name -> loop.BigInt
	10, // 9: loop.internal.pb.ccip.TokenAmount.amount:type_name -> loop.BigInt
	2,  // 10: loop.internal.pb.ccip.ExecutionReport.evm_to_evm_messages:type_name -> loop.internal.pb.ccip.EVM2EVMMessage
	7,  // 11: loop.internal.pb.ccip.ExecutionReport.offchain_token_data:type_name -> loop.internal.pb.ccip.TokenData
	10, // 12: loop.internal.pb.ccip.ExecutionReport.proof_flag_bits:type_name -> loop.BigInt
	10, // 13: loop.internal.pb.ccip.TokenPrice.value:type_name -> loop.BigInt
	10, // 14: loop.internal.pb.ccip.GasPrice.value:type_name -> loop.BigInt
	15, // [15:15] is the sub-list for method output_type
	15, // [15:15] is the sub-list for method input_type
	15, // [15:15] is the sub-list for extension type_name
	15, // [15:15] is the sub-list for extension extendee
	0,  // [0:15] is the sub-list for field type_name
}

func init() { file_models_proto_init() }
func file_models_proto_init() {
	if File_models_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_models_proto_rawDesc), len(file_models_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_models_proto_goTypes,
		DependencyIndexes: file_models_proto_depIdxs,
		EnumInfos:         file_models_proto_enumTypes,
		MessageInfos:      file_models_proto_msgTypes,
	}.Build()
	File_models_proto = out.File
	file_models_proto_goTypes = nil
	file_models_proto_depIdxs = nil
}
