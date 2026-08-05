package flytek8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	pluginsCoreMock "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginsIOMock "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	config2 "github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/flytestdlib/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
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
		StorageConfig: config.StorageConfigSources{
			ConfigMapName: "flyte-config",
			ConfigMapKeys: []string{"003-storage.yaml"},
			SecretName:    "flyte-config-secret",
			SecretKeys:    []string{"013-storage-secrets.yaml"},
			MountPath:     "/etc/flyte/copilot",
		},
	}

	t.Run("happy stow backend", func(t *testing.T) {
		storage.GetConfig().Stow.Kind = "S3"
		storage.GetConfig().Stow.Config = map[string]string{
			"path": "config.yaml",
		}
		c, err := FlyteCoPilotContainer("x", cfg, []string{"hello"})
		assert.NoError(t, err)

		expectedCommand, err := CopilotCommandArgs(storage.GetConfig(), cfg.StorageConfig.ConfigGlob())
		assert.NoError(t, err)

		assert.Equal(t, "test-x", c.Name)
		assert.Equal(t, "test", c.Image)
		assert.Equal(t, expectedCommand, c.Command)
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
			"json":       "service-account-key",
			"project_id": "flyte-gcp",
			"scope":      "read_write",
		}

		command, err := CopilotCommandArgs(storage.GetConfig(), cfg.StorageConfig.ConfigGlob())
		assert.NoError(t, err)

		// Every backend keeps its credentials off the command line, not just S3: the
		// stow config is never rendered here, so a backend whose secret lives under a
		// different key (GCP's service-account JSON) is covered without enumerating it.
		assert.NotContains(t, command, "--storage.stow.config")
		for _, val := range storage.GetConfig().Stow.Config {
			for _, arg := range command {
				assert.NotContains(t, arg, val)
			}
		}
		assert.Contains(t, command, "--config")
		assert.Contains(t, command, "/etc/flyte/copilot/*.yaml")
	})

	t.Run("unconfigured storage config secret is rejected", func(t *testing.T) {
		// Failing loudly is deliberate: falling back to the old flag-rendered config
		// would put the storage credentials back into every task's pod spec.
		_, err := CopilotCommandArgs(storage.GetConfig(), "")
		assert.Error(t, err)

		bare := cfg
		bare.StorageConfig = config.StorageConfigSources{}
		_, err = FlyteCoPilotContainer("x", bare, []string{"hello"})
		assert.Error(t, err)
	})

	t.Run("storage override", func(t *testing.T) {

		storageConfigOverride := storage.Config{}

		storageConfigOverride.Type = storage.TypeStow
		storageConfigOverride.InitContainer = "bucket"
		storageConfigOverride.Stow.Kind = "google"
		storageConfigOverride.Stow.Config = map[string]string{
			"json":       "",
			"project_id": "flyte-gcp",
		}
		cfg.StorageConfigOverride = &storageConfigOverride

		c, err := FlyteCoPilotContainer("x", cfg, []string{"hello"}, v1.VolumeMount{Name: "X", MountPath: "/"})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(c.VolumeMounts))

		expectedCommand, err := CopilotCommandArgs(&storageConfigOverride, cfg.StorageConfig.ConfigGlob())
		assert.NoError(t, err)

		assert.ElementsMatch(t, c.Command, expectedCommand)
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
}

func TestDownloadCommandArgs(t *testing.T) {
	_, err := DownloadCommandArgs("", "", "", core.DataLoadingConfig_YAML, core.DataLoadingConfig_DIRECT, nil)
	assert.Error(t, err)

	iFace := &core.VariableMap{
		Variables: []*core.VariableEntry{
			{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
			{Key: "y", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
		},
	}
	d, err := DownloadCommandArgs("s3://from", "s3://output-meta", "/to", core.DataLoadingConfig_JSON, core.DataLoadingConfig_NAMED_DIR, iFace)
	assert.NoError(t, err)
	expected := []string{"download", "--from-remote", "s3://from", "--to-output-prefix", "s3://output-meta", "--to-local-dir", "/to", "--format", "JSON", "--file-input-layout", "NAMED_DIR", "--input-interface", "<interface>"}
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
			assert.Len(t, vm.Variables, 2)
			for _, entry := range iFace.Variables {
				v2 := utils.GetVariable(vm, entry.Key)
				assert.NotNil(t, v2, "variable %s should exist", entry.Key)
				assert.Equal(t, entry.Value.Type.GetSimple(), v2.Type.GetSimple(), "for %s, types do not match", entry.Key)
			}
		}
	}
}

