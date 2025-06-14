// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: capabilities/v2/chain-capabilities/evm/capability.proto

package evm

import (
	context "context"
	evm "github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Client_CallContract_FullMethodName           = "/cre.sdk.v2.evm.Client/CallContract"
	Client_FilterLogs_FullMethodName             = "/cre.sdk.v2.evm.Client/FilterLogs"
	Client_BalanceAt_FullMethodName              = "/cre.sdk.v2.evm.Client/BalanceAt"
	Client_EstimateGas_FullMethodName            = "/cre.sdk.v2.evm.Client/EstimateGas"
	Client_GetTransactionByHash_FullMethodName   = "/cre.sdk.v2.evm.Client/GetTransactionByHash"
	Client_GetTransactionReceipt_FullMethodName  = "/cre.sdk.v2.evm.Client/GetTransactionReceipt"
	Client_LatestAndFinalizedHead_FullMethodName = "/cre.sdk.v2.evm.Client/LatestAndFinalizedHead"
	Client_QueryTrackedLogs_FullMethodName       = "/cre.sdk.v2.evm.Client/QueryTrackedLogs"
	Client_RegisterLogTracking_FullMethodName    = "/cre.sdk.v2.evm.Client/RegisterLogTracking"
	Client_UnregisterLogTracking_FullMethodName  = "/cre.sdk.v2.evm.Client/UnregisterLogTracking"
)

