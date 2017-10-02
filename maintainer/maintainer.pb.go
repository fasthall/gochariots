// Code generated by protoc-gen-go. DO NOT EDIT.
// source: maintainer/maintainer.proto

/*
Package maintainer is a generated protocol buffer package.

It is generated from these files:
	maintainer/maintainer.proto

It has these top-level messages:
	RPCRecord
	RPCCausality
	RPCRecords
	RPCReply
	RPCBatchers
	RPCIndexer
	RPCLId
*/
package maintainer

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
	Host      int32             `protobuf:"varint,3,opt,name=host" json:"host,omitempty"`
	Lid       int32             `protobuf:"varint,4,opt,name=lid" json:"lid,omitempty"`
	Tags      map[string]string `protobuf:"bytes,5,rep,name=tags" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Hash      []uint64          `protobuf:"varint,6,rep,packed,name=hash" json:"hash,omitempty"`
	Seed      uint64            `protobuf:"varint,7,opt,name=seed" json:"seed,omitempty"`
	// for TOID record
	Toid      int32         `protobuf:"varint,8,opt,name=toid" json:"toid,omitempty"`
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

func (m *RPCRecord) GetHost() int32 {
	if m != nil {
		return m.Host
	}
	return 0
}

func (m *RPCRecord) GetLid() int32 {
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

func (m *RPCRecord) GetToid() int32 {
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
	Host int32 `protobuf:"varint,1,opt,name=host" json:"host,omitempty"`
	Toid int32 `protobuf:"varint,2,opt,name=toid" json:"toid,omitempty"`
}

func (m *RPCCausality) Reset()                    { *m = RPCCausality{} }
func (m *RPCCausality) String() string            { return proto.CompactTextString(m) }
func (*RPCCausality) ProtoMessage()               {}
func (*RPCCausality) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RPCCausality) GetHost() int32 {
	if m != nil {
		return m.Host
	}
	return 0
}

func (m *RPCCausality) GetToid() int32 {
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

type RPCBatchers struct {
	Version int32    `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	Batcher []string `protobuf:"bytes,2,rep,name=batcher" json:"batcher,omitempty"`
}

func (m *RPCBatchers) Reset()                    { *m = RPCBatchers{} }
func (m *RPCBatchers) String() string            { return proto.CompactTextString(m) }
func (*RPCBatchers) ProtoMessage()               {}
func (*RPCBatchers) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *RPCBatchers) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *RPCBatchers) GetBatcher() []string {
	if m != nil {
		return m.Batcher
	}
	return nil
}

