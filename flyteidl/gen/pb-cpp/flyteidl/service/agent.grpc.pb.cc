// Generated by the gRPC C++ plugin.
// If you make any local change, they will be lost.
// source: flyteidl/service/agent.proto

#include "flyteidl/service/agent.pb.h"
#include "flyteidl/service/agent.grpc.pb.h"

#include <functional>
#include <grpcpp/impl/codegen/async_stream.h>
#include <grpcpp/impl/codegen/async_unary_call.h>
#include <grpcpp/impl/codegen/channel_interface.h>
#include <grpcpp/impl/codegen/client_unary_call.h>
#include <grpcpp/impl/codegen/client_callback.h>
#include <grpcpp/impl/codegen/method_handler_impl.h>
#include <grpcpp/impl/codegen/rpc_service_method.h>
#include <grpcpp/impl/codegen/server_callback.h>
#include <grpcpp/impl/codegen/service_type.h>
#include <grpcpp/impl/codegen/sync_stream.h>
namespace flyteidl {
namespace service {

static const char* AgentService_method_names[] = {
  "/flyteidl.service.AgentService/CreateTask",
  "/flyteidl.service.AgentService/GetTask",
  "/flyteidl.service.AgentService/DeleteTask",
};

std::unique_ptr< AgentService::Stub> AgentService::NewStub(const std::shared_ptr< ::grpc::ChannelInterface>& channel, const ::grpc::StubOptions& options) {
  (void)options;
  std::unique_ptr< AgentService::Stub> stub(new AgentService::Stub(channel));
  return stub;
}

AgentService::Stub::Stub(const std::shared_ptr< ::grpc::ChannelInterface>& channel)
  : channel_(channel), rpcmethod_CreateTask_(AgentService_method_names[0], ::grpc::internal::RpcMethod::NORMAL_RPC, channel)
  , rpcmethod_GetTask_(AgentService_method_names[1], ::grpc::internal::RpcMethod::NORMAL_RPC, channel)
  , rpcmethod_DeleteTask_(AgentService_method_names[2], ::grpc::internal::RpcMethod::NORMAL_RPC, channel)
  {}

::grpc::Status AgentService::Stub::CreateTask(::grpc::ClientContext* context, const ::flyteidl::admin::CreateTaskRequest& request, ::flyteidl::admin::CreateTaskResponse* response) {
  return ::grpc::internal::BlockingUnaryCall(channel_.get(), rpcmethod_CreateTask_, context, request, response);
}

void AgentService::Stub::experimental_async::CreateTask(::grpc::ClientContext* context, const ::flyteidl::admin::CreateTaskRequest* request, ::flyteidl::admin::CreateTaskResponse* response, std::function<void(::grpc::Status)> f) {
  ::grpc::internal::CallbackUnaryCall(stub_->channel_.get(), stub_->rpcmethod_CreateTask_, context, request, response, std::move(f));
}

void AgentService::Stub::experimental_async::CreateTask(::grpc::ClientContext* context, const ::grpc::ByteBuffer* request, ::flyteidl::admin::CreateTaskResponse* response, std::function<void(::grpc::Status)> f) {
  ::grpc::internal::CallbackUnaryCall(stub_->channel_.get(), stub_->rpcmethod_CreateTask_, context, request, response, std::move(f));
}

void AgentService::Stub::experimental_async::CreateTask(::grpc::ClientContext* context, const ::flyteidl::admin::CreateTaskRequest* request, ::flyteidl::admin::CreateTaskResponse* response, ::grpc::experimental::ClientUnaryReactor* reactor) {
  ::grpc::internal::ClientCallbackUnaryFactory::Create(stub_->channel_.get(), stub_->rpcmethod_CreateTask_, context, request, response, reactor);
}

void AgentService::Stub::experimental_async::CreateTask(::grpc::ClientContext* context, const ::grpc::ByteBuffer* request, ::flyteidl::admin::CreateTaskResponse* response, ::grpc::experimental::ClientUnaryReactor* reactor) {
  ::grpc::internal::ClientCallbackUnaryFactory::Create(stub_->channel_.get(), stub_->rpcmethod_CreateTask_, context, request, response, reactor);
}

::grpc::ClientAsyncResponseReader< ::flyteidl::admin::CreateTaskResponse>* AgentService::Stub::AsyncCreateTaskRaw(::grpc::ClientContext* context, const ::flyteidl::admin::CreateTaskRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::flyteidl::admin::CreateTaskResponse>::Create(channel_.get(), cq, rpcmethod_CreateTask_, context, request, true);
}

::grpc::ClientAsyncResponseReader< ::flyteidl::admin::CreateTaskResponse>* AgentService::Stub::PrepareAsyncCreateTaskRaw(::grpc::ClientContext* context, const ::flyteidl::admin::CreateTaskRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::flyteidl::admin::CreateTaskResponse>::Create(channel_.get(), cq, rpcmethod_CreateTask_, context, request, false);
}

::grpc::Status AgentService::Stub::GetTask(::grpc::ClientContext* context, const ::flyteidl::admin::GetTaskRequest& request, ::flyteidl::admin::GetTaskResponse* response) {
  return ::grpc::internal::BlockingUnaryCall(channel_.get(), rpcmethod_GetTask_, context, request, response);
}

void AgentService::Stub::experimental_async::GetTask(::grpc::ClientContext* context, const ::flyteidl::admin::GetTaskRequest* request, ::flyteidl::admin::GetTaskResponse* response, std::function<void(::grpc::Status)> f) {
  ::grpc::internal::CallbackUnaryCall(stub_->channel_.get(), stub_->rpcmethod_GetTask_, context, request, response, std::move(f));
}

void AgentService::Stub::experimental_async::GetTask(::grpc::ClientContext* context, const ::grpc::ByteBuffer* request, ::flyteidl::admin::GetTaskResponse* response, std::function<void(::grpc::Status)> f) {
  ::grpc::internal::CallbackUnaryCall(stub_->channel_.get(), stub_->rpcmethod_GetTask_, context, request, response, std::move(f));
}

void AgentService::Stub::experimental_async::GetTask(::grpc::ClientContext* context, const ::flyteidl::admin::GetTaskRequest* request, ::flyteidl::admin::GetTaskResponse* response, ::grpc::experimental::ClientUnaryReactor* reactor) {
  ::grpc::internal::ClientCallbackUnaryFactory::Create(stub_->channel_.get(), stub_->rpcmethod_GetTask_, context, request, response, reactor);
}

void AgentService::Stub::experimental_async::GetTask(::grpc::ClientContext* context, const ::grpc::ByteBuffer* request, ::flyteidl::admin::GetTaskResponse* response, ::grpc::experimental::ClientUnaryReactor* reactor) {
  ::grpc::internal::ClientCallbackUnaryFactory::Create(stub_->channel_.get(), stub_->rpcmethod_GetTask_, context, request, response, reactor);
}

::grpc::ClientAsyncResponseReader< ::flyteidl::admin::GetTaskResponse>* AgentService::Stub::AsyncGetTaskRaw(::grpc::ClientContext* context, const ::flyteidl::admin::GetTaskRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::flyteidl::admin::GetTaskResponse>::Create(channel_.get(), cq, rpcmethod_GetTask_, context, request, true);
}

::grpc::ClientAsyncResponseReader< ::flyteidl::admin::GetTaskResponse>* AgentService::Stub::PrepareAsyncGetTaskRaw(::grpc::ClientContext* context, const ::flyteidl::admin::GetTaskRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::flyteidl::admin::GetTaskResponse>::Create(channel_.get(), cq, rpcmethod_GetTask_, context, request, false);
}

::grpc::Status AgentService::Stub::DeleteTask(::grpc::ClientContext* context, const ::flyteidl::admin::DeleteTaskRequest& request, ::flyteidl::admin::DeleteTaskResponse* response) {
  return ::grpc::internal::BlockingUnaryCall(channel_.get(), rpcmethod_DeleteTask_, context, request, response);
}

void AgentService::Stub::experimental_async::DeleteTask(::grpc::ClientContext* context, const ::flyteidl::admin::DeleteTaskRequest* request, ::flyteidl::admin::DeleteTaskResponse* response, std::function<void(::grpc::Status)> f) {
  ::grpc::internal::CallbackUnaryCall(stub_->channel_.get(), stub_->rpcmethod_DeleteTask_, context, request, response, std::move(f));
}

void AgentService::Stub::experimental_async::DeleteTask(::grpc::ClientContext* context, const ::grpc::ByteBuffer* request, ::flyteidl::admin::DeleteTaskResponse* response, std::function<void(::grpc::Status)> f) {
  ::grpc::internal::CallbackUnaryCall(stub_->channel_.get(), stub_->rpcmethod_DeleteTask_, context, request, response, std::move(f));
}

void AgentService::Stub::experimental_async::DeleteTask(::grpc::ClientContext* context, const ::flyteidl::admin::DeleteTaskRequest* request, ::flyteidl::admin::DeleteTaskResponse* response, ::grpc::experimental::ClientUnaryReactor* reactor) {
  ::grpc::internal::ClientCallbackUnaryFactory::Create(stub_->channel_.get(), stub_->rpcmethod_DeleteTask_, context, request, response, reactor);
}

void AgentService::Stub::experimental_async::DeleteTask(::grpc::ClientContext* context, const ::grpc::ByteBuffer* request, ::flyteidl::admin::DeleteTaskResponse* response, ::grpc::experimental::ClientUnaryReactor* reactor) {
  ::grpc::internal::ClientCallbackUnaryFactory::Create(stub_->channel_.get(), stub_->rpcmethod_DeleteTask_, context, request, response, reactor);
}

::grpc::ClientAsyncResponseReader< ::flyteidl::admin::DeleteTaskResponse>* AgentService::Stub::AsyncDeleteTaskRaw(::grpc::ClientContext* context, const ::flyteidl::admin::DeleteTaskRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::flyteidl::admin::DeleteTaskResponse>::Create(channel_.get(), cq, rpcmethod_DeleteTask_, context, request, true);
}

::grpc::ClientAsyncResponseReader< ::flyteidl::admin::DeleteTaskResponse>* AgentService::Stub::PrepareAsyncDeleteTaskRaw(::grpc::ClientContext* context, const ::flyteidl::admin::DeleteTaskRequest& request, ::grpc::CompletionQueue* cq) {
  return ::grpc::internal::ClientAsyncResponseReaderFactory< ::flyteidl::admin::DeleteTaskResponse>::Create(channel_.get(), cq, rpcmethod_DeleteTask_, context, request, false);
}

AgentService::Service::Service() {
  AddMethod(new ::grpc::internal::RpcServiceMethod(
      AgentService_method_names[0],
      ::grpc::internal::RpcMethod::NORMAL_RPC,
      new ::grpc::internal::RpcMethodHandler< AgentService::Service, ::flyteidl::admin::CreateTaskRequest, ::flyteidl::admin::CreateTaskResponse>(
          std::mem_fn(&AgentService::Service::CreateTask), this)));
  AddMethod(new ::grpc::internal::RpcServiceMethod(
      AgentService_method_names[1],
      ::grpc::internal::RpcMethod::NORMAL_RPC,
      new ::grpc::internal::RpcMethodHandler< AgentService::Service, ::flyteidl::admin::GetTaskRequest, ::flyteidl::admin::GetTaskResponse>(
          std::mem_fn(&AgentService::Service::GetTask), this)));
  AddMethod(new ::grpc::internal::RpcServiceMethod(
      AgentService_method_names[2],
      ::grpc::internal::RpcMethod::NORMAL_RPC,
      new ::grpc::internal::RpcMethodHandler< AgentService::Service, ::flyteidl::admin::DeleteTaskRequest, ::flyteidl::admin::DeleteTaskResponse>(
          std::mem_fn(&AgentService::Service::DeleteTask), this)));
}

AgentService::Service::~Service() {
}

::grpc::Status AgentService::Service::CreateTask(::grpc::ServerContext* context, const ::flyteidl::admin::CreateTaskRequest* request, ::flyteidl::admin::CreateTaskResponse* response) {
  (void) context;
  (void) request;
  (void) response;
  return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
}

::grpc::Status AgentService::Service::GetTask(::grpc::ServerContext* context, const ::flyteidl::admin::GetTaskRequest* request, ::flyteidl::admin::GetTaskResponse* response) {
  (void) context;
  (void) request;
  (void) response;
  return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
}

::grpc::Status AgentService::Service::DeleteTask(::grpc::ServerContext* context, const ::flyteidl::admin::DeleteTaskRequest* request, ::flyteidl::admin::DeleteTaskResponse* response) {
  (void) context;
  (void) request;
  (void) response;
  return ::grpc::Status(::grpc::StatusCode::UNIMPLEMENTED, "");
}


}  // namespace flyteidl
}  // namespace service

