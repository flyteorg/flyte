// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: flyteidl/service/external_plugin_service.proto

package serviceconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	service "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// ExternalPluginServiceName is the fully-qualified name of the ExternalPluginService service.
	ExternalPluginServiceName = "flyteidl.service.ExternalPluginService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ExternalPluginServiceCreateTaskProcedure is the fully-qualified name of the
	// ExternalPluginService's CreateTask RPC.
	ExternalPluginServiceCreateTaskProcedure = "/flyteidl.service.ExternalPluginService/CreateTask"
	// ExternalPluginServiceGetTaskProcedure is the fully-qualified name of the ExternalPluginService's
	// GetTask RPC.
	ExternalPluginServiceGetTaskProcedure = "/flyteidl.service.ExternalPluginService/GetTask"
	// ExternalPluginServiceDeleteTaskProcedure is the fully-qualified name of the
	// ExternalPluginService's DeleteTask RPC.
	ExternalPluginServiceDeleteTaskProcedure = "/flyteidl.service.ExternalPluginService/DeleteTask"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	externalPluginServiceServiceDescriptor          = service.File_flyteidl_service_external_plugin_service_proto.Services().ByName("ExternalPluginService")
	externalPluginServiceCreateTaskMethodDescriptor = externalPluginServiceServiceDescriptor.Methods().ByName("CreateTask")
	externalPluginServiceGetTaskMethodDescriptor    = externalPluginServiceServiceDescriptor.Methods().ByName("GetTask")
	externalPluginServiceDeleteTaskMethodDescriptor = externalPluginServiceServiceDescriptor.Methods().ByName("DeleteTask")
)

// ExternalPluginServiceClient is a client for the flyteidl.service.ExternalPluginService service.
type ExternalPluginServiceClient interface {
	// Send a task create request to the backend plugin server.
	//
	// Deprecated: do not use.
	CreateTask(context.Context, *connect.Request[service.TaskCreateRequest]) (*connect.Response[service.TaskCreateResponse], error)
	// Get job status.
	//
	// Deprecated: do not use.
	GetTask(context.Context, *connect.Request[service.TaskGetRequest]) (*connect.Response[service.TaskGetResponse], error)
	// Delete the task resource.
	//
	// Deprecated: do not use.
	DeleteTask(context.Context, *connect.Request[service.TaskDeleteRequest]) (*connect.Response[service.TaskDeleteResponse], error)
}

// NewExternalPluginServiceClient constructs a client for the flyteidl.service.ExternalPluginService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewExternalPluginServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ExternalPluginServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &externalPluginServiceClient{
		createTask: connect.NewClient[service.TaskCreateRequest, service.TaskCreateResponse](
			httpClient,
			baseURL+ExternalPluginServiceCreateTaskProcedure,
			connect.WithSchema(externalPluginServiceCreateTaskMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getTask: connect.NewClient[service.TaskGetRequest, service.TaskGetResponse](
			httpClient,
			baseURL+ExternalPluginServiceGetTaskProcedure,
			connect.WithSchema(externalPluginServiceGetTaskMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		deleteTask: connect.NewClient[service.TaskDeleteRequest, service.TaskDeleteResponse](
			httpClient,
			baseURL+ExternalPluginServiceDeleteTaskProcedure,
			connect.WithSchema(externalPluginServiceDeleteTaskMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// externalPluginServiceClient implements ExternalPluginServiceClient.
type externalPluginServiceClient struct {
	createTask *connect.Client[service.TaskCreateRequest, service.TaskCreateResponse]
	getTask    *connect.Client[service.TaskGetRequest, service.TaskGetResponse]
	deleteTask *connect.Client[service.TaskDeleteRequest, service.TaskDeleteResponse]
}

// CreateTask calls flyteidl.service.ExternalPluginService.CreateTask.
//
// Deprecated: do not use.
func (c *externalPluginServiceClient) CreateTask(ctx context.Context, req *connect.Request[service.TaskCreateRequest]) (*connect.Response[service.TaskCreateResponse], error) {
	return c.createTask.CallUnary(ctx, req)
}

// GetTask calls flyteidl.service.ExternalPluginService.GetTask.
//
// Deprecated: do not use.
func (c *externalPluginServiceClient) GetTask(ctx context.Context, req *connect.Request[service.TaskGetRequest]) (*connect.Response[service.TaskGetResponse], error) {
	return c.getTask.CallUnary(ctx, req)
}

// DeleteTask calls flyteidl.service.ExternalPluginService.DeleteTask.
//
// Deprecated: do not use.
func (c *externalPluginServiceClient) DeleteTask(ctx context.Context, req *connect.Request[service.TaskDeleteRequest]) (*connect.Response[service.TaskDeleteResponse], error) {
	return c.deleteTask.CallUnary(ctx, req)
}

// ExternalPluginServiceHandler is an implementation of the flyteidl.service.ExternalPluginService
// service.
type ExternalPluginServiceHandler interface {
	// Send a task create request to the backend plugin server.
	//
	// Deprecated: do not use.
	CreateTask(context.Context, *connect.Request[service.TaskCreateRequest]) (*connect.Response[service.TaskCreateResponse], error)
	// Get job status.
	//
	// Deprecated: do not use.
	GetTask(context.Context, *connect.Request[service.TaskGetRequest]) (*connect.Response[service.TaskGetResponse], error)
	// Delete the task resource.
	//
	// Deprecated: do not use.
	DeleteTask(context.Context, *connect.Request[service.TaskDeleteRequest]) (*connect.Response[service.TaskDeleteResponse], error)
}

// NewExternalPluginServiceHandler builds an HTTP handler from the service implementation. It
// returns the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewExternalPluginServiceHandler(svc ExternalPluginServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	externalPluginServiceCreateTaskHandler := connect.NewUnaryHandler(
		ExternalPluginServiceCreateTaskProcedure,
		svc.CreateTask,
		connect.WithSchema(externalPluginServiceCreateTaskMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	externalPluginServiceGetTaskHandler := connect.NewUnaryHandler(
		ExternalPluginServiceGetTaskProcedure,
		svc.GetTask,
		connect.WithSchema(externalPluginServiceGetTaskMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	externalPluginServiceDeleteTaskHandler := connect.NewUnaryHandler(
		ExternalPluginServiceDeleteTaskProcedure,
		svc.DeleteTask,
		connect.WithSchema(externalPluginServiceDeleteTaskMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/flyteidl.service.ExternalPluginService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ExternalPluginServiceCreateTaskProcedure:
			externalPluginServiceCreateTaskHandler.ServeHTTP(w, r)
		case ExternalPluginServiceGetTaskProcedure:
			externalPluginServiceGetTaskHandler.ServeHTTP(w, r)
		case ExternalPluginServiceDeleteTaskProcedure:
			externalPluginServiceDeleteTaskHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedExternalPluginServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedExternalPluginServiceHandler struct{}

func (UnimplementedExternalPluginServiceHandler) CreateTask(context.Context, *connect.Request[service.TaskCreateRequest]) (*connect.Response[service.TaskCreateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("flyteidl.service.ExternalPluginService.CreateTask is not implemented"))
}

func (UnimplementedExternalPluginServiceHandler) GetTask(context.Context, *connect.Request[service.TaskGetRequest]) (*connect.Response[service.TaskGetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("flyteidl.service.ExternalPluginService.GetTask is not implemented"))
}

func (UnimplementedExternalPluginServiceHandler) DeleteTask(context.Context, *connect.Request[service.TaskDeleteRequest]) (*connect.Response[service.TaskDeleteResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("flyteidl.service.ExternalPluginService.DeleteTask is not implemented"))
}
