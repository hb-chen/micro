// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: runtime/build/build.proto

package build

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type BuildRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data    []byte   `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Options *Options `protobuf:"bytes,2,opt,name=options,proto3" json:"options,omitempty"`
}

func (x *BuildRequest) Reset() {
	*x = BuildRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_runtime_build_build_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildRequest) ProtoMessage() {}

func (x *BuildRequest) ProtoReflect() protoreflect.Message {
	mi := &file_runtime_build_build_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildRequest.ProtoReflect.Descriptor instead.
func (*BuildRequest) Descriptor() ([]byte, []int) {
	return file_runtime_build_build_proto_rawDescGZIP(), []int{0}
}

func (x *BuildRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *BuildRequest) GetOptions() *Options {
	if x != nil {
		return x.Options
	}
	return nil
}

type Options struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Archive    string `protobuf:"bytes,1,opt,name=archive,proto3" json:"archive,omitempty"`
	Entrypoint string `protobuf:"bytes,2,opt,name=entrypoint,proto3" json:"entrypoint,omitempty"`
}

func (x *Options) Reset() {
	*x = Options{}
	if protoimpl.UnsafeEnabled {
		mi := &file_runtime_build_build_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Options) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Options) ProtoMessage() {}

func (x *Options) ProtoReflect() protoreflect.Message {
	mi := &file_runtime_build_build_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Options.ProtoReflect.Descriptor instead.
func (*Options) Descriptor() ([]byte, []int) {
	return file_runtime_build_build_proto_rawDescGZIP(), []int{1}
}

func (x *Options) GetArchive() string {
	if x != nil {
		return x.Archive
	}
	return ""
}

func (x *Options) GetEntrypoint() string {
	if x != nil {
		return x.Entrypoint
	}
	return ""
}

type Result struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Result) Reset() {
	*x = Result{}
	if protoimpl.UnsafeEnabled {
		mi := &file_runtime_build_build_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Result) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Result) ProtoMessage() {}

func (x *Result) ProtoReflect() protoreflect.Message {
	mi := &file_runtime_build_build_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Result.ProtoReflect.Descriptor instead.
func (*Result) Descriptor() ([]byte, []int) {
	return file_runtime_build_build_proto_rawDescGZIP(), []int{2}
}

func (x *Result) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_runtime_build_build_proto protoreflect.FileDescriptor

var file_runtime_build_build_proto_rawDesc = []byte{
	0x0a, 0x19, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f,
	0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x72, 0x75, 0x6e,
	0x74, 0x69, 0x6d, 0x65, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x22, 0x54, 0x0a, 0x0c, 0x42, 0x75,
	0x69, 0x6c, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x30,
	0x0a, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x16, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x22, 0x43, 0x0a, 0x07, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x61,
	0x72, 0x63, 0x68, 0x69, 0x76, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x72,
	0x63, 0x68, 0x69, 0x76, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65, 0x6e, 0x74, 0x72, 0x79,
	0x70, 0x6f, 0x69, 0x6e, 0x74, 0x22, 0x1c, 0x0a, 0x06, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12,
	0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x32, 0x4a, 0x0a, 0x05, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x12, 0x41, 0x0a, 0x05,
	0x42, 0x75, 0x69, 0x6c, 0x64, 0x12, 0x1b, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e,
	0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x15, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x62, 0x75, 0x69,
	0x6c, 0x64, 0x2e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42,
	0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x69,
	0x63, 0x72, 0x6f, 0x2f, 0x6d, 0x69, 0x63, 0x72, 0x6f, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x3b, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_runtime_build_build_proto_rawDescOnce sync.Once
	file_runtime_build_build_proto_rawDescData = file_runtime_build_build_proto_rawDesc
)

func file_runtime_build_build_proto_rawDescGZIP() []byte {
	file_runtime_build_build_proto_rawDescOnce.Do(func() {
		file_runtime_build_build_proto_rawDescData = protoimpl.X.CompressGZIP(file_runtime_build_build_proto_rawDescData)
	})
	return file_runtime_build_build_proto_rawDescData
}

var file_runtime_build_build_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_runtime_build_build_proto_goTypes = []interface{}{
	(*BuildRequest)(nil), // 0: runtime.build.BuildRequest
	(*Options)(nil),      // 1: runtime.build.Options
	(*Result)(nil),       // 2: runtime.build.Result
}
var file_runtime_build_build_proto_depIdxs = []int32{
	1, // 0: runtime.build.BuildRequest.options:type_name -> runtime.build.Options
	0, // 1: runtime.build.Build.Build:input_type -> runtime.build.BuildRequest
	2, // 2: runtime.build.Build.Build:output_type -> runtime.build.Result
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_runtime_build_build_proto_init() }
func file_runtime_build_build_proto_init() {
	if File_runtime_build_build_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_runtime_build_build_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_runtime_build_build_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Options); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_runtime_build_build_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Result); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_runtime_build_build_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_runtime_build_build_proto_goTypes,
		DependencyIndexes: file_runtime_build_build_proto_depIdxs,
		MessageInfos:      file_runtime_build_build_proto_msgTypes,
	}.Build()
	File_runtime_build_build_proto = out.File
	file_runtime_build_build_proto_rawDesc = nil
	file_runtime_build_build_proto_goTypes = nil
	file_runtime_build_build_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// BuildClient is the client API for Build service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BuildClient interface {
	Build(ctx context.Context, opts ...grpc.CallOption) (Build_BuildClient, error)
}

type buildClient struct {
	cc grpc.ClientConnInterface
}

func NewBuildClient(cc grpc.ClientConnInterface) BuildClient {
	return &buildClient{cc}
}

func (c *buildClient) Build(ctx context.Context, opts ...grpc.CallOption) (Build_BuildClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Build_serviceDesc.Streams[0], "/runtime.build.Build/Build", opts...)
	if err != nil {
		return nil, err
	}
	x := &buildBuildClient{stream}
	return x, nil
}

type Build_BuildClient interface {
	Send(*BuildRequest) error
	Recv() (*Result, error)
	grpc.ClientStream
}

type buildBuildClient struct {
	grpc.ClientStream
}

func (x *buildBuildClient) Send(m *BuildRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *buildBuildClient) Recv() (*Result, error) {
	m := new(Result)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// BuildServer is the server API for Build service.
type BuildServer interface {
	Build(Build_BuildServer) error
}

// UnimplementedBuildServer can be embedded to have forward compatible implementations.
type UnimplementedBuildServer struct {
}

func (*UnimplementedBuildServer) Build(Build_BuildServer) error {
	return status.Errorf(codes.Unimplemented, "method Build not implemented")
}

func RegisterBuildServer(s *grpc.Server, srv BuildServer) {
	s.RegisterService(&_Build_serviceDesc, srv)
}

func _Build_Build_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BuildServer).Build(&buildBuildServer{stream})
}

type Build_BuildServer interface {
	Send(*Result) error
	Recv() (*BuildRequest, error)
	grpc.ServerStream
}

type buildBuildServer struct {
	grpc.ServerStream
}

func (x *buildBuildServer) Send(m *Result) error {
	return x.ServerStream.SendMsg(m)
}

func (x *buildBuildServer) Recv() (*BuildRequest, error) {
	m := new(BuildRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Build_serviceDesc = grpc.ServiceDesc{
	ServiceName: "runtime.build.Build",
	HandlerType: (*BuildServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Build",
			Handler:       _Build_Build_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "runtime/build/build.proto",
}
