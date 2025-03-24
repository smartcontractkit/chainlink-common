// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: oraclefactory.proto

package oraclefactory

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
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

type LocalConfig struct {
	state                              protoimpl.MessageState `protogen:"open.v1"`
	BlockchainTimeout                  *durationpb.Duration   `protobuf:"bytes,1,opt,name=blockchain_timeout,json=blockchainTimeout,proto3" json:"blockchain_timeout,omitempty"`
	ContractConfigConfirmations        uint32                 `protobuf:"varint,2,opt,name=contract_config_confirmations,json=contractConfigConfirmations,proto3" json:"contract_config_confirmations,omitempty"`
	SkipContractConfigConfirmations    bool                   `protobuf:"varint,3,opt,name=skip_contract_config_confirmations,json=skipContractConfigConfirmations,proto3" json:"skip_contract_config_confirmations,omitempty"`
	ContractConfigTrackerPollInterval  *durationpb.Duration   `protobuf:"bytes,4,opt,name=contract_config_tracker_poll_interval,json=contractConfigTrackerPollInterval,proto3" json:"contract_config_tracker_poll_interval,omitempty"`
	ContractTransmitterTransmitTimeout *durationpb.Duration   `protobuf:"bytes,5,opt,name=contract_transmitter_transmit_timeout,json=contractTransmitterTransmitTimeout,proto3" json:"contract_transmitter_transmit_timeout,omitempty"`
	DatabaseTimeout                    *durationpb.Duration   `protobuf:"bytes,6,opt,name=database_timeout,json=databaseTimeout,proto3" json:"database_timeout,omitempty"`
	MinOcr2MaxDurationQuery            *durationpb.Duration   `protobuf:"bytes,7,opt,name=min_ocr2_max_duration_query,json=minOcr2MaxDurationQuery,proto3" json:"min_ocr2_max_duration_query,omitempty"`
	DevelopmentMode                    string                 `protobuf:"bytes,8,opt,name=development_mode,json=developmentMode,proto3" json:"development_mode,omitempty"`
	unknownFields                      protoimpl.UnknownFields
	sizeCache                          protoimpl.SizeCache
}

func (x *LocalConfig) Reset() {
	*x = LocalConfig{}
	mi := &file_oraclefactory_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LocalConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LocalConfig) ProtoMessage() {}

