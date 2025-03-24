// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: pricegetter.proto

package ccippb

import (
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
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

// FilterConfiguredTokensRequest is a request for which tokens of a list of addresses are configured and which aren't.. It is a gRPC adapter for the input arguments of
// [github.com/smartcontractkit/chainlink-common/chainlink-common/pkg/types/ccip/PriceGetter.FilterConfiguredTokens]
type FilterConfiguredTokensRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tokens        []string               `protobuf:"bytes,1,rep,name=tokens,proto3" json:"tokens,omitempty"` // []Address
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FilterConfiguredTokensRequest) Reset() {
	*x = FilterConfiguredTokensRequest{}
	mi := &file_pricegetter_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FilterConfiguredTokensRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FilterConfiguredTokensRequest) ProtoMessage() {}

func (x *FilterConfiguredTokensRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pricegetter_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FilterConfiguredTokensRequest.ProtoReflect.Descriptor instead.
func (*FilterConfiguredTokensRequest) Descriptor() ([]byte, []int) {
	return file_pricegetter_proto_rawDescGZIP(), []int{0}
}

func (x *FilterConfiguredTokensRequest) GetTokens() []string {
	if x != nil {
		return x.Tokens
	}
	return nil
}

// FilterConfiguredTokensResponse is a response for which tokens of a list of addresses are configured and which aren't. It is a gRPC adapter for the return values of
// [github.com/smartcontractkit/chainlink-common/chainlink-common/pkg/types/ccip/PriceGetter.FilterConfiguredTokens]
type FilterConfiguredTokensResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Configured    []string               `protobuf:"bytes,1,rep,name=configured,proto3" json:"configured,omitempty"`     // []Address
	Unconfigured  []string               `protobuf:"bytes,2,rep,name=unconfigured,proto3" json:"unconfigured,omitempty"` // []Address
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FilterConfiguredTokensResponse) Reset() {
	*x = FilterConfiguredTokensResponse{}
	mi := &file_pricegetter_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FilterConfiguredTokensResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FilterConfiguredTokensResponse) ProtoMessage() {}

func (x *FilterConfiguredTokensResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pricegetter_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FilterConfiguredTokensResponse.ProtoReflect.Descriptor instead.
func (*FilterConfiguredTokensResponse) Descriptor() ([]byte, []int) {
	return file_pricegetter_proto_rawDescGZIP(), []int{1}
}

func (x *FilterConfiguredTokensResponse) GetConfigured() []string {
	if x != nil {
		return x.Configured
	}
	return nil
}

func (x *FilterConfiguredTokensResponse) GetUnconfigured() []string {
	if x != nil {
		return x.Unconfigured
	}
	return nil
}

// TokenPricesRequest is a request for the price of a token in USD. It is a gRPC adapter for the input arguments of
// [github.com/smartcontractkit/chainlink-common/chainlink-common/pkg/types/ccip/PriceGetter.TokenPricesUSD]]
type TokenPricesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tokens        []string               `protobuf:"bytes,1,rep,name=tokens,proto3" json:"tokens,omitempty"` // []Address
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TokenPricesRequest) Reset() {
	*x = TokenPricesRequest{}
	mi := &file_pricegetter_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TokenPricesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenPricesRequest) ProtoMessage() {}

func (x *TokenPricesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pricegetter_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenPricesRequest.ProtoReflect.Descriptor instead.
func (*TokenPricesRequest) Descriptor() ([]byte, []int) {
	return file_pricegetter_proto_rawDescGZIP(), []int{2}
}

func (x *TokenPricesRequest) GetTokens() []string {
	if x != nil {
		return x.Tokens
	}
	return nil
}

// TokenPricesResponse is a response for the price of a token in USD. It is a gRPC adapter for the return values of
// [github.com/smartcontractkit/chainlink-common/chainlink-common/pkg/types/ccip/CommitStoreReader]
type TokenPricesResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Prices        map[string]*pb.BigInt  `protobuf:"bytes,1,rep,name=prices,proto3" json:"prices,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"` // map[Address]price
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TokenPricesResponse) Reset() {
	*x = TokenPricesResponse{}
	mi := &file_pricegetter_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TokenPricesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenPricesResponse) ProtoMessage() {}

