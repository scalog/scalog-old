// Code generated by protoc-gen-go.
// source: orderpb/order.proto
// DO NOT EDIT!

/*
Package orderpb is a generated protocol buffer package.

It is generated from these files:
	orderpb/order.proto

It has these top-level messages:
	Cut
	ShardView
	ReportRequest
	ReportResponse
	RegisterRequest
	RegisterResponse
	FinalizeRequest
	FinalizeResponse
*/
package orderpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Cut struct {
	// Array of len numReplicas
	Cut []int32 `protobuf:"varint,1,rep,packed,name=cut" json:"cut,omitempty"`
}

func (m *Cut) Reset()                    { *m = Cut{} }
func (m *Cut) String() string            { return proto.CompactTextString(m) }
func (*Cut) ProtoMessage()               {}
func (*Cut) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Cut) GetCut() []int32 {
	if m != nil {
		return m.Cut
	}
	return nil
}

type ShardView struct {
	// Key: ReplicaID, Value: replica cuts
	Replicas map[int32]*Cut `protobuf:"bytes,1,rep,name=replicas" json:"replicas,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *ShardView) Reset()                    { *m = ShardView{} }
func (m *ShardView) String() string            { return proto.CompactTextString(m) }
func (*ShardView) ProtoMessage()               {}
func (*ShardView) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ShardView) GetReplicas() map[int32]*Cut {
	if m != nil {
		return m.Replicas
	}
	return nil
}

type ReportRequest struct {
	// Key: ShardID, Value: Shard cuts
	Shards map[int32]*ShardView `protobuf:"bytes,1,rep,name=shards" json:"shards,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *ReportRequest) Reset()                    { *m = ReportRequest{} }
func (m *ReportRequest) String() string            { return proto.CompactTextString(m) }
func (*ReportRequest) ProtoMessage()               {}
func (*ReportRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ReportRequest) GetShards() map[int32]*ShardView {
	if m != nil {
		return m.Shards
	}
	return nil
}

type ReportResponse struct {
	// Key: ShardID, Value: Commited cuts
	CommitedCuts map[int32]*Cut `protobuf:"bytes,1,rep,name=commitedCuts" json:"commitedCuts,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// GSN to start computing at
	StartGSN         int32   `protobuf:"varint,2,opt,name=startGSN" json:"startGSN,omitempty"`
	ViewID           int32   `protobuf:"varint,3,opt,name=viewID" json:"viewID,omitempty"`
	FinalizeShardIDs []int32 `protobuf:"varint,4,rep,packed,name=finalizeShardIDs" json:"finalizeShardIDs,omitempty"`
}

func (m *ReportResponse) Reset()                    { *m = ReportResponse{} }
func (m *ReportResponse) String() string            { return proto.CompactTextString(m) }
func (*ReportResponse) ProtoMessage()               {}
func (*ReportResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ReportResponse) GetCommitedCuts() map[int32]*Cut {
	if m != nil {
		return m.CommitedCuts
	}
	return nil
}

func (m *ReportResponse) GetStartGSN() int32 {
	if m != nil {
		return m.StartGSN
	}
	return 0
}

func (m *ReportResponse) GetViewID() int32 {
	if m != nil {
		return m.ViewID
	}
	return 0
}

func (m *ReportResponse) GetFinalizeShardIDs() []int32 {
	if m != nil {
		return m.FinalizeShardIDs
	}
	return nil
}

type RegisterRequest struct {
	ShardID   int32 `protobuf:"varint,1,opt,name=shardID" json:"shardID,omitempty"`
	ReplicaID int32 `protobuf:"varint,2,opt,name=replicaID" json:"replicaID,omitempty"`
}

func (m *RegisterRequest) Reset()                    { *m = RegisterRequest{} }
func (m *RegisterRequest) String() string            { return proto.CompactTextString(m) }
func (*RegisterRequest) ProtoMessage()               {}
func (*RegisterRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *RegisterRequest) GetShardID() int32 {
	if m != nil {
		return m.ShardID
	}
	return 0
}

func (m *RegisterRequest) GetReplicaID() int32 {
	if m != nil {
		return m.ReplicaID
	}
	return 0
}

type RegisterResponse struct {
	ViewID int32 `protobuf:"varint,1,opt,name=viewID" json:"viewID,omitempty"`
}

func (m *RegisterResponse) Reset()                    { *m = RegisterResponse{} }
func (m *RegisterResponse) String() string            { return proto.CompactTextString(m) }
func (*RegisterResponse) ProtoMessage()               {}
func (*RegisterResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *RegisterResponse) GetViewID() int32 {
	if m != nil {
		return m.ViewID
	}
	return 0
}

type FinalizeRequest struct {
	ShardIDs []int32 `protobuf:"varint,1,rep,packed,name=shardIDs" json:"shardIDs,omitempty"`
	Limit    int32   `protobuf:"varint,2,opt,name=limit" json:"limit,omitempty"`
}

func (m *FinalizeRequest) Reset()                    { *m = FinalizeRequest{} }
func (m *FinalizeRequest) String() string            { return proto.CompactTextString(m) }
func (*FinalizeRequest) ProtoMessage()               {}
func (*FinalizeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *FinalizeRequest) GetShardIDs() []int32 {
	if m != nil {
		return m.ShardIDs
	}
	return nil
}

func (m *FinalizeRequest) GetLimit() int32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type FinalizeResponse struct {
}

func (m *FinalizeResponse) Reset()                    { *m = FinalizeResponse{} }
func (m *FinalizeResponse) String() string            { return proto.CompactTextString(m) }
func (*FinalizeResponse) ProtoMessage()               {}
func (*FinalizeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func init() {
	proto.RegisterType((*Cut)(nil), "orderpb.Cut")
	proto.RegisterType((*ShardView)(nil), "orderpb.ShardView")
	proto.RegisterType((*ReportRequest)(nil), "orderpb.ReportRequest")
	proto.RegisterType((*ReportResponse)(nil), "orderpb.ReportResponse")
	proto.RegisterType((*RegisterRequest)(nil), "orderpb.RegisterRequest")
	proto.RegisterType((*RegisterResponse)(nil), "orderpb.RegisterResponse")
	proto.RegisterType((*FinalizeRequest)(nil), "orderpb.FinalizeRequest")
	proto.RegisterType((*FinalizeResponse)(nil), "orderpb.FinalizeResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Order service

type OrderClient interface {
	Report(ctx context.Context, opts ...grpc.CallOption) (Order_ReportClient, error)
	Forward(ctx context.Context, opts ...grpc.CallOption) (Order_ForwardClient, error)
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	Finalize(ctx context.Context, in *FinalizeRequest, opts ...grpc.CallOption) (*FinalizeResponse, error)
}

type orderClient struct {
	cc *grpc.ClientConn
}

func NewOrderClient(cc *grpc.ClientConn) OrderClient {
	return &orderClient{cc}
}

func (c *orderClient) Report(ctx context.Context, opts ...grpc.CallOption) (Order_ReportClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Order_serviceDesc.Streams[0], c.cc, "/orderpb.Order/Report", opts...)
	if err != nil {
		return nil, err
	}
	x := &orderReportClient{stream}
	return x, nil
}

type Order_ReportClient interface {
	Send(*ReportRequest) error
	Recv() (*ReportResponse, error)
	grpc.ClientStream
}

type orderReportClient struct {
	grpc.ClientStream
}

func (x *orderReportClient) Send(m *ReportRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *orderReportClient) Recv() (*ReportResponse, error) {
	m := new(ReportResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *orderClient) Forward(ctx context.Context, opts ...grpc.CallOption) (Order_ForwardClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Order_serviceDesc.Streams[1], c.cc, "/orderpb.Order/Forward", opts...)
	if err != nil {
		return nil, err
	}
	x := &orderForwardClient{stream}
	return x, nil
}

type Order_ForwardClient interface {
	Send(*ReportRequest) error
	CloseAndRecv() (*ReportResponse, error)
	grpc.ClientStream
}

type orderForwardClient struct {
	grpc.ClientStream
}

func (x *orderForwardClient) Send(m *ReportRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *orderForwardClient) CloseAndRecv() (*ReportResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ReportResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *orderClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := grpc.Invoke(ctx, "/orderpb.Order/Register", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderClient) Finalize(ctx context.Context, in *FinalizeRequest, opts ...grpc.CallOption) (*FinalizeResponse, error) {
	out := new(FinalizeResponse)
	err := grpc.Invoke(ctx, "/orderpb.Order/Finalize", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Order service

type OrderServer interface {
	Report(Order_ReportServer) error
	Forward(Order_ForwardServer) error
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	Finalize(context.Context, *FinalizeRequest) (*FinalizeResponse, error)
}

func RegisterOrderServer(s *grpc.Server, srv OrderServer) {
	s.RegisterService(&_Order_serviceDesc, srv)
}

func _Order_Report_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(OrderServer).Report(&orderReportServer{stream})
}

type Order_ReportServer interface {
	Send(*ReportResponse) error
	Recv() (*ReportRequest, error)
	grpc.ServerStream
}

type orderReportServer struct {
	grpc.ServerStream
}

func (x *orderReportServer) Send(m *ReportResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *orderReportServer) Recv() (*ReportRequest, error) {
	m := new(ReportRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Order_Forward_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(OrderServer).Forward(&orderForwardServer{stream})
}

type Order_ForwardServer interface {
	SendAndClose(*ReportResponse) error
	Recv() (*ReportRequest, error)
	grpc.ServerStream
}

type orderForwardServer struct {
	grpc.ServerStream
}

func (x *orderForwardServer) SendAndClose(m *ReportResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *orderForwardServer) Recv() (*ReportRequest, error) {
	m := new(ReportRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Order_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/orderpb.Order/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Order_Finalize_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FinalizeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServer).Finalize(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/orderpb.Order/Finalize",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServer).Finalize(ctx, req.(*FinalizeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Order_serviceDesc = grpc.ServiceDesc{
	ServiceName: "orderpb.Order",
	HandlerType: (*OrderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _Order_Register_Handler,
		},
		{
			MethodName: "Finalize",
			Handler:    _Order_Finalize_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Report",
			Handler:       _Order_Report_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Forward",
			Handler:       _Order_Forward_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "orderpb/order.proto",
}

func init() { proto.RegisterFile("orderpb/order.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 473 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xa4, 0x54, 0x4b, 0x6f, 0xd3, 0x40,
	0x10, 0xee, 0x26, 0x38, 0x49, 0x27, 0x85, 0x86, 0x01, 0xb5, 0xc6, 0xe2, 0x10, 0xed, 0x29, 0xf4,
	0x60, 0x50, 0xb8, 0xa0, 0x0a, 0x81, 0x90, 0x43, 0x91, 0x0f, 0x01, 0xc9, 0x95, 0xb8, 0xbb, 0xc9,
	0x02, 0xab, 0x26, 0x71, 0xd8, 0x5d, 0x37, 0x0a, 0x3f, 0xa2, 0x27, 0x4e, 0xfc, 0x5a, 0xe4, 0x7d,
	0x38, 0x76, 0x1d, 0x2e, 0xf4, 0x94, 0xcc, 0xeb, 0x7b, 0xcc, 0xac, 0x0c, 0x4f, 0x32, 0x31, 0x67,
	0x62, 0x7d, 0xf5, 0x52, 0xff, 0x86, 0x6b, 0x91, 0xa9, 0x0c, 0xbb, 0x36, 0x49, 0x4f, 0xa1, 0x1d,
	0xe5, 0x0a, 0x07, 0xd0, 0x9e, 0xe5, 0xca, 0x27, 0xc3, 0xf6, 0xc8, 0x4b, 0x8a, 0xbf, 0xf4, 0x37,
	0x81, 0xc3, 0xcb, 0x1f, 0xa9, 0x98, 0x7f, 0xe5, 0x6c, 0x83, 0x6f, 0xa1, 0x27, 0xd8, 0x7a, 0xc1,
	0x67, 0xa9, 0xd4, 0x4d, 0xfd, 0xf1, 0x30, 0xb4, 0x10, 0x61, 0xd9, 0x15, 0x26, 0xb6, 0xe5, 0xe3,
	0x4a, 0x89, 0x6d, 0x52, 0x4e, 0x04, 0x31, 0x3c, 0xac, 0x95, 0x0a, 0xba, 0x6b, 0xb6, 0xf5, 0xc9,
	0x90, 0x14, 0x74, 0xd7, 0x6c, 0x8b, 0x14, 0xbc, 0x9b, 0x74, 0x91, 0x33, 0xbf, 0x35, 0x24, 0xa3,
	0xfe, 0xf8, 0xa8, 0x44, 0x8f, 0x72, 0x95, 0x98, 0xd2, 0x79, 0xeb, 0x0d, 0xa1, 0x7f, 0x88, 0xc6,
	0xca, 0x84, 0x4a, 0xd8, 0xcf, 0x9c, 0x49, 0x85, 0xe7, 0xd0, 0x91, 0x85, 0x02, 0x27, 0x8c, 0x96,
	0xa3, 0xb5, 0x3e, 0x23, 0xd3, 0x4a, 0xb3, 0x13, 0xc1, 0x14, 0xfa, 0x95, 0xf4, 0x1e, 0x59, 0xa3,
	0xba, 0x2c, 0x6c, 0x9a, 0xae, 0x8a, 0xbb, 0x6d, 0xc1, 0x23, 0x47, 0x2a, 0xd7, 0xd9, 0x4a, 0x32,
	0x9c, 0xc2, 0xd1, 0x2c, 0x5b, 0x2e, 0xb9, 0x62, 0xf3, 0x28, 0x57, 0x4e, 0xe3, 0x8b, 0x86, 0x46,
	0xd3, 0x1e, 0x46, 0x95, 0x5e, 0x23, 0xb5, 0x36, 0x8e, 0x01, 0xf4, 0xa4, 0x4a, 0x85, 0xfa, 0x74,
	0xf9, 0x59, 0x4b, 0xf2, 0x92, 0x32, 0xc6, 0x13, 0xe8, 0xdc, 0x70, 0xb6, 0x89, 0x27, 0x7e, 0x5b,
	0x57, 0x6c, 0x84, 0x67, 0x30, 0xf8, 0xc6, 0x57, 0xe9, 0x82, 0xff, 0x62, 0x5a, 0x75, 0x3c, 0x91,
	0xfe, 0x03, 0x7d, 0xe8, 0x46, 0x3e, 0x98, 0xc2, 0xe3, 0x86, 0x84, 0x7b, 0x5c, 0x2b, 0x86, 0xe3,
	0x84, 0x7d, 0xe7, 0x52, 0x31, 0xe1, 0xce, 0xe5, 0x43, 0x57, 0x1a, 0x36, 0x0b, 0xe8, 0x42, 0x7c,
	0x0e, 0x87, 0xf6, 0xc5, 0xc4, 0x13, 0x6b, 0x6e, 0x97, 0xa0, 0x67, 0x30, 0xd8, 0x41, 0xd9, 0xe5,
	0xee, 0x1c, 0x93, 0xaa, 0x63, 0x1a, 0xc1, 0xf1, 0x85, 0x75, 0xe6, 0x68, 0x8b, 0xc5, 0x39, 0xf3,
	0xe6, 0x95, 0x97, 0x31, 0x3e, 0x05, 0x6f, 0xc1, 0x97, 0x5c, 0x59, 0x52, 0x13, 0x50, 0x84, 0xc1,
	0x0e, 0xc4, 0x10, 0x8e, 0x6f, 0x5b, 0xe0, 0x7d, 0x29, 0xac, 0xe2, 0x7b, 0xe8, 0x98, 0xd3, 0xe1,
	0xc9, 0xfe, 0xf7, 0x16, 0x9c, 0xfe, 0xe3, 0xc6, 0xf4, 0x60, 0x44, 0x5e, 0x11, 0x7c, 0x07, 0xdd,
	0x8b, 0x4c, 0x6c, 0x52, 0x31, 0xff, 0x2f, 0x04, 0xfc, 0x00, 0x3d, 0xb7, 0x0f, 0xf4, 0x2b, 0x8d,
	0xb5, 0x6d, 0x07, 0xcf, 0xf6, 0x54, 0x1c, 0x48, 0x01, 0xe1, 0x1c, 0x56, 0x20, 0xee, 0x6c, 0xae,
	0x02, 0x71, 0x77, 0x1d, 0xf4, 0xe0, 0xaa, 0xa3, 0x3f, 0x27, 0xaf, 0xff, 0x06, 0x00, 0x00, 0xff,
	0xff, 0xb1, 0xe0, 0x43, 0xa6, 0x65, 0x04, 0x00, 0x00,
}
