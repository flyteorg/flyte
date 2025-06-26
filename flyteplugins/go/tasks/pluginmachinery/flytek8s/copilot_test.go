package flytek8s

import (
	"context"
	"encoding/base64"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCoreMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginsIOMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	config2 "github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var resourceRequirements = &v1.ResourceRequirements{
	Limits: v1.ResourceList{
		v1.ResourceCPU:     resource.MustParse("1024m"),
		v1.ResourceStorage: resource.MustParse("100M"),
	},
}

func TestFlyteCoPilotContainer(t *testing.T) {
	cfg := config.FlyteCoPilotConfig{
		NamePrefix:           "test-",
		Image:                "test",
		DefaultInputDataPath: "/in",
		DefaultOutputPath:    "/out",
		InputVolumeName:      "inp",
		OutputVolumeName:     "out",
		StartTimeout: config2.Duration{
			Duration: time.Second * 1,
		},
		CPU:    "1024m",
		Memory: "1024Mi",
	}

	t.Run("happy", func(t *testing.T) {
		c, err := FlyteCoPilotContainer("x", cfg, []string{"hello"})
		assert.NoError(t, err)
		assert.Equal(t, "test-x", c.Name)
		assert.Equal(t, "test", c.Image)
		assert.Equal(t, CopilotCommandArgs(storage.GetConfig()), c.Command)
		assert.Equal(t, []string{"hello"}, c.Args)
		assert.Equal(t, 0, len(c.VolumeMounts))
		assert.Equal(t, "/", c.WorkingDir)
		assert.Equal(t, 2, len(c.Resources.Limits))
		assert.Equal(t, 2, len(c.Resources.Requests))
	})

	t.Run("happy stow backend", func(t *testing.T) {
		storage.GetConfig().Stow.Kind = "S3"
		storage.GetConfig().Stow.Config = map[string]string{
			"path": "config.yaml",
		}
		c, err := FlyteCoPilotContainer("x", cfg, []string{"hello"})
		assert.NoError(t, err)
		assert.Equal(t, "test-x", c.Name)
		assert.Equal(t, "test", c.Image)
		assert.Equal(t, CopilotCommandArgs(storage.GetConfig()), c.Command)
		assert.Equal(t, []string{"hello"}, c.Args)
		assert.Equal(t, 0, len(c.VolumeMounts))
		assert.Equal(t, "/", c.WorkingDir)
		assert.Equal(t, 2, len(c.Resources.Limits))
		assert.Equal(t, 2, len(c.Resources.Requests))
	})

	t.Run("happy-vols", func(t *testing.T) {
		c, err := FlyteCoPilotContainer("x", cfg, []string{"hello"}, v1.VolumeMount{Name: "X", MountPath: "/"})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(c.VolumeMounts))
	})

	t.Run("happy stow GCP backend", func(t *testing.T) {
		storage.GetConfig().Type = storage.TypeStow
		storage.GetConfig().InitContainer = "bucket"
		storage.GetConfig().Stow.Kind = "google"
		storage.GetConfig().Stow.Config = map[string]string{
			"json":       "",
			"project_id": "flyte-gcp",
			"scope":      "read_write",
		}
		assert.Equal(t, 12, len(CopilotCommandArgs(storage.GetConfig())))
	})

	t.Run("bad-res-cpu", func(t *testing.T) {
		old := cfg.CPU
		cfg.CPU = "x"
		_, err := FlyteCoPilotContainer("x", cfg, []string{"hello"}, v1.VolumeMount{Name: "X", MountPath: "/"})
		assert.Error(t, err)
		cfg.CPU = old
	})

	t.Run("bad-res-mem", func(t *testing.T) {
		old := cfg.Memory
		cfg.Memory = "x"
		_, err := FlyteCoPilotContainer("x", cfg, []string{"hello"}, v1.VolumeMount{Name: "X", MountPath: "/"})
		assert.Error(t, err)
		cfg.Memory = old
	})

	t.Run("sidecar-container-name-change", func(t *testing.T) {
		c, err := FlyteCoPilotContainer(flyteSidecarContainerName, cfg, []string{"hello"})
		assert.NoError(t, err)
		assert.Equal(t, "uploader", strings.Split(c.Name, "-")[1])
	})
}

