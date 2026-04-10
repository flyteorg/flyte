package controller

import (
	"context"
	"fmt"
	"strings"
	"testing"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
)

// mockPluginResolver is a test double for pluginResolver.
type mockPluginResolver struct {
	plugin pluginsCore.Plugin
	err    error
}

func (m *mockPluginResolver) ResolvePlugin(_ string) (pluginsCore.Plugin, error) {
	return m.plugin, m.err
}

// mockPlugin is a minimal pluginsCore.Plugin implementation for tests.
type mockPlugin struct{}

func (mockPlugin) GetID() string                               { return "mock" }
func (mockPlugin) GetProperties() pluginsCore.PluginProperties { return pluginsCore.PluginProperties{} }
func (mockPlugin) Handle(_ context.Context, _ pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	return pluginsCore.Transition{}, nil
}
func (mockPlugin) Abort(_ context.Context, _ pluginsCore.TaskExecutionContext) error    { return nil }
func (mockPlugin) Finalize(_ context.Context, _ pluginsCore.TaskExecutionContext) error { return nil }

func validTaskAction() *flyteorgv1.TaskAction {
	return &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			RunName:       "my-run",
			Project:       "my-project",
			Domain:        "my-domain",
			ActionName:    "my-action",
			TaskType:      "container",
			TaskTemplate:  []byte(`{}`),
			InputURI:      "s3://bucket/input",
			RunOutputBase: "s3://bucket/output",
		},
	}
}

func TestValidateTaskAction_ValidSpec(t *testing.T) {
	resolver := &mockPluginResolver{plugin: mockPlugin{}}
	p, reason, err := validateTaskAction(validTaskAction(), resolver)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if reason != "" {
		t.Fatalf("expected empty reason, got: %q", reason)
	}
	if p == nil {
		t.Fatal("expected non-nil plugin")
	}
}

func TestValidateTaskAction_MissingFields(t *testing.T) {
	cases := []struct {
		name          string
		mutate        func(*flyteorgv1.TaskAction)
		expectedField string
	}{
		{"missing runName", func(ta *flyteorgv1.TaskAction) { ta.Spec.RunName = "" }, "runName"},
		{"missing project", func(ta *flyteorgv1.TaskAction) { ta.Spec.Project = "" }, "project"},
		{"missing domain", func(ta *flyteorgv1.TaskAction) { ta.Spec.Domain = "" }, "domain"},
		{"missing actionName", func(ta *flyteorgv1.TaskAction) { ta.Spec.ActionName = "" }, "actionName"},
		{"missing taskType", func(ta *flyteorgv1.TaskAction) { ta.Spec.TaskType = "" }, "taskType"},
		{"missing taskTemplate", func(ta *flyteorgv1.TaskAction) { ta.Spec.TaskTemplate = nil }, "taskTemplate"},
		{"missing inputUri", func(ta *flyteorgv1.TaskAction) { ta.Spec.InputURI = "" }, "inputUri"},
		{"missing runOutputBase", func(ta *flyteorgv1.TaskAction) { ta.Spec.RunOutputBase = "" }, "runOutputBase"},
	}

	resolver := &mockPluginResolver{plugin: mockPlugin{}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ta := validTaskAction()
			tc.mutate(ta)
			_, reason, err := validateTaskAction(ta, resolver)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.expectedField) {
				t.Errorf("expected error to mention %q, got: %v", tc.expectedField, err)
			}
			if reason != flyteorgv1.ConditionReasonInvalidSpec {
				t.Errorf("expected reason %q, got %q", flyteorgv1.ConditionReasonInvalidSpec, reason)
			}
		})
	}
}

func TestValidateTaskAction_PluginNotFound(t *testing.T) {
	resolver := &mockPluginResolver{
		plugin: nil,
		err:    fmt.Errorf("no plugin registered for task type %q", "container"),
	}
	_, reason, err := validateTaskAction(validTaskAction(), resolver)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if reason != flyteorgv1.ConditionReasonPluginNotFound {
		t.Errorf("expected reason %q, got %q", flyteorgv1.ConditionReasonPluginNotFound, reason)
	}
}
