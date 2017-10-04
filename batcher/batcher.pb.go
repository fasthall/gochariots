// Code generated by protoc-gen-go. DO NOT EDIT.
// source: batcher/batcher.proto

/*
Package batcher is a generated protocol buffer package.

It is generated from these files:
	batcher/batcher.proto

It has these top-level messages:
	RPCRecord
	RPCCausality
	RPCRecords
	RPCReply
	RPCQueues
*/
package batcher

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

type RPCRecord struct {
	Id        uint64            `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Timestamp int64             `protobuf:"varint,2,opt,name=timestamp" json:"timestamp,omitempty"`
	Host      uint32            `protobuf:"varint,3,opt,name=host" json:"host,omitempty"`
	Lid       uint32            `protobuf:"varint,4,opt,name=lid" json:"lid,omitempty"`
	Tags      map[string]string `protobuf:"bytes,5,rep,name=tags" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Hash      []uint64          `protobuf:"varint,6,rep,packed,name=hash" json:"hash,omitempty"`
	Seed      uint64            `protobuf:"varint,7,opt,name=seed" json:"seed,omitempty"`
	// for TOID record
	Toid      uint32        `protobuf:"varint,8,opt,name=toid" json:"toid,omitempty"`
	Causality *RPCCausality `protobuf:"bytes,9,opt,name=causality" json:"causality,omitempty"`
}

func (m *RPCRecord) Reset()                    { *m = RPCRecord{} }
func (m *RPCRecord) String() string            { return proto.CompactTextString(m) }
func (*RPCRecord) ProtoMessage()               {}
func (*RPCRecord) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *RPCRecord) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *RPCRecord) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *RPCRecord) GetHost() uint32 {
	if m != nil {
		return m.Host
	}
	return 0
}

func (m *RPCRecord) GetLid() uint32 {
	if m != nil {
		return m.Lid
	}
	return 0
}

func (m *RPCRecord) GetTags() map[string]string {
	if m != nil {
		return m.Tags
	}
	return nil
}

func (m *RPCRecord) GetHash() []uint64 {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *RPCRecord) GetSeed() uint64 {
	if m != nil {
		return m.Seed
	}
	return 0
}

func (m *RPCRecord) GetToid() uint32 {
	if m != nil {
		return m.Toid
	}
	return 0
}

func (m *RPCRecord) GetCausality() *RPCCausality {
	if m != nil {
		return m.Causality
	}
	return nil
}

type RPCCausality struct {
	Host uint32 `protobuf:"varint,1,opt,name=host" json:"host,omitempty"`
	Toid uint32 `protobuf:"varint,2,opt,name=toid" json:"toid,omitempty"`
}

func (m *RPCCausality) Reset()                    { *m = RPCCausality{} }
func (m *RPCCausality) String() string            { return proto.CompactTextString(m) }
func (*RPCCausality) ProtoMessage()               {}
func (*RPCCausality) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RPCCausality) GetHost() uint32 {
	if m != nil {
		return m.Host
	}
	return 0
}

func (m *RPCCausality) GetToid() uint32 {
	if m != nil {
		return m.Toid
	}
	return 0
}

type RPCRecords struct {
	Records []*RPCRecord `protobuf:"bytes,1,rep,name=records" json:"records,omitempty"`
}

func (m *RPCRecords) Reset()                    { *m = RPCRecords{} }
func (m *RPCRecords) String() string            { return proto.CompactTextString(m) }
func (*RPCRecords) ProtoMessage()               {}
func (*RPCRecords) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *RPCRecords) GetRecords() []*RPCRecord {
	if m != nil {
		return m.Records
	}
	return nil
}

type RPCReply struct {
	Message string `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
}

func (m *RPCReply) Reset()                    { *m = RPCReply{} }
func (m *RPCReply) String() string            { return proto.CompactTextString(m) }
func (*RPCReply) ProtoMessage()               {}
func (*RPCReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *RPCReply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type RPCQueues struct {
	Version uint32   `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	Queues  []string `protobuf:"bytes,2,rep,name=queues" json:"queues,omitempty"`
}

func (m *RPCQueues) Reset()                    { *m = RPCQueues{} }
func (m *RPCQueues) String() string            { return proto.CompactTextString(m) }
func (*RPCQueues) ProtoMessage()               {}
func (*RPCQueues) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *RPCQueues) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *RPCQueues) GetQueues() []string {
	if m != nil {
		return m.Queues
	}
	return nil
}