func TestDownloadCommandArgs(t *testing.T) {
	_, err := DownloadCommandArgs("", "", "", core.DataLoadingConfig_YAML, nil)
	assert.Error(t, err)

	iFace := &core.VariableMap{
		Variables: map[string]*core.Variable{
			"x": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
			"y": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
		},
	}
	d, err := DownloadCommandArgs("s3://from", "s3://output-meta", "/to", core.DataLoadingConfig_JSON, iFace)
	assert.NoError(t, err)
	expected := []string{"download", "--from-remote", "s3://from", "--to-output-prefix", "s3://output-meta", "--to-local-dir", "/to", "--format", "JSON", "--input-interface", "<interface>"}
	if assert.Len(t, d, len(expected)) {
		for i := 0; i < len(expected)-1; i++ {
			assert.Equal(t, expected[i], d[i])
		}
		// We cannot compare the last one, as the interface is a map the order is not guaranteed.
		ifaceB64 := d[len(expected)-1]
		serIFaceBytes, err := base64.StdEncoding.DecodeString(ifaceB64)
		if assert.NoError(t, err) {
			vm := &core.VariableMap{}
			assert.NoError(t, proto.Unmarshal(serIFaceBytes, vm))
			assert.Len(t, vm.GetVariables(), 2)
			for k, v := range iFace.GetVariables() {
				v2, ok := vm.GetVariables()[k]
				assert.True(t, ok)
				assert.Equal(t, v.GetType().GetSimple(), v2.GetType().GetSimple(), "for %s, types do not match", k)
			}
		}
	}
}

func TestSidecarCommandArgs(t *testing.T) {
	_, err := SidecarCommandArgs("", "", "", time.Second*10, nil)
	assert.Error(t, err)

	iFace := &core.TypedInterface{
		Outputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				"y": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
			},
		},
	}
	d, err := SidecarCommandArgs("/from", "s3://output-meta", "s3://raw-output", time.Hour*1, iFace)
	assert.NoError(t, err)
	expected := []string{"sidecar", "--timeout", "1h0m0s", "--to-raw-output", "s3://raw-output", "--to-output-prefix", "s3://output-meta", "--from-local-dir", "/from", "--interface", "<interface>"}
	if assert.Len(t, d, len(expected)) {
		for i := 0; i < len(expected)-1; i++ {
			assert.Equal(t, expected[i], d[i])
		}
		// We cannot compare the last one, as the interface is a map the order is not guaranteed.
		ifaceB64 := d[len(expected)-1]
		serIFaceBytes, err := base64.StdEncoding.DecodeString(ifaceB64)
		if assert.NoError(t, err) {
			if2 := &core.TypedInterface{}
			assert.NoError(t, proto.Unmarshal(serIFaceBytes, if2))
			assert.Len(t, if2.GetOutputs().GetVariables(), 2)
			for k, v := range iFace.GetOutputs().GetVariables() {
				v2, ok := if2.GetOutputs().GetVariables()[k]
				assert.True(t, ok)
				assert.Equal(t, v.GetType().GetSimple(), v2.GetType().GetSimple(), "for %s, types do not match", k)
			}
		}
	}
}

func TestDataVolume(t *testing.T) {
	v := DataVolume("x", nil)
	assert.Equal(t, "x", v.Name)
	assert.NotNil(t, v.EmptyDir)
	assert.Nil(t, v.EmptyDir.SizeLimit)
	assert.Equal(t, v1.StorageMediumDefault, v.EmptyDir.Medium)

	q := resource.MustParse("1024Mi")
	v = DataVolume("x", &q)
	assert.NotNil(t, v.EmptyDir.SizeLimit)
	assert.Equal(t, q, *v.EmptyDir.SizeLimit)
}

func assertContainerHasVolumeMounts(t *testing.T, cfg config.FlyteCoPilotConfig, pilot *core.DataLoadingConfig, iFace *core.TypedInterface, c *v1.Container) {
	if iFace != nil {
		vmap := map[string]v1.VolumeMount{}
		for _, v := range c.VolumeMounts {
			vmap[v.Name] = v
		}
		if iFace.GetInputs() != nil {
			path := cfg.DefaultInputDataPath
			if pilot.GetInputPath() != "" {
				path = pilot.GetInputPath()
			}
			v, found := vmap[cfg.InputVolumeName]
			assert.Equal(t, path, v.MountPath, "Input Path does not match")
			assert.True(t, found, "Input volume mount expected but not found!")
		}

		if iFace.GetOutputs() != nil {
			path := cfg.DefaultOutputPath
			if pilot.GetOutputPath() != "" {
				path = pilot.GetOutputPath()
			}
			v, found := vmap[cfg.OutputVolumeName]
			assert.Equal(t, path, v.MountPath, "Output Path does not match")
			assert.True(t, found, "Output volume mount expected but not found!")
		}
	} else {
		assert.Len(t, c.VolumeMounts, 0)
	}
}

