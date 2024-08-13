// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: codec.proto

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

type VersionedBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Data    []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *VersionedBytes) Reset() {
	*x = VersionedBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_codec_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VersionedBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VersionedBytes) ProtoMessage() {}

func (x *VersionedBytes) ProtoReflect() protoreflect.Message {
	mi := &file_codec_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VersionedBytes.ProtoReflect.Descriptor instead.
func (*VersionedBytes) Descriptor() ([]byte, []int) {
	return file_codec_proto_rawDescGZIP(), []int{0}
}

func (x *VersionedBytes) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *VersionedBytes) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type GetEncodingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Params   *VersionedBytes `protobuf:"bytes,1,opt,name=params,proto3" json:"params,omitempty"`
	ItemType string          `protobuf:"bytes,2,opt,name=itemType,proto3" json:"itemType,omitempty"`
}

func (x *GetEncodingRequest) Reset() {
	*x = GetEncodingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_codec_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetEncodingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetEncodingRequest) ProtoMessage() {}

func (x *GetEncodingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_codec_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetEncodingRequest.ProtoReflect.Descriptor instead.
func (*GetEncodingRequest) Descriptor() ([]byte, []int) {
	return file_codec_proto_rawDescGZIP(), []int{1}
}

func (x *GetEncodingRequest) GetParams() *VersionedBytes {
	if x != nil {
		return x.Params
	}
	return nil
}

func (x *GetEncodingRequest) GetItemType() string {
	if x != nil {
		return x.ItemType
	}
	return ""
}

type GetEncodingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RetVal []byte `protobuf:"bytes,1,opt,name=retVal,proto3" json:"retVal,omitempty"`
}

func (x *GetEncodingResponse) Reset() {
	*x = GetEncodingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_codec_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetEncodingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetEncodingResponse) ProtoMessage() {}

func (x *GetEncodingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_codec_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetEncodingResponse.ProtoReflect.Descriptor instead.
func (*GetEncodingResponse) Descriptor() ([]byte, []int) {
	return file_codec_proto_rawDescGZIP(), []int{2}
}

func (x *GetEncodingResponse) GetRetVal() []byte {
	if x != nil {
		return x.RetVal
	}
	return nil
}

type GetDecodingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Encoded             []byte `protobuf:"bytes,1,opt,name=encoded,proto3" json:"encoded,omitempty"`
	ItemType            string `protobuf:"bytes,2,opt,name=itemType,proto3" json:"itemType,omitempty"`
	WireEncodingVersion uint32 `protobuf:"varint,3,opt,name=wireEncodingVersion,proto3" json:"wireEncodingVersion,omitempty"`
}

func (x *GetDecodingRequest) Reset() {
	*x = GetDecodingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_codec_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDecodingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDecodingRequest) ProtoMessage() {}

func (x *GetDecodingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_codec_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDecodingRequest.ProtoReflect.Descriptor instead.
func (*GetDecodingRequest) Descriptor() ([]byte, []int) {
	return file_codec_proto_rawDescGZIP(), []int{3}
}

func (x *GetDecodingRequest) GetEncoded() []byte {
	if x != nil {
		return x.Encoded
	}
	return nil
}

func (x *GetDecodingRequest) GetItemType() string {
	if x != nil {
		return x.ItemType
	}
	return ""
}

func (x *GetDecodingRequest) GetWireEncodingVersion() uint32 {
	if x != nil {
		return x.WireEncodingVersion
	}
	return 0
}

type GetDecodingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RetVal *VersionedBytes `protobuf:"bytes,1,opt,name=retVal,proto3" json:"retVal,omitempty"`
}

func (x *GetDecodingResponse) Reset() {
	*x = GetDecodingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_codec_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDecodingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDecodingResponse) ProtoMessage() {}

