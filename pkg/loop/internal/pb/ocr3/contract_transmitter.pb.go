// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.3
// source: contract_transmitter.proto

package ocr3pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
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

type Signature struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Signature     []byte                 `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	Signer        uint32                 `protobuf:"varint,2,opt,name=signer,proto3" json:"signer,omitempty"` // NOTE: this is actually a uint8, but proto doesn't support this.
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Signature) Reset() {
	*x = Signature{}
	mi := &file_contract_transmitter_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Signature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signature) ProtoMessage() {}

func (x *Signature) ProtoReflect() protoreflect.Message {
	mi := &file_contract_transmitter_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signature.ProtoReflect.Descriptor instead.
func (*Signature) Descriptor() ([]byte, []int) {
	return file_contract_transmitter_proto_rawDescGZIP(), []int{0}
}

func (x *Signature) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *Signature) GetSigner() uint32 {
	if x != nil {
		return x.Signer
	}
	return 0
}

type TransmitRequest struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	ConfigDigest   []byte                 `protobuf:"bytes,1,opt,name=configDigest,proto3" json:"configDigest,omitempty"` // NOTE: this is actually [32]byte
	SeqNr          uint64                 `protobuf:"varint,2,opt,name=seqNr,proto3" json:"seqNr,omitempty"`
	ReportWithInfo *ReportWithInfo        `protobuf:"bytes,3,opt,name=reportWithInfo,proto3" json:"reportWithInfo,omitempty"`
	Signatures     []*Signature           `protobuf:"bytes,4,rep,name=signatures,proto3" json:"signatures,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *TransmitRequest) Reset() {
	*x = TransmitRequest{}
	mi := &file_contract_transmitter_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransmitRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransmitRequest) ProtoMessage() {}