func (x *TokenPricesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pricegetter_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenPricesResponse.ProtoReflect.Descriptor instead.
func (*TokenPricesResponse) Descriptor() ([]byte, []int) {
	return file_pricegetter_proto_rawDescGZIP(), []int{3}
}

func (x *TokenPricesResponse) GetPrices() map[string]*pb.BigInt {
	if x != nil {
		return x.Prices
	}
	return nil
}

var File_pricegetter_proto protoreflect.FileDescriptor

const file_pricegetter_proto_rawDesc = "" +
	"\n" +
	"\x11pricegetter.proto\x12\x15loop.internal.pb.ccip\x1a\rrelayer.proto\x1a\x1bgoogle/protobuf/empty.proto\"7\n" +
	"\x1dFilterConfiguredTokensRequest\x12\x16\n" +
	"\x06tokens\x18\x01 \x03(\tR\x06tokens\"d\n" +
	"\x1eFilterConfiguredTokensResponse\x12\x1e\n" +
	"\n" +
	"configured\x18\x01 \x03(\tR\n" +
	"configured\x12\"\n" +
	"\funconfigured\x18\x02 \x03(\tR\funconfigured\",\n" +
	"\x12TokenPricesRequest\x12\x16\n" +
	"\x06tokens\x18\x01 \x03(\tR\x06tokens\"\xae\x01\n" +
	"\x13TokenPricesResponse\x12N\n" +
	"\x06prices\x18\x01 \x03(\v26.loop.internal.pb.ccip.TokenPricesResponse.PricesEntryR\x06prices\x1aG\n" +
	"\vPricesEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\"\n" +
	"\x05value\x18\x02 \x01(\v2\f.loop.BigIntR\x05value:\x028\x012\xbd\x02\n" +
	"\vPriceGetter\x12\x87\x01\n" +
	"\x16FilterConfiguredTokens\x124.loop.internal.pb.ccip.FilterConfiguredTokensRequest\x1a5.loop.internal.pb.ccip.FilterConfiguredTokensResponse\"\x00\x12i\n" +
	"\x0eTokenPricesUSD\x12).loop.internal.pb.ccip.TokenPricesRequest\x1a*.loop.internal.pb.ccip.TokenPricesResponse\"\x00\x129\n" +
	"\x05Close\x12\x16.google.protobuf.Empty\x1a\x16.google.protobuf.Empty\"\x00BOZMgithub.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip;ccippbb\x06proto3"

var (
	file_pricegetter_proto_rawDescOnce sync.Once
	file_pricegetter_proto_rawDescData []byte
)

func file_pricegetter_proto_rawDescGZIP() []byte {
	file_pricegetter_proto_rawDescOnce.Do(func() {
		file_pricegetter_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_pricegetter_proto_rawDesc), len(file_pricegetter_proto_rawDesc)))
	})
	return file_pricegetter_proto_rawDescData
}

var file_pricegetter_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pricegetter_proto_goTypes = []any{
	(*FilterConfiguredTokensRequest)(nil),  // 0: loop.internal.pb.ccip.FilterConfiguredTokensRequest
	(*FilterConfiguredTokensResponse)(nil), // 1: loop.internal.pb.ccip.FilterConfiguredTokensResponse
	(*TokenPricesRequest)(nil),             // 2: loop.internal.pb.ccip.TokenPricesRequest
	(*TokenPricesResponse)(nil),            // 3: loop.internal.pb.ccip.TokenPricesResponse
	nil,                                    // 4: loop.internal.pb.ccip.TokenPricesResponse.PricesEntry
	(*pb.BigInt)(nil),                      // 5: loop.BigInt
	(*emptypb.Empty)(nil),                  // 6: google.protobuf.Empty
}
var file_pricegetter_proto_depIdxs = []int32{
	4, // 0: loop.internal.pb.ccip.TokenPricesResponse.prices:type_name -> loop.internal.pb.ccip.TokenPricesResponse.PricesEntry
	5, // 1: loop.internal.pb.ccip.TokenPricesResponse.PricesEntry.value:type_name -> loop.BigInt
	0, // 2: loop.internal.pb.ccip.PriceGetter.FilterConfiguredTokens:input_type -> loop.internal.pb.ccip.FilterConfiguredTokensRequest
	2, // 3: loop.internal.pb.ccip.PriceGetter.TokenPricesUSD:input_type -> loop.internal.pb.ccip.TokenPricesRequest
	6, // 4: loop.internal.pb.ccip.PriceGetter.Close:input_type -> google.protobuf.Empty
	1, // 5: loop.internal.pb.ccip.PriceGetter.FilterConfiguredTokens:output_type -> loop.internal.pb.ccip.FilterConfiguredTokensResponse
	3, // 6: loop.internal.pb.ccip.PriceGetter.TokenPricesUSD:output_type -> loop.internal.pb.ccip.TokenPricesResponse
	6, // 7: loop.internal.pb.ccip.PriceGetter.Close:output_type -> google.protobuf.Empty
	5, // [5:8] is the sub-list for method output_type
	2, // [2:5] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pricegetter_proto_init() }
func file_pricegetter_proto_init() {
	if File_pricegetter_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_pricegetter_proto_rawDesc), len(file_pricegetter_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pricegetter_proto_goTypes,
		DependencyIndexes: file_pricegetter_proto_depIdxs,
		MessageInfos:      file_pricegetter_proto_msgTypes,
	}.Build()
	File_pricegetter_proto = out.File
	file_pricegetter_proto_goTypes = nil
	file_pricegetter_proto_depIdxs = nil
}
