// Code generated by protoc-gen-go. DO NOT EDIT.
// source: queue/queue.proto

/*
Package queue is a generated protocol buffer package.

It is generated from these files:
	queue/queue.proto

It has these top-level messages:
	RPCRecord
	RPCCausality
	RPCRecords
	RPCReply
	RPCQueue
	RPCToken
	RPCMaintainers
	RPCIndexers
	RPCTOIDToken
*/
package queue

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

type RPCQueue struct {
	Version uint32 `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	Queue   string `protobuf:"bytes,2,opt,name=queue" json:"queue,omitempty"`
}

func (m *RPCQueue) Reset()                    { *m = RPCQueue{} }
func (m *RPCQueue) String() string            { return proto.CompactTextString(m) }
func (*RPCQueue) ProtoMessage()               {}
func (*RPCQueue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *RPCQueue) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *RPCQueue) GetQueue() string {
	if m != nil {
		return m.Queue
	}
	return ""
}

type RPCToken struct {
	Lastlid uint32 `protobuf:"varint,1,opt,name=lastlid" json:"lastlid,omitempty"`
}

func (m *RPCToken) Reset()                    { *m = RPCToken{} }
func (m *RPCToken) String() string            { return proto.CompactTextString(m) }
func (*RPCToken) ProtoMessage()               {}
func (*RPCToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *RPCToken) GetLastlid() uint32 {
	if m != nil {
		return m.Lastlid
	}
	return 0
}

type RPCMaintainers struct {
	Version    uint32   `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	Maintainer []string `protobuf:"bytes,2,rep,name=maintainer" json:"maintainer,omitempty"`
}

func (m *RPCMaintainers) Reset()                    { *m = RPCMaintainers{} }
func (m *RPCMaintainers) String() string            { return proto.CompactTextString(m) }
func (*RPCMaintainers) ProtoMessage()               {}
func (*RPCMaintainers) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *RPCMaintainers) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *RPCMaintainers) GetMaintainer() []string {
	if m != nil {
		return m.Maintainer
	}
	return nil
}

