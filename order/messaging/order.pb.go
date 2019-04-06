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

type ReportRequest struct {
	ShardID      int32   `protobuf:"varint,1,opt,name=shardID,proto3" json:"shardID,omitempty"`
	ReplicaID    int32   `protobuf:"varint,2,opt,name=replicaID,proto3" json:"replicaID,omitempty"`
	TentativeCut []int32 `protobuf:"varint,3,rep,packed,name=tentativeCut,proto3" json:"tentativeCut,omitempty"`
	// True only if the ordering layer wishes to propose a finalization of this particular shard
	Finalized bool `protobuf:"varint,4,opt,name=finalized,proto3" json:"finalized,omitempty"`
	// Non-zero if data server wants to request missing cuts in range [minLogNum, maxLogNum] (inclusive)
	MinLogNum            int32    `protobuf:"varint,5,opt,name=minLogNum,proto3" json:"minLogNum,omitempty"`
	MaxLogNum            int32    `protobuf:"varint,6,opt,name=maxLogNum,proto3" json:"maxLogNum,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReportRequest) Reset()         { *m = ReportRequest{} }
func (m *ReportRequest) String() string { return proto.CompactTextString(m) }
func (*ReportRequest) ProtoMessage()    {}
func (*ReportRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{0}
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

func (m *ReportRequest) GetShardID() int32 {
	if m != nil {
		return m.ShardID
	}
	return 0
}

func (m *ReportRequest) GetReplicaID() int32 {
	if m != nil {
		return m.ReplicaID
	}
	return 0
}

func (m *ReportRequest) GetTentativeCut() []int32 {
	if m != nil {
		return m.TentativeCut
	}
	return nil
}

func (m *ReportRequest) GetFinalized() bool {
	if m != nil {
		return m.Finalized
	}
	return false
}

func (m *ReportRequest) GetMinLogNum() int32 {
	if m != nil {
		return m.MinLogNum
	}
	return 0
}

func (m *ReportRequest) GetMaxLogNum() int32 {
	if m != nil {
		return m.MaxLogNum
	}
	return 0
}

type IntList struct {
	List                 []int32  `protobuf:"varint,1,rep,packed,name=list,proto3" json:"list,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IntList) Reset()         { *m = IntList{} }
func (m *IntList) String() string { return proto.CompactTextString(m) }
func (*IntList) ProtoMessage()    {}
func (*IntList) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{1}
}

func (m *IntList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IntList.Unmarshal(m, b)
}
func (m *IntList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IntList.Marshal(b, m, deterministic)
}
func (m *IntList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IntList.Merge(m, src)
}
func (m *IntList) XXX_Size() int {
	return xxx_messageInfo_IntList.Size(m)
}
func (m *IntList) XXX_DiscardUnknown() {
	xxx_messageInfo_IntList.DiscardUnknown(m)
}

var xxx_messageInfo_IntList proto.InternalMessageInfo

func (m *IntList) GetList() []int32 {
	if m != nil {
		return m.List
	}
	return nil
}

type ForwardResponse struct {
	StartGlobalSequenceNums map[int32]int32    `protobuf:"bytes,1,rep,name=startGlobalSequenceNums,proto3" json:"startGlobalSequenceNums,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	CommittedCuts           map[int32]*IntList `protobuf:"bytes,2,rep,name=committedCuts,proto3" json:"committedCuts,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	StartGlobalSequenceNum  int32              `protobuf:"varint,3,opt,name=startGlobalSequenceNum,proto3" json:"startGlobalSequenceNum,omitempty"`
	XXX_NoUnkeyedLiteral    struct{}           `json:"-"`
	XXX_unrecognized        []byte             `json:"-"`
	XXX_sizecache           int32              `json:"-"`
}

func (m *ForwardResponse) Reset()         { *m = ForwardResponse{} }
func (m *ForwardResponse) String() string { return proto.CompactTextString(m) }
func (*ForwardResponse) ProtoMessage()    {}
func (*ForwardResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f8af3e412e3326d2, []int{2}
}

