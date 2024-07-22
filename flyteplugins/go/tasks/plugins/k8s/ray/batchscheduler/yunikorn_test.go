package batchscheduler

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateTaskGroupName(t *testing.T) {
	var tests = []struct{
		master bool
		index int
		expect string
	}{
		{true, 0, fmt.Sprintf("%s-%s", GenerateTaskGroupName, "head")},
		{false, 0, fmt.Sprintf("%s-%s-%d", GenerateTaskGroupName, "worker", 0)},
		{false, 1, fmt.Sprintf("%s-%s-%d", GenerateTaskGroupName, "worker", 1)},
	}
	for _, tt := range tests {
		t.Run("Generating Task group name", func(t *testing.T) {
			if got := GenerateTaskGroupName(tt.master, tt.index); got != tt.expect {
				t.Errorf("got %s, expect %s", got, tt.expect)
			}
		})
	}
}

func TestRemoveGangSchedulingAnnotations(t *testing.T) {
	var tests = []struct{
		input *metav1.ObjectMeta
		expect int
	}{
		{input: &metav1.ObjectMeta{"others": "extra", TaskGroupNameKey: "TGName", TaskGroupsKey: "TGs", TaskGroupPrarameters: "parameters"}, 1},
		{input: &metav1.ObjectMeta{TaskGroupNameKey: "TGName", TaskGroupsKey: "TGs", TaskGroupPrarameters: "parameters"}, 0},
		{input: &metav1.ObjectMeta{TaskGroupNameKey: "TGName", TaskGroupsKey: "TGs"}, 0},
		{input: &metav1.ObjectMeta{TaskGroupNameKey: "TGName"}, 0},
		{input: &metav1.ObjectMeta{}, 0},
	}
	for _, tt := range tests {
		t.Run("Remove Gang scheduling labels", func(t *testing.T){
			RemoveGangSchedulingAnnotations(tt.input)
			if got := len(tt.input); got != tt.expect {
				t.Errorf("got %d, expect %d", got, tt.expect)
			}
		})
	}
}

