// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.2
// source: reporting_plugin_service.proto

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

type ReportingPluginServiceConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProviderType  string `protobuf:"bytes,1,opt,name=providerType,proto3" json:"providerType,omitempty"`
	Command       string `protobuf:"bytes,2,opt,name=command,proto3" json:"command,omitempty"`
	PluginName    string `protobuf:"bytes,3,opt,name=pluginName,proto3" json:"pluginName,omitempty"`
	PluginConfig  string `protobuf:"bytes,4,opt,name=pluginConfig,proto3" json:"pluginConfig,omitempty"`
	TelemetryType string `protobuf:"bytes,5,opt,name=telemetryType,proto3" json:"telemetryType,omitempty"`
}

func (x *ReportingPluginServiceConfig) Reset() {
	*x = ReportingPluginServiceConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reporting_plugin_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReportingPluginServiceConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReportingPluginServiceConfig) ProtoMessage() {}

func (x *ReportingPluginServiceConfig) ProtoReflect() protoreflect.Message {
	mi := &file_reporting_plugin_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReportingPluginServiceConfig.ProtoReflect.Descriptor instead.
func (*ReportingPluginServiceConfig) Descriptor() ([]byte, []int) {
	return file_reporting_plugin_service_proto_rawDescGZIP(), []int{0}
}

func (x *ReportingPluginServiceConfig) GetProviderType() string {
	if x != nil {
		return x.ProviderType
	}
	return ""
}

func (x *ReportingPluginServiceConfig) GetCommand() string {
	if x != nil {
		return x.Command
	}
	return ""
}

func (x *ReportingPluginServiceConfig) GetPluginName() string {
	if x != nil {
		return x.PluginName
	}
	return ""
}

func (x *ReportingPluginServiceConfig) GetPluginConfig() string {
	if x != nil {
		return x.PluginConfig
	}
	return ""
}

func (x *ReportingPluginServiceConfig) GetTelemetryType() string {
	if x != nil {
		return x.TelemetryType
	}
	return ""
}

// NewReportingPluginFactoryRequest has arguments for [github.com/smartcontractkit/chainlink-common/pkg/loop/reporting_plugins/LOOPPService.NewReportingPluginFactory].
type NewReportingPluginFactoryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProviderID                   uint32                        `protobuf:"varint,1,opt,name=providerID,proto3" json:"providerID,omitempty"`
	ErrorLogID                   uint32                        `protobuf:"varint,2,opt,name=errorLogID,proto3" json:"errorLogID,omitempty"`
	PipelineRunnerID             uint32                        `protobuf:"varint,3,opt,name=pipelineRunnerID,proto3" json:"pipelineRunnerID,omitempty"`
	TelemetryID                  uint32                        `protobuf:"varint,4,opt,name=telemetryID,proto3" json:"telemetryID,omitempty"`
	ReportingPluginServiceConfig *ReportingPluginServiceConfig `protobuf:"bytes,5,opt,name=ReportingPluginServiceConfig,proto3" json:"ReportingPluginServiceConfig,omitempty"`
}

func (x *NewReportingPluginFactoryRequest) Reset() {
	*x = NewReportingPluginFactoryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reporting_plugin_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewReportingPluginFactoryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewReportingPluginFactoryRequest) ProtoMessage() {}

func (x *NewReportingPluginFactoryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_reporting_plugin_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewReportingPluginFactoryRequest.ProtoReflect.Descriptor instead.
func (*NewReportingPluginFactoryRequest) Descriptor() ([]byte, []int) {
	return file_reporting_plugin_service_proto_rawDescGZIP(), []int{1}
}

func (x *NewReportingPluginFactoryRequest) GetProviderID() uint32 {
	if x != nil {
		return x.ProviderID
	}
	return 0
}

func (x *NewReportingPluginFactoryRequest) GetErrorLogID() uint32 {
	if x != nil {
		return x.ErrorLogID
	}
	return 0
}

func (x *NewReportingPluginFactoryRequest) GetPipelineRunnerID() uint32 {
	if x != nil {
		return x.PipelineRunnerID
	}
	return 0
}

func (x *NewReportingPluginFactoryRequest) GetTelemetryID() uint32 {
	if x != nil {
		return x.TelemetryID
	}
	return 0
}

func (x *NewReportingPluginFactoryRequest) GetReportingPluginServiceConfig() *ReportingPluginServiceConfig {
	if x != nil {
		return x.ReportingPluginServiceConfig
	}
	return nil
}

// NewReportingPluginFactoryReply has return arguments for [github.com/smartcontractkit/chainlink-common/pkg/loop/reporting_plugins/LOOPPService.NewReportingPluginFactory].
type NewReportingPluginFactoryReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID uint32 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
}

