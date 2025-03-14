// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: microservice/v1/microservice_grpc.proto

package v1

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
	MicroService_Sender_FullMethodName = "/microservice.v1.MicroService/Sender"
)

// MicroServiceClient is the client API for MicroService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MicroServiceClient interface {
	Sender(ctx context.Context, in *SenderRequest, opts ...grpc.CallOption) (*SenderResponse, error)
}

type microServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMicroServiceClient(cc grpc.ClientConnInterface) MicroServiceClient {
	return &microServiceClient{cc}
}

func (c *microServiceClient) Sender(ctx context.Context, in *SenderRequest, opts ...grpc.CallOption) (*SenderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SenderResponse)
	err := c.cc.Invoke(ctx, MicroService_Sender_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MicroServiceServer is the server API for MicroService service.
// All implementations must embed UnimplementedMicroServiceServer
// for forward compatibility.
type MicroServiceServer interface {
	Sender(context.Context, *SenderRequest) (*SenderResponse, error)
	mustEmbedUnimplementedMicroServiceServer()
}

// UnimplementedMicroServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMicroServiceServer struct{}

func (UnimplementedMicroServiceServer) Sender(context.Context, *SenderRequest) (*SenderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sender not implemented")
}
func (UnimplementedMicroServiceServer) mustEmbedUnimplementedMicroServiceServer() {}
func (UnimplementedMicroServiceServer) testEmbeddedByValue()                      {}

// UnsafeMicroServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MicroServiceServer will
// result in compilation errors.
type UnsafeMicroServiceServer interface {
	mustEmbedUnimplementedMicroServiceServer()
}

func RegisterMicroServiceServer(s grpc.ServiceRegistrar, srv MicroServiceServer) {
	// If the following call pancis, it indicates UnimplementedMicroServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MicroService_ServiceDesc, srv)
}

func _MicroService_Sender_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SenderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MicroServiceServer).Sender(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MicroService_Sender_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MicroServiceServer).Sender(ctx, req.(*SenderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MicroService_ServiceDesc is the grpc.ServiceDesc for MicroService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MicroService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "microservice.v1.MicroService",
	HandlerType: (*MicroServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Sender",
			Handler:    _MicroService_Sender_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "microservice/v1/microservice_grpc.proto",
}
