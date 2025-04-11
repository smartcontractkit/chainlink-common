// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: evm.proto

package evm

import (
	_ "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/pb"
	crosschain "github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/crosschain"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ConfidenceLevel int32

const (
	ConfidenceLevel_CONFIDENCE_LEVEL_UNSPECIFIED ConfidenceLevel = 0
	ConfidenceLevel_LOW                          ConfidenceLevel = 1
	ConfidenceLevel_MEDIUM                       ConfidenceLevel = 2
	ConfidenceLevel_HIGH                         ConfidenceLevel = 3
	ConfidenceLevel_FINALITY                     ConfidenceLevel = 4
)

// Enum value maps for ConfidenceLevel.
var (
	ConfidenceLevel_name = map[int32]string{
		0: "CONFIDENCE_LEVEL_UNSPECIFIED",
		1: "LOW",
		2: "MEDIUM",
		3: "HIGH",
		4: "FINALITY",
	}
	ConfidenceLevel_value = map[string]int32{
		"CONFIDENCE_LEVEL_UNSPECIFIED": 0,
		"LOW":                          1,
		"MEDIUM":                       2,
		"HIGH":                         3,
		"FINALITY":                     4,
	}
)

func (x ConfidenceLevel) Enum() *ConfidenceLevel {
	p := new(ConfidenceLevel)
	*p = x
	return p
}

func (x ConfidenceLevel) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ConfidenceLevel) Descriptor() protoreflect.EnumDescriptor {
	return file_evm_proto_enumTypes[0].Descriptor()
}

func (ConfidenceLevel) Type() protoreflect.EnumType {
	return &file_evm_proto_enumTypes[0]
}

func (x ConfidenceLevel) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ConfidenceLevel.Descriptor instead.
func (ConfidenceLevel) EnumDescriptor() ([]byte, []int) {
	return file_evm_proto_rawDescGZIP(), []int{0}
}

type TxID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *TxID) Reset() {
	*x = TxID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evm_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TxID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TxID) ProtoMessage() {}

