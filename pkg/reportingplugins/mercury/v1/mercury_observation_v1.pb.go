// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.2
// source: mercury_observation_v1.proto

package mercury_v1

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// TODO: what about different report formats for different clients?
type MercuryObservationProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp uint32 `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// Prices
	BenchmarkPrice []byte `protobuf:"bytes,2,opt,name=benchmarkPrice,proto3" json:"benchmarkPrice,omitempty"`
	Bid            []byte `protobuf:"bytes,3,opt,name=bid,proto3" json:"bid,omitempty"`
	Ask            []byte `protobuf:"bytes,4,opt,name=ask,proto3" json:"ask,omitempty"`
	// All three prices must be valid, or none are (they all should come from one API query and hold invariant bid <= bm <= ask)
	PricesValid bool `protobuf:"varint,5,opt,name=pricesValid,proto3" json:"pricesValid,omitempty"`
	// Current block
	CurrentBlockNum       int64  `protobuf:"varint,6,opt,name=currentBlockNum,proto3" json:"currentBlockNum,omitempty"`
	CurrentBlockHash      []byte `protobuf:"bytes,7,opt,name=currentBlockHash,proto3" json:"currentBlockHash,omitempty"`
	CurrentBlockTimestamp uint64 `protobuf:"varint,8,opt,name=currentBlockTimestamp,proto3" json:"currentBlockTimestamp,omitempty"`
	// All three block observations must be valid, or none are (they all come from the same block)
	CurrentBlockValid bool `protobuf:"varint,9,opt,name=currentBlockValid,proto3" json:"currentBlockValid,omitempty"`
	// MaxFinalizedBlockNumber comes from previous report when present and is
	// only observed from mercury server when previous report is nil
	MaxFinalizedBlockNumber      int64 `protobuf:"varint,10,opt,name=maxFinalizedBlockNumber,proto3" json:"maxFinalizedBlockNumber,omitempty"`
	MaxFinalizedBlockNumberValid bool  `protobuf:"varint,11,opt,name=maxFinalizedBlockNumberValid,proto3" json:"maxFinalizedBlockNumberValid,omitempty"`
}

func (x *MercuryObservationProto) Reset() {
	*x = MercuryObservationProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_observation_v1_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MercuryObservationProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MercuryObservationProto) ProtoMessage() {}

func (x *MercuryObservationProto) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_observation_v1_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MercuryObservationProto.ProtoReflect.Descriptor instead.
func (*MercuryObservationProto) Descriptor() ([]byte, []int) {
	return file_mercury_observation_v1_proto_rawDescGZIP(), []int{0}
}

func (x *MercuryObservationProto) GetTimestamp() uint32 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *MercuryObservationProto) GetBenchmarkPrice() []byte {
	if x != nil {
		return x.BenchmarkPrice
	}
	return nil
}

func (x *MercuryObservationProto) GetBid() []byte {
	if x != nil {
		return x.Bid
	}
	return nil
}

func (x *MercuryObservationProto) GetAsk() []byte {
	if x != nil {
		return x.Ask
	}
	return nil
}

func (x *MercuryObservationProto) GetPricesValid() bool {
	if x != nil {
		return x.PricesValid
	}
	return false
}

func (x *MercuryObservationProto) GetCurrentBlockNum() int64 {
	if x != nil {
		return x.CurrentBlockNum
	}
	return 0
}

func (x *MercuryObservationProto) GetCurrentBlockHash() []byte {
	if x != nil {
		return x.CurrentBlockHash
	}
	return nil
}

func (x *MercuryObservationProto) GetCurrentBlockTimestamp() uint64 {
	if x != nil {
		return x.CurrentBlockTimestamp
	}
	return 0
}

func (x *MercuryObservationProto) GetCurrentBlockValid() bool {
	if x != nil {
		return x.CurrentBlockValid
	}
	return false
}

func (x *MercuryObservationProto) GetMaxFinalizedBlockNumber() int64 {
	if x != nil {
		return x.MaxFinalizedBlockNumber
	}
	return 0
}

func (x *MercuryObservationProto) GetMaxFinalizedBlockNumberValid() bool {
	if x != nil {
		return x.MaxFinalizedBlockNumberValid
	}
	return false
}

var File_mercury_observation_v1_proto protoreflect.FileDescriptor

var file_mercury_observation_v1_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x5f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x76, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a,
	0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x5f, 0x76, 0x31, 0x22, 0xdd, 0x03, 0x0a, 0x17, 0x4d,
	0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x12, 0x26, 0x0a, 0x0e, 0x62, 0x65, 0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72,
	0x6b, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x62, 0x65,
	0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72, 0x6b, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x62, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x62, 0x69, 0x64, 0x12, 0x10,
	0x0a, 0x03, 0x61, 0x73, 0x6b, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x61, 0x73, 0x6b,
	0x12, 0x20, 0x0a, 0x0b, 0x70, 0x72, 0x69, 0x63, 0x65, 0x73, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x70, 0x72, 0x69, 0x63, 0x65, 0x73, 0x56, 0x61, 0x6c,
	0x69, 0x64, 0x12, 0x28, 0x0a, 0x0f, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x63, 0x75, 0x72,
	0x72, 0x65, 0x6e, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x12, 0x2a, 0x0a, 0x10,
	0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x10, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x12, 0x34, 0x0a, 0x15, 0x63, 0x75, 0x72, 0x72,
	0x65, 0x6e, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x18, 0x08, 0x20, 0x01, 0x28, 0x04, 0x52, 0x15, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x2c,
	0x0a, 0x11, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x56, 0x61,
	0x6c, 0x69, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x11, 0x63, 0x75, 0x72, 0x72, 0x65,
	0x6e, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x12, 0x38, 0x0a, 0x17,
	0x6d, 0x61, 0x78, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x17, 0x6d,
	0x61, 0x78, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x42, 0x0a, 0x1c, 0x6d, 0x61, 0x78, 0x46, 0x69, 0x6e,
	0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65,
	0x72, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x08, 0x52, 0x1c, 0x6d, 0x61,
	0x78, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x42, 0x0e, 0x5a, 0x0c, 0x2e, 0x3b,
	0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x5f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_mercury_observation_v1_proto_rawDescOnce sync.Once
	file_mercury_observation_v1_proto_rawDescData = file_mercury_observation_v1_proto_rawDesc
)

func file_mercury_observation_v1_proto_rawDescGZIP() []byte {
	file_mercury_observation_v1_proto_rawDescOnce.Do(func() {
		file_mercury_observation_v1_proto_rawDescData = protoimpl.X.CompressGZIP(file_mercury_observation_v1_proto_rawDescData)
	})
	return file_mercury_observation_v1_proto_rawDescData
}

var file_mercury_observation_v1_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_mercury_observation_v1_proto_goTypes = []interface{}{
	(*MercuryObservationProto)(nil), // 0: mercury_v1.MercuryObservationProto
}
var file_mercury_observation_v1_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_mercury_observation_v1_proto_init() }
func file_mercury_observation_v1_proto_init() {
	if File_mercury_observation_v1_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mercury_observation_v1_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MercuryObservationProto); i {
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
			RawDescriptor: file_mercury_observation_v1_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mercury_observation_v1_proto_goTypes,
		DependencyIndexes: file_mercury_observation_v1_proto_depIdxs,
		MessageInfos:      file_mercury_observation_v1_proto_msgTypes,
	}.Build()
	File_mercury_observation_v1_proto = out.File
	file_mercury_observation_v1_proto_rawDesc = nil
	file_mercury_observation_v1_proto_goTypes = nil
	file_mercury_observation_v1_proto_depIdxs = nil
}