func (x *NewReportingPluginFactoryReply) Reset() {
	*x = NewReportingPluginFactoryReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reporting_plugin_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewReportingPluginFactoryReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewReportingPluginFactoryReply) ProtoMessage() {}

func (x *NewReportingPluginFactoryReply) ProtoReflect() protoreflect.Message {
	mi := &file_reporting_plugin_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewReportingPluginFactoryReply.ProtoReflect.Descriptor instead.
func (*NewReportingPluginFactoryReply) Descriptor() ([]byte, []int) {
	return file_reporting_plugin_service_proto_rawDescGZIP(), []int{2}
}

func (x *NewReportingPluginFactoryReply) GetID() uint32 {
	if x != nil {
		return x.ID
	}
	return 0
}

var File_reporting_plugin_service_proto protoreflect.FileDescriptor

var file_reporting_plugin_service_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x04, 0x6c, 0x6f, 0x6f, 0x70, 0x22, 0xc6, 0x01, 0x0a, 0x1c, 0x52, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x69, 0x6e, 0x67, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x22, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70,
	0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x24, 0x0a, 0x0d, 0x74, 0x65, 0x6c,
	0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x54, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x54, 0x79, 0x70, 0x65, 0x22,
	0x98, 0x02, 0x0a, 0x20, 0x4e, 0x65, 0x77, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64,
	0x65, 0x72, 0x49, 0x44, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x4c, 0x6f, 0x67,
	0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x4c,
	0x6f, 0x67, 0x49, 0x44, 0x12, 0x2a, 0x0a, 0x10, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x10,
	0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x49, 0x44,
	0x12, 0x20, 0x0a, 0x0b, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x49, 0x44, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79,
	0x49, 0x44, 0x12, 0x66, 0x0a, 0x1c, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e,
	0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x1c, 0x52, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x30, 0x0a, 0x1e, 0x4e, 0x65,
	0x77, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x0e, 0x0a, 0x02,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x49, 0x44, 0x32, 0x85, 0x01, 0x0a,
	0x16, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x6b, 0x0a, 0x19, 0x4e, 0x65, 0x77, 0x52, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x46, 0x61, 0x63,
	0x74, 0x6f, 0x72, 0x79, 0x12, 0x26, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x4e, 0x65, 0x77, 0x52,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x46, 0x61,
	0x63, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x6c,
	0x6f, 0x6f, 0x70, 0x2e, 0x4e, 0x65, 0x77, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x22, 0x00, 0x42, 0x43, 0x5a, 0x41, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74,
	0x6b, 0x69, 0x74, 0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x2d, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6c, 0x6f, 0x6f, 0x70, 0x2f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_reporting_plugin_service_proto_rawDescOnce sync.Once
	file_reporting_plugin_service_proto_rawDescData = file_reporting_plugin_service_proto_rawDesc
)

func file_reporting_plugin_service_proto_rawDescGZIP() []byte {
	file_reporting_plugin_service_proto_rawDescOnce.Do(func() {
		file_reporting_plugin_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_reporting_plugin_service_proto_rawDescData)
	})
	return file_reporting_plugin_service_proto_rawDescData
}

var file_reporting_plugin_service_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_reporting_plugin_service_proto_goTypes = []interface{}{
	(*ReportingPluginServiceConfig)(nil),     // 0: loop.ReportingPluginServiceConfig
	(*NewReportingPluginFactoryRequest)(nil), // 1: loop.NewReportingPluginFactoryRequest
	(*NewReportingPluginFactoryReply)(nil),   // 2: loop.NewReportingPluginFactoryReply
}
var file_reporting_plugin_service_proto_depIdxs = []int32{
	0, // 0: loop.NewReportingPluginFactoryRequest.ReportingPluginServiceConfig:type_name -> loop.ReportingPluginServiceConfig
	1, // 1: loop.ReportingPluginService.NewReportingPluginFactory:input_type -> loop.NewReportingPluginFactoryRequest
	2, // 2: loop.ReportingPluginService.NewReportingPluginFactory:output_type -> loop.NewReportingPluginFactoryReply
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_reporting_plugin_service_proto_init() }
func file_reporting_plugin_service_proto_init() {
	if File_reporting_plugin_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_reporting_plugin_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReportingPluginServiceConfig); i {
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
		file_reporting_plugin_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewReportingPluginFactoryRequest); i {
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
		file_reporting_plugin_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewReportingPluginFactoryReply); i {
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
			RawDescriptor: file_reporting_plugin_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_reporting_plugin_service_proto_goTypes,
		DependencyIndexes: file_reporting_plugin_service_proto_depIdxs,
		MessageInfos:      file_reporting_plugin_service_proto_msgTypes,
	}.Build()
	File_reporting_plugin_service_proto = out.File
	file_reporting_plugin_service_proto_rawDesc = nil
	file_reporting_plugin_service_proto_goTypes = nil
	file_reporting_plugin_service_proto_depIdxs = nil
}
