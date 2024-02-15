// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: pkg/loop/internal/pb/mercury/mercury_plugin.proto

// note: the generate.go file in this dir specifies the import path of the relative proto files

package mercurypb

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
	MercuryPluginFactory_NewMercuryPlugin_FullMethodName = "/loop.internal.pb.mercury.MercuryPluginFactory/NewMercuryPlugin"
)

// MercuryPluginFactoryClient is the client API for MercuryPluginFactory service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MercuryPluginFactoryClient interface {
	NewMercuryPlugin(ctx context.Context, in *NewMercuryPluginRequest, opts ...grpc.CallOption) (*NewMercuryPluginResponse, error)
}

type mercuryPluginFactoryClient struct {
	cc grpc.ClientConnInterface
}

func NewMercuryPluginFactoryClient(cc grpc.ClientConnInterface) MercuryPluginFactoryClient {
	return &mercuryPluginFactoryClient{cc}
}

func (c *mercuryPluginFactoryClient) NewMercuryPlugin(ctx context.Context, in *NewMercuryPluginRequest, opts ...grpc.CallOption) (*NewMercuryPluginResponse, error) {
	out := new(NewMercuryPluginResponse)
	err := c.cc.Invoke(ctx, MercuryPluginFactory_NewMercuryPlugin_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MercuryPluginFactoryServer is the server API for MercuryPluginFactory service.
// All implementations must embed UnimplementedMercuryPluginFactoryServer
// for forward compatibility
type MercuryPluginFactoryServer interface {
	NewMercuryPlugin(context.Context, *NewMercuryPluginRequest) (*NewMercuryPluginResponse, error)
	mustEmbedUnimplementedMercuryPluginFactoryServer()
}

// UnimplementedMercuryPluginFactoryServer must be embedded to have forward compatible implementations.
type UnimplementedMercuryPluginFactoryServer struct {
}

func (UnimplementedMercuryPluginFactoryServer) NewMercuryPlugin(context.Context, *NewMercuryPluginRequest) (*NewMercuryPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewMercuryPlugin not implemented")
}
func (UnimplementedMercuryPluginFactoryServer) mustEmbedUnimplementedMercuryPluginFactoryServer() {}

// UnsafeMercuryPluginFactoryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MercuryPluginFactoryServer will
// result in compilation errors.
type UnsafeMercuryPluginFactoryServer interface {
	mustEmbedUnimplementedMercuryPluginFactoryServer()
}

func RegisterMercuryPluginFactoryServer(s grpc.ServiceRegistrar, srv MercuryPluginFactoryServer) {
	s.RegisterService(&MercuryPluginFactory_ServiceDesc, srv)
}

func _MercuryPluginFactory_NewMercuryPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewMercuryPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MercuryPluginFactoryServer).NewMercuryPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MercuryPluginFactory_NewMercuryPlugin_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MercuryPluginFactoryServer).NewMercuryPlugin(ctx, req.(*NewMercuryPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MercuryPluginFactory_ServiceDesc is the grpc.ServiceDesc for MercuryPluginFactory service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MercuryPluginFactory_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.internal.pb.mercury.MercuryPluginFactory",
	HandlerType: (*MercuryPluginFactoryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewMercuryPlugin",
			Handler:    _MercuryPluginFactory_NewMercuryPlugin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/loop/internal/pb/mercury/mercury_plugin.proto",
}

const (
	MercuryPlugin_Observation_FullMethodName = "/loop.internal.pb.mercury.MercuryPlugin/Observation"
	MercuryPlugin_Report_FullMethodName      = "/loop.internal.pb.mercury.MercuryPlugin/Report"
	MercuryPlugin_Close_FullMethodName       = "/loop.internal.pb.mercury.MercuryPlugin/Close"
)

// MercuryPluginClient is the client API for MercuryPlugin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MercuryPluginClient interface {
	Observation(ctx context.Context, in *ObservationRequest, opts ...grpc.CallOption) (*ObservationResponse, error)
	Report(ctx context.Context, in *ReportRequest, opts ...grpc.CallOption) (*ReportResponse, error)
	Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type mercuryPluginClient struct {
	cc grpc.ClientConnInterface
}

func NewMercuryPluginClient(cc grpc.ClientConnInterface) MercuryPluginClient {
	return &mercuryPluginClient{cc}
}

func (c *mercuryPluginClient) Observation(ctx context.Context, in *ObservationRequest, opts ...grpc.CallOption) (*ObservationResponse, error) {
	out := new(ObservationResponse)
	err := c.cc.Invoke(ctx, MercuryPlugin_Observation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mercuryPluginClient) Report(ctx context.Context, in *ReportRequest, opts ...grpc.CallOption) (*ReportResponse, error) {
	out := new(ReportResponse)
	err := c.cc.Invoke(ctx, MercuryPlugin_Report_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mercuryPluginClient) Close(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, MercuryPlugin_Close_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MercuryPluginServer is the server API for MercuryPlugin service.
// All implementations must embed UnimplementedMercuryPluginServer
// for forward compatibility
type MercuryPluginServer interface {
	Observation(context.Context, *ObservationRequest) (*ObservationResponse, error)
	Report(context.Context, *ReportRequest) (*ReportResponse, error)
	Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedMercuryPluginServer()
}

// UnimplementedMercuryPluginServer must be embedded to have forward compatible implementations.
type UnimplementedMercuryPluginServer struct {
}

func (UnimplementedMercuryPluginServer) Observation(context.Context, *ObservationRequest) (*ObservationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Observation not implemented")
}
func (UnimplementedMercuryPluginServer) Report(context.Context, *ReportRequest) (*ReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Report not implemented")
}
func (UnimplementedMercuryPluginServer) Close(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
func (UnimplementedMercuryPluginServer) mustEmbedUnimplementedMercuryPluginServer() {}

// UnsafeMercuryPluginServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MercuryPluginServer will
// result in compilation errors.
type UnsafeMercuryPluginServer interface {
	mustEmbedUnimplementedMercuryPluginServer()
}

func RegisterMercuryPluginServer(s grpc.ServiceRegistrar, srv MercuryPluginServer) {
	s.RegisterService(&MercuryPlugin_ServiceDesc, srv)
}

func _MercuryPlugin_Observation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ObservationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MercuryPluginServer).Observation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MercuryPlugin_Observation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MercuryPluginServer).Observation(ctx, req.(*ObservationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MercuryPlugin_Report_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MercuryPluginServer).Report(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MercuryPlugin_Report_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MercuryPluginServer).Report(ctx, req.(*ReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MercuryPlugin_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MercuryPluginServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MercuryPlugin_Close_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MercuryPluginServer).Close(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// MercuryPlugin_ServiceDesc is the grpc.ServiceDesc for MercuryPlugin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MercuryPlugin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.internal.pb.mercury.MercuryPlugin",
	HandlerType: (*MercuryPluginServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Observation",
			Handler:    _MercuryPlugin_Observation_Handler,
		},
		{
			MethodName: "Report",
			Handler:    _MercuryPlugin_Report_Handler,
		},
		{
			MethodName: "Close",
			Handler:    _MercuryPlugin_Close_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/loop/internal/pb/mercury/mercury_plugin.proto",
}
