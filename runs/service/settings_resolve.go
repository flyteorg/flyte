package service

import (
	"context"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// fetchLevels returns the scope levels covered by key, broadest first, together with
// the stored row for each. The two slices align by position, and a nil row means no
// record exists at that level. Order comes from the key ladder; the repo returns rows
// unordered.
func fetchLevels(ctx context.Context, repo interfaces.SettingsRepo, key *settings.SettingsKey) ([]*settings.SettingsKey, []*models.Settings, error) {
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

	rows, err := repo.GetSettingsByKeys(ctx, storageKeys)
	if err != nil {
		return nil, nil, err
	}
	rowsByKey := make(map[string]*models.Settings, len(rows))
	for _, row := range rows {
		rowsByKey[row.Key] = row
	}

	aligned := make([]*models.Settings, len(levelKeys))
	for i := range levelKeys {
		aligned[i] = rowsByKey[storageKeys[i]]
	}
	return levelKeys, aligned, nil
}

// decodeStored unmarshals one stored row into a Settings message. Unknown fields are
// discarded so a row written by a newer server still loads on an older one.
func decodeStored(row *models.Settings) (*settings.Settings, error) {
	stored := &settings.Settings{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(row.Data, stored); err != nil {
		return nil, err
	}
	return stored, nil
}

// resolveSettings returns the effective settings at the scope named by key, merging
// every level of the chain. Callers inside this package use this directly; GetSettings
// only wraps it in a response.
func resolveSettings(ctx context.Context, repo interfaces.SettingsRepo, key *settings.SettingsKey) (*settings.Settings, error) {
	_, rows, err := fetchLevels(ctx, repo, key)
	if err != nil {
		return nil, err
	}

	levels := make([]*settings.Settings, len(rows))
	for i, row := range rows {
		if row == nil {
			continue
		}
		stored, err := decodeStored(row)
		if err != nil {
			return nil, err
		}
		levels[i] = stored
	}
	return mergeSettings(levels), nil
}
