// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: values/pb/values.proto

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

type Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Value:
	//
	//	*Value_StringValue
	//	*Value_BoolValue
	//	*Value_BytesValue
	//	*Value_MapValue
	//	*Value_ListValue
	//	*Value_DecimalValue
	//	*Value_Int64Value
	//	*Value_BigintValue
	Value isValue_Value `protobuf_oneof:"value"`
}

func (x *Value) Reset() {
	*x = Value{}
	if protoimpl.UnsafeEnabled {
		mi := &file_values_pb_values_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Value) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Value) ProtoMessage() {}

func (x *Value) ProtoReflect() protoreflect.Message {
	mi := &file_values_pb_values_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Value.ProtoReflect.Descriptor instead.
func (*Value) Descriptor() ([]byte, []int) {
	return file_values_pb_values_proto_rawDescGZIP(), []int{0}
}

func (m *Value) GetValue() isValue_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *Value) GetStringValue() string {
	if x, ok := x.GetValue().(*Value_StringValue); ok {
		return x.StringValue
	}
	return ""
}

func (x *Value) GetBoolValue() bool {
	if x, ok := x.GetValue().(*Value_BoolValue); ok {
		return x.BoolValue
	}
	return false
}

func (x *Value) GetBytesValue() []byte {
	if x, ok := x.GetValue().(*Value_BytesValue); ok {
		return x.BytesValue
	}
	return nil
}

func (x *Value) GetMapValue() *Map {
	if x, ok := x.GetValue().(*Value_MapValue); ok {
		return x.MapValue
	}
	return nil
}

func (x *Value) GetListValue() *List {
	if x, ok := x.GetValue().(*Value_ListValue); ok {
		return x.ListValue
	}
	return nil
}

func (x *Value) GetDecimalValue() string {
	if x, ok := x.GetValue().(*Value_DecimalValue); ok {
		return x.DecimalValue
	}
	return ""
}

func (x *Value) GetInt64Value() int64 {
	if x, ok := x.GetValue().(*Value_Int64Value); ok {
		return x.Int64Value
	}
	return 0
}

func (x *Value) GetBigintValue() []byte {
	if x, ok := x.GetValue().(*Value_BigintValue); ok {
		return x.BigintValue
	}
	return nil
}

type isValue_Value interface {
	isValue_Value()
}

type Value_StringValue struct {
	StringValue string `protobuf:"bytes,1,opt,name=string_value,json=stringValue,proto3,oneof"`
}

type Value_BoolValue struct {
	BoolValue bool `protobuf:"varint,2,opt,name=bool_value,json=boolValue,proto3,oneof"`
}

type Value_BytesValue struct {
	BytesValue []byte `protobuf:"bytes,3,opt,name=bytes_value,json=bytesValue,proto3,oneof"`
}

type Value_MapValue struct {
	MapValue *Map `protobuf:"bytes,4,opt,name=map_value,json=mapValue,proto3,oneof"`
}

type Value_ListValue struct {
	ListValue *List `protobuf:"bytes,5,opt,name=list_value,json=listValue,proto3,oneof"`
}

type Value_DecimalValue struct {
	DecimalValue string `protobuf:"bytes,6,opt,name=decimal_value,json=decimalValue,proto3,oneof"`
}

type Value_Int64Value struct {
	Int64Value int64 `protobuf:"varint,7,opt,name=int64_value,json=int64Value,proto3,oneof"`
}

type Value_BigintValue struct {
	BigintValue []byte `protobuf:"bytes,8,opt,name=bigint_value,json=bigintValue,proto3,oneof"`
}

func (*Value_StringValue) isValue_Value() {}

func (*Value_BoolValue) isValue_Value() {}

func (*Value_BytesValue) isValue_Value() {}

func (*Value_MapValue) isValue_Value() {}

func (*Value_ListValue) isValue_Value() {}

func (*Value_DecimalValue) isValue_Value() {}

func (*Value_Int64Value) isValue_Value() {}

func (*Value_BigintValue) isValue_Value() {}

type Map struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fields map[string]*Value `protobuf:"bytes,1,rep,name=fields,proto3" json:"fields,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Map) Reset() {
	*x = Map{}
	if protoimpl.UnsafeEnabled {
		mi := &file_values_pb_values_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Map) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Map) ProtoMessage() {}

func (x *Map) ProtoReflect() protoreflect.Message {
	mi := &file_values_pb_values_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Map.ProtoReflect.Descriptor instead.
func (*Map) Descriptor() ([]byte, []int) {
	return file_values_pb_values_proto_rawDescGZIP(), []int{1}
}

func (x *Map) GetFields() map[string]*Value {
	if x != nil {
		return x.Fields
	}
	return nil
}

type List struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fields []*Value `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty"`
}

