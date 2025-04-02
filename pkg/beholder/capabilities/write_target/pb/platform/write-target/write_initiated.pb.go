// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v4.25.4
// source: write_initiated.proto

package write_target

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

// WT initiated the processing of the write request
type WriteInitiated struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Node      string `protobuf:"bytes,1,opt,name=node,proto3" json:"node,omitempty"`
	Forwarder string `protobuf:"bytes,2,opt,name=forwarder,proto3" json:"forwarder,omitempty"`
	Receiver  string `protobuf:"bytes,3,opt,name=receiver,proto3" json:"receiver,omitempty"`
	// Report Info
	ReportId uint32 `protobuf:"varint,4,opt,name=report_id,json=reportId,proto3" json:"report_id,omitempty"`
	// [Execution Context]
	// TODO: replace with a proto reference once supported
	// Execution Context - Source
	MetaSourceId string `protobuf:"bytes,20,opt,name=meta_source_id,json=metaSourceId,proto3" json:"meta_source_id,omitempty"`
	// Execution Context - Chain
	MetaChainFamilyName string `protobuf:"bytes,21,opt,name=meta_chain_family_name,json=metaChainFamilyName,proto3" json:"meta_chain_family_name,omitempty"`
	MetaChainId         string `protobuf:"bytes,22,opt,name=meta_chain_id,json=metaChainId,proto3" json:"meta_chain_id,omitempty"`
	MetaNetworkName     string `protobuf:"bytes,23,opt,name=meta_network_name,json=metaNetworkName,proto3" json:"meta_network_name,omitempty"`
	MetaNetworkNameFull string `protobuf:"bytes,24,opt,name=meta_network_name_full,json=metaNetworkNameFull,proto3" json:"meta_network_name_full,omitempty"`
	// Execution Context - Workflow (capabilities.RequestMetadata)
	MetaWorkflowId               string `protobuf:"bytes,25,opt,name=meta_workflow_id,json=metaWorkflowId,proto3" json:"meta_workflow_id,omitempty"`
	MetaWorkflowOwner            string `protobuf:"bytes,26,opt,name=meta_workflow_owner,json=metaWorkflowOwner,proto3" json:"meta_workflow_owner,omitempty"`
	MetaWorkflowExecutionId      string `protobuf:"bytes,27,opt,name=meta_workflow_execution_id,json=metaWorkflowExecutionId,proto3" json:"meta_workflow_execution_id,omitempty"`
	MetaWorkflowName             string `protobuf:"bytes,28,opt,name=meta_workflow_name,json=metaWorkflowName,proto3" json:"meta_workflow_name,omitempty"`
	MetaWorkflowDonId            uint32 `protobuf:"varint,29,opt,name=meta_workflow_don_id,json=metaWorkflowDonId,proto3" json:"meta_workflow_don_id,omitempty"`
	MetaWorkflowDonConfigVersion uint32 `protobuf:"varint,30,opt,name=meta_workflow_don_config_version,json=metaWorkflowDonConfigVersion,proto3" json:"meta_workflow_don_config_version,omitempty"`
	MetaReferenceId              string `protobuf:"bytes,31,opt,name=meta_reference_id,json=metaReferenceId,proto3" json:"meta_reference_id,omitempty"`
	// Execution Context - Capability
	MetaCapabilityType           string `protobuf:"bytes,32,opt,name=meta_capability_type,json=metaCapabilityType,proto3" json:"meta_capability_type,omitempty"`
	MetaCapabilityId             string `protobuf:"bytes,33,opt,name=meta_capability_id,json=metaCapabilityId,proto3" json:"meta_capability_id,omitempty"`
	MetaCapabilityTimestampStart uint64 `protobuf:"varint,34,opt,name=meta_capability_timestamp_start,json=metaCapabilityTimestampStart,proto3" json:"meta_capability_timestamp_start,omitempty"`
	MetaCapabilityTimestampEmit  uint64 `protobuf:"varint,35,opt,name=meta_capability_timestamp_emit,json=metaCapabilityTimestampEmit,proto3" json:"meta_capability_timestamp_emit,omitempty"`
}

func (x *WriteInitiated) Reset() {
	*x = WriteInitiated{}
	if protoimpl.UnsafeEnabled {
		mi := &file_write_initiated_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteInitiated) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteInitiated) ProtoMessage() {}

