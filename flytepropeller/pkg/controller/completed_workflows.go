package controller

import (
	"strconv"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	controllerAgentName          = "flyteworkflow-controller"
	workflowTerminationStatusKey = "termination-status"
	workflowTerminatedValue      = "terminated"
	hourOfDayCompletedKey        = "hour-of-day"
	completedTimeKey             = "completed-time"
	// Layout string for time.Format() function expects the time ("Mon, 02 Jan 2006 15:04:05 MST") formatted
	// in the desired layout.
	labelTimeFormat = "2006-01-02.15"
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
	w.Labels[completedTimeKey] = FormatTimeForLabel(currentTime)
}

func HasCompletedLabel(w *v1alpha1.FlyteWorkflow) bool {
	if w.Labels != nil {
		v, ok := w.Labels[workflowTerminationStatusKey]
		if ok {
			return v == workflowTerminatedValue
		}
	}
	return false
}

// CalculateHoursToDelete calculates a list of all the hours that should be deleted given the current hour of the day
// and the retention period in hours. Usually this is a list of all hours out of the 24 hours in the day - retention
// period - the current hour of the day
func CalculateHoursToDelete(retentionPeriodHours, currentHourOfDay int) []string {
	numberOfHoursToDelete := 24 - retentionPeriodHours
	hoursToDelete := make([]string, 0, numberOfHoursToDelete)

	for i := 0; i < currentHourOfDay-retentionPeriodHours; i++ {
		hoursToDelete = append(hoursToDelete, strconv.Itoa(i))
	}

	maxHourOfDay := 24
	if currentHourOfDay-retentionPeriodHours < 0 {
		maxHourOfDay = 24 + (currentHourOfDay - retentionPeriodHours)
	}

	for i := currentHourOfDay + 1; i < maxHourOfDay; i++ {
		hoursToDelete = append(hoursToDelete, strconv.Itoa(i))
	}

	return hoursToDelete
}

// CalculateHoursToKeep calculates a list of all the hours that should be kept given the current time and the retention
// period in hours
func CalculateHoursToKeep(retentionPeriodHours int, currentTime time.Time) []string {
	hoursToKeep := make([]string, 0, retentionPeriodHours+1)
	for i := 0; i <= retentionPeriodHours; i++ {
		hoursToKeep = append(hoursToKeep, FormatTimeForLabel(currentTime))
		currentTime = currentTime.Add(-1 * time.Hour)
	}
	return hoursToKeep
}

// CompletedWorkflowsSelectorOutsideRetentionPeriod creates a new selector that selects all completed workflows and
// workflows with completed hour label outside the retention window.
func CompletedWorkflowsSelectorOutsideRetentionPeriod(retentionPeriodHours int, currentTime time.Time) *v1.LabelSelector {
	hoursToKeep := CalculateHoursToKeep(retentionPeriodHours, currentTime)
	s := CompletedWorkflowsLabelSelector()
	s.MatchExpressions = append(s.MatchExpressions, v1.LabelSelectorRequirement{
		Key:      completedTimeKey,
		Operator: v1.LabelSelectorOpNotIn,
		Values:   hoursToKeep,
	})

	s.MatchExpressions = append(s.MatchExpressions, v1.LabelSelectorRequirement{
		Key:      hourOfDayCompletedKey,
		Operator: v1.LabelSelectorOpDoesNotExist,
	})

	return s
}

// DeprecatedCompletedWorkflowsSelectorOutsideRetentionPeriod
// Deprecated
func DeprecatedCompletedWorkflowsSelectorOutsideRetentionPeriod(retentionPeriodHours int, currentTime time.Time) *v1.LabelSelector {
	hoursToDelete := CalculateHoursToDelete(retentionPeriodHours-1, currentTime.Hour())
	s := CompletedWorkflowsLabelSelector()
	s.MatchExpressions = append(s.MatchExpressions, v1.LabelSelectorRequirement{
		Key:      hourOfDayCompletedKey,
		Operator: v1.LabelSelectorOpIn,
		Values:   hoursToDelete,
	})

	return s
}

// FormatTimeForLabel returns a string representation of the time with the following properties:
// 1. It's safe to put as a label in k8s objects' metadata
// 2. The granularity is up to the hour only.
// 3. The format is YYYY-MM-DD.HH
// 4. Is always in UTC.
func FormatTimeForLabel(currentTime time.Time) string {
	// We only care about the date and the hour within.
	// We convert to UTC to avoid possible issues with changing timezones between deployments.
	return currentTime.UTC().Format(labelTimeFormat)
}
