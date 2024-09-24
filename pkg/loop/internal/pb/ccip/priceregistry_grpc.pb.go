// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.2
// source: priceregistry.proto

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
	PriceRegistryReader_GetTokenPriceUpdatesCreatedAfter_FullMethodName  = "/loop.internal.pb.ccip.PriceRegistryReader/GetTokenPriceUpdatesCreatedAfter"
	PriceRegistryReader_GetGasPriceUpdatesCreatedAfter_FullMethodName    = "/loop.internal.pb.ccip.PriceRegistryReader/GetGasPriceUpdatesCreatedAfter"
	PriceRegistryReader_GetAllGasPriceUpdatesCreatedAfter_FullMethodName = "/loop.internal.pb.ccip.PriceRegistryReader/GetAllGasPriceUpdatesCreatedAfter"
	PriceRegistryReader_GetAddress_FullMethodName                        = "/loop.internal.pb.ccip.PriceRegistryReader/GetAddress"
	PriceRegistryReader_GetFeeTokens_FullMethodName                      = "/loop.internal.pb.ccip.PriceRegistryReader/GetFeeTokens"
	PriceRegistryReader_GetTokenPrices_FullMethodName                    = "/loop.internal.pb.ccip.PriceRegistryReader/GetTokenPrices"
	PriceRegistryReader_GetTokensDecimals_FullMethodName                 = "/loop.internal.pb.ccip.PriceRegistryReader/GetTokensDecimals"
	PriceRegistryReader_Close_FullMethodName                             = "/loop.internal.pb.ccip.PriceRegistryReader/Close"
)

// PriceRegistryReaderClient is the client API for PriceRegistryReader service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PriceRegistryReaderClient interface {
	GetTokenPriceUpdatesCreatedAfter(ctx context.Context, in *GetTokenPriceUpdatesCreatedAfterRequest, opts ...grpc.CallOption) (*GetTokenPriceUpdatesCreatedAfterResponse, error)
	GetGasPriceUpdatesCreatedAfter(ctx context.Context, in *GetGasPriceUpdatesCreatedAfterRequest, opts ...grpc.CallOption) (*GetGasPriceUpdatesCreatedAfterResponse, error)
	GetAllGasPriceUpdatesCreatedAfter(ctx context.Context, in *GetAllGasPriceUpdatesCreatedAfterRequest, opts ...grpc.CallOption) (*GetAllGasPriceUpdatesCreatedAfterResponse, error)
	GetAddress(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetPriceRegistryAddressResponse, error)
	GetFeeTokens(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetFeeTokensResponse, error)
	GetTokenPrices(ctx context.Context, in *GetTokenPricesRequest, opts ...grpc.CallOption) (*GetTokenPricesResponse, error)
	GetTokensDecimals(ctx context.Context, in *GetTokensDecimalsRequest, opts ...grpc.CallOption) (*GetTokensDecimalsResponse, error)
	Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type priceRegistryReaderClient struct {
	cc grpc.ClientConnInterface
}

func NewPriceRegistryReaderClient(cc grpc.ClientConnInterface) PriceRegistryReaderClient {
	return &priceRegistryReaderClient{cc}
}