type RPCIndexers struct {
	Version uint32   `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	Indexer []string `protobuf:"bytes,2,rep,name=indexer" json:"indexer,omitempty"`
}

func (m *RPCIndexers) Reset()                    { *m = RPCIndexers{} }
func (m *RPCIndexers) String() string            { return proto.CompactTextString(m) }
func (*RPCIndexers) ProtoMessage()               {}
func (*RPCIndexers) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *RPCIndexers) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *RPCIndexers) GetIndexer() []string {
	if m != nil {
		return m.Indexer
	}
	return nil
}

type RPCTOIDToken struct {
	MaxTOId         []uint32     `protobuf:"varint,1,rep,packed,name=maxTOId" json:"maxTOId,omitempty"`
	LastLId         uint32       `protobuf:"varint,2,opt,name=lastLId" json:"lastLId,omitempty"`
	DeferredRecords []*RPCRecord `protobuf:"bytes,3,rep,name=deferredRecords" json:"deferredRecords,omitempty"`
}

func (m *RPCTOIDToken) Reset()                    { *m = RPCTOIDToken{} }
func (m *RPCTOIDToken) String() string            { return proto.CompactTextString(m) }
func (*RPCTOIDToken) ProtoMessage()               {}
func (*RPCTOIDToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *RPCTOIDToken) GetMaxTOId() []uint32 {
	if m != nil {
		return m.MaxTOId
	}
	return nil
}

func (m *RPCTOIDToken) GetLastLId() uint32 {
	if m != nil {
		return m.LastLId
	}
	return 0
}

func (m *RPCTOIDToken) GetDeferredRecords() []*RPCRecord {
	if m != nil {
		return m.DeferredRecords
	}
	return nil
}

func init() {
	proto.RegisterType((*RPCRecord)(nil), "RPCRecord")
	proto.RegisterType((*RPCCausality)(nil), "RPCCausality")
	proto.RegisterType((*RPCRecords)(nil), "RPCRecords")
	proto.RegisterType((*RPCReply)(nil), "RPCReply")
	proto.RegisterType((*RPCQueue)(nil), "RPCQueue")
	proto.RegisterType((*RPCToken)(nil), "RPCToken")
	proto.RegisterType((*RPCMaintainers)(nil), "RPCMaintainers")
	proto.RegisterType((*RPCIndexers)(nil), "RPCIndexers")
	proto.RegisterType((*RPCTOIDToken)(nil), "RPCTOIDToken")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Queue service

type QueueClient interface {
	ReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error)
	ReceiveToken(ctx context.Context, in *RPCToken, opts ...grpc.CallOption) (*RPCReply, error)
	UpdateNextQueue(ctx context.Context, in *RPCQueue, opts ...grpc.CallOption) (*RPCReply, error)
	UpdateMaintainers(ctx context.Context, in *RPCMaintainers, opts ...grpc.CallOption) (*RPCReply, error)
	UpdateIndexers(ctx context.Context, in *RPCIndexers, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDReceiveToken(ctx context.Context, in *RPCTOIDToken, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDUpdateNextQueue(ctx context.Context, in *RPCQueue, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDUpdateMaintainers(ctx context.Context, in *RPCMaintainers, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDUpdateIndexers(ctx context.Context, in *RPCIndexers, opts ...grpc.CallOption) (*RPCReply, error)
}

type queueClient struct {
	cc *grpc.ClientConn
}

func NewQueueClient(cc *grpc.ClientConn) QueueClient {
	return &queueClient{cc}
}

func (c *queueClient) ReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/ReceiveRecords", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) ReceiveToken(ctx context.Context, in *RPCToken, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/ReceiveToken", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) UpdateNextQueue(ctx context.Context, in *RPCQueue, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/UpdateNextQueue", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) UpdateMaintainers(ctx context.Context, in *RPCMaintainers, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/UpdateMaintainers", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) UpdateIndexers(ctx context.Context, in *RPCIndexers, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/UpdateIndexers", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) TOIDReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/TOIDReceiveRecords", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) TOIDReceiveToken(ctx context.Context, in *RPCTOIDToken, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/TOIDReceiveToken", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) TOIDUpdateNextQueue(ctx context.Context, in *RPCQueue, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/TOIDUpdateNextQueue", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) TOIDUpdateMaintainers(ctx context.Context, in *RPCMaintainers, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/TOIDUpdateMaintainers", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueClient) TOIDUpdateIndexers(ctx context.Context, in *RPCIndexers, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Queue/TOIDUpdateIndexers", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Queue service

type QueueServer interface {
	ReceiveRecords(context.Context, *RPCRecords) (*RPCReply, error)
	ReceiveToken(context.Context, *RPCToken) (*RPCReply, error)
	UpdateNextQueue(context.Context, *RPCQueue) (*RPCReply, error)
	UpdateMaintainers(context.Context, *RPCMaintainers) (*RPCReply, error)
	UpdateIndexers(context.Context, *RPCIndexers) (*RPCReply, error)
	TOIDReceiveRecords(context.Context, *RPCRecords) (*RPCReply, error)
	TOIDReceiveToken(context.Context, *RPCTOIDToken) (*RPCReply, error)
	TOIDUpdateNextQueue(context.Context, *RPCQueue) (*RPCReply, error)
	TOIDUpdateMaintainers(context.Context, *RPCMaintainers) (*RPCReply, error)
	TOIDUpdateIndexers(context.Context, *RPCIndexers) (*RPCReply, error)
}

func RegisterQueueServer(s *grpc.Server, srv QueueServer) {
	s.RegisterService(&_Queue_serviceDesc, srv)
}

func _Queue_ReceiveRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCRecords)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).ReceiveRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/ReceiveRecords",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).ReceiveRecords(ctx, req.(*RPCRecords))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_ReceiveToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCToken)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).ReceiveToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/ReceiveToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).ReceiveToken(ctx, req.(*RPCToken))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_UpdateNextQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCQueue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).UpdateNextQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/UpdateNextQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).UpdateNextQueue(ctx, req.(*RPCQueue))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_UpdateMaintainers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCMaintainers)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).UpdateMaintainers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/UpdateMaintainers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).UpdateMaintainers(ctx, req.(*RPCMaintainers))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_UpdateIndexers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCIndexers)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).UpdateIndexers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/UpdateIndexers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).UpdateIndexers(ctx, req.(*RPCIndexers))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_TOIDReceiveRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCRecords)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).TOIDReceiveRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/TOIDReceiveRecords",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).TOIDReceiveRecords(ctx, req.(*RPCRecords))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_TOIDReceiveToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCTOIDToken)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).TOIDReceiveToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/TOIDReceiveToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).TOIDReceiveToken(ctx, req.(*RPCTOIDToken))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_TOIDUpdateNextQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCQueue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).TOIDUpdateNextQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/TOIDUpdateNextQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).TOIDUpdateNextQueue(ctx, req.(*RPCQueue))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_TOIDUpdateMaintainers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCMaintainers)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).TOIDUpdateMaintainers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/TOIDUpdateMaintainers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).TOIDUpdateMaintainers(ctx, req.(*RPCMaintainers))
	}
	return interceptor(ctx, in, info, handler)
}

func _Queue_TOIDUpdateIndexers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCIndexers)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueServer).TOIDUpdateIndexers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Queue/TOIDUpdateIndexers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueServer).TOIDUpdateIndexers(ctx, req.(*RPCIndexers))
	}
	return interceptor(ctx, in, info, handler)
}

var _Queue_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Queue",
	HandlerType: (*QueueServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReceiveRecords",
			Handler:    _Queue_ReceiveRecords_Handler,
		},
		{
			MethodName: "ReceiveToken",
			Handler:    _Queue_ReceiveToken_Handler,
		},
		{
			MethodName: "UpdateNextQueue",
			Handler:    _Queue_UpdateNextQueue_Handler,
		},
		{
			MethodName: "UpdateMaintainers",
			Handler:    _Queue_UpdateMaintainers_Handler,
		},
		{
			MethodName: "UpdateIndexers",
			Handler:    _Queue_UpdateIndexers_Handler,
		},
		{
			MethodName: "TOIDReceiveRecords",
			Handler:    _Queue_TOIDReceiveRecords_Handler,
		},
		{
			MethodName: "TOIDReceiveToken",
			Handler:    _Queue_TOIDReceiveToken_Handler,
		},
		{
			MethodName: "TOIDUpdateNextQueue",
			Handler:    _Queue_TOIDUpdateNextQueue_Handler,
		},
		{
			MethodName: "TOIDUpdateMaintainers",
			Handler:    _Queue_TOIDUpdateMaintainers_Handler,
		},
		{
			MethodName: "TOIDUpdateIndexers",
			Handler:    _Queue_TOIDUpdateIndexers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "queue/queue.proto",
}

func init() { proto.RegisterFile("queue/queue.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 569 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x54, 0xdd, 0x6e, 0xd3, 0x4c,
	0x10, 0xfd, 0x6c, 0xa7, 0x4d, 0x3d, 0x69, 0x92, 0x76, 0xbf, 0x22, 0xad, 0x2a, 0x84, 0x2c, 0xab,
	0x42, 0x86, 0x0a, 0x57, 0x84, 0x5f, 0xf5, 0x0e, 0x05, 0x2e, 0x82, 0x80, 0x86, 0x55, 0x78, 0x80,
	0x25, 0x5e, 0x52, 0xab, 0x8e, 0x1d, 0xbc, 0x9b, 0x28, 0x79, 0x26, 0x9e, 0x89, 0x77, 0x41, 0x3b,
	0xf6, 0xfa, 0x07, 0xa4, 0x2a, 0x37, 0xd1, 0x99, 0x99, 0xb3, 0x33, 0x73, 0xe6, 0x28, 0x86, 0xd3,
	0x9f, 0x6b, 0xb1, 0x16, 0x57, 0xf8, 0x1b, 0xae, 0xf2, 0x4c, 0x65, 0xfe, 0x2f, 0x1b, 0x5c, 0x36,
	0x1d, 0x33, 0x31, 0xcf, 0xf2, 0x88, 0x0c, 0xc0, 0x8e, 0x23, 0x6a, 0x79, 0x56, 0xd0, 0x61, 0x76,
	0x1c, 0x91, 0x87, 0xe0, 0xaa, 0x78, 0x29, 0xa4, 0xe2, 0xcb, 0x15, 0xb5, 0x3d, 0x2b, 0x70, 0x58,
	0x9d, 0x20, 0x04, 0x3a, 0xb7, 0x99, 0x54, 0xd4, 0xf1, 0xac, 0xa0, 0xcf, 0x10, 0x93, 0x13, 0x70,
	0x92, 0x38, 0xa2, 0x1d, 0x4c, 0x69, 0x48, 0x02, 0xe8, 0x28, 0xbe, 0x90, 0xf4, 0xc0, 0x73, 0x82,
	0xde, 0xe8, 0x2c, 0xac, 0xa6, 0x85, 0x33, 0xbe, 0x90, 0x1f, 0x52, 0x95, 0xef, 0x18, 0x32, 0xb0,
	0x1f, 0x97, 0xb7, 0xf4, 0xd0, 0x73, 0x82, 0x0e, 0x43, 0xac, 0x73, 0x52, 0x88, 0x88, 0x76, 0x71,
	0x27, 0xc4, 0x3a, 0xa7, 0xb2, 0x38, 0xa2, 0x47, 0xc5, 0x5c, 0x8d, 0xc9, 0x25, 0xb8, 0x73, 0xbe,
	0x96, 0x3c, 0x89, 0xd5, 0x8e, 0xba, 0x9e, 0x15, 0xf4, 0x46, 0x7d, 0x3d, 0x6a, 0x6c, 0x92, 0xac,
	0xae, 0x9f, 0xbf, 0x01, 0xb7, 0x9a, 0xad, 0x37, 0xbe, 0x13, 0x3b, 0x14, 0xed, 0x32, 0x0d, 0xc9,
	0x19, 0x1c, 0x6c, 0x78, 0xb2, 0x16, 0xa8, 0xd8, 0x65, 0x45, 0x70, 0x6d, 0xbf, 0xb5, 0xfc, 0xd7,
	0x70, 0xdc, 0xec, 0x59, 0x5d, 0xc0, 0x6a, 0x5c, 0xc0, 0x6c, 0x67, 0xd7, 0xdb, 0xf9, 0x23, 0x80,
	0x4a, 0xb6, 0x24, 0x17, 0xd0, 0xcd, 0x0b, 0x48, 0x2d, 0x3c, 0x0a, 0xd4, 0x47, 0x61, 0xa6, 0xe4,
	0x5f, 0xc0, 0x11, 0x66, 0x57, 0xc9, 0x8e, 0x50, 0xe8, 0x2e, 0x85, 0x94, 0x7c, 0x21, 0xca, 0x3d,
	0x4d, 0xe8, 0x5f, 0x23, 0xeb, 0xab, 0x76, 0x54, 0xb3, 0x36, 0x22, 0x97, 0x71, 0x96, 0x96, 0x0b,
	0x99, 0x50, 0x2b, 0x42, 0xd3, 0x8d, 0x22, 0x0c, 0xca, 0x09, 0xb3, 0xec, 0x4e, 0xa4, 0xfa, 0x6d,
	0xc2, 0xa5, 0x4a, 0x4a, 0xfb, 0xfb, 0xcc, 0x84, 0xfe, 0x47, 0x18, 0xb0, 0xe9, 0xf8, 0x33, 0x8f,
	0x53, 0xc5, 0xe3, 0x54, 0xe4, 0xf2, 0x9e, 0x39, 0x8f, 0x00, 0x96, 0x15, 0x91, 0xda, 0x9e, 0x13,
	0xb8, 0xac, 0x91, 0xf1, 0xdf, 0x41, 0x8f, 0x4d, 0xc7, 0x93, 0x34, 0x12, 0xdb, 0xfb, 0x1b, 0x51,
	0xe8, 0xc6, 0x05, 0xab, 0xec, 0x62, 0x42, 0x7f, 0x8b, 0x16, 0xcc, 0x6e, 0x26, 0xef, 0xab, 0xc5,
	0x97, 0x7c, 0x3b, 0xbb, 0x99, 0x44, 0x78, 0xcc, 0x3e, 0x33, 0xa1, 0x91, 0xf4, 0x69, 0x62, 0xbc,
	0x30, 0x21, 0x79, 0x09, 0xc3, 0x48, 0xfc, 0x10, 0x79, 0x2e, 0xa2, 0xd2, 0x13, 0xea, 0xfc, 0x63,
	0xc4, 0xdf, 0x94, 0xd1, 0x6f, 0x07, 0x0e, 0x8a, 0x43, 0x3f, 0x85, 0x01, 0x13, 0x73, 0x11, 0x6f,
	0x84, 0xb1, 0xb4, 0x57, 0x3f, 0x94, 0xe7, 0x6e, 0x68, 0x8c, 0xf3, 0xff, 0x23, 0x8f, 0xe1, 0xb8,
	0xe4, 0x16, 0xfb, 0x62, 0x11, 0x61, 0x9b, 0xf7, 0x04, 0x86, 0xdf, 0x56, 0x11, 0x57, 0xe2, 0x8b,
	0xd8, 0xaa, 0x62, 0x0c, 0xd6, 0x11, 0xb6, 0xa9, 0xcf, 0xe1, 0xb4, 0xa0, 0x36, 0x4d, 0x19, 0x86,
	0x6d, 0x97, 0xda, 0x4f, 0x2e, 0x61, 0x50, 0x3c, 0xa9, 0x6e, 0x7f, 0x1c, 0x36, 0x9c, 0x68, 0x93,
	0x43, 0x20, 0xfa, 0xbe, 0x7b, 0x4b, 0x0c, 0xe1, 0xa4, 0xc1, 0x2f, 0x64, 0xe2, 0x9f, 0xaf, 0x72,
	0xa9, 0xcd, 0x7f, 0x06, 0xff, 0xeb, 0xca, 0xbe, 0x72, 0x5f, 0xc1, 0x83, 0x9a, 0xbe, 0xbf, 0xe4,
	0xab, 0x42, 0xc5, 0xde, 0xb2, 0xbf, 0x1f, 0xe2, 0x17, 0xf1, 0xc5, 0x9f, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xbd, 0x1d, 0xc8, 0xd0, 0x26, 0x05, 0x00, 0x00,
}