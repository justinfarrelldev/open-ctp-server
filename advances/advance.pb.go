// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.1
// source: advances/advance.proto

package advances

import (
	ages "github.com/justinfarrelldev/open-ctp-server/ages"
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

type Advance struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prerequisites []*Advance `protobuf:"bytes,1,rep,name=prerequisites,proto3" json:"prerequisites,omitempty"`
	Cost          *int32     `protobuf:"varint,2,opt,name=cost,proto3,oneof" json:"cost,omitempty"`
	Age           *ages.Age  `protobuf:"bytes,3,opt,name=age,proto3,oneof" json:"age,omitempty"`
}

func (x *Advance) Reset() {
	*x = Advance{}
	if protoimpl.UnsafeEnabled {
		mi := &file_advances_advance_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Advance) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Advance) ProtoMessage() {}

func (x *Advance) ProtoReflect() protoreflect.Message {
	mi := &file_advances_advance_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Advance.ProtoReflect.Descriptor instead.
func (*Advance) Descriptor() ([]byte, []int) {
	return file_advances_advance_proto_rawDescGZIP(), []int{0}
}

func (x *Advance) GetPrerequisites() []*Advance {
	if x != nil {
		return x.Prerequisites
	}
	return nil
}

func (x *Advance) GetCost() int32 {
	if x != nil && x.Cost != nil {
		return *x.Cost
	}
	return 0
}

func (x *Advance) GetAge() *ages.Age {
	if x != nil {
		return x.Age
	}
	return nil
}

var File_advances_advance_proto protoreflect.FileDescriptor

var file_advances_advance_proto_rawDesc = []byte{
	0x0a, 0x16, 0x61, 0x64, 0x76, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x2f, 0x61, 0x64, 0x76, 0x61, 0x6e,
	0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x61, 0x64, 0x76, 0x61, 0x6e, 0x63,
	0x65, 0x73, 0x1a, 0x0e, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x8e, 0x01, 0x0a, 0x07, 0x41, 0x64, 0x76, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x37,
	0x0a, 0x0d, 0x70, 0x72, 0x65, 0x72, 0x65, 0x71, 0x75, 0x69, 0x73, 0x69, 0x74, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x61, 0x64, 0x76, 0x61, 0x6e, 0x63, 0x65, 0x73,
	0x2e, 0x41, 0x64, 0x76, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x0d, 0x70, 0x72, 0x65, 0x72, 0x65, 0x71,
	0x75, 0x69, 0x73, 0x69, 0x74, 0x65, 0x73, 0x12, 0x17, 0x0a, 0x04, 0x63, 0x6f, 0x73, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x48, 0x00, 0x52, 0x04, 0x63, 0x6f, 0x73, 0x74, 0x88, 0x01, 0x01,
	0x12, 0x20, 0x0a, 0x03, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e,
	0x61, 0x67, 0x65, 0x73, 0x2e, 0x41, 0x67, 0x65, 0x48, 0x01, 0x52, 0x03, 0x61, 0x67, 0x65, 0x88,
	0x01, 0x01, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x63, 0x6f, 0x73, 0x74, 0x42, 0x06, 0x0a, 0x04, 0x5f,
	0x61, 0x67, 0x65, 0x42, 0x36, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6a, 0x75, 0x73, 0x74, 0x69, 0x6e, 0x66, 0x61, 0x72, 0x72, 0x65, 0x6c, 0x6c, 0x64,
	0x65, 0x76, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x2d, 0x63, 0x74, 0x70, 0x2d, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2f, 0x61, 0x64, 0x76, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_advances_advance_proto_rawDescOnce sync.Once
	file_advances_advance_proto_rawDescData = file_advances_advance_proto_rawDesc
)

func file_advances_advance_proto_rawDescGZIP() []byte {
	file_advances_advance_proto_rawDescOnce.Do(func() {
		file_advances_advance_proto_rawDescData = protoimpl.X.CompressGZIP(file_advances_advance_proto_rawDescData)
	})
	return file_advances_advance_proto_rawDescData
}

var file_advances_advance_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_advances_advance_proto_goTypes = []interface{}{
	(*Advance)(nil),  // 0: advances.Advance
	(*ages.Age)(nil), // 1: ages.Age
}
var file_advances_advance_proto_depIdxs = []int32{
	0, // 0: advances.Advance.prerequisites:type_name -> advances.Advance
	1, // 1: advances.Advance.age:type_name -> ages.Age
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_advances_advance_proto_init() }
func file_advances_advance_proto_init() {
	if File_advances_advance_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_advances_advance_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Advance); i {
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
	file_advances_advance_proto_msgTypes[0].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_advances_advance_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_advances_advance_proto_goTypes,
		DependencyIndexes: file_advances_advance_proto_depIdxs,
		MessageInfos:      file_advances_advance_proto_msgTypes,
	}.Build()
	File_advances_advance_proto = out.File
	file_advances_advance_proto_rawDesc = nil
	file_advances_advance_proto_goTypes = nil
	file_advances_advance_proto_depIdxs = nil
}
