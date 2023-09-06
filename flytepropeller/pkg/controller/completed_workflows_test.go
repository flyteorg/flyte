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
		"2009-11-10.06", "2009-11-10.05", "2009-11-10.04", "2009-11-10.03", "2009-11-10.02", "2009-11-10.01", "2009-11-10.00",
	}, CalculateHoursToKeep(6, time.Date(2009, time.November, 10, 6, 0, 0, 0, time.UTC)))

	assert.Equal(t, []string{
		"2009-10-01.03", "2009-10-01.02", "2009-10-01.01", "2009-10-01.00", "2009-09-30.23", "2009-09-30.22", "2009-09-30.21",
	}, CalculateHoursToKeep(6, time.Date(2009, time.October, 1, 3, 0, 0, 0, time.UTC)))

	assert.Equal(t, []string{
		"2009-01-01.00", "2008-12-31.23", "2008-12-31.22", "2008-12-31.21", "2008-12-31.20", "2008-12-31.19", "2008-12-31.18",
	}, CalculateHoursToKeep(6, time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC)))

	assert.Equal(t, []string{
		"2009-11-10.22", "2009-11-10.21", "2009-11-10.20", "2009-11-10.19", "2009-11-10.18", "2009-11-10.17", "2009-11-10.16",
	}, CalculateHoursToKeep(6, time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC)))

	assert.Equal(t, []string{
		"2009-11-10.23", "2009-11-10.22", "2009-11-10.21", "2009-11-10.20", "2009-11-10.19", "2009-11-10.18", "2009-11-10.17",
	}, CalculateHoursToKeep(6, time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)))

	assert.Equal(t, []string{"2009-11-10.20", "2009-11-10.19"}, CalculateHoursToKeep(1, time.Date(2009, time.November, 10, 20, 0, 0, 0, time.UTC)))
	assert.Equal(t, []string{"2009-11-10.23", "2009-11-10.22", "2009-11-10.21"}, CalculateHoursToKeep(2, time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)))

	assert.Equal(t, []string{
		"2009-11-10.12", "2009-11-10.11", "2009-11-10.10", "2009-11-10.09", "2009-11-10.08", "2009-11-10.07", "2009-11-10.06", "2009-11-10.05", "2009-11-10.04", "2009-11-10.03", "2009-11-10.02", "2009-11-10.01", "2009-11-10.00", "2009-11-09.23", "2009-11-09.22", "2009-11-09.21", "2009-11-09.20", "2009-11-09.19", "2009-11-09.18", "2009-11-09.17", "2009-11-09.16", "2009-11-09.15", "2009-11-09.14",
	}, CalculateHoursToKeep(22, time.Date(2009, time.November, 10, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, []string{
		"2009-11-10.00", "2009-11-09.23", "2009-11-09.22", "2009-11-09.21", "2009-11-09.20", "2009-11-09.19", "2009-11-09.18", "2009-11-09.17", "2009-11-09.16", "2009-11-09.15", "2009-11-09.14", "2009-11-09.13", "2009-11-09.12", "2009-11-09.11", "2009-11-09.10", "2009-11-09.09", "2009-11-09.08", "2009-11-09.07", "2009-11-09.06", "2009-11-09.05", "2009-11-09.04", "2009-11-09.03", "2009-11-09.02",
	}, CalculateHoursToKeep(22, time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)))

	assert.Equal(t, []string{"2022-03-30.12", "2022-03-30.11"}, CalculateHoursToKeep(1, time.Date(2022, time.March, 30, 12, 10, 0, 0, time.UTC)))
}

func TestCompletedWorkflowsSelectorOutsideRetentionPeriod(t *testing.T) {
	n := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	s := CompletedWorkflowsSelectorOutsideRetentionPeriod(2, n)
	v, ok := s.MatchLabels[workflowTerminationStatusKey]
	assert.True(t, ok)
	assert.Equal(t, workflowTerminatedValue, v)
	assert.NotEmpty(t, s.MatchExpressions)
	r := s.MatchExpressions[0]
	assert.Equal(t, completedTimeKey, r.Key)
	assert.Equal(t, v1.LabelSelectorOpNotIn, r.Operator)
	assert.Equal(t, 3, len(r.Values))
	assert.Equal(t, []string{
		"2009-11-10.23", "2009-11-10.22", "2009-11-10.21",
	}, r.Values)
}

func TestFormatTimeForLabel(t *testing.T) {
	assert.Equal(t, "1970-01-12.13", FormatTimeForLabel(time.Unix(1000000, 100000)))
}
