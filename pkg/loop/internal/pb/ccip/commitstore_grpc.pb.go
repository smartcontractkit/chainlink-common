// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.4
// source: commitstore.proto

package ccippb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	CommitStoreReader_ChangeConfig_FullMethodName                          = "/loop.internal.pb.ccip.CommitStoreReader/ChangeConfig"
	CommitStoreReader_DecodeCommitReport_FullMethodName                    = "/loop.internal.pb.ccip.CommitStoreReader/DecodeCommitReport"
	CommitStoreReader_EncodeCommitReport_FullMethodName                    = "/loop.internal.pb.ccip.CommitStoreReader/EncodeCommitReport"
	CommitStoreReader_GetAcceptedCommitReportsGteTimestamp_FullMethodName  = "/loop.internal.pb.ccip.CommitStoreReader/GetAcceptedCommitReportsGteTimestamp"
	CommitStoreReader_GetCommitGasPriceEstimator_FullMethodName            = "/loop.internal.pb.ccip.CommitStoreReader/GetCommitGasPriceEstimator"
	CommitStoreReader_GetCommitReportMatchingSequenceNumber_FullMethodName = "/loop.internal.pb.ccip.CommitStoreReader/GetCommitReportMatchingSequenceNumber"
	CommitStoreReader_GetCommitStoreStaticConfig_FullMethodName            = "/loop.internal.pb.ccip.CommitStoreReader/GetCommitStoreStaticConfig"
	CommitStoreReader_GetExpectedNextSequenceNumber_FullMethodName         = "/loop.internal.pb.ccip.CommitStoreReader/GetExpectedNextSequenceNumber"
	CommitStoreReader_GetLatestPriceEpochAndRound_FullMethodName           = "/loop.internal.pb.ccip.CommitStoreReader/GetLatestPriceEpochAndRound"
	CommitStoreReader_GetOffchainConfig_FullMethodName                     = "/loop.internal.pb.ccip.CommitStoreReader/GetOffchainConfig"
	CommitStoreReader_IsBlessed_FullMethodName                             = "/loop.internal.pb.ccip.CommitStoreReader/IsBlessed"
	CommitStoreReader_IsDestChainHealthy_FullMethodName                    = "/loop.internal.pb.ccip.CommitStoreReader/IsDestChainHealthy"
	CommitStoreReader_IsDown_FullMethodName                                = "/loop.internal.pb.ccip.CommitStoreReader/IsDown"
	CommitStoreReader_VerifyExecutionReport_FullMethodName                 = "/loop.internal.pb.ccip.CommitStoreReader/VerifyExecutionReport"
	CommitStoreReader_Close_FullMethodName                                 = "/loop.internal.pb.ccip.CommitStoreReader/Close"
)

// CommitStoreReaderClient is the client API for CommitStoreReader service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CommitStoreReaderClient interface {
	ChangeConfig(ctx context.Context, in *CommitStoreChangeConfigRequest, opts ...grpc.CallOption) (*CommitStoreChangeConfigResponse, error)
	DecodeCommitReport(ctx context.Context, in *DecodeCommitReportRequest, opts ...grpc.CallOption) (*DecodeCommitReportResponse, error)
	EncodeCommitReport(ctx context.Context, in *EncodeCommitReportRequest, opts ...grpc.CallOption) (*EncodeCommitReportResponse, error)
	GetAcceptedCommitReportsGteTimestamp(ctx context.Context, in *GetAcceptedCommitReportsGteTimestampRequest, opts ...grpc.CallOption) (*GetAcceptedCommitReportsGteTimestampResponse, error)
	GetCommitGasPriceEstimator(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetCommitGasPriceEstimatorResponse, error)
	GetCommitReportMatchingSequenceNumber(ctx context.Context, in *GetCommitReportMatchingSequenceNumberRequest, opts ...grpc.CallOption) (*GetCommitReportMatchingSequenceNumberResponse, error)
	GetCommitStoreStaticConfig(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetCommitStoreStaticConfigResponse, error)
	GetExpectedNextSequenceNumber(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetExpectedNextSequenceNumberResponse, error)
	GetLatestPriceEpochAndRound(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetLatestPriceEpochAndRoundResponse, error)
	GetOffchainConfig(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetOffchainConfigResponse, error)
	IsBlessed(ctx context.Context, in *IsBlessedRequest, opts ...grpc.CallOption) (*IsBlessedResponse, error)
	IsDestChainHealthy(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IsDestChainHealthyResponse, error)
	IsDown(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IsDownResponse, error)
	VerifyExecutionReport(ctx context.Context, in *VerifyExecutionReportRequest, opts ...grpc.CallOption) (*VerifyExecutionReportResponse, error)
	Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type commitStoreReaderClient struct {
	cc grpc.ClientConnInterface
}