func TestSidecarCommandArgs(t *testing.T) {
	_, err := SidecarCommandArgs("", "", "", time.Second*10, nil)
	assert.Error(t, err)

	iFace := &core.TypedInterface{
		Outputs: &core.VariableMap{
			Variables: []*core.VariableEntry{
				{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				{Key: "y", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
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
			assert.Len(t, if2.Outputs.Variables, 2)
			for _, entry := range iFace.Outputs.Variables {
				v2 := utils.GetVariable(if2.Outputs, entry.Key)
				assert.NotNil(t, v2, "variable %s should exist", entry.Key)
				assert.Equal(t, entry.Value.Type.GetSimple(), v2.Type.GetSimple(), "for %s, types do not match", entry.Key)
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
		if iFace.Inputs != nil {
			path := cfg.DefaultInputDataPath
			if pilot.InputPath != "" {
				path = pilot.InputPath
			}
			v, found := vmap[cfg.InputVolumeName]
			assert.Equal(t, path, v.MountPath, "Input Path does not match")
			assert.True(t, found, "Input volume mount expected but not found!")
		}

		if iFace.Outputs != nil {
			path := cfg.DefaultOutputPath
			if pilot.OutputPath != "" {
				path = pilot.OutputPath
			}
			v, found := vmap[cfg.OutputVolumeName]
			assert.Equal(t, path, v.MountPath, "Output Path does not match")
			assert.True(t, found, "Output volume mount expected but not found!")
		}
	} else {
		assert.Len(t, c.VolumeMounts, 0)
	}
}

// TestAddCoPilotToPod_StorageCredentialsNeverInPodSpec is the regression guard for the
// credential leak: the rendered pod must not contain any value from the stow config,
// whichever backend is configured. It scans the serialized spec for the values rather
// than checking known field names, so a backend whose secret lives under a key nobody
// thought to exclude (Azure's "key", GCP's "json") is covered too.
func TestAddCoPilotToPod_StorageCredentialsNeverInPodSpec(t *testing.T) {
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
		StorageConfig: config.StorageConfigSources{
			ConfigMapName: "flyte-config",
			ConfigMapKeys: []string{"003-storage.yaml"},
			SecretName:    "flyte-config-secret",
			SecretKeys:    []string{"013-storage-secrets.yaml"},
			MountPath:     "/etc/flyte/copilot",
		},
	}

	secrets := map[string]string{
		"access_key_id": "AKIAEXAMPLE12345",
		"secret_key":    "s3-secret-value",
		"key":           "azure-account-key",
		"json":          "gcp-service-account-json",
		"password":      "oracle-password",
	}
	original := storage.GetConfig().Stow.Config
	defer func() { storage.GetConfig().Stow.Config = original }()
	storage.GetConfig().Stow.Config = secrets
	storage.GetConfig().Stow.Kind = "s3"

	taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
	overrides := &pluginsCoreMock.TaskOverrides{}
	overrides.EXPECT().GetResources().Return(resourceRequirements)
	taskMetadata.EXPECT().GetOverrides().Return(overrides)
	taskExecutionID := &pluginsCoreMock.TaskExecutionID{}
	taskExecutionID.EXPECT().GetID().Return(&core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{Name: "task"},
	})
	taskMetadata.EXPECT().GetTaskExecutionID().Return(taskExecutionID)

	inputPaths := &pluginsIOMock.InputFilePaths{}
	inputPaths.EXPECT().GetInputPath().Return("s3://input/inputs.pb")
	opath := &pluginsIOMock.OutputFilePaths{}
	opath.EXPECT().GetRawOutputPrefix().Return("s3://raw")
	opath.EXPECT().GetOutputPrefixPath().Return("s3://output")

	pod := v1.PodSpec{}
	iface := &core.TypedInterface{
		Inputs: &core.VariableMap{Variables: []*core.VariableEntry{
			{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
		}},
		Outputs: &core.VariableMap{Variables: []*core.VariableEntry{
			{Key: "o", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
		}},
	}
	pilot := &core.DataLoadingConfig{Enabled: true, InputPath: "in", OutputPath: "out"}

	assert.NoError(t, AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot))

	rendered, err := json.Marshal(pod)
	assert.NoError(t, err)
	for name, value := range secrets {
		assert.NotContains(t, string(rendered), value,
			"stow config value for %q leaked into the pod spec", name)
	}

	// The credentials are absent because co-pilot reads them from a mounted Secret, and
	// that mount must reach the co-pilot containers only.
	// Exactly one. This task has both inputs and outputs, so both co-pilot containers are
	// added and each needs the config; adding the volume per container instead of per pod
	// would be a duplicate volume name and the API server would reject the whole pod.
	var volume *v1.Volume
	volumeCount := 0
	for i := range pod.Volumes {
		if pod.Volumes[i].Name == copilotConfigVolumeName {
			volume = &pod.Volumes[i]
			volumeCount++
		}
	}
	assert.Equal(t, 1, volumeCount, "storage config volume must be added exactly once")
	if assert.NotNil(t, volume, "storage config volume missing") {
		if assert.NotNil(t, volume.Projected, "both config sources must be projected together") {
			assert.Len(t, volume.Projected.Sources, 2)
			cm := volume.Projected.Sources[0].ConfigMap
			sec := volume.Projected.Sources[1].Secret
			assert.Equal(t, cfg.StorageConfig.ConfigMapName, cm.Name)
			assert.Equal(t, cfg.StorageConfig.SecretName, sec.Name)
			// Only the listed keys: both objects also hold entries co-pilot must not see
			// (plugin config, database credentials) and cannot parse under strict mode.
			assert.Len(t, cm.Items, len(cfg.StorageConfig.ConfigMapKeys))
			assert.Len(t, sec.Items, len(cfg.StorageConfig.SecretKeys))
		}
	}
	for _, c := range pod.InitContainers {
		mounts := make([]string, 0, len(c.VolumeMounts))
		for _, m := range c.VolumeMounts {
			mounts = append(mounts, m.Name)
		}
		assert.Contains(t, mounts, copilotConfigVolumeName, "co-pilot container %q cannot read its config", c.Name)
	}
	for _, c := range pod.Containers {
		for _, m := range c.VolumeMounts {
			assert.NotEqual(t, copilotConfigVolumeName, m.Name,
				"primary container %q must not be able to read the storage credentials", c.Name)
		}
	}
}

func assertPodHasCoPilot(t *testing.T, cfg config.FlyteCoPilotConfig, pilot *core.DataLoadingConfig, iFace *core.TypedInterface, pod *v1.PodSpec) {
	containers := append(pod.Containers, pod.InitContainers...)
	for _, c := range containers {
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
					if iFace.Inputs != nil {
						path := cfg.DefaultInputDataPath
						if pilot != nil {
							path = pilot.InputPath
						}
						v, found := vmap[cfg.InputVolumeName]
						if c.Name == cfg.NamePrefix+flyteDownloaderContainerName {
							assert.Equal(t, path, v.MountPath, "Input Path does not match")
							assert.True(t, found, "Input volume mount expected but not found!")
						} else {
							assert.False(t, found, "Input volume mount not expected but found!")
						}
					}

					if iFace.Outputs != nil {
						path := cfg.DefaultOutputPath
						if pilot != nil {
							path = pilot.OutputPath
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
				Variables: []*core.VariableEntry{
					{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
					{Key: "y", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
			Outputs: &core.VariableMap{
				Variables: []*core.VariableEntry{
					{Key: "o", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
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
				Variables: []*core.VariableEntry{
					{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
					{Key: "y", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
			Outputs: &core.VariableMap{
				Variables: []*core.VariableEntry{
					{Key: "o", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
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
				Variables: []*core.VariableEntry{
					{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
					{Key: "y", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
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
				Variables: []*core.VariableEntry{
					{Key: "o", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
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
		StorageConfig: config.StorageConfigSources{
			ConfigMapName: "flyte-config",
			ConfigMapKeys: []string{"003-storage.yaml"},
			SecretName:    "flyte-config-secret",
			SecretKeys:    []string{"013-storage-secrets.yaml"},
			MountPath:     "/etc/flyte/copilot",
		},
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
	tID.EXPECT().GetID().Return(&core.TaskExecutionIdentifier{
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
				Variables: []*core.VariableEntry{
					{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
					{Key: "y", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
			Outputs: &core.VariableMap{
				Variables: []*core.VariableEntry{
					{Key: "o", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		assert.NoError(t, AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot))
		assert.Equal(t, pod.InitContainers[0].Name, cfg.NamePrefix+flyteSidecarContainerName)
		assert.Equal(t, pod.InitContainers[1].Name, cfg.NamePrefix+flyteDownloaderContainerName)
		assertPodHasCoPilot(t, cfg, pilot, iface, &pod)
	})

	t.Run("nil-task-id", func(t *testing.T) {
		pod := v1.PodSpec{}
		iface := &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: []*core.VariableEntry{
					{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
			Outputs: &core.VariableMap{
				Variables: []*core.VariableEntry{
					{Key: "o", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		tID := &pluginsCoreMock.TaskExecutionID{}
		tID.EXPECT().GetID().Return(&core.TaskExecutionIdentifier{})
		metadata := &pluginsCoreMock.TaskExecutionMetadata{}
		metadata.EXPECT().GetTaskExecutionID().Return(tID)
		overrides := &pluginsCoreMock.TaskOverrides{}
		overrides.EXPECT().GetResources().Return(resourceRequirements)
		metadata.EXPECT().GetOverrides().Return(overrides)

		inputPaths := &pluginsIOMock.InputFilePaths{}
		inputPaths.EXPECT().GetInputPath().Return(storage.DataReference("/base/inputs/inputs.pb"))

		outputPaths := &pluginsIOMock.OutputFilePaths{}
		outputPaths.EXPECT().GetOutputPrefixPath().Return(storage.DataReference("/output"))
		outputPaths.EXPECT().GetRawOutputPrefix().Return(storage.DataReference("/raw"))

		var err error
		assert.NotPanics(t, func() {
			err = AddCoPilotToPod(ctx, cfg, &pod, iface, metadata, inputPaths, outputPaths, pilot)
		})
		assert.NoError(t, err)
		assertPodHasCoPilot(t, cfg, pilot, iface, &pod)
	})

	t.Run("happy-nil-iface", func(t *testing.T) {
		pod := v1.PodSpec{}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		assert.NoError(t, AddCoPilotToPod(ctx, cfg, &pod, nil, taskMetadata, inputPaths, opath, pilot))
		assertPodHasCoPilot(t, cfg, pilot, nil, &pod)
	})

	t.Run("happy-inputs-only", func(t *testing.T) {
		pod := v1.PodSpec{}
		iface := &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: []*core.VariableEntry{
					{Key: "x", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
					{Key: "y", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		assert.NoError(t, AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot))
		assertPodHasCoPilot(t, cfg, pilot, iface, &pod)
	})

	t.Run("happy-outputs-only", func(t *testing.T) {
		pod := v1.PodSpec{}
		iface := &core.TypedInterface{
			Outputs: &core.VariableMap{
				Variables: []*core.VariableEntry{
					{Key: "o", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		assert.NoError(t, AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot))
		assertPodHasCoPilot(t, cfg, pilot, iface, &pod)
	})

	t.Run("empty-iface-adds-no-config-volume", func(t *testing.T) {
		// A non-nil interface with neither inputs nor outputs adds no co-pilot container,
		// so it must add no storage config volume either — otherwise the pod would depend
		// on a Secret that nothing mounts, and would fail to start if it were absent.
		pod := v1.PodSpec{}
		iface := &core.TypedInterface{}
		pilot := &core.DataLoadingConfig{
			Enabled:    true,
			InputPath:  "in",
			OutputPath: "out",
		}
		assert.NoError(t, AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot))
		assert.Empty(t, pod.InitContainers)
		for _, v := range pod.Volumes {
			assert.NotEqual(t, copilotConfigVolumeName, v.Name)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		pod := v1.PodSpec{}
		iface := &core.TypedInterface{
			Outputs: &core.VariableMap{
				Variables: []*core.VariableEntry{
					{Key: "o", Value: &core.Variable{Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
		}
		pilot := &core.DataLoadingConfig{
			Enabled:    false,
			InputPath:  "in",
			OutputPath: "out",
		}
		assert.NoError(t, AddCoPilotToPod(ctx, cfg, &pod, iface, taskMetadata, inputPaths, opath, pilot))
		assert.Len(t, pod.Volumes, 0)
	})

	t.Run("nil", func(t *testing.T) {
		assert.NoError(t, AddCoPilotToPod(ctx, cfg, nil, nil, taskMetadata, inputPaths, opath, nil))
	})
}
