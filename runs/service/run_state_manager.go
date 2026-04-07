package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type node struct {
	Parent   *node
	Action   *models.Action
	Children []*node

	ChildPhaseCounts        map[common.ActionPhase]int
	MatchingDescendantCount int
}

type nodeUpdate struct {
	Node        *node
	MeetsFilter bool
}

type runStateManager struct {
	AllNodes map[string]*node
	// VisibleNodes tracks the nodes that should currently be rendered by the frontend.
	// This includes both nodes that match the filter directly and ancestor nodes that
	// must stay visible to preserve the tree path to a matching descendant.
	VisibleNodes map[string]*node

	// DirectlyMatchingNodes tracks only the nodes whose own fields match the active
	// filter. Ancestors that are shown only because they lead to a matching descendant
	// are not included here.
	DirectlyMatchingNodes map[string]*node

	phaseFilters map[common.ActionPhase]struct{}
	nameFilter   string
}

// newRunStateManager creates a per-watch tree state manager that tracks filter
// visibility and child phase aggregates for a run.
func newRunStateManager(filters []*common.Filter) (*runStateManager, error) {
	phaseFilters, nameFilter, err := validateAndTransformFilters(filters)
	if err != nil {
		return nil, err
	}

	return &runStateManager{
		AllNodes:              make(map[string]*node),
		VisibleNodes:          make(map[string]*node),
		DirectlyMatchingNodes: make(map[string]*node),
		phaseFilters:          phaseFilters,
		nameFilter:            nameFilter,
	}, nil
}

