// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: pricegetter.proto

package ccippb

import (
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// TokenPricesRequest is a request for the price of a token in USD. It is a gRPC adapter for the input arguments of
// [github.com/smartcontractkit/chainlink-common/chainlink-common/pkg/types/ccip/PriceGetter.TokenPricesUSD]]
type TokenPricesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tokens []string `protobuf:"bytes,1,rep,name=tokens,proto3" json:"tokens,omitempty"` // []Address
}

func (x *TokenPricesRequest) Reset() {
	*x = TokenPricesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pricegetter_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TokenPricesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenPricesRequest) ProtoMessage() {}

func (x *TokenPricesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pricegetter_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_pricegetter_proto_rawDescGZIP(), []int{0}
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
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prices map[string]*pb.BigInt `protobuf:"bytes,1,rep,name=prices,proto3" json:"prices,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // map[Address]price
}

func (x *TokenPricesResponse) Reset() {
	*x = TokenPricesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pricegetter_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TokenPricesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenPricesResponse) ProtoMessage() {}

func (x *TokenPricesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pricegetter_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_pricegetter_proto_rawDescGZIP(), []int{1}
}

func (x *TokenPricesResponse) GetPrices() map[string]*pb.BigInt {
	if x != nil {
		return x.Prices
	}
	return nil
}

// TokenConfiguredRequest is a request for if a token address is configured to be able to get a price. It is a gRPC adapter for the input arguments of
// [github.com/smartcontractkit/chainlink-common/chainlink-common/pkg/types/ccip/PriceGetter.IsTokenConfigured]]
type TokenConfiguredRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"` // Address
}

func (x *TokenConfiguredRequest) Reset() {
	*x = TokenConfiguredRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pricegetter_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TokenConfiguredRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenConfiguredRequest) ProtoMessage() {}

