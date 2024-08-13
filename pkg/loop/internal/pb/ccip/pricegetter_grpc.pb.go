// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.4
// source: pricegetter.proto

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
	PriceGetter_FilterConfiguredTokens_FullMethodName = "/loop.internal.pb.ccip.PriceGetter/FilterConfiguredTokens"
	PriceGetter_TokenPricesUSD_FullMethodName         = "/loop.internal.pb.ccip.PriceGetter/TokenPricesUSD"
	PriceGetter_Close_FullMethodName                  = "/loop.internal.pb.ccip.PriceGetter/Close"
)

// PriceGetterClient is the client API for PriceGetter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PriceGetterClient interface {
	FilterConfiguredTokens(ctx context.Context, in *FilterConfiguredTokensRequest, opts ...grpc.CallOption) (*FilterConfiguredTokensResponse, error)
	TokenPricesUSD(ctx context.Context, in *TokenPricesRequest, opts ...grpc.CallOption) (*TokenPricesResponse, error)
	Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type priceGetterClient struct {
	cc grpc.ClientConnInterface
}

func NewPriceGetterClient(cc grpc.ClientConnInterface) PriceGetterClient {
	return &priceGetterClient{cc}
}

func (c *priceGetterClient) FilterConfiguredTokens(ctx context.Context, in *FilterConfiguredTokensRequest, opts ...grpc.CallOption) (*FilterConfiguredTokensResponse, error) {
	out := new(FilterConfiguredTokensResponse)
	err := c.cc.Invoke(ctx, PriceGetter_FilterConfiguredTokens_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceGetterClient) TokenPricesUSD(ctx context.Context, in *TokenPricesRequest, opts ...grpc.CallOption) (*TokenPricesResponse, error) {
	out := new(TokenPricesResponse)
	err := c.cc.Invoke(ctx, PriceGetter_TokenPricesUSD_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *priceGetterClient) Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, PriceGetter_Close_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PriceGetterServer is the server API for PriceGetter service.
// All implementations must embed UnimplementedPriceGetterServer
// for forward compatibility
type PriceGetterServer interface {
	FilterConfiguredTokens(context.Context, *FilterConfiguredTokensRequest) (*FilterConfiguredTokensResponse, error)
	TokenPricesUSD(context.Context, *TokenPricesRequest) (*TokenPricesResponse, error)
	Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedPriceGetterServer()
}

// UnimplementedPriceGetterServer must be embedded to have forward compatible implementations.
type UnimplementedPriceGetterServer struct {
}

func (UnimplementedPriceGetterServer) FilterConfiguredTokens(context.Context, *FilterConfiguredTokensRequest) (*FilterConfiguredTokensResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FilterConfiguredTokens not implemented")
}
func (UnimplementedPriceGetterServer) TokenPricesUSD(context.Context, *TokenPricesRequest) (*TokenPricesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TokenPricesUSD not implemented")
}
func (UnimplementedPriceGetterServer) Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
func (UnimplementedPriceGetterServer) mustEmbedUnimplementedPriceGetterServer() {}

// UnsafePriceGetterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PriceGetterServer will
// result in compilation errors.
type UnsafePriceGetterServer interface {
	mustEmbedUnimplementedPriceGetterServer()
}

func RegisterPriceGetterServer(s grpc.ServiceRegistrar, srv PriceGetterServer) {
	s.RegisterService(&PriceGetter_ServiceDesc, srv)
}

func _PriceGetter_FilterConfiguredTokens_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FilterConfiguredTokensRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceGetterServer).FilterConfiguredTokens(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceGetter_FilterConfiguredTokens_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceGetterServer).FilterConfiguredTokens(ctx, req.(*FilterConfiguredTokensRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceGetter_TokenPricesUSD_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TokenPricesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceGetterServer).TokenPricesUSD(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceGetter_TokenPricesUSD_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceGetterServer).TokenPricesUSD(ctx, req.(*TokenPricesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PriceGetter_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceGetterServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PriceGetter_Close_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceGetterServer).Close(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// PriceGetter_ServiceDesc is the grpc.ServiceDesc for PriceGetter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PriceGetter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.internal.pb.ccip.PriceGetter",
	HandlerType: (*PriceGetterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FilterConfiguredTokens",
			Handler:    _PriceGetter_FilterConfiguredTokens_Handler,
		},
		{
			MethodName: "TokenPricesUSD",
			Handler:    _PriceGetter_TokenPricesUSD_Handler,
		},
		{
			MethodName: "Close",
			Handler:    _PriceGetter_Close_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pricegetter.proto",
}