// upsertActions updates the run tree from the incoming actions and returns the
// subset of nodes whose displayed state changed.
func (rsm *runStateManager) upsertActions(ctx context.Context, actions []*models.Action) ([]*nodeUpdate, error) {
	changed := make(map[string]struct{})
	// First, upsert each incoming action into the tree and collect the set of nodes
	// whose state changed directly or through parent/child aggregate updates.
	for _, action := range actions {
		if action == nil {
			continue
		}

		changedFromUpsert, err := rsm.upsertAction(ctx, action)
		if err != nil {
			return nil, err
		}
		for name := range changedFromUpsert {
			changed[name] = struct{}{}
		}
	}

	// Second, reconcile which changed nodes now match the active filters directly.
	// This also updates ancestor MatchingDescendantCount values so visibility can be
	// recomputed from a fully settled matching state in the next pass.
	for nodeID := range changed {
		node := rsm.GetActionTreeNodeByName(nodeID)
		if node == nil {
			continue
		}

		matchesFilter := rsm.doesActionMatchFilters(node)
		_, wasDirectlyMatching := rsm.DirectlyMatchingNodes[nodeID]

		if matchesFilter && !wasDirectlyMatching {
			// Matched the filter but didn't match before
			// Increase count for its parent
			rsm.DirectlyMatchingNodes[nodeID] = node
			rsm.modifyMatchingDescendantCount(node.Parent, +1)
		} else if !matchesFilter && wasDirectlyMatching {
			// Didn't match filter but matched before
			// Decrease count for its parent
			delete(rsm.DirectlyMatchingNodes, nodeID)
			rsm.modifyMatchingDescendantCount(node.Parent, -1)
		}
	}

	resultSet := make(map[string]bool)
	// Third, recompute frontend visibility for the changed nodes using the updated
	// direct-match state. A node is visible if it matches itself or if it must stay
	// visible to preserve the path to a matching descendant.
	for nodeID := range changed {
		node := rsm.GetActionTreeNodeByName(nodeID)
		if node == nil {
			continue
		}
		if _, ok := resultSet[nodeID]; ok {
			continue
		}

		_, isDirectlyMatching := rsm.DirectlyMatchingNodes[nodeID]
		_, wasVisible := rsm.VisibleNodes[nodeID]

		switch {
		case isDirectlyMatching:
			rsm.VisibleNodes[nodeID] = node
			// Update parents to be visible
			for parent := node.Parent; parent != nil; parent = parent.Parent {
				if _, parentVisible := rsm.VisibleNodes[parent.Action.Name]; !parentVisible {
					rsm.VisibleNodes[parent.Action.Name] = parent
					resultSet[parent.Action.Name] = true
				}
			}
			resultSet[nodeID] = true
		case wasVisible:
			if rsm.hasMatchingDescendants(node) {
				resultSet[nodeID] = true
			} else {
				// If node does not match filter, and didn't have children match, make it invisible
				delete(rsm.VisibleNodes, nodeID)
				resultSet[nodeID] = false
				rsm.pruneAncestorsWithoutMatchingDescendants(node.Parent, resultSet)
			}
		}
	}

	result := make([]*nodeUpdate, 0, len(resultSet))
	// Finally, build the response payload from the nodes whose visible state changed.
	for nodeID, show := range resultSet {
		node := rsm.GetActionTreeNodeByName(nodeID)
		if node == nil {
			continue
		}
		result = append(result, &nodeUpdate{
			Node:        node,
			MeetsFilter: show,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Node.Action.CreatedAt.Before(result[j].Node.Action.CreatedAt)
	})

	return result, nil
}

// upsertAction inserts a new action into the tree or updates an existing node,
// then returns the set of node names whose state changed and must be reconciled
// for filtering, visibility, or child phase counts.
func (rsm *runStateManager) upsertAction(_ context.Context, action *models.Action) (map[string]struct{}, error) {
	changed := make(map[string]struct{})

	current := rsm.GetActionTreeNodeByName(action.Name)
	// Insert node if not exists
	if current == nil {
		node := &node{
			Action:           action.Clone(),
			ChildPhaseCounts: make(map[common.ActionPhase]int),
		}
		rsm.AllNodes[action.Name] = node
		if err := rsm.insertAction(node, changed); err != nil {
			delete(rsm.AllNodes, action.Name)
			return nil, err
		}
		return changed, nil
	}

	// Update node
	oldPhase := common.ActionPhase(current.Action.Phase)
	newPhase := common.ActionPhase(action.Phase)
	parentName := getParentActionName(current)
	if parentName != "" && current.Parent == nil {
		return nil, fmt.Errorf("parent node [%s] not found for action [%s]", parentName, current.Action.Name)
	}
	current.Action = action.Clone()

	rsm.modifyPhaseCounters(current.Parent, newPhase, oldPhase, changed)
	changed[current.Action.Name] = struct{}{}
	return changed, nil
}

func (rsm *runStateManager) insertAction(actionNode *node, changed map[string]struct{}) error {
	if actionNode == nil || actionNode.Action == nil {
		return nil
	}

	changed[actionNode.Action.Name] = struct{}{}
	parentName := getParentActionName(actionNode)
	if parentName == "" {
		return nil
	}

	parent := rsm.GetActionTreeNodeByName(parentName)
	if parent == nil {
		return fmt.Errorf("parent node [%s] not found for action [%s]", parentName, actionNode.Action.Name)
	}

	rsm.attachChild(parent, actionNode, changed)
	return nil
}

// attachChild attach given child to the parent node, and update ChildPhaseCounts if have changed nodes
func (rsm *runStateManager) attachChild(parent, child *node, changed map[string]struct{}) {
	if parent == nil || child == nil || child.Parent == parent {
		return
	}

	for _, existing := range parent.Children {
		if existing == child {
			child.Parent = parent
			return
		}
	}

	child.Parent = parent
	parent.Children = append(parent.Children, child)
	rsm.modifyPhaseCounters(parent, common.ActionPhase(child.Action.Phase), common.ActionPhase_ACTION_PHASE_UNSPECIFIED, changed)
	changed[child.Action.Name] = struct{}{}
}

func (rsm *runStateManager) modifyPhaseCounters(current *node, toPhase, fromPhase common.ActionPhase, changed map[string]struct{}) {
	for current != nil {
		if toPhase != common.ActionPhase_ACTION_PHASE_UNSPECIFIED {
			current.ChildPhaseCounts[toPhase]++
		}
		if fromPhase != common.ActionPhase_ACTION_PHASE_UNSPECIFIED {
			count := current.ChildPhaseCounts[fromPhase]
			if count <= 1 {
				current.ChildPhaseCounts[fromPhase] = 0
			} else {
				current.ChildPhaseCounts[fromPhase] = count - 1
			}
		}

		changed[current.Action.Name] = struct{}{}
		current = current.Parent
	}
}

func (rsm *runStateManager) modifyMatchingDescendantCount(current *node, delta int) {
	for current != nil {
		current.MatchingDescendantCount += delta
		current = current.Parent
	}
}

func (rsm *runStateManager) hasMatchingDescendants(node *node) bool {
	return node != nil && node.MatchingDescendantCount > 0
}

func (rsm *runStateManager) pruneAncestorsWithoutMatchingDescendants(node *node, resultSet map[string]bool) {
	for current := node; current != nil; current = current.Parent {
		nodeID := current.Action.Name
		if _, isVisible := rsm.VisibleNodes[nodeID]; !isVisible {
			break
		}
		if rsm.hasMatchingDescendants(current) {
			break
		}
		if _, directlyMatches := rsm.DirectlyMatchingNodes[nodeID]; directlyMatches {
			break
		}

		delete(rsm.VisibleNodes, nodeID)
		if _, inResult := resultSet[nodeID]; !inResult {
			resultSet[nodeID] = false
		}
	}
}

func (rsm *runStateManager) doesActionMatchFilters(node *node) bool {
	if node == nil || node.Action == nil {
		return false
	}
	if rsm.nameFilter == "" && len(rsm.phaseFilters) == 0 {
		return true
	}

	_, phaseMatched := rsm.phaseFilters[common.ActionPhase(node.Action.Phase)]
	matchesPhase := len(rsm.phaseFilters) == 0 || phaseMatched

	taskName := actionFilterName(node.Action)
	matchesName := rsm.nameFilter == "" || strings.Contains(strings.ToLower(taskName), rsm.nameFilter)

	return matchesPhase && matchesName
}

func (rsm *runStateManager) GetActionTreeNodeByName(name string) *node {
	return rsm.AllNodes[name]
}

func getParentActionName(actionNode *node) string {
	if actionNode == nil || actionNode.Action == nil || actionNode.Action.ParentActionName.Valid == false {
		return ""
	}
	return actionNode.Action.ParentActionName.String
}

func validateAndTransformFilters(filters []*common.Filter) (map[common.ActionPhase]struct{}, string, error) {
	if len(filters) == 0 {
		return nil, "", nil
	}

	phaseFilters := make(map[common.ActionPhase]struct{})
	var nameFilter string

	for _, filter := range filters {
		if filter == nil {
			continue
		}
		if !isSupportedFilterFunction(filter) {
			return nil, "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("function %s is not supported for filter %s", filter.GetFunction(), filter.GetField()))
		}

		switch filter.GetField() {
		case "PHASE":
			for _, value := range filter.GetValues() {
				phaseInt, err := strconv.Atoi(value)
				if err != nil {
					return nil, "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid phase value: %s", value))
				}
				phaseFilters[common.ActionPhase(phaseInt)] = struct{}{}
			}
		case "NAME":
			if len(filter.GetValues()) == 0 {
				continue
			}
			nameFilter = strings.ToLower(filter.GetValues()[0])
		default:
			return nil, "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported filter field: %s", filter.GetField()))
		}
	}

	return phaseFilters, nameFilter, nil
}

func isSupportedFilterFunction(filter *common.Filter) bool {
	if filter == nil {
		return false
	}
	switch filter.GetField() {
	case "PHASE":
		return filter.GetFunction() == common.Filter_VALUE_IN
	case "NAME":
		return filter.GetFunction() == common.Filter_CONTAINS_CASE_INSENSITIVE
	default:
		return false
	}
}

func actionFilterName(action *models.Action) string {
	if action == nil {
		return ""
	}

	taskName := extractShortName(extractTaskName(action.ActionSpec))
	if taskName != "" {
		return taskName
	}

	return action.Name
}
