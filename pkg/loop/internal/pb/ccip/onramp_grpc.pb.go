// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.4
// source: onramp.proto

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
	OnRampReader_Address_FullMethodName                       = "/loop.internal.pb.ccip.OnRampReader/Address"
	OnRampReader_GetDynamicConfig_FullMethodName              = "/loop.internal.pb.ccip.OnRampReader/GetDynamicConfig"
	OnRampReader_GetSendRequestsBetweenSeqNums_FullMethodName = "/loop.internal.pb.ccip.OnRampReader/GetSendRequestsBetweenSeqNums"
	OnRampReader_IsSourceChainHealthy_FullMethodName          = "/loop.internal.pb.ccip.OnRampReader/IsSourceChainHealthy"
	OnRampReader_IsSourceCursed_FullMethodName                = "/loop.internal.pb.ccip.OnRampReader/IsSourceCursed"
	OnRampReader_RouterAddress_FullMethodName                 = "/loop.internal.pb.ccip.OnRampReader/RouterAddress"
	OnRampReader_SourcePriceRegistryAddress_FullMethodName    = "/loop.internal.pb.ccip.OnRampReader/SourcePriceRegistryAddress"
	OnRampReader_Close_FullMethodName                         = "/loop.internal.pb.ccip.OnRampReader/Close"
)

// OnRampReaderClient is the client API for OnRampReader service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OnRampReaderClient interface {
	Address(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*OnrampAddressResponse, error)
	GetDynamicConfig(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetDynamicConfigResponse, error)
	GetSendRequestsBetweenSeqNums(ctx context.Context, in *GetSendRequestsBetweenSeqNumsRequest, opts ...grpc.CallOption) (*GetSendRequestsBetweenSeqNumsResponse, error)
	IsSourceChainHealthy(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IsSourceChainHealthyResponse, error)
	IsSourceCursed(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IsSourceCursedResponse, error)
	RouterAddress(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*RouterAddressResponse, error)
	SourcePriceRegistryAddress(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*SourcePriceRegistryAddressResponse, error)
	Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type onRampReaderClient struct {
	cc grpc.ClientConnInterface
}

func NewOnRampReaderClient(cc grpc.ClientConnInterface) OnRampReaderClient {
	return &onRampReaderClient{cc}
}

