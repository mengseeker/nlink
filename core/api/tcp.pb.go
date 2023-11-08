// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.25.0
// source: core/api/tcp.proto

package api

import (
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

type SockData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *SockData) Reset() {
	*x = SockData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_api_tcp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SockData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SockData) ProtoMessage() {}

func (x *SockData) ProtoReflect() protoreflect.Message {
	mi := &file_core_api_tcp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SockData.ProtoReflect.Descriptor instead.
func (*SockData) Descriptor() ([]byte, []int) {
	return file_core_api_tcp_proto_rawDescGZIP(), []int{0}
}

func (x *SockData) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type SockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Req  *SockRequest_Sock `protobuf:"bytes,1,opt,name=req,proto3,oneof" json:"req,omitempty"`
	Data *SockData         `protobuf:"bytes,2,opt,name=data,proto3,oneof" json:"data,omitempty"`
}

func (x *SockRequest) Reset() {
	*x = SockRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_api_tcp_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SockRequest) ProtoMessage() {}

func (x *SockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_core_api_tcp_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SockRequest.ProtoReflect.Descriptor instead.
func (*SockRequest) Descriptor() ([]byte, []int) {
	return file_core_api_tcp_proto_rawDescGZIP(), []int{1}
}

func (x *SockRequest) GetReq() *SockRequest_Sock {
	if x != nil {
		return x.Req
	}
	return nil
}

func (x *SockRequest) GetData() *SockData {
	if x != nil {
		return x.Data
	}
	return nil
}

type SockRequest_Sock struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Port int32  `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
}

func (x *SockRequest_Sock) Reset() {
	*x = SockRequest_Sock{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_api_tcp_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SockRequest_Sock) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SockRequest_Sock) ProtoMessage() {}

func (x *SockRequest_Sock) ProtoReflect() protoreflect.Message {
	mi := &file_core_api_tcp_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SockRequest_Sock.ProtoReflect.Descriptor instead.
func (*SockRequest_Sock) Descriptor() ([]byte, []int) {
	return file_core_api_tcp_proto_rawDescGZIP(), []int{1, 0}
}

func (x *SockRequest_Sock) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *SockRequest_Sock) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

var File_core_api_tcp_proto protoreflect.FileDescriptor

var file_core_api_tcp_proto_rawDesc = []byte{
	0x0a, 0x12, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x63, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x61, 0x70, 0x69, 0x22, 0x1e, 0x0a, 0x08, 0x53, 0x6f, 0x63,
	0x6b, 0x44, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0xa4, 0x01, 0x0a, 0x0b, 0x53, 0x6f,
	0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x03, 0x72, 0x65, 0x71,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x53, 0x6f, 0x63,
	0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x53, 0x6f, 0x63, 0x6b, 0x48, 0x00, 0x52,
	0x03, 0x72, 0x65, 0x71, 0x88, 0x01, 0x01, 0x12, 0x26, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x53, 0x6f, 0x63, 0x6b,
	0x44, 0x61, 0x74, 0x61, 0x48, 0x01, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x88, 0x01, 0x01, 0x1a,
	0x2e, 0x0a, 0x04, 0x53, 0x6f, 0x63, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70,
	0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x42,
	0x06, 0x0a, 0x04, 0x5f, 0x72, 0x65, 0x71, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x64, 0x61, 0x74, 0x61,
	0x32, 0x3d, 0x0a, 0x09, 0x54, 0x43, 0x50, 0x43, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x12, 0x30, 0x0a,
	0x07, 0x54, 0x43, 0x50, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x10, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x53,
	0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x53, 0x6f, 0x63, 0x6b, 0x44, 0x61, 0x74, 0x61, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42,
	0x26, 0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x65,
	0x6e, 0x67, 0x73, 0x65, 0x65, 0x6b, 0x65, 0x72, 0x2f, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x2f, 0x63,
	0x6f, 0x72, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_core_api_tcp_proto_rawDescOnce sync.Once
	file_core_api_tcp_proto_rawDescData = file_core_api_tcp_proto_rawDesc
)

func file_core_api_tcp_proto_rawDescGZIP() []byte {
	file_core_api_tcp_proto_rawDescOnce.Do(func() {
		file_core_api_tcp_proto_rawDescData = protoimpl.X.CompressGZIP(file_core_api_tcp_proto_rawDescData)
	})
	return file_core_api_tcp_proto_rawDescData
}

var file_core_api_tcp_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_core_api_tcp_proto_goTypes = []interface{}{
	(*SockData)(nil),         // 0: api.SockData
	(*SockRequest)(nil),      // 1: api.SockRequest
	(*SockRequest_Sock)(nil), // 2: api.SockRequest.Sock
}
var file_core_api_tcp_proto_depIdxs = []int32{
	2, // 0: api.SockRequest.req:type_name -> api.SockRequest.Sock
	0, // 1: api.SockRequest.data:type_name -> api.SockData
	1, // 2: api.TCPCaller.TCPCall:input_type -> api.SockRequest
	0, // 3: api.TCPCaller.TCPCall:output_type -> api.SockData
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_core_api_tcp_proto_init() }
func file_core_api_tcp_proto_init() {
	if File_core_api_tcp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_core_api_tcp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SockData); i {
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
		file_core_api_tcp_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SockRequest); i {
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
		file_core_api_tcp_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SockRequest_Sock); i {
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
	file_core_api_tcp_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_core_api_tcp_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_core_api_tcp_proto_goTypes,
		DependencyIndexes: file_core_api_tcp_proto_depIdxs,
		MessageInfos:      file_core_api_tcp_proto_msgTypes,
	}.Build()
	File_core_api_tcp_proto = out.File
	file_core_api_tcp_proto_rawDesc = nil
	file_core_api_tcp_proto_goTypes = nil
	file_core_api_tcp_proto_depIdxs = nil
}
