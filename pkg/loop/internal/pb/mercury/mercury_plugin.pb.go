// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: mercury_plugin.proto

// note: the generate.go file in this dir specifies the import path of the relative proto files

package mercurypb

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

type NewMercuryPluginRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MercuryPluginConfig *MercuryPluginConfig `protobuf:"bytes,1,opt,name=mercuryPluginConfig,proto3" json:"mercuryPluginConfig,omitempty"`
}

func (x *NewMercuryPluginRequest) Reset() {
	*x = NewMercuryPluginRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewMercuryPluginRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewMercuryPluginRequest) ProtoMessage() {}

func (x *NewMercuryPluginRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewMercuryPluginRequest.ProtoReflect.Descriptor instead.
func (*NewMercuryPluginRequest) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{0}
}

func (x *NewMercuryPluginRequest) GetMercuryPluginConfig() *MercuryPluginConfig {
	if x != nil {
		return x.MercuryPluginConfig
	}
	return nil
}

type NewMercuryPluginResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MercuryPluginID   uint32             `protobuf:"varint,1,opt,name=mercuryPluginID,proto3" json:"mercuryPluginID,omitempty"`
	MercuryPluginInfo *MercuryPluginInfo `protobuf:"bytes,2,opt,name=mercuryPluginInfo,proto3" json:"mercuryPluginInfo,omitempty"`
}

func (x *NewMercuryPluginResponse) Reset() {
	*x = NewMercuryPluginResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewMercuryPluginResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewMercuryPluginResponse) ProtoMessage() {}

func (x *NewMercuryPluginResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewMercuryPluginResponse.ProtoReflect.Descriptor instead.
func (*NewMercuryPluginResponse) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{1}
}

func (x *NewMercuryPluginResponse) GetMercuryPluginID() uint32 {
	if x != nil {
		return x.MercuryPluginID
	}
	return 0
}

func (x *NewMercuryPluginResponse) GetMercuryPluginInfo() *MercuryPluginInfo {
	if x != nil {
		return x.MercuryPluginInfo
	}
	return nil
}

// MercuryPluginConfig represents [github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types.MercuryPluginConfig]
type MercuryPluginConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ConfigDigest           []byte `protobuf:"bytes,1,opt,name=configDigest,proto3" json:"configDigest,omitempty"` // [32]byte
	OracleID               uint32 `protobuf:"varint,2,opt,name=oracleID,proto3" json:"oracleID,omitempty"`        // uint8
	N                      uint32 `protobuf:"varint,3,opt,name=n,proto3" json:"n,omitempty"`
	F                      uint32 `protobuf:"varint,4,opt,name=f,proto3" json:"f,omitempty"`
	OnchainConfig          []byte `protobuf:"bytes,5,opt,name=onchainConfig,proto3" json:"onchainConfig,omitempty"`
	OffchainConfig         []byte `protobuf:"bytes,6,opt,name=offchainConfig,proto3" json:"offchainConfig,omitempty"`
	EstimatedRoundInterval int64  `protobuf:"varint,7,opt,name=estimatedRoundInterval,proto3" json:"estimatedRoundInterval,omitempty"`
	MaxDurationObservation int64  `protobuf:"varint,8,opt,name=maxDurationObservation,proto3" json:"maxDurationObservation,omitempty"`
}

func (x *MercuryPluginConfig) Reset() {
	*x = MercuryPluginConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MercuryPluginConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MercuryPluginConfig) ProtoMessage() {}

func (x *MercuryPluginConfig) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MercuryPluginConfig.ProtoReflect.Descriptor instead.
func (*MercuryPluginConfig) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{2}
}

func (x *MercuryPluginConfig) GetConfigDigest() []byte {
	if x != nil {
		return x.ConfigDigest
	}
	return nil
}

func (x *MercuryPluginConfig) GetOracleID() uint32 {
	if x != nil {
		return x.OracleID
	}
	return 0
}

func (x *MercuryPluginConfig) GetN() uint32 {
	if x != nil {
		return x.N
	}
	return 0
}

func (x *MercuryPluginConfig) GetF() uint32 {
	if x != nil {
		return x.F
	}
	return 0
}

func (x *MercuryPluginConfig) GetOnchainConfig() []byte {
	if x != nil {
		return x.OnchainConfig
	}
	return nil
}

func (x *MercuryPluginConfig) GetOffchainConfig() []byte {
	if x != nil {
		return x.OffchainConfig
	}
	return nil
}

func (x *MercuryPluginConfig) GetEstimatedRoundInterval() int64 {
	if x != nil {
		return x.EstimatedRoundInterval
	}
	return 0
}

