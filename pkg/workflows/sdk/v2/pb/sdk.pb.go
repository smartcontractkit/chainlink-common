// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: workflows/sdk/v2/pb/sdk.proto

package pb

import (
	pb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
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

type AggregationType int32

const (
	AggregationType_MEDIAN        AggregationType = 0
	AggregationType_IDENTICAL     AggregationType = 1
	AggregationType_COMMON_PREFIX AggregationType = 2
	AggregationType_COMMON_SUFFIX AggregationType = 3
)

// Enum value maps for AggregationType.
var (
	AggregationType_name = map[int32]string{
		0: "MEDIAN",
		1: "IDENTICAL",
		2: "COMMON_PREFIX",
		3: "COMMON_SUFFIX",
	}
	AggregationType_value = map[string]int32{
		"MEDIAN":        0,
		"IDENTICAL":     1,
		"COMMON_PREFIX": 2,
		"COMMON_SUFFIX": 3,
	}
)

func (x AggregationType) Enum() *AggregationType {
	p := new(AggregationType)
	*p = x
	return p
}

func (x AggregationType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AggregationType) Descriptor() protoreflect.EnumDescriptor {
	return file_workflows_sdk_v2_pb_sdk_proto_enumTypes[0].Descriptor()
}

func (AggregationType) Type() protoreflect.EnumType {
	return &file_workflows_sdk_v2_pb_sdk_proto_enumTypes[0]
}

func (x AggregationType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AggregationType.Descriptor instead.
func (AggregationType) EnumDescriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{0}
}

type Mode int32

const (
	Mode_DON  Mode = 0
	Mode_Node Mode = 1
)

// Enum value maps for Mode.
var (
	Mode_name = map[int32]string{
		0: "DON",
		1: "Node",
	}
	Mode_value = map[string]int32{
		"DON":  0,
		"Node": 1,
	}
)

func (x Mode) Enum() *Mode {
	p := new(Mode)
	*p = x
	return p
}

func (x Mode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Mode) Descriptor() protoreflect.EnumDescriptor {
	return file_workflows_sdk_v2_pb_sdk_proto_enumTypes[1].Descriptor()
}

func (Mode) Type() protoreflect.EnumType {
	return &file_workflows_sdk_v2_pb_sdk_proto_enumTypes[1]
}

func (x Mode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Mode.Descriptor instead.
func (Mode) EnumDescriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{1}
}

type SimpleConsensusInputs struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Observation:
	//
	//	*SimpleConsensusInputs_Value
	//	*SimpleConsensusInputs_Error
	Observation   isSimpleConsensusInputs_Observation `protobuf_oneof:"observation"`
	Descriptors   *ConsensusDescriptor                `protobuf:"bytes,3,opt,name=descriptors,proto3" json:"descriptors,omitempty"`
	Default       *pb.Value                           `protobuf:"bytes,4,opt,name=default,proto3" json:"default,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SimpleConsensusInputs) Reset() {
	*x = SimpleConsensusInputs{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SimpleConsensusInputs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimpleConsensusInputs) ProtoMessage() {}

func (x *SimpleConsensusInputs) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SimpleConsensusInputs.ProtoReflect.Descriptor instead.
func (*SimpleConsensusInputs) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{0}
}

func (x *SimpleConsensusInputs) GetObservation() isSimpleConsensusInputs_Observation {
	if x != nil {
		return x.Observation
	}
	return nil
}

func (x *SimpleConsensusInputs) GetValue() *pb.Value {
	if x != nil {
		if x, ok := x.Observation.(*SimpleConsensusInputs_Value); ok {
			return x.Value
		}
	}
	return nil
}

func (x *SimpleConsensusInputs) GetError() string {
	if x != nil {
		if x, ok := x.Observation.(*SimpleConsensusInputs_Error); ok {
			return x.Error
		}
	}
	return ""
}

func (x *SimpleConsensusInputs) GetDescriptors() *ConsensusDescriptor {
	if x != nil {
		return x.Descriptors
	}
	return nil
}

func (x *SimpleConsensusInputs) GetDefault() *pb.Value {
	if x != nil {
		return x.Default
	}
	return nil
}

type isSimpleConsensusInputs_Observation interface {
	isSimpleConsensusInputs_Observation()
}

type SimpleConsensusInputs_Value struct {
	Value *pb.Value `protobuf:"bytes,1,opt,name=value,proto3,oneof"`
}

type SimpleConsensusInputs_Error struct {
	Error string `protobuf:"bytes,2,opt,name=error,proto3,oneof"`
}

func (*SimpleConsensusInputs_Value) isSimpleConsensusInputs_Observation() {}

func (*SimpleConsensusInputs_Error) isSimpleConsensusInputs_Observation() {}

type FieldsMap struct {
	state         protoimpl.MessageState          `protogen:"open.v1"`
	Fields        map[string]*ConsensusDescriptor `protobuf:"bytes,1,rep,name=fields,proto3" json:"fields,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FieldsMap) Reset() {
	*x = FieldsMap{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FieldsMap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldsMap) ProtoMessage() {}

func (x *FieldsMap) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldsMap.ProtoReflect.Descriptor instead.
func (*FieldsMap) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{1}
}

