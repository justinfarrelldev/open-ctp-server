// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.1
// source: units/unit_category.proto

package units

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

type UnitCategory int32

const (
	UnitCategory_GENERIC  UnitCategory = 0
	UnitCategory_SETTLER  UnitCategory = 1
	UnitCategory_AERIAL   UnitCategory = 2
	UnitCategory_NAVAL    UnitCategory = 3
	UnitCategory_UNDERSEA UnitCategory = 4
	UnitCategory_ATTACK   UnitCategory = 5
	UnitCategory_DEFENSE  UnitCategory = 6
	UnitCategory_RANGED   UnitCategory = 7
	UnitCategory_FLANKER  UnitCategory = 8
	UnitCategory_SPECIAL  UnitCategory = 9
)

// Enum value maps for UnitCategory.
var (
	UnitCategory_name = map[int32]string{
		0: "GENERIC",
		1: "SETTLER",
		2: "AERIAL",
		3: "NAVAL",
		4: "UNDERSEA",
		5: "ATTACK",
		6: "DEFENSE",
		7: "RANGED",
		8: "FLANKER",
		9: "SPECIAL",
	}
	UnitCategory_value = map[string]int32{
		"GENERIC":  0,
		"SETTLER":  1,
		"AERIAL":   2,
		"NAVAL":    3,
		"UNDERSEA": 4,
		"ATTACK":   5,
		"DEFENSE":  6,
		"RANGED":   7,
		"FLANKER":  8,
		"SPECIAL":  9,
	}
)

func (x UnitCategory) Enum() *UnitCategory {
	p := new(UnitCategory)
	*p = x
	return p
}

func (x UnitCategory) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UnitCategory) Descriptor() protoreflect.EnumDescriptor {
	return file_units_unit_category_proto_enumTypes[0].Descriptor()
}

func (UnitCategory) Type() protoreflect.EnumType {
	return &file_units_unit_category_proto_enumTypes[0]
}

func (x UnitCategory) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *UnitCategory) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = UnitCategory(num)
	return nil
}

// Deprecated: Use UnitCategory.Descriptor instead.
func (UnitCategory) EnumDescriptor() ([]byte, []int) {
	return file_units_unit_category_proto_rawDescGZIP(), []int{0}
}

var File_units_unit_category_proto protoreflect.FileDescriptor

var file_units_unit_category_proto_rawDesc = []byte{
	0x0a, 0x19, 0x75, 0x6e, 0x69, 0x74, 0x73, 0x2f, 0x75, 0x6e, 0x69, 0x74, 0x5f, 0x63, 0x61, 0x74,
	0x65, 0x67, 0x6f, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x75, 0x6e, 0x69,
	0x74, 0x73, 0x2a, 0x8c, 0x01, 0x0a, 0x0c, 0x55, 0x6e, 0x69, 0x74, 0x43, 0x61, 0x74, 0x65, 0x67,
	0x6f, 0x72, 0x79, 0x12, 0x0b, 0x0a, 0x07, 0x47, 0x45, 0x4e, 0x45, 0x52, 0x49, 0x43, 0x10, 0x00,
	0x12, 0x0b, 0x0a, 0x07, 0x53, 0x45, 0x54, 0x54, 0x4c, 0x45, 0x52, 0x10, 0x01, 0x12, 0x0a, 0x0a,
	0x06, 0x41, 0x45, 0x52, 0x49, 0x41, 0x4c, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x4e, 0x41, 0x56,
	0x41, 0x4c, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08, 0x55, 0x4e, 0x44, 0x45, 0x52, 0x53, 0x45, 0x41,
	0x10, 0x04, 0x12, 0x0a, 0x0a, 0x06, 0x41, 0x54, 0x54, 0x41, 0x43, 0x4b, 0x10, 0x05, 0x12, 0x0b,
	0x0a, 0x07, 0x44, 0x45, 0x46, 0x45, 0x4e, 0x53, 0x45, 0x10, 0x06, 0x12, 0x0a, 0x0a, 0x06, 0x52,
	0x41, 0x4e, 0x47, 0x45, 0x44, 0x10, 0x07, 0x12, 0x0b, 0x0a, 0x07, 0x46, 0x4c, 0x41, 0x4e, 0x4b,
	0x45, 0x52, 0x10, 0x08, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x50, 0x45, 0x43, 0x49, 0x41, 0x4c, 0x10,
	0x09, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6a, 0x75, 0x73, 0x74, 0x69, 0x6e, 0x66, 0x61, 0x72, 0x72, 0x65, 0x6c, 0x6c, 0x64, 0x65, 0x76,
	0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x2d, 0x63, 0x74, 0x70, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2f, 0x75, 0x6e, 0x69, 0x74, 0x73,
}

var (
	file_units_unit_category_proto_rawDescOnce sync.Once
	file_units_unit_category_proto_rawDescData = file_units_unit_category_proto_rawDesc
)

func file_units_unit_category_proto_rawDescGZIP() []byte {
	file_units_unit_category_proto_rawDescOnce.Do(func() {
		file_units_unit_category_proto_rawDescData = protoimpl.X.CompressGZIP(file_units_unit_category_proto_rawDescData)
	})
	return file_units_unit_category_proto_rawDescData
}

var file_units_unit_category_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_units_unit_category_proto_goTypes = []interface{}{
	(UnitCategory)(0), // 0: units.UnitCategory
}
var file_units_unit_category_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_units_unit_category_proto_init() }
func file_units_unit_category_proto_init() {
	if File_units_unit_category_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_units_unit_category_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_units_unit_category_proto_goTypes,
		DependencyIndexes: file_units_unit_category_proto_depIdxs,
		EnumInfos:         file_units_unit_category_proto_enumTypes,
	}.Build()
	File_units_unit_category_proto = out.File
	file_units_unit_category_proto_rawDesc = nil
	file_units_unit_category_proto_goTypes = nil
	file_units_unit_category_proto_depIdxs = nil
}