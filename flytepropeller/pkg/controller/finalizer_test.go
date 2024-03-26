package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/batch/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFinalizersIdentical(t *testing.T) {
	noFinalizer := &v1.Job{}
	withFinalizer := &v1.Job{}
	withFinalizer.SetFinalizers([]string{"t1"})

	assert.True(t, FinalizersIdentical(noFinalizer, noFinalizer))
	assert.True(t, FinalizersIdentical(withFinalizer, withFinalizer))
	assert.False(t, FinalizersIdentical(noFinalizer, withFinalizer))
	withMultipleFinalizers := &v1.Job{}
	withMultipleFinalizers.SetFinalizers([]string{"f1", "f2"})
	assert.False(t, FinalizersIdentical(withMultipleFinalizers, withFinalizer))

	withDiffFinalizer := &v1.Job{}
	withDiffFinalizer.SetFinalizers([]string{"f1"})
	assert.True(t, FinalizersIdentical(withFinalizer, withDiffFinalizer))
}

func TestIsDeleted(t *testing.T) {
	noTermTS := &v1.Job{}
	termedTS := &v1.Job{}
	n := v12.Now()
	termedTS.SetDeletionTimestamp(&n)

	assert.True(t, IsDeleted(termedTS))
	assert.False(t, IsDeleted(noTermTS))
}

func TestHasFinalizer(t *testing.T) {
	noFinalizer := &v1.Job{}
	withFinalizer := &v1.Job{}
	withFinalizer.SetFinalizers([]string{"t1"})

	assert.False(t, HasFinalizer(noFinalizer))
	assert.True(t, HasFinalizer(withFinalizer))
}

func TestSetFinalizerIfEmpty(t *testing.T) {
	noFinalizer := &v1.Job{}
	withFinalizer := &v1.Job{}
	withFinalizer.SetFinalizers([]string{"t1"})

	assert.False(t, HasFinalizer(noFinalizer))
	SetFinalizerIfEmpty(noFinalizer, "f1")
	assert.True(t, HasFinalizer(noFinalizer))
	assert.Equal(t, []string{"f1"}, noFinalizer.GetFinalizers())

	SetFinalizerIfEmpty(withFinalizer, "f1")
	assert.Equal(t, []string{"t1"}, withFinalizer.GetFinalizers())
}

func TestResetFinalizer(t *testing.T) {
	noFinalizer := &v1.Job{}
	ResetFinalizers(noFinalizer)
	assert.Equal(t, []string{}, noFinalizer.GetFinalizers())

	withFinalizer := &v1.Job{}
	withFinalizer.SetFinalizers([]string{"t1"})
	ResetFinalizers(withFinalizer)
	assert.Equal(t, []string{}, withFinalizer.GetFinalizers())
}
