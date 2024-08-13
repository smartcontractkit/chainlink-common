// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: datasource_v3.proto

package mercuryv3pb

import (
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// ObserveRequest is the request payload for the Observe method, which is a gRPC adapter for input arguments of [pkg/types/mercury/v1/DataSource.Observe]
type ObserveRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReportTimestamp           *pb.ReportTimestamp `protobuf:"bytes,1,opt,name=report_timestamp,json=reportTimestamp,proto3" json:"report_timestamp,omitempty"`
	FetchMaxFinalizedBlockNum bool                `protobuf:"varint,2,opt,name=fetchMaxFinalizedBlockNum,proto3" json:"fetchMaxFinalizedBlockNum,omitempty"`
}

func (x *ObserveRequest) Reset() {
	*x = ObserveRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datasource_v3_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObserveRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObserveRequest) ProtoMessage() {}

func (x *ObserveRequest) ProtoReflect() protoreflect.Message {
	mi := &file_datasource_v3_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObserveRequest.ProtoReflect.Descriptor instead.
func (*ObserveRequest) Descriptor() ([]byte, []int) {
	return file_datasource_v3_proto_rawDescGZIP(), []int{0}
}

func (x *ObserveRequest) GetReportTimestamp() *pb.ReportTimestamp {
	if x != nil {
		return x.ReportTimestamp
	}
	return nil
}

func (x *ObserveRequest) GetFetchMaxFinalizedBlockNum() bool {
	if x != nil {
		return x.FetchMaxFinalizedBlockNum
	}
	return false
}

// Block is a gRPC adapter for [pkg/types/mercury/v1/Block]
type Block struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Number    int64  `protobuf:"varint,1,opt,name=number,proto3" json:"number,omitempty"`
	Hash      []byte `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	Timestamp uint64 `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Block) Reset() {
	*x = Block{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datasource_v3_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Block) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Block) ProtoMessage() {}

func (x *Block) ProtoReflect() protoreflect.Message {
	mi := &file_datasource_v3_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Block.ProtoReflect.Descriptor instead.
func (*Block) Descriptor() ([]byte, []int) {
	return file_datasource_v3_proto_rawDescGZIP(), []int{1}
}

func (x *Block) GetNumber() int64 {
	if x != nil {
		return x.Number
	}
	return 0
}

func (x *Block) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (x *Block) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

// Observation is a gRPC adapter for [pkg/types/mercury/v1/Observation]
type Observation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BenchmarkPrice        *pb.BigInt `protobuf:"bytes,1,opt,name=benchmarkPrice,proto3" json:"benchmarkPrice,omitempty"`
	Bid                   *pb.BigInt `protobuf:"bytes,2,opt,name=bid,proto3" json:"bid,omitempty"`
	Ask                   *pb.BigInt `protobuf:"bytes,3,opt,name=ask,proto3" json:"ask,omitempty"`
	MaxFinalizedTimestamp int64      `protobuf:"varint,4,opt,name=maxFinalizedTimestamp,proto3" json:"maxFinalizedTimestamp,omitempty"`
	LinkPrice             *pb.BigInt `protobuf:"bytes,5,opt,name=linkPrice,proto3" json:"linkPrice,omitempty"`
	NativePrice           *pb.BigInt `protobuf:"bytes,6,opt,name=nativePrice,proto3" json:"nativePrice,omitempty"`
}

func (x *Observation) Reset() {
	*x = Observation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datasource_v3_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Observation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Observation) ProtoMessage() {}