func (c *priceRegistryReaderClient) GetTokenPriceUpdatesCreatedAfter(ctx context.Context, in *GetTokenPriceUpdatesCreatedAfterRequest, opts ...grpc.CallOption) (*GetTokenPriceUpdatesCreatedAfterResponse, error) {
	out := new(GetTokenPriceUpdatesCreatedAfterResponse)
	err := c.cc.Invoke(ctx, PriceRegistryReader_GetTokenPriceUpdatesCreatedAfter_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceRegistryReaderClient) GetGasPriceUpdatesCreatedAfter(ctx context.Context, in *GetGasPriceUpdatesCreatedAfterRequest, opts ...grpc.CallOption) (*GetGasPriceUpdatesCreatedAfterResponse, error) {
	out := new(GetGasPriceUpdatesCreatedAfterResponse)
	err := c.cc.Invoke(ctx, PriceRegistryReader_GetGasPriceUpdatesCreatedAfter_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceRegistryReaderClient) GetAllGasPriceUpdatesCreatedAfter(ctx context.Context, in *GetAllGasPriceUpdatesCreatedAfterRequest, opts ...grpc.CallOption) (*GetAllGasPriceUpdatesCreatedAfterResponse, error) {
	out := new(GetAllGasPriceUpdatesCreatedAfterResponse)
	err := c.cc.Invoke(ctx, PriceRegistryReader_GetAllGasPriceUpdatesCreatedAfter_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceRegistryReaderClient) GetAddress(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetPriceRegistryAddressResponse, error) {
	out := new(GetPriceRegistryAddressResponse)
	err := c.cc.Invoke(ctx, PriceRegistryReader_GetAddress_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceRegistryReaderClient) GetFeeTokens(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetFeeTokensResponse, error) {
	out := new(GetFeeTokensResponse)
	err := c.cc.Invoke(ctx, PriceRegistryReader_GetFeeTokens_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceRegistryReaderClient) GetTokenPrices(ctx context.Context, in *GetTokenPricesRequest, opts ...grpc.CallOption) (*GetTokenPricesResponse, error) {
	out := new(GetTokenPricesResponse)
	err := c.cc.Invoke(ctx, PriceRegistryReader_GetTokenPrices_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceRegistryReaderClient) GetTokensDecimals(ctx context.Context, in *GetTokensDecimalsRequest, opts ...grpc.CallOption) (*GetTokensDecimalsResponse, error) {
	out := new(GetTokensDecimalsResponse)
	err := c.cc.Invoke(ctx, PriceRegistryReader_GetTokensDecimals_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceRegistryReaderClient) Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, PriceRegistryReader_Close_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PriceRegistryReaderServer is the server API for PriceRegistryReader service.
// All implementations must embed UnimplementedPriceRegistryReaderServer
// for forward compatibility
type PriceRegistryReaderServer interface {
	GetTokenPriceUpdatesCreatedAfter(context.Context, *GetTokenPriceUpdatesCreatedAfterRequest) (*GetTokenPriceUpdatesCreatedAfterResponse, error)
	GetGasPriceUpdatesCreatedAfter(context.Context, *GetGasPriceUpdatesCreatedAfterRequest) (*GetGasPriceUpdatesCreatedAfterResponse, error)
	GetAllGasPriceUpdatesCreatedAfter(context.Context, *GetAllGasPriceUpdatesCreatedAfterRequest) (*GetAllGasPriceUpdatesCreatedAfterResponse, error)
	GetAddress(context.Context, *emptypb.Empty) (*GetPriceRegistryAddressResponse, error)
	GetFeeTokens(context.Context, *emptypb.Empty) (*GetFeeTokensResponse, error)
	GetTokenPrices(context.Context, *GetTokenPricesRequest) (*GetTokenPricesResponse, error)
	GetTokensDecimals(context.Context, *GetTokensDecimalsRequest) (*GetTokensDecimalsResponse, error)
	Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedPriceRegistryReaderServer()
}

// UnimplementedPriceRegistryReaderServer must be embedded to have forward compatible implementations.
type UnimplementedPriceRegistryReaderServer struct {
}

func (UnimplementedPriceRegistryReaderServer) GetTokenPriceUpdatesCreatedAfter(context.Context, *GetTokenPriceUpdatesCreatedAfterRequest) (*GetTokenPriceUpdatesCreatedAfterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTokenPriceUpdatesCreatedAfter not implemented")
}
func (UnimplementedPriceRegistryReaderServer) GetGasPriceUpdatesCreatedAfter(context.Context, *GetGasPriceUpdatesCreatedAfterRequest) (*GetGasPriceUpdatesCreatedAfterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGasPriceUpdatesCreatedAfter not implemented")
}
func (UnimplementedPriceRegistryReaderServer) GetAllGasPriceUpdatesCreatedAfter(context.Context, *GetAllGasPriceUpdatesCreatedAfterRequest) (*GetAllGasPriceUpdatesCreatedAfterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllGasPriceUpdatesCreatedAfter not implemented")
}
func (UnimplementedPriceRegistryReaderServer) GetAddress(context.Context, *emptypb.Empty) (*GetPriceRegistryAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAddress not implemented")
}
func (UnimplementedPriceRegistryReaderServer) GetFeeTokens(context.Context, *emptypb.Empty) (*GetFeeTokensResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFeeTokens not implemented")
}
func (UnimplementedPriceRegistryReaderServer) GetTokenPrices(context.Context, *GetTokenPricesRequest) (*GetTokenPricesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTokenPrices not implemented")
}
func (UnimplementedPriceRegistryReaderServer) GetTokensDecimals(context.Context, *GetTokensDecimalsRequest) (*GetTokensDecimalsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTokensDecimals not implemented")
}
func (UnimplementedPriceRegistryReaderServer) Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
func (UnimplementedPriceRegistryReaderServer) mustEmbedUnimplementedPriceRegistryReaderServer() {}

// UnsafePriceRegistryReaderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PriceRegistryReaderServer will
// result in compilation errors.
type UnsafePriceRegistryReaderServer interface {
	mustEmbedUnimplementedPriceRegistryReaderServer()
}

func RegisterPriceRegistryReaderServer(s grpc.ServiceRegistrar, srv PriceRegistryReaderServer) {
	s.RegisterService(&PriceRegistryReader_ServiceDesc, srv)
}

func _PriceRegistryReader_GetTokenPriceUpdatesCreatedAfter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTokenPriceUpdatesCreatedAfterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceRegistryReaderServer).GetTokenPriceUpdatesCreatedAfter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceRegistryReader_GetTokenPriceUpdatesCreatedAfter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceRegistryReaderServer).GetTokenPriceUpdatesCreatedAfter(ctx, req.(*GetTokenPriceUpdatesCreatedAfterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceRegistryReader_GetGasPriceUpdatesCreatedAfter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGasPriceUpdatesCreatedAfterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceRegistryReaderServer).GetGasPriceUpdatesCreatedAfter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceRegistryReader_GetGasPriceUpdatesCreatedAfter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceRegistryReaderServer).GetGasPriceUpdatesCreatedAfter(ctx, req.(*GetGasPriceUpdatesCreatedAfterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceRegistryReader_GetAllGasPriceUpdatesCreatedAfter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllGasPriceUpdatesCreatedAfterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceRegistryReaderServer).GetAllGasPriceUpdatesCreatedAfter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceRegistryReader_GetAllGasPriceUpdatesCreatedAfter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceRegistryReaderServer).GetAllGasPriceUpdatesCreatedAfter(ctx, req.(*GetAllGasPriceUpdatesCreatedAfterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceRegistryReader_GetAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceRegistryReaderServer).GetAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceRegistryReader_GetAddress_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceRegistryReaderServer).GetAddress(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceRegistryReader_GetFeeTokens_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceRegistryReaderServer).GetFeeTokens(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceRegistryReader_GetFeeTokens_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceRegistryReaderServer).GetFeeTokens(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceRegistryReader_GetTokenPrices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTokenPricesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceRegistryReaderServer).GetTokenPrices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceRegistryReader_GetTokenPrices_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceRegistryReaderServer).GetTokenPrices(ctx, req.(*GetTokenPricesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceRegistryReader_GetTokensDecimals_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTokensDecimalsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceRegistryReaderServer).GetTokensDecimals(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceRegistryReader_GetTokensDecimals_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceRegistryReaderServer).GetTokensDecimals(ctx, req.(*GetTokensDecimalsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceRegistryReader_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceRegistryReaderServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceRegistryReader_Close_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceRegistryReaderServer).Close(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// PriceRegistryReader_ServiceDesc is the grpc.ServiceDesc for PriceRegistryReader service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PriceRegistryReader_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.internal.pb.ccip.PriceRegistryReader",
	HandlerType: (*PriceRegistryReaderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetTokenPriceUpdatesCreatedAfter",
			Handler:    _PriceRegistryReader_GetTokenPriceUpdatesCreatedAfter_Handler,
		},
		{
			MethodName: "GetGasPriceUpdatesCreatedAfter",
			Handler:    _PriceRegistryReader_GetGasPriceUpdatesCreatedAfter_Handler,
		},
		{
			MethodName: "GetAllGasPriceUpdatesCreatedAfter",
			Handler:    _PriceRegistryReader_GetAllGasPriceUpdatesCreatedAfter_Handler,
		},
		{
			MethodName: "GetAddress",
			Handler:    _PriceRegistryReader_GetAddress_Handler,
		},
		{
			MethodName: "GetFeeTokens",
			Handler:    _PriceRegistryReader_GetFeeTokens_Handler,
		},
		{
			MethodName: "GetTokenPrices",
			Handler:    _PriceRegistryReader_GetTokenPrices_Handler,
		},
		{
			MethodName: "GetTokensDecimals",
			Handler:    _PriceRegistryReader_GetTokensDecimals_Handler,
		},
		{
			MethodName: "Close",
			Handler:    _PriceRegistryReader_Close_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "priceregistry.proto",
}
