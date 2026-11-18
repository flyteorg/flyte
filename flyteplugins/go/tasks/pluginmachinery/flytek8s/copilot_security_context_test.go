package flytek8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/flytestdlib/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func coPilotTestSecurityContext(uid int64) *v1.SecurityContext {
	nonRoot, escalation := true, false
	return &v1.SecurityContext{
		RunAsUser:                &uid,
		RunAsNonRoot:             &nonRoot,
		AllowPrivilegeEscalation: &escalation,
		Capabilities:             &v1.Capabilities{Drop: []v1.Capability{"ALL"}},
		SeccompProfile:           &v1.SeccompProfile{Type: v1.SeccompProfileTypeRuntimeDefault},
	}
}

func TestCoPilotSecurityContext(t *testing.T) {
	uid := int64(2000)
	partialContext := &v1.SecurityContext{RunAsUser: &uid}
	tests := []struct {
		name                                            string
		podTarget                                       bool
		inputs, outputs                                 bool
		defaultContext, primaryContext, templateContext *v1.SecurityContext
		initTemplates                                   []v1.Container
		deviceTemplate                                  bool
		disabled, noInterface, noDataConfig             bool
	}{
		{name: "container defaults", inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000)},
		{name: "pod defaults", podTarget: true, inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000)},
		{name: "input only", inputs: true, defaultContext: coPilotTestSecurityContext(1000)},
		{name: "output only", outputs: true, defaultContext: coPilotTestSecurityContext(1000)},
		{name: "explicit primary", podTarget: true, inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000), primaryContext: coPilotTestSecurityContext(2000)},
		{name: "partial primary suppresses default", podTarget: true, inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000), primaryContext: partialContext},
		{name: "empty primary suppresses default", podTarget: true, inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000), primaryContext: &v1.SecurityContext{}},
		{name: "primary from base template", inputs: true, outputs: true, templateContext: coPilotTestSecurityContext(2000)},
		{name: "primary from device template", inputs: true, outputs: true, templateContext: coPilotTestSecurityContext(2000), deviceTemplate: true},
		{name: "default init template", inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000), initTemplates: []v1.Container{{Name: defaultInitContainerTemplateName, SecurityContext: coPilotTestSecurityContext(3000)}}},
		{name: "named init template", inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000), initTemplates: []v1.Container{{Name: "custom-downloader", SecurityContext: coPilotTestSecurityContext(3000)}}},
		{name: "partial init context", inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000), initTemplates: []v1.Container{{Name: "custom-downloader", SecurityContext: partialContext}}},
		{name: "empty init context", inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000), initTemplates: []v1.Container{{Name: "custom-downloader", SecurityContext: &v1.SecurityContext{}}}},
		{name: "no configured context", inputs: true, outputs: true},
		{name: "disabled", podTarget: true, inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000), disabled: true},
		{name: "nil interface", podTarget: true, defaultContext: coPilotTestSecurityContext(1000), noInterface: true},
		{name: "empty interface", podTarget: true, defaultContext: coPilotTestSecurityContext(1000)},
		{name: "nil data config", podTarget: true, inputs: true, outputs: true, defaultContext: coPilotTestSecurityContext(1000), noDataConfig: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous := config.GetK8sPluginConfig()
			t.Cleanup(func() { require.NoError(t, config.SetK8sPluginConfig(previous)) })
			nonRoot := true
			cfg := &config.K8sPluginConfig{
				DefaultCPURequest:         resource.MustParse("100m"),
				DefaultMemoryRequest:      resource.MustParse("100Mi"),
				DefaultSecurityContext:    tt.defaultContext.DeepCopy(),
				DefaultPodSecurityContext: &v1.PodSecurityContext{RunAsNonRoot: &nonRoot},
				CoPilot: config.FlyteCoPilotConfig{
					NamePrefix: "custom-", Image: "copilot:test",
					DefaultInputDataPath: "/in", DefaultOutputPath: "/out",
					InputVolumeName: "inputs", OutputVolumeName: "outputs",
					CPU: "100m", Memory: "100Mi",
					StorageConfigOverride: &storage.Config{
						Type: storage.TypeStow,
						Stow: storage.StowConfig{Kind: "s3", Config: map[string]string{"region": "us-east-1"}},
					},
				},
			}
			defaultSnapshot := cfg.DefaultSecurityContext.DeepCopy()
			iface := &core.TypedInterface{}
			if tt.inputs {
				iface.Inputs = coPilotTestVariables()
			}
			if tt.outputs {
				iface.Outputs = coPilotTestVariables()
			}
			if tt.noInterface {
				iface = nil
			}
			pilot := &core.DataLoadingConfig{Enabled: !tt.disabled}
			if tt.noDataConfig {
				pilot = nil
			}
			task := dummyTaskTemplate()
			task.Interface = iface
			task.GetContainer().DataConfig = pilot
			generated := !tt.disabled && !tt.noInterface && !tt.noDataConfig
			primaryName := "some-acceptable-name"
			var existing []v1.Container
			if tt.podTarget {
				primaryName = "task-primary"
				existing = []v1.Container{{Name: "unrelated-init"}, {Name: "configured-init", SecurityContext: coPilotTestSecurityContext(4000)}}
				if !generated || !tt.inputs && !tt.outputs {
					existing = append(existing, v1.Container{Name: "custom-downloader"}, v1.Container{Name: "custom-uploader"})
				}
				spec, err := utils.MarshalObjToStruct(v1.PodSpec{
					Containers: []v1.Container{
						{Name: "unrelated-sidecar", SecurityContext: coPilotTestSecurityContext(5000)},
						{Name: primaryName, SecurityContext: tt.primaryContext.DeepCopy()},
					},
					InitContainers: existing,
				})
				require.NoError(t, err)
				task.Target = &core.TaskTemplate_K8SPod{K8SPod: &core.K8SPod{PodSpec: spec, DataConfig: pilot}}
				task.Config = map[string]string{PrimaryContainerKey: primaryName}
			}
			template := &v1.PodTemplate{
				ObjectMeta: metav1.ObjectMeta{Name: "copilot-security-template", Namespace: "test-namespace"},
				Template: v1.PodTemplateSpec{Spec: v1.PodSpec{
					Containers:     []v1.Container{{Name: primaryContainerTemplateName, SecurityContext: tt.templateContext.DeepCopy()}},
					InitContainers: tt.initTemplates,
				}},
			}
			templateSnapshot := template.DeepCopy()
			if tt.deviceTemplate {
				cfg.AcceleratorDeviceClasses = map[string]config.AcceleratorDeviceClassConfig{
					"AMAZON_NEURON": {PodTemplate: template},
				}
				task.ExtendedResources = &core.ExtendedResources{GpuAccelerator: &core.GPUAccelerator{DeviceClass: core.GPUAccelerator_AMAZON_NEURON}}
			} else if tt.templateContext != nil || len(tt.initTemplates) != 0 {
				DefaultPodTemplateStore.Store(template)
				t.Cleanup(func() { DefaultPodTemplateStore.Delete(template) })
				task.Metadata = &core.TaskMetadata{PodTemplateName: template.Name}
			}
			require.NoError(t, config.SetK8sPluginConfig(cfg))
			taskContext := dummyExecContext(task, &v1.ResourceRequirements{}, nil, "", nil)
			pod, _, name, err := ToK8sPodSpec(context.Background(), taskContext)
			require.NoError(t, err)
			require.Equal(t, primaryName, name)
			primary := coPilotTestContainer(t, pod.Containers, primaryName)
			wantPrimary := tt.primaryContext.DeepCopy()
			if wantPrimary == nil {
				wantPrimary = tt.defaultContext.DeepCopy()
			}
			if tt.templateContext != nil {
				wantPrimary = tt.templateContext.DeepCopy()
			}
			if generated {
				if wantPrimary == nil {
					wantPrimary = &v1.SecurityContext{}
				}
				if wantPrimary.Capabilities == nil {
					wantPrimary.Capabilities = &v1.Capabilities{}
				}
			}
			require.Equal(t, wantPrimary, primary.SecurityContext)
			assert.Equal(t, cfg.DefaultPodSecurityContext, pod.SecurityContext)
			for _, c := range existing {
				assert.Equal(t, c.SecurityContext, coPilotTestContainer(t, pod.InitContainers, c.Name).SecurityContext)
			}
			if tt.podTarget {
				assert.Equal(t, coPilotTestSecurityContext(5000), coPilotTestContainer(t, pod.Containers, "unrelated-sidecar").SecurityContext)
			}
			expectedInitCount := len(existing)
			var inherited []*v1.Container
			for _, helper := range []struct {
				name   string
				needed bool
			}{
				{"custom-downloader", generated && tt.inputs},
				{"custom-uploader", generated && tt.outputs},
			} {
				if !helper.needed {
					continue
				}
				expectedInitCount++
				container := coPilotTestContainer(t, pod.InitContainers, helper.name)
				expected := primary.SecurityContext
				for _, initTemplate := range tt.initTemplates {
					if initTemplate.Name == defaultInitContainerTemplateName || initTemplate.Name == helper.name {
						expected = initTemplate.SecurityContext
					}
				}
				require.Equal(t, expected, container.SecurityContext)
				if expected == primary.SecurityContext {
					inherited = append(inherited, container)
				}
				if helper.name == "custom-uploader" {
					require.Equal(t, helper.name, pod.InitContainers[0].Name)
					require.NotNil(t, container.RestartPolicy)
					assert.Equal(t, v1.ContainerRestartPolicyAlways, *container.RestartPolicy)
				}
			}
			require.Len(t, pod.InitContainers, expectedInitCount)
			primarySnapshot := primary.SecurityContext.DeepCopy()
			for _, container := range inherited {
				require.NotSame(t, primary.SecurityContext, container.SecurityContext)
				if container.SecurityContext.RunAsUser != nil {
					*container.SecurityContext.RunAsUser = 9999
					if len(container.SecurityContext.Capabilities.Drop) > 0 {
						container.SecurityContext.Capabilities.Drop[0] = "CHANGED"
					}
					assert.Equal(t, primarySnapshot, primary.SecurityContext)
					for _, sibling := range inherited {
						if sibling != container {
							assert.Equal(t, primarySnapshot, sibling.SecurityContext)
						}
					}
					container.SecurityContext = primarySnapshot.DeepCopy()
				}
			}
			assert.Equal(t, defaultSnapshot, cfg.DefaultSecurityContext)
			for idx := range template.Template.Spec.Containers {
				assert.Equal(t, templateSnapshot.Template.Spec.Containers[idx].SecurityContext, template.Template.Spec.Containers[idx].SecurityContext)
			}
			for idx := range template.Template.Spec.InitContainers {
				assert.Equal(t, templateSnapshot.Template.Spec.InitContainers[idx].SecurityContext, template.Template.Spec.InitContainers[idx].SecurityContext)
			}
		})
	}
}

func coPilotTestVariables() *core.VariableMap {
	return &core.VariableMap{Variables: []*core.VariableEntry{{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}}}}
}

func coPilotTestContainer(t *testing.T, containers []v1.Container, name string) *v1.Container {
	t.Helper()
	for idx := range containers {
		if containers[idx].Name == name {
			return &containers[idx]
		}
	}
	t.Fatalf("container %q not found", name)
	return nil
}