func (x *List) Reset() {
	*x = List{}
	if protoimpl.UnsafeEnabled {
		mi := &file_values_pb_values_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *List) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*List) ProtoMessage() {}

func (x *List) ProtoReflect() protoreflect.Message {
	mi := &file_values_pb_values_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use List.ProtoReflect.Descriptor instead.
func (*List) Descriptor() ([]byte, []int) {
	return file_values_pb_values_proto_rawDescGZIP(), []int{2}
}

func (x *List) GetFields() []*Value {
	if x != nil {
		return x.Fields
	}
	return nil
}

var File_values_pb_values_proto protoreflect.FileDescriptor

var file_values_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0xc3, 0x02, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x12, 0x23, 0x0a, 0x0c, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0b, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x1f, 0x0a, 0x0a, 0x62, 0x6f, 0x6f, 0x6c, 0x5f, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x48, 0x00, 0x52, 0x09, 0x62, 0x6f, 0x6f,
	0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x21, 0x0a, 0x0b, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x0a, 0x62,
	0x79, 0x74, 0x65, 0x73, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x2a, 0x0a, 0x09, 0x6d, 0x61, 0x70,
	0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x73, 0x2e, 0x4d, 0x61, 0x70, 0x48, 0x00, 0x52, 0x08, 0x6d, 0x61, 0x70,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x2d, 0x0a, 0x0a, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x48, 0x00, 0x52, 0x09, 0x6c, 0x69, 0x73, 0x74, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x12, 0x25, 0x0a, 0x0d, 0x64, 0x65, 0x63, 0x69, 0x6d, 0x61, 0x6c, 0x5f,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0c, 0x64,
	0x65, 0x63, 0x69, 0x6d, 0x61, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x21, 0x0a, 0x0b, 0x69,
	0x6e, 0x74, 0x36, 0x34, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03,
	0x48, 0x00, 0x52, 0x0a, 0x69, 0x6e, 0x74, 0x36, 0x34, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x23,
	0x0a, 0x0c, 0x62, 0x69, 0x67, 0x69, 0x6e, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x0b, 0x62, 0x69, 0x67, 0x69, 0x6e, 0x74, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x42, 0x07, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x80, 0x01, 0x0a,
	0x03, 0x4d, 0x61, 0x70, 0x12, 0x2f, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2e, 0x4d, 0x61,
	0x70, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x66,
	0x69, 0x65, 0x6c, 0x64, 0x73, 0x1a, 0x48, 0x0a, 0x0b, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x23, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2e, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0x2d, 0x0a, 0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x25, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73,
	0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x42, 0x3c,
	0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6d, 0x61,
	0x72, 0x74, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6b, 0x69, 0x74, 0x2f, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_values_pb_values_proto_rawDescOnce sync.Once
	file_values_pb_values_proto_rawDescData = file_values_pb_values_proto_rawDesc
)

func file_values_pb_values_proto_rawDescGZIP() []byte {
	file_values_pb_values_proto_rawDescOnce.Do(func() {
		file_values_pb_values_proto_rawDescData = protoimpl.X.CompressGZIP(file_values_pb_values_proto_rawDescData)
	})
	return file_values_pb_values_proto_rawDescData
}

var file_values_pb_values_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_values_pb_values_proto_goTypes = []interface{}{
	(*Value)(nil), // 0: values.Value
	(*Map)(nil),   // 1: values.Map
	(*List)(nil),  // 2: values.List
	nil,           // 3: values.Map.FieldsEntry
}
var file_values_pb_values_proto_depIdxs = []int32{
	1, // 0: values.Value.map_value:type_name -> values.Map
	2, // 1: values.Value.list_value:type_name -> values.List
	3, // 2: values.Map.fields:type_name -> values.Map.FieldsEntry
	0, // 3: values.List.fields:type_name -> values.Value
	0, // 4: values.Map.FieldsEntry.value:type_name -> values.Value
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_values_pb_values_proto_init() }
func file_values_pb_values_proto_init() {
	if File_values_pb_values_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_values_pb_values_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Value); i {
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
		file_values_pb_values_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Map); i {
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
		file_values_pb_values_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*List); i {
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
	file_values_pb_values_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Value_StringValue)(nil),
		(*Value_BoolValue)(nil),
		(*Value_BytesValue)(nil),
		(*Value_MapValue)(nil),
		(*Value_ListValue)(nil),
		(*Value_DecimalValue)(nil),
		(*Value_Int64Value)(nil),
		(*Value_BigintValue)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_values_pb_values_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_values_pb_values_proto_goTypes,
		DependencyIndexes: file_values_pb_values_proto_depIdxs,
		MessageInfos:      file_values_pb_values_proto_msgTypes,
	}.Build()
	File_values_pb_values_proto = out.File
	file_values_pb_values_proto_rawDesc = nil
	file_values_pb_values_proto_goTypes = nil
	file_values_pb_values_proto_depIdxs = nil
}
