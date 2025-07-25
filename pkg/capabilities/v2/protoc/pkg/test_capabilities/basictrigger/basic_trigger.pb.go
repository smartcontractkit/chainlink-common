// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: capabilities/internal/basictrigger/v1/basic_trigger.proto

package basictrigger

import (
	_ "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
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

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Number        int32                  `protobuf:"varint,2,opt,name=number,proto3" json:"number,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_capabilities_internal_basictrigger_v1_basic_trigger_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_capabilities_internal_basictrigger_v1_basic_trigger_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Config) GetNumber() int32 {
	if x != nil {
		return x.Number
	}
	return 0
}

type Outputs struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	CoolOutput    string                 `protobuf:"bytes,1,opt,name=cool_output,json=coolOutput,proto3" json:"cool_output,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Outputs) Reset() {
	*x = Outputs{}
	mi := &file_capabilities_internal_basictrigger_v1_basic_trigger_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Outputs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Outputs) ProtoMessage() {}

func (x *Outputs) ProtoReflect() protoreflect.Message {
	mi := &file_capabilities_internal_basictrigger_v1_basic_trigger_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Outputs.ProtoReflect.Descriptor instead.
func (*Outputs) Descriptor() ([]byte, []int) {
	return file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDescGZIP(), []int{1}
}

func (x *Outputs) GetCoolOutput() string {
	if x != nil {
		return x.CoolOutput
	}
	return ""
}

var File_capabilities_internal_basictrigger_v1_basic_trigger_proto protoreflect.FileDescriptor

const file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDesc = "" +
	"\n" +
	"9capabilities/internal/basictrigger/v1/basic_trigger.proto\x12%capabilities.internal.basictrigger.v1\x1a*tools/generator/v1alpha/cre_metadata.proto\"4\n" +
	"\x06Config\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x16\n" +
	"\x06number\x18\x02 \x01(\x05R\x06number\"*\n" +
	"\aOutputs\x12\x1f\n" +
	"\vcool_output\x18\x01 \x01(\tR\n" +
	"coolOutput2\x95\x01\n" +
	"\x05Basic\x12j\n" +
	"\aTrigger\x12-.capabilities.internal.basictrigger.v1.Config\x1a..capabilities.internal.basictrigger.v1.Outputs0\x01\x1a \x82\xb5\x18\x1c\b\x01\x12\x18basic-test-trigger@1.0.0b\x06proto3"

var (
	file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDescOnce sync.Once
	file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDescData []byte
)

func file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDescGZIP() []byte {
	file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDescOnce.Do(func() {
		file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDesc), len(file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDesc)))
	})
	return file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDescData
}

var file_capabilities_internal_basictrigger_v1_basic_trigger_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_capabilities_internal_basictrigger_v1_basic_trigger_proto_goTypes = []any{
	(*Config)(nil),  // 0: capabilities.internal.basictrigger.v1.Config
	(*Outputs)(nil), // 1: capabilities.internal.basictrigger.v1.Outputs
}
var file_capabilities_internal_basictrigger_v1_basic_trigger_proto_depIdxs = []int32{
	0, // 0: capabilities.internal.basictrigger.v1.Basic.Trigger:input_type -> capabilities.internal.basictrigger.v1.Config
	1, // 1: capabilities.internal.basictrigger.v1.Basic.Trigger:output_type -> capabilities.internal.basictrigger.v1.Outputs
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_capabilities_internal_basictrigger_v1_basic_trigger_proto_init() }
func file_capabilities_internal_basictrigger_v1_basic_trigger_proto_init() {
	if File_capabilities_internal_basictrigger_v1_basic_trigger_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDesc), len(file_capabilities_internal_basictrigger_v1_basic_trigger_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_capabilities_internal_basictrigger_v1_basic_trigger_proto_goTypes,
		DependencyIndexes: file_capabilities_internal_basictrigger_v1_basic_trigger_proto_depIdxs,
		MessageInfos:      file_capabilities_internal_basictrigger_v1_basic_trigger_proto_msgTypes,
	}.Build()
	File_capabilities_internal_basictrigger_v1_basic_trigger_proto = out.File
	file_capabilities_internal_basictrigger_v1_basic_trigger_proto_goTypes = nil
	file_capabilities_internal_basictrigger_v1_basic_trigger_proto_depIdxs = nil
}