func (x *MercuryPluginConfig) GetMaxDurationObservation() int64 {
	if x != nil {
		return x.MaxDurationObservation
	}
	return 0
}

// MercuryPluginLimits represents [github.com/smartcontractkit/libocr/offchainreporting2plus/types.MercuryPluginLimits]
type MercuryPluginLimits struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MaxObservationLength uint64 `protobuf:"varint,1,opt,name=maxObservationLength,proto3" json:"maxObservationLength,omitempty"`
	MaxReportLength      uint64 `protobuf:"varint,2,opt,name=maxReportLength,proto3" json:"maxReportLength,omitempty"`
}

func (x *MercuryPluginLimits) Reset() {
	*x = MercuryPluginLimits{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MercuryPluginLimits) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MercuryPluginLimits) ProtoMessage() {}

func (x *MercuryPluginLimits) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MercuryPluginLimits.ProtoReflect.Descriptor instead.
func (*MercuryPluginLimits) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{3}
}

func (x *MercuryPluginLimits) GetMaxObservationLength() uint64 {
	if x != nil {
		return x.MaxObservationLength
	}
	return 0
}

func (x *MercuryPluginLimits) GetMaxReportLength() uint64 {
	if x != nil {
		return x.MaxReportLength
	}
	return 0
}

// MercuryPluginInfo represents [github.com/smartcontractkit/libocr/offchainreporting2plus/types.MercuryPluginInfo]
type MercuryPluginInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name                string               `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	MercuryPluginLimits *MercuryPluginLimits `protobuf:"bytes,3,opt,name=mercuryPluginLimits,proto3" json:"mercuryPluginLimits,omitempty"`
}

func (x *MercuryPluginInfo) Reset() {
	*x = MercuryPluginInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MercuryPluginInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MercuryPluginInfo) ProtoMessage() {}

func (x *MercuryPluginInfo) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MercuryPluginInfo.ProtoReflect.Descriptor instead.
func (*MercuryPluginInfo) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{4}
}

func (x *MercuryPluginInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *MercuryPluginInfo) GetMercuryPluginLimits() *MercuryPluginLimits {
	if x != nil {
		return x.MercuryPluginLimits
	}
	return nil
}

// ObservationRequest has arguments for [github.com/smartcontractkit/libocr/offchainreporting2plus/types.MercuryPlugin.Observation].
type ObservationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReportTimestamp *pb.ReportTimestamp `protobuf:"bytes,1,opt,name=reportTimestamp,proto3" json:"reportTimestamp,omitempty"`
	PreviousReport  []byte              `protobuf:"bytes,2,opt,name=previousReport,proto3" json:"previousReport,omitempty"`
}

func (x *ObservationRequest) Reset() {
	*x = ObservationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObservationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObservationRequest) ProtoMessage() {}

func (x *ObservationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObservationRequest.ProtoReflect.Descriptor instead.
func (*ObservationRequest) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{5}
}

func (x *ObservationRequest) GetReportTimestamp() *pb.ReportTimestamp {
	if x != nil {
		return x.ReportTimestamp
	}
	return nil
}

func (x *ObservationRequest) GetPreviousReport() []byte {
	if x != nil {
		return x.PreviousReport
	}
	return nil
}

// ObservationResponse has return arguments for [github.com/smartcontractkit/libocr/offchainreporting2plus/types.MercuryPlugin.Observation].
type ObservationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Observation []byte `protobuf:"bytes,1,opt,name=observation,proto3" json:"observation,omitempty"`
}

func (x *ObservationResponse) Reset() {
	*x = ObservationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObservationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObservationResponse) ProtoMessage() {}

func (x *ObservationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObservationResponse.ProtoReflect.Descriptor instead.
func (*ObservationResponse) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{6}
}

func (x *ObservationResponse) GetObservation() []byte {
	if x != nil {
		return x.Observation
	}
	return nil
}

// TODO some definitions are shared with reporting plugin for ocr2. not sure to copy or share. copy for now...
// AttributedObservation represents [github.com/smartcontractkit/libocr/offchainreporting2plus/types.AttributedObservation]
type AttributedObservation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Observation []byte `protobuf:"bytes,1,opt,name=observation,proto3" json:"observation,omitempty"`
	Observer    uint32 `protobuf:"varint,2,opt,name=observer,proto3" json:"observer,omitempty"` // uint8
}

func (x *AttributedObservation) Reset() {
	*x = AttributedObservation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AttributedObservation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AttributedObservation) ProtoMessage() {}

func (x *AttributedObservation) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AttributedObservation.ProtoReflect.Descriptor instead.
func (*AttributedObservation) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{7}
}

func (x *AttributedObservation) GetObservation() []byte {
	if x != nil {
		return x.Observation
	}
	return nil
}

func (x *AttributedObservation) GetObserver() uint32 {
	if x != nil {
		return x.Observer
	}
	return 0
}

// ReportRequest has arguments for [github.com/smartcontractkit/libocr/offchainreporting2plus/types.MercuryPlugin.Report].
type ReportRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReportTimestamp *pb.ReportTimestamp      `protobuf:"bytes,1,opt,name=reportTimestamp,proto3" json:"reportTimestamp,omitempty"`
	PreviousReport  []byte                   `protobuf:"bytes,2,opt,name=previousReport,proto3" json:"previousReport,omitempty"`
	Observations    []*AttributedObservation `protobuf:"bytes,3,rep,name=observations,proto3" json:"observations,omitempty"`
}

func (x *ReportRequest) Reset() {
	*x = ReportRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReportRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReportRequest) ProtoMessage() {}

func (x *ReportRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReportRequest.ProtoReflect.Descriptor instead.
func (*ReportRequest) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{8}
}

func (x *ReportRequest) GetReportTimestamp() *pb.ReportTimestamp {
	if x != nil {
		return x.ReportTimestamp
	}
	return nil
}

func (x *ReportRequest) GetPreviousReport() []byte {
	if x != nil {
		return x.PreviousReport
	}
	return nil
}

func (x *ReportRequest) GetObservations() []*AttributedObservation {
	if x != nil {
		return x.Observations
	}
	return nil
}

// ReportResponse has return arguments for [github.com/smartcontractkit/libocr/offchainreporting2plus/types.MercuryPlugin.Report].
type ReportResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShouldReport bool   `protobuf:"varint,1,opt,name=shouldReport,proto3" json:"shouldReport,omitempty"`
	Report       []byte `protobuf:"bytes,2,opt,name=report,proto3" json:"report,omitempty"`
}

func (x *ReportResponse) Reset() {
	*x = ReportResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mercury_plugin_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReportResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReportResponse) ProtoMessage() {}

func (x *ReportResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mercury_plugin_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReportResponse.ProtoReflect.Descriptor instead.
func (*ReportResponse) Descriptor() ([]byte, []int) {
	return file_mercury_plugin_proto_rawDescGZIP(), []int{9}
}

func (x *ReportResponse) GetShouldReport() bool {
	if x != nil {
		return x.ShouldReport
	}
	return false
}

func (x *ReportResponse) GetReport() []byte {
	if x != nil {
		return x.Report
	}
	return nil
}

var File_mercury_plugin_proto protoreflect.FileDescriptor

var file_mercury_plugin_proto_rawDesc = []byte{
	0x0a, 0x14, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x5f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79,
	0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x72,
	0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7a, 0x0a, 0x17,
	0x4e, 0x65, 0x77, 0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x5f, 0x0a, 0x13, 0x6d, 0x65, 0x72, 0x63, 0x75,
	0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x2e,
	0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x13, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x9f, 0x01, 0x0a, 0x18, 0x4e, 0x65, 0x77,
	0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x0f, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0f,
	0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x49, 0x44, 0x12,
	0x59, 0x0a, 0x11, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x49, 0x6e, 0x66, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x6c, 0x6f, 0x6f,
	0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65,
	0x72, 0x63, 0x75, 0x72, 0x79, 0x2e, 0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x11, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0xaf, 0x02, 0x0a, 0x13, 0x4d,
	0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0x22, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x44, 0x69, 0x67, 0x65,
	0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x44, 0x69, 0x67, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6f, 0x72, 0x61, 0x63, 0x6c, 0x65,
	0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x6f, 0x72, 0x61, 0x63, 0x6c, 0x65,
	0x49, 0x44, 0x12, 0x0c, 0x0a, 0x01, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x01, 0x6e,
	0x12, 0x0c, 0x0a, 0x01, 0x66, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x01, 0x66, 0x12, 0x24,
	0x0a, 0x0d, 0x6f, 0x6e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x6f, 0x6e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x26, 0x0a, 0x0e, 0x6f, 0x66, 0x66, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x6f, 0x66,
	0x66, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x36, 0x0a, 0x16,
	0x65, 0x73, 0x74, 0x69, 0x6d, 0x61, 0x74, 0x65, 0x64, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x49, 0x6e,
	0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x16, 0x65, 0x73,
	0x74, 0x69, 0x6d, 0x61, 0x74, 0x65, 0x64, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x49, 0x6e, 0x74, 0x65,
	0x72, 0x76, 0x61, 0x6c, 0x12, 0x36, 0x0a, 0x16, 0x6d, 0x61, 0x78, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x16, 0x6d, 0x61, 0x78, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x73, 0x0a, 0x13,
	0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4c, 0x69, 0x6d,
	0x69, 0x74, 0x73, 0x12, 0x32, 0x0a, 0x14, 0x6d, 0x61, 0x78, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x14, 0x6d, 0x61, 0x78, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x28, 0x0a, 0x0f, 0x6d, 0x61, 0x78, 0x52, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x0f, 0x6d, 0x61, 0x78, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x4c, 0x65, 0x6e, 0x67, 0x74,
	0x68, 0x22, 0x88, 0x01, 0x0a, 0x11, 0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x5f, 0x0a, 0x13, 0x6d,
	0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4c, 0x69, 0x6d, 0x69,
	0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63,
	0x75, 0x72, 0x79, 0x2e, 0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x73, 0x52, 0x13, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x73, 0x22, 0x7d, 0x0a, 0x12,
	0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x3f, 0x0a, 0x0f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6c, 0x6f,
	0x6f, 0x70, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x0f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x12, 0x26, 0x0a, 0x0e, 0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x52,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x70, 0x72, 0x65,
	0x76, 0x69, 0x6f, 0x75, 0x73, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x37, 0x0a, 0x13, 0x4f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x55, 0x0a, 0x15, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x64, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x20, 0x0a,
	0x0b, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x0b, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x1a, 0x0a, 0x08, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x08, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x22, 0xcd, 0x01, 0x0a, 0x0d,
	0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3f, 0x0a,
	0x0f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x52, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0f, 0x72,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x26,
	0x0a, 0x0e, 0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73,
	0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x53, 0x0a, 0x0c, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x6c,
	0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e,
	0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x2e, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x64, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x6f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x4c, 0x0a, 0x0e, 0x52,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x22, 0x0a,
	0x0c, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x0c, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x52, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x06, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x32, 0x93, 0x01, 0x0a, 0x14, 0x4d, 0x65,
	0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x46, 0x61, 0x63, 0x74, 0x6f,
	0x72, 0x79, 0x12, 0x7b, 0x0a, 0x10, 0x4e, 0x65, 0x77, 0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x31, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72,
	0x79, 0x2e, 0x4e, 0x65, 0x77, 0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x32, 0x2e, 0x6c, 0x6f, 0x6f, 0x70,
	0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72,
	0x63, 0x75, 0x72, 0x79, 0x2e, 0x4e, 0x65, 0x77, 0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x32,
	0x97, 0x02, 0x0a, 0x0d, 0x4d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x12, 0x6c, 0x0a, 0x0b, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x2c, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x2e, 0x4f, 0x62, 0x73, 0x65,
	0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d,
	0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70,
	0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x2e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x5d, 0x0a, 0x06, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x27, 0x2e, 0x6c, 0x6f, 0x6f, 0x70,
	0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72,
	0x63, 0x75, 0x72, 0x79, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x28, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2e, 0x70, 0x62, 0x2e, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x2e, 0x52, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x39,
	0x0a, 0x05, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a,
	0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x42, 0x55, 0x5a, 0x53, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x63, 0x6f, 0x6e,
	0x74, 0x72, 0x61, 0x63, 0x74, 0x6b, 0x69, 0x74, 0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x6c, 0x69,
	0x6e, 0x6b, 0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6c, 0x6f,
	0x6f, 0x70, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x62, 0x2f, 0x6d,
	0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x3b, 0x6d, 0x65, 0x72, 0x63, 0x75, 0x72, 0x79, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mercury_plugin_proto_rawDescOnce sync.Once
	file_mercury_plugin_proto_rawDescData = file_mercury_plugin_proto_rawDesc
)

func file_mercury_plugin_proto_rawDescGZIP() []byte {
	file_mercury_plugin_proto_rawDescOnce.Do(func() {
		file_mercury_plugin_proto_rawDescData = protoimpl.X.CompressGZIP(file_mercury_plugin_proto_rawDescData)
	})
	return file_mercury_plugin_proto_rawDescData
}

var file_mercury_plugin_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_mercury_plugin_proto_goTypes = []interface{}{
	(*NewMercuryPluginRequest)(nil),  // 0: loop.internal.pb.mercury.NewMercuryPluginRequest
	(*NewMercuryPluginResponse)(nil), // 1: loop.internal.pb.mercury.NewMercuryPluginResponse
	(*MercuryPluginConfig)(nil),      // 2: loop.internal.pb.mercury.MercuryPluginConfig
	(*MercuryPluginLimits)(nil),      // 3: loop.internal.pb.mercury.MercuryPluginLimits
	(*MercuryPluginInfo)(nil),        // 4: loop.internal.pb.mercury.MercuryPluginInfo
	(*ObservationRequest)(nil),       // 5: loop.internal.pb.mercury.ObservationRequest
	(*ObservationResponse)(nil),      // 6: loop.internal.pb.mercury.ObservationResponse
	(*AttributedObservation)(nil),    // 7: loop.internal.pb.mercury.AttributedObservation
	(*ReportRequest)(nil),            // 8: loop.internal.pb.mercury.ReportRequest
	(*ReportResponse)(nil),           // 9: loop.internal.pb.mercury.ReportResponse
	(*pb.ReportTimestamp)(nil),       // 10: loop.ReportTimestamp
	(*emptypb.Empty)(nil),            // 11: google.protobuf.Empty
}
var file_mercury_plugin_proto_depIdxs = []int32{
	2,  // 0: loop.internal.pb.mercury.NewMercuryPluginRequest.mercuryPluginConfig:type_name -> loop.internal.pb.mercury.MercuryPluginConfig
	4,  // 1: loop.internal.pb.mercury.NewMercuryPluginResponse.mercuryPluginInfo:type_name -> loop.internal.pb.mercury.MercuryPluginInfo
	3,  // 2: loop.internal.pb.mercury.MercuryPluginInfo.mercuryPluginLimits:type_name -> loop.internal.pb.mercury.MercuryPluginLimits
	10, // 3: loop.internal.pb.mercury.ObservationRequest.reportTimestamp:type_name -> loop.ReportTimestamp
	10, // 4: loop.internal.pb.mercury.ReportRequest.reportTimestamp:type_name -> loop.ReportTimestamp
	7,  // 5: loop.internal.pb.mercury.ReportRequest.observations:type_name -> loop.internal.pb.mercury.AttributedObservation
	0,  // 6: loop.internal.pb.mercury.MercuryPluginFactory.NewMercuryPlugin:input_type -> loop.internal.pb.mercury.NewMercuryPluginRequest
	5,  // 7: loop.internal.pb.mercury.MercuryPlugin.Observation:input_type -> loop.internal.pb.mercury.ObservationRequest
	8,  // 8: loop.internal.pb.mercury.MercuryPlugin.Report:input_type -> loop.internal.pb.mercury.ReportRequest
	11, // 9: loop.internal.pb.mercury.MercuryPlugin.Close:input_type -> google.protobuf.Empty
	1,  // 10: loop.internal.pb.mercury.MercuryPluginFactory.NewMercuryPlugin:output_type -> loop.internal.pb.mercury.NewMercuryPluginResponse
	6,  // 11: loop.internal.pb.mercury.MercuryPlugin.Observation:output_type -> loop.internal.pb.mercury.ObservationResponse
	9,  // 12: loop.internal.pb.mercury.MercuryPlugin.Report:output_type -> loop.internal.pb.mercury.ReportResponse
	11, // 13: loop.internal.pb.mercury.MercuryPlugin.Close:output_type -> google.protobuf.Empty
	10, // [10:14] is the sub-list for method output_type
	6,  // [6:10] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_mercury_plugin_proto_init() }
func file_mercury_plugin_proto_init() {
	if File_mercury_plugin_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mercury_plugin_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewMercuryPluginRequest); i {
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
		file_mercury_plugin_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewMercuryPluginResponse); i {
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
		file_mercury_plugin_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MercuryPluginConfig); i {
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
		file_mercury_plugin_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MercuryPluginLimits); i {
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
		file_mercury_plugin_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MercuryPluginInfo); i {
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
		file_mercury_plugin_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObservationRequest); i {
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
		file_mercury_plugin_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObservationResponse); i {
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
		file_mercury_plugin_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AttributedObservation); i {
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
		file_mercury_plugin_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReportRequest); i {
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
		file_mercury_plugin_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReportResponse); i {
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
			RawDescriptor: file_mercury_plugin_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_mercury_plugin_proto_goTypes,
		DependencyIndexes: file_mercury_plugin_proto_depIdxs,
		MessageInfos:      file_mercury_plugin_proto_msgTypes,
	}.Build()
	File_mercury_plugin_proto = out.File
	file_mercury_plugin_proto_rawDesc = nil
	file_mercury_plugin_proto_goTypes = nil
	file_mercury_plugin_proto_depIdxs = nil
}