func (x *GetDecodingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_codec_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDecodingResponse.ProtoReflect.Descriptor instead.
func (*GetDecodingResponse) Descriptor() ([]byte, []int) {
	return file_codec_proto_rawDescGZIP(), []int{4}
}

func (x *GetDecodingResponse) GetRetVal() *VersionedBytes {
	if x != nil {
		return x.RetVal
	}
	return nil
}

type GetMaxSizeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	N           int32  `protobuf:"varint,1,opt,name=n,proto3" json:"n,omitempty"`
	ItemType    string `protobuf:"bytes,2,opt,name=itemType,proto3" json:"itemType,omitempty"`
	ForEncoding bool   `protobuf:"varint,3,opt,name=forEncoding,proto3" json:"forEncoding,omitempty"`
}

func (x *GetMaxSizeRequest) Reset() {
	*x = GetMaxSizeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_codec_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetMaxSizeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMaxSizeRequest) ProtoMessage() {}

func (x *GetMaxSizeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_codec_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMaxSizeRequest.ProtoReflect.Descriptor instead.
func (*GetMaxSizeRequest) Descriptor() ([]byte, []int) {
	return file_codec_proto_rawDescGZIP(), []int{5}
}

func (x *GetMaxSizeRequest) GetN() int32 {
	if x != nil {
		return x.N
	}
	return 0
}

func (x *GetMaxSizeRequest) GetItemType() string {
	if x != nil {
		return x.ItemType
	}
	return ""
}

func (x *GetMaxSizeRequest) GetForEncoding() bool {
	if x != nil {
		return x.ForEncoding
	}
	return false
}

type GetMaxSizeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SizeInBytes int32 `protobuf:"varint,1,opt,name=sizeInBytes,proto3" json:"sizeInBytes,omitempty"`
}

func (x *GetMaxSizeResponse) Reset() {
	*x = GetMaxSizeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_codec_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetMaxSizeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMaxSizeResponse) ProtoMessage() {}

func (x *GetMaxSizeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_codec_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMaxSizeResponse.ProtoReflect.Descriptor instead.
func (*GetMaxSizeResponse) Descriptor() ([]byte, []int) {
	return file_codec_proto_rawDescGZIP(), []int{6}
}

func (x *GetMaxSizeResponse) GetSizeInBytes() int32 {
	if x != nil {
		return x.SizeInBytes
	}
	return 0
}

var File_codec_proto protoreflect.FileDescriptor

var file_codec_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x63, 0x6f, 0x64, 0x65, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x6c,
	0x6f, 0x6f, 0x70, 0x22, 0x3e, 0x0a, 0x0e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x64,
	0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x22, 0x5e, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69,
	0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x06, 0x70, 0x61, 0x72,
	0x61, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6c, 0x6f, 0x6f, 0x70,
	0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x64, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52,
	0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x74, 0x65, 0x6d, 0x54,
	0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x74, 0x65, 0x6d, 0x54,
	0x79, 0x70, 0x65, 0x22, 0x2d, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69,
	0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65,
	0x74, 0x56, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x72, 0x65, 0x74, 0x56,
	0x61, 0x6c, 0x22, 0x7c, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x69, 0x6e,
	0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x63, 0x6f,
	0x64, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x65, 0x6e, 0x63, 0x6f, 0x64,
	0x65, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x74, 0x65, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x74, 0x65, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x12, 0x30,
	0x0a, 0x13, 0x77, 0x69, 0x72, 0x65, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x13, 0x77, 0x69, 0x72,
	0x65, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x22, 0x43, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x72, 0x65, 0x74, 0x56, 0x61,
	0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x64, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52, 0x06, 0x72,
	0x65, 0x74, 0x56, 0x61, 0x6c, 0x22, 0x5f, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x4d, 0x61, 0x78, 0x53,
	0x69, 0x7a, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0c, 0x0a, 0x01, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x74, 0x65, 0x6d,
	0x54, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x74, 0x65, 0x6d,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x66, 0x6f, 0x72, 0x45, 0x6e, 0x63, 0x6f, 0x64,
	0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x66, 0x6f, 0x72, 0x45, 0x6e,
	0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x22, 0x36, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x4d, 0x61, 0x78,
	0x53, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x20, 0x0a, 0x0b,
	0x73, 0x69, 0x7a, 0x65, 0x49, 0x6e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0b, 0x73, 0x69, 0x7a, 0x65, 0x49, 0x6e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x32, 0xd0,
	0x01, 0x0a, 0x05, 0x43, 0x6f, 0x64, 0x65, 0x63, 0x12, 0x42, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x45,
	0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x18, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x47,
	0x65, 0x74, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x19, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x47, 0x65, 0x74, 0x45, 0x6e, 0x63, 0x6f,
	0x64, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x42, 0x0a, 0x0b,
	0x47, 0x65, 0x74, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x18, 0x2e, 0x6c, 0x6f,
	0x6f, 0x70, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x47, 0x65, 0x74,
	0x44, 0x65, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x3f, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x4d, 0x61, 0x78, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x17,
	0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x47, 0x65, 0x74, 0x4d, 0x61, 0x78, 0x53, 0x69, 0x7a, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x47,
	0x65, 0x74, 0x4d, 0x61, 0x78, 0x53, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x43, 0x5a, 0x41, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x73, 0x6d, 0x61, 0x72, 0x74, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6b, 0x69, 0x74,
	0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6c, 0x6f, 0x6f, 0x70, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_codec_proto_rawDescOnce sync.Once
	file_codec_proto_rawDescData = file_codec_proto_rawDesc
)