func (x *FieldsMap) GetFields() map[string]*ConsensusDescriptor {
	if x != nil {
		return x.Fields
	}
	return nil
}

type ConsensusDescriptor struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Descriptor_:
	//
	//	*ConsensusDescriptor_Aggregation
	//	*ConsensusDescriptor_FieldsMap
	Descriptor_   isConsensusDescriptor_Descriptor_ `protobuf_oneof:"descriptor"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConsensusDescriptor) Reset() {
	*x = ConsensusDescriptor{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConsensusDescriptor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConsensusDescriptor) ProtoMessage() {}

func (x *ConsensusDescriptor) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConsensusDescriptor.ProtoReflect.Descriptor instead.
func (*ConsensusDescriptor) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{2}
}

func (x *ConsensusDescriptor) GetDescriptor_() isConsensusDescriptor_Descriptor_ {
	if x != nil {
		return x.Descriptor_
	}
	return nil
}

func (x *ConsensusDescriptor) GetAggregation() AggregationType {
	if x != nil {
		if x, ok := x.Descriptor_.(*ConsensusDescriptor_Aggregation); ok {
			return x.Aggregation
		}
	}
	return AggregationType_MEDIAN
}

func (x *ConsensusDescriptor) GetFieldsMap() *FieldsMap {
	if x != nil {
		if x, ok := x.Descriptor_.(*ConsensusDescriptor_FieldsMap); ok {
			return x.FieldsMap
		}
	}
	return nil
}

type isConsensusDescriptor_Descriptor_ interface {
	isConsensusDescriptor_Descriptor_()
}

type ConsensusDescriptor_Aggregation struct {
	Aggregation AggregationType `protobuf:"varint,1,opt,name=aggregation,proto3,enum=cre.sdk.v2.AggregationType,oneof"`
}

type ConsensusDescriptor_FieldsMap struct {
	FieldsMap *FieldsMap `protobuf:"bytes,2,opt,name=fieldsMap,proto3,oneof"`
}

func (*ConsensusDescriptor_Aggregation) isConsensusDescriptor_Descriptor_() {}

func (*ConsensusDescriptor_FieldsMap) isConsensusDescriptor_Descriptor_() {}

type CapabilityRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Payload       *anypb.Any             `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	Method        string                 `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	CallbackId    int32                  `protobuf:"varint,4,opt,name=callbackId,proto3" json:"callbackId,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CapabilityRequest) Reset() {
	*x = CapabilityRequest{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CapabilityRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CapabilityRequest) ProtoMessage() {}

func (x *CapabilityRequest) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CapabilityRequest.ProtoReflect.Descriptor instead.
func (*CapabilityRequest) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{3}
}

