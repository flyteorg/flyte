package rpc

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings/settingsconnect"
	"github.com/flyteorg/flyte/v2/runs/service"
)

type settingsHandler struct{ settings service.Settings }

// NewSettingsHandler exposes the settings service through the Connect API.
func NewSettingsHandler(settings service.Settings) settingsconnect.SettingsServiceHandler {
	return &settingsHandler{settings: settings}
}

func (h *settingsHandler) GetSettings(ctx context.Context, req *connect.Request[settings.GetSettingsRequest]) (*connect.Response[settings.GetSettingsResponse], error) {
	return rpcResponse(h.settings.GetSettings(ctx, req.Msg))
}

func (h *settingsHandler) GetSettingsForEdit(ctx context.Context, req *connect.Request[settings.GetSettingsForEditRequest]) (*connect.Response[settings.GetSettingsForEditResponse], error) {
	return rpcResponse(h.settings.GetSettingsForEdit(ctx, req.Msg))
}

func (h *settingsHandler) CreateSettings(ctx context.Context, req *connect.Request[settings.CreateSettingsRequest]) (*connect.Response[settings.CreateSettingsResponse], error) {
	return rpcResponse(h.settings.CreateSettings(ctx, req.Msg))
}

func (h *settingsHandler) UpdateSettings(ctx context.Context, req *connect.Request[settings.UpdateSettingsRequest]) (*connect.Response[settings.UpdateSettingsResponse], error) {
	return rpcResponse(h.settings.UpdateSettings(ctx, req.Msg))
}
