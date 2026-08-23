package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/apimachinery/pkg/api/resource"

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

// GetSettings returns resolved, merged settings.
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
	key := req.Msg.GetKey()
	if key == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("key is required"))
	}
	if key.GetProject() != "" && key.GetDomain() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("a project-scope key requires a domain"))
	}

	// One partial key per scope level covered by the request, broadest first,
	// as GetSettingsForEditResponse requires. Order comes from this list; the
	// repo returns rows unordered.
	levelKeys := []*settings.SettingsKey{{Org: key.GetOrg()}}
	if key.GetDomain() != "" {
		levelKeys = append(levelKeys, &settings.SettingsKey{Org: key.GetOrg(), Domain: key.GetDomain()})
	}
	if key.GetProject() != "" {
		levelKeys = append(levelKeys, &settings.SettingsKey{Org: key.GetOrg(), Domain: key.GetDomain(), Project: key.GetProject()})
	}

	storageKeys := make([]string, 0, len(levelKeys))
	for _, lk := range levelKeys {
		storageKeys = append(storageKeys, models.EncodeSettingsKey(lk.GetOrg(), lk.GetDomain(), lk.GetProject()))
	}

	rows, err := s.settingsRepo.GetSettingsByKeys(ctx, storageKeys)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	rowsByKey := make(map[string]*models.Settings, len(rows))
	for _, row := range rows {
		rowsByKey[row.Key] = row
	}

	levels := make([]*settings.SettingsRecord, 0, len(levelKeys))
	for i, lk := range levelKeys {
		// Absent level: empty settings, version 0 — the client's signal to use
		// CreateSettings there (see SettingsRecord in the proto).
		record := &settings.SettingsRecord{Key: lk, Settings: &settings.Settings{}}
		if row := rowsByKey[storageKeys[i]]; row != nil {
			stored := &settings.Settings{}
			if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(row.Data, stored); err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
			record.Settings = stored
			record.Version = row.Version
		}
		levels = append(levels, record)
	}

	return connect.NewResponse(&settings.GetSettingsForEditResponse{
		RequestedKey: key,
		Levels:       levels,
	}), nil
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
	key := req.Msg.GetKey()
	if key == nil || req.Msg.GetSettings() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("key and settings are required"))
	}
	if key.GetProject() != "" && key.GetDomain() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("a project-scope key requires a domain"))
	}
	if req.Msg.GetVersion() == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("a version is required; use CreateSettings for a new record"))
	}

	data, err := protojson.Marshal(req.Msg.GetSettings())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	model := &models.Settings{
		Key:     models.EncodeSettingsKey(key.GetOrg(), key.GetDomain(), key.GetProject()),
		Data:    data,
		Version: req.Msg.GetVersion(),
	}

	// Pure update, not upsert: a missing record returns NotFound (use
	// CreateSettings), a stale version returns FailedPrecondition.
	if err := s.settingsRepo.UpdateSettings(ctx, model); err != nil {
		if errors.Is(err, interfaces.ErrSettingsNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.Is(err, interfaces.ErrSettingsVersionConflict) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&settings.UpdateSettingsResponse{
		SettingsRecord: &settings.SettingsRecord{
			Key:      key,
			Settings: req.Msg.GetSettings(),
			Version:  model.Version,
		},
	}), nil
}

func validateMaxActionConcurrency(setting *settings.Int64Setting) error {
	if setting.GetState() != settings.SettingState_SETTING_STATE_VALUE {
		return nil
	}
	if v := setting.GetIntValue(); v < 0 || v == 1 || v > math.MaxUint32 {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid max_action_concurrency %d: must be 0 (unlimited) or at least 2", v))
	}
	return nil
}

func validateQuantity(name string, setting *settings.QuantitySetting) error {
	if setting.GetState() != settings.SettingState_SETTING_STATE_VALUE {
		return nil
	}
	if _, err := resource.ParseQuantity(setting.GetQuantityValue()); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid %s %q: %w", name, setting.GetQuantityValue(), err))
	}
	return nil
}

func validateTaskResourceDefaults(bound string, d *settings.TaskResourceDefaults) error {
	if err := validateQuantity(bound+".cpu", d.GetCpu()); err != nil {
		return err
	}
	if err := validateQuantity(bound+".gpu", d.GetGpu()); err != nil {
		return err
	}
	if err := validateQuantity(bound+".memory", d.GetMemory()); err != nil {
		return err
	}
	return validateQuantity(bound+".storage", d.GetStorage())
}

func validateSettings(s *settings.Settings) error {
	if err := validateMaxActionConcurrency(s.GetRun().GetMaxActionConcurrency()); err != nil {
		return err
	}
	if err := validateTaskResourceDefaults("task_resource.min", s.GetTaskResource().GetMin()); err != nil {
		return err
	}
	return validateTaskResourceDefaults("task_resource.max", s.GetTaskResource().GetMax())
}

var _ settingsconnect.SettingsServiceHandler = (*SettingsService)(nil)