func (x *WriteInitiated) ProtoReflect() protoreflect.Message {
	mi := &file_write_initiated_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteInitiated.ProtoReflect.Descriptor instead.
func (*WriteInitiated) Descriptor() ([]byte, []int) {
	return file_write_initiated_proto_rawDescGZIP(), []int{0}
}

func (x *WriteInitiated) GetNode() string {
	if x != nil {
		return x.Node
	}
	return ""
}

func (x *WriteInitiated) GetForwarder() string {
	if x != nil {
		return x.Forwarder
	}
	return ""
}

func (x *WriteInitiated) GetReceiver() string {
	if x != nil {
		return x.Receiver
	}
	return ""
}

func (x *WriteInitiated) GetReportId() uint32 {
	if x != nil {
		return x.ReportId
	}
	return 0
}

func (x *WriteInitiated) GetMetaSourceId() string {
	if x != nil {
		return x.MetaSourceId
	}
	return ""
}

func (x *WriteInitiated) GetMetaChainFamilyName() string {
	if x != nil {
		return x.MetaChainFamilyName
	}
	return ""
}

func (x *WriteInitiated) GetMetaChainId() string {
	if x != nil {
		return x.MetaChainId
	}
	return ""
}

func (x *WriteInitiated) GetMetaNetworkName() string {
	if x != nil {
		return x.MetaNetworkName
	}
	return ""
}

func (x *WriteInitiated) GetMetaNetworkNameFull() string {
	if x != nil {
		return x.MetaNetworkNameFull
	}
	return ""
}

func (x *WriteInitiated) GetMetaWorkflowId() string {
	if x != nil {
		return x.MetaWorkflowId
	}
	return ""
}

func (x *WriteInitiated) GetMetaWorkflowOwner() string {
	if x != nil {
		return x.MetaWorkflowOwner
	}
	return ""
}

func (x *WriteInitiated) GetMetaWorkflowExecutionId() string {
	if x != nil {
		return x.MetaWorkflowExecutionId
	}
	return ""
}

func (x *WriteInitiated) GetMetaWorkflowName() string {
	if x != nil {
		return x.MetaWorkflowName
	}
	return ""
}

func (x *WriteInitiated) GetMetaWorkflowDonId() uint32 {
	if x != nil {
		return x.MetaWorkflowDonId
	}
	return 0
}

func (x *WriteInitiated) GetMetaWorkflowDonConfigVersion() uint32 {
	if x != nil {
		return x.MetaWorkflowDonConfigVersion
	}
	return 0
}

func (x *WriteInitiated) GetMetaReferenceId() string {
	if x != nil {
		return x.MetaReferenceId
	}
	return ""
}

func (x *WriteInitiated) GetMetaCapabilityType() string {
	if x != nil {
		return x.MetaCapabilityType
	}
	return ""
}

func (x *WriteInitiated) GetMetaCapabilityId() string {
	if x != nil {
		return x.MetaCapabilityId
	}
	return ""
}

func (x *WriteInitiated) GetMetaCapabilityTimestampStart() uint64 {
	if x != nil {
		return x.MetaCapabilityTimestampStart
	}
	return 0
}

func (x *WriteInitiated) GetMetaCapabilityTimestampEmit() uint64 {
	if x != nil {
		return x.MetaCapabilityTimestampEmit
	}
	return 0
}

var File_write_initiated_proto protoreflect.FileDescriptor

var file_write_initiated_proto_rawDesc = []byte{
	0x0a, 0x15, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x74, 0x65,
	0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72,
	0x6d, 0x2e, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x22, 0xb1,
	0x07, 0x0a, 0x0e, 0x57, 0x72, 0x69, 0x74, 0x65, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x74, 0x65,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x6f, 0x64, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x66, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x66, 0x6f, 0x72, 0x77, 0x61, 0x72,
	0x64, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x12,
	0x1b, 0x0a, 0x09, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x08, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x49, 0x64, 0x12, 0x24, 0x0a, 0x0e,
	0x6d, 0x65, 0x74, 0x61, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x14,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6d, 0x65, 0x74, 0x61, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x49, 0x64, 0x12, 0x33, 0x0a, 0x16, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x5f, 0x66, 0x61, 0x6d, 0x69, 0x6c, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x15, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x13, 0x6d, 0x65, 0x74, 0x61, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x46, 0x61, 0x6d,
	0x69, 0x6c, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x22, 0x0a, 0x0d, 0x6d, 0x65, 0x74, 0x61, 0x5f,
	0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x16, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x6d, 0x65, 0x74, 0x61, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x12, 0x2a, 0x0a, 0x11, 0x6d,
	0x65, 0x74, 0x61, 0x5f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x17, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x6d, 0x65, 0x74, 0x61, 0x4e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x33, 0x0a, 0x16, 0x6d, 0x65, 0x74, 0x61, 0x5f,
	0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x5f, 0x66, 0x75, 0x6c,
	0x6c, 0x18, 0x18, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13, 0x6d, 0x65, 0x74, 0x61, 0x4e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x4e, 0x61, 0x6d, 0x65, 0x46, 0x75, 0x6c, 0x6c, 0x12, 0x28, 0x0a, 0x10,
	0x6d, 0x65, 0x74, 0x61, 0x5f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x69, 0x64,
	0x18, 0x19, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x6d, 0x65, 0x74, 0x61, 0x57, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x12, 0x2e, 0x0a, 0x13, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x77,
	0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x1a, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x11, 0x6d, 0x65, 0x74, 0x61, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f,
	0x77, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x3b, 0x0a, 0x1a, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x77,
	0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x69, 0x64, 0x18, 0x1b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x17, 0x6d, 0x65, 0x74, 0x61,
	0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f,
	0x6e, 0x49, 0x64, 0x12, 0x2c, 0x0a, 0x12, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x1c, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x10, 0x6d, 0x65, 0x74, 0x61, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x2f, 0x0a, 0x14, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x5f, 0x64, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x1d, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x11, 0x6d, 0x65, 0x74, 0x61, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x44, 0x6f, 0x6e,
	0x49, 0x64, 0x12, 0x46, 0x0a, 0x20, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x77, 0x6f, 0x72, 0x6b, 0x66,
	0x6c, 0x6f, 0x77, 0x5f, 0x64, 0x6f, 0x6e, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x1e, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x1c, 0x6d, 0x65,
	0x74, 0x61, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x44, 0x6f, 0x6e, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x2a, 0x0a, 0x11, 0x6d, 0x65,
	0x74, 0x61, 0x5f, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18,
	0x1f, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x6d, 0x65, 0x74, 0x61, 0x52, 0x65, 0x66, 0x65, 0x72,
	0x65, 0x6e, 0x63, 0x65, 0x49, 0x64, 0x12, 0x30, 0x0a, 0x14, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x63,
	0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x20,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x6d, 0x65, 0x74, 0x61, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69,
	0x6c, 0x69, 0x74, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x2c, 0x0a, 0x12, 0x6d, 0x65, 0x74, 0x61,
	0x5f, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x21,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x6d, 0x65, 0x74, 0x61, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69,
	0x6c, 0x69, 0x74, 0x79, 0x49, 0x64, 0x12, 0x45, 0x0a, 0x1f, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x63,
	0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x22, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x1c, 0x6d, 0x65, 0x74, 0x61, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x53, 0x74, 0x61, 0x72, 0x74, 0x12, 0x43, 0x0a,
	0x1e, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x5f, 0x65, 0x6d, 0x69, 0x74, 0x18,
	0x23, 0x20, 0x01, 0x28, 0x04, 0x52, 0x1b, 0x6d, 0x65, 0x74, 0x61, 0x43, 0x61, 0x70, 0x61, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x79, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x45, 0x6d,
	0x69, 0x74, 0x42, 0x10, 0x5a, 0x0e, 0x2e, 0x3b, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_write_initiated_proto_rawDescOnce sync.Once
	file_write_initiated_proto_rawDescData = file_write_initiated_proto_rawDesc
)

func file_write_initiated_proto_rawDescGZIP() []byte {
	file_write_initiated_proto_rawDescOnce.Do(func() {
		file_write_initiated_proto_rawDescData = protoimpl.X.CompressGZIP(file_write_initiated_proto_rawDescData)
	})
	return file_write_initiated_proto_rawDescData
}

var file_write_initiated_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_write_initiated_proto_goTypes = []any{
	(*WriteInitiated)(nil), // 0: platform.write_target.WriteInitiated
}
var file_write_initiated_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_write_initiated_proto_init() }
func file_write_initiated_proto_init() {
	if File_write_initiated_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_write_initiated_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*WriteInitiated); i {
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
			RawDescriptor: file_write_initiated_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_write_initiated_proto_goTypes,
		DependencyIndexes: file_write_initiated_proto_depIdxs,
		MessageInfos:      file_write_initiated_proto_msgTypes,
	}.Build()
	File_write_initiated_proto = out.File
	file_write_initiated_proto_rawDesc = nil
	file_write_initiated_proto_goTypes = nil
	file_write_initiated_proto_depIdxs = nil
}
