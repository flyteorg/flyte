package controller

import (
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

const (
	controllerAgentName          = "flyteworkflow-controller"
	workflowTerminationStatusKey = "termination-status"
	workflowTerminatedValue      = "terminated"
	completedHourTimeKey         = "completed-time"
	completedMinuteTimeKey       = "completed-time-minutes"
	minuteGranularity            = 5 * time.Minute
	// Layout string for time.Format() function expects the time ("Mon, 02 Jan 2006 15:04:05 MST") formatted
	// in the desired layout.
	labelHourTimeFormat   = "2006-01-02.15"
	labelMinuteTimeFormat = "2006-01-02T15_04"
)

// IgnoreCompletedWorkflowsLabelSelector this function creates a label selector, that will ignore all objects (in this
// case workflow) that DOES NOT have a label key=workflowTerminationStatusKey with a value=workflowTerminatedValue
func IgnoreCompletedWorkflowsLabelSelector() *v1.LabelSelector {
	return &v1.LabelSelector{
		MatchExpressions: []v1.LabelSelectorRequirement{
			{
				Key:      workflowTerminationStatusKey,
				Operator: v1.LabelSelectorOpNotIn,
				Values:   []string{workflowTerminatedValue},
			},
		},
	}
}

// CompletedWorkflowsLabelSelector creates a new LabelSelector that selects all workflows that have the completed Label
func CompletedWorkflowsLabelSelector() *v1.LabelSelector {
	return &v1.LabelSelector{
		MatchLabels: map[string]string{
			workflowTerminationStatusKey: workflowTerminatedValue,
		},
	}
}

func SetCompletedLabel(w *v1alpha1.FlyteWorkflow, currentTime time.Time) {
	if w.Labels == nil {
		w.Labels = make(map[string]string)
	}
	w.Labels[workflowTerminationStatusKey] = workflowTerminatedValue
	w.Labels[completedHourTimeKey] = FormatTimeForHourLabel(currentTime)
	w.Labels[completedMinuteTimeKey] = FormatTimeForMinuteLabel(currentTime)
}

func HasCompletedLabel(w *v1alpha1.FlyteWorkflow) bool {
	return w.Labels[workflowTerminationStatusKey] == workflowTerminatedValue
}

// CalculateRetentionLabelValues calculates a list of values for the retention label of CRDs that should be kept
// given the current time and the retention period
func CalculateRetentionLabelValues(ttl time.Duration, currentTime time.Time) []string {
	step := minuteGranularity
	if ttl >= time.Hour {
		step = time.Hour
	}
	n := int64(ttl / step)
	labelValues := make([]string, 0, n+1)
	for i := int64(0); i <= n; i++ {
		if ttl >= time.Hour {
			labelValues = append(labelValues, FormatTimeForHourLabel(currentTime))
		} else {
			labelValues = append(labelValues, FormatTimeForMinuteLabel(currentTime))
		}
		currentTime = currentTime.Add(-step)
	}
	return labelValues
}

// CompletedWorkflowsSelectorOutsideRetentionPeriod creates a new selector that selects all completed workflows and
// workflows with completed hour label outside the retention window.
func CompletedWorkflowsSelectorOutsideRetentionPeriod(ttl time.Duration, currentTime time.Time) *v1.LabelSelector {
	key := completedMinuteTimeKey
	if ttl >= time.Hour {
		key = completedHourTimeKey
	}
	labelValues := CalculateRetentionLabelValues(ttl, currentTime)
	s := CompletedWorkflowsLabelSelector()
	s.MatchExpressions = append(s.MatchExpressions, v1.LabelSelectorRequirement{
		Key:      key,
		Operator: v1.LabelSelectorOpNotIn,
		Values:   labelValues,
	})

	return s
}

// FormatTimeForHourLabel returns a string representation of the time with the following properties:
// 1. It's safe to put as a label in k8s objects' metadata
// 2. The granularity is up to the hour only.
// 3. The format is YYYY-MM-DD.HH
// 4. Is always in UTC.
func FormatTimeForHourLabel(currentTime time.Time) string {
	// We only care about the date and the hour within.
	// We convert to UTC to avoid possible issues with changing timezones between deployments.
	return currentTime.UTC().Format(labelHourTimeFormat)
}

// FormatTimeForMinuteLabel returns value for completed-time-minute label with a 5-min granularity
func FormatTimeForMinuteLabel(currentTime time.Time) string {
	return currentTime.Truncate(minuteGranularity).UTC().Format(labelMinuteTimeFormat)
}
