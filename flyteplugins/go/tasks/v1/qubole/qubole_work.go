package qubole

import (
	"fmt"
	"encoding/json"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/v1/events"
	"github.com/lyft/flyteplugins/go/tasks/v1/qubole/client"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"
)

// This struct is supposed to represent all the details of one query/unit of work on Qubole.  For instance, a user's
// @qubole_hive_task will get unpacked to one of these for each query contained in the task.
// It is intentionally vaguely named, in an effort to potentially support extensibility to other things Qubole
// is capable of executing in the future.
// Retries and Status are the only two fields that should get changed.
type QuboleWorkItem struct {
	// This ID is the cache key and so will need to be unique across all objects in the cache (it will probably be
	// unique across all of Flyte) and needs to be deterministic.
	// This will also be used as the allocation token for now.
	UniqueWorkCacheKey string `json:"unique_work_cache_key"`

	// This will store the command ID from Qubole
	CommandId string `json:"command_id,omitempty"`

	// Our representation of the status of this work item
	Status QuboleWorkItemStatus `json:"status,omitempty"`

	// The Qubole cluster to do this work
	ClusterLabel string `json:"cluster_label,omitempty"`

	// These are Qubole Tags that show up on their UI
	Tags []string `json:"tags,omitempty"`

	// This number keeps track of the number of retries within the sync function.  Without this, what happens in
	// the sync function is entirely opaque.  Note that this field is not meant to represent the number of retries
	// of the work itself, just errors with the Qubole API when attempting to sync
	Retries int `json:"retries,omitempty"`

	// For Hive jobs, this is the query that will be run
	// Not necessary for other Qubole task types necessarily
	Query string `json:"query,omitempty"`

	TimeoutSec uint32 `json:"timeout,omitempty"`
}

// This ID will be used in a process-wide cache, so it needs to be unique across all concurrent work being done by
// that process, but does not necessarily need to be universally unique
func (q QuboleWorkItem) ID() string {
	return q.UniqueWorkCacheKey
}

func constructQuboleWorkItem(uniqueWorkCacheKey string, quboleCommandId string, status QuboleWorkItemStatus) QuboleWorkItem {
	return QuboleWorkItem{
		UniqueWorkCacheKey: uniqueWorkCacheKey,
		CommandId:          quboleCommandId,
		Status:             status,
	}
}

func NewQuboleWorkItem(uniqueWorkCacheKey string, quboleCommandId string, status QuboleWorkItemStatus, clusterLabel string,
	tags []string, retries int) QuboleWorkItem {
	return QuboleWorkItem{
		UniqueWorkCacheKey: uniqueWorkCacheKey,
		CommandId:          quboleCommandId,
		Status:             status,
		ClusterLabel:       clusterLabel,
		Tags:               tags,
		Retries:            retries,
	}
}

// This status encapsulates all possible states for our custom object.  It is different from the QuboleStatus type
// in that this is our Flyte type.  It represents the same thing as QuboleStatus, but will actually persist in etcd.
// It is also different from the TaskStatuses in that this is on the qubole job level, not the task level.  A task,
// can contain many queries/spark jobs, etc.
type QuboleWorkItemStatus int

const (
	QuboleWorkNotStarted QuboleWorkItemStatus = iota
	QuboleWorkUnknown
	QuboleWorkRunnable
	QuboleWorkRunning
	QuboleWorkExecutionFailed
	QuboleWorkExecutionSucceeded
	QuboleWorkFailed
	QuboleWorkSucceeded
)

func QuboleWorkIsTerminalState(status QuboleWorkItemStatus) bool {
	return status == QuboleWorkFailed || status == QuboleWorkSucceeded
}

func (q QuboleWorkItemStatus) String() string {
	switch q {
	case QuboleWorkNotStarted:
		return "NotStarted"
	case QuboleWorkUnknown:
		return "Unknown"
	case QuboleWorkRunnable:
		return "Runnable"
	case QuboleWorkRunning:
		return "Running"
	case QuboleWorkExecutionFailed:
		return "ExecutionFailed"
	case QuboleWorkExecutionSucceeded:
		return "ExecutionSucceeded"
	case QuboleWorkFailed:
		return "Failed"
	case QuboleWorkSucceeded:
		return "Succeeded"
	}
	return "IllegalQuboleWorkStatus"
}

func (q QuboleWorkItem) EqualTo(other QuboleWorkItem) bool {
	if q.UniqueWorkCacheKey != other.UniqueWorkCacheKey ||
		q.Status != other.Status || q.CommandId != other.CommandId ||
		q.Retries != other.Retries || len(q.Tags) != len(other.Tags) {
		return false
	}

	return true
}

func workItemMapsAreEqual(old map[string]QuboleWorkItem, new map[string]interface{}) bool {
	if len(old) != len(new) {
		return false
	}

	for k, oldItem := range old {
		if x, ok := new[k]; ok {
			newItem := x.(QuboleWorkItem)
			if !oldItem.EqualTo(newItem) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func constructEventInfoFromQuboleWorkItems(taskCtx types.TaskContext, quboleWorkItems map[string]interface{}) (*events.TaskEventInfo, error) {
	logs := make([]*core.TaskLog, 0, len(quboleWorkItems))
	for _, v := range quboleWorkItems {
		workItem := v.(QuboleWorkItem)
		if workItem.CommandId != "" {
			logs = append(logs, &core.TaskLog{
				Name:          fmt.Sprintf("Retry: %d Status: %s [%s]",
					taskCtx.GetTaskExecutionID().GetID().RetryAttempt, workItem.Status, workItem.CommandId),
				MessageFormat: core.TaskLog_UNKNOWN,
				Uri:           fmt.Sprintf(client.QuboleLogLinkFormat, workItem.CommandId),
			})
		}
	}

	customInfo, err := utils.MarshalObjToStruct(quboleWorkItems)
	if err != nil {
		return nil, err
	}

	return &events.TaskEventInfo{
		CustomInfo: customInfo,
		Logs:       logs,
	}, nil
}

func QuboleStatusToWorkItemStatus(s client.QuboleStatus) QuboleWorkItemStatus {
	switch s {
	case client.QuboleStatusDone:
		return QuboleWorkSucceeded
	case client.QuboleStatusCancelled:
		return QuboleWorkFailed
	case client.QuboleStatusError:
		return QuboleWorkFailed
	case client.QuboleStatusUnknown:
		return QuboleWorkUnknown
	case client.QuboleStatusWaiting:
		return QuboleWorkRunning
	case client.QuboleStatusRunning:
		return QuboleWorkRunning
	default:
		return QuboleWorkRunning
	}
}

func InterfaceConverter(cachedInterface interface{}) (QuboleWorkItem, error) {
	raw, err := json.Marshal(cachedInterface)
	if err != nil {
		return QuboleWorkItem{}, err
	}

	item := &QuboleWorkItem{}
	err = json.Unmarshal(raw, item)
	if err != nil {
		return QuboleWorkItem{}, fmt.Errorf("Failed to unmarshal state into Qubole work item")
	}

	return *item, nil
}