// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: flyteidl/service/signal.proto

#include "flyteidl/service/signal.pb.h"

#include <algorithm>

#include <google/protobuf/stubs/common.h>
#include <google/protobuf/io/coded_stream.h>
#include <google/protobuf/extension_set.h>
#include <google/protobuf/wire_format_lite_inl.h>
#include <google/protobuf/descriptor.h>
#include <google/protobuf/generated_message_reflection.h>
#include <google/protobuf/reflection_ops.h>
#include <google/protobuf/wire_format.h>
// @@protoc_insertion_point(includes)
#include <google/protobuf/port_def.inc>

namespace flyteidl {
namespace service {
}  // namespace service
}  // namespace flyteidl
void InitDefaults_flyteidl_2fservice_2fsignal_2eproto() {
}

constexpr ::google::protobuf::Metadata* file_level_metadata_flyteidl_2fservice_2fsignal_2eproto = nullptr;
constexpr ::google::protobuf::EnumDescriptor const** file_level_enum_descriptors_flyteidl_2fservice_2fsignal_2eproto = nullptr;
constexpr ::google::protobuf::ServiceDescriptor const** file_level_service_descriptors_flyteidl_2fservice_2fsignal_2eproto = nullptr;
const ::google::protobuf::uint32 TableStruct_flyteidl_2fservice_2fsignal_2eproto::offsets[1] = {};
static constexpr ::google::protobuf::internal::MigrationSchema* schemas = nullptr;
static constexpr ::google::protobuf::Message* const* file_default_instances = nullptr;

::google::protobuf::internal::AssignDescriptorsTable assign_descriptors_table_flyteidl_2fservice_2fsignal_2eproto = {
  {}, AddDescriptors_flyteidl_2fservice_2fsignal_2eproto, "flyteidl/service/signal.proto", schemas,
  file_default_instances, TableStruct_flyteidl_2fservice_2fsignal_2eproto::offsets,
  file_level_metadata_flyteidl_2fservice_2fsignal_2eproto, 0, file_level_enum_descriptors_flyteidl_2fservice_2fsignal_2eproto, file_level_service_descriptors_flyteidl_2fservice_2fsignal_2eproto,
};

const char descriptor_table_protodef_flyteidl_2fservice_2fsignal_2eproto[] =
  "\n\035flyteidl/service/signal.proto\022\020flyteid"
  "l.service\032\034google/api/annotations.proto\032"
  "\033flyteidl/admin/signal.proto2\232\003\n\rSignalS"
  "ervice\022W\n\021GetOrCreateSignal\022(.flyteidl.a"
  "dmin.SignalGetOrCreateRequest\032\026.flyteidl"
  ".admin.Signal\"\000\022\301\001\n\013ListSignals\022!.flytei"
  "dl.admin.SignalListRequest\032\032.flyteidl.ad"
  "min.SignalList\"s\202\323\344\223\002m\022k/api/v1/signals/"
  "{workflow_execution_id.project}/{workflo"
  "w_execution_id.domain}/{workflow_executi"
  "on_id.name}\022l\n\tSetSignal\022 .flyteidl.admi"
  "n.SignalSetRequest\032!.flyteidl.admin.Sign"
  "alSetResponse\"\032\202\323\344\223\002\024\"\017/api/v1/signals:\001"
  "*B\?Z=github.com/flyteorg/flyte/flyteidl/"
  "gen/pb-go/flyteidl/serviceb\006proto3"
  ;
::google::protobuf::internal::DescriptorTable descriptor_table_flyteidl_2fservice_2fsignal_2eproto = {
  false, InitDefaults_flyteidl_2fservice_2fsignal_2eproto, 
  descriptor_table_protodef_flyteidl_2fservice_2fsignal_2eproto,
  "flyteidl/service/signal.proto", &assign_descriptors_table_flyteidl_2fservice_2fsignal_2eproto, 594,
};

void AddDescriptors_flyteidl_2fservice_2fsignal_2eproto() {
  static constexpr ::google::protobuf::internal::InitFunc deps[2] =
  {
    ::AddDescriptors_google_2fapi_2fannotations_2eproto,
    ::AddDescriptors_flyteidl_2fadmin_2fsignal_2eproto,
  };
 ::google::protobuf::internal::AddDescriptors(&descriptor_table_flyteidl_2fservice_2fsignal_2eproto, deps, 2);
}

// Force running AddDescriptors() at dynamic initialization time.
static bool dynamic_init_dummy_flyteidl_2fservice_2fsignal_2eproto = []() { AddDescriptors_flyteidl_2fservice_2fsignal_2eproto(); return true; }();
namespace flyteidl {
namespace service {

// @@protoc_insertion_point(namespace_scope)
}  // namespace service
}  // namespace flyteidl
namespace google {
namespace protobuf {
}  // namespace protobuf
}  // namespace google

// @@protoc_insertion_point(global_scope)
#include <google/protobuf/port_undef.inc>
