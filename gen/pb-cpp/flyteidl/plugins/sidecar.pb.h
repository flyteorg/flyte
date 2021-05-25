// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: flyteidl/plugins/sidecar.proto

#ifndef PROTOBUF_INCLUDED_flyteidl_2fplugins_2fsidecar_2eproto
#define PROTOBUF_INCLUDED_flyteidl_2fplugins_2fsidecar_2eproto

#include <limits>
#include <string>

#include <google/protobuf/port_def.inc>
#if PROTOBUF_VERSION < 3007000
#error This file was generated by a newer version of protoc which is
#error incompatible with your Protocol Buffer headers. Please update
#error your headers.
#endif
#if 3007000 < PROTOBUF_MIN_PROTOC_VERSION
#error This file was generated by an older version of protoc which is
#error incompatible with your Protocol Buffer headers. Please
#error regenerate this file with a newer version of protoc.
#endif

#include <google/protobuf/port_undef.inc>
#include <google/protobuf/io/coded_stream.h>
#include <google/protobuf/arena.h>
#include <google/protobuf/arenastring.h>
#include <google/protobuf/generated_message_table_driven.h>
#include <google/protobuf/generated_message_util.h>
#include <google/protobuf/inlined_string_field.h>
#include <google/protobuf/metadata.h>
#include <google/protobuf/message.h>
#include <google/protobuf/repeated_field.h>  // IWYU pragma: export
#include <google/protobuf/extension_set.h>  // IWYU pragma: export
#include <google/protobuf/map.h>  // IWYU pragma: export
#include <google/protobuf/map_entry.h>
#include <google/protobuf/map_field_inl.h>
#include <google/protobuf/unknown_field_set.h>
#include "k8s.io/api/core/v1/generated.pb.h"
// @@protoc_insertion_point(includes)
#include <google/protobuf/port_def.inc>
#define PROTOBUF_INTERNAL_EXPORT_flyteidl_2fplugins_2fsidecar_2eproto

