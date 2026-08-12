package flytek8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
		CPU:                  "1024m",
		Memory:               "1024Mi",
		CopilotStorageConfig: "flyte-copilot-storage-config-secret",
	}

	t.Run("happy stow backend", func(t *testing.T) {
		storage.GetConfig().Stow.Kind = "S3"
		storage.GetConfig().Stow.Config = map[string]string{
			"path": "config.yaml",
		}
		c, err := FlyteCoPilotContainer("x", cfg, []string{"hello"})
		assert.NoError(t, err)

		expectedCommand, err := CopilotCommandArgs(storage.GetConfig(), cfg.CopilotStorageConfig, false)
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

	t.Run("configured sources keep the whole stow config off the command line", func(t *testing.T) {
		storage.GetConfig().Type = storage.TypeStow
		storage.GetConfig().InitContainer = "bucket"
		storage.GetConfig().Stow.Kind = "s3"
		storage.GetConfig().Stow.Config = map[string]string{
			"access_key_id":   "AKIAEXAMPLE12345",
			"secret_key":      "s3-secret-value",
			"secret_key_path": "/etc/flyte/storage/secret_key",
			"auth_type":       "accesskey",
			"region":          "us-east-1",
			"endpoint":        "http://minio:9000",
		}

		command, err := CopilotCommandArgs(storage.GetConfig(), cfg.CopilotStorageConfig, false)
		assert.NoError(t, err)

		joined := strings.Join(command, " ")
		assert.NotContains(t, joined, "AKIAEXAMPLE12345")
		assert.NotContains(t, joined, "s3-secret-value")
		assert.NotContains(t, joined, "/etc/flyte/storage/secret_key")

		// All-or-nothing: not even the non-sensitive keys travel on the command line.
		assert.NotContains(t, joined, "--storage.stow.config")
		assert.NotContains(t, joined, "region=us-east-1")
		assert.NotContains(t, joined, "endpoint=http://minio:9000")

		// flytestdlib globs the pattern itself; the container has no shell to expand it.
		assert.Contains(t, joined, "--config /etc/flyte/copilot/*.yaml")

		// Scalar flags still travel, and still override the mounted config key by key.
		assert.Contains(t, joined, "--storage.container=bucket")
		assert.Contains(t, joined, "--storage.limits.maxDownloadMBs=0")
	})

	t.Run("unconfigured sources fall back to the old rendering", func(t *testing.T) {
		// Backwards compatibility: a deployment that has not adopted the new setting keeps
		// working, still exposed, rather than failing to launch every ContainerTask.
		storage.GetConfig().Type = storage.TypeStow
		storage.GetConfig().InitContainer = "bucket"
		storage.GetConfig().Stow.Kind = "s3"
		storage.GetConfig().Stow.Config = map[string]string{
			"access_key_id":   "AKIAFALLBACK",
			"secret_key":      "fallback-secret",
			"secret_key_path": "/etc/flyte/storage/secret_key",
			"auth_type":       "accesskey",
			"region":          "us-east-1",
		}

		command, err := CopilotCommandArgs(storage.GetConfig(), "", false)
		assert.NoError(t, err)

		joined := strings.Join(command, " ")
		assert.Contains(t, joined, "access_key_id=AKIAFALLBACK")
		assert.Contains(t, joined, "secret_key=fallback-secret")
		assert.Contains(t, joined, "auth_type=accesskey")
		assert.NotContains(t, joined, "--config")
		// secret_key_path is dropped even here: it names a file that does not exist in a
		// task pod, and forwarding it stops co-pilot from starting at all.
		assert.NotContains(t, joined, "secret_key_path")

		bare := cfg
		bare.CopilotStorageConfig = ""
		c, err := FlyteCoPilotContainer("x", bare, []string{"hello"})
		assert.NoError(t, err)
		assert.Empty(t, c.VolumeMounts, "nothing to mount when no source is configured")
	})

	t.Run("the Secret mounts for backends that need no credentials", func(t *testing.T) {
		// S3 authType=iam and gcs/azure carry no credentials, but the storage settings
		// still reach co-pilot through the same Secret — the projection does not vary by
		// backend.
		storage.GetConfig().Type = storage.TypeStow
		storage.GetConfig().InitContainer = "bucket"
		storage.GetConfig().Stow.Kind = "s3"
		storage.GetConfig().Stow.Config = map[string]string{
			"auth_type": "iam",
			"region":    "us-east-1",
		}

		secretName := "flyte-copilot-storage-config-secret"

		command, err := CopilotCommandArgs(storage.GetConfig(), secretName, false)
		assert.NoError(t, err)
		assert.Contains(t, strings.Join(command, " "), "--config /etc/flyte/copilot/*.yaml")

		vol := storageConfigVolume(secretName)
		assert.Len(t, vol.Projected.Sources, 1)
		assert.Equal(t, secretName, vol.Projected.Sources[0].Secret.Name)
		// No Items: the Secret holds co-pilot's config, so every key in it is projected
		// and the *.yaml glob decides what is read.
		assert.Nil(t, vol.Projected.Sources[0].Secret.Items)
	})

	t.Run("non-s3 backends are mounted the same way", func(t *testing.T) {
		storage.GetConfig().Type = storage.TypeStow
		storage.GetConfig().InitContainer = "bucket"
		storage.GetConfig().Stow.Kind = "google"
		storage.GetConfig().Stow.Config = map[string]string{
			"json":       "service-account-key",
			"project_id": "flyte-gcp",
		}

		command, err := CopilotCommandArgs(storage.GetConfig(), cfg.CopilotStorageConfig, false)
		assert.NoError(t, err)

		joined := strings.Join(command, " ")
		assert.NotContains(t, joined, "service-account-key")
		assert.NotContains(t, joined, "project_id=flyte-gcp")
		assert.Contains(t, joined, "--config /etc/flyte/copilot/*.yaml")
	})

	t.Run("storage override puts the stow config back on the command line", func(t *testing.T) {
		storageConfigOverride := storage.Config{}

		storageConfigOverride.Type = storage.TypeStow
		storageConfigOverride.InitContainer = "bucket"
		storageConfigOverride.Stow.Kind = "google"
		storageConfigOverride.Stow.Config = map[string]string{
			"json":       "",
			"project_id": "flyte-gcp",
		}
		cfg.StorageConfigOverride = &storageConfigOverride
		defer func() { cfg.StorageConfigOverride = nil }()

		c, err := FlyteCoPilotContainer("x", cfg, []string{"hello"}, v1.VolumeMount{Name: "X", MountPath: "/"})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(c.VolumeMounts))

		expectedCommand, err := CopilotCommandArgs(&storageConfigOverride, cfg.CopilotStorageConfig, true)
		assert.NoError(t, err)
		assert.ElementsMatch(t, c.Command, expectedCommand)

		// The override cannot be expressed in the mounted config, so honouring it means
		// taking all of stow.config from the command line.
		joined := strings.Join(c.Command, " ")
		assert.Contains(t, joined, "project_id=flyte-gcp")
		assert.Contains(t, joined, "--storage.stow.kind=google")
		assert.Contains(t, joined, "--config /etc/flyte/copilot/*.yaml")
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

// Regression guard for the credential leak. It scans the serialized spec for the values
// rather than checking known field names, so a backend whose secret lives under a key
// nobody thought to exclude (Azure's "key", GCP's "json") is covered too.
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
		CopilotStorageConfig: "flyte-copilot-storage-config-secret",
	}

	secrets := map[string]string{
		"access_key_id": "AKIAEXAMPLE12345",
		"secret_key":    "s3-secret-value",
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

	var projected []v1.Volume
	for _, v := range pod.Volumes {
		if v.Name == copilotConfigVolumeName {
			projected = append(projected, v)
		}
	}
	assert.Len(t, projected, 1, "the storage config volume must be added exactly once")
	assert.Equal(t, int32(0400), *projected[0].Projected.DefaultMode)
	assert.Len(t, projected[0].Projected.Sources, 1, "the whole configuration lives in one Secret")
	assert.Equal(t, "flyte-copilot-storage-config-secret", projected[0].Projected.Sources[0].Secret.Name)
	assert.Nil(t, projected[0].Projected.Sources[0].Secret.Items)

	// Only the co-pilot containers: the primary one runs user code.
	hasConfigMount := func(c v1.Container) bool {
		for _, m := range c.VolumeMounts {
			if m.Name == copilotConfigVolumeName {
				assert.True(t, m.ReadOnly, "container %q mounts the storage config writable", c.Name)
				assert.Equal(t, "/etc/flyte/copilot", m.MountPath)
				return true
			}
		}
		return false
	}
	for _, c := range pod.InitContainers {
		assert.True(t, hasConfigMount(c), "co-pilot container %q has no storage config", c.Name)
	}
	for _, c := range pod.Containers {
		assert.False(t, hasConfigMount(c),
			"primary container %q must not receive the storage config", c.Name)
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
		CPU:                  "1024m",
		Memory:               "1024Mi",
		CopilotStorageConfig: "flyte-copilot-storage-config-secret",
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