func (x *CapabilityRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *CapabilityRequest) GetPayload() *anypb.Any {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *CapabilityRequest) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *CapabilityRequest) GetCallbackId() int32 {
	if x != nil {
		return x.CallbackId
	}
	return 0
}

type CapabilityResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Response:
	//
	//	*CapabilityResponse_Payload
	//	*CapabilityResponse_Error
	Response      isCapabilityResponse_Response `protobuf_oneof:"response"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CapabilityResponse) Reset() {
	*x = CapabilityResponse{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CapabilityResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CapabilityResponse) ProtoMessage() {}

func (x *CapabilityResponse) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CapabilityResponse.ProtoReflect.Descriptor instead.
func (*CapabilityResponse) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{4}
}

func (x *CapabilityResponse) GetResponse() isCapabilityResponse_Response {
	if x != nil {
		return x.Response
	}
	return nil
}

func (x *CapabilityResponse) GetPayload() *anypb.Any {
	if x != nil {
		if x, ok := x.Response.(*CapabilityResponse_Payload); ok {
			return x.Payload
		}
	}
	return nil
}

func (x *CapabilityResponse) GetError() string {
	if x != nil {
		if x, ok := x.Response.(*CapabilityResponse_Error); ok {
			return x.Error
		}
	}
	return ""
}

type isCapabilityResponse_Response interface {
	isCapabilityResponse_Response()
}

type CapabilityResponse_Payload struct {
	Payload *anypb.Any `protobuf:"bytes,1,opt,name=payload,proto3,oneof"`
}

type CapabilityResponse_Error struct {
	Error string `protobuf:"bytes,2,opt,name=error,proto3,oneof"`
}

func (*CapabilityResponse_Payload) isCapabilityResponse_Response() {}

func (*CapabilityResponse_Error) isCapabilityResponse_Response() {}

type TriggerSubscription struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Payload       *anypb.Any             `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	Method        string                 `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TriggerSubscription) Reset() {
	*x = TriggerSubscription{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TriggerSubscription) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TriggerSubscription) ProtoMessage() {}

func (x *TriggerSubscription) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TriggerSubscription.ProtoReflect.Descriptor instead.
func (*TriggerSubscription) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{5}
}

func (x *TriggerSubscription) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TriggerSubscription) GetPayload() *anypb.Any {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *TriggerSubscription) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

type TriggerSubscriptionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Subscriptions []*TriggerSubscription `protobuf:"bytes,1,rep,name=subscriptions,proto3" json:"subscriptions,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TriggerSubscriptionRequest) Reset() {
	*x = TriggerSubscriptionRequest{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TriggerSubscriptionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TriggerSubscriptionRequest) ProtoMessage() {}

func (x *TriggerSubscriptionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TriggerSubscriptionRequest.ProtoReflect.Descriptor instead.
func (*TriggerSubscriptionRequest) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{6}
}

func (x *TriggerSubscriptionRequest) GetSubscriptions() []*TriggerSubscription {
	if x != nil {
		return x.Subscriptions
	}
	return nil
}

type Trigger struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            uint64                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Payload       *anypb.Any             `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Trigger) Reset() {
	*x = Trigger{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Trigger) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Trigger) ProtoMessage() {}

func (x *Trigger) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Trigger.ProtoReflect.Descriptor instead.
func (*Trigger) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{7}
}

func (x *Trigger) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Trigger) GetPayload() *anypb.Any {
	if x != nil {
		return x.Payload
	}
	return nil
}

type AwaitCapabilitiesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Ids           []int32                `protobuf:"varint,1,rep,packed,name=ids,proto3" json:"ids,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AwaitCapabilitiesRequest) Reset() {
	*x = AwaitCapabilitiesRequest{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AwaitCapabilitiesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AwaitCapabilitiesRequest) ProtoMessage() {}

