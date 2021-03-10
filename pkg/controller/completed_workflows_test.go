package controller

import (
	"testing"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIgnoreCompletedWorkflowsLabelSelector(t *testing.T) {
	s := IgnoreCompletedWorkflowsLabelSelector()
	assert.NotNil(t, s)
	assert.Empty(t, s.MatchLabels)
	assert.NotEmpty(t, s.MatchExpressions)
	r := s.MatchExpressions[0]
	assert.Equal(t, workflowTerminationStatusKey, r.Key)
	assert.Equal(t, v1.LabelSelectorOpNotIn, r.Operator)
	assert.Equal(t, []string{workflowTerminatedValue}, r.Values)
}

func TestCompletedWorkflowsLabelSelector(t *testing.T) {
	s := CompletedWorkflowsLabelSelector()
	assert.NotEmpty(t, s.MatchLabels)
	v, ok := s.MatchLabels[workflowTerminationStatusKey]
	assert.True(t, ok)
	assert.Equal(t, workflowTerminatedValue, v)
}

func TestHasCompletedLabel(t *testing.T) {

	n := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	t.Run("no-labels", func(t *testing.T) {

		w := &v1alpha1.FlyteWorkflow{}
		assert.Empty(t, w.Labels)
		assert.False(t, HasCompletedLabel(w))
		SetCompletedLabel(w, n)
		assert.NotEmpty(t, w.Labels)
		v, ok := w.Labels[workflowTerminationStatusKey]
		assert.True(t, ok)
		assert.Equal(t, workflowTerminatedValue, v)
		assert.True(t, HasCompletedLabel(w))
	})

	t.Run("existing-lables", func(t *testing.T) {
		w := &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"x": "v",
				},
			},
		}
		assert.NotEmpty(t, w.Labels)
		assert.False(t, HasCompletedLabel(w))
		SetCompletedLabel(w, n)
		assert.NotEmpty(t, w.Labels)
		v, ok := w.Labels[workflowTerminationStatusKey]
		assert.True(t, ok)
		assert.Equal(t, workflowTerminatedValue, v)
		v, ok = w.Labels["x"]
		assert.True(t, ok)
		assert.Equal(t, "v", v)
		assert.True(t, HasCompletedLabel(w))
	})
}

func TestSetCompletedLabel(t *testing.T) {
	n := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	t.Run("no-labels", func(t *testing.T) {

		w := &v1alpha1.FlyteWorkflow{}
		assert.Empty(t, w.Labels)
		SetCompletedLabel(w, n)
		assert.NotEmpty(t, w.Labels)
		v, ok := w.Labels[workflowTerminationStatusKey]
		assert.True(t, ok)
		assert.Equal(t, workflowTerminatedValue, v)
	})

	t.Run("existing-lables", func(t *testing.T) {
		w := &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"x": "v",
				},
			},
		}
		assert.NotEmpty(t, w.Labels)
		SetCompletedLabel(w, n)
		assert.NotEmpty(t, w.Labels)
		v, ok := w.Labels[workflowTerminationStatusKey]
		assert.True(t, ok)
		assert.Equal(t, workflowTerminatedValue, v)
		v, ok = w.Labels["x"]
		assert.True(t, ok)
		assert.Equal(t, "v", v)
	})

}

func TestCalculateHoursToDelete(t *testing.T) {
	assert.Equal(t, []string{
		"6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22",
	}, CalculateHoursToDelete(6, 5))

	assert.Equal(t, []string{
		"7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23",
	}, CalculateHoursToDelete(6, 6))

	assert.Equal(t, []string{
		"0", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23",
	}, CalculateHoursToDelete(6, 7))

	assert.Equal(t, []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "23",
	}, CalculateHoursToDelete(6, 22))

	assert.Equal(t, []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16",
	}, CalculateHoursToDelete(6, 23))

	assert.Equal(t, []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22",
	}, CalculateHoursToDelete(0, 23))

	assert.Equal(t, []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "21", "22", "23",
	}, CalculateHoursToDelete(0, 20))

	assert.Equal(t, []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23",
	}, CalculateHoursToDelete(0, 0))

	assert.Equal(t, []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23",
	}, CalculateHoursToDelete(0, 12))

	assert.Equal(t, []string{"13"}, CalculateHoursToDelete(22, 12))
	assert.Equal(t, []string{"1"}, CalculateHoursToDelete(22, 0))
	assert.Equal(t, []string{"0"}, CalculateHoursToDelete(22, 23))
	assert.Equal(t, []string{"23"}, CalculateHoursToDelete(22, 22))
}

func TestCompletedWorkflowsSelectorOutsideRetentionPeriod(t *testing.T) {
	n := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	s := CompletedWorkflowsSelectorOutsideRetentionPeriod(2, n)
	v, ok := s.MatchLabels[workflowTerminationStatusKey]
	assert.True(t, ok)
	assert.Equal(t, workflowTerminatedValue, v)
	assert.NotEmpty(t, s.MatchExpressions)
	r := s.MatchExpressions[0]
	assert.Equal(t, hourOfDayCompletedKey, r.Key)
	assert.Equal(t, v1.LabelSelectorOpIn, r.Operator)
	assert.Equal(t, 21, len(r.Values))
	assert.Equal(t, []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20",
	}, r.Values)
}
