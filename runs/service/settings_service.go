package service

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings/settingsconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

type SettingsService struct {
	settingsRepo interfaces.SettingsRepo
}

// NewSettingsService returns the Connect handler implementation for the
// settings service, backed by the given repository.
func NewSettingsService(settingsRepo interfaces.SettingsRepo) *SettingsService {
	return &SettingsService{settingsRepo: settingsRepo}
}

// GetSettings returns resolved, merged settings. Resolution requires the
// merge engine (RFC #7775 task 2.2), which is not implemented yet.
func (s *SettingsService) GetSettings(
	ctx context.Context,
	req *connect.Request[settings.GetSettingsRequest],
) (*connect.Response[settings.GetSettingsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("settings resolution engine not implemented yet"))
}

func (s *SettingsService) GetSettingsForEdit(
	ctx context.Context,
	req *connect.Request[settings.GetSettingsForEditRequest],
) (*connect.Response[settings.GetSettingsForEditResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented yet"))
}

func (s *SettingsService) CreateSettings(
	ctx context.Context,
	req *connect.Request[settings.CreateSettingsRequest],
) (*connect.Response[settings.CreateSettingsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented yet"))
}

func (s *SettingsService) UpdateSettings(
	ctx context.Context,
	req *connect.Request[settings.UpdateSettingsRequest],
) (*connect.Response[settings.UpdateSettingsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented yet"))
}

var _ settingsconnect.SettingsServiceHandler = (*SettingsService)(nil)