func assertPodHasCoPilot(t *testing.T, cfg config.FlyteCoPilotConfig, pilot *core.DataLoadingConfig, iFace *core.TypedInterface, pod *v1.PodSpec) {
	for _, c := range pod.Containers {
		if c.Name == "test" {
			cntr := c
			assertContainerHasVolumeMounts(t, cfg, pilot, iFace, &cntr)
		} else {
			if c.Name == cfg.NamePrefix+flyteDownloaderContainerName || c.Name == cfg.NamePrefix+flyteSidecarContainerName {
				if iFace != nil {
					vmap := map[string]v1.VolumeMount{}
					for _, v := range c.VolumeMounts {
						vmap[v.Name] = v
					}
					if iFace.GetInputs() != nil {
						path := cfg.DefaultInputDataPath
						if pilot != nil {
							path = pilot.GetInputPath()
						}
						v, found := vmap[cfg.InputVolumeName]
						if c.Name == cfg.NamePrefix+flyteDownloaderContainerName {
							assert.Equal(t, path, v.MountPath, "Input Path does not match")
							assert.True(t, found, "Input volume mount expected but not found!")
						} else {
							assert.False(t, found, "Input volume mount not expected but found!")
						}
					}

					if iFace.GetOutputs() != nil {
						path := cfg.DefaultOutputPath
						if pilot != nil {
							path = pilot.GetOutputPath()
						}
						v, found := vmap[cfg.OutputVolumeName]
						if c.Name == cfg.NamePrefix+flyteDownloaderContainerName {
							assert.False(t, found, "Output volume mount not expected but found on init container!")
						} else {
							assert.Equal(t, path, v.MountPath, "Output Path does not match")
							assert.True(t, found, "Output volume mount expected but not found!")
						}
					}

				} else {
					assert.Len(t, c.VolumeMounts, 0)
				}
			}
		}
	}
}

