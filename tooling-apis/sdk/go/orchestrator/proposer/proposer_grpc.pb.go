// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.3
// source: orchestrator/proposer/proposer.proto

package proposer

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	FeedsManager_ApprovedJob_FullMethodName  = "/tooling.orchestrator.proposer.FeedsManager/ApprovedJob"
	FeedsManager_Healthcheck_FullMethodName  = "/tooling.orchestrator.proposer.FeedsManager/Healthcheck"
	FeedsManager_UpdateNode_FullMethodName   = "/tooling.orchestrator.proposer.FeedsManager/UpdateNode"
	FeedsManager_RejectedJob_FullMethodName  = "/tooling.orchestrator.proposer.FeedsManager/RejectedJob"
	FeedsManager_CancelledJob_FullMethodName = "/tooling.orchestrator.proposer.FeedsManager/CancelledJob"
)

// FeedsManagerClient is the client API for FeedsManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FeedsManagerClient interface {
	ApprovedJob(ctx context.Context, in *ApprovedJobRequest, opts ...grpc.CallOption) (*ApprovedJobResponse, error)
	Healthcheck(ctx context.Context, in *HealthcheckRequest, opts ...grpc.CallOption) (*HealthcheckResponse, error)
	UpdateNode(ctx context.Context, in *UpdateNodeRequest, opts ...grpc.CallOption) (*UpdateNodeResponse, error)
	RejectedJob(ctx context.Context, in *RejectedJobRequest, opts ...grpc.CallOption) (*RejectedJobResponse, error)
	CancelledJob(ctx context.Context, in *CancelledJobRequest, opts ...grpc.CallOption) (*CancelledJobResponse, error)
}

type feedsManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewFeedsManagerClient(cc grpc.ClientConnInterface) FeedsManagerClient {
	return &feedsManagerClient{cc}
}

func (c *feedsManagerClient) ApprovedJob(ctx context.Context, in *ApprovedJobRequest, opts ...grpc.CallOption) (*ApprovedJobResponse, error) {
	out := new(ApprovedJobResponse)
	err := c.cc.Invoke(ctx, FeedsManager_ApprovedJob_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feedsManagerClient) Healthcheck(ctx context.Context, in *HealthcheckRequest, opts ...grpc.CallOption) (*HealthcheckResponse, error) {
	out := new(HealthcheckResponse)
	err := c.cc.Invoke(ctx, FeedsManager_Healthcheck_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feedsManagerClient) UpdateNode(ctx context.Context, in *UpdateNodeRequest, opts ...grpc.CallOption) (*UpdateNodeResponse, error) {
	out := new(UpdateNodeResponse)
	err := c.cc.Invoke(ctx, FeedsManager_UpdateNode_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feedsManagerClient) RejectedJob(ctx context.Context, in *RejectedJobRequest, opts ...grpc.CallOption) (*RejectedJobResponse, error) {
	out := new(RejectedJobResponse)
	err := c.cc.Invoke(ctx, FeedsManager_RejectedJob_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feedsManagerClient) CancelledJob(ctx context.Context, in *CancelledJobRequest, opts ...grpc.CallOption) (*CancelledJobResponse, error) {
	out := new(CancelledJobResponse)
	err := c.cc.Invoke(ctx, FeedsManager_CancelledJob_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FeedsManagerServer is the server API for FeedsManager service.
// All implementations must embed UnimplementedFeedsManagerServer
// for forward compatibility
type FeedsManagerServer interface {
	ApprovedJob(context.Context, *ApprovedJobRequest) (*ApprovedJobResponse, error)
	Healthcheck(context.Context, *HealthcheckRequest) (*HealthcheckResponse, error)
	UpdateNode(context.Context, *UpdateNodeRequest) (*UpdateNodeResponse, error)
	RejectedJob(context.Context, *RejectedJobRequest) (*RejectedJobResponse, error)
	CancelledJob(context.Context, *CancelledJobRequest) (*CancelledJobResponse, error)
	mustEmbedUnimplementedFeedsManagerServer()
}

// UnimplementedFeedsManagerServer must be embedded to have forward compatible implementations.
type UnimplementedFeedsManagerServer struct {
}

func (UnimplementedFeedsManagerServer) ApprovedJob(context.Context, *ApprovedJobRequest) (*ApprovedJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApprovedJob not implemented")
}
func (UnimplementedFeedsManagerServer) Healthcheck(context.Context, *HealthcheckRequest) (*HealthcheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Healthcheck not implemented")
}
func (UnimplementedFeedsManagerServer) UpdateNode(context.Context, *UpdateNodeRequest) (*UpdateNodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateNode not implemented")
}
func (UnimplementedFeedsManagerServer) RejectedJob(context.Context, *RejectedJobRequest) (*RejectedJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RejectedJob not implemented")
}
func (UnimplementedFeedsManagerServer) CancelledJob(context.Context, *CancelledJobRequest) (*CancelledJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelledJob not implemented")
}
func (UnimplementedFeedsManagerServer) mustEmbedUnimplementedFeedsManagerServer() {}

// UnsafeFeedsManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FeedsManagerServer will
// result in compilation errors.
type UnsafeFeedsManagerServer interface {
	mustEmbedUnimplementedFeedsManagerServer()
}

func RegisterFeedsManagerServer(s grpc.ServiceRegistrar, srv FeedsManagerServer) {
	s.RegisterService(&FeedsManager_ServiceDesc, srv)
}

func _FeedsManager_ApprovedJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApprovedJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeedsManagerServer).ApprovedJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FeedsManager_ApprovedJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeedsManagerServer).ApprovedJob(ctx, req.(*ApprovedJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeedsManager_Healthcheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthcheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeedsManagerServer).Healthcheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FeedsManager_Healthcheck_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeedsManagerServer).Healthcheck(ctx, req.(*HealthcheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeedsManager_UpdateNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeedsManagerServer).UpdateNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FeedsManager_UpdateNode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeedsManagerServer).UpdateNode(ctx, req.(*UpdateNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeedsManager_RejectedJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RejectedJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeedsManagerServer).RejectedJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FeedsManager_RejectedJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeedsManagerServer).RejectedJob(ctx, req.(*RejectedJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeedsManager_CancelledJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelledJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeedsManagerServer).CancelledJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FeedsManager_CancelledJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeedsManagerServer).CancelledJob(ctx, req.(*CancelledJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FeedsManager_ServiceDesc is the grpc.ServiceDesc for FeedsManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FeedsManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tooling.orchestrator.proposer.FeedsManager",
	HandlerType: (*FeedsManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ApprovedJob",
			Handler:    _FeedsManager_ApprovedJob_Handler,
		},
		{
			MethodName: "Healthcheck",
			Handler:    _FeedsManager_Healthcheck_Handler,
		},
		{
			MethodName: "UpdateNode",
			Handler:    _FeedsManager_UpdateNode_Handler,
		},
		{
			MethodName: "RejectedJob",
			Handler:    _FeedsManager_RejectedJob_Handler,
		},
		{
			MethodName: "CancelledJob",
			Handler:    _FeedsManager_CancelledJob_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "orchestrator/proposer/proposer.proto",
}
