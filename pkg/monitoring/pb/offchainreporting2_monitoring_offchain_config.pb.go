// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: offchainreporting2_monitoring_offchain_config.proto

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

type OffchainConfigProto struct {
	state                                              protoimpl.MessageState        `protogen:"open.v1"`
	DeltaProgressNanoseconds                           uint64                        `protobuf:"varint,1,opt,name=delta_progress_nanoseconds,json=deltaProgressNanoseconds,proto3" json:"delta_progress_nanoseconds,omitempty"`
	DeltaResendNanoseconds                             uint64                        `protobuf:"varint,2,opt,name=delta_resend_nanoseconds,json=deltaResendNanoseconds,proto3" json:"delta_resend_nanoseconds,omitempty"`
	DeltaRoundNanoseconds                              uint64                        `protobuf:"varint,3,opt,name=delta_round_nanoseconds,json=deltaRoundNanoseconds,proto3" json:"delta_round_nanoseconds,omitempty"`
	DeltaGraceNanoseconds                              uint64                        `protobuf:"varint,4,opt,name=delta_grace_nanoseconds,json=deltaGraceNanoseconds,proto3" json:"delta_grace_nanoseconds,omitempty"`
	DeltaStageNanoseconds                              uint64                        `protobuf:"varint,5,opt,name=delta_stage_nanoseconds,json=deltaStageNanoseconds,proto3" json:"delta_stage_nanoseconds,omitempty"`
	RMax                                               uint32                        `protobuf:"varint,6,opt,name=r_max,json=rMax,proto3" json:"r_max,omitempty"`
	S                                                  []uint32                      `protobuf:"varint,7,rep,packed,name=s,proto3" json:"s,omitempty"`
	OffchainPublicKeys                                 [][]byte                      `protobuf:"bytes,8,rep,name=offchain_public_keys,json=offchainPublicKeys,proto3" json:"offchain_public_keys,omitempty"`
	PeerIds                                            []string                      `protobuf:"bytes,9,rep,name=peer_ids,json=peerIds,proto3" json:"peer_ids,omitempty"`
	ReportingPluginConfig                              []byte                        `protobuf:"bytes,10,opt,name=reporting_plugin_config,json=reportingPluginConfig,proto3" json:"reporting_plugin_config,omitempty"`
	MaxDurationQueryNanoseconds                        uint64                        `protobuf:"varint,11,opt,name=max_duration_query_nanoseconds,json=maxDurationQueryNanoseconds,proto3" json:"max_duration_query_nanoseconds,omitempty"`
	MaxDurationObservationNanoseconds                  uint64                        `protobuf:"varint,12,opt,name=max_duration_observation_nanoseconds,json=maxDurationObservationNanoseconds,proto3" json:"max_duration_observation_nanoseconds,omitempty"`
	MaxDurationReportNanoseconds                       uint64                        `protobuf:"varint,13,opt,name=max_duration_report_nanoseconds,json=maxDurationReportNanoseconds,proto3" json:"max_duration_report_nanoseconds,omitempty"`
	MaxDurationShouldAcceptFinalizedReportNanoseconds  uint64                        `protobuf:"varint,14,opt,name=max_duration_should_accept_finalized_report_nanoseconds,json=maxDurationShouldAcceptFinalizedReportNanoseconds,proto3" json:"max_duration_should_accept_finalized_report_nanoseconds,omitempty"`
	MaxDurationShouldTransmitAcceptedReportNanoseconds uint64                        `protobuf:"varint,15,opt,name=max_duration_should_transmit_accepted_report_nanoseconds,json=maxDurationShouldTransmitAcceptedReportNanoseconds,proto3" json:"max_duration_should_transmit_accepted_report_nanoseconds,omitempty"`
	SharedSecretEncryptions                            *SharedSecretEncryptionsProto `protobuf:"bytes,16,opt,name=shared_secret_encryptions,json=sharedSecretEncryptions,proto3" json:"shared_secret_encryptions,omitempty"`
	unknownFields                                      protoimpl.UnknownFields
	sizeCache                                          protoimpl.SizeCache
}

func (x *OffchainConfigProto) Reset() {
	*x = OffchainConfigProto{}
	mi := &file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OffchainConfigProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OffchainConfigProto) ProtoMessage() {}