func TestCalculateStorageSize(t *testing.T) {
	twoG := resource.MustParse("2048Mi")
	oneG := resource.MustParse("1024Mi")
	tests := []struct {
		name string
		args *v1.ResourceRequirements
		want *resource.Quantity
	}{
		{"nil", nil, nil},
		{"empty", &v1.ResourceRequirements{}, nil},
		{"limits", &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceStorage: twoG,
			}}, &twoG},
		{"requests", &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceStorage: oneG,
			}}, &oneG},

		{"max", &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceStorage: twoG,
			},
			Requests: v1.ResourceList{
				v1.ResourceStorage: oneG,
			}}, &twoG},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateStorageSize(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateStorageSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddCoPilotToContainer(t *testing.T) {
	ctx := context.TODO()
	cfg := config.FlyteCoPilotConfig{
		NamePrefix:           "test-",
		Image:                "test",
		DefaultInputDataPath: "/in",
		DefaultOutputPath:    "/out",
		InputVolumeName:      "inp",
		OutputVolumeName:     "out",
		CPU:                  "1024m",
		Memory:               "1024Mi",
	}

	t.Run("dataload-config-nil", func(t *testing.T) {
		pilot := &core.DataLoadingConfig{}
		assert.NoError(t, AddCoPilotToContainer(ctx, cfg, nil, nil, pilot))
	})

	t.Run("disabled", func(t *testing.T) {
		pilot := &core.DataLoadingConfig{}
		assert.NoError(t, AddCoPilotToContainer(ctx, cfg, nil, nil, pilot))
	})

	t.Run("nil-iface", func(t *testing.T) {
		c := v1.Container{}
		pilot := &core.DataLoadingConfig{Enabled: true}
		assert.NoError(t, AddCoPilotToContainer(ctx, cfg, &c, nil, pilot))
		assertContainerHasVolumeMounts(t, cfg, pilot, nil, &c)
	})

	t.Run("happy-iface-empty-config", func(t *testing.T) {

		c := v1.Container{}
		iface := &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"x": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
					"y": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"o": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{Enabled: true}
		assert.NoError(t, AddCoPilotToContainer(ctx, cfg, &c, iface, pilot))
		assertContainerHasVolumeMounts(t, cfg, pilot, iface, &c)
	})

	t.Run("happy-iface-set-config", func(t *testing.T) {

		c := v1.Container{}
		iface := &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"x": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
					"y": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"o": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		assert.NoError(t, AddCoPilotToContainer(ctx, cfg, &c, iface, pilot))
		assertContainerHasVolumeMounts(t, cfg, pilot, iface, &c)
	})

	t.Run("happy-iface-inputs", func(t *testing.T) {

		c := v1.Container{}
		iface := &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"x": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
					"y": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		assert.NoError(t, AddCoPilotToContainer(ctx, cfg, &c, iface, pilot))
		assertContainerHasVolumeMounts(t, cfg, pilot, iface, &c)
	})

	t.Run("happy-iface-outputs", func(t *testing.T) {

		c := v1.Container{}
		iface := &core.TypedInterface{
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"o": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		assert.NoError(t, AddCoPilotToContainer(ctx, cfg, &c, iface, pilot))
		assertContainerHasVolumeMounts(t, cfg, pilot, iface, &c)
	})
}

func TestAddCoPilotToPod(t *testing.T) {
	ctx := context.TODO()
	cfg := config.FlyteCoPilotConfig{
		NamePrefix:           "test-",
		Image:                "test",
		DefaultInputDataPath: "/in",
		DefaultOutputPath:    "/out",
		InputVolumeName:      "inp",
		OutputVolumeName:     "out",
		StartTimeout: config2.Duration{
			Duration: time.Second * 1,
		},
		CPU:    "1024m",
		Memory: "1024Mi",
	}

	taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
	taskMetadata.EXPECT().GetNamespace().Return("test-namespace")
	taskMetadata.EXPECT().GetAnnotations().Return(map[string]string{"annotation-1": "val1"})
	taskMetadata.EXPECT().GetLabels().Return(map[string]string{"label-1": "val1"})
	taskMetadata.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskMetadata.EXPECT().GetK8sServiceAccount().Return("")
	taskMetadata.EXPECT().GetOwnerID().Return(types.NamespacedName{
		Namespace: "test-namespace",
		Name:      "test-owner-name",
	})
	taskMetadata.EXPECT().IsInterruptible().Return(false)

	tID := &pluginsCoreMock.TaskExecutionID{}
	tID.EXPECT().GetID().Return(core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			Name: "my-task",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.EXPECT().GetGeneratedName().Return("name")
	taskMetadata.EXPECT().GetTaskExecutionID().Return(tID)

	to := &pluginsCoreMock.TaskOverrides{}
	to.EXPECT().GetResources().Return(resourceRequirements)
	taskMetadata.EXPECT().GetOverrides().Return(to)

	inputPaths := &pluginsIOMock.InputFilePaths{}
	inputs := "/base/inputs"
	inputPaths.EXPECT().GetInputPrefixPath().Return(storage.DataReference(inputs))
	inputPaths.EXPECT().GetInputPath().Return(storage.DataReference(inputs + "/inputs.pb"))

	opath := &pluginsIOMock.OutputFilePaths{}
	opath.EXPECT().GetRawOutputPrefix().Return("/raw")
	opath.EXPECT().GetOutputPrefixPath().Return("/output")

	t.Run("happy", func(t *testing.T) {
		pod := v1.PodSpec{}
		iface := &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"x": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
					"y": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"o": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		primaryInitContainerName, err := AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot)
		assert.NoError(t, err)
		assert.Equal(t, "test-downloader", primaryInitContainerName)
		assertPodHasCoPilot(t, cfg, pilot, iface, &pod)
	})

	t.Run("happy-nil-iface", func(t *testing.T) {
		pod := v1.PodSpec{}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		primaryInitContainerName, err := AddCoPilotToPod(ctx, cfg, &pod, nil, taskMetadata, inputPaths, opath, pilot)
		assert.NoError(t, err)
		assert.Empty(t, primaryInitContainerName)
		assertPodHasCoPilot(t, cfg, pilot, nil, &pod)
	})

	t.Run("happy-inputs-only", func(t *testing.T) {
		pod := v1.PodSpec{}
		iface := &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"x": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
					"y": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		primaryInitContainerName, err := AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot)
		assert.NoError(t, err)
		assert.Equal(t, "test-downloader", primaryInitContainerName)
		assertPodHasCoPilot(t, cfg, pilot, iface, &pod)
	})

	t.Run("happy-outputs-only", func(t *testing.T) {
		pod := v1.PodSpec{}
		iface := &core.TypedInterface{
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"o": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		primaryInitContainerName, err := AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot)
		assert.NoError(t, err)
		assert.Empty(t, primaryInitContainerName)
		assertPodHasCoPilot(t, cfg, pilot, iface, &pod)
	})

	t.Run("disabled", func(t *testing.T) {
		pod := v1.PodSpec{}
		iface := &core.TypedInterface{
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"o": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    false,
			InputPath:  "in",
			OutputPath: "out",
		}
		primaryInitContainerName, err := AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot)
		assert.NoError(t, err)
		assert.Empty(t, primaryInitContainerName)
		assert.Len(t, pod.Volumes, 0)
	})

	t.Run("nil", func(t *testing.T) {
		primaryInitContainerName, err := AddCoPilotToPod(ctx, cfg, nil, nil, taskMetadata, inputPaths, opath, nil)
		assert.NoError(t, err)
		assert.Empty(t, primaryInitContainerName)
	})
}
