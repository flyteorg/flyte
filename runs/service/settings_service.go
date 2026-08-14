package service

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"google.golang.org/protobuf/encoding/protojson"

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
	// The buf.validate annotations on these protos are not enforced by the
	// generated Go code, so required fields and key shape are checked by hand
	key := req.Msg.GetKey()
	if key == nil || req.Msg.GetSettings() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("key and settings are required"))
	}
	if key.GetProject() != "" && key.GetDomain() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("a project-scope key requires a domain"))
	}

	// Currently we store the Settings message as a raw protojson, sparse storage
	// (prune/hydrate) arrives with the resolution work (RFC #7775 task 2.1)
	data, err := protojson.Marshal(req.Msg.GetSettings())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	model := &models.Settings{
		Key:  models.EncodeSettingsKey(key.GetOrg(), key.GetDomain(), key.GetProject()),
		Data: data,
	}

	if err := s.settingsRepo.CreateSettings(ctx, model); err != nil {
		if errors.Is(err, interfaces.ErrSettingsAlreadyExists) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&settings.CreateSettingsResponse{
		SettingsRecord: &settings.SettingsRecord{
			Key:      key,
			Settings: req.Msg.GetSettings(),
			Version:  model.Version,
		},
	}), nil
}

func (s *SettingsService) UpdateSettings(
	ctx context.Context,
	req *connect.Request[settings.UpdateSettingsRequest],
) (*connect.Response[settings.UpdateSettingsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented yet"))
}

var _ settingsconnect.SettingsServiceHandler = (*SettingsService)(nil)
