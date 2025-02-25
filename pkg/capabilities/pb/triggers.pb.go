// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.3
// source: capabilities/pb/triggers.proto

package pb

import (
	pb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
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

type OCRTriggerEvent struct {
	state         protoimpl.MessageState           `protogen:"open.v1"`
	ConfigDigest  []byte                           `protobuf:"bytes,1,opt,name=configDigest,proto3" json:"configDigest,omitempty"`
	SeqNr         uint64                           `protobuf:"varint,2,opt,name=seqNr,proto3" json:"seqNr,omitempty"`
	Report        []byte                           `protobuf:"bytes,3,opt,name=report,proto3" json:"report,omitempty"` // marshalled OCRTriggerReport
	Sigs          []*OCRAttributedOnchainSignature `protobuf:"bytes,4,rep,name=sigs,proto3" json:"sigs,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OCRTriggerEvent) Reset() {
	*x = OCRTriggerEvent{}
	mi := &file_capabilities_pb_triggers_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OCRTriggerEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OCRTriggerEvent) ProtoMessage() {}

func (x *OCRTriggerEvent) ProtoReflect() protoreflect.Message {
	mi := &file_capabilities_pb_triggers_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OCRTriggerEvent.ProtoReflect.Descriptor instead.
func (*OCRTriggerEvent) Descriptor() ([]byte, []int) {
	return file_capabilities_pb_triggers_proto_rawDescGZIP(), []int{0}
}

func (x *OCRTriggerEvent) GetConfigDigest() []byte {
	if x != nil {
		return x.ConfigDigest
	}
	return nil
}

func (x *OCRTriggerEvent) GetSeqNr() uint64 {
	if x != nil {
		return x.SeqNr
	}
	return 0
}

func (x *OCRTriggerEvent) GetReport() []byte {
	if x != nil {
		return x.Report
	}
	return nil
}

func (x *OCRTriggerEvent) GetSigs() []*OCRAttributedOnchainSignature {
	if x != nil {
		return x.Sigs
	}
	return nil
}

type OCRAttributedOnchainSignature struct {
	state     protoimpl.MessageState `protogen:"open.v1"`
	Signature []byte                 `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	// signer is actually a uint8 but uint32 is the smallest supported by protobuf
	Signer        uint32 `protobuf:"varint,2,opt,name=signer,proto3" json:"signer,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OCRAttributedOnchainSignature) Reset() {
	*x = OCRAttributedOnchainSignature{}
	mi := &file_capabilities_pb_triggers_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OCRAttributedOnchainSignature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OCRAttributedOnchainSignature) ProtoMessage() {}

func (x *OCRAttributedOnchainSignature) ProtoReflect() protoreflect.Message {
	mi := &file_capabilities_pb_triggers_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OCRAttributedOnchainSignature.ProtoReflect.Descriptor instead.
func (*OCRAttributedOnchainSignature) Descriptor() ([]byte, []int) {
	return file_capabilities_pb_triggers_proto_rawDescGZIP(), []int{1}
}

func (x *OCRAttributedOnchainSignature) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *OCRAttributedOnchainSignature) GetSigner() uint32 {
	if x != nil {
		return x.Signer
	}
	return 0
}

type OCRTriggerReport struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	EventID       string                 `protobuf:"bytes,1,opt,name=eventID,proto3" json:"eventID,omitempty"`      // unique, scoped to the trigger capability
	Timestamp     int64                  `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"` // used to enforce freshness
	Outputs       *pb.Map                `protobuf:"bytes,3,opt,name=outputs,proto3" json:"outputs,omitempty"`      // contains trigger-specific data
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OCRTriggerReport) Reset() {
	*x = OCRTriggerReport{}
	mi := &file_capabilities_pb_triggers_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OCRTriggerReport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OCRTriggerReport) ProtoMessage() {}

func (x *OCRTriggerReport) ProtoReflect() protoreflect.Message {
	mi := &file_capabilities_pb_triggers_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OCRTriggerReport.ProtoReflect.Descriptor instead.
func (*OCRTriggerReport) Descriptor() ([]byte, []int) {
	return file_capabilities_pb_triggers_proto_rawDescGZIP(), []int{2}
}

func (x *OCRTriggerReport) GetEventID() string {
	if x != nil {
		return x.EventID
	}
	return ""
}

func (x *OCRTriggerReport) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *OCRTriggerReport) GetOutputs() *pb.Map {
	if x != nil {
		return x.Outputs
	}
	return nil
}

var File_capabilities_pb_triggers_proto protoreflect.FileDescriptor

var file_capabilities_pb_triggers_proto_rawDesc = string([]byte{
	0x0a, 0x1e, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x2f, 0x70,
	0x62, 0x2f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x02, 0x76, 0x31, 0x1a, 0x16, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2f, 0x70, 0x62, 0x2f,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9a, 0x01, 0x0a,
	0x0f, 0x4f, 0x43, 0x52, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x12, 0x22, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x44, 0x69, 0x67, 0x65, 0x73, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x44, 0x69,
	0x67, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x65, 0x71, 0x4e, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x05, 0x73, 0x65, 0x71, 0x4e, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x72, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x12, 0x35, 0x0a, 0x04, 0x73, 0x69, 0x67, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x21, 0x2e, 0x76, 0x31, 0x2e, 0x4f, 0x43, 0x52, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
	0x74, 0x65, 0x64, 0x4f, 0x6e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x52, 0x04, 0x73, 0x69, 0x67, 0x73, 0x22, 0x55, 0x0a, 0x1d, 0x4f, 0x43, 0x52,
	0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x64, 0x4f, 0x6e, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69,
	0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x69, 0x67, 0x6e,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72,
	0x22, 0x71, 0x0a, 0x10, 0x4f, 0x43, 0x52, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x12, 0x1c,
	0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x25, 0x0a, 0x07,
	0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2e, 0x4d, 0x61, 0x70, 0x52, 0x07, 0x6f, 0x75, 0x74, 0x70,
	0x75, 0x74, 0x73, 0x42, 0x42, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6b,
	0x69, 0x74, 0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x2d, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69,
	0x74, 0x69, 0x65, 0x73, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_capabilities_pb_triggers_proto_rawDescOnce sync.Once
	file_capabilities_pb_triggers_proto_rawDescData []byte
)

func file_capabilities_pb_triggers_proto_rawDescGZIP() []byte {
	file_capabilities_pb_triggers_proto_rawDescOnce.Do(func() {
		file_capabilities_pb_triggers_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_capabilities_pb_triggers_proto_rawDesc), len(file_capabilities_pb_triggers_proto_rawDesc)))
	})
	return file_capabilities_pb_triggers_proto_rawDescData
}

