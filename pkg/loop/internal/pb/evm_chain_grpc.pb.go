// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: evm_chain.proto

package pb

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
	EVMChain_ReadContract_FullMethodName = "/loop.EVMChain/ReadContract"
)

// EVMChainClient is the client API for EVMChain service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// EVM chain
type EVMChainClient interface {
	ReadContract(ctx context.Context, in *ReadContractRequest, opts ...grpc.CallOption) (*ReadContractReply, error)
}

type eVMChainClient struct {
	cc grpc.ClientConnInterface
}

func NewEVMChainClient(cc grpc.ClientConnInterface) EVMChainClient {
	return &eVMChainClient{cc}
}

func (c *eVMChainClient) ReadContract(ctx context.Context, in *ReadContractRequest, opts ...grpc.CallOption) (*ReadContractReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReadContractReply)
	err := c.cc.Invoke(ctx, EVMChain_ReadContract_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EVMChainServer is the server API for EVMChain service.
// All implementations must embed UnimplementedEVMChainServer
// for forward compatibility.
//
// EVM chain
type EVMChainServer interface {
	ReadContract(context.Context, *ReadContractRequest) (*ReadContractReply, error)
	mustEmbedUnimplementedEVMChainServer()
}

// UnimplementedEVMChainServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedEVMChainServer struct{}

func (UnimplementedEVMChainServer) ReadContract(context.Context, *ReadContractRequest) (*ReadContractReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadContract not implemented")
}
func (UnimplementedEVMChainServer) mustEmbedUnimplementedEVMChainServer() {}
func (UnimplementedEVMChainServer) testEmbeddedByValue()                  {}

// UnsafeEVMChainServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EVMChainServer will
// result in compilation errors.
type UnsafeEVMChainServer interface {
	mustEmbedUnimplementedEVMChainServer()
}

func RegisterEVMChainServer(s grpc.ServiceRegistrar, srv EVMChainServer) {
	// If the following call pancis, it indicates UnimplementedEVMChainServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&EVMChain_ServiceDesc, srv)
}

func _EVMChain_ReadContract_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadContractRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EVMChainServer).ReadContract(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EVMChain_ReadContract_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EVMChainServer).ReadContract(ctx, req.(*ReadContractRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// EVMChain_ServiceDesc is the grpc.ServiceDesc for EVMChain service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EVMChain_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "loop.EVMChain",
	HandlerType: (*EVMChainServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReadContract",
			Handler:    _EVMChain_ReadContract_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "evm_chain.proto",
}