func init() {
	proto.RegisterType((*RPCRecord)(nil), "RPCRecord")
	proto.RegisterType((*RPCCausality)(nil), "RPCCausality")
	proto.RegisterType((*RPCRecords)(nil), "RPCRecords")
	proto.RegisterType((*RPCReply)(nil), "RPCReply")
	proto.RegisterType((*RPCQueues)(nil), "RPCQueues")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Batcher service

type BatcherClient interface {
	ReceiveRecord(ctx context.Context, in *RPCRecord, opts ...grpc.CallOption) (*RPCReply, error)
	ReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error)
	UpdateQueue(ctx context.Context, in *RPCQueues, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDReceiveRecord(ctx context.Context, in *RPCRecord, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDUpdateQueue(ctx context.Context, in *RPCQueues, opts ...grpc.CallOption) (*RPCReply, error)
}

type batcherClient struct {
	cc *grpc.ClientConn
}

func NewBatcherClient(cc *grpc.ClientConn) BatcherClient {
	return &batcherClient{cc}
}

func (c *batcherClient) ReceiveRecord(ctx context.Context, in *RPCRecord, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Batcher/ReceiveRecord", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *batcherClient) ReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Batcher/ReceiveRecords", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *batcherClient) UpdateQueue(ctx context.Context, in *RPCQueues, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Batcher/UpdateQueue", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *batcherClient) TOIDReceiveRecord(ctx context.Context, in *RPCRecord, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Batcher/TOIDReceiveRecord", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *batcherClient) TOIDReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Batcher/TOIDReceiveRecords", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *batcherClient) TOIDUpdateQueue(ctx context.Context, in *RPCQueues, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Batcher/TOIDUpdateQueue", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Batcher service

type BatcherServer interface {
	ReceiveRecord(context.Context, *RPCRecord) (*RPCReply, error)
	ReceiveRecords(context.Context, *RPCRecords) (*RPCReply, error)
	UpdateQueue(context.Context, *RPCQueues) (*RPCReply, error)
	TOIDReceiveRecord(context.Context, *RPCRecord) (*RPCReply, error)
	TOIDReceiveRecords(context.Context, *RPCRecords) (*RPCReply, error)
	TOIDUpdateQueue(context.Context, *RPCQueues) (*RPCReply, error)
}

func RegisterBatcherServer(s *grpc.Server, srv BatcherServer) {
	s.RegisterService(&_Batcher_serviceDesc, srv)
}

func _Batcher_ReceiveRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCRecord)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BatcherServer).ReceiveRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Batcher/ReceiveRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BatcherServer).ReceiveRecord(ctx, req.(*RPCRecord))
	}
	return interceptor(ctx, in, info, handler)
}

func _Batcher_ReceiveRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCRecords)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BatcherServer).ReceiveRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Batcher/ReceiveRecords",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BatcherServer).ReceiveRecords(ctx, req.(*RPCRecords))
	}
	return interceptor(ctx, in, info, handler)
}

func _Batcher_UpdateQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCQueues)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BatcherServer).UpdateQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Batcher/UpdateQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BatcherServer).UpdateQueue(ctx, req.(*RPCQueues))
	}
	return interceptor(ctx, in, info, handler)
}

func _Batcher_TOIDReceiveRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCRecord)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BatcherServer).TOIDReceiveRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Batcher/TOIDReceiveRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BatcherServer).TOIDReceiveRecord(ctx, req.(*RPCRecord))
	}
	return interceptor(ctx, in, info, handler)
}

func _Batcher_TOIDReceiveRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCRecords)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BatcherServer).TOIDReceiveRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Batcher/TOIDReceiveRecords",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BatcherServer).TOIDReceiveRecords(ctx, req.(*RPCRecords))
	}
	return interceptor(ctx, in, info, handler)
}

func _Batcher_TOIDUpdateQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCQueues)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BatcherServer).TOIDUpdateQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Batcher/TOIDUpdateQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BatcherServer).TOIDUpdateQueue(ctx, req.(*RPCQueues))
	}
	return interceptor(ctx, in, info, handler)
}

var _Batcher_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Batcher",
	HandlerType: (*BatcherServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReceiveRecord",
			Handler:    _Batcher_ReceiveRecord_Handler,
		},
		{
			MethodName: "ReceiveRecords",
			Handler:    _Batcher_ReceiveRecords_Handler,
		},
		{
			MethodName: "UpdateQueue",
			Handler:    _Batcher_UpdateQueue_Handler,
		},
		{
			MethodName: "TOIDReceiveRecord",
			Handler:    _Batcher_TOIDReceiveRecord_Handler,
		},
		{
			MethodName: "TOIDReceiveRecords",
			Handler:    _Batcher_TOIDReceiveRecords_Handler,
		},
		{
			MethodName: "TOIDUpdateQueue",
			Handler:    _Batcher_TOIDUpdateQueue_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "batcher/batcher.proto",
}