// Internal implementation detail -- do not use these members.
struct TableStruct_flyteidl_2fplugins_2fsidecar_2eproto {
  static const ::google::protobuf::internal::ParseTableField entries[]
    PROTOBUF_SECTION_VARIABLE(protodesc_cold);
  static const ::google::protobuf::internal::AuxillaryParseTableField aux[]
    PROTOBUF_SECTION_VARIABLE(protodesc_cold);
  static const ::google::protobuf::internal::ParseTable schema[3]
    PROTOBUF_SECTION_VARIABLE(protodesc_cold);
  static const ::google::protobuf::internal::FieldMetadata field_metadata[];
  static const ::google::protobuf::internal::SerializationTable serialization_table[];
  static const ::google::protobuf::uint32 offsets[];
};
void AddDescriptors_flyteidl_2fplugins_2fsidecar_2eproto();
namespace flyteidl {
namespace plugins {
class SidecarJob;
class SidecarJobDefaultTypeInternal;
extern SidecarJobDefaultTypeInternal _SidecarJob_default_instance_;
class SidecarJob_AnnotationsEntry_DoNotUse;
class SidecarJob_AnnotationsEntry_DoNotUseDefaultTypeInternal;
extern SidecarJob_AnnotationsEntry_DoNotUseDefaultTypeInternal _SidecarJob_AnnotationsEntry_DoNotUse_default_instance_;
class SidecarJob_LabelsEntry_DoNotUse;
class SidecarJob_LabelsEntry_DoNotUseDefaultTypeInternal;
extern SidecarJob_LabelsEntry_DoNotUseDefaultTypeInternal _SidecarJob_LabelsEntry_DoNotUse_default_instance_;
}  // namespace plugins
}  // namespace flyteidl
namespace google {
namespace protobuf {
template<> ::flyteidl::plugins::SidecarJob* Arena::CreateMaybeMessage<::flyteidl::plugins::SidecarJob>(Arena*);
template<> ::flyteidl::plugins::SidecarJob_AnnotationsEntry_DoNotUse* Arena::CreateMaybeMessage<::flyteidl::plugins::SidecarJob_AnnotationsEntry_DoNotUse>(Arena*);
template<> ::flyteidl::plugins::SidecarJob_LabelsEntry_DoNotUse* Arena::CreateMaybeMessage<::flyteidl::plugins::SidecarJob_LabelsEntry_DoNotUse>(Arena*);
}  // namespace protobuf
}  // namespace google
namespace flyteidl {
namespace plugins {

// ===================================================================

class SidecarJob_AnnotationsEntry_DoNotUse : public ::google::protobuf::internal::MapEntry<SidecarJob_AnnotationsEntry_DoNotUse, 
    ::std::string, ::std::string,
    ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
    ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
    0 > {
public:
#if GOOGLE_PROTOBUF_ENABLE_EXPERIMENTAL_PARSER
static bool _ParseMap(const char* begin, const char* end, void* object, ::google::protobuf::internal::ParseContext* ctx);
#endif  // GOOGLE_PROTOBUF_ENABLE_EXPERIMENTAL_PARSER
  typedef ::google::protobuf::internal::MapEntry<SidecarJob_AnnotationsEntry_DoNotUse, 
    ::std::string, ::std::string,
    ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
    ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
    0 > SuperType;
  SidecarJob_AnnotationsEntry_DoNotUse();
  SidecarJob_AnnotationsEntry_DoNotUse(::google::protobuf::Arena* arena);
  void MergeFrom(const SidecarJob_AnnotationsEntry_DoNotUse& other);
  static const SidecarJob_AnnotationsEntry_DoNotUse* internal_default_instance() { return reinterpret_cast<const SidecarJob_AnnotationsEntry_DoNotUse*>(&_SidecarJob_AnnotationsEntry_DoNotUse_default_instance_); }
  void MergeFrom(const ::google::protobuf::Message& other) final;
  ::google::protobuf::Metadata GetMetadata() const;
};

// -------------------------------------------------------------------

class SidecarJob_LabelsEntry_DoNotUse : public ::google::protobuf::internal::MapEntry<SidecarJob_LabelsEntry_DoNotUse, 
    ::std::string, ::std::string,
    ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
    ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
    0 > {
public:
#if GOOGLE_PROTOBUF_ENABLE_EXPERIMENTAL_PARSER
static bool _ParseMap(const char* begin, const char* end, void* object, ::google::protobuf::internal::ParseContext* ctx);
#endif  // GOOGLE_PROTOBUF_ENABLE_EXPERIMENTAL_PARSER
  typedef ::google::protobuf::internal::MapEntry<SidecarJob_LabelsEntry_DoNotUse, 
    ::std::string, ::std::string,
    ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
    ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
    0 > SuperType;
  SidecarJob_LabelsEntry_DoNotUse();
  SidecarJob_LabelsEntry_DoNotUse(::google::protobuf::Arena* arena);
  void MergeFrom(const SidecarJob_LabelsEntry_DoNotUse& other);
  static const SidecarJob_LabelsEntry_DoNotUse* internal_default_instance() { return reinterpret_cast<const SidecarJob_LabelsEntry_DoNotUse*>(&_SidecarJob_LabelsEntry_DoNotUse_default_instance_); }
  void MergeFrom(const ::google::protobuf::Message& other) final;
  ::google::protobuf::Metadata GetMetadata() const;
};

// -------------------------------------------------------------------

class SidecarJob final :
    public ::google::protobuf::Message /* @@protoc_insertion_point(class_definition:flyteidl.plugins.SidecarJob) */ {
 public:
  SidecarJob();
  virtual ~SidecarJob();

  SidecarJob(const SidecarJob& from);

  inline SidecarJob& operator=(const SidecarJob& from) {
    CopyFrom(from);
    return *this;
  }
  #if LANG_CXX11
  SidecarJob(SidecarJob&& from) noexcept
    : SidecarJob() {
    *this = ::std::move(from);
  }

  inline SidecarJob& operator=(SidecarJob&& from) noexcept {
    if (GetArenaNoVirtual() == from.GetArenaNoVirtual()) {
      if (this != &from) InternalSwap(&from);
    } else {
      CopyFrom(from);
    }
    return *this;
  }
  #endif
  static const ::google::protobuf::Descriptor* descriptor() {
    return default_instance().GetDescriptor();
  }
  static const SidecarJob& default_instance();

  static void InitAsDefaultInstance();  // FOR INTERNAL USE ONLY
  static inline const SidecarJob* internal_default_instance() {
    return reinterpret_cast<const SidecarJob*>(
               &_SidecarJob_default_instance_);
  }
  static constexpr int kIndexInFileMessages =
    2;

  void Swap(SidecarJob* other);
  friend void swap(SidecarJob& a, SidecarJob& b) {
    a.Swap(&b);
  }

  // implements Message ----------------------------------------------

  inline SidecarJob* New() const final {
    return CreateMaybeMessage<SidecarJob>(nullptr);
  }

  SidecarJob* New(::google::protobuf::Arena* arena) const final {
    return CreateMaybeMessage<SidecarJob>(arena);
  }
  void CopyFrom(const ::google::protobuf::Message& from) final;
  void MergeFrom(const ::google::protobuf::Message& from) final;
  void CopyFrom(const SidecarJob& from);
  void MergeFrom(const SidecarJob& from);
  PROTOBUF_ATTRIBUTE_REINITIALIZES void Clear() final;
  bool IsInitialized() const final;

  size_t ByteSizeLong() const final;
  #if GOOGLE_PROTOBUF_ENABLE_EXPERIMENTAL_PARSER
  static const char* _InternalParse(const char* begin, const char* end, void* object, ::google::protobuf::internal::ParseContext* ctx);
  ::google::protobuf::internal::ParseFunc _ParseFunc() const final { return _InternalParse; }
  #else
  bool MergePartialFromCodedStream(
      ::google::protobuf::io::CodedInputStream* input) final;
  #endif  // GOOGLE_PROTOBUF_ENABLE_EXPERIMENTAL_PARSER
  void SerializeWithCachedSizes(
      ::google::protobuf::io::CodedOutputStream* output) const final;
  ::google::protobuf::uint8* InternalSerializeWithCachedSizesToArray(
      ::google::protobuf::uint8* target) const final;
  int GetCachedSize() const final { return _cached_size_.Get(); }

  private:
  void SharedCtor();
  void SharedDtor();
  void SetCachedSize(int size) const final;
  void InternalSwap(SidecarJob* other);
  private:
  inline ::google::protobuf::Arena* GetArenaNoVirtual() const {
    return nullptr;
  }
  inline void* MaybeArenaPtr() const {
    return nullptr;
  }
  public:

  ::google::protobuf::Metadata GetMetadata() const final;

  // nested types ----------------------------------------------------


  // accessors -------------------------------------------------------

  // map<string, string> annotations = 3;
  int annotations_size() const;
  void clear_annotations();
  static const int kAnnotationsFieldNumber = 3;
  const ::google::protobuf::Map< ::std::string, ::std::string >&
      annotations() const;
  ::google::protobuf::Map< ::std::string, ::std::string >*
      mutable_annotations();

  // map<string, string> labels = 4;
  int labels_size() const;
  void clear_labels();
  static const int kLabelsFieldNumber = 4;
  const ::google::protobuf::Map< ::std::string, ::std::string >&
      labels() const;
  ::google::protobuf::Map< ::std::string, ::std::string >*
      mutable_labels();

  // string primary_container_name = 2;
  void clear_primary_container_name();
  static const int kPrimaryContainerNameFieldNumber = 2;
  const ::std::string& primary_container_name() const;
  void set_primary_container_name(const ::std::string& value);
  #if LANG_CXX11
  void set_primary_container_name(::std::string&& value);
  #endif
  void set_primary_container_name(const char* value);
  void set_primary_container_name(const char* value, size_t size);
  ::std::string* mutable_primary_container_name();
  ::std::string* release_primary_container_name();
  void set_allocated_primary_container_name(::std::string* primary_container_name);

  // .k8s.io.api.core.v1.PodSpec pod_spec = 1;
  bool has_pod_spec() const;
  void clear_pod_spec();
  static const int kPodSpecFieldNumber = 1;
  const ::k8s::io::api::core::v1::PodSpec& pod_spec() const;
  ::k8s::io::api::core::v1::PodSpec* release_pod_spec();
  ::k8s::io::api::core::v1::PodSpec* mutable_pod_spec();
  void set_allocated_pod_spec(::k8s::io::api::core::v1::PodSpec* pod_spec);

  // @@protoc_insertion_point(class_scope:flyteidl.plugins.SidecarJob)
 private:
  class HasBitSetters;

  ::google::protobuf::internal::InternalMetadataWithArena _internal_metadata_;
  ::google::protobuf::internal::MapField<
      SidecarJob_AnnotationsEntry_DoNotUse,
      ::std::string, ::std::string,
      ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
      ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
      0 > annotations_;
  ::google::protobuf::internal::MapField<
      SidecarJob_LabelsEntry_DoNotUse,
      ::std::string, ::std::string,
      ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
      ::google::protobuf::internal::WireFormatLite::TYPE_STRING,
      0 > labels_;
  ::google::protobuf::internal::ArenaStringPtr primary_container_name_;
  ::k8s::io::api::core::v1::PodSpec* pod_spec_;
  mutable ::google::protobuf::internal::CachedSize _cached_size_;
  friend struct ::TableStruct_flyteidl_2fplugins_2fsidecar_2eproto;
};
// ===================================================================


// ===================================================================

#ifdef __GNUC__
  #pragma GCC diagnostic push
  #pragma GCC diagnostic ignored "-Wstrict-aliasing"
#endif  // __GNUC__
// -------------------------------------------------------------------

// -------------------------------------------------------------------

// SidecarJob

// .k8s.io.api.core.v1.PodSpec pod_spec = 1;
inline bool SidecarJob::has_pod_spec() const {
  return this != internal_default_instance() && pod_spec_ != nullptr;
}
inline const ::k8s::io::api::core::v1::PodSpec& SidecarJob::pod_spec() const {
  const ::k8s::io::api::core::v1::PodSpec* p = pod_spec_;
  // @@protoc_insertion_point(field_get:flyteidl.plugins.SidecarJob.pod_spec)
  return p != nullptr ? *p : *reinterpret_cast<const ::k8s::io::api::core::v1::PodSpec*>(
      &::k8s::io::api::core::v1::_PodSpec_default_instance_);
}
inline ::k8s::io::api::core::v1::PodSpec* SidecarJob::release_pod_spec() {
  // @@protoc_insertion_point(field_release:flyteidl.plugins.SidecarJob.pod_spec)
  
  ::k8s::io::api::core::v1::PodSpec* temp = pod_spec_;
  pod_spec_ = nullptr;
  return temp;
}
inline ::k8s::io::api::core::v1::PodSpec* SidecarJob::mutable_pod_spec() {
  
  if (pod_spec_ == nullptr) {
    auto* p = CreateMaybeMessage<::k8s::io::api::core::v1::PodSpec>(GetArenaNoVirtual());
    pod_spec_ = p;
  }
  // @@protoc_insertion_point(field_mutable:flyteidl.plugins.SidecarJob.pod_spec)
  return pod_spec_;
}
inline void SidecarJob::set_allocated_pod_spec(::k8s::io::api::core::v1::PodSpec* pod_spec) {
  ::google::protobuf::Arena* message_arena = GetArenaNoVirtual();
  if (message_arena == nullptr) {
    delete reinterpret_cast< ::google::protobuf::MessageLite*>(pod_spec_);
  }
  if (pod_spec) {
    ::google::protobuf::Arena* submessage_arena = nullptr;
    if (message_arena != submessage_arena) {
      pod_spec = ::google::protobuf::internal::GetOwnedMessage(
          message_arena, pod_spec, submessage_arena);
    }
    
  } else {
    
  }
  pod_spec_ = pod_spec;
  // @@protoc_insertion_point(field_set_allocated:flyteidl.plugins.SidecarJob.pod_spec)
}

// string primary_container_name = 2;
inline void SidecarJob::clear_primary_container_name() {
  primary_container_name_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& SidecarJob::primary_container_name() const {
  // @@protoc_insertion_point(field_get:flyteidl.plugins.SidecarJob.primary_container_name)
  return primary_container_name_.GetNoArena();
}
inline void SidecarJob::set_primary_container_name(const ::std::string& value) {
  
  primary_container_name_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:flyteidl.plugins.SidecarJob.primary_container_name)
}
#if LANG_CXX11
inline void SidecarJob::set_primary_container_name(::std::string&& value) {
  
  primary_container_name_.SetNoArena(
    &::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::move(value));
  // @@protoc_insertion_point(field_set_rvalue:flyteidl.plugins.SidecarJob.primary_container_name)
}
#endif
inline void SidecarJob::set_primary_container_name(const char* value) {
  GOOGLE_DCHECK(value != nullptr);
  
  primary_container_name_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:flyteidl.plugins.SidecarJob.primary_container_name)
}
inline void SidecarJob::set_primary_container_name(const char* value, size_t size) {
  
  primary_container_name_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:flyteidl.plugins.SidecarJob.primary_container_name)
}
inline ::std::string* SidecarJob::mutable_primary_container_name() {
  
  // @@protoc_insertion_point(field_mutable:flyteidl.plugins.SidecarJob.primary_container_name)
  return primary_container_name_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* SidecarJob::release_primary_container_name() {
  // @@protoc_insertion_point(field_release:flyteidl.plugins.SidecarJob.primary_container_name)
  
  return primary_container_name_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void SidecarJob::set_allocated_primary_container_name(::std::string* primary_container_name) {
  if (primary_container_name != nullptr) {
    
  } else {
    
  }
  primary_container_name_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), primary_container_name);
  // @@protoc_insertion_point(field_set_allocated:flyteidl.plugins.SidecarJob.primary_container_name)
}

