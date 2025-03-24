// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: datasource_v4.proto

package mercuryv4pb

import (
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
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

// ObserveRequest is the request payload for the Observe method, which is a gRPC adapter for input arguments of [pkg/types/mercury/v1/DataSource.Observe]
type ObserveRequest struct {
	state                     protoimpl.MessageState `protogen:"open.v1"`
	ReportTimestamp           *pb.ReportTimestamp    `protobuf:"bytes,1,opt,name=report_timestamp,json=reportTimestamp,proto3" json:"report_timestamp,omitempty"`
	FetchMaxFinalizedBlockNum bool                   `protobuf:"varint,2,opt,name=fetchMaxFinalizedBlockNum,proto3" json:"fetchMaxFinalizedBlockNum,omitempty"`
	unknownFields             protoimpl.UnknownFields
	sizeCache                 protoimpl.SizeCache
}

func (x *ObserveRequest) Reset() {
	*x = ObserveRequest{}
	mi := &file_datasource_v4_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ObserveRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObserveRequest) ProtoMessage() {}

func (x *ObserveRequest) ProtoReflect() protoreflect.Message {
	mi := &file_datasource_v4_proto_msgTypes[0]
	if x != nil {
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
	return file_datasource_v4_proto_rawDescGZIP(), []int{0}
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	Number        int64                  `protobuf:"varint,1,opt,name=number,proto3" json:"number,omitempty"`
	Hash          []byte                 `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	Timestamp     uint64                 `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Block) Reset() {
	*x = Block{}
	mi := &file_datasource_v4_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Block) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Block) ProtoMessage() {}

func (x *Block) ProtoReflect() protoreflect.Message {
	mi := &file_datasource_v4_proto_msgTypes[1]
	if x != nil {
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
	return file_datasource_v4_proto_rawDescGZIP(), []int{1}
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
	state          protoimpl.MessageState `protogen:"open.v1"`
	BenchmarkPrice *pb.BigInt             `protobuf:"bytes,1,opt,name=benchmarkPrice,proto3" json:"benchmarkPrice,omitempty"`
	// Deprecated: Marked as deprecated in datasource_v4.proto.
	Bid *pb.BigInt `protobuf:"bytes,2,opt,name=bid,proto3" json:"bid,omitempty"` // Field not used for v4.
	// Deprecated: Marked as deprecated in datasource_v4.proto.
	Ask                   *pb.BigInt `protobuf:"bytes,3,opt,name=ask,proto3" json:"ask,omitempty"` // Field not used for v4.
	MaxFinalizedTimestamp int64      `protobuf:"varint,4,opt,name=maxFinalizedTimestamp,proto3" json:"maxFinalizedTimestamp,omitempty"`
	LinkPrice             *pb.BigInt `protobuf:"bytes,5,opt,name=linkPrice,proto3" json:"linkPrice,omitempty"`
	NativePrice           *pb.BigInt `protobuf:"bytes,6,opt,name=nativePrice,proto3" json:"nativePrice,omitempty"`
	MarketStatus          uint32     `protobuf:"varint,7,opt,name=marketStatus,proto3" json:"marketStatus,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *Observation) Reset() {
	*x = Observation{}
	mi := &file_datasource_v4_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Observation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Observation) ProtoMessage() {}

func (x *Observation) ProtoReflect() protoreflect.Message {
	mi := &file_datasource_v4_proto_msgTypes[2]
	if x != nil {
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
	return file_datasource_v4_proto_rawDescGZIP(), []int{2}
}

func (x *Observation) GetBenchmarkPrice() *pb.BigInt {
	if x != nil {
		return x.BenchmarkPrice
	}
	return nil
}

// Deprecated: Marked as deprecated in datasource_v4.proto.
func (x *Observation) GetBid() *pb.BigInt {
	if x != nil {
		return x.Bid
	}
	return nil
}

// Deprecated: Marked as deprecated in datasource_v4.proto.
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

func (x *Observation) GetMarketStatus() uint32 {
	if x != nil {
		return x.MarketStatus
	}
	return 0
}

// ObserveResponse is the response payload for the Observe method, which is a gRPC adapter for output arguments of [pkg/types/mercury/v1/DataSource.Observe]
type ObserveResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Observation   *Observation           `protobuf:"bytes,1,opt,name=observation,proto3" json:"observation,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ObserveResponse) Reset() {
	*x = ObserveResponse{}
	mi := &file_datasource_v4_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ObserveResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObserveResponse) ProtoMessage() {}

func (x *ObserveResponse) ProtoReflect() protoreflect.Message {
	mi := &file_datasource_v4_proto_msgTypes[3]
	if x != nil {
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
	return file_datasource_v4_proto_rawDescGZIP(), []int{3}
}

func (x *ObserveResponse) GetObservation() *Observation {
	if x != nil {
		return x.Observation
	}
	return nil
}

var File_datasource_v4_proto protoreflect.FileDescriptor

const file_datasource_v4_proto_rawDesc = "" +
	"\n" +
	"\x13datasource_v4.proto\x12\x1bloop.internal.pb.mercury.v4\x1a\rrelayer.proto\"\x90\x01\n" +
	"\x0eObserveRequest\x12@\n" +
	"\x10report_timestamp\x18\x01 \x01(\v2\x15.loop.ReportTimestampR\x0freportTimestamp\x12<\n" +
	"\x19fetchMaxFinalizedBlockNum\x18\x02 \x01(\bR\x19fetchMaxFinalizedBlockNum\"Q\n" +
	"\x05Block\x12\x16\n" +
	"\x06number\x18\x01 \x01(\x03R\x06number\x12\x12\n" +
	"\x04hash\x18\x02 \x01(\fR\x04hash\x12\x1c\n" +
	"\ttimestamp\x18\x03 \x01(\x04R\ttimestamp\"\xc1\x02\n" +
	"\vObservation\x124\n" +
	"\x0ebenchmarkPrice\x18\x01 \x01(\v2\f.loop.BigIntR\x0ebenchmarkPrice\x12\"\n" +
	"\x03bid\x18\x02 \x01(\v2\f.loop.BigIntB\x02\x18\x01R\x03bid\x12\"\n" +
	"\x03ask\x18\x03 \x01(\v2\f.loop.BigIntB\x02\x18\x01R\x03ask\x124\n" +
	"\x15maxFinalizedTimestamp\x18\x04 \x01(\x03R\x15maxFinalizedTimestamp\x12*\n" +
	"\tlinkPrice\x18\x05 \x01(\v2\f.loop.BigIntR\tlinkPrice\x12.\n" +
	"\vnativePrice\x18\x06 \x01(\v2\f.loop.BigIntR\vnativePrice\x12\"\n" +
	"\fmarketStatus\x18\a \x01(\rR\fmarketStatus\"]\n" +
	"\x0fObserveResponse\x12J\n" +
	"\vobservation\x18\x01 \x01(\v2(.loop.internal.pb.mercury.v4.ObservationR\vobservation2t\n" +
	"\n" +
	"DataSource\x12f\n" +
	"\aObserve\x12+.loop.internal.pb.mercury.v4.ObserveRequest\x1a,.loop.internal.pb.mercury.v4.ObserveResponse\"\x00BZZXgithub.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/mercury/v4;mercuryv4pbb\x06proto3"

var (
	file_datasource_v4_proto_rawDescOnce sync.Once
	file_datasource_v4_proto_rawDescData []byte
)

func file_datasource_v4_proto_rawDescGZIP() []byte {
	file_datasource_v4_proto_rawDescOnce.Do(func() {
		file_datasource_v4_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_datasource_v4_proto_rawDesc), len(file_datasource_v4_proto_rawDesc)))
	})
	return file_datasource_v4_proto_rawDescData
}

