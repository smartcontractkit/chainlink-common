// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: loop/internal/pb/ccip/tokendata.proto

package ccippb

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

// TokenDataRequest is a gRPC adapter for the input arguments of
// [github.com/smartcontractkit/chainlink-common/chainlink-common/pkg/types/ccip/TokenDataReader.ReadTokenData]
type TokenDataRequest struct {
	state         protoimpl.MessageState                  `protogen:"open.v1"`
	Msg           *EVM2EVMOnRampCCIPSendRequestedWithMeta `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	TokenIndex    uint64                                  `protobuf:"varint,2,opt,name=token_index,json=tokenIndex,proto3" json:"token_index,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TokenDataRequest) Reset() {
	*x = TokenDataRequest{}
	mi := &file_loop_internal_pb_ccip_tokendata_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TokenDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenDataRequest) ProtoMessage() {}

func (x *TokenDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_loop_internal_pb_ccip_tokendata_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenDataRequest.ProtoReflect.Descriptor instead.
func (*TokenDataRequest) Descriptor() ([]byte, []int) {
	return file_loop_internal_pb_ccip_tokendata_proto_rawDescGZIP(), []int{0}
}

func (x *TokenDataRequest) GetMsg() *EVM2EVMOnRampCCIPSendRequestedWithMeta {
	if x != nil {
		return x.Msg
	}
	return nil
}

func (x *TokenDataRequest) GetTokenIndex() uint64 {
	if x != nil {
		return x.TokenIndex
	}
	return 0
}

// TokenDataResponse is a gRPC adapter for the return value of
// [github.com/smartcontractkit/chainlink-common/chainlink-common/pkg/types/ccip/TokenDataReader.ReadTokenData]
type TokenDataResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TokenData     []byte                 `protobuf:"bytes,1,opt,name=token_data,json=tokenData,proto3" json:"token_data,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TokenDataResponse) Reset() {
	*x = TokenDataResponse{}
	mi := &file_loop_internal_pb_ccip_tokendata_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TokenDataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenDataResponse) ProtoMessage() {}

func (x *TokenDataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_loop_internal_pb_ccip_tokendata_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenDataResponse.ProtoReflect.Descriptor instead.
func (*TokenDataResponse) Descriptor() ([]byte, []int) {
	return file_loop_internal_pb_ccip_tokendata_proto_rawDescGZIP(), []int{1}
}

func (x *TokenDataResponse) GetTokenData() []byte {
	if x != nil {
		return x.TokenData
	}
	return nil
}

var File_loop_internal_pb_ccip_tokendata_proto protoreflect.FileDescriptor

const file_loop_internal_pb_ccip_tokendata_proto_rawDesc = "" +
	"\n" +
	"%loop/internal/pb/ccip/tokendata.proto\x12\x15loop.internal.pb.ccip\x1a\x1bgoogle/protobuf/empty.proto\x1a\"loop/internal/pb/ccip/models.proto\"\x84\x01\n" +
	"\x10TokenDataRequest\x12O\n" +
	"\x03msg\x18\x01 \x01(\v2=.loop.internal.pb.ccip.EVM2EVMOnRampCCIPSendRequestedWithMetaR\x03msg\x12\x1f\n" +
	"\vtoken_index\x18\x02 \x01(\x04R\n" +
	"tokenIndex\"2\n" +
	"\x11TokenDataResponse\x12\x1d\n" +
	"\n" +
	"token_data\x18\x01 \x01(\fR\ttokenData2\xb2\x01\n" +
	"\x0fTokenDataReader\x12d\n" +
	"\rReadTokenData\x12'.loop.internal.pb.ccip.TokenDataRequest\x1a(.loop.internal.pb.ccip.TokenDataResponse\"\x00\x129\n" +
	"\x05Close\x12\x16.google.protobuf.Empty\x1a\x16.google.protobuf.Empty\"\x00BOZMgithub.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip;ccippbb\x06proto3"

var (
	file_loop_internal_pb_ccip_tokendata_proto_rawDescOnce sync.Once
	file_loop_internal_pb_ccip_tokendata_proto_rawDescData []byte
)

func file_loop_internal_pb_ccip_tokendata_proto_rawDescGZIP() []byte {
	file_loop_internal_pb_ccip_tokendata_proto_rawDescOnce.Do(func() {
		file_loop_internal_pb_ccip_tokendata_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_loop_internal_pb_ccip_tokendata_proto_rawDesc), len(file_loop_internal_pb_ccip_tokendata_proto_rawDesc)))
	})
	return file_loop_internal_pb_ccip_tokendata_proto_rawDescData
}

var file_loop_internal_pb_ccip_tokendata_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_loop_internal_pb_ccip_tokendata_proto_goTypes = []any{
	(*TokenDataRequest)(nil),                       // 0: loop.internal.pb.ccip.TokenDataRequest
	(*TokenDataResponse)(nil),                      // 1: loop.internal.pb.ccip.TokenDataResponse
	(*EVM2EVMOnRampCCIPSendRequestedWithMeta)(nil), // 2: loop.internal.pb.ccip.EVM2EVMOnRampCCIPSendRequestedWithMeta
	(*emptypb.Empty)(nil),                          // 3: google.protobuf.Empty
}
var file_loop_internal_pb_ccip_tokendata_proto_depIdxs = []int32{
	2, // 0: loop.internal.pb.ccip.TokenDataRequest.msg:type_name -> loop.internal.pb.ccip.EVM2EVMOnRampCCIPSendRequestedWithMeta
	0, // 1: loop.internal.pb.ccip.TokenDataReader.ReadTokenData:input_type -> loop.internal.pb.ccip.TokenDataRequest
	3, // 2: loop.internal.pb.ccip.TokenDataReader.Close:input_type -> google.protobuf.Empty
	1, // 3: loop.internal.pb.ccip.TokenDataReader.ReadTokenData:output_type -> loop.internal.pb.ccip.TokenDataResponse
	3, // 4: loop.internal.pb.ccip.TokenDataReader.Close:output_type -> google.protobuf.Empty
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_loop_internal_pb_ccip_tokendata_proto_init() }
func file_loop_internal_pb_ccip_tokendata_proto_init() {
	if File_loop_internal_pb_ccip_tokendata_proto != nil {
		return
	}
	file_loop_internal_pb_ccip_models_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_loop_internal_pb_ccip_tokendata_proto_rawDesc), len(file_loop_internal_pb_ccip_tokendata_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_loop_internal_pb_ccip_tokendata_proto_goTypes,
		DependencyIndexes: file_loop_internal_pb_ccip_tokendata_proto_depIdxs,
		MessageInfos:      file_loop_internal_pb_ccip_tokendata_proto_msgTypes,
	}.Build()
	File_loop_internal_pb_ccip_tokendata_proto = out.File
	file_loop_internal_pb_ccip_tokendata_proto_goTypes = nil
	file_loop_internal_pb_ccip_tokendata_proto_depIdxs = nil
}