// ClientClient is the client API for Client service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClientClient interface {
	CallContract(ctx context.Context, in *evm.CallContractRequest, opts ...grpc.CallOption) (*evm.CallContractReply, error)
	FilterLogs(ctx context.Context, in *evm.FilterLogsRequest, opts ...grpc.CallOption) (*evm.FilterLogsReply, error)
	BalanceAt(ctx context.Context, in *evm.BalanceAtRequest, opts ...grpc.CallOption) (*evm.BalanceAtReply, error)
	EstimateGas(ctx context.Context, in *evm.EstimateGasRequest, opts ...grpc.CallOption) (*evm.EstimateGasReply, error)
	GetTransactionByHash(ctx context.Context, in *evm.GetTransactionByHashRequest, opts ...grpc.CallOption) (*evm.GetTransactionByHashReply, error)
	GetTransactionReceipt(ctx context.Context, in *evm.GetTransactionReceiptRequest, opts ...grpc.CallOption) (*evm.GetTransactionReceiptReply, error)
	LatestAndFinalizedHead(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*evm.LatestAndFinalizedHeadReply, error)
	QueryTrackedLogs(ctx context.Context, in *evm.QueryTrackedLogsRequest, opts ...grpc.CallOption) (*evm.QueryTrackedLogsReply, error)
	RegisterLogTracking(ctx context.Context, in *evm.RegisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	UnregisterLogTracking(ctx context.Context, in *evm.UnregisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type clientClient struct {
	cc grpc.ClientConnInterface
}

func NewClientClient(cc grpc.ClientConnInterface) ClientClient {
	return &clientClient{cc}
}

func (c *clientClient) CallContract(ctx context.Context, in *evm.CallContractRequest, opts ...grpc.CallOption) (*evm.CallContractReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(evm.CallContractReply)
	err := c.cc.Invoke(ctx, Client_CallContract_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientClient) FilterLogs(ctx context.Context, in *evm.FilterLogsRequest, opts ...grpc.CallOption) (*evm.FilterLogsReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(evm.FilterLogsReply)
	err := c.cc.Invoke(ctx, Client_FilterLogs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientClient) BalanceAt(ctx context.Context, in *evm.BalanceAtRequest, opts ...grpc.CallOption) (*evm.BalanceAtReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(evm.BalanceAtReply)
	err := c.cc.Invoke(ctx, Client_BalanceAt_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientClient) EstimateGas(ctx context.Context, in *evm.EstimateGasRequest, opts ...grpc.CallOption) (*evm.EstimateGasReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(evm.EstimateGasReply)
	err := c.cc.Invoke(ctx, Client_EstimateGas_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientClient) GetTransactionByHash(ctx context.Context, in *evm.GetTransactionByHashRequest, opts ...grpc.CallOption) (*evm.GetTransactionByHashReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(evm.GetTransactionByHashReply)
	err := c.cc.Invoke(ctx, Client_GetTransactionByHash_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientClient) GetTransactionReceipt(ctx context.Context, in *evm.GetTransactionReceiptRequest, opts ...grpc.CallOption) (*evm.GetTransactionReceiptReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(evm.GetTransactionReceiptReply)
	err := c.cc.Invoke(ctx, Client_GetTransactionReceipt_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientClient) LatestAndFinalizedHead(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*evm.LatestAndFinalizedHeadReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(evm.LatestAndFinalizedHeadReply)
	err := c.cc.Invoke(ctx, Client_LatestAndFinalizedHead_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientClient) QueryTrackedLogs(ctx context.Context, in *evm.QueryTrackedLogsRequest, opts ...grpc.CallOption) (*evm.QueryTrackedLogsReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(evm.QueryTrackedLogsReply)
	err := c.cc.Invoke(ctx, Client_QueryTrackedLogs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientClient) RegisterLogTracking(ctx context.Context, in *evm.RegisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Client_RegisterLogTracking_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientClient) UnregisterLogTracking(ctx context.Context, in *evm.UnregisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Client_UnregisterLogTracking_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClientServer is the server API for Client service.
// All implementations must embed UnimplementedClientServer
// for forward compatibility.
type ClientServer interface {
	CallContract(context.Context, *evm.CallContractRequest) (*evm.CallContractReply, error)
	FilterLogs(context.Context, *evm.FilterLogsRequest) (*evm.FilterLogsReply, error)
	BalanceAt(context.Context, *evm.BalanceAtRequest) (*evm.BalanceAtReply, error)
	EstimateGas(context.Context, *evm.EstimateGasRequest) (*evm.EstimateGasReply, error)
	GetTransactionByHash(context.Context, *evm.GetTransactionByHashRequest) (*evm.GetTransactionByHashReply, error)
	GetTransactionReceipt(context.Context, *evm.GetTransactionReceiptRequest) (*evm.GetTransactionReceiptReply, error)
	LatestAndFinalizedHead(context.Context, *emptypb.Empty) (*evm.LatestAndFinalizedHeadReply, error)
	QueryTrackedLogs(context.Context, *evm.QueryTrackedLogsRequest) (*evm.QueryTrackedLogsReply, error)
	RegisterLogTracking(context.Context, *evm.RegisterLogTrackingRequest) (*emptypb.Empty, error)
	UnregisterLogTracking(context.Context, *evm.UnregisterLogTrackingRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedClientServer()
}

// UnimplementedClientServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedClientServer struct{}

func (UnimplementedClientServer) CallContract(context.Context, *evm.CallContractRequest) (*evm.CallContractReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CallContract not implemented")
}
func (UnimplementedClientServer) FilterLogs(context.Context, *evm.FilterLogsRequest) (*evm.FilterLogsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FilterLogs not implemented")
}
func (UnimplementedClientServer) BalanceAt(context.Context, *evm.BalanceAtRequest) (*evm.BalanceAtReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BalanceAt not implemented")
}
func (UnimplementedClientServer) EstimateGas(context.Context, *evm.EstimateGasRequest) (*evm.EstimateGasReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EstimateGas not implemented")
}
func (UnimplementedClientServer) GetTransactionByHash(context.Context, *evm.GetTransactionByHashRequest) (*evm.GetTransactionByHashReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransactionByHash not implemented")
}
func (UnimplementedClientServer) GetTransactionReceipt(context.Context, *evm.GetTransactionReceiptRequest) (*evm.GetTransactionReceiptReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransactionReceipt not implemented")
}
func (UnimplementedClientServer) LatestAndFinalizedHead(context.Context, *emptypb.Empty) (*evm.LatestAndFinalizedHeadReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LatestAndFinalizedHead not implemented")
}
func (UnimplementedClientServer) QueryTrackedLogs(context.Context, *evm.QueryTrackedLogsRequest) (*evm.QueryTrackedLogsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryTrackedLogs not implemented")
}
func (UnimplementedClientServer) RegisterLogTracking(context.Context, *evm.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterLogTracking not implemented")
}
func (UnimplementedClientServer) UnregisterLogTracking(context.Context, *evm.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnregisterLogTracking not implemented")
}
func (UnimplementedClientServer) mustEmbedUnimplementedClientServer() {}
func (UnimplementedClientServer) testEmbeddedByValue()                {}

// UnsafeClientServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientServer will
// result in compilation errors.
type UnsafeClientServer interface {
	mustEmbedUnimplementedClientServer()
}

func RegisterClientServer(s grpc.ServiceRegistrar, srv ClientServer) {
	// If the following call pancis, it indicates UnimplementedClientServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Client_ServiceDesc, srv)
}

func _Client_CallContract_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(evm.CallContractRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).CallContract(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_CallContract_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).CallContract(ctx, req.(*evm.CallContractRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Client_FilterLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(evm.FilterLogsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).FilterLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_FilterLogs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).FilterLogs(ctx, req.(*evm.FilterLogsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Client_BalanceAt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(evm.BalanceAtRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).BalanceAt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_BalanceAt_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).BalanceAt(ctx, req.(*evm.BalanceAtRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Client_EstimateGas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(evm.EstimateGasRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).EstimateGas(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_EstimateGas_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).EstimateGas(ctx, req.(*evm.EstimateGasRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Client_GetTransactionByHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(evm.GetTransactionByHashRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).GetTransactionByHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_GetTransactionByHash_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).GetTransactionByHash(ctx, req.(*evm.GetTransactionByHashRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Client_GetTransactionReceipt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(evm.GetTransactionReceiptRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).GetTransactionReceipt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_GetTransactionReceipt_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).GetTransactionReceipt(ctx, req.(*evm.GetTransactionReceiptRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Client_LatestAndFinalizedHead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).LatestAndFinalizedHead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_LatestAndFinalizedHead_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).LatestAndFinalizedHead(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Client_QueryTrackedLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(evm.QueryTrackedLogsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).QueryTrackedLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_QueryTrackedLogs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).QueryTrackedLogs(ctx, req.(*evm.QueryTrackedLogsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Client_RegisterLogTracking_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(evm.RegisterLogTrackingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).RegisterLogTracking(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_RegisterLogTracking_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).RegisterLogTracking(ctx, req.(*evm.RegisterLogTrackingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Client_UnregisterLogTracking_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(evm.UnregisterLogTrackingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServer).UnregisterLogTracking(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Client_UnregisterLogTracking_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServer).UnregisterLogTracking(ctx, req.(*evm.UnregisterLogTrackingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Client_ServiceDesc is the grpc.ServiceDesc for Client service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Client_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cre.sdk.v2.evm.Client",
	HandlerType: (*ClientServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CallContract",
			Handler:    _Client_CallContract_Handler,
		},
		{
			MethodName: "FilterLogs",
			Handler:    _Client_FilterLogs_Handler,
		},
		{
			MethodName: "BalanceAt",
			Handler:    _Client_BalanceAt_Handler,
		},
		{
			MethodName: "EstimateGas",
			Handler:    _Client_EstimateGas_Handler,
		},
		{
			MethodName: "GetTransactionByHash",
			Handler:    _Client_GetTransactionByHash_Handler,
		},
		{
			MethodName: "GetTransactionReceipt",
			Handler:    _Client_GetTransactionReceipt_Handler,
		},
		{
			MethodName: "LatestAndFinalizedHead",
			Handler:    _Client_LatestAndFinalizedHead_Handler,
		},
		{
			MethodName: "QueryTrackedLogs",
			Handler:    _Client_QueryTrackedLogs_Handler,
		},
		{
			MethodName: "RegisterLogTracking",
			Handler:    _Client_RegisterLogTracking_Handler,
		},
		{
			MethodName: "UnregisterLogTracking",
			Handler:    _Client_UnregisterLogTracking_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "capabilities/v2/chain-capabilities/evm/capability.proto",
}