func file_codec_proto_rawDescGZIP() []byte {
	file_codec_proto_rawDescOnce.Do(func() {
		file_codec_proto_rawDescData = protoimpl.X.CompressGZIP(file_codec_proto_rawDescData)
	})
	return file_codec_proto_rawDescData
}

var file_codec_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_codec_proto_goTypes = []interface{}{
	(*VersionedBytes)(nil),      // 0: loop.VersionedBytes
	(*GetEncodingRequest)(nil),  // 1: loop.GetEncodingRequest
	(*GetEncodingResponse)(nil), // 2: loop.GetEncodingResponse
	(*GetDecodingRequest)(nil),  // 3: loop.GetDecodingRequest
	(*GetDecodingResponse)(nil), // 4: loop.GetDecodingResponse
	(*GetMaxSizeRequest)(nil),   // 5: loop.GetMaxSizeRequest
	(*GetMaxSizeResponse)(nil),  // 6: loop.GetMaxSizeResponse
}
var file_codec_proto_depIdxs = []int32{
	0, // 0: loop.GetEncodingRequest.params:type_name -> loop.VersionedBytes
	0, // 1: loop.GetDecodingResponse.retVal:type_name -> loop.VersionedBytes
	1, // 2: loop.Codec.GetEncoding:input_type -> loop.GetEncodingRequest
	3, // 3: loop.Codec.GetDecoding:input_type -> loop.GetDecodingRequest
	5, // 4: loop.Codec.GetMaxSize:input_type -> loop.GetMaxSizeRequest
	2, // 5: loop.Codec.GetEncoding:output_type -> loop.GetEncodingResponse
	4, // 6: loop.Codec.GetDecoding:output_type -> loop.GetDecodingResponse
	6, // 7: loop.Codec.GetMaxSize:output_type -> loop.GetMaxSizeResponse
	5, // [5:8] is the sub-list for method output_type
	2, // [2:5] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_codec_proto_init() }
func file_codec_proto_init() {
	if File_codec_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_codec_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VersionedBytes); i {
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
		file_codec_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetEncodingRequest); i {
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
		file_codec_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetEncodingResponse); i {
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
		file_codec_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDecodingRequest); i {
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
		file_codec_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDecodingResponse); i {
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
		file_codec_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetMaxSizeRequest); i {
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
		file_codec_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetMaxSizeResponse); i {
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
			RawDescriptor: file_codec_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_codec_proto_goTypes,
		DependencyIndexes: file_codec_proto_depIdxs,
		MessageInfos:      file_codec_proto_msgTypes,
	}.Build()
	File_codec_proto = out.File
	file_codec_proto_rawDesc = nil
	file_codec_proto_goTypes = nil
	file_codec_proto_depIdxs = nil
}