func init() { proto.RegisterFile("batcher/batcher.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 423 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x93, 0xdf, 0x8a, 0x13, 0x31,
	0x14, 0xc6, 0x4d, 0x66, 0xb6, 0xdd, 0x9c, 0xda, 0x55, 0xc3, 0x2a, 0x61, 0xf1, 0x62, 0x18, 0x16,
	0x09, 0xab, 0x8c, 0x50, 0x41, 0x45, 0xf0, 0xc6, 0xd5, 0x0b, 0xaf, 0xd4, 0xb0, 0x3e, 0x40, 0xb6,
	0x73, 0x68, 0x83, 0xed, 0xce, 0x38, 0xc9, 0x14, 0xe6, 0x31, 0x7c, 0x0e, 0x5f, 0x52, 0x92, 0xf9,
	0xd7, 0x22, 0x2c, 0xbd, 0xea, 0x77, 0xbe, 0x7c, 0xe4, 0xfc, 0xce, 0x69, 0x06, 0x9e, 0xde, 0x6a,
	0xb7, 0x5c, 0x63, 0xf5, 0xba, 0xfb, 0xcd, 0xca, 0xaa, 0x70, 0x45, 0xfa, 0x97, 0x02, 0x53, 0xdf,
	0xaf, 0x15, 0x2e, 0x8b, 0x2a, 0xe7, 0x67, 0x40, 0x4d, 0x2e, 0x48, 0x42, 0x64, 0xac, 0xa8, 0xc9,
	0xf9, 0x73, 0x60, 0xce, 0x6c, 0xd1, 0x3a, 0xbd, 0x2d, 0x05, 0x4d, 0x88, 0x8c, 0xd4, 0x68, 0x70,
	0x0e, 0xf1, 0xba, 0xb0, 0x4e, 0x44, 0x09, 0x91, 0x73, 0x15, 0x34, 0x7f, 0x0c, 0xd1, 0xc6, 0xe4,
	0x22, 0x0e, 0x96, 0x97, 0x5c, 0x42, 0xec, 0xf4, 0xca, 0x8a, 0x93, 0x24, 0x92, 0xb3, 0xc5, 0x79,
	0x36, 0x74, 0xcb, 0x6e, 0xf4, 0xca, 0x7e, 0xb9, 0x73, 0x55, 0xa3, 0x42, 0x22, 0xdc, 0xa7, 0xed,
	0x5a, 0x4c, 0x92, 0x48, 0xc6, 0x2a, 0x68, 0xef, 0x59, 0xc4, 0x5c, 0x4c, 0x03, 0x53, 0xd0, 0xde,
	0x73, 0x85, 0xc9, 0xc5, 0x69, 0xdb, 0xd7, 0x6b, 0xfe, 0x12, 0xd8, 0x52, 0xd7, 0x56, 0x6f, 0x8c,
	0x6b, 0x04, 0x4b, 0x88, 0x9c, 0x2d, 0xe6, 0xbe, 0xd5, 0x75, 0x6f, 0xaa, 0xf1, 0xfc, 0xe2, 0x1d,
	0xb0, 0xa1, 0xb7, 0x27, 0xfe, 0x85, 0x4d, 0x18, 0x9a, 0x29, 0x2f, 0xf9, 0x39, 0x9c, 0xec, 0xf4,
	0xa6, 0xc6, 0x30, 0x31, 0x53, 0x6d, 0xf1, 0x81, 0xbe, 0x27, 0xe9, 0x5b, 0x78, 0xb8, 0x7f, 0xe7,
	0xb0, 0x01, 0xb2, 0xb7, 0x81, 0x9e, 0x8e, 0x8e, 0x74, 0xe9, 0x02, 0x60, 0x18, 0xdb, 0xf2, 0x4b,
	0x98, 0x56, 0xad, 0x14, 0x24, 0x2c, 0x05, 0xc6, 0xa5, 0xa8, 0xfe, 0x28, 0xbd, 0x84, 0xd3, 0xe0,
	0x96, 0x9b, 0x86, 0x0b, 0x98, 0x6e, 0xd1, 0x5a, 0xbd, 0xc2, 0x8e, 0xb3, 0x2f, 0xd3, 0x8f, 0xe1,
	0xef, 0xfb, 0x51, 0x63, 0x8d, 0xd6, 0xc7, 0x76, 0x58, 0x59, 0x53, 0xdc, 0x75, 0x44, 0x7d, 0xc9,
	0x9f, 0xc1, 0xe4, 0x77, 0xc8, 0x08, 0x9a, 0x44, 0x92, 0xa9, 0xae, 0x5a, 0xfc, 0xa1, 0x30, 0xfd,
	0xd4, 0x3e, 0x08, 0x2e, 0x61, 0xae, 0x70, 0x89, 0x66, 0x87, 0xdd, 0x6b, 0xd8, 0xc3, 0xba, 0x60,
	0x59, 0x0f, 0x93, 0x3e, 0xe0, 0x57, 0x70, 0x76, 0x90, 0xb4, 0x7c, 0x36, 0x46, 0xed, 0x61, 0xf6,
	0x05, 0xcc, 0x7e, 0x96, 0xb9, 0x76, 0x18, 0x18, 0xdb, 0x3b, 0x5b, 0xdc, 0xc3, 0xdc, 0x2b, 0x78,
	0x72, 0xf3, 0xed, 0xeb, 0xe7, 0x23, 0x09, 0x32, 0xe0, 0xff, 0xa5, 0xef, 0xa3, 0xb8, 0x82, 0x47,
	0x3e, 0x7f, 0x0c, 0xc9, 0xed, 0x24, 0x7c, 0x19, 0x6f, 0xfe, 0x05, 0x00, 0x00, 0xff, 0xff, 0x2c,
	0x41, 0xc5, 0x65, 0x32, 0x03, 0x00, 0x00,
}