// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: offchainreporting2_monitoring_offchain_config.proto

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

type OffchainConfigProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

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
}

func (x *OffchainConfigProto) Reset() {
	*x = OffchainConfigProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OffchainConfigProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OffchainConfigProto) ProtoMessage() {}

func (x *OffchainConfigProto) ProtoReflect() protoreflect.Message {
	mi := &file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
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
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DiffieHellmanPoint []byte   `protobuf:"bytes,1,opt,name=diffieHellmanPoint,proto3" json:"diffieHellmanPoint,omitempty"`
	SharedSecretHash   []byte   `protobuf:"bytes,2,opt,name=sharedSecretHash,proto3" json:"sharedSecretHash,omitempty"`
	Encryptions        [][]byte `protobuf:"bytes,3,rep,name=encryptions,proto3" json:"encryptions,omitempty"`
}

func (x *SharedSecretEncryptionsProto) Reset() {
	*x = SharedSecretEncryptionsProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SharedSecretEncryptionsProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SharedSecretEncryptionsProto) ProtoMessage() {}

func (x *SharedSecretEncryptionsProto) ProtoReflect() protoreflect.Message {
	mi := &file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
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

var file_offchainreporting2_monitoring_offchain_config_proto_rawDesc = []byte{
	0x0a, 0x33, 0x6f, 0x66, 0x66, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x69, 0x6e, 0x67, 0x32, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x5f,
	0x6f, 0x66, 0x66, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e,
	0x67, 0x5f, 0x70, 0x62, 0x22, 0x8d, 0x08, 0x0a, 0x13, 0x4f, 0x66, 0x66, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x3c, 0x0a, 0x1a,
	0x64, 0x65, 0x6c, 0x74, 0x61, 0x5f, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x5f, 0x6e,
	0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x18, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x50, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x4e,
	0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x38, 0x0a, 0x18, 0x64, 0x65,
	0x6c, 0x74, 0x61, 0x5f, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x64, 0x5f, 0x6e, 0x61, 0x6e, 0x6f, 0x73,
	0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x16, 0x64, 0x65,
	0x6c, 0x74, 0x61, 0x52, 0x65, 0x73, 0x65, 0x6e, 0x64, 0x4e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63,
	0x6f, 0x6e, 0x64, 0x73, 0x12, 0x36, 0x0a, 0x17, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x5f, 0x72, 0x6f,
	0x75, 0x6e, 0x64, 0x5f, 0x6e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x15, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x52, 0x6f, 0x75, 0x6e,
	0x64, 0x4e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x36, 0x0a, 0x17,
	0x64, 0x65, 0x6c, 0x74, 0x61, 0x5f, 0x67, 0x72, 0x61, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6e, 0x6f,
	0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x15, 0x64,
	0x65, 0x6c, 0x74, 0x61, 0x47, 0x72, 0x61, 0x63, 0x65, 0x4e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63,
	0x6f, 0x6e, 0x64, 0x73, 0x12, 0x36, 0x0a, 0x17, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x5f, 0x73, 0x74,
	0x61, 0x67, 0x65, 0x5f, 0x6e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x15, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x53, 0x74, 0x61, 0x67,
	0x65, 0x4e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x13, 0x0a, 0x05,
	0x72, 0x5f, 0x6d, 0x61, 0x78, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x72, 0x4d, 0x61,
	0x78, 0x12, 0x0c, 0x0a, 0x01, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x01, 0x73, 0x12,
	0x30, 0x0a, 0x14, 0x6f, 0x66, 0x66, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x70, 0x75, 0x62, 0x6c,
	0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x12, 0x6f,
	0x66, 0x66, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79,
	0x73, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x09, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x07, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64, 0x73, 0x12, 0x36, 0x0a, 0x17,
	0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x15, 0x72,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x43, 0x0a, 0x1e, 0x6d, 0x61, 0x78, 0x5f, 0x64, 0x75, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x6e, 0x61, 0x6e, 0x6f, 0x73,
	0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x04, 0x52, 0x1b, 0x6d, 0x61,
	0x78, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4e, 0x61,
	0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x4f, 0x0a, 0x24, 0x6d, 0x61, 0x78,
	0x5f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64,
	0x73, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x04, 0x52, 0x21, 0x6d, 0x61, 0x78, 0x44, 0x75, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4e,
	0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x45, 0x0a, 0x1f, 0x6d, 0x61,
	0x78, 0x5f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x5f, 0x6e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x0d, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x1c, 0x6d, 0x61, 0x78, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x4e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64,
	0x73, 0x12, 0x72, 0x0a, 0x37, 0x6d, 0x61, 0x78, 0x5f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x5f,
	0x66, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x5f, 0x6e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x0e, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x31, 0x6d, 0x61, 0x78, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53,
	0x68, 0x6f, 0x75, 0x6c, 0x64, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x46, 0x69, 0x6e, 0x61, 0x6c,
	0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x4e, 0x61, 0x6e, 0x6f, 0x73, 0x65,
	0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x74, 0x0a, 0x38, 0x6d, 0x61, 0x78, 0x5f, 0x64, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x5f, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x6d, 0x69, 0x74, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x5f, 0x72,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x6e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64,
	0x73, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x04, 0x52, 0x32, 0x6d, 0x61, 0x78, 0x44, 0x75, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x53, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6d,
	0x69, 0x74, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x4e, 0x61, 0x6e, 0x6f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x12, 0x67, 0x0a, 0x19, 0x73,
	0x68, 0x61, 0x72, 0x65, 0x64, 0x5f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x65, 0x6e, 0x63,
	0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x10, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b,
	0x2e, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x70, 0x62, 0x2e, 0x53,
	0x68, 0x61, 0x72, 0x65, 0x64, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x45, 0x6e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x17, 0x73, 0x68, 0x61,
	0x72, 0x65, 0x64, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x22, 0x9c, 0x01, 0x0a, 0x1c, 0x53, 0x68, 0x61, 0x72, 0x65, 0x64, 0x53,
	0x65, 0x63, 0x72, 0x65, 0x74, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x2e, 0x0a, 0x12, 0x64, 0x69, 0x66, 0x66, 0x69, 0x65, 0x48,
	0x65, 0x6c, 0x6c, 0x6d, 0x61, 0x6e, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x12, 0x64, 0x69, 0x66, 0x66, 0x69, 0x65, 0x48, 0x65, 0x6c, 0x6c, 0x6d, 0x61, 0x6e,
	0x50, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x2a, 0x0a, 0x10, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x53,
	0x65, 0x63, 0x72, 0x65, 0x74, 0x48, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x10, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x48, 0x61, 0x73,
	0x68, 0x12, 0x20, 0x0a, 0x0b, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0b, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_offchainreporting2_monitoring_offchain_config_proto_rawDescOnce sync.Once
	file_offchainreporting2_monitoring_offchain_config_proto_rawDescData = file_offchainreporting2_monitoring_offchain_config_proto_rawDesc
)

func file_offchainreporting2_monitoring_offchain_config_proto_rawDescGZIP() []byte {
	file_offchainreporting2_monitoring_offchain_config_proto_rawDescOnce.Do(func() {
		file_offchainreporting2_monitoring_offchain_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_offchainreporting2_monitoring_offchain_config_proto_rawDescData)
	})
	return file_offchainreporting2_monitoring_offchain_config_proto_rawDescData
}

var file_offchainreporting2_monitoring_offchain_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_offchainreporting2_monitoring_offchain_config_proto_goTypes = []interface{}{
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
	if !protoimpl.UnsafeEnabled {
		file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OffchainConfigProto); i {
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
		file_offchainreporting2_monitoring_offchain_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SharedSecretEncryptionsProto); i {
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
			RawDescriptor: file_offchainreporting2_monitoring_offchain_config_proto_rawDesc,
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
	file_offchainreporting2_monitoring_offchain_config_proto_rawDesc = nil
	file_offchainreporting2_monitoring_offchain_config_proto_goTypes = nil
	file_offchainreporting2_monitoring_offchain_config_proto_depIdxs = nil
}
