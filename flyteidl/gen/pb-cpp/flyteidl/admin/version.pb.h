// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: flyteidl/admin/version.proto

#ifndef PROTOBUF_INCLUDED_flyteidl_2fadmin_2fversion_2eproto
#define PROTOBUF_INCLUDED_flyteidl_2fadmin_2fversion_2eproto

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
#include <google/protobuf/unknown_field_set.h>
// @@protoc_insertion_point(includes)
#include <google/protobuf/port_def.inc>
#define PROTOBUF_INTERNAL_EXPORT_flyteidl_2fadmin_2fversion_2eproto

// Internal implementation detail -- do not use these members.
struct TableStruct_flyteidl_2fadmin_2fversion_2eproto {
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
void AddDescriptors_flyteidl_2fadmin_2fversion_2eproto();
namespace flyteidl {
namespace admin {
class GetVersionRequest;
class GetVersionRequestDefaultTypeInternal;
extern GetVersionRequestDefaultTypeInternal _GetVersionRequest_default_instance_;
class GetVersionResponse;
class GetVersionResponseDefaultTypeInternal;
extern GetVersionResponseDefaultTypeInternal _GetVersionResponse_default_instance_;
class Version;
class VersionDefaultTypeInternal;
extern VersionDefaultTypeInternal _Version_default_instance_;
}  // namespace admin
}  // namespace flyteidl
namespace google {
namespace protobuf {
template<> ::flyteidl::admin::GetVersionRequest* Arena::CreateMaybeMessage<::flyteidl::admin::GetVersionRequest>(Arena*);
template<> ::flyteidl::admin::GetVersionResponse* Arena::CreateMaybeMessage<::flyteidl::admin::GetVersionResponse>(Arena*);
template<> ::flyteidl::admin::Version* Arena::CreateMaybeMessage<::flyteidl::admin::Version>(Arena*);
}  // namespace protobuf
}  // namespace google
namespace flyteidl {
namespace admin {

// ===================================================================

class GetVersionResponse final :
    public ::google::protobuf::Message /* @@protoc_insertion_point(class_definition:flyteidl.admin.GetVersionResponse) */ {
 public:
  GetVersionResponse();
  virtual ~GetVersionResponse();

  GetVersionResponse(const GetVersionResponse& from);

  inline GetVersionResponse& operator=(const GetVersionResponse& from) {
    CopyFrom(from);
    return *this;
  }
  #if LANG_CXX11
  GetVersionResponse(GetVersionResponse&& from) noexcept
    : GetVersionResponse() {
    *this = ::std::move(from);
  }

  inline GetVersionResponse& operator=(GetVersionResponse&& from) noexcept {
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
  static const GetVersionResponse& default_instance();

  static void InitAsDefaultInstance();  // FOR INTERNAL USE ONLY
  static inline const GetVersionResponse* internal_default_instance() {
    return reinterpret_cast<const GetVersionResponse*>(
               &_GetVersionResponse_default_instance_);
  }
  static constexpr int kIndexInFileMessages =
    0;

  void Swap(GetVersionResponse* other);
  friend void swap(GetVersionResponse& a, GetVersionResponse& b) {
    a.Swap(&b);
  }

  // implements Message ----------------------------------------------

  inline GetVersionResponse* New() const final {
    return CreateMaybeMessage<GetVersionResponse>(nullptr);
  }

  GetVersionResponse* New(::google::protobuf::Arena* arena) const final {
    return CreateMaybeMessage<GetVersionResponse>(arena);
  }
  void CopyFrom(const ::google::protobuf::Message& from) final;
  void MergeFrom(const ::google::protobuf::Message& from) final;
  void CopyFrom(const GetVersionResponse& from);
  void MergeFrom(const GetVersionResponse& from);
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
  void InternalSwap(GetVersionResponse* other);
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

  // .flyteidl.admin.Version control_plane_version = 1;
  bool has_control_plane_version() const;
  void clear_control_plane_version();
  static const int kControlPlaneVersionFieldNumber = 1;
  const ::flyteidl::admin::Version& control_plane_version() const;
  ::flyteidl::admin::Version* release_control_plane_version();
  ::flyteidl::admin::Version* mutable_control_plane_version();
  void set_allocated_control_plane_version(::flyteidl::admin::Version* control_plane_version);

  // @@protoc_insertion_point(class_scope:flyteidl.admin.GetVersionResponse)
 private:
  class HasBitSetters;

  ::google::protobuf::internal::InternalMetadataWithArena _internal_metadata_;
  ::flyteidl::admin::Version* control_plane_version_;
  mutable ::google::protobuf::internal::CachedSize _cached_size_;
  friend struct ::TableStruct_flyteidl_2fadmin_2fversion_2eproto;
};
// -------------------------------------------------------------------

class Version final :
    public ::google::protobuf::Message /* @@protoc_insertion_point(class_definition:flyteidl.admin.Version) */ {
 public:
  Version();
  virtual ~Version();

  Version(const Version& from);

  inline Version& operator=(const Version& from) {
    CopyFrom(from);
    return *this;
  }
  #if LANG_CXX11
  Version(Version&& from) noexcept
    : Version() {
    *this = ::std::move(from);
  }

  inline Version& operator=(Version&& from) noexcept {
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
  static const Version& default_instance();

  static void InitAsDefaultInstance();  // FOR INTERNAL USE ONLY
  static inline const Version* internal_default_instance() {
    return reinterpret_cast<const Version*>(
               &_Version_default_instance_);
  }
  static constexpr int kIndexInFileMessages =
    1;

  void Swap(Version* other);
  friend void swap(Version& a, Version& b) {
    a.Swap(&b);
  }

  // implements Message ----------------------------------------------

  inline Version* New() const final {
    return CreateMaybeMessage<Version>(nullptr);
  }

  Version* New(::google::protobuf::Arena* arena) const final {
    return CreateMaybeMessage<Version>(arena);
  }
  void CopyFrom(const ::google::protobuf::Message& from) final;
  void MergeFrom(const ::google::protobuf::Message& from) final;
  void CopyFrom(const Version& from);
  void MergeFrom(const Version& from);
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
  void InternalSwap(Version* other);
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

  // string Build = 1;
  void clear_build();
  static const int kBuildFieldNumber = 1;
  const ::std::string& build() const;
  void set_build(const ::std::string& value);
  #if LANG_CXX11
  void set_build(::std::string&& value);
  #endif
  void set_build(const char* value);
  void set_build(const char* value, size_t size);
  ::std::string* mutable_build();
  ::std::string* release_build();
  void set_allocated_build(::std::string* build);

  // string Version = 2;
  void clear_version();
  static const int kVersionFieldNumber = 2;
  const ::std::string& version() const;
  void set_version(const ::std::string& value);
  #if LANG_CXX11
  void set_version(::std::string&& value);
  #endif
  void set_version(const char* value);
  void set_version(const char* value, size_t size);
  ::std::string* mutable_version();
  ::std::string* release_version();
  void set_allocated_version(::std::string* version);

  // string BuildTime = 3;
  void clear_buildtime();
  static const int kBuildTimeFieldNumber = 3;
  const ::std::string& buildtime() const;
  void set_buildtime(const ::std::string& value);
  #if LANG_CXX11
  void set_buildtime(::std::string&& value);
  #endif
  void set_buildtime(const char* value);
  void set_buildtime(const char* value, size_t size);
  ::std::string* mutable_buildtime();
  ::std::string* release_buildtime();
  void set_allocated_buildtime(::std::string* buildtime);

  // @@protoc_insertion_point(class_scope:flyteidl.admin.Version)
 private:
  class HasBitSetters;

  ::google::protobuf::internal::InternalMetadataWithArena _internal_metadata_;
  ::google::protobuf::internal::ArenaStringPtr build_;
  ::google::protobuf::internal::ArenaStringPtr version_;
  ::google::protobuf::internal::ArenaStringPtr buildtime_;
  mutable ::google::protobuf::internal::CachedSize _cached_size_;
  friend struct ::TableStruct_flyteidl_2fadmin_2fversion_2eproto;
};
// -------------------------------------------------------------------

class GetVersionRequest final :
    public ::google::protobuf::Message /* @@protoc_insertion_point(class_definition:flyteidl.admin.GetVersionRequest) */ {
 public:
  GetVersionRequest();
  virtual ~GetVersionRequest();

  GetVersionRequest(const GetVersionRequest& from);

  inline GetVersionRequest& operator=(const GetVersionRequest& from) {
    CopyFrom(from);
    return *this;
  }
  #if LANG_CXX11
  GetVersionRequest(GetVersionRequest&& from) noexcept
    : GetVersionRequest() {
    *this = ::std::move(from);
  }

  inline GetVersionRequest& operator=(GetVersionRequest&& from) noexcept {
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
  static const GetVersionRequest& default_instance();

  static void InitAsDefaultInstance();  // FOR INTERNAL USE ONLY
  static inline const GetVersionRequest* internal_default_instance() {
    return reinterpret_cast<const GetVersionRequest*>(
               &_GetVersionRequest_default_instance_);
  }
  static constexpr int kIndexInFileMessages =
    2;

  void Swap(GetVersionRequest* other);
  friend void swap(GetVersionRequest& a, GetVersionRequest& b) {
    a.Swap(&b);
  }

  // implements Message ----------------------------------------------

  inline GetVersionRequest* New() const final {
    return CreateMaybeMessage<GetVersionRequest>(nullptr);
  }

  GetVersionRequest* New(::google::protobuf::Arena* arena) const final {
    return CreateMaybeMessage<GetVersionRequest>(arena);
  }
  void CopyFrom(const ::google::protobuf::Message& from) final;
  void MergeFrom(const ::google::protobuf::Message& from) final;
  void CopyFrom(const GetVersionRequest& from);
  void MergeFrom(const GetVersionRequest& from);
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
  void InternalSwap(GetVersionRequest* other);
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

  // @@protoc_insertion_point(class_scope:flyteidl.admin.GetVersionRequest)
 private:
  class HasBitSetters;

  ::google::protobuf::internal::InternalMetadataWithArena _internal_metadata_;
  mutable ::google::protobuf::internal::CachedSize _cached_size_;
  friend struct ::TableStruct_flyteidl_2fadmin_2fversion_2eproto;
};
// ===================================================================


// ===================================================================

#ifdef __GNUC__
  #pragma GCC diagnostic push
  #pragma GCC diagnostic ignored "-Wstrict-aliasing"
#endif  // __GNUC__
// GetVersionResponse

// .flyteidl.admin.Version control_plane_version = 1;
inline bool GetVersionResponse::has_control_plane_version() const {
  return this != internal_default_instance() && control_plane_version_ != nullptr;
}
inline void GetVersionResponse::clear_control_plane_version() {
  if (GetArenaNoVirtual() == nullptr && control_plane_version_ != nullptr) {
    delete control_plane_version_;
  }
  control_plane_version_ = nullptr;
}
inline const ::flyteidl::admin::Version& GetVersionResponse::control_plane_version() const {
  const ::flyteidl::admin::Version* p = control_plane_version_;
  // @@protoc_insertion_point(field_get:flyteidl.admin.GetVersionResponse.control_plane_version)
  return p != nullptr ? *p : *reinterpret_cast<const ::flyteidl::admin::Version*>(
      &::flyteidl::admin::_Version_default_instance_);
}
inline ::flyteidl::admin::Version* GetVersionResponse::release_control_plane_version() {
  // @@protoc_insertion_point(field_release:flyteidl.admin.GetVersionResponse.control_plane_version)
  
  ::flyteidl::admin::Version* temp = control_plane_version_;
  control_plane_version_ = nullptr;
  return temp;
}
inline ::flyteidl::admin::Version* GetVersionResponse::mutable_control_plane_version() {
  
  if (control_plane_version_ == nullptr) {
    auto* p = CreateMaybeMessage<::flyteidl::admin::Version>(GetArenaNoVirtual());
    control_plane_version_ = p;
  }
  // @@protoc_insertion_point(field_mutable:flyteidl.admin.GetVersionResponse.control_plane_version)
  return control_plane_version_;
}
inline void GetVersionResponse::set_allocated_control_plane_version(::flyteidl::admin::Version* control_plane_version) {
  ::google::protobuf::Arena* message_arena = GetArenaNoVirtual();
  if (message_arena == nullptr) {
    delete control_plane_version_;
  }
  if (control_plane_version) {
    ::google::protobuf::Arena* submessage_arena = nullptr;
    if (message_arena != submessage_arena) {
      control_plane_version = ::google::protobuf::internal::GetOwnedMessage(
          message_arena, control_plane_version, submessage_arena);
    }
    
  } else {
    
  }
  control_plane_version_ = control_plane_version;
  // @@protoc_insertion_point(field_set_allocated:flyteidl.admin.GetVersionResponse.control_plane_version)
}

// -------------------------------------------------------------------

// Version

// string Build = 1;
inline void Version::clear_build() {
  build_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Version::build() const {
  // @@protoc_insertion_point(field_get:flyteidl.admin.Version.Build)
  return build_.GetNoArena();
}
inline void Version::set_build(const ::std::string& value) {
  
  build_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:flyteidl.admin.Version.Build)
}
#if LANG_CXX11
inline void Version::set_build(::std::string&& value) {
  
  build_.SetNoArena(
    &::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::move(value));
  // @@protoc_insertion_point(field_set_rvalue:flyteidl.admin.Version.Build)
}
#endif
inline void Version::set_build(const char* value) {
  GOOGLE_DCHECK(value != nullptr);
  
  build_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:flyteidl.admin.Version.Build)
}
inline void Version::set_build(const char* value, size_t size) {
  
  build_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:flyteidl.admin.Version.Build)
}
inline ::std::string* Version::mutable_build() {
  
  // @@protoc_insertion_point(field_mutable:flyteidl.admin.Version.Build)
  return build_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Version::release_build() {
  // @@protoc_insertion_point(field_release:flyteidl.admin.Version.Build)
  
  return build_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Version::set_allocated_build(::std::string* build) {
  if (build != nullptr) {
    
  } else {
    
  }
  build_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), build);
  // @@protoc_insertion_point(field_set_allocated:flyteidl.admin.Version.Build)
}

// string Version = 2;
inline void Version::clear_version() {
  version_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Version::version() const {
  // @@protoc_insertion_point(field_get:flyteidl.admin.Version.Version)
  return version_.GetNoArena();
}
inline void Version::set_version(const ::std::string& value) {
  
  version_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:flyteidl.admin.Version.Version)
}
#if LANG_CXX11
inline void Version::set_version(::std::string&& value) {
  
  version_.SetNoArena(
    &::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::move(value));
  // @@protoc_insertion_point(field_set_rvalue:flyteidl.admin.Version.Version)
}
#endif
inline void Version::set_version(const char* value) {
  GOOGLE_DCHECK(value != nullptr);
  
  version_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:flyteidl.admin.Version.Version)
}
inline void Version::set_version(const char* value, size_t size) {
  
  version_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:flyteidl.admin.Version.Version)
}
inline ::std::string* Version::mutable_version() {
  
  // @@protoc_insertion_point(field_mutable:flyteidl.admin.Version.Version)
  return version_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Version::release_version() {
  // @@protoc_insertion_point(field_release:flyteidl.admin.Version.Version)
  
  return version_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Version::set_allocated_version(::std::string* version) {
  if (version != nullptr) {
    
  } else {
    
  }
  version_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), version);
  // @@protoc_insertion_point(field_set_allocated:flyteidl.admin.Version.Version)
}

// string BuildTime = 3;
inline void Version::clear_buildtime() {
  buildtime_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Version::buildtime() const {
  // @@protoc_insertion_point(field_get:flyteidl.admin.Version.BuildTime)
  return buildtime_.GetNoArena();
}
inline void Version::set_buildtime(const ::std::string& value) {
  
  buildtime_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:flyteidl.admin.Version.BuildTime)
}
#if LANG_CXX11
inline void Version::set_buildtime(::std::string&& value) {
  
  buildtime_.SetNoArena(
    &::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::move(value));
  // @@protoc_insertion_point(field_set_rvalue:flyteidl.admin.Version.BuildTime)
}
#endif
inline void Version::set_buildtime(const char* value) {
  GOOGLE_DCHECK(value != nullptr);
  
  buildtime_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:flyteidl.admin.Version.BuildTime)
}
inline void Version::set_buildtime(const char* value, size_t size) {
  
  buildtime_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:flyteidl.admin.Version.BuildTime)
}
inline ::std::string* Version::mutable_buildtime() {
  
  // @@protoc_insertion_point(field_mutable:flyteidl.admin.Version.BuildTime)
  return buildtime_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Version::release_buildtime() {
  // @@protoc_insertion_point(field_release:flyteidl.admin.Version.BuildTime)
  
  return buildtime_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Version::set_allocated_buildtime(::std::string* buildtime) {
  if (buildtime != nullptr) {
    
  } else {
    
  }
  buildtime_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), buildtime);
  // @@protoc_insertion_point(field_set_allocated:flyteidl.admin.Version.BuildTime)
}

// -------------------------------------------------------------------

// GetVersionRequest

#ifdef __GNUC__
  #pragma GCC diagnostic pop
#endif  // __GNUC__
// -------------------------------------------------------------------

// -------------------------------------------------------------------


// @@protoc_insertion_point(namespace_scope)

}  // namespace admin
}  // namespace flyteidl

// @@protoc_insertion_point(global_scope)

#include <google/protobuf/port_undef.inc>
#endif  // PROTOBUF_INCLUDED_flyteidl_2fadmin_2fversion_2eproto