func (x *TxID) ProtoReflect() protoreflect.Message {
	mi := &file_evm_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TxID.ProtoReflect.Descriptor instead.
func (*TxID) Descriptor() ([]byte, []int) {
	return file_evm_proto_rawDescGZIP(), []int{0}
}

func (x *TxID) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type ReadMethodRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address         string          `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Calldata        []byte          `protobuf:"bytes,2,opt,name=calldata,proto3" json:"calldata,omitempty"`
	ConfidenceLevel ConfidenceLevel `protobuf:"varint,3,opt,name=confidence_level,json=confidenceLevel,proto3,enum=don.evm.v1.ConfidenceLevel" json:"confidence_level,omitempty"`
}

func (x *ReadMethodRequest) Reset() {
	*x = ReadMethodRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evm_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadMethodRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadMethodRequest) ProtoMessage() {}

func (x *ReadMethodRequest) ProtoReflect() protoreflect.Message {
	mi := &file_evm_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadMethodRequest.ProtoReflect.Descriptor instead.
func (*ReadMethodRequest) Descriptor() ([]byte, []int) {
	return file_evm_proto_rawDescGZIP(), []int{1}
}

func (x *ReadMethodRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ReadMethodRequest) GetCalldata() []byte {
	if x != nil {
		return x.Calldata
	}
	return nil
}

func (x *ReadMethodRequest) GetConfidenceLevel() ConfidenceLevel {
	if x != nil {
		return x.ConfidenceLevel
	}
	return ConfidenceLevel_CONFIDENCE_LEVEL_UNSPECIFIED
}

// Represents a request to query logs
type QueryLogsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Filter          *FilterQuery    `protobuf:"bytes,1,opt,name=filter,proto3" json:"filter,omitempty"`
	ConfidenceLevel ConfidenceLevel `protobuf:"varint,2,opt,name=confidence_level,json=confidenceLevel,proto3,enum=don.evm.v1.ConfidenceLevel" json:"confidence_level,omitempty"`
}

func (x *QueryLogsRequest) Reset() {
	*x = QueryLogsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evm_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryLogsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryLogsRequest) ProtoMessage() {}

func (x *QueryLogsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_evm_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryLogsRequest.ProtoReflect.Descriptor instead.
func (*QueryLogsRequest) Descriptor() ([]byte, []int) {
	return file_evm_proto_rawDescGZIP(), []int{2}
}

func (x *QueryLogsRequest) GetFilter() *FilterQuery {
	if x != nil {
		return x.Filter
	}
	return nil
}

func (x *QueryLogsRequest) GetConfidenceLevel() ConfidenceLevel {
	if x != nil {
		return x.ConfidenceLevel
	}
	return ConfidenceLevel_CONFIDENCE_LEVEL_UNSPECIFIED
}

// Represents a request to submit a transaction
type SubmitTransactionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ToAddress string     `protobuf:"bytes,1,opt,name=to_address,json=toAddress,proto3" json:"to_address,omitempty"`
	GasConfig *GasConfig `protobuf:"bytes,2,opt,name=gas_config,json=gasConfig,proto3" json:"gas_config,omitempty"`
	Calldata  []byte     `protobuf:"bytes,3,opt,name=calldata,proto3" json:"calldata,omitempty"` // Optional
}

func (x *SubmitTransactionRequest) Reset() {
	*x = SubmitTransactionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evm_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubmitTransactionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitTransactionRequest) ProtoMessage() {}

func (x *SubmitTransactionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_evm_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmitTransactionRequest.ProtoReflect.Descriptor instead.
func (*SubmitTransactionRequest) Descriptor() ([]byte, []int) {
	return file_evm_proto_rawDescGZIP(), []int{3}
}

func (x *SubmitTransactionRequest) GetToAddress() string {
	if x != nil {
		return x.ToAddress
	}
	return ""
}

func (x *SubmitTransactionRequest) GetGasConfig() *GasConfig {
	if x != nil {
		return x.GasConfig
	}
	return nil
}

func (x *SubmitTransactionRequest) GetCalldata() []byte {
	if x != nil {
		return x.Calldata
	}
	return nil
}

// Placeholder messages for types defined elsewhere
type GasConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GasConfig) Reset() {
	*x = GasConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evm_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GasConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GasConfig) ProtoMessage() {}

func (x *GasConfig) ProtoReflect() protoreflect.Message {
	mi := &file_evm_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GasConfig.ProtoReflect.Descriptor instead.
func (*GasConfig) Descriptor() ([]byte, []int) {
	return file_evm_proto_rawDescGZIP(), []int{4}
}

type FilterQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *FilterQuery) Reset() {
	*x = FilterQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evm_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FilterQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FilterQuery) ProtoMessage() {}

func (x *FilterQuery) ProtoReflect() protoreflect.Message {
	mi := &file_evm_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FilterQuery.ProtoReflect.Descriptor instead.
func (*FilterQuery) Descriptor() ([]byte, []int) {
	return file_evm_proto_rawDescGZIP(), []int{5}
}

type Log struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Log) Reset() {
	*x = Log{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evm_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Log) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Log) ProtoMessage() {}

func (x *Log) ProtoReflect() protoreflect.Message {
	mi := &file_evm_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Log.ProtoReflect.Descriptor instead.
func (*Log) Descriptor() ([]byte, []int) {
	return file_evm_proto_rawDescGZIP(), []int{6}
}

type LogList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Logs []*Log `protobuf:"bytes,1,rep,name=logs,proto3" json:"logs,omitempty"`
}

func (x *LogList) Reset() {
	*x = LogList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evm_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogList) ProtoMessage() {}

func (x *LogList) ProtoReflect() protoreflect.Message {
	mi := &file_evm_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogList.ProtoReflect.Descriptor instead.
func (*LogList) Descriptor() ([]byte, []int) {
	return file_evm_proto_rawDescGZIP(), []int{7}
}

func (x *LogList) GetLogs() []*Log {
	if x != nil {
		return x.Logs
	}
	return nil
}

var File_evm_proto protoreflect.FileDescriptor

var file_evm_proto_rawDesc = []byte{
	0x0a, 0x09, 0x65, 0x76, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x64, 0x6f, 0x6e,
	0x2e, 0x65, 0x76, 0x6d, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x63, 0x72, 0x6f, 0x73, 0x73, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0e, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1c, 0x0a, 0x04, 0x54, 0x78, 0x49, 0x44, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x22, 0x91, 0x01, 0x0a, 0x11, 0x52, 0x65, 0x61, 0x64, 0x4d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x61, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x63, 0x61, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61,
	0x12, 0x46, 0x0a, 0x10, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x5f, 0x6c,
	0x65, 0x76, 0x65, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x64, 0x6f, 0x6e,
	0x2e, 0x65, 0x76, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x64, 0x65, 0x6e,
	0x63, 0x65, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x52, 0x0f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x64, 0x65,
	0x6e, 0x63, 0x65, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x22, 0x8b, 0x01, 0x0a, 0x10, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2f, 0x0a,
	0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e,
	0x64, 0x6f, 0x6e, 0x2e, 0x65, 0x76, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x74, 0x65,
	0x72, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x46,
	0x0a, 0x10, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x5f, 0x6c, 0x65, 0x76,
	0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x65,
	0x76, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x64, 0x65, 0x6e, 0x63, 0x65,
	0x4c, 0x65, 0x76, 0x65, 0x6c, 0x52, 0x0f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x64, 0x65, 0x6e, 0x63,
	0x65, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x22, 0x8b, 0x01, 0x0a, 0x18, 0x53, 0x75, 0x62, 0x6d, 0x69,
	0x74, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x74, 0x6f, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x6f, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x12, 0x34, 0x0a, 0x0a, 0x67, 0x61, 0x73, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x65, 0x76, 0x6d,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x61, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x09, 0x67,
	0x61, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x61, 0x6c, 0x6c,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x63, 0x61, 0x6c, 0x6c,
	0x64, 0x61, 0x74, 0x61, 0x22, 0x0b, 0x0a, 0x09, 0x47, 0x61, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x22, 0x0d, 0x0a, 0x0b, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x51, 0x75, 0x65, 0x72, 0x79,
	0x22, 0x05, 0x0a, 0x03, 0x4c, 0x6f, 0x67, 0x22, 0x2e, 0x0a, 0x07, 0x4c, 0x6f, 0x67, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x23, 0x0a, 0x04, 0x6c, 0x6f, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x0f, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x65, 0x76, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x6f,
	0x67, 0x52, 0x04, 0x6c, 0x6f, 0x67, 0x73, 0x2a, 0x60, 0x0a, 0x0f, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x64, 0x65, 0x6e, 0x63, 0x65, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x20, 0x0a, 0x1c, 0x43, 0x4f,
	0x4e, 0x46, 0x49, 0x44, 0x45, 0x4e, 0x43, 0x45, 0x5f, 0x4c, 0x45, 0x56, 0x45, 0x4c, 0x5f, 0x55,
	0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03,
	0x4c, 0x4f, 0x57, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x4d, 0x45, 0x44, 0x49, 0x55, 0x4d, 0x10,
	0x02, 0x12, 0x08, 0x0a, 0x04, 0x48, 0x49, 0x47, 0x48, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08, 0x46,
	0x49, 0x4e, 0x41, 0x4c, 0x49, 0x54, 0x59, 0x10, 0x04, 0x32, 0xf2, 0x02, 0x0a, 0x06, 0x43, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x12, 0x3c, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x54, 0x78, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x12, 0x10, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x65, 0x76, 0x6d, 0x2e, 0x76, 0x31,
	0x2e, 0x54, 0x78, 0x49, 0x44, 0x1a, 0x1b, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x63, 0x72, 0x6f, 0x73,
	0x73, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x78, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x12, 0x49, 0x0a, 0x0a, 0x52, 0x65, 0x61, 0x64, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64,
	0x12, 0x1d, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x65, 0x76, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65,
	0x61, 0x64, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1c, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x63, 0x72, 0x6f, 0x73, 0x73, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x2e, 0x76, 0x31, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x41, 0x72, 0x72, 0x61, 0x79, 0x12, 0x3e, 0x0a,
	0x09, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4c, 0x6f, 0x67, 0x73, 0x12, 0x1c, 0x2e, 0x64, 0x6f, 0x6e,
	0x2e, 0x65, 0x76, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4c, 0x6f, 0x67,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x65,
	0x76, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x4b, 0x0a,
	0x11, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x24, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x65, 0x76, 0x6d, 0x2e, 0x76, 0x31, 0x2e,
	0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10, 0x2e, 0x64, 0x6f, 0x6e, 0x2e, 0x65,
	0x76, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x78, 0x49, 0x44, 0x12, 0x52, 0x0a, 0x13, 0x4f, 0x6e,
	0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79, 0x56, 0x69, 0x6f, 0x6c, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1d, 0x2e, 0x64, 0x6f, 0x6e, 0x2e,
	0x63, 0x72, 0x6f, 0x73, 0x73, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x22, 0x04, 0xd8, 0xb5, 0x18, 0x01, 0x42, 0x4d,
	0x5a, 0x4b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6d, 0x61,
	0x72, 0x74, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6b, 0x69, 0x74, 0x2f, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x2f,
	0x73, 0x74, 0x75, 0x62, 0x73, 0x2f, 0x64, 0x6f, 0x6e, 0x2f, 0x65, 0x76, 0x6d, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_evm_proto_rawDescOnce sync.Once
	file_evm_proto_rawDescData = file_evm_proto_rawDesc
)

func file_evm_proto_rawDescGZIP() []byte {
	file_evm_proto_rawDescOnce.Do(func() {
		file_evm_proto_rawDescData = protoimpl.X.CompressGZIP(file_evm_proto_rawDescData)
	})
	return file_evm_proto_rawDescData
}

var file_evm_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_evm_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_evm_proto_goTypes = []interface{}{
	(ConfidenceLevel)(0),             // 0: don.evm.v1.ConfidenceLevel
	(*TxID)(nil),                     // 1: don.evm.v1.TxID
	(*ReadMethodRequest)(nil),        // 2: don.evm.v1.ReadMethodRequest
	(*QueryLogsRequest)(nil),         // 3: don.evm.v1.QueryLogsRequest
	(*SubmitTransactionRequest)(nil), // 4: don.evm.v1.SubmitTransactionRequest
	(*GasConfig)(nil),                // 5: don.evm.v1.GasConfig
	(*FilterQuery)(nil),              // 6: don.evm.v1.FilterQuery
	(*Log)(nil),                      // 7: don.evm.v1.Log
	(*LogList)(nil),                  // 8: don.evm.v1.LogList
	(*emptypb.Empty)(nil),            // 9: google.protobuf.Empty
	(*crosschain.TxResult)(nil),      // 10: don.crosschain.v1.TxResult
	(*crosschain.ByteArray)(nil),     // 11: don.crosschain.v1.ByteArray
	(*crosschain.BlockRange)(nil),    // 12: don.crosschain.v1.BlockRange
}
var file_evm_proto_depIdxs = []int32{
	0,  // 0: don.evm.v1.ReadMethodRequest.confidence_level:type_name -> don.evm.v1.ConfidenceLevel
	6,  // 1: don.evm.v1.QueryLogsRequest.filter:type_name -> don.evm.v1.FilterQuery
	0,  // 2: don.evm.v1.QueryLogsRequest.confidence_level:type_name -> don.evm.v1.ConfidenceLevel
	5,  // 3: don.evm.v1.SubmitTransactionRequest.gas_config:type_name -> don.evm.v1.GasConfig
	7,  // 4: don.evm.v1.LogList.logs:type_name -> don.evm.v1.Log
	1,  // 5: don.evm.v1.Client.GetTxResult:input_type -> don.evm.v1.TxID
	2,  // 6: don.evm.v1.Client.ReadMethod:input_type -> don.evm.v1.ReadMethodRequest
	3,  // 7: don.evm.v1.Client.QueryLogs:input_type -> don.evm.v1.QueryLogsRequest
	4,  // 8: don.evm.v1.Client.SubmitTransaction:input_type -> don.evm.v1.SubmitTransactionRequest
	9,  // 9: don.evm.v1.Client.OnFinalityViolation:input_type -> google.protobuf.Empty
	10, // 10: don.evm.v1.Client.GetTxResult:output_type -> don.crosschain.v1.TxResult
	11, // 11: don.evm.v1.Client.ReadMethod:output_type -> don.crosschain.v1.ByteArray
	8,  // 12: don.evm.v1.Client.QueryLogs:output_type -> don.evm.v1.LogList
	1,  // 13: don.evm.v1.Client.SubmitTransaction:output_type -> don.evm.v1.TxID
	12, // 14: don.evm.v1.Client.OnFinalityViolation:output_type -> don.crosschain.v1.BlockRange
	10, // [10:15] is the sub-list for method output_type
	5,  // [5:10] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_evm_proto_init() }
func file_evm_proto_init() {
	if File_evm_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_evm_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TxID); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evm_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadMethodRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evm_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryLogsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evm_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubmitTransactionRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evm_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GasConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evm_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FilterQuery); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evm_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Log); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evm_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogList); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_evm_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_evm_proto_goTypes,
		DependencyIndexes: file_evm_proto_depIdxs,
		EnumInfos:         file_evm_proto_enumTypes,
		MessageInfos:      file_evm_proto_msgTypes,
	}.Build()
	File_evm_proto = out.File
	file_evm_proto_rawDesc = nil
	file_evm_proto_goTypes = nil
	file_evm_proto_depIdxs = nil
}