func (x *TransmitRequest) ProtoReflect() protoreflect.Message {
	mi := &file_contract_transmitter_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransmitRequest.ProtoReflect.Descriptor instead.
func (*TransmitRequest) Descriptor() ([]byte, []int) {
	return file_contract_transmitter_proto_rawDescGZIP(), []int{1}
}

func (x *TransmitRequest) GetConfigDigest() []byte {
	if x != nil {
		return x.ConfigDigest
	}
	return nil
}

func (x *TransmitRequest) GetSeqNr() uint64 {
	if x != nil {
		return x.SeqNr
	}
	return 0
}

func (x *TransmitRequest) GetReportWithInfo() *ReportWithInfo {
	if x != nil {
		return x.ReportWithInfo
	}
	return nil
}

func (x *TransmitRequest) GetSignatures() []*Signature {
	if x != nil {
		return x.Signatures
	}
	return nil
}

type FromAccountReply struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Account       string                 `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FromAccountReply) Reset() {
	*x = FromAccountReply{}
	mi := &file_contract_transmitter_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FromAccountReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FromAccountReply) ProtoMessage() {}

func (x *FromAccountReply) ProtoReflect() protoreflect.Message {
	mi := &file_contract_transmitter_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FromAccountReply.ProtoReflect.Descriptor instead.
func (*FromAccountReply) Descriptor() ([]byte, []int) {
	return file_contract_transmitter_proto_rawDescGZIP(), []int{2}
}

func (x *FromAccountReply) GetAccount() string {
	if x != nil {
		return x.Account
	}
	return ""
}

var File_contract_transmitter_proto protoreflect.FileDescriptor

var file_contract_transmitter_proto_rawDesc = string([]byte{
	0x0a, 0x1a, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x6d, 0x69, 0x74, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x6c, 0x6f,
	0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6f,
	0x63, 0x72, 0x33, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x14, 0x6f, 0x63, 0x72, 0x33, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x41, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x06, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x22, 0xdc, 0x01, 0x0a, 0x0f, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x6d, 0x69, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x22, 0x0a,
	0x0c, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x44, 0x69, 0x67, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x44, 0x69, 0x67, 0x65, 0x73,
	0x74, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x65, 0x71, 0x4e, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x05, 0x73, 0x65, 0x71, 0x4e, 0x72, 0x12, 0x4d, 0x0a, 0x0e, 0x72, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x57, 0x69, 0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x25, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e,
	0x70, 0x62, 0x2e, 0x6f, 0x63, 0x72, 0x33, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x57, 0x69,
	0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0e, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x57, 0x69,
	0x74, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x40, 0x0a, 0x0a, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6c, 0x6f, 0x6f,
	0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6f, 0x63,
	0x72, 0x33, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x0a, 0x73, 0x69,
	0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x22, 0x2c, 0x0a, 0x10, 0x46, 0x72, 0x6f, 0x6d,
	0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a, 0x07,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x32, 0xb5, 0x01, 0x0a, 0x13, 0x43, 0x6f, 0x6e, 0x74, 0x72,
	0x61, 0x63, 0x74, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x72, 0x12, 0x4c,
	0x0a, 0x08, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6d, 0x69, 0x74, 0x12, 0x26, 0x2e, 0x6c, 0x6f, 0x6f,
	0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6f, 0x63,
	0x72, 0x33, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6d, 0x69, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x50, 0x0a, 0x0b,
	0x46, 0x72, 0x6f, 0x6d, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x1a, 0x27, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6f, 0x63, 0x72, 0x33, 0x2e, 0x46, 0x72, 0x6f, 0x6d,
	0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x4f,
	0x5a, 0x4d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6d, 0x61,
	0x72, 0x74, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6b, 0x69, 0x74, 0x2f, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x6c, 0x6f, 0x6f, 0x70, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2f, 0x70, 0x62, 0x2f, 0x6f, 0x63, 0x72, 0x33, 0x3b, 0x6f, 0x63, 0x72, 0x33, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_contract_transmitter_proto_rawDescOnce sync.Once
	file_contract_transmitter_proto_rawDescData []byte
)

func file_contract_transmitter_proto_rawDescGZIP() []byte {
	file_contract_transmitter_proto_rawDescOnce.Do(func() {
		file_contract_transmitter_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_contract_transmitter_proto_rawDesc), len(file_contract_transmitter_proto_rawDesc)))
	})
	return file_contract_transmitter_proto_rawDescData
}

var file_contract_transmitter_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_contract_transmitter_proto_goTypes = []any{
	(*Signature)(nil),        // 0: loop.internal.pb.ocr3.Signature
	(*TransmitRequest)(nil),  // 1: loop.internal.pb.ocr3.TransmitRequest
	(*FromAccountReply)(nil), // 2: loop.internal.pb.ocr3.FromAccountReply
	(*ReportWithInfo)(nil),   // 3: loop.internal.pb.ocr3.ReportWithInfo
	(*emptypb.Empty)(nil),    // 4: google.protobuf.Empty
}
var file_contract_transmitter_proto_depIdxs = []int32{
	3, // 0: loop.internal.pb.ocr3.TransmitRequest.reportWithInfo:type_name -> loop.internal.pb.ocr3.ReportWithInfo
	0, // 1: loop.internal.pb.ocr3.TransmitRequest.signatures:type_name -> loop.internal.pb.ocr3.Signature
	1, // 2: loop.internal.pb.ocr3.ContractTransmitter.Transmit:input_type -> loop.internal.pb.ocr3.TransmitRequest
	4, // 3: loop.internal.pb.ocr3.ContractTransmitter.FromAccount:input_type -> google.protobuf.Empty
	4, // 4: loop.internal.pb.ocr3.ContractTransmitter.Transmit:output_type -> google.protobuf.Empty
	2, // 5: loop.internal.pb.ocr3.ContractTransmitter.FromAccount:output_type -> loop.internal.pb.ocr3.FromAccountReply
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_contract_transmitter_proto_init() }
func file_contract_transmitter_proto_init() {
	if File_contract_transmitter_proto != nil {
		return
	}
	file_ocr3_reporting_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_contract_transmitter_proto_rawDesc), len(file_contract_transmitter_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_contract_transmitter_proto_goTypes,
		DependencyIndexes: file_contract_transmitter_proto_depIdxs,
		MessageInfos:      file_contract_transmitter_proto_msgTypes,
	}.Build()
	File_contract_transmitter_proto = out.File
	file_contract_transmitter_proto_goTypes = nil
	file_contract_transmitter_proto_depIdxs = nil
}
