// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: pkg/capabilities/consensus/ocr3/types/ocr3_types.proto

package types

import (
	pb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
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

// per-workflow aggregation outcome
type AggregationOutcome struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EncodableOutcome *pb.Value `protobuf:"bytes,1,opt,name=encodableOutcome,proto3" json:"encodableOutcome,omitempty"` // passed to encoders
	Metadata         []byte    `protobuf:"bytes,2,opt,name=metadata,proto3" json:"metadata,omitempty"`                 // internal to the aggregator
	ShouldReport     bool      `protobuf:"varint,3,opt,name=shouldReport,proto3" json:"shouldReport,omitempty"`
}

func (x *AggregationOutcome) Reset() {
	*x = AggregationOutcome{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AggregationOutcome) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AggregationOutcome) ProtoMessage() {}

func (x *AggregationOutcome) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AggregationOutcome.ProtoReflect.Descriptor instead.
func (*AggregationOutcome) Descriptor() ([]byte, []int) {
	return file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDescGZIP(), []int{0}
}

func (x *AggregationOutcome) GetEncodableOutcome() *pb.Value {
	if x != nil {
		return x.EncodableOutcome
	}
	return nil
}

func (x *AggregationOutcome) GetMetadata() []byte {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *AggregationOutcome) GetShouldReport() bool {
	if x != nil {
		return x.ShouldReport
	}
	return false
}

type Query struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// the requests to get consensus on.
	RequestIds []string `protobuf:"bytes,1,rep,name=requestIds,proto3" json:"requestIds,omitempty"`
}

func (x *Query) Reset() {
	*x = Query{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Query) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Query) ProtoMessage() {}

func (x *Query) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Query.ProtoReflect.Descriptor instead.
func (*Query) Descriptor() ([]byte, []int) {
	return file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDescGZIP(), []int{1}
}

func (x *Query) GetRequestIds() []string {
	if x != nil {
		return x.RequestIds
	}
	return nil
}

var File_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto protoreflect.FileDescriptor

var file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDesc = []byte{
	0x0a, 0x36, 0x70, 0x6b, 0x67, 0x2f, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69,
	0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x2f, 0x6f, 0x63, 0x72,
	0x33, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x6f, 0x63, 0x72, 0x33, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6f, 0x63, 0x72, 0x33, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x1a, 0x1a, 0x70, 0x6b, 0x67, 0x2f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73,
	0x2f, 0x70, 0x62, 0x2f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x8f, 0x01, 0x0a, 0x12, 0x41, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x4f, 0x75, 0x74, 0x63, 0x6f, 0x6d, 0x65, 0x12, 0x39, 0x0a, 0x10, 0x65, 0x6e, 0x63, 0x6f, 0x64,
	0x61, 0x62, 0x6c, 0x65, 0x4f, 0x75, 0x74, 0x63, 0x6f, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0d, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x52, 0x10, 0x65, 0x6e, 0x63, 0x6f, 0x64, 0x61, 0x62, 0x6c, 0x65, 0x4f, 0x75, 0x74, 0x63, 0x6f,
	0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x22,
	0x0a, 0x0c, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x52, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x22, 0x27, 0x0a, 0x05, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x1e, 0x0a, 0x0a, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x73, 0x42, 0x23, 0x5a, 0x21, 0x63,
	0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x73,
	0x65, 0x6e, 0x73, 0x75, 0x73, 0x2f, 0x6f, 0x63, 0x72, 0x33, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDescOnce sync.Once
	file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDescData = file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDesc
)

func file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDescGZIP() []byte {
	file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDescOnce.Do(func() {
		file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDescData)
	})
	return file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDescData
}

var file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_goTypes = []interface{}{
	(*AggregationOutcome)(nil), // 0: ocr3_types.AggregationOutcome
	(*Query)(nil),              // 1: ocr3_types.Query
	(*pb.Value)(nil),           // 2: values.Value
}
var file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_depIdxs = []int32{
	2, // 0: ocr3_types.AggregationOutcome.encodableOutcome:type_name -> values.Value
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_init() }
func file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_init() {
	if File_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AggregationOutcome); i {
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
		file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Query); i {
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
			RawDescriptor: file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_goTypes,
		DependencyIndexes: file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_depIdxs,
		MessageInfos:      file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_msgTypes,
	}.Build()
	File_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto = out.File
	file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_rawDesc = nil
	file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_goTypes = nil
	file_pkg_capabilities_consensus_ocr3_types_ocr3_types_proto_depIdxs = nil
}