func (x *AwaitCapabilitiesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AwaitCapabilitiesRequest.ProtoReflect.Descriptor instead.
func (*AwaitCapabilitiesRequest) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{8}
}

func (x *AwaitCapabilitiesRequest) GetIds() []int32 {
	if x != nil {
		return x.Ids
	}
	return nil
}

type AwaitCapabilitiesResponse struct {
	state         protoimpl.MessageState        `protogen:"open.v1"`
	Responses     map[int32]*CapabilityResponse `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AwaitCapabilitiesResponse) Reset() {
	*x = AwaitCapabilitiesResponse{}
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AwaitCapabilitiesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AwaitCapabilitiesResponse) ProtoMessage() {}

func (x *AwaitCapabilitiesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_workflows_sdk_v2_pb_sdk_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AwaitCapabilitiesResponse.ProtoReflect.Descriptor instead.
func (*AwaitCapabilitiesResponse) Descriptor() ([]byte, []int) {
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP(), []int{9}
}

func (x *AwaitCapabilitiesResponse) GetResponses() map[int32]*CapabilityResponse {
	if x != nil {
		return x.Responses
	}
	return nil
}

var File_workflows_sdk_v2_pb_sdk_proto protoreflect.FileDescriptor

const file_workflows_sdk_v2_pb_sdk_proto_rawDesc = "" +
	"\n" +
	"\x1dworkflows/sdk/v2/pb/sdk.proto\x12\n" +
	"cre.sdk.v2\x1a\x19google/protobuf/any.proto\x1a\x16values/pb/values.proto\"\xd1\x01\n" +
	"\x15SimpleConsensusInputs\x12%\n" +
	"\x05value\x18\x01 \x01(\v2\r.values.ValueH\x00R\x05value\x12\x16\n" +
	"\x05error\x18\x02 \x01(\tH\x00R\x05error\x12A\n" +
	"\vdescriptors\x18\x03 \x01(\v2\x1f.cre.sdk.v2.ConsensusDescriptorR\vdescriptors\x12'\n" +
	"\adefault\x18\x04 \x01(\v2\r.values.ValueR\adefaultB\r\n" +
	"\vobservation\"\xa2\x01\n" +
	"\tFieldsMap\x129\n" +
	"\x06fields\x18\x01 \x03(\v2!.cre.sdk.v2.FieldsMap.FieldsEntryR\x06fields\x1aZ\n" +
	"\vFieldsEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x125\n" +
	"\x05value\x18\x02 \x01(\v2\x1f.cre.sdk.v2.ConsensusDescriptorR\x05value:\x028\x01\"\x9b\x01\n" +
	"\x13ConsensusDescriptor\x12?\n" +
	"\vaggregation\x18\x01 \x01(\x0e2\x1b.cre.sdk.v2.AggregationTypeH\x00R\vaggregation\x125\n" +
	"\tfieldsMap\x18\x02 \x01(\v2\x15.cre.sdk.v2.FieldsMapH\x00R\tfieldsMapB\f\n" +
	"\n" +
	"descriptor\"\x8b\x01\n" +
	"\x11CapabilityRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12.\n" +
	"\apayload\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\apayload\x12\x16\n" +
	"\x06method\x18\x03 \x01(\tR\x06method\x12\x1e\n" +
	"\n" +
	"callbackId\x18\x04 \x01(\x05R\n" +
	"callbackId\"j\n" +
	"\x12CapabilityResponse\x120\n" +
	"\apayload\x18\x01 \x01(\v2\x14.google.protobuf.AnyH\x00R\apayload\x12\x16\n" +
	"\x05error\x18\x02 \x01(\tH\x00R\x05errorB\n" +
	"\n" +
	"\bresponse\"m\n" +
	"\x13TriggerSubscription\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12.\n" +
	"\apayload\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\apayload\x12\x16\n" +
	"\x06method\x18\x03 \x01(\tR\x06method\"c\n" +
	"\x1aTriggerSubscriptionRequest\x12E\n" +
	"\rsubscriptions\x18\x01 \x03(\v2\x1f.cre.sdk.v2.TriggerSubscriptionR\rsubscriptions\"I\n" +
	"\aTrigger\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x04R\x02id\x12.\n" +
	"\apayload\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\apayload\",\n" +
	"\x18AwaitCapabilitiesRequest\x12\x10\n" +
	"\x03ids\x18\x01 \x03(\x05R\x03ids\"\xcd\x01\n" +
	"\x19AwaitCapabilitiesResponse\x12R\n" +
	"\tresponses\x18\x01 \x03(\v24.cre.sdk.v2.AwaitCapabilitiesResponse.ResponsesEntryR\tresponses\x1a\\\n" +
	"\x0eResponsesEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\x05R\x03key\x124\n" +
	"\x05value\x18\x02 \x01(\v2\x1e.cre.sdk.v2.CapabilityResponseR\x05value:\x028\x01*R\n" +
	"\x0fAggregationType\x12\n" +
	"\n" +
	"\x06MEDIAN\x10\x00\x12\r\n" +
	"\tIDENTICAL\x10\x01\x12\x11\n" +
	"\rCOMMON_PREFIX\x10\x02\x12\x11\n" +
	"\rCOMMON_SUFFIX\x10\x03*\x19\n" +
	"\x04Mode\x12\a\n" +
	"\x03DON\x10\x00\x12\b\n" +
	"\x04Node\x10\x01BFZDgithub.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pbb\x06proto3"

var (
	file_workflows_sdk_v2_pb_sdk_proto_rawDescOnce sync.Once
	file_workflows_sdk_v2_pb_sdk_proto_rawDescData []byte
)

func file_workflows_sdk_v2_pb_sdk_proto_rawDescGZIP() []byte {
	file_workflows_sdk_v2_pb_sdk_proto_rawDescOnce.Do(func() {
		file_workflows_sdk_v2_pb_sdk_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_workflows_sdk_v2_pb_sdk_proto_rawDesc), len(file_workflows_sdk_v2_pb_sdk_proto_rawDesc)))
	})
	return file_workflows_sdk_v2_pb_sdk_proto_rawDescData
}

var file_workflows_sdk_v2_pb_sdk_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_workflows_sdk_v2_pb_sdk_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_workflows_sdk_v2_pb_sdk_proto_goTypes = []any{
	(AggregationType)(0),               // 0: cre.sdk.v2.AggregationType
	(Mode)(0),                          // 1: cre.sdk.v2.Mode
	(*SimpleConsensusInputs)(nil),      // 2: cre.sdk.v2.SimpleConsensusInputs
	(*FieldsMap)(nil),                  // 3: cre.sdk.v2.FieldsMap
	(*ConsensusDescriptor)(nil),        // 4: cre.sdk.v2.ConsensusDescriptor
	(*CapabilityRequest)(nil),          // 5: cre.sdk.v2.CapabilityRequest
	(*CapabilityResponse)(nil),         // 6: cre.sdk.v2.CapabilityResponse
	(*TriggerSubscription)(nil),        // 7: cre.sdk.v2.TriggerSubscription
	(*TriggerSubscriptionRequest)(nil), // 8: cre.sdk.v2.TriggerSubscriptionRequest
	(*Trigger)(nil),                    // 9: cre.sdk.v2.Trigger
	(*AwaitCapabilitiesRequest)(nil),   // 10: cre.sdk.v2.AwaitCapabilitiesRequest
	(*AwaitCapabilitiesResponse)(nil),  // 11: cre.sdk.v2.AwaitCapabilitiesResponse
	nil,                                // 12: cre.sdk.v2.FieldsMap.FieldsEntry
	nil,                                // 13: cre.sdk.v2.AwaitCapabilitiesResponse.ResponsesEntry
	(*pb.Value)(nil),                   // 14: values.Value
	(*anypb.Any)(nil),                  // 15: google.protobuf.Any
}
var file_workflows_sdk_v2_pb_sdk_proto_depIdxs = []int32{
	14, // 0: cre.sdk.v2.SimpleConsensusInputs.value:type_name -> values.Value
	4,  // 1: cre.sdk.v2.SimpleConsensusInputs.descriptors:type_name -> cre.sdk.v2.ConsensusDescriptor
	14, // 2: cre.sdk.v2.SimpleConsensusInputs.default:type_name -> values.Value
	12, // 3: cre.sdk.v2.FieldsMap.fields:type_name -> cre.sdk.v2.FieldsMap.FieldsEntry
	0,  // 4: cre.sdk.v2.ConsensusDescriptor.aggregation:type_name -> cre.sdk.v2.AggregationType
	3,  // 5: cre.sdk.v2.ConsensusDescriptor.fieldsMap:type_name -> cre.sdk.v2.FieldsMap
	15, // 6: cre.sdk.v2.CapabilityRequest.payload:type_name -> google.protobuf.Any
	15, // 7: cre.sdk.v2.CapabilityResponse.payload:type_name -> google.protobuf.Any
	15, // 8: cre.sdk.v2.TriggerSubscription.payload:type_name -> google.protobuf.Any
	7,  // 9: cre.sdk.v2.TriggerSubscriptionRequest.subscriptions:type_name -> cre.sdk.v2.TriggerSubscription
	15, // 10: cre.sdk.v2.Trigger.payload:type_name -> google.protobuf.Any
	13, // 11: cre.sdk.v2.AwaitCapabilitiesResponse.responses:type_name -> cre.sdk.v2.AwaitCapabilitiesResponse.ResponsesEntry
	4,  // 12: cre.sdk.v2.FieldsMap.FieldsEntry.value:type_name -> cre.sdk.v2.ConsensusDescriptor
	6,  // 13: cre.sdk.v2.AwaitCapabilitiesResponse.ResponsesEntry.value:type_name -> cre.sdk.v2.CapabilityResponse
	14, // [14:14] is the sub-list for method output_type
	14, // [14:14] is the sub-list for method input_type
	14, // [14:14] is the sub-list for extension type_name
	14, // [14:14] is the sub-list for extension extendee
	0,  // [0:14] is the sub-list for field type_name
}

func init() { file_workflows_sdk_v2_pb_sdk_proto_init() }
func file_workflows_sdk_v2_pb_sdk_proto_init() {
	if File_workflows_sdk_v2_pb_sdk_proto != nil {
		return
	}
	file_workflows_sdk_v2_pb_sdk_proto_msgTypes[0].OneofWrappers = []any{
		(*SimpleConsensusInputs_Value)(nil),
		(*SimpleConsensusInputs_Error)(nil),
	}
	file_workflows_sdk_v2_pb_sdk_proto_msgTypes[2].OneofWrappers = []any{
		(*ConsensusDescriptor_Aggregation)(nil),
		(*ConsensusDescriptor_FieldsMap)(nil),
	}
	file_workflows_sdk_v2_pb_sdk_proto_msgTypes[4].OneofWrappers = []any{
		(*CapabilityResponse_Payload)(nil),
		(*CapabilityResponse_Error)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_workflows_sdk_v2_pb_sdk_proto_rawDesc), len(file_workflows_sdk_v2_pb_sdk_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_workflows_sdk_v2_pb_sdk_proto_goTypes,
		DependencyIndexes: file_workflows_sdk_v2_pb_sdk_proto_depIdxs,
		EnumInfos:         file_workflows_sdk_v2_pb_sdk_proto_enumTypes,
		MessageInfos:      file_workflows_sdk_v2_pb_sdk_proto_msgTypes,
	}.Build()
	File_workflows_sdk_v2_pb_sdk_proto = out.File
	file_workflows_sdk_v2_pb_sdk_proto_goTypes = nil
	file_workflows_sdk_v2_pb_sdk_proto_depIdxs = nil
}
