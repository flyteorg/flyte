package controller

import (
	"strconv"
	"strings"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const controllerAgentName = "flyteworkflow-controller"
const workflowTerminationStatusKey = "termination-status"
const workflowTerminatedValue = "terminated"
const hourOfDayCompletedKey = "hour-of-day"
const completedTimeKey = "completed-time"

// This function creates a label selector, that will ignore all objects (in this case workflow) that DOES NOT have a
// label key=workflowTerminationStatusKey with a value=workflowTerminatedValue
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

// Creates a new LabelSelector that selects all workflows that have the completed Label
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
	w.Labels[completedTimeKey] = SplitTimeToMeetsLabelFormat(currentTime)
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

// Calculates a list of all the hours that should be deleted given the current hour of the day and the retentionperiod in hours
// Usually this is a list of all hours out of the 24 hours in the day - retention period - the current hour of the day
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

//Calculates a list of all the hours that should be kept given the current time and the retentionperiod in hours
func CalculateHoursToKeep(retentionPeriodHours int, currentTime time.Time) []string {
	hoursToKeep := make([]string, 0, retentionPeriodHours+1)
	for i := 0; i <= retentionPeriodHours; i++ {
		hoursToKeep = append(hoursToKeep, SplitTimeToMeetsLabelFormat(currentTime))
		currentTime = currentTime.Add(-1 * time.Hour)
	}
	return hoursToKeep
}

// Creates a new selector that selects all completed workflows and workflows with completed hour label outside of the
// retention window
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
		Operator: v1.LabelSelectorOpIn,
		Values:   []string{""},
	})
	return s
}

func CompletedWorkflowsSelectorOutsideRetentionPeriodAbandon(retentionPeriodHours int, currentTime time.Time) *v1.LabelSelector {
	hoursToDelete := CalculateHoursToDelete(retentionPeriodHours-1, currentTime.Hour())
	s := CompletedWorkflowsLabelSelector()
	s.MatchExpressions = append(s.MatchExpressions, v1.LabelSelectorRequirement{
		Key:      hourOfDayCompletedKey,
		Operator: v1.LabelSelectorOpIn,
		Values:   hoursToDelete,
	})
	return s
}

func SplitTimeToMeetsLabelFormat(currentTime time.Time) string {
	return strings.ReplaceAll(
		strings.Split(currentTime.Round(time.Hour).String(), ":")[0], " ", ".")
}
