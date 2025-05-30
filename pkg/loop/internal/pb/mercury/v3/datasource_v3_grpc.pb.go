// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: loop/internal/pb/mercury/v3/datasource_v3.proto

package mercuryv3pb

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
	DataSource_Observe_FullMethodName = "/loop.internal.pb.mercury.v3.DataSource/Observe"
)

// DataSourceClient is the client API for DataSource service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// DataSource is a gRPC adapter for [pkg/types/mercury/v1/DataSource]
type DataSourceClient interface {
	Observe(ctx context.Context, in *ObserveRequest, opts ...grpc.CallOption) (*ObserveResponse, error)
}

type dataSourceClient struct {
	cc grpc.ClientConnInterface
}

func NewDataSourceClient(cc grpc.ClientConnInterface) DataSourceClient {
	return &dataSourceClient{cc}
}

func (c *dataSourceClient) Observe(ctx context.Context, in *ObserveRequest, opts ...grpc.CallOption) (*ObserveResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ObserveResponse)
	err := c.cc.Invoke(ctx, DataSource_Observe_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataSourceServer is the server API for DataSource service.
// All implementations must embed UnimplementedDataSourceServer
// for forward compatibility.
//
// DataSource is a gRPC adapter for [pkg/types/mercury/v1/DataSource]
type DataSourceServer interface {
	Observe(context.Context, *ObserveRequest) (*ObserveResponse, error)
	mustEmbedUnimplementedDataSourceServer()
}

// UnimplementedDataSourceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedDataSourceServer struct{}

func (UnimplementedDataSourceServer) Observe(context.Context, *ObserveRequest) (*ObserveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Observe not implemented")
}
func (UnimplementedDataSourceServer) mustEmbedUnimplementedDataSourceServer() {}
func (UnimplementedDataSourceServer) testEmbeddedByValue()                    {}

// UnsafeDataSourceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataSourceServer will
// result in compilation errors.
type UnsafeDataSourceServer interface {
	mustEmbedUnimplementedDataSourceServer()
}

func RegisterDataSourceServer(s grpc.ServiceRegistrar, srv DataSourceServer) {
	// If the following call pancis, it indicates UnimplementedDataSourceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&DataSource_ServiceDesc, srv)
}

func _DataSource_Observe_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ObserveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataSourceServer).Observe(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataSource_Observe_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataSourceServer).Observe(ctx, req.(*ObserveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DataSource_ServiceDesc is the grpc.ServiceDesc for DataSource service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataSource_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.internal.pb.mercury.v3.DataSource",
	HandlerType: (*DataSourceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Observe",
			Handler:    _DataSource_Observe_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "loop/internal/pb/mercury/v3/datasource_v3.proto",
}