var file_capabilities_pb_triggers_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_capabilities_pb_triggers_proto_goTypes = []any{
	(*OCRTriggerEvent)(nil),               // 0: v1.OCRTriggerEvent
	(*OCRAttributedOnchainSignature)(nil), // 1: v1.OCRAttributedOnchainSignature
	(*OCRTriggerReport)(nil),              // 2: v1.OCRTriggerReport
	(*pb.Map)(nil),                        // 3: values.Map
}
var file_capabilities_pb_triggers_proto_depIdxs = []int32{
	1, // 0: v1.OCRTriggerEvent.sigs:type_name -> v1.OCRAttributedOnchainSignature
	3, // 1: v1.OCRTriggerReport.outputs:type_name -> values.Map
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_capabilities_pb_triggers_proto_init() }
func file_capabilities_pb_triggers_proto_init() {
	if File_capabilities_pb_triggers_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_capabilities_pb_triggers_proto_rawDesc), len(file_capabilities_pb_triggers_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_capabilities_pb_triggers_proto_goTypes,
		DependencyIndexes: file_capabilities_pb_triggers_proto_depIdxs,
		MessageInfos:      file_capabilities_pb_triggers_proto_msgTypes,
	}.Build()
	File_capabilities_pb_triggers_proto = out.File
	file_capabilities_pb_triggers_proto_goTypes = nil
	file_capabilities_pb_triggers_proto_depIdxs = nil
}