type RPCIndexer struct {
	Version int32  `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	Indexer string `protobuf:"bytes,2,opt,name=indexer" json:"indexer,omitempty"`
}

func (m *RPCIndexer) Reset()                    { *m = RPCIndexer{} }
func (m *RPCIndexer) String() string            { return proto.CompactTextString(m) }
func (*RPCIndexer) ProtoMessage()               {}
func (*RPCIndexer) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *RPCIndexer) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *RPCIndexer) GetIndexer() string {
	if m != nil {
		return m.Indexer
	}
	return ""
}

type RPCLId struct {
	Lid int32 `protobuf:"varint,1,opt,name=lid" json:"lid,omitempty"`
}

func (m *RPCLId) Reset()                    { *m = RPCLId{} }
func (m *RPCLId) String() string            { return proto.CompactTextString(m) }
func (*RPCLId) ProtoMessage()               {}
func (*RPCLId) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *RPCLId) GetLid() int32 {
	if m != nil {
		return m.Lid
	}
	return 0
}

func init() {
	proto.RegisterType((*RPCRecord)(nil), "RPCRecord")
	proto.RegisterType((*RPCCausality)(nil), "RPCCausality")
	proto.RegisterType((*RPCRecords)(nil), "RPCRecords")
	proto.RegisterType((*RPCReply)(nil), "RPCReply")
	proto.RegisterType((*RPCBatchers)(nil), "RPCBatchers")
	proto.RegisterType((*RPCIndexer)(nil), "RPCIndexer")
	proto.RegisterType((*RPCLId)(nil), "RPCLId")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Maintainer service

type MaintainerClient interface {
	ReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error)
	UpdateBatchers(ctx context.Context, in *RPCBatchers, opts ...grpc.CallOption) (*RPCReply, error)
	UpdateIndexer(ctx context.Context, in *RPCIndexer, opts ...grpc.CallOption) (*RPCReply, error)
	ReadByLId(ctx context.Context, in *RPCLId, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDUpdateBatchers(ctx context.Context, in *RPCBatchers, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDUpdateIndexer(ctx context.Context, in *RPCIndexer, opts ...grpc.CallOption) (*RPCReply, error)
	TOIDReadByLId(ctx context.Context, in *RPCLId, opts ...grpc.CallOption) (*RPCReply, error)
}

type maintainerClient struct {
	cc *grpc.ClientConn
}

func NewMaintainerClient(cc *grpc.ClientConn) MaintainerClient {
	return &maintainerClient{cc}
}

func (c *maintainerClient) ReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Maintainer/ReceiveRecords", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *maintainerClient) UpdateBatchers(ctx context.Context, in *RPCBatchers, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Maintainer/UpdateBatchers", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *maintainerClient) UpdateIndexer(ctx context.Context, in *RPCIndexer, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Maintainer/UpdateIndexer", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *maintainerClient) ReadByLId(ctx context.Context, in *RPCLId, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Maintainer/ReadByLId", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *maintainerClient) TOIDReceiveRecords(ctx context.Context, in *RPCRecords, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Maintainer/TOIDReceiveRecords", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *maintainerClient) TOIDUpdateBatchers(ctx context.Context, in *RPCBatchers, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Maintainer/TOIDUpdateBatchers", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *maintainerClient) TOIDUpdateIndexer(ctx context.Context, in *RPCIndexer, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Maintainer/TOIDUpdateIndexer", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *maintainerClient) TOIDReadByLId(ctx context.Context, in *RPCLId, opts ...grpc.CallOption) (*RPCReply, error) {
	out := new(RPCReply)
	err := grpc.Invoke(ctx, "/Maintainer/TOIDReadByLId", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Maintainer service

type MaintainerServer interface {
	ReceiveRecords(context.Context, *RPCRecords) (*RPCReply, error)
	UpdateBatchers(context.Context, *RPCBatchers) (*RPCReply, error)
	UpdateIndexer(context.Context, *RPCIndexer) (*RPCReply, error)
	ReadByLId(context.Context, *RPCLId) (*RPCReply, error)
	TOIDReceiveRecords(context.Context, *RPCRecords) (*RPCReply, error)
	TOIDUpdateBatchers(context.Context, *RPCBatchers) (*RPCReply, error)
	TOIDUpdateIndexer(context.Context, *RPCIndexer) (*RPCReply, error)
	TOIDReadByLId(context.Context, *RPCLId) (*RPCReply, error)
}

func RegisterMaintainerServer(s *grpc.Server, srv MaintainerServer) {
	s.RegisterService(&_Maintainer_serviceDesc, srv)
}

func _Maintainer_ReceiveRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCRecords)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaintainerServer).ReceiveRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Maintainer/ReceiveRecords",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaintainerServer).ReceiveRecords(ctx, req.(*RPCRecords))
	}
	return interceptor(ctx, in, info, handler)
}

func _Maintainer_UpdateBatchers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCBatchers)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaintainerServer).UpdateBatchers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Maintainer/UpdateBatchers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaintainerServer).UpdateBatchers(ctx, req.(*RPCBatchers))
	}
	return interceptor(ctx, in, info, handler)
}

func _Maintainer_UpdateIndexer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCIndexer)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaintainerServer).UpdateIndexer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Maintainer/UpdateIndexer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaintainerServer).UpdateIndexer(ctx, req.(*RPCIndexer))
	}
	return interceptor(ctx, in, info, handler)
}

func _Maintainer_ReadByLId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCLId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaintainerServer).ReadByLId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Maintainer/ReadByLId",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaintainerServer).ReadByLId(ctx, req.(*RPCLId))
	}
	return interceptor(ctx, in, info, handler)
}

func _Maintainer_TOIDReceiveRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCRecords)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaintainerServer).TOIDReceiveRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Maintainer/TOIDReceiveRecords",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaintainerServer).TOIDReceiveRecords(ctx, req.(*RPCRecords))
	}
	return interceptor(ctx, in, info, handler)
}

func _Maintainer_TOIDUpdateBatchers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCBatchers)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaintainerServer).TOIDUpdateBatchers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Maintainer/TOIDUpdateBatchers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaintainerServer).TOIDUpdateBatchers(ctx, req.(*RPCBatchers))
	}
	return interceptor(ctx, in, info, handler)
}

func _Maintainer_TOIDUpdateIndexer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCIndexer)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaintainerServer).TOIDUpdateIndexer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Maintainer/TOIDUpdateIndexer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaintainerServer).TOIDUpdateIndexer(ctx, req.(*RPCIndexer))
	}
	return interceptor(ctx, in, info, handler)
}

func _Maintainer_TOIDReadByLId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RPCLId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MaintainerServer).TOIDReadByLId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Maintainer/TOIDReadByLId",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MaintainerServer).TOIDReadByLId(ctx, req.(*RPCLId))
	}
	return interceptor(ctx, in, info, handler)
}

var _Maintainer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Maintainer",
	HandlerType: (*MaintainerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReceiveRecords",
			Handler:    _Maintainer_ReceiveRecords_Handler,
		},
		{
			MethodName: "UpdateBatchers",
			Handler:    _Maintainer_UpdateBatchers_Handler,
		},
		{
			MethodName: "UpdateIndexer",
			Handler:    _Maintainer_UpdateIndexer_Handler,
		},
		{
			MethodName: "ReadByLId",
			Handler:    _Maintainer_ReadByLId_Handler,
		},
		{
			MethodName: "TOIDReceiveRecords",
			Handler:    _Maintainer_TOIDReceiveRecords_Handler,
		},
		{
			MethodName: "TOIDUpdateBatchers",
			Handler:    _Maintainer_TOIDUpdateBatchers_Handler,
		},
		{
			MethodName: "TOIDUpdateIndexer",
			Handler:    _Maintainer_TOIDUpdateIndexer_Handler,
		},
		{
			MethodName: "TOIDReadByLId",
			Handler:    _Maintainer_TOIDReadByLId_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "maintainer/maintainer.proto",
}

func init() { proto.RegisterFile("maintainer/maintainer.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 478 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x93, 0xdf, 0x6e, 0xd3, 0x30,
	0x14, 0xc6, 0xc9, 0x9f, 0xb6, 0xcb, 0xe9, 0x5a, 0x81, 0xb5, 0x0b, 0xab, 0x70, 0x11, 0xa2, 0x21,
	0x05, 0x26, 0x32, 0xa9, 0x48, 0x80, 0xb8, 0x82, 0x15, 0x2e, 0x2a, 0x81, 0xa8, 0xac, 0xf1, 0x00,
	0x5e, 0x7d, 0xd4, 0x5a, 0xb4, 0x49, 0x15, 0x7b, 0x15, 0x79, 0x26, 0xde, 0x8c, 0xa7, 0x40, 0x76,
	0xe2, 0xa4, 0xbd, 0xd9, 0x7a, 0xf7, 0x9d, 0xe3, 0xaf, 0xc7, 0xbf, 0xef, 0xb8, 0x81, 0xe7, 0x5b,
	0x2e, 0x73, 0xcd, 0x65, 0x8e, 0xe5, 0x75, 0x27, 0xb3, 0x5d, 0x59, 0xe8, 0x22, 0xf9, 0xeb, 0x43,
	0xc4, 0x16, 0x33, 0x86, 0xcb, 0xa2, 0x14, 0x64, 0x0c, 0xbe, 0x14, 0xd4, 0x8b, 0xbd, 0x34, 0x64,
	0xbe, 0x14, 0xe4, 0x05, 0x44, 0x5a, 0x6e, 0x51, 0x69, 0xbe, 0xdd, 0x51, 0x3f, 0xf6, 0xd2, 0x80,
	0x75, 0x0d, 0x42, 0x20, 0x5c, 0x17, 0x4a, 0xd3, 0x20, 0xf6, 0xd2, 0x1e, 0xb3, 0x9a, 0x3c, 0x85,
	0x60, 0x23, 0x05, 0x0d, 0x6d, 0xcb, 0x48, 0x92, 0x42, 0xa8, 0xf9, 0x4a, 0xd1, 0x5e, 0x1c, 0xa4,
	0xc3, 0xe9, 0x45, 0xd6, 0xde, 0x96, 0xdd, 0xf2, 0x95, 0xfa, 0x96, 0xeb, 0xb2, 0x62, 0xd6, 0x61,
	0xe7, 0x71, 0xb5, 0xa6, 0xfd, 0x38, 0x48, 0x43, 0x66, 0xb5, 0xe9, 0x29, 0x44, 0x41, 0x07, 0x96,
	0xc9, 0x6a, 0xd3, 0xd3, 0x85, 0x14, 0xf4, 0xac, 0xbe, 0xd7, 0x68, 0x72, 0x05, 0xd1, 0x92, 0xdf,
	0x2b, 0xbe, 0x91, 0xba, 0xa2, 0x51, 0xec, 0xa5, 0xc3, 0xe9, 0xc8, 0x5c, 0x35, 0x73, 0x4d, 0xd6,
	0x9d, 0x4f, 0x3e, 0x40, 0xd4, 0xde, 0x6d, 0x88, 0x7f, 0x63, 0x65, 0x43, 0x47, 0xcc, 0x48, 0x72,
	0x01, 0xbd, 0x3d, 0xdf, 0xdc, 0xa3, 0x4d, 0x1c, 0xb1, 0xba, 0xf8, 0xe4, 0x7f, 0xf4, 0x92, 0xf7,
	0x70, 0x7e, 0x38, 0xb3, 0xdd, 0x80, 0x77, 0xb0, 0x01, 0x47, 0xe7, 0x77, 0x74, 0xc9, 0x14, 0xa0,
	0x8d, 0xad, 0xc8, 0x25, 0x0c, 0xca, 0x5a, 0x52, 0xcf, 0x2e, 0x05, 0xba, 0xa5, 0x30, 0x77, 0x94,
	0x5c, 0xc2, 0x99, 0xed, 0xee, 0x36, 0x15, 0xa1, 0x30, 0xd8, 0xa2, 0x52, 0x7c, 0x85, 0x0d, 0xa7,
	0x2b, 0x93, 0x2f, 0x30, 0x64, 0x8b, 0xd9, 0x0d, 0xd7, 0xcb, 0x35, 0x96, 0xca, 0x18, 0xf7, 0x58,
	0x2a, 0x59, 0xe4, 0x0d, 0x93, 0x2b, 0xcd, 0xc9, 0x5d, 0xed, 0xa2, 0x7e, 0x1c, 0x98, 0x11, 0x4d,
	0x99, 0x7c, 0xb6, 0x70, 0xf3, 0x5c, 0xe0, 0x1f, 0x2c, 0x1f, 0x9e, 0x20, 0x6b, 0x53, 0xb3, 0x18,
	0x57, 0x26, 0x13, 0xe8, 0xb3, 0xc5, 0xec, 0xfb, 0x5c, 0xb8, 0xe7, 0xf7, 0xda, 0xe7, 0x9f, 0xfe,
	0xf3, 0x01, 0x7e, 0xb4, 0xff, 0x3a, 0xf2, 0x06, 0xc6, 0x0c, 0x97, 0x28, 0xf7, 0xe8, 0xb6, 0x31,
	0xec, 0xc2, 0xab, 0x49, 0x94, 0xb9, 0xcc, 0xc9, 0x13, 0x72, 0x05, 0xe3, 0x5f, 0x3b, 0xc1, 0x35,
	0xb6, 0xf1, 0xce, 0xb3, 0x83, 0xb0, 0xc7, 0xe6, 0xd7, 0x30, 0xaa, 0xcd, 0x2e, 0x88, 0x9d, 0xdb,
	0x14, 0xc7, 0xd6, 0x97, 0x10, 0x31, 0xe4, 0xe2, 0xa6, 0x32, 0xc4, 0x83, 0xac, 0x46, 0x3f, 0xb6,
	0x64, 0x40, 0x6e, 0x7f, 0xce, 0xbf, 0x9e, 0x8c, 0x7a, 0x5d, 0xfb, 0x4f, 0xc7, 0x7d, 0x0b, 0xcf,
	0xba, 0x1f, 0x3c, 0x8e, 0xfc, 0x0a, 0x46, 0x35, 0xcf, 0x83, 0xd8, 0x77, 0x7d, 0xfb, 0x51, 0xbf,
	0xfb, 0x1f, 0x00, 0x00, 0xff, 0xff, 0x11, 0x50, 0xfe, 0x28, 0xf3, 0x03, 0x00, 0x00,
}