func (c *onRampReaderClient) Address(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*OnrampAddressResponse, error) {
	out := new(OnrampAddressResponse)
	err := c.cc.Invoke(ctx, OnRampReader_Address_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onRampReaderClient) GetDynamicConfig(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetDynamicConfigResponse, error) {
	out := new(GetDynamicConfigResponse)
	err := c.cc.Invoke(ctx, OnRampReader_GetDynamicConfig_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onRampReaderClient) GetSendRequestsBetweenSeqNums(ctx context.Context, in *GetSendRequestsBetweenSeqNumsRequest, opts ...grpc.CallOption) (*GetSendRequestsBetweenSeqNumsResponse, error) {
	out := new(GetSendRequestsBetweenSeqNumsResponse)
	err := c.cc.Invoke(ctx, OnRampReader_GetSendRequestsBetweenSeqNums_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onRampReaderClient) IsSourceChainHealthy(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IsSourceChainHealthyResponse, error) {
	out := new(IsSourceChainHealthyResponse)
	err := c.cc.Invoke(ctx, OnRampReader_IsSourceChainHealthy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onRampReaderClient) IsSourceCursed(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*IsSourceCursedResponse, error) {
	out := new(IsSourceCursedResponse)
	err := c.cc.Invoke(ctx, OnRampReader_IsSourceCursed_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onRampReaderClient) RouterAddress(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*RouterAddressResponse, error) {
	out := new(RouterAddressResponse)
	err := c.cc.Invoke(ctx, OnRampReader_RouterAddress_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onRampReaderClient) SourcePriceRegistryAddress(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*SourcePriceRegistryAddressResponse, error) {
	out := new(SourcePriceRegistryAddressResponse)
	err := c.cc.Invoke(ctx, OnRampReader_SourcePriceRegistryAddress_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *onRampReaderClient) Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, OnRampReader_Close_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OnRampReaderServer is the server API for OnRampReader service.
// All implementations must embed UnimplementedOnRampReaderServer
// for forward compatibility
type OnRampReaderServer interface {
	Address(context.Context, *emptypb.Empty) (*OnrampAddressResponse, error)
	GetDynamicConfig(context.Context, *emptypb.Empty) (*GetDynamicConfigResponse, error)
	GetSendRequestsBetweenSeqNums(context.Context, *GetSendRequestsBetweenSeqNumsRequest) (*GetSendRequestsBetweenSeqNumsResponse, error)
	IsSourceChainHealthy(context.Context, *emptypb.Empty) (*IsSourceChainHealthyResponse, error)
	IsSourceCursed(context.Context, *emptypb.Empty) (*IsSourceCursedResponse, error)
	RouterAddress(context.Context, *emptypb.Empty) (*RouterAddressResponse, error)
	SourcePriceRegistryAddress(context.Context, *emptypb.Empty) (*SourcePriceRegistryAddressResponse, error)
	Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedOnRampReaderServer()
}

// UnimplementedOnRampReaderServer must be embedded to have forward compatible implementations.
type UnimplementedOnRampReaderServer struct {
}

func (UnimplementedOnRampReaderServer) Address(context.Context, *emptypb.Empty) (*OnrampAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Address not implemented")
}
func (UnimplementedOnRampReaderServer) GetDynamicConfig(context.Context, *emptypb.Empty) (*GetDynamicConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDynamicConfig not implemented")
}
func (UnimplementedOnRampReaderServer) GetSendRequestsBetweenSeqNums(context.Context, *GetSendRequestsBetweenSeqNumsRequest) (*GetSendRequestsBetweenSeqNumsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSendRequestsBetweenSeqNums not implemented")
}
func (UnimplementedOnRampReaderServer) IsSourceChainHealthy(context.Context, *emptypb.Empty) (*IsSourceChainHealthyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsSourceChainHealthy not implemented")
}
func (UnimplementedOnRampReaderServer) IsSourceCursed(context.Context, *emptypb.Empty) (*IsSourceCursedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsSourceCursed not implemented")
}
func (UnimplementedOnRampReaderServer) RouterAddress(context.Context, *emptypb.Empty) (*RouterAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RouterAddress not implemented")
}
func (UnimplementedOnRampReaderServer) SourcePriceRegistryAddress(context.Context, *emptypb.Empty) (*SourcePriceRegistryAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SourcePriceRegistryAddress not implemented")
}
func (UnimplementedOnRampReaderServer) Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
func (UnimplementedOnRampReaderServer) mustEmbedUnimplementedOnRampReaderServer() {}

// UnsafeOnRampReaderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OnRampReaderServer will
// result in compilation errors.
type UnsafeOnRampReaderServer interface {
	mustEmbedUnimplementedOnRampReaderServer()
}

func RegisterOnRampReaderServer(s grpc.ServiceRegistrar, srv OnRampReaderServer) {
	s.RegisterService(&OnRampReader_ServiceDesc, srv)
}

func _OnRampReader_Address_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnRampReaderServer).Address(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OnRampReader_Address_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnRampReaderServer).Address(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnRampReader_GetDynamicConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnRampReaderServer).GetDynamicConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OnRampReader_GetDynamicConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnRampReaderServer).GetDynamicConfig(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnRampReader_GetSendRequestsBetweenSeqNums_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSendRequestsBetweenSeqNumsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnRampReaderServer).GetSendRequestsBetweenSeqNums(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OnRampReader_GetSendRequestsBetweenSeqNums_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnRampReaderServer).GetSendRequestsBetweenSeqNums(ctx, req.(*GetSendRequestsBetweenSeqNumsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnRampReader_IsSourceChainHealthy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnRampReaderServer).IsSourceChainHealthy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OnRampReader_IsSourceChainHealthy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnRampReaderServer).IsSourceChainHealthy(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnRampReader_IsSourceCursed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnRampReaderServer).IsSourceCursed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OnRampReader_IsSourceCursed_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnRampReaderServer).IsSourceCursed(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnRampReader_RouterAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnRampReaderServer).RouterAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OnRampReader_RouterAddress_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnRampReaderServer).RouterAddress(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnRampReader_SourcePriceRegistryAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnRampReaderServer).SourcePriceRegistryAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OnRampReader_SourcePriceRegistryAddress_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnRampReaderServer).SourcePriceRegistryAddress(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _OnRampReader_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnRampReaderServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OnRampReader_Close_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnRampReaderServer).Close(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// OnRampReader_ServiceDesc is the grpc.ServiceDesc for OnRampReader service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OnRampReader_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.internal.pb.ccip.OnRampReader",
	HandlerType: (*OnRampReaderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Address",
			Handler:    _OnRampReader_Address_Handler,
		},
		{
			MethodName: "GetDynamicConfig",
			Handler:    _OnRampReader_GetDynamicConfig_Handler,
		},
		{
			MethodName: "GetSendRequestsBetweenSeqNums",
			Handler:    _OnRampReader_GetSendRequestsBetweenSeqNums_Handler,
		},
		{
			MethodName: "IsSourceChainHealthy",
			Handler:    _OnRampReader_IsSourceChainHealthy_Handler,
		},
		{
			MethodName: "IsSourceCursed",
			Handler:    _OnRampReader_IsSourceCursed_Handler,
		},
		{
			MethodName: "RouterAddress",
			Handler:    _OnRampReader_RouterAddress_Handler,
		},
		{
			MethodName: "SourcePriceRegistryAddress",
			Handler:    _OnRampReader_SourcePriceRegistryAddress_Handler,
		},
		{
			MethodName: "Close",
			Handler:    _OnRampReader_Close_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "onramp.proto",
}
