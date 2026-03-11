package service

import (
	"reflect"

	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type runStateManager struct {
	actions map[string]*models.Action
}

// newRunStateManager creates a per-watch in-memory action store.
func newRunStateManager() *runStateManager {
	return &runStateManager{
		actions: make(map[string]*models.Action),
	}
}

// upsertActions merges the incoming actions into the in-memory state and
// returns only the actions whose stored value changed.
func (rsm *runStateManager) upsertActions(actions []*models.Action) []*models.Action {
	if len(actions) == 0 {
		return nil
	}

	changed := make([]*models.Action, 0, len(actions))
	changedIdxByName := make(map[string]int, len(actions))

	// Loop through actions to update, update the action map and return the updated actions
	for _, action := range actions {
		if action == nil {
			continue
		}

		next := action.Clone()
		existing, ok := rsm.actions[action.Name]
		// Continue if the new action has no changes
		if ok && reflect.DeepEqual(existing, next) {
			continue
		}

		rsm.actions[action.Name] = next
		if idx, ok := changedIdxByName[action.Name]; ok {
			changed[idx] = next
			continue
		}

		changedIdxByName[action.Name] = len(changed)
		changed = append(changed, next)
	}

	return changed
}
