// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: metering/pb/meteringdetail.proto

package pb

import (
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

// MeteringReportNodeDetail is consumed by
// capability responses and by Metering Report.
// It currently lives in its own file due to a
// restriction in proto registration for Beholder.
type MeteringReportNodeDetail struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Peer_2PeerId  string                 `protobuf:"bytes,1,opt,name=peer_2_peer_id,json=peer2PeerId,proto3" json:"peer_2_peer_id,omitempty"`
	SpendUnit     string                 `protobuf:"bytes,2,opt,name=spend_unit,json=spendUnit,proto3" json:"spend_unit,omitempty"`
	SpendValue    string                 `protobuf:"bytes,3,opt,name=spend_value,json=spendValue,proto3" json:"spend_value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MeteringReportNodeDetail) Reset() {
	*x = MeteringReportNodeDetail{}
	mi := &file_metering_pb_meteringdetail_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MeteringReportNodeDetail) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeteringReportNodeDetail) ProtoMessage() {}

func (x *MeteringReportNodeDetail) ProtoReflect() protoreflect.Message {
	mi := &file_metering_pb_meteringdetail_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeteringReportNodeDetail.ProtoReflect.Descriptor instead.
func (*MeteringReportNodeDetail) Descriptor() ([]byte, []int) {
	return file_metering_pb_meteringdetail_proto_rawDescGZIP(), []int{0}
}

func (x *MeteringReportNodeDetail) GetPeer_2PeerId() string {
	if x != nil {
		return x.Peer_2PeerId
	}
	return ""
}

func (x *MeteringReportNodeDetail) GetSpendUnit() string {
	if x != nil {
		return x.SpendUnit
	}
	return ""
}

func (x *MeteringReportNodeDetail) GetSpendValue() string {
	if x != nil {
		return x.SpendValue
	}
	return ""
}

var File_metering_pb_meteringdetail_proto protoreflect.FileDescriptor

const file_metering_pb_meteringdetail_proto_rawDesc = "" +
	"\n" +
	" metering/pb/meteringdetail.proto\x12\bmetering\"\x7f\n" +
	"\x18MeteringReportNodeDetail\x12#\n" +
	"\x0epeer_2_peer_id\x18\x01 \x01(\tR\vpeer2PeerId\x12\x1d\n" +
	"\n" +
	"spend_unit\x18\x02 \x01(\tR\tspendUnit\x12\x1f\n" +
	"\vspend_value\x18\x03 \x01(\tR\n" +
	"spendValueB>Z<github.com/smartcontractkit/chainlink-common/pkg/metering/pbb\x06proto3"

var (
	file_metering_pb_meteringdetail_proto_rawDescOnce sync.Once
	file_metering_pb_meteringdetail_proto_rawDescData []byte
)

func file_metering_pb_meteringdetail_proto_rawDescGZIP() []byte {
	file_metering_pb_meteringdetail_proto_rawDescOnce.Do(func() {
		file_metering_pb_meteringdetail_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_metering_pb_meteringdetail_proto_rawDesc), len(file_metering_pb_meteringdetail_proto_rawDesc)))
	})
	return file_metering_pb_meteringdetail_proto_rawDescData
}

var file_metering_pb_meteringdetail_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_metering_pb_meteringdetail_proto_goTypes = []any{
	(*MeteringReportNodeDetail)(nil), // 0: metering.MeteringReportNodeDetail
}
var file_metering_pb_meteringdetail_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_metering_pb_meteringdetail_proto_init() }
func file_metering_pb_meteringdetail_proto_init() {
	if File_metering_pb_meteringdetail_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_metering_pb_meteringdetail_proto_rawDesc), len(file_metering_pb_meteringdetail_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_metering_pb_meteringdetail_proto_goTypes,
		DependencyIndexes: file_metering_pb_meteringdetail_proto_depIdxs,
		MessageInfos:      file_metering_pb_meteringdetail_proto_msgTypes,
	}.Build()
	File_metering_pb_meteringdetail_proto = out.File
	file_metering_pb_meteringdetail_proto_goTypes = nil
	file_metering_pb_meteringdetail_proto_depIdxs = nil
}