func (m *ForwardResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ForwardResponse.Unmarshal(m, b)
}
func (m *ForwardResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ForwardResponse.Marshal(b, m, deterministic)
}
func (m *ForwardResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ForwardResponse.Merge(m, src)
}
func (m *ForwardResponse) XXX_Size() int {
	return xxx_messageInfo_ForwardResponse.Size(m)
}
func (m *ForwardResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ForwardResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ForwardResponse proto.InternalMessageInfo

func (m *ForwardResponse) GetStartGlobalSequenceNums() map[int32]int32 {
	if m != nil {
		return m.StartGlobalSequenceNums
	}
	return nil
}

func (m *ForwardResponse) GetCommittedCuts() map[int32]*IntList {
	if m != nil {
		return m.CommittedCuts
	}
	return nil
}

func (m *ForwardResponse) GetStartGlobalSequenceNum() int32 {
	if m != nil {
		return m.StartGlobalSequenceNum
	}
	return 0
}

type ReportResponse struct {
	StartGlobalSequenceNum int32   `protobuf:"varint,1,opt,name=startGlobalSequenceNum,proto3" json:"startGlobalSequenceNum,omitempty"`
	CommittedCuts          []int32 `protobuf:"varint,2,rep,packed,name=committedCuts,proto3" json:"committedCuts,omitempty"`
	// range [minLogNum, maxLogNum] includes all newly applied entries & config changes (inclusive)
	MinLogNum int32 `protobuf:"varint,3,opt,name=minLogNum,proto3" json:"minLogNum,omitempty"`
	MaxLogNum int32 `protobuf:"varint,4,opt,name=maxLogNum,proto3" json:"maxLogNum,omitempty"`
	// True only if the ordering layer is finalizing this data replica
	Finalized            bool     `protobuf:"varint,5,opt,name=finalized,proto3" json:"finalized,omitempty"`
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

func (m *ReportResponse) GetStartGlobalSequenceNum() int32 {
	if m != nil {
		return m.StartGlobalSequenceNum
	}
	return 0
}

func (m *ReportResponse) GetCommittedCuts() []int32 {
	if m != nil {
		return m.CommittedCuts
	}
	return nil
}

func (m *ReportResponse) GetMinLogNum() int32 {
	if m != nil {
		return m.MinLogNum
	}
	return 0
}

func (m *ReportResponse) GetMaxLogNum() int32 {
	if m != nil {
		return m.MaxLogNum
	}
	return 0
}

func (m *ReportResponse) GetFinalized() bool {
	if m != nil {
		return m.Finalized
	}
	return false
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
	// Array of shards to finalize
	ShardIDs []int32 `protobuf:"varint,1,rep,packed,name=shardIDs,proto3" json:"shardIDs,omitempty"`
	// Finalize these shards after [finalizeAfterCuts] more cuts
	FinalizeAfterCuts    int32    `protobuf:"varint,2,opt,name=finalizeAfterCuts,proto3" json:"finalizeAfterCuts,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
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

func (m *FinalizeRequest) GetShardIDs() []int32 {
	if m != nil {
		return m.ShardIDs
	}
	return nil
}

func (m *FinalizeRequest) GetFinalizeAfterCuts() int32 {
	if m != nil {
		return m.FinalizeAfterCuts
	}
	return 0
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
	proto.RegisterType((*ReportRequest)(nil), "messaging.ReportRequest")
	proto.RegisterType((*IntList)(nil), "messaging.IntList")
	proto.RegisterType((*ForwardResponse)(nil), "messaging.ForwardResponse")
	proto.RegisterMapType((map[int32]*IntList)(nil), "messaging.ForwardResponse.CommittedCutsEntry")
	proto.RegisterMapType((map[int32]int32)(nil), "messaging.ForwardResponse.StartGlobalSequenceNumsEntry")
	proto.RegisterType((*ReportResponse)(nil), "messaging.ReportResponse")
	proto.RegisterType((*RegisterRequest)(nil), "messaging.RegisterRequest")
	proto.RegisterType((*RegisterResponse)(nil), "messaging.RegisterResponse")
	proto.RegisterType((*FinalizeRequest)(nil), "messaging.FinalizeRequest")
	proto.RegisterType((*FinalizeResponse)(nil), "messaging.FinalizeResponse")
}

func init() { proto.RegisterFile("messaging/order.proto", fileDescriptor_f8af3e412e3326d2) }

var fileDescriptor_f8af3e412e3326d2 = []byte{
	// 534 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x94, 0xdd, 0x6e, 0xd3, 0x30,
	0x14, 0xc7, 0xe7, 0xa6, 0x69, 0xbb, 0x33, 0x46, 0xc7, 0x11, 0x1f, 0x26, 0x0c, 0xa9, 0x8a, 0xb8,
	0xe8, 0x05, 0x14, 0x54, 0x24, 0x40, 0xdc, 0xa1, 0x6e, 0x43, 0x45, 0xd3, 0x90, 0x52, 0xee, 0xb8,
	0xf2, 0x5a, 0xaf, 0x58, 0xe4, 0xa3, 0xb3, 0x9d, 0xc1, 0x78, 0x1b, 0x9e, 0x81, 0x67, 0x40, 0xe2,
	0xb1, 0x50, 0x12, 0xa7, 0x4d, 0xb2, 0x34, 0x37, 0xbb, 0x8b, 0xcf, 0x3f, 0xfe, 0x1d, 0x9f, 0xff,
	0x39, 0x36, 0x3c, 0x08, 0xb8, 0x52, 0x6c, 0x29, 0xc2, 0xe5, 0xcb, 0x48, 0x2e, 0xb8, 0x1c, 0xad,
	0x64, 0xa4, 0x23, 0xdc, 0x5d, 0x87, 0xdd, 0xbf, 0x04, 0xf6, 0x3d, 0xbe, 0x8a, 0xa4, 0xf6, 0xf8,
	0x65, 0xcc, 0x95, 0x46, 0x0a, 0x5d, 0xf5, 0x8d, 0xc9, 0xc5, 0xf4, 0x88, 0x92, 0x01, 0x19, 0xda,
	0x5e, 0xbe, 0xc4, 0x43, 0xd8, 0x95, 0x7c, 0xe5, 0x8b, 0x39, 0x9b, 0x1e, 0xd1, 0x56, 0xaa, 0x6d,
	0x02, 0xe8, 0xc2, 0x1d, 0xcd, 0x43, 0xcd, 0xb4, 0xb8, 0xe2, 0x93, 0x58, 0x53, 0x6b, 0x60, 0x0d,
	0x6d, 0xaf, 0x14, 0x4b, 0x08, 0x17, 0x22, 0x64, 0xbe, 0xf8, 0xc5, 0x17, 0xb4, 0x3d, 0x20, 0xc3,
	0x9e, 0xb7, 0x09, 0x24, 0x6a, 0x20, 0xc2, 0xd3, 0x68, 0x79, 0x16, 0x07, 0xd4, 0xce, 0xf8, 0xeb,
	0x40, 0xaa, 0xb2, 0x9f, 0x46, 0xed, 0x18, 0x35, 0x0f, 0xb8, 0x4f, 0xa1, 0x3b, 0x0d, 0xf5, 0xa9,
	0x50, 0x1a, 0x11, 0xda, 0xbe, 0x50, 0x9a, 0x92, 0xf4, 0x00, 0xe9, 0xb7, 0xfb, 0xc7, 0x82, 0xfe,
	0x49, 0x24, 0x7f, 0x30, 0xb9, 0xf0, 0xb8, 0x5a, 0x45, 0xa1, 0xe2, 0x78, 0x09, 0x8f, 0x94, 0x66,
	0x52, 0x7f, 0xf4, 0xa3, 0x73, 0xe6, 0xcf, 0x92, 0xf2, 0xc3, 0x39, 0x3f, 0x8b, 0x03, 0x95, 0x6e,
	0xdd, 0x1b, 0xbf, 0x1d, 0xad, 0x7d, 0x1a, 0x55, 0x36, 0x8f, 0x66, 0xf5, 0x3b, 0x8f, 0x43, 0x2d,
	0xaf, 0xbd, 0x6d, 0x5c, 0x9c, 0xc1, 0xfe, 0x3c, 0x0a, 0x02, 0xa1, 0x35, 0x5f, 0x4c, 0x62, 0xad,
	0x68, 0x2b, 0x4d, 0xf4, 0xa2, 0x21, 0xd1, 0xa4, 0xf8, 0x7f, 0x86, 0x2f, 0x33, 0xf0, 0x0d, 0x3c,
	0xac, 0xcf, 0x47, 0xad, 0xd4, 0xa5, 0x2d, 0xaa, 0xf3, 0x09, 0x0e, 0x9b, 0xaa, 0xc0, 0x03, 0xb0,
	0xbe, 0xf3, 0x6b, 0x33, 0x04, 0xc9, 0x27, 0xde, 0x07, 0xfb, 0x8a, 0xf9, 0x31, 0x37, 0xcd, 0xcf,
	0x16, 0xef, 0x5b, 0xef, 0x88, 0xf3, 0x05, 0xf0, 0xe6, 0x41, 0x6b, 0x08, 0xc3, 0x22, 0x61, 0x6f,
	0x8c, 0x85, 0xc2, 0x4d, 0xfb, 0x0a, 0x54, 0xf7, 0x1f, 0x81, 0xbb, 0xf9, 0x70, 0x9a, 0xa6, 0x6d,
	0x2f, 0x96, 0x34, 0x15, 0x8b, 0xcf, 0xea, 0x9c, 0xb7, 0xab, 0x56, 0x96, 0x26, 0xd0, 0x6a, 0x9c,
	0xc0, 0x76, 0x65, 0x02, 0xcb, 0xb3, 0x6d, 0x57, 0x66, 0xdb, 0x9d, 0x42, 0xdf, 0xe3, 0x4b, 0xa1,
	0x34, 0x97, 0xb7, 0xbc, 0x68, 0x2e, 0xc2, 0xc1, 0x06, 0x95, 0xd9, 0xe2, 0x7e, 0x85, 0xfe, 0x89,
	0xc9, 0x95, 0xe3, 0x1d, 0xe8, 0x19, 0x9e, 0x32, 0x57, 0x61, 0xbd, 0xc6, 0xe7, 0x70, 0x2f, 0x3f,
	0xda, 0x87, 0x0b, 0xcd, 0xa5, 0x71, 0x24, 0x49, 0x74, 0x53, 0x48, 0x12, 0x6e, 0xe0, 0x59, 0xc2,
	0xf1, 0xef, 0x16, 0xd8, 0x9f, 0x93, 0x27, 0x05, 0x27, 0xd0, 0xc9, 0x7a, 0x84, 0xb4, 0xd0, 0xcd,
	0xd2, 0x9b, 0xe2, 0x3c, 0xae, 0x51, 0xcc, 0xc9, 0x77, 0x86, 0xe4, 0x15, 0xc1, 0x63, 0xe8, 0x9a,
	0xc1, 0x6f, 0xa0, 0x38, 0xdb, 0xaf, 0xc9, 0x1a, 0xd3, 0xcb, 0xad, 0x41, 0xa7, 0xc4, 0x29, 0x59,
	0xef, 0x3c, 0xa9, 0xd5, 0x72, 0x54, 0x82, 0xc9, 0x0b, 0x2e, 0x61, 0x2a, 0x16, 0x97, 0x30, 0x55,
	0x87, 0xdc, 0x9d, 0xf3, 0x4e, 0xfa, 0xda, 0xbe, 0xfe, 0x1f, 0x00, 0x00, 0xff, 0xff, 0x6e, 0xd7,
	0xe6, 0x3d, 0x86, 0x05, 0x00, 0x00,
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

func (c *orderClient) Forward(ctx context.Context, opts ...grpc.CallOption) (Order_ForwardClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Order_serviceDesc.Streams[1], "/messaging.Order/Forward", opts...)
	if err != nil {
		return nil, err
	}
	x := &orderForwardClient{stream}
	return x, nil
}

type Order_ForwardClient interface {
	Send(*ReportRequest) error
	Recv() (*ForwardResponse, error)
	grpc.ClientStream
}

type orderForwardClient struct {
	grpc.ClientStream
}

func (x *orderForwardClient) Send(m *ReportRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *orderForwardClient) Recv() (*ForwardResponse, error) {
	m := new(ForwardResponse)
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
	Forward(Order_ForwardServer) error
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	Finalize(context.Context, *FinalizeRequest) (*FinalizeResponse, error)
}

// UnimplementedOrderServer can be embedded to have forward compatible implementations.
type UnimplementedOrderServer struct {
}

func (*UnimplementedOrderServer) Report(srv Order_ReportServer) error {
	return status.Errorf(codes.Unimplemented, "method Report not implemented")
}
func (*UnimplementedOrderServer) Forward(srv Order_ForwardServer) error {
	return status.Errorf(codes.Unimplemented, "method Forward not implemented")
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

func _Order_Forward_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(OrderServer).Forward(&orderForwardServer{stream})
}

type Order_ForwardServer interface {
	Send(*ForwardResponse) error
	Recv() (*ReportRequest, error)
	grpc.ServerStream
}

type orderForwardServer struct {
	grpc.ServerStream
}

func (x *orderForwardServer) Send(m *ForwardResponse) error {
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
		{
			StreamName:    "Forward",
			Handler:       _Order_Forward_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "messaging/order.proto",
}
