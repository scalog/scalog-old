// Code generated by protoc-gen-go. DO NOT EDIT.
// source: messaging/order.proto

package messaging

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Cut struct {
	// Array of len numReplicas
	Cut                  []int32  `protobuf:"varint,1,rep,packed,name=cut,proto3" json:"cut,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Cut) Reset()         { *m = Cut{} }
func (m *Cut) String() string { return proto.CompactTextString(m) }
func (*Cut) ProtoMessage()    {}
func (*Cut) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{0}
}

func (m *Cut) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Cut.Unmarshal(m, b)
}
func (m *Cut) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Cut.Marshal(b, m, deterministic)
}
func (m *Cut) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Cut.Merge(m, src)
}
func (m *Cut) XXX_Size() int {
	return xxx_messageInfo_Cut.Size(m)
}
func (m *Cut) XXX_DiscardUnknown() {
	xxx_messageInfo_Cut.DiscardUnknown(m)
}

var xxx_messageInfo_Cut proto.InternalMessageInfo

func (m *Cut) GetCut() []int32 {
	if m != nil {
		return m.Cut
	}
	return nil
}

type ShardView struct {
	// Key: ReplicaID, Value: replica cuts
	Replicas             map[int32]*Cut `protobuf:"bytes,1,rep,name=replicas,proto3" json:"replicas,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *ShardView) Reset()         { *m = ShardView{} }
func (m *ShardView) String() string { return proto.CompactTextString(m) }
func (*ShardView) ProtoMessage()    {}
func (*ShardView) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{1}
}

func (m *ShardView) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ShardView.Unmarshal(m, b)
}
func (m *ShardView) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ShardView.Marshal(b, m, deterministic)
}
func (m *ShardView) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ShardView.Merge(m, src)
}
func (m *ShardView) XXX_Size() int {
	return xxx_messageInfo_ShardView.Size(m)
}
func (m *ShardView) XXX_DiscardUnknown() {
	xxx_messageInfo_ShardView.DiscardUnknown(m)
}

var xxx_messageInfo_ShardView proto.InternalMessageInfo

func (m *ShardView) GetReplicas() map[int32]*Cut {
	if m != nil {
		return m.Replicas
	}
	return nil
}

type ReportRequest struct {
	// Key: ShardID, Value: Shard cuts
	Shards map[int32]*ShardView `protobuf:"bytes,1,rep,name=shards,proto3" json:"shards,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Non-nil only if the ordering layer wishes to propose a finalization these particular shards after k cuts
	Finalize map[int32]int32 `protobuf:"bytes,2,rep,name=finalize,proto3" json:"finalize,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	// True if this is a request to batch and send the committedcuts
	Batch                bool     `protobuf:"varint,3,opt,name=batch,proto3" json:"batch,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReportRequest) Reset()         { *m = ReportRequest{} }
func (m *ReportRequest) String() string { return proto.CompactTextString(m) }
func (*ReportRequest) ProtoMessage()    {}
func (*ReportRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{2}
}

func (m *ReportRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReportRequest.Unmarshal(m, b)
}
func (m *ReportRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReportRequest.Marshal(b, m, deterministic)
}
func (m *ReportRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReportRequest.Merge(m, src)
}
func (m *ReportRequest) XXX_Size() int {
	return xxx_messageInfo_ReportRequest.Size(m)
}
func (m *ReportRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ReportRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ReportRequest proto.InternalMessageInfo

func (m *ReportRequest) GetShards() map[int32]*ShardView {
	if m != nil {
		return m.Shards
	}
	return nil
}

func (m *ReportRequest) GetFinalize() map[int32]int32 {
	if m != nil {
		return m.Finalize
	}
	return nil
}

func (m *ReportRequest) GetBatch() bool {
	if m != nil {
		return m.Batch
	}
	return false
}

type ReportResponse struct {
	// Key: ShardID, Value: Commited cuts
	CommitedCuts map[int32]*Cut `protobuf:"bytes,1,rep,name=commitedCuts,proto3" json:"commitedCuts,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// GSN to start computing at
	StartGSN int32 `protobuf:"varint,2,opt,name=startGSN,proto3" json:"startGSN,omitempty"`
	// Array of shards that have been finalized in this batch
	Finalize             []int32  `protobuf:"varint,3,rep,packed,name=finalize,proto3" json:"finalize,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReportResponse) Reset()         { *m = ReportResponse{} }
func (m *ReportResponse) String() string { return proto.CompactTextString(m) }
func (*ReportResponse) ProtoMessage()    {}
func (*ReportResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{3}
}

func (m *ReportResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReportResponse.Unmarshal(m, b)
}
func (m *ReportResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReportResponse.Marshal(b, m, deterministic)
}
func (m *ReportResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReportResponse.Merge(m, src)
}
func (m *ReportResponse) XXX_Size() int {
	return xxx_messageInfo_ReportResponse.Size(m)
}
func (m *ReportResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ReportResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ReportResponse proto.InternalMessageInfo

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

func (m *ReportResponse) GetFinalize() []int32 {
	if m != nil {
		return m.Finalize
	}
	return nil
}

type RegisterRequest struct {
	ShardID              int32    `protobuf:"varint,1,opt,name=shardID,proto3" json:"shardID,omitempty"`
	ReplicaID            int32    `protobuf:"varint,2,opt,name=replicaID,proto3" json:"replicaID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegisterRequest) Reset()         { *m = RegisterRequest{} }
func (m *RegisterRequest) String() string { return proto.CompactTextString(m) }
func (*RegisterRequest) ProtoMessage()    {}
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{4}
}

