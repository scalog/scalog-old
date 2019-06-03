// Code generated by protoc-gen-go.
// source: rpc/discovery.proto
// DO NOT EDIT!

/*
Package rpc is a generated protocol buffer package.

It is generated from these files:
	rpc/discovery.proto

It has these top-level messages:
	DiscoverRequest
	DiscoverResponse
	Shard
	DataServer
*/
package rpc

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

type DiscoverRequest struct {
}

func (m *DiscoverRequest) Reset()                    { *m = DiscoverRequest{} }
func (m *DiscoverRequest) String() string            { return proto.CompactTextString(m) }
func (*DiscoverRequest) ProtoMessage()               {}
func (*DiscoverRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type DiscoverResponse struct {
	Shards []*Shard `protobuf:"bytes,1,rep,name=shards" json:"shards,omitempty"`
}

func (m *DiscoverResponse) Reset()                    { *m = DiscoverResponse{} }
func (m *DiscoverResponse) String() string            { return proto.CompactTextString(m) }
func (*DiscoverResponse) ProtoMessage()               {}
func (*DiscoverResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *DiscoverResponse) GetShards() []*Shard {
	if m != nil {
		return m.Shards
	}
	return nil
}

type Shard struct {
	ShardID int32         `protobuf:"varint,1,opt,name=shardID" json:"shardID,omitempty"`
	Servers []*DataServer `protobuf:"bytes,2,rep,name=servers" json:"servers,omitempty"`
}

func (m *Shard) Reset()                    { *m = Shard{} }
func (m *Shard) String() string            { return proto.CompactTextString(m) }
func (*Shard) ProtoMessage()               {}
func (*Shard) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Shard) GetShardID() int32 {
	if m != nil {
		return m.ShardID
	}
	return 0
}

func (m *Shard) GetServers() []*DataServer {
	if m != nil {
		return m.Servers
	}
	return nil
}

type DataServer struct {
	ServerID int32  `protobuf:"varint,1,opt,name=serverID" json:"serverID,omitempty"`
	Port     int32  `protobuf:"varint,2,opt,name=port" json:"port,omitempty"`
	Ip       string `protobuf:"bytes,3,opt,name=ip" json:"ip,omitempty"`
}

func (m *DataServer) Reset()                    { *m = DataServer{} }
func (m *DataServer) String() string            { return proto.CompactTextString(m) }
func (*DataServer) ProtoMessage()               {}
func (*DataServer) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *DataServer) GetServerID() int32 {
	if m != nil {
		return m.ServerID
	}
	return 0
}

func (m *DataServer) GetPort() int32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *DataServer) GetIp() string {
	if m != nil {
		return m.Ip
	}
	return ""
}

func init() {
	proto.RegisterType((*DiscoverRequest)(nil), "rpc.DiscoverRequest")
	proto.RegisterType((*DiscoverResponse)(nil), "rpc.DiscoverResponse")
	proto.RegisterType((*Shard)(nil), "rpc.Shard")
	proto.RegisterType((*DataServer)(nil), "rpc.DataServer")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Discovery service

type DiscoveryClient interface {
	DiscoverServers(ctx context.Context, in *DiscoverRequest, opts ...grpc.CallOption) (*DiscoverResponse, error)
}

type discoveryClient struct {
	cc *grpc.ClientConn
}

func NewDiscoveryClient(cc *grpc.ClientConn) DiscoveryClient {
	return &discoveryClient{cc}
}

func (c *discoveryClient) DiscoverServers(ctx context.Context, in *DiscoverRequest, opts ...grpc.CallOption) (*DiscoverResponse, error) {
	out := new(DiscoverResponse)
	err := grpc.Invoke(ctx, "/rpc.Discovery/DiscoverServers", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Discovery service

type DiscoveryServer interface {
	DiscoverServers(context.Context, *DiscoverRequest) (*DiscoverResponse, error)
}

func RegisterDiscoveryServer(s *grpc.Server, srv DiscoveryServer) {
	s.RegisterService(&_Discovery_serviceDesc, srv)
}

func _Discovery_DiscoverServers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DiscoverRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiscoveryServer).DiscoverServers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Discovery/DiscoverServers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiscoveryServer).DiscoverServers(ctx, req.(*DiscoverRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Discovery_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.Discovery",
	HandlerType: (*DiscoveryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DiscoverServers",
			Handler:    _Discovery_DiscoverServers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/discovery.proto",
}

func init() { proto.RegisterFile("rpc/discovery.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 228 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x5c, 0xd0, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x06, 0x60, 0x93, 0xd8, 0xd6, 0x8e, 0x60, 0x75, 0x54, 0x58, 0x7a, 0x0a, 0x7b, 0x8a, 0x97,
	0x08, 0x15, 0x3c, 0x7b, 0xc8, 0x45, 0xa8, 0x97, 0xed, 0x13, 0xc4, 0x74, 0xc1, 0x5c, 0xdc, 0x71,
	0x66, 0x2d, 0xf4, 0xed, 0xc5, 0xd9, 0x26, 0x41, 0x6f, 0x3b, 0xdf, 0xc0, 0xcf, 0xce, 0x0f, 0xb7,
	0x4c, 0xdd, 0xe3, 0xbe, 0x97, 0x2e, 0x1c, 0x3c, 0x1f, 0x6b, 0xe2, 0x10, 0x03, 0x16, 0x4c, 0x9d,
	0xbd, 0x81, 0x55, 0x73, 0x72, 0xe7, 0xbf, 0xbe, 0xbd, 0x44, 0xfb, 0x0c, 0xd7, 0x13, 0x09, 0x85,
	0x4f, 0xf1, 0x68, 0x61, 0x2e, 0x1f, 0x2d, 0xef, 0xc5, 0x64, 0x65, 0x51, 0x5d, 0x6e, 0xa0, 0x66,
	0xea, 0xea, 0xdd, 0x2f, 0xb9, 0xd3, 0xc6, 0x6e, 0x61, 0xa6, 0x80, 0x06, 0x16, 0x4a, 0xaf, 0x8d,
	0xc9, 0xca, 0xac, 0x9a, 0xb9, 0x61, 0xc4, 0x07, 0x58, 0x88, 0xe7, 0x83, 0x67, 0x31, 0xb9, 0xe6,
	0xac, 0x34, 0xa7, 0x69, 0x63, 0xbb, 0x53, 0x77, 0xc3, 0xde, 0x6e, 0x01, 0x26, 0xc6, 0x35, 0x5c,
	0xa4, 0xc5, 0x98, 0x39, 0xce, 0x88, 0x70, 0x4e, 0x81, 0xa3, 0xc9, 0xd5, 0xf5, 0x8d, 0x57, 0x90,
	0xf7, 0x64, 0x8a, 0x32, 0xab, 0x96, 0x2e, 0xef, 0x69, 0xf3, 0x06, 0xcb, 0xe1, 0xa6, 0x23, 0xbe,
	0x4c, 0x37, 0xa7, 0x78, 0xc1, 0xbb, 0xf4, 0x8f, 0xbf, 0x4d, 0xac, 0xef, 0xff, 0x69, 0x2a, 0xc3,
	0x9e, 0xbd, 0xcf, 0xb5, 0xc1, 0xa7, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xbd, 0xfb, 0x18, 0xcd,
	0x58, 0x01, 0x00, 0x00,
}
