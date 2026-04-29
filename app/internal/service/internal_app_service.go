package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	appk8s "github.com/flyteorg/flyte/v2/app/internal/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

// InternalAppService is the data plane implementation of the AppService.
// It has direct K8s access via AppK8sClientInterface.
type InternalAppService struct {
	appconnect.UnimplementedAppServiceHandler
	k8s appk8s.AppK8sClientInterface
}

// NewInternalAppService creates a new InternalAppService.
func NewInternalAppService(k8s appk8s.AppK8sClientInterface) *InternalAppService {
	return &InternalAppService{k8s: k8s}
}

// Ensure InternalAppService satisfies the generated handler interface.
var _ appconnect.AppServiceHandler = (*InternalAppService)(nil)

// Create deploys a new app as a KService CRD.
func (s *InternalAppService) Create(
	ctx context.Context,
	req *connect.Request[flyteapp.CreateRequest],
) (*connect.Response[flyteapp.CreateResponse], error) {
	app := req.Msg.GetApp()
	if app.GetMetadata().GetId() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("app identifier is required"))
	}
	if app.GetSpec() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("app spec is required"))
	}
	if app.GetSpec().GetAppPayload() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("app spec must include a container or pod payload"))
	}

	if err := s.k8s.Deploy(ctx, app); err != nil {
		logger.Errorf(ctx, "Failed to deploy app %s: %v", app.GetMetadata().GetId().GetName(), err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	app.Status = &flyteapp.Status{
		Ingress: s.k8s.PublicIngress(app.GetMetadata().GetId()),
	}

	return connect.NewResponse(&flyteapp.CreateResponse{App: app}), nil
}

// Get retrieves an app and its live status from the KService CRD.
// Note: App.Spec is not populated — status and ingress URL are the authoritative fields.
func (s *InternalAppService) Get(
	ctx context.Context,
	req *connect.Request[flyteapp.GetRequest],
) (*connect.Response[flyteapp.GetResponse], error) {
	appID, ok := req.Msg.GetIdentifier().(*flyteapp.GetRequest_AppId)
	if !ok || appID.AppId == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("app_id is required"))
	}

	app, err := s.getApp(ctx, appID.AppId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&flyteapp.GetResponse{App: app}), nil
}

// getApp fetches an app from K8s with its live status.
func (s *InternalAppService) getApp(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.App, error) {
	app, err := s.k8s.GetApp(ctx, appID)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return app, nil
}

// Update modifies an app's spec or desired state.
// When Spec.DesiredState is STOPPED, the app is scaled to zero (KService kept).
// When Spec.DesiredState is STARTED or ACTIVE, the app is redeployed/resumed.
// Otherwise the spec update is applied and the app is redeployed.
func (s *InternalAppService) Update(
	ctx context.Context,
	req *connect.Request[flyteapp.UpdateRequest],
) (*connect.Response[flyteapp.UpdateResponse], error) {
	app := req.Msg.GetApp()
	if app.GetMetadata().GetId() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("app identifier is required"))
	}

	appID := app.GetMetadata().GetId()

	switch app.GetSpec().GetDesiredState() {
	case flyteapp.Spec_DESIRED_STATE_STOPPED:
		if err := s.k8s.Stop(ctx, appID); err != nil {
			logger.Errorf(ctx, "Failed to stop app %s: %v", appID.GetName(), err)
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	default:
		// UNSPECIFIED, STARTED, ACTIVE — deploy/redeploy the spec.
		if err := s.k8s.Deploy(ctx, app); err != nil {
			logger.Errorf(ctx, "Failed to update app %s: %v", appID.GetName(), err)
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	freshApp, err := s.getApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	app.Status = freshApp.Status

	return connect.NewResponse(&flyteapp.UpdateResponse{App: app}), nil
}

// Delete removes the KService CRD for the given app entirely.
func (s *InternalAppService) Delete(
	ctx context.Context,
	req *connect.Request[flyteapp.DeleteRequest],
) (*connect.Response[flyteapp.DeleteResponse], error) {
	appID := req.Msg.GetAppId()
	if appID == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("app_id is required"))
	}

	if err := s.k8s.Delete(ctx, appID); err != nil {
		logger.Errorf(ctx, "Failed to delete app %s: %v", appID.GetName(), err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&flyteapp.DeleteResponse{}), nil
}

// List returns apps for the requested scope with pagination.
func (s *InternalAppService) List(
	ctx context.Context,
	req *connect.Request[flyteapp.ListRequest],
) (*connect.Response[flyteapp.ListResponse], error) {
	var project, domain string

	switch f := req.Msg.GetFilterBy().(type) {
	case *flyteapp.ListRequest_Project:
		project = f.Project.GetName()
		domain = f.Project.GetDomain()
	case *flyteapp.ListRequest_Org, *flyteapp.ListRequest_ClusterId:
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("org and cluster_id filters are not supported by the data plane"))
	}

	var limit uint32
	var token string
	if r := req.Msg.GetRequest(); r != nil {
		limit = r.GetLimit()
		token = r.GetToken()
	}

	apps, nextToken, err := s.k8s.List(ctx, project, domain, limit, token)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&flyteapp.ListResponse{Apps: apps, Token: nextToken}), nil
}

// Watch streams live KService events to the client.
// It first sends the current state as CreateEvents (initial snapshot), then streams changes.
func (s *InternalAppService) Watch(
	ctx context.Context,
	req *connect.Request[flyteapp.WatchRequest],
	stream *connect.ServerStream[flyteapp.WatchResponse],
) error {
	var project, domain, appName string

	switch t := req.Msg.GetTarget().(type) {
	case *flyteapp.WatchRequest_AppId:
		project = t.AppId.GetProject()
		domain = t.AppId.GetDomain()
		appName = t.AppId.GetName()
	case *flyteapp.WatchRequest_Project, *flyteapp.WatchRequest_Org, *flyteapp.WatchRequest_ClusterId:
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("only app_id watch target is supported"))
	}

	// Subscribe before listing so no events are lost between the two calls.
	ch := s.k8s.Subscribe(appName)
	defer s.k8s.Unsubscribe(appName, ch)

	// Send initial snapshot so the client has current state before streaming changes.
	snapshot, _, err := s.k8s.List(ctx, project, domain, 0, "")
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	for _, app := range snapshot {
		if appName != "" && app.GetMetadata().GetId().GetName() != appName {
			continue
		}
		if err := stream.Send(&flyteapp.WatchResponse{
			Event: &flyteapp.WatchResponse_CreateEvent{
				CreateEvent: &flyteapp.CreateEvent{App: app},
			},
		}); err != nil {
			return err
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}
