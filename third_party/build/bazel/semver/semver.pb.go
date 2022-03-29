// Copyright 2018 The Bazel Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.6.1
// source: build/bazel/semver/semver.proto

package semver

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

// The full version of a given tool.
type SemVer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The major version, e.g 10 for 10.2.3.
	Major int32 `protobuf:"varint,1,opt,name=major,proto3" json:"major,omitempty"`
	// The minor version, e.g. 2 for 10.2.3.
	Minor int32 `protobuf:"varint,2,opt,name=minor,proto3" json:"minor,omitempty"`
	// The patch version, e.g 3 for 10.2.3.
	Patch int32 `protobuf:"varint,3,opt,name=patch,proto3" json:"patch,omitempty"`
	// The pre-release version. Either this field or major/minor/patch fields
	// must be filled. They are mutually exclusive. Pre-release versions are
	// assumed to be earlier than any released versions.
	Prerelease string `protobuf:"bytes,4,opt,name=prerelease,proto3" json:"prerelease,omitempty"`
}

func (x *SemVer) Reset() {
	*x = SemVer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_build_bazel_semver_semver_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SemVer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SemVer) ProtoMessage() {}

func (x *SemVer) ProtoReflect() protoreflect.Message {
	mi := &file_build_bazel_semver_semver_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SemVer.ProtoReflect.Descriptor instead.
func (*SemVer) Descriptor() ([]byte, []int) {
	return file_build_bazel_semver_semver_proto_rawDescGZIP(), []int{0}
}

func (x *SemVer) GetMajor() int32 {
	if x != nil {
		return x.Major
	}
	return 0
}

func (x *SemVer) GetMinor() int32 {
	if x != nil {
		return x.Minor
	}
	return 0
}

func (x *SemVer) GetPatch() int32 {
	if x != nil {
		return x.Patch
	}
	return 0
}

func (x *SemVer) GetPrerelease() string {
	if x != nil {
		return x.Prerelease
	}
	return ""
}

var File_build_bazel_semver_semver_proto protoreflect.FileDescriptor

var file_build_bazel_semver_semver_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f, 0x62, 0x61, 0x7a, 0x65, 0x6c, 0x2f, 0x73, 0x65,
	0x6d, 0x76, 0x65, 0x72, 0x2f, 0x73, 0x65, 0x6d, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x12, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x62, 0x61, 0x7a, 0x65, 0x6c, 0x2e, 0x73,
	0x65, 0x6d, 0x76, 0x65, 0x72, 0x22, 0x6a, 0x0a, 0x06, 0x53, 0x65, 0x6d, 0x56, 0x65, 0x72, 0x12,
	0x14, 0x0a, 0x05, 0x6d, 0x61, 0x6a, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05,
	0x6d, 0x61, 0x6a, 0x6f, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x69, 0x6e, 0x6f, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6d, 0x69, 0x6e, 0x6f, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x70,
	0x61, 0x74, 0x63, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x70, 0x61, 0x74, 0x63,
	0x68, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x65, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x72, 0x65, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73,
	0x65, 0x42, 0x75, 0x0a, 0x12, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x62, 0x61, 0x7a, 0x65, 0x6c,
	0x2e, 0x73, 0x65, 0x6d, 0x76, 0x65, 0x72, 0x42, 0x0b, 0x53, 0x65, 0x6d, 0x76, 0x65, 0x72, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x35, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x72, 0x75, 0x65, 0x6c, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x2f, 0x74,
	0x68, 0x69, 0x72, 0x64, 0x5f, 0x70, 0x61, 0x72, 0x74, 0x79, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x2f, 0x62, 0x61, 0x7a, 0x65, 0x6c, 0x2f, 0x73, 0x65, 0x6d, 0x76, 0x65, 0x72, 0xa2, 0x02, 0x03,
	0x53, 0x4d, 0x56, 0xaa, 0x02, 0x12, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x42, 0x61, 0x7a, 0x65,
	0x6c, 0x2e, 0x53, 0x65, 0x6d, 0x76, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_build_bazel_semver_semver_proto_rawDescOnce sync.Once
	file_build_bazel_semver_semver_proto_rawDescData = file_build_bazel_semver_semver_proto_rawDesc
)

func file_build_bazel_semver_semver_proto_rawDescGZIP() []byte {
	file_build_bazel_semver_semver_proto_rawDescOnce.Do(func() {
		file_build_bazel_semver_semver_proto_rawDescData = protoimpl.X.CompressGZIP(file_build_bazel_semver_semver_proto_rawDescData)
	})
	return file_build_bazel_semver_semver_proto_rawDescData
}

var file_build_bazel_semver_semver_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_build_bazel_semver_semver_proto_goTypes = []interface{}{
	(*SemVer)(nil), // 0: build.bazel.semver.SemVer
}
var file_build_bazel_semver_semver_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_build_bazel_semver_semver_proto_init() }
func file_build_bazel_semver_semver_proto_init() {
	if File_build_bazel_semver_semver_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_build_bazel_semver_semver_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SemVer); i {
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
			RawDescriptor: file_build_bazel_semver_semver_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_build_bazel_semver_semver_proto_goTypes,
		DependencyIndexes: file_build_bazel_semver_semver_proto_depIdxs,
		MessageInfos:      file_build_bazel_semver_semver_proto_msgTypes,
	}.Build()
	File_build_bazel_semver_semver_proto = out.File
	file_build_bazel_semver_semver_proto_rawDesc = nil
	file_build_bazel_semver_semver_proto_goTypes = nil
	file_build_bazel_semver_semver_proto_depIdxs = nil
}
