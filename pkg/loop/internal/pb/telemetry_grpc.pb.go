// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: telemetry.proto

package pb

import (
	context "context"
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
	Telemetry_Send_FullMethodName = "/loop.Telemetry/Send"
)

// TelemetryClient is the client API for Telemetry service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TelemetryClient interface {
	Send(ctx context.Context, in *TelemetryMessage, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type telemetryClient struct {
	cc grpc.ClientConnInterface
}

func NewTelemetryClient(cc grpc.ClientConnInterface) TelemetryClient {
	return &telemetryClient{cc}
}

func (c *telemetryClient) Send(ctx context.Context, in *TelemetryMessage, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Telemetry_Send_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TelemetryServer is the server API for Telemetry service.
// All implementations must embed UnimplementedTelemetryServer
// for forward compatibility.
type TelemetryServer interface {
	Send(context.Context, *TelemetryMessage) (*emptypb.Empty, error)
	mustEmbedUnimplementedTelemetryServer()
}

// UnimplementedTelemetryServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTelemetryServer struct{}

func (UnimplementedTelemetryServer) Send(context.Context, *TelemetryMessage) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedTelemetryServer) mustEmbedUnimplementedTelemetryServer() {}
func (UnimplementedTelemetryServer) testEmbeddedByValue()                   {}

// UnsafeTelemetryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TelemetryServer will
// result in compilation errors.
type UnsafeTelemetryServer interface {
	mustEmbedUnimplementedTelemetryServer()
}

func RegisterTelemetryServer(s grpc.ServiceRegistrar, srv TelemetryServer) {
	// If the following call pancis, it indicates UnimplementedTelemetryServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Telemetry_ServiceDesc, srv)
}

func _Telemetry_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TelemetryMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TelemetryServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Telemetry_Send_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TelemetryServer).Send(ctx, req.(*TelemetryMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// Telemetry_ServiceDesc is the grpc.ServiceDesc for Telemetry service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Telemetry_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.Telemetry",
	HandlerType: (*TelemetryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _Telemetry_Send_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "telemetry.proto",
}