var file_datasource_v4_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_datasource_v4_proto_goTypes = []any{
	(*ObserveRequest)(nil),     // 0: loop.internal.pb.mercury.v4.ObserveRequest
	(*Block)(nil),              // 1: loop.internal.pb.mercury.v4.Block
	(*Observation)(nil),        // 2: loop.internal.pb.mercury.v4.Observation
	(*ObserveResponse)(nil),    // 3: loop.internal.pb.mercury.v4.ObserveResponse
	(*pb.ReportTimestamp)(nil), // 4: loop.ReportTimestamp
	(*pb.BigInt)(nil),          // 5: loop.BigInt
}
var file_datasource_v4_proto_depIdxs = []int32{
	4, // 0: loop.internal.pb.mercury.v4.ObserveRequest.report_timestamp:type_name -> loop.ReportTimestamp
	5, // 1: loop.internal.pb.mercury.v4.Observation.benchmarkPrice:type_name -> loop.BigInt
	5, // 2: loop.internal.pb.mercury.v4.Observation.bid:type_name -> loop.BigInt
	5, // 3: loop.internal.pb.mercury.v4.Observation.ask:type_name -> loop.BigInt
	5, // 4: loop.internal.pb.mercury.v4.Observation.linkPrice:type_name -> loop.BigInt
	5, // 5: loop.internal.pb.mercury.v4.Observation.nativePrice:type_name -> loop.BigInt
	2, // 6: loop.internal.pb.mercury.v4.ObserveResponse.observation:type_name -> loop.internal.pb.mercury.v4.Observation
	0, // 7: loop.internal.pb.mercury.v4.DataSource.Observe:input_type -> loop.internal.pb.mercury.v4.ObserveRequest
	3, // 8: loop.internal.pb.mercury.v4.DataSource.Observe:output_type -> loop.internal.pb.mercury.v4.ObserveResponse
	8, // [8:9] is the sub-list for method output_type
	7, // [7:8] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_datasource_v4_proto_init() }
func file_datasource_v4_proto_init() {
	if File_datasource_v4_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_datasource_v4_proto_rawDesc), len(file_datasource_v4_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_datasource_v4_proto_goTypes,
		DependencyIndexes: file_datasource_v4_proto_depIdxs,
		MessageInfos:      file_datasource_v4_proto_msgTypes,
	}.Build()
	File_datasource_v4_proto = out.File
	file_datasource_v4_proto_goTypes = nil
	file_datasource_v4_proto_depIdxs = nil
}