func (x *Observation) ProtoReflect() protoreflect.Message {
	mi := &file_datasource_v3_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Observation.ProtoReflect.Descriptor instead.
func (*Observation) Descriptor() ([]byte, []int) {
	return file_datasource_v3_proto_rawDescGZIP(), []int{2}
}

func (x *Observation) GetBenchmarkPrice() *pb.BigInt {
	if x != nil {
		return x.BenchmarkPrice
	}
	return nil
}

func (x *Observation) GetBid() *pb.BigInt {
	if x != nil {
		return x.Bid
	}
	return nil
}

func (x *Observation) GetAsk() *pb.BigInt {
	if x != nil {
		return x.Ask
	}
	return nil
}

func (x *Observation) GetMaxFinalizedTimestamp() int64 {
	if x != nil {
		return x.MaxFinalizedTimestamp
	}
	return 0
}

func (x *Observation) GetLinkPrice() *pb.BigInt {
	if x != nil {
		return x.LinkPrice
	}
	return nil
}

func (x *Observation) GetNativePrice() *pb.BigInt {
	if x != nil {
		return x.NativePrice
	}
	return nil
}

// ObserveResponse is the response payload for the Observe method, which is a gRPC adapter for output arguments of [pkg/types/mercury/v1/DataSource.Observe]
type ObserveResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Observation *Observation `protobuf:"bytes,1,opt,name=observation,proto3" json:"observation,omitempty"`
}

func (x *ObserveResponse) Reset() {
	*x = ObserveResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datasource_v3_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObserveResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObserveResponse) ProtoMessage() {}

