// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: oraclefactory.proto

package oraclefactory

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	OracleFactory_NewOracle_FullMethodName = "/loop.OracleFactory/NewOracle"
)

// OracleFactoryClient is the client API for OracleFactory service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OracleFactoryClient interface {
	NewOracle(ctx context.Context, in *NewOracleRequest, opts ...grpc.CallOption) (*NewOracleReply, error)
}

type oracleFactoryClient struct {
	cc grpc.ClientConnInterface
}

func NewOracleFactoryClient(cc grpc.ClientConnInterface) OracleFactoryClient {
	return &oracleFactoryClient{cc}
}

func (c *oracleFactoryClient) NewOracle(ctx context.Context, in *NewOracleRequest, opts ...grpc.CallOption) (*NewOracleReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(NewOracleReply)
	err := c.cc.Invoke(ctx, OracleFactory_NewOracle_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OracleFactoryServer is the server API for OracleFactory service.
// All implementations must embed UnimplementedOracleFactoryServer
// for forward compatibility.
type OracleFactoryServer interface {
	NewOracle(context.Context, *NewOracleRequest) (*NewOracleReply, error)
	mustEmbedUnimplementedOracleFactoryServer()
}

// UnimplementedOracleFactoryServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedOracleFactoryServer struct{}

func (UnimplementedOracleFactoryServer) NewOracle(context.Context, *NewOracleRequest) (*NewOracleReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewOracle not implemented")
}
func (UnimplementedOracleFactoryServer) mustEmbedUnimplementedOracleFactoryServer() {}
func (UnimplementedOracleFactoryServer) testEmbeddedByValue()                       {}

// UnsafeOracleFactoryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OracleFactoryServer will
// result in compilation errors.
type UnsafeOracleFactoryServer interface {
	mustEmbedUnimplementedOracleFactoryServer()
}

func RegisterOracleFactoryServer(s grpc.ServiceRegistrar, srv OracleFactoryServer) {
	// If the following call pancis, it indicates UnimplementedOracleFactoryServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&OracleFactory_ServiceDesc, srv)
}

func _OracleFactory_NewOracle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewOracleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OracleFactoryServer).NewOracle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OracleFactory_NewOracle_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OracleFactoryServer).NewOracle(ctx, req.(*NewOracleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// OracleFactory_ServiceDesc is the grpc.ServiceDesc for OracleFactory service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OracleFactory_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.OracleFactory",
	HandlerType: (*OracleFactoryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewOracle",
			Handler:    _OracleFactory_NewOracle_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "oraclefactory.proto",
}