func (x *OffchainConfigProto) ProtoReflect() protoreflect.Message {
	mi := &file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OffchainConfigProto.ProtoReflect.Descriptor instead.
func (*OffchainConfigProto) Descriptor() ([]byte, []int) {
	return file_offchainreporting2_monitoring_offchain_config_proto_rawDescGZIP(), []int{0}
}

func (x *OffchainConfigProto) GetDeltaProgressNanoseconds() uint64 {
	if x != nil {
		return x.DeltaProgressNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetDeltaResendNanoseconds() uint64 {
	if x != nil {
		return x.DeltaResendNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetDeltaRoundNanoseconds() uint64 {
	if x != nil {
		return x.DeltaRoundNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetDeltaGraceNanoseconds() uint64 {
	if x != nil {
		return x.DeltaGraceNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetDeltaStageNanoseconds() uint64 {
	if x != nil {
		return x.DeltaStageNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetRMax() uint32 {
	if x != nil {
		return x.RMax
	}
	return 0
}

func (x *OffchainConfigProto) GetS() []uint32 {
	if x != nil {
		return x.S
	}
	return nil
}

func (x *OffchainConfigProto) GetOffchainPublicKeys() [][]byte {
	if x != nil {
		return x.OffchainPublicKeys
	}
	return nil
}

func (x *OffchainConfigProto) GetPeerIds() []string {
	if x != nil {
		return x.PeerIds
	}
	return nil
}

func (x *OffchainConfigProto) GetReportingPluginConfig() []byte {
	if x != nil {
		return x.ReportingPluginConfig
	}
	return nil
}

func (x *OffchainConfigProto) GetMaxDurationQueryNanoseconds() uint64 {
	if x != nil {
		return x.MaxDurationQueryNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetMaxDurationObservationNanoseconds() uint64 {
	if x != nil {
		return x.MaxDurationObservationNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetMaxDurationReportNanoseconds() uint64 {
	if x != nil {
		return x.MaxDurationReportNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetMaxDurationShouldAcceptFinalizedReportNanoseconds() uint64 {
	if x != nil {
		return x.MaxDurationShouldAcceptFinalizedReportNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetMaxDurationShouldTransmitAcceptedReportNanoseconds() uint64 {
	if x != nil {
		return x.MaxDurationShouldTransmitAcceptedReportNanoseconds
	}
	return 0
}

func (x *OffchainConfigProto) GetSharedSecretEncryptions() *SharedSecretEncryptionsProto {
	if x != nil {
		return x.SharedSecretEncryptions
	}
	return nil
}

type SharedSecretEncryptionsProto struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	DiffieHellmanPoint []byte                 `protobuf:"bytes,1,opt,name=diffieHellmanPoint,proto3" json:"diffieHellmanPoint,omitempty"`
	SharedSecretHash   []byte                 `protobuf:"bytes,2,opt,name=sharedSecretHash,proto3" json:"sharedSecretHash,omitempty"`
	Encryptions        [][]byte               `protobuf:"bytes,3,rep,name=encryptions,proto3" json:"encryptions,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *SharedSecretEncryptionsProto) Reset() {
	*x = SharedSecretEncryptionsProto{}
	mi := &file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SharedSecretEncryptionsProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SharedSecretEncryptionsProto) ProtoMessage() {}

func (x *SharedSecretEncryptionsProto) ProtoReflect() protoreflect.Message {
	mi := &file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SharedSecretEncryptionsProto.ProtoReflect.Descriptor instead.
func (*SharedSecretEncryptionsProto) Descriptor() ([]byte, []int) {
	return file_offchainreporting2_monitoring_offchain_config_proto_rawDescGZIP(), []int{1}
}

func (x *SharedSecretEncryptionsProto) GetDiffieHellmanPoint() []byte {
	if x != nil {
		return x.DiffieHellmanPoint
	}
	return nil
}

func (x *SharedSecretEncryptionsProto) GetSharedSecretHash() []byte {
	if x != nil {
		return x.SharedSecretHash
	}
	return nil
}

func (x *SharedSecretEncryptionsProto) GetEncryptions() [][]byte {
	if x != nil {
		return x.Encryptions
	}
	return nil
}

var File_offchainreporting2_monitoring_offchain_config_proto protoreflect.FileDescriptor

const file_offchainreporting2_monitoring_offchain_config_proto_rawDesc = "" +
	"\n" +
	"3offchainreporting2_monitoring_offchain_config.proto\x12\rmonitoring_pb\"\x8d\b\n" +
	"\x13OffchainConfigProto\x12<\n" +
	"\x1adelta_progress_nanoseconds\x18\x01 \x01(\x04R\x18deltaProgressNanoseconds\x128\n" +
	"\x18delta_resend_nanoseconds\x18\x02 \x01(\x04R\x16deltaResendNanoseconds\x126\n" +
	"\x17delta_round_nanoseconds\x18\x03 \x01(\x04R\x15deltaRoundNanoseconds\x126\n" +
	"\x17delta_grace_nanoseconds\x18\x04 \x01(\x04R\x15deltaGraceNanoseconds\x126\n" +
	"\x17delta_stage_nanoseconds\x18\x05 \x01(\x04R\x15deltaStageNanoseconds\x12\x13\n" +
	"\x05r_max\x18\x06 \x01(\rR\x04rMax\x12\f\n" +
	"\x01s\x18\a \x03(\rR\x01s\x120\n" +
	"\x14offchain_public_keys\x18\b \x03(\fR\x12offchainPublicKeys\x12\x19\n" +
	"\bpeer_ids\x18\t \x03(\tR\apeerIds\x126\n" +
	"\x17reporting_plugin_config\x18\n" +
	" \x01(\fR\x15reportingPluginConfig\x12C\n" +
	"\x1emax_duration_query_nanoseconds\x18\v \x01(\x04R\x1bmaxDurationQueryNanoseconds\x12O\n" +
	"$max_duration_observation_nanoseconds\x18\f \x01(\x04R!maxDurationObservationNanoseconds\x12E\n" +
	"\x1fmax_duration_report_nanoseconds\x18\r \x01(\x04R\x1cmaxDurationReportNanoseconds\x12r\n" +
	"7max_duration_should_accept_finalized_report_nanoseconds\x18\x0e \x01(\x04R1maxDurationShouldAcceptFinalizedReportNanoseconds\x12t\n" +
	"8max_duration_should_transmit_accepted_report_nanoseconds\x18\x0f \x01(\x04R2maxDurationShouldTransmitAcceptedReportNanoseconds\x12g\n" +
	"\x19shared_secret_encryptions\x18\x10 \x01(\v2+.monitoring_pb.SharedSecretEncryptionsProtoR\x17sharedSecretEncryptions\"\x9c\x01\n" +
	"\x1cSharedSecretEncryptionsProto\x12.\n" +
	"\x12diffieHellmanPoint\x18\x01 \x01(\fR\x12diffieHellmanPoint\x12*\n" +
	"\x10sharedSecretHash\x18\x02 \x01(\fR\x10sharedSecretHash\x12 \n" +
	"\vencryptions\x18\x03 \x03(\fR\vencryptionsB\x06Z\x04.;pbb\x06proto3"

var (
	file_offchainreporting2_monitoring_offchain_config_proto_rawDescOnce sync.Once
	file_offchainreporting2_monitoring_offchain_config_proto_rawDescData []byte
)

func file_offchainreporting2_monitoring_offchain_config_proto_rawDescGZIP() []byte {
	file_offchainreporting2_monitoring_offchain_config_proto_rawDescOnce.Do(func() {
		file_offchainreporting2_monitoring_offchain_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_offchainreporting2_monitoring_offchain_config_proto_rawDesc), len(file_offchainreporting2_monitoring_offchain_config_proto_rawDesc)))
	})
	return file_offchainreporting2_monitoring_offchain_config_proto_rawDescData
}

var file_offchainreporting2_monitoring_offchain_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_offchainreporting2_monitoring_offchain_config_proto_goTypes = []any{
	(*OffchainConfigProto)(nil),          // 0: monitoring_pb.OffchainConfigProto
	(*SharedSecretEncryptionsProto)(nil), // 1: monitoring_pb.SharedSecretEncryptionsProto
}
var file_offchainreporting2_monitoring_offchain_config_proto_depIdxs = []int32{
	1, // 0: monitoring_pb.OffchainConfigProto.shared_secret_encryptions:type_name -> monitoring_pb.SharedSecretEncryptionsProto
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_offchainreporting2_monitoring_offchain_config_proto_init() }
func file_offchainreporting2_monitoring_offchain_config_proto_init() {
	if File_offchainreporting2_monitoring_offchain_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_offchainreporting2_monitoring_offchain_config_proto_rawDesc), len(file_offchainreporting2_monitoring_offchain_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_offchainreporting2_monitoring_offchain_config_proto_goTypes,
		DependencyIndexes: file_offchainreporting2_monitoring_offchain_config_proto_depIdxs,
		MessageInfos:      file_offchainreporting2_monitoring_offchain_config_proto_msgTypes,
	}.Build()
	File_offchainreporting2_monitoring_offchain_config_proto = out.File
	file_offchainreporting2_monitoring_offchain_config_proto_goTypes = nil
	file_offchainreporting2_monitoring_offchain_config_proto_depIdxs = nil
}