func (x *ObserveResponse) ProtoReflect() protoreflect.Message {
	mi := &file_datasource_v3_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObserveResponse.ProtoReflect.Descriptor instead.
func (*ObserveResponse) Descriptor() ([]byte, []int) {
	return file_datasource_v3_proto_rawDescGZIP(), []int{3}
}

func (x *ObserveResponse) GetObservation() *Observation {
	if x != nil {
		return x.Observation
	}
	return nil
}

var File_datasource_v3_proto protoreflect.FileDescriptor

var file_datasource_v3_proto_rawDesc = []byte{
	0x0a, 0x13, 0x64, 0x61, 0x74, 0x61, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x76, 0x33, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1b, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x2e,
	0x76, 0x33, 0x1a, 0x0d, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x90, 0x01, 0x0a, 0x0e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x40, 0x0a, 0x10, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15,
	0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x3c, 0x0a, 0x19, 0x66, 0x65, 0x74, 0x63, 0x68, 0x4d,
	0x61, 0x78, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x4e, 0x75, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x19, 0x66, 0x65, 0x74, 0x63, 0x68,
	0x4d, 0x61, 0x78, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x4e, 0x75, 0x6d, 0x22, 0x51, 0x0a, 0x05, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x16, 0x0a,
	0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x95, 0x02, 0x0a, 0x0b, 0x4f, 0x62, 0x73, 0x65,
	0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x34, 0x0a, 0x0e, 0x62, 0x65, 0x6e, 0x63, 0x68,
	0x6d, 0x61, 0x72, 0x6b, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0c, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x42, 0x69, 0x67, 0x49, 0x6e, 0x74, 0x52, 0x0e, 0x62,
	0x65, 0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72, 0x6b, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12, 0x1e, 0x0a,
	0x03, 0x62, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6c, 0x6f, 0x6f,
	0x70, 0x2e, 0x42, 0x69, 0x67, 0x49, 0x6e, 0x74, 0x52, 0x03, 0x62, 0x69, 0x64, 0x12, 0x1e, 0x0a,
	0x03, 0x61, 0x73, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6c, 0x6f, 0x6f,
	0x70, 0x2e, 0x42, 0x69, 0x67, 0x49, 0x6e, 0x74, 0x52, 0x03, 0x61, 0x73, 0x6b, 0x12, 0x34, 0x0a,
	0x15, 0x6d, 0x61, 0x78, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x15, 0x6d, 0x61,
	0x78, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x12, 0x2a, 0x0a, 0x09, 0x6c, 0x69, 0x6e, 0x6b, 0x50, 0x72, 0x69, 0x63, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x42, 0x69,
	0x67, 0x49, 0x6e, 0x74, 0x52, 0x09, 0x6c, 0x69, 0x6e, 0x6b, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12,
	0x2e, 0x0a, 0x0b, 0x6e, 0x61, 0x74, 0x69, 0x76, 0x65, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x42, 0x69, 0x67, 0x49,
	0x6e, 0x74, 0x52, 0x0b, 0x6e, 0x61, 0x74, 0x69, 0x76, 0x65, 0x50, 0x72, 0x69, 0x63, 0x65, 0x22,
	0x5d, 0x0a, 0x0f, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x4a, 0x0a, 0x0b, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75,
	0x72, 0x79, 0x2e, 0x76, 0x33, 0x2e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x0b, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x32, 0x74,
	0x0a, 0x0a, 0x44, 0x61, 0x74, 0x61, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x66, 0x0a, 0x07,
	0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x12, 0x2b, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75,
	0x72, 0x79, 0x2e, 0x76, 0x33, 0x2e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x2e,
	0x76, 0x33, 0x2e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x42, 0x5a, 0x5a, 0x58, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74,
	0x6b, 0x69, 0x74, 0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x2d, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6c, 0x6f, 0x6f, 0x70, 0x2f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x62, 0x2f, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72,
	0x79, 0x2f, 0x76, 0x33, 0x3b, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x76, 0x33, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_datasource_v3_proto_rawDescOnce sync.Once
	file_datasource_v3_proto_rawDescData = file_datasource_v3_proto_rawDesc
)

func file_datasource_v3_proto_rawDescGZIP() []byte {
	file_datasource_v3_proto_rawDescOnce.Do(func() {
		file_datasource_v3_proto_rawDescData = protoimpl.X.CompressGZIP(file_datasource_v3_proto_rawDescData)
	})
	return file_datasource_v3_proto_rawDescData
}

var file_datasource_v3_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_datasource_v3_proto_goTypes = []interface{}{
	(*ObserveRequest)(nil),     // 0: loop.internal.pb.mercury.v3.ObserveRequest
	(*Block)(nil),              // 1: loop.internal.pb.mercury.v3.Block
	(*Observation)(nil),        // 2: loop.internal.pb.mercury.v3.Observation
	(*ObserveResponse)(nil),    // 3: loop.internal.pb.mercury.v3.ObserveResponse
	(*pb.ReportTimestamp)(nil), // 4: loop.ReportTimestamp
	(*pb.BigInt)(nil),          // 5: loop.BigInt
}
var file_datasource_v3_proto_depIdxs = []int32{
	4, // 0: loop.internal.pb.mercury.v3.ObserveRequest.report_timestamp:type_name -> loop.ReportTimestamp
	5, // 1: loop.internal.pb.mercury.v3.Observation.benchmarkPrice:type_name -> loop.BigInt
	5, // 2: loop.internal.pb.mercury.v3.Observation.bid:type_name -> loop.BigInt
	5, // 3: loop.internal.pb.mercury.v3.Observation.ask:type_name -> loop.BigInt
	5, // 4: loop.internal.pb.mercury.v3.Observation.linkPrice:type_name -> loop.BigInt
	5, // 5: loop.internal.pb.mercury.v3.Observation.nativePrice:type_name -> loop.BigInt
	2, // 6: loop.internal.pb.mercury.v3.ObserveResponse.observation:type_name -> loop.internal.pb.mercury.v3.Observation
	0, // 7: loop.internal.pb.mercury.v3.DataSource.Observe:input_type -> loop.internal.pb.mercury.v3.ObserveRequest
	3, // 8: loop.internal.pb.mercury.v3.DataSource.Observe:output_type -> loop.internal.pb.mercury.v3.ObserveResponse
	8, // [8:9] is the sub-list for method output_type
	7, // [7:8] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_datasource_v3_proto_init() }
func file_datasource_v3_proto_init() {
	if File_datasource_v3_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_datasource_v3_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObserveRequest); i {
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
		file_datasource_v3_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Block); i {
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
		file_datasource_v3_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Observation); i {
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
		file_datasource_v3_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObserveResponse); i {
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
			RawDescriptor: file_datasource_v3_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_datasource_v3_proto_goTypes,
		DependencyIndexes: file_datasource_v3_proto_depIdxs,
		MessageInfos:      file_datasource_v3_proto_msgTypes,
	}.Build()
	File_datasource_v3_proto = out.File
	file_datasource_v3_proto_rawDesc = nil
	file_datasource_v3_proto_goTypes = nil
	file_datasource_v3_proto_depIdxs = nil
}