func (m *RegisterRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegisterRequest.Unmarshal(m, b)
}
func (m *RegisterRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegisterRequest.Marshal(b, m, deterministic)
}
func (m *RegisterRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegisterRequest.Merge(m, src)
}
func (m *RegisterRequest) XXX_Size() int {
	return xxx_messageInfo_RegisterRequest.Size(m)
}
func (m *RegisterRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RegisterRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RegisterRequest proto.InternalMessageInfo

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
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegisterResponse) Reset()         { *m = RegisterResponse{} }
func (m *RegisterResponse) String() string { return proto.CompactTextString(m) }
func (*RegisterResponse) ProtoMessage()    {}
func (*RegisterResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{5}
}

func (m *RegisterResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegisterResponse.Unmarshal(m, b)
}
func (m *RegisterResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegisterResponse.Marshal(b, m, deterministic)
}
func (m *RegisterResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegisterResponse.Merge(m, src)
}
func (m *RegisterResponse) XXX_Size() int {
	return xxx_messageInfo_RegisterResponse.Size(m)
}
func (m *RegisterResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RegisterResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RegisterResponse proto.InternalMessageInfo

type FinalizeRequest struct {
	// Array of shards to finalize after k cuts
	Shards               map[int32]int32 `protobuf:"bytes,1,rep,name=shards,proto3" json:"shards,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *FinalizeRequest) Reset()         { *m = FinalizeRequest{} }
func (m *FinalizeRequest) String() string { return proto.CompactTextString(m) }
func (*FinalizeRequest) ProtoMessage()    {}
func (*FinalizeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{6}
}

func (m *FinalizeRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FinalizeRequest.Unmarshal(m, b)
}
func (m *FinalizeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FinalizeRequest.Marshal(b, m, deterministic)
}
func (m *FinalizeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FinalizeRequest.Merge(m, src)
}
func (m *FinalizeRequest) XXX_Size() int {
	return xxx_messageInfo_FinalizeRequest.Size(m)
}
func (m *FinalizeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_FinalizeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_FinalizeRequest proto.InternalMessageInfo

func (m *FinalizeRequest) GetShards() map[int32]int32 {
	if m != nil {
		return m.Shards
	}
	return nil
}

type FinalizeResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FinalizeResponse) Reset()         { *m = FinalizeResponse{} }
func (m *FinalizeResponse) String() string { return proto.CompactTextString(m) }
func (*FinalizeResponse) ProtoMessage()    {}
func (*FinalizeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{7}
}

func (m *FinalizeResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FinalizeResponse.Unmarshal(m, b)
}
func (m *FinalizeResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FinalizeResponse.Marshal(b, m, deterministic)
}
func (m *FinalizeResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FinalizeResponse.Merge(m, src)
}
func (m *FinalizeResponse) XXX_Size() int {
	return xxx_messageInfo_FinalizeResponse.Size(m)
}
func (m *FinalizeResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_FinalizeResponse.DiscardUnknown(m)
}

var xxx_messageInfo_FinalizeResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Cut)(nil), "messaging.Cut")
	proto.RegisterType((*ShardView)(nil), "messaging.ShardView")
	proto.RegisterMapType((map[int32]*Cut)(nil), "messaging.ShardView.ReplicasEntry")
	proto.RegisterType((*ReportRequest)(nil), "messaging.ReportRequest")
	proto.RegisterMapType((map[int32]int32)(nil), "messaging.ReportRequest.FinalizeEntry")
	proto.RegisterMapType((map[int32]*ShardView)(nil), "messaging.ReportRequest.ShardsEntry")
	proto.RegisterType((*ReportResponse)(nil), "messaging.ReportResponse")
	proto.RegisterMapType((map[int32]*Cut)(nil), "messaging.ReportResponse.CommitedCutsEntry")
	proto.RegisterType((*RegisterRequest)(nil), "messaging.RegisterRequest")
	proto.RegisterType((*RegisterResponse)(nil), "messaging.RegisterResponse")
	proto.RegisterType((*FinalizeRequest)(nil), "messaging.FinalizeRequest")
	proto.RegisterMapType((map[int32]int32)(nil), "messaging.FinalizeRequest.ShardsEntry")
	proto.RegisterType((*FinalizeResponse)(nil), "messaging.FinalizeResponse")
}

func init() { proto.RegisterFile("messaging/order.proto", fileDescriptor_f8af3e412e3326d2) }

var fileDescriptor_f8af3e412e3326d2 = []byte{
	// 496 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0xdd, 0x6e, 0xd3, 0x30,
	0x14, 0x9e, 0x1b, 0xa5, 0xa4, 0xa7, 0x6c, 0x2b, 0x47, 0x45, 0x04, 0xc3, 0x45, 0x15, 0x4d, 0xa8,
	0x02, 0x29, 0xa0, 0x72, 0xc3, 0x9f, 0x76, 0x41, 0x36, 0x50, 0x85, 0x44, 0xa5, 0x4c, 0xe2, 0x3e,
	0x6b, 0x4d, 0x17, 0xad, 0x6d, 0x8a, 0xed, 0x80, 0xca, 0x33, 0xf0, 0x00, 0xbc, 0x08, 0x2f, 0xc3,
	0x23, 0xf0, 0x14, 0xc8, 0x89, 0x9d, 0xc6, 0x25, 0x15, 0x17, 0xbb, 0xab, 0xed, 0xef, 0x7c, 0xe7,
	0x3b, 0xdf, 0xf9, 0x1a, 0xb8, 0xbb, 0x64, 0x42, 0x24, 0xf3, 0x74, 0x35, 0x7f, 0x9a, 0xf1, 0x19,
	0xe3, 0xe1, 0x9a, 0x67, 0x32, 0xc3, 0x4e, 0x75, 0x1d, 0xdc, 0x03, 0x27, 0xca, 0x25, 0xf6, 0xc0,
	0x99, 0xe6, 0xd2, 0x27, 0x03, 0x67, 0xe8, 0xc6, 0xea, 0x67, 0xf0, 0x93, 0x40, 0xe7, 0xe2, 0x2a,
	0xe1, 0xb3, 0x4f, 0x29, 0xfb, 0x86, 0xa7, 0xe0, 0x71, 0xb6, 0x5e, 0xa4, 0xd3, 0x44, 0x14, 0xa0,
	0xee, 0x28, 0x08, 0x2b, 0x92, 0xb0, 0xc2, 0x85, 0xb1, 0x06, 0x9d, 0xaf, 0x24, 0xdf, 0xc4, 0x55,
	0x0d, 0xfd, 0x00, 0x87, 0xd6, 0x93, 0x6a, 0x78, 0xcd, 0x36, 0x3e, 0x19, 0x10, 0xd5, 0xf0, 0x9a,
	0x6d, 0xf0, 0x04, 0xdc, 0xaf, 0xc9, 0x22, 0x67, 0x7e, 0x6b, 0x40, 0x86, 0xdd, 0xd1, 0x51, 0x8d,
	0x3f, 0xca, 0x65, 0x5c, 0x3e, 0xbe, 0x6a, 0xbd, 0x20, 0xc1, 0xaf, 0x56, 0xc1, 0x96, 0x71, 0x19,
	0xb3, 0x2f, 0x39, 0x13, 0x12, 0xdf, 0x40, 0x5b, 0x28, 0x0d, 0x46, 0xdc, 0x49, 0xad, 0xd8, 0x42,
	0x96, 0x52, 0xb5, 0x3c, 0x5d, 0x83, 0x6f, 0xc1, 0xfb, 0x9c, 0xae, 0x92, 0x45, 0xfa, 0x5d, 0x35,
	0x57, 0xf5, 0x8f, 0xf6, 0xd6, 0xbf, 0xd3, 0x40, 0x3d, 0xa0, 0xa9, 0xc3, 0x3e, 0xb8, 0x97, 0x89,
	0x9c, 0x5e, 0xf9, 0xce, 0x80, 0x0c, 0xbd, 0xb8, 0x3c, 0xd0, 0x09, 0x74, 0x6b, 0x0d, 0x1b, 0x86,
	0x7e, 0x6c, 0x0f, 0xdd, 0x6f, 0x32, 0xb5, 0x36, 0x3a, 0x7d, 0x0d, 0x87, 0x96, 0x82, 0x06, 0xca,
	0x7e, 0x9d, 0xd2, 0xad, 0xfb, 0xf6, 0x87, 0xc0, 0x91, 0x99, 0x46, 0xac, 0xb3, 0x95, 0x60, 0x38,
	0x81, 0xdb, 0xd3, 0x6c, 0xb9, 0x4c, 0x25, 0x9b, 0x45, 0xb9, 0x34, 0xf6, 0x3d, 0x69, 0x18, 0xbf,
	0x2c, 0x08, 0xa3, 0x1a, 0xba, 0xf4, 0xc0, 0x22, 0x40, 0x0a, 0x9e, 0x90, 0x09, 0x97, 0xef, 0x2f,
	0x3e, 0x6a, 0x01, 0xd5, 0x59, 0xbd, 0x55, 0x3e, 0x3b, 0x45, 0xd2, 0xaa, 0x33, 0x9d, 0xc0, 0x9d,
	0x7f, 0xa8, 0x6f, 0x14, 0x92, 0x31, 0x1c, 0xc7, 0x6c, 0x9e, 0x0a, 0xc9, 0xb8, 0x49, 0x89, 0x0f,
	0xb7, 0x8a, 0x8d, 0x8f, 0xcf, 0x34, 0xa5, 0x39, 0xe2, 0x43, 0xe8, 0xe8, 0xa8, 0x8e, 0xcf, 0xb4,
	0xec, 0xed, 0x45, 0x80, 0xd0, 0xdb, 0x52, 0x95, 0x3e, 0x04, 0x3f, 0x08, 0x1c, 0x9b, 0x4d, 0x18,
	0xfe, 0xd3, 0x9d, 0x14, 0xd6, 0x53, 0xb4, 0x83, 0x6d, 0xca, 0x21, 0x7d, 0xf9, 0xbf, 0xb4, 0xec,
	0x5f, 0x2d, 0x42, 0x6f, 0xdb, 0xa1, 0x94, 0x38, 0xfa, 0x4d, 0xc0, 0x9d, 0xa8, 0x7f, 0x3d, 0x46,
	0xd0, 0x2e, 0xd7, 0x88, 0xfe, 0xbe, 0x60, 0xd3, 0xfb, 0x7b, 0x77, 0x1e, 0x1c, 0x0c, 0xc9, 0x33,
	0x82, 0xe7, 0xe0, 0x19, 0x17, 0x90, 0x5a, 0x60, 0xcb, 0x65, 0xfa, 0xa0, 0xf1, 0xcd, 0x50, 0x29,
	0x1a, 0xa3, 0xd4, 0xa2, 0xd9, 0x31, 0xc8, 0xa2, 0xd9, 0x1d, 0x2d, 0x38, 0xb8, 0x6c, 0x17, 0x5f,
	0xb2, 0xe7, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x81, 0x38, 0x5b, 0x37, 0xe2, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// OrderClient is the client API for Order service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type OrderClient interface {
	Report(ctx context.Context, opts ...grpc.CallOption) (Order_ReportClient, error)
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
	stream, err := c.cc.NewStream(ctx, &_Order_serviceDesc.Streams[0], "/messaging.Order/Report", opts...)
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

func (c *orderClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, "/messaging.Order/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderClient) Finalize(ctx context.Context, in *FinalizeRequest, opts ...grpc.CallOption) (*FinalizeResponse, error) {
	out := new(FinalizeResponse)
	err := c.cc.Invoke(ctx, "/messaging.Order/Finalize", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrderServer is the server API for Order service.
type OrderServer interface {
	Report(Order_ReportServer) error
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	Finalize(context.Context, *FinalizeRequest) (*FinalizeResponse, error)
}

// UnimplementedOrderServer can be embedded to have forward compatible implementations.
type UnimplementedOrderServer struct {
}

func (*UnimplementedOrderServer) Report(srv Order_ReportServer) error {
	return status.Errorf(codes.Unimplemented, "method Report not implemented")
}
func (*UnimplementedOrderServer) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (*UnimplementedOrderServer) Finalize(ctx context.Context, req *FinalizeRequest) (*FinalizeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Finalize not implemented")
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
		FullMethod: "/messaging.Order/Register",
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
		FullMethod: "/messaging.Order/Finalize",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServer).Finalize(ctx, req.(*FinalizeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Order_serviceDesc = grpc.ServiceDesc{
	ServiceName: "messaging.Order",
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
	},
	Metadata: "messaging/order.proto",
}