func (x *LocalConfig) ProtoReflect() protoreflect.Message {
	mi := &file_oraclefactory_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LocalConfig.ProtoReflect.Descriptor instead.
func (*LocalConfig) Descriptor() ([]byte, []int) {
	return file_oraclefactory_proto_rawDescGZIP(), []int{0}
}

func (x *LocalConfig) GetBlockchainTimeout() *durationpb.Duration {
	if x != nil {
		return x.BlockchainTimeout
	}
	return nil
}

func (x *LocalConfig) GetContractConfigConfirmations() uint32 {
	if x != nil {
		return x.ContractConfigConfirmations
	}
	return 0
}

func (x *LocalConfig) GetSkipContractConfigConfirmations() bool {
	if x != nil {
		return x.SkipContractConfigConfirmations
	}
	return false
}

func (x *LocalConfig) GetContractConfigTrackerPollInterval() *durationpb.Duration {
	if x != nil {
		return x.ContractConfigTrackerPollInterval
	}
	return nil
}

func (x *LocalConfig) GetContractTransmitterTransmitTimeout() *durationpb.Duration {
	if x != nil {
		return x.ContractTransmitterTransmitTimeout
	}
	return nil
}

func (x *LocalConfig) GetDatabaseTimeout() *durationpb.Duration {
	if x != nil {
		return x.DatabaseTimeout
	}
	return nil
}

func (x *LocalConfig) GetMinOcr2MaxDurationQuery() *durationpb.Duration {
	if x != nil {
		return x.MinOcr2MaxDurationQuery
	}
	return nil
}

func (x *LocalConfig) GetDevelopmentMode() string {
	if x != nil {
		return x.DevelopmentMode
	}
	return ""
}

type NewOracleRequest struct {
	state                           protoimpl.MessageState `protogen:"open.v1"`
	LocalConfig                     *LocalConfig           `protobuf:"bytes,1,opt,name=local_config,json=localConfig,proto3" json:"local_config,omitempty"`
	ReportingPluginFactoryServiceId uint32                 `protobuf:"varint,2,opt,name=reporting_plugin_factory_service_id,json=reportingPluginFactoryServiceId,proto3" json:"reporting_plugin_factory_service_id,omitempty"`
	ContractConfigTrackerId         uint32                 `protobuf:"varint,3,opt,name=contract_config_tracker_id,json=contractConfigTrackerId,proto3" json:"contract_config_tracker_id,omitempty"`
	ContractTransmitterId           uint32                 `protobuf:"varint,4,opt,name=contract_transmitter_id,json=contractTransmitterId,proto3" json:"contract_transmitter_id,omitempty"`
	OffchainConfigDigesterId        uint32                 `protobuf:"varint,5,opt,name=offchain_config_digester_id,json=offchainConfigDigesterId,proto3" json:"offchain_config_digester_id,omitempty"`
	unknownFields                   protoimpl.UnknownFields
	sizeCache                       protoimpl.SizeCache
}

func (x *NewOracleRequest) Reset() {
	*x = NewOracleRequest{}
	mi := &file_oraclefactory_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NewOracleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewOracleRequest) ProtoMessage() {}

func (x *NewOracleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_oraclefactory_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewOracleRequest.ProtoReflect.Descriptor instead.
func (*NewOracleRequest) Descriptor() ([]byte, []int) {
	return file_oraclefactory_proto_rawDescGZIP(), []int{1}
}

func (x *NewOracleRequest) GetLocalConfig() *LocalConfig {
	if x != nil {
		return x.LocalConfig
	}
	return nil
}

func (x *NewOracleRequest) GetReportingPluginFactoryServiceId() uint32 {
	if x != nil {
		return x.ReportingPluginFactoryServiceId
	}
	return 0
}

func (x *NewOracleRequest) GetContractConfigTrackerId() uint32 {
	if x != nil {
		return x.ContractConfigTrackerId
	}
	return 0
}

func (x *NewOracleRequest) GetContractTransmitterId() uint32 {
	if x != nil {
		return x.ContractTransmitterId
	}
	return 0
}

func (x *NewOracleRequest) GetOffchainConfigDigesterId() uint32 {
	if x != nil {
		return x.OffchainConfigDigesterId
	}
	return 0
}

type NewOracleReply struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	OracleId      uint32                 `protobuf:"varint,1,opt,name=oracle_id,json=oracleId,proto3" json:"oracle_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *NewOracleReply) Reset() {
	*x = NewOracleReply{}
	mi := &file_oraclefactory_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NewOracleReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewOracleReply) ProtoMessage() {}

func (x *NewOracleReply) ProtoReflect() protoreflect.Message {
	mi := &file_oraclefactory_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewOracleReply.ProtoReflect.Descriptor instead.
func (*NewOracleReply) Descriptor() ([]byte, []int) {
	return file_oraclefactory_proto_rawDescGZIP(), []int{2}
}

func (x *NewOracleReply) GetOracleId() uint32 {
	if x != nil {
		return x.OracleId
	}
	return 0
}

var File_oraclefactory_proto protoreflect.FileDescriptor

const file_oraclefactory_proto_rawDesc = "" +
	"\n" +
	"\x13oraclefactory.proto\x12\x04loop\x1a\x1egoogle/protobuf/duration.proto\"\x8d\x05\n" +
	"\vLocalConfig\x12H\n" +
	"\x12blockchain_timeout\x18\x01 \x01(\v2\x19.google.protobuf.DurationR\x11blockchainTimeout\x12B\n" +
	"\x1dcontract_config_confirmations\x18\x02 \x01(\rR\x1bcontractConfigConfirmations\x12K\n" +
	"\"skip_contract_config_confirmations\x18\x03 \x01(\bR\x1fskipContractConfigConfirmations\x12k\n" +
	"%contract_config_tracker_poll_interval\x18\x04 \x01(\v2\x19.google.protobuf.DurationR!contractConfigTrackerPollInterval\x12l\n" +
	"%contract_transmitter_transmit_timeout\x18\x05 \x01(\v2\x19.google.protobuf.DurationR\"contractTransmitterTransmitTimeout\x12D\n" +
	"\x10database_timeout\x18\x06 \x01(\v2\x19.google.protobuf.DurationR\x0fdatabaseTimeout\x12W\n" +
	"\x1bmin_ocr2_max_duration_query\x18\a \x01(\v2\x19.google.protobuf.DurationR\x17minOcr2MaxDurationQuery\x12)\n" +
	"\x10development_mode\x18\b \x01(\tR\x0fdevelopmentMode\"\xca\x02\n" +
	"\x10NewOracleRequest\x124\n" +
	"\flocal_config\x18\x01 \x01(\v2\x11.loop.LocalConfigR\vlocalConfig\x12L\n" +
	"#reporting_plugin_factory_service_id\x18\x02 \x01(\rR\x1freportingPluginFactoryServiceId\x12;\n" +
	"\x1acontract_config_tracker_id\x18\x03 \x01(\rR\x17contractConfigTrackerId\x126\n" +
	"\x17contract_transmitter_id\x18\x04 \x01(\rR\x15contractTransmitterId\x12=\n" +
	"\x1boffchain_config_digester_id\x18\x05 \x01(\rR\x18offchainConfigDigesterId\"-\n" +
	"\x0eNewOracleReply\x12\x1b\n" +
	"\toracle_id\x18\x01 \x01(\rR\boracleId2L\n" +
	"\rOracleFactory\x12;\n" +
	"\tNewOracle\x12\x16.loop.NewOracleRequest\x1a\x14.loop.NewOracleReply\"\x00BNZLgithub.com/smartcontractkit/chainlink-common/pkg/loop/internal/oraclefactoryb\x06proto3"

var (
	file_oraclefactory_proto_rawDescOnce sync.Once
	file_oraclefactory_proto_rawDescData []byte
)

func file_oraclefactory_proto_rawDescGZIP() []byte {
	file_oraclefactory_proto_rawDescOnce.Do(func() {
		file_oraclefactory_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_oraclefactory_proto_rawDesc), len(file_oraclefactory_proto_rawDesc)))
	})
	return file_oraclefactory_proto_rawDescData
}

var file_oraclefactory_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_oraclefactory_proto_goTypes = []any{
	(*LocalConfig)(nil),         // 0: loop.LocalConfig
	(*NewOracleRequest)(nil),    // 1: loop.NewOracleRequest
	(*NewOracleReply)(nil),      // 2: loop.NewOracleReply
	(*durationpb.Duration)(nil), // 3: google.protobuf.Duration
}
var file_oraclefactory_proto_depIdxs = []int32{
	3, // 0: loop.LocalConfig.blockchain_timeout:type_name -> google.protobuf.Duration
	3, // 1: loop.LocalConfig.contract_config_tracker_poll_interval:type_name -> google.protobuf.Duration
	3, // 2: loop.LocalConfig.contract_transmitter_transmit_timeout:type_name -> google.protobuf.Duration
	3, // 3: loop.LocalConfig.database_timeout:type_name -> google.protobuf.Duration
	3, // 4: loop.LocalConfig.min_ocr2_max_duration_query:type_name -> google.protobuf.Duration
	0, // 5: loop.NewOracleRequest.local_config:type_name -> loop.LocalConfig
	1, // 6: loop.OracleFactory.NewOracle:input_type -> loop.NewOracleRequest
	2, // 7: loop.OracleFactory.NewOracle:output_type -> loop.NewOracleReply
	7, // [7:8] is the sub-list for method output_type
	6, // [6:7] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_oraclefactory_proto_init() }
func file_oraclefactory_proto_init() {
	if File_oraclefactory_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_oraclefactory_proto_rawDesc), len(file_oraclefactory_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_oraclefactory_proto_goTypes,
		DependencyIndexes: file_oraclefactory_proto_depIdxs,
		MessageInfos:      file_oraclefactory_proto_msgTypes,
	}.Build()
	File_oraclefactory_proto = out.File
	file_oraclefactory_proto_goTypes = nil
	file_oraclefactory_proto_depIdxs = nil
}