func (x *TokenConfiguredRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pricegetter_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenConfiguredRequest.ProtoReflect.Descriptor instead.
func (*TokenConfiguredRequest) Descriptor() ([]byte, []int) {
	return file_pricegetter_proto_rawDescGZIP(), []int{2}
}

func (x *TokenConfiguredRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

// TokenConfiguredResponse is a response for if a token address is configured to be able to get a price. It is a gRPC adapter for the return values of
// [github.com/smartcontractkit/chainlink-common/chainlink-common/pkg/types/ccip/PriceGetter.IsTokenConfigured]]
type TokenConfiguredResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsConfigured bool `protobuf:"varint,1,opt,name=isConfigured,proto3" json:"isConfigured,omitempty"` // bool
}

func (x *TokenConfiguredResponse) Reset() {
	*x = TokenConfiguredResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pricegetter_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TokenConfiguredResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenConfiguredResponse) ProtoMessage() {}

func (x *TokenConfiguredResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pricegetter_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenConfiguredResponse.ProtoReflect.Descriptor instead.
func (*TokenConfiguredResponse) Descriptor() ([]byte, []int) {
	return file_pricegetter_proto_rawDescGZIP(), []int{3}
}

func (x *TokenConfiguredResponse) GetIsConfigured() bool {
	if x != nil {
		return x.IsConfigured
	}
	return false
}

var File_pricegetter_proto protoreflect.FileDescriptor

var file_pricegetter_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x72, 0x69, 0x63, 0x65, 0x67, 0x65, 0x74, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x15, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x63, 0x63, 0x69, 0x70, 0x1a, 0x0d, 0x72, 0x65, 0x6c, 0x61,
	0x79, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2c, 0x0a, 0x12, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x50,
	0x72, 0x69, 0x63, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x74, 0x6f,
	0x6b, 0x65, 0x6e, 0x73, 0x22, 0xae, 0x01, 0x0a, 0x13, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x50, 0x72,
	0x69, 0x63, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4e, 0x0a, 0x06,
	0x70, 0x72, 0x69, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x36, 0x2e, 0x6c,
	0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e,
	0x63, 0x63, 0x69, 0x70, 0x2e, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x50, 0x72, 0x69, 0x63, 0x65, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x50, 0x72, 0x69, 0x63, 0x65, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x70, 0x72, 0x69, 0x63, 0x65, 0x73, 0x1a, 0x47, 0x0a, 0x0b,
	0x50, 0x72, 0x69, 0x63, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x22, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6c,
	0x6f, 0x6f, 0x70, 0x2e, 0x42, 0x69, 0x67, 0x49, 0x6e, 0x74, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x2e, 0x0a, 0x16, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x3d, 0x0a, 0x17, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x22, 0x0a, 0x0c, 0x69, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x69, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x75, 0x72, 0x65, 0x64, 0x32, 0xa9, 0x02, 0x0a, 0x0b, 0x50, 0x72, 0x69, 0x63, 0x65, 0x47, 0x65,
	0x74, 0x74, 0x65, 0x72, 0x12, 0x69, 0x0a, 0x0e, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x50, 0x72, 0x69,
	0x63, 0x65, 0x73, 0x55, 0x53, 0x44, 0x12, 0x29, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x63, 0x63, 0x69, 0x70, 0x2e, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x50, 0x72, 0x69, 0x63, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x2a, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x63, 0x63, 0x69, 0x70, 0x2e, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x50,
	0x72, 0x69, 0x63, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x74, 0x0a, 0x11, 0x49, 0x73, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x75, 0x72, 0x65, 0x64, 0x12, 0x2d, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x63, 0x63, 0x69, 0x70, 0x2e, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x63, 0x63, 0x69, 0x70, 0x2e, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x39, 0x0a, 0x05, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x12, 0x16,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00,
	0x42, 0x4f, 0x5a, 0x4d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73,
	0x6d, 0x61, 0x72, 0x74, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6b, 0x69, 0x74, 0x2f,
	0x63, 0x68, 0x61, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6c, 0x6f, 0x6f, 0x70, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2f, 0x70, 0x62, 0x2f, 0x63, 0x63, 0x69, 0x70, 0x3b, 0x63, 0x63, 0x69, 0x70, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pricegetter_proto_rawDescOnce sync.Once
	file_pricegetter_proto_rawDescData = file_pricegetter_proto_rawDesc
)

func file_pricegetter_proto_rawDescGZIP() []byte {
	file_pricegetter_proto_rawDescOnce.Do(func() {
		file_pricegetter_proto_rawDescData = protoimpl.X.CompressGZIP(file_pricegetter_proto_rawDescData)
	})
	return file_pricegetter_proto_rawDescData
}

var file_pricegetter_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pricegetter_proto_goTypes = []interface{}{
	(*TokenPricesRequest)(nil),      // 0: loop.internal.pb.ccip.TokenPricesRequest
	(*TokenPricesResponse)(nil),     // 1: loop.internal.pb.ccip.TokenPricesResponse
	(*TokenConfiguredRequest)(nil),  // 2: loop.internal.pb.ccip.TokenConfiguredRequest
	(*TokenConfiguredResponse)(nil), // 3: loop.internal.pb.ccip.TokenConfiguredResponse
	nil,                             // 4: loop.internal.pb.ccip.TokenPricesResponse.PricesEntry
	(*pb.BigInt)(nil),               // 5: loop.BigInt
	(*emptypb.Empty)(nil),           // 6: google.protobuf.Empty
}
var file_pricegetter_proto_depIdxs = []int32{
	4, // 0: loop.internal.pb.ccip.TokenPricesResponse.prices:type_name -> loop.internal.pb.ccip.TokenPricesResponse.PricesEntry
	5, // 1: loop.internal.pb.ccip.TokenPricesResponse.PricesEntry.value:type_name -> loop.BigInt
	0, // 2: loop.internal.pb.ccip.PriceGetter.TokenPricesUSD:input_type -> loop.internal.pb.ccip.TokenPricesRequest
	2, // 3: loop.internal.pb.ccip.PriceGetter.IsTokenConfigured:input_type -> loop.internal.pb.ccip.TokenConfiguredRequest
	6, // 4: loop.internal.pb.ccip.PriceGetter.Close:input_type -> google.protobuf.Empty
	1, // 5: loop.internal.pb.ccip.PriceGetter.TokenPricesUSD:output_type -> loop.internal.pb.ccip.TokenPricesResponse
	3, // 6: loop.internal.pb.ccip.PriceGetter.IsTokenConfigured:output_type -> loop.internal.pb.ccip.TokenConfiguredResponse
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
	if !protoimpl.UnsafeEnabled {
		file_pricegetter_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TokenPricesRequest); i {
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
		file_pricegetter_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TokenPricesResponse); i {
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
		file_pricegetter_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TokenConfiguredRequest); i {
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
		file_pricegetter_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TokenConfiguredResponse); i {
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
			RawDescriptor: file_pricegetter_proto_rawDesc,
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
	file_pricegetter_proto_rawDesc = nil
	file_pricegetter_proto_goTypes = nil
	file_pricegetter_proto_depIdxs = nil
}