// map<string, string> annotations = 3;
inline int SidecarJob::annotations_size() const {
  return annotations_.size();
}
inline void SidecarJob::clear_annotations() {
  annotations_.Clear();
}
inline const ::google::protobuf::Map< ::std::string, ::std::string >&
SidecarJob::annotations() const {
  // @@protoc_insertion_point(field_map:flyteidl.plugins.SidecarJob.annotations)
  return annotations_.GetMap();
}
inline ::google::protobuf::Map< ::std::string, ::std::string >*
SidecarJob::mutable_annotations() {
  // @@protoc_insertion_point(field_mutable_map:flyteidl.plugins.SidecarJob.annotations)
  return annotations_.MutableMap();
}

// map<string, string> labels = 4;
inline int SidecarJob::labels_size() const {
  return labels_.size();
}
inline void SidecarJob::clear_labels() {
  labels_.Clear();
}
inline const ::google::protobuf::Map< ::std::string, ::std::string >&
SidecarJob::labels() const {
  // @@protoc_insertion_point(field_map:flyteidl.plugins.SidecarJob.labels)
  return labels_.GetMap();
}
inline ::google::protobuf::Map< ::std::string, ::std::string >*
SidecarJob::mutable_labels() {
  // @@protoc_insertion_point(field_mutable_map:flyteidl.plugins.SidecarJob.labels)
  return labels_.MutableMap();
}

#ifdef __GNUC__
  #pragma GCC diagnostic pop
#endif  // __GNUC__
// -------------------------------------------------------------------

// -------------------------------------------------------------------


// @@protoc_insertion_point(namespace_scope)

}  // namespace plugins
}  // namespace flyteidl

// @@protoc_insertion_point(global_scope)

#include <google/protobuf/port_undef.inc>
#endif  // PROTOBUF_INCLUDED_flyteidl_2fplugins_2fsidecar_2eproto
