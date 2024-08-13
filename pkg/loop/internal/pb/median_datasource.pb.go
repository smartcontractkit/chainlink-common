// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: median_datasource.proto

package pb

import (
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

// ObserveRequest has arguments for [github.com/smartcontractkit/chainlink-common/pkg/loop.DataSource.Observe].
type ObserveRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReportTimestamp *ReportTimestamp `protobuf:"bytes,1,opt,name=reportTimestamp,proto3" json:"reportTimestamp,omitempty"`
}

func (x *ObserveRequest) Reset() {
	*x = ObserveRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_median_datasource_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObserveRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObserveRequest) ProtoMessage() {}

func (x *ObserveRequest) ProtoReflect() protoreflect.Message {
	mi := &file_median_datasource_proto_msgTypes[0]
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
	return file_median_datasource_proto_rawDescGZIP(), []int{0}
}

func (x *ObserveRequest) GetReportTimestamp() *ReportTimestamp {
	if x != nil {
		return x.ReportTimestamp
	}
	return nil
}

// ObserveReply has return arguments for [github.com/smartcontractkit/chainlink-common/pkg/loop.DataSource.Observe].
type ObserveReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value *BigInt `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *ObserveReply) Reset() {
	*x = ObserveReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_median_datasource_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObserveReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObserveReply) ProtoMessage() {}

func (x *ObserveReply) ProtoReflect() protoreflect.Message {
	mi := &file_median_datasource_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObserveReply.ProtoReflect.Descriptor instead.
func (*ObserveReply) Descriptor() ([]byte, []int) {
	return file_median_datasource_proto_rawDescGZIP(), []int{1}
}

func (x *ObserveReply) GetValue() *BigInt {
	if x != nil {
		return x.Value
	}
	return nil
}

var File_median_datasource_proto protoreflect.FileDescriptor

var file_median_datasource_proto_rawDesc = []byte{
	0x0a, 0x17, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x6e, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x6c, 0x6f, 0x6f, 0x70, 0x1a,
	0x0d, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x51,
	0x0a, 0x0e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x3f, 0x0a, 0x0f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6c, 0x6f, 0x6f, 0x70,
	0x2e, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x0f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x22, 0x32, 0x0a, 0x0c, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x12, 0x22, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0c, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x42, 0x69, 0x67, 0x49, 0x6e, 0x74, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0x43, 0x0a, 0x0a, 0x44, 0x61, 0x74, 0x61, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x12, 0x35, 0x0a, 0x07, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x12, 0x14,
	0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x4f, 0x62, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x43, 0x5a, 0x41, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6b, 0x69, 0x74, 0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x6c,
	0x69, 0x6e, 0x6b, 0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6c,
	0x6f, 0x6f, 0x70, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_median_datasource_proto_rawDescOnce sync.Once
	file_median_datasource_proto_rawDescData = file_median_datasource_proto_rawDesc
)

func file_median_datasource_proto_rawDescGZIP() []byte {
	file_median_datasource_proto_rawDescOnce.Do(func() {
		file_median_datasource_proto_rawDescData = protoimpl.X.CompressGZIP(file_median_datasource_proto_rawDescData)
	})
	return file_median_datasource_proto_rawDescData
}

var file_median_datasource_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_median_datasource_proto_goTypes = []interface{}{
	(*ObserveRequest)(nil),  // 0: loop.ObserveRequest
	(*ObserveReply)(nil),    // 1: loop.ObserveReply
	(*ReportTimestamp)(nil), // 2: loop.ReportTimestamp
	(*BigInt)(nil),          // 3: loop.BigInt
}
var file_median_datasource_proto_depIdxs = []int32{
	2, // 0: loop.ObserveRequest.reportTimestamp:type_name -> loop.ReportTimestamp
	3, // 1: loop.ObserveReply.value:type_name -> loop.BigInt
	0, // 2: loop.DataSource.Observe:input_type -> loop.ObserveRequest
	1, // 3: loop.DataSource.Observe:output_type -> loop.ObserveReply
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_median_datasource_proto_init() }
func file_median_datasource_proto_init() {
	if File_median_datasource_proto != nil {
		return
	}
	file_relayer_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_median_datasource_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_median_datasource_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObserveReply); i {
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
			RawDescriptor: file_median_datasource_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_median_datasource_proto_goTypes,
		DependencyIndexes: file_median_datasource_proto_depIdxs,
		MessageInfos:      file_median_datasource_proto_msgTypes,
	}.Build()
	File_median_datasource_proto = out.File
	file_median_datasource_proto_rawDesc = nil
	file_median_datasource_proto_goTypes = nil
	file_median_datasource_proto_depIdxs = nil
}