func NewCommitStoreReaderClient(cc grpc.ClientConnInterface) CommitStoreReaderClient {
	return &commitStoreReaderClient{cc}
}

func (c *commitStoreReaderClient) ChangeConfig(ctx context.Context, in *CommitStoreChangeConfigRequest, opts ...grpc.CallOption) (*CommitStoreChangeConfigResponse, error) {
	out := new(CommitStoreChangeConfigResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_ChangeConfig_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) DecodeCommitReport(ctx context.Context, in *DecodeCommitReportRequest, opts ...grpc.CallOption) (*DecodeCommitReportResponse, error) {
	out := new(DecodeCommitReportResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_DecodeCommitReport_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) EncodeCommitReport(ctx context.Context, in *EncodeCommitReportRequest, opts ...grpc.CallOption) (*EncodeCommitReportResponse, error) {
	out := new(EncodeCommitReportResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_EncodeCommitReport_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) GetAcceptedCommitReportsGteTimestamp(ctx context.Context, in *GetAcceptedCommitReportsGteTimestampRequest, opts ...grpc.CallOption) (*GetAcceptedCommitReportsGteTimestampResponse, error) {
	out := new(GetAcceptedCommitReportsGteTimestampResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_GetAcceptedCommitReportsGteTimestamp_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) GetCommitGasPriceEstimator(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetCommitGasPriceEstimatorResponse, error) {
	out := new(GetCommitGasPriceEstimatorResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_GetCommitGasPriceEstimator_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) GetCommitReportMatchingSequenceNumber(ctx context.Context, in *GetCommitReportMatchingSequenceNumberRequest, opts ...grpc.CallOption) (*GetCommitReportMatchingSequenceNumberResponse, error) {
	out := new(GetCommitReportMatchingSequenceNumberResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_GetCommitReportMatchingSequenceNumber_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) GetCommitStoreStaticConfig(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetCommitStoreStaticConfigResponse, error) {
	out := new(GetCommitStoreStaticConfigResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_GetCommitStoreStaticConfig_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) GetExpectedNextSequenceNumber(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetExpectedNextSequenceNumberResponse, error) {
	out := new(GetExpectedNextSequenceNumberResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_GetExpectedNextSequenceNumber_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) GetLatestPriceEpochAndRound(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetLatestPriceEpochAndRoundResponse, error) {
	out := new(GetLatestPriceEpochAndRoundResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_GetLatestPriceEpochAndRound_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) GetOffchainConfig(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetOffchainConfigResponse, error) {
	out := new(GetOffchainConfigResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_GetOffchainConfig_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) IsBlessed(ctx context.Context, in *IsBlessedRequest, opts ...grpc.CallOption) (*IsBlessedResponse, error) {
	out := new(IsBlessedResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_IsBlessed_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) IsDestChainHealthy(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IsDestChainHealthyResponse, error) {
	out := new(IsDestChainHealthyResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_IsDestChainHealthy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) IsDown(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IsDownResponse, error) {
	out := new(IsDownResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_IsDown_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) VerifyExecutionReport(ctx context.Context, in *VerifyExecutionReportRequest, opts ...grpc.CallOption) (*VerifyExecutionReportResponse, error) {
	out := new(VerifyExecutionReportResponse)
	err := c.cc.Invoke(ctx, CommitStoreReader_VerifyExecutionReport_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commitStoreReaderClient) Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, CommitStoreReader_Close_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CommitStoreReaderServer is the server API for CommitStoreReader service.
// All implementations must embed UnimplementedCommitStoreReaderServer
// for forward compatibility
type CommitStoreReaderServer interface {
	ChangeConfig(context.Context, *CommitStoreChangeConfigRequest) (*CommitStoreChangeConfigResponse, error)
	DecodeCommitReport(context.Context, *DecodeCommitReportRequest) (*DecodeCommitReportResponse, error)
	EncodeCommitReport(context.Context, *EncodeCommitReportRequest) (*EncodeCommitReportResponse, error)
	GetAcceptedCommitReportsGteTimestamp(context.Context, *GetAcceptedCommitReportsGteTimestampRequest) (*GetAcceptedCommitReportsGteTimestampResponse, error)
	GetCommitGasPriceEstimator(context.Context, *emptypb.Empty) (*GetCommitGasPriceEstimatorResponse, error)
	GetCommitReportMatchingSequenceNumber(context.Context, *GetCommitReportMatchingSequenceNumberRequest) (*GetCommitReportMatchingSequenceNumberResponse, error)
	GetCommitStoreStaticConfig(context.Context, *emptypb.Empty) (*GetCommitStoreStaticConfigResponse, error)
	GetExpectedNextSequenceNumber(context.Context, *emptypb.Empty) (*GetExpectedNextSequenceNumberResponse, error)
	GetLatestPriceEpochAndRound(context.Context, *emptypb.Empty) (*GetLatestPriceEpochAndRoundResponse, error)
	GetOffchainConfig(context.Context, *emptypb.Empty) (*GetOffchainConfigResponse, error)
	IsBlessed(context.Context, *IsBlessedRequest) (*IsBlessedResponse, error)
	IsDestChainHealthy(context.Context, *emptypb.Empty) (*IsDestChainHealthyResponse, error)
	IsDown(context.Context, *emptypb.Empty) (*IsDownResponse, error)
	VerifyExecutionReport(context.Context, *VerifyExecutionReportRequest) (*VerifyExecutionReportResponse, error)
	Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedCommitStoreReaderServer()
}

// UnimplementedCommitStoreReaderServer must be embedded to have forward compatible implementations.
type UnimplementedCommitStoreReaderServer struct {
}

func (UnimplementedCommitStoreReaderServer) ChangeConfig(context.Context, *CommitStoreChangeConfigRequest) (*CommitStoreChangeConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeConfig not implemented")
}
func (UnimplementedCommitStoreReaderServer) DecodeCommitReport(context.Context, *DecodeCommitReportRequest) (*DecodeCommitReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DecodeCommitReport not implemented")
}
func (UnimplementedCommitStoreReaderServer) EncodeCommitReport(context.Context, *EncodeCommitReportRequest) (*EncodeCommitReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EncodeCommitReport not implemented")
}
func (UnimplementedCommitStoreReaderServer) GetAcceptedCommitReportsGteTimestamp(context.Context, *GetAcceptedCommitReportsGteTimestampRequest) (*GetAcceptedCommitReportsGteTimestampResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAcceptedCommitReportsGteTimestamp not implemented")
}
func (UnimplementedCommitStoreReaderServer) GetCommitGasPriceEstimator(context.Context, *emptypb.Empty) (*GetCommitGasPriceEstimatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCommitGasPriceEstimator not implemented")
}
func (UnimplementedCommitStoreReaderServer) GetCommitReportMatchingSequenceNumber(context.Context, *GetCommitReportMatchingSequenceNumberRequest) (*GetCommitReportMatchingSequenceNumberResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCommitReportMatchingSequenceNumber not implemented")
}
func (UnimplementedCommitStoreReaderServer) GetCommitStoreStaticConfig(context.Context, *emptypb.Empty) (*GetCommitStoreStaticConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCommitStoreStaticConfig not implemented")
}
func (UnimplementedCommitStoreReaderServer) GetExpectedNextSequenceNumber(context.Context, *emptypb.Empty) (*GetExpectedNextSequenceNumberResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetExpectedNextSequenceNumber not implemented")
}
func (UnimplementedCommitStoreReaderServer) GetLatestPriceEpochAndRound(context.Context, *emptypb.Empty) (*GetLatestPriceEpochAndRoundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLatestPriceEpochAndRound not implemented")
}
func (UnimplementedCommitStoreReaderServer) GetOffchainConfig(context.Context, *emptypb.Empty) (*GetOffchainConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOffchainConfig not implemented")
}
func (UnimplementedCommitStoreReaderServer) IsBlessed(context.Context, *IsBlessedRequest) (*IsBlessedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsBlessed not implemented")
}
func (UnimplementedCommitStoreReaderServer) IsDestChainHealthy(context.Context, *emptypb.Empty) (*IsDestChainHealthyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsDestChainHealthy not implemented")
}
func (UnimplementedCommitStoreReaderServer) IsDown(context.Context, *emptypb.Empty) (*IsDownResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsDown not implemented")
}
func (UnimplementedCommitStoreReaderServer) VerifyExecutionReport(context.Context, *VerifyExecutionReportRequest) (*VerifyExecutionReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyExecutionReport not implemented")
}
func (UnimplementedCommitStoreReaderServer) Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
func (UnimplementedCommitStoreReaderServer) mustEmbedUnimplementedCommitStoreReaderServer() {}

// UnsafeCommitStoreReaderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CommitStoreReaderServer will
// result in compilation errors.
type UnsafeCommitStoreReaderServer interface {
	mustEmbedUnimplementedCommitStoreReaderServer()
}

func RegisterCommitStoreReaderServer(s grpc.ServiceRegistrar, srv CommitStoreReaderServer) {
	s.RegisterService(&CommitStoreReader_ServiceDesc, srv)
}

func _CommitStoreReader_ChangeConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitStoreChangeConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).ChangeConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_ChangeConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).ChangeConfig(ctx, req.(*CommitStoreChangeConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_DecodeCommitReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DecodeCommitReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).DecodeCommitReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_DecodeCommitReport_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).DecodeCommitReport(ctx, req.(*DecodeCommitReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_EncodeCommitReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EncodeCommitReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).EncodeCommitReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_EncodeCommitReport_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).EncodeCommitReport(ctx, req.(*EncodeCommitReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_GetAcceptedCommitReportsGteTimestamp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAcceptedCommitReportsGteTimestampRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).GetAcceptedCommitReportsGteTimestamp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_GetAcceptedCommitReportsGteTimestamp_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).GetAcceptedCommitReportsGteTimestamp(ctx, req.(*GetAcceptedCommitReportsGteTimestampRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_GetCommitGasPriceEstimator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).GetCommitGasPriceEstimator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_GetCommitGasPriceEstimator_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).GetCommitGasPriceEstimator(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_GetCommitReportMatchingSequenceNumber_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCommitReportMatchingSequenceNumberRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).GetCommitReportMatchingSequenceNumber(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_GetCommitReportMatchingSequenceNumber_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).GetCommitReportMatchingSequenceNumber(ctx, req.(*GetCommitReportMatchingSequenceNumberRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_GetCommitStoreStaticConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).GetCommitStoreStaticConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_GetCommitStoreStaticConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).GetCommitStoreStaticConfig(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_GetExpectedNextSequenceNumber_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).GetExpectedNextSequenceNumber(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_GetExpectedNextSequenceNumber_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).GetExpectedNextSequenceNumber(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_GetLatestPriceEpochAndRound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).GetLatestPriceEpochAndRound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_GetLatestPriceEpochAndRound_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).GetLatestPriceEpochAndRound(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_GetOffchainConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).GetOffchainConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_GetOffchainConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).GetOffchainConfig(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_IsBlessed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsBlessedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).IsBlessed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_IsBlessed_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).IsBlessed(ctx, req.(*IsBlessedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_IsDestChainHealthy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).IsDestChainHealthy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_IsDestChainHealthy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).IsDestChainHealthy(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_IsDown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).IsDown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_IsDown_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).IsDown(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_VerifyExecutionReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyExecutionReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).VerifyExecutionReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_VerifyExecutionReport_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).VerifyExecutionReport(ctx, req.(*VerifyExecutionReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommitStoreReader_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommitStoreReaderServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommitStoreReader_Close_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommitStoreReaderServer).Close(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// CommitStoreReader_ServiceDesc is the grpc.ServiceDesc for CommitStoreReader service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CommitStoreReader_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.internal.pb.ccip.CommitStoreReader",
	HandlerType: (*CommitStoreReaderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ChangeConfig",
			Handler:    _CommitStoreReader_ChangeConfig_Handler,
		},
		{
			MethodName: "DecodeCommitReport",
			Handler:    _CommitStoreReader_DecodeCommitReport_Handler,
		},
		{
			MethodName: "EncodeCommitReport",
			Handler:    _CommitStoreReader_EncodeCommitReport_Handler,
		},
		{
			MethodName: "GetAcceptedCommitReportsGteTimestamp",
			Handler:    _CommitStoreReader_GetAcceptedCommitReportsGteTimestamp_Handler,
		},
		{
			MethodName: "GetCommitGasPriceEstimator",
			Handler:    _CommitStoreReader_GetCommitGasPriceEstimator_Handler,
		},
		{
			MethodName: "GetCommitReportMatchingSequenceNumber",
			Handler:    _CommitStoreReader_GetCommitReportMatchingSequenceNumber_Handler,
		},
		{
			MethodName: "GetCommitStoreStaticConfig",
			Handler:    _CommitStoreReader_GetCommitStoreStaticConfig_Handler,
		},
		{
			MethodName: "GetExpectedNextSequenceNumber",
			Handler:    _CommitStoreReader_GetExpectedNextSequenceNumber_Handler,
		},
		{
			MethodName: "GetLatestPriceEpochAndRound",
			Handler:    _CommitStoreReader_GetLatestPriceEpochAndRound_Handler,
		},
		{
			MethodName: "GetOffchainConfig",
			Handler:    _CommitStoreReader_GetOffchainConfig_Handler,
		},
		{
			MethodName: "IsBlessed",
			Handler:    _CommitStoreReader_IsBlessed_Handler,
		},
		{
			MethodName: "IsDestChainHealthy",
			Handler:    _CommitStoreReader_IsDestChainHealthy_Handler,
		},
		{
			MethodName: "IsDown",
			Handler:    _CommitStoreReader_IsDown_Handler,
		},
		{
			MethodName: "VerifyExecutionReport",
			Handler:    _CommitStoreReader_VerifyExecutionReport_Handler,
		},
		{
			MethodName: "Close",
			Handler:    _CommitStoreReader_Close_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "commitstore.proto",
}
