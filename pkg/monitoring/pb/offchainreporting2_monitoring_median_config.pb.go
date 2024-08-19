// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: offchainreporting2_monitoring_median_config.proto

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

type NumericalMedianConfigProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AlphaReportInfinite bool   `protobuf:"varint,1,opt,name=alpha_report_infinite,json=alphaReportInfinite,proto3" json:"alpha_report_infinite,omitempty"`
	AlphaReportPpb      uint64 `protobuf:"varint,2,opt,name=alpha_report_ppb,json=alphaReportPpb,proto3" json:"alpha_report_ppb,omitempty"`
	AlphaAcceptInfinite bool   `protobuf:"varint,3,opt,name=alpha_accept_infinite,json=alphaAcceptInfinite,proto3" json:"alpha_accept_infinite,omitempty"`
	AlphaAcceptPpb      uint64 `protobuf:"varint,4,opt,name=alpha_accept_ppb,json=alphaAcceptPpb,proto3" json:"alpha_accept_ppb,omitempty"`
	DeltaCNanoseconds   uint64 `protobuf:"varint,5,opt,name=delta_c_nanoseconds,json=deltaCNanoseconds,proto3" json:"delta_c_nanoseconds,omitempty"`
}

func (x *NumericalMedianConfigProto) Reset() {
	*x = NumericalMedianConfigProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_offchainreporting2_monitoring_median_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NumericalMedianConfigProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NumericalMedianConfigProto) ProtoMessage() {}

func (x *NumericalMedianConfigProto) ProtoReflect() protoreflect.Message {
	mi := &file_offchainreporting2_monitoring_median_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NumericalMedianConfigProto.ProtoReflect.Descriptor instead.
func (*NumericalMedianConfigProto) Descriptor() ([]byte, []int) {
	return file_offchainreporting2_monitoring_median_config_proto_rawDescGZIP(), []int{0}
}

func (x *NumericalMedianConfigProto) GetAlphaReportInfinite() bool {
	if x != nil {
		return x.AlphaReportInfinite
	}
	return false
}

func (x *NumericalMedianConfigProto) GetAlphaReportPpb() uint64 {
	if x != nil {
		return x.AlphaReportPpb
	}
	return 0
}

func (x *NumericalMedianConfigProto) GetAlphaAcceptInfinite() bool {
	if x != nil {
		return x.AlphaAcceptInfinite
	}
	return false
}

func (x *NumericalMedianConfigProto) GetAlphaAcceptPpb() uint64 {
	if x != nil {
		return x.AlphaAcceptPpb
	}
	return 0
}

func (x *NumericalMedianConfigProto) GetDeltaCNanoseconds() uint64 {
	if x != nil {
		return x.DeltaCNanoseconds
	}
	return 0
}

var File_offchainreporting2_monitoring_median_config_proto protoreflect.FileDescriptor

var file_offchainreporting2_monitoring_median_config_proto_rawDesc = []byte{
	0x0a, 0x31, 0x6f, 0x66, 0x66, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x69, 0x6e, 0x67, 0x32, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x5f,
	0x6d, 0x65, 0x64, 0x69, 0x61, 0x6e, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x22, 0x88, 0x02, 0x0a, 0x1a, 0x4e, 0x75, 0x6d, 0x65,
	0x72, 0x69, 0x63, 0x61, 0x6c, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x32, 0x0a, 0x15, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x5f,
	0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x69, 0x6e, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x13, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x52, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x49, 0x6e, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x65, 0x12, 0x28, 0x0a, 0x10, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x70, 0x70, 0x62, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x0e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x52, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x50, 0x70, 0x62, 0x12, 0x32, 0x0a, 0x15, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x5f, 0x61, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x5f, 0x69, 0x6e, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x13, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74,
	0x49, 0x6e, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x65, 0x12, 0x28, 0x0a, 0x10, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x5f, 0x70, 0x70, 0x62, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x0e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x50,
	0x70, 0x62, 0x12, 0x2e, 0x0a, 0x13, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x5f, 0x63, 0x5f, 0x6e, 0x61,
	0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x11, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x43, 0x4e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e,
	0x64, 0x73, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_offchainreporting2_monitoring_median_config_proto_rawDescOnce sync.Once
	file_offchainreporting2_monitoring_median_config_proto_rawDescData = file_offchainreporting2_monitoring_median_config_proto_rawDesc
)

func file_offchainreporting2_monitoring_median_config_proto_rawDescGZIP() []byte {
	file_offchainreporting2_monitoring_median_config_proto_rawDescOnce.Do(func() {
		file_offchainreporting2_monitoring_median_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_offchainreporting2_monitoring_median_config_proto_rawDescData)
	})
	return file_offchainreporting2_monitoring_median_config_proto_rawDescData
}

var file_offchainreporting2_monitoring_median_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_offchainreporting2_monitoring_median_config_proto_goTypes = []interface{}{
	(*NumericalMedianConfigProto)(nil), // 0: pb.NumericalMedianConfigProto
}
var file_offchainreporting2_monitoring_median_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_offchainreporting2_monitoring_median_config_proto_init() }
func file_offchainreporting2_monitoring_median_config_proto_init() {
	if File_offchainreporting2_monitoring_median_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_offchainreporting2_monitoring_median_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NumericalMedianConfigProto); i {
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
			RawDescriptor: file_offchainreporting2_monitoring_median_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_offchainreporting2_monitoring_median_config_proto_goTypes,
		DependencyIndexes: file_offchainreporting2_monitoring_median_config_proto_depIdxs,
		MessageInfos:      file_offchainreporting2_monitoring_median_config_proto_msgTypes,
	}.Build()
	File_offchainreporting2_monitoring_median_config_proto = out.File
	file_offchainreporting2_monitoring_median_config_proto_rawDesc = nil
	file_offchainreporting2_monitoring_median_config_proto_goTypes = nil
	file_offchainreporting2_monitoring_median_config_proto_depIdxs = nil
}
