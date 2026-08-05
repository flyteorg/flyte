package flytek8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto" //nolint: staticcheck
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	core2 "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	flyteSidecarContainerName    = "uploader"
	flyteDownloaderContainerName = "downloader"
	copilotConfigVolumeName      = "flyte-copilot-config"
)

func FlyteCoPilotContainer(name string, cfg config.FlyteCoPilotConfig, args []string, volumeMounts ...v1.VolumeMount) (v1.Container, error) {
	cpu, err := resource.ParseQuantity(cfg.CPU)
	if err != nil {
		return v1.Container{}, err
	}

	mem, err := resource.ParseQuantity(cfg.Memory)
	if err != nil {
		return v1.Container{}, err
	}

	var storageCfg *storage.Config
	if cfg.StorageConfigOverride != nil {
		storageCfg = cfg.StorageConfigOverride
	} else {
		storageCfg = storage.GetConfig()
	}

	command, err := CopilotCommandArgs(storageCfg, cfg.StorageConfig.ConfigGlob())
	if err != nil {
		return v1.Container{}, err
	}

	return v1.Container{
		Name:       cfg.NamePrefix + name,
		Image:      cfg.Image,
		Command:    command,
		Args:       args,
		WorkingDir: "/",
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    cpu,
				v1.ResourceMemory: mem,
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    cpu,
				v1.ResourceMemory: mem,
			},
		},
		VolumeMounts:             volumeMounts,
		TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		ImagePullPolicy:          v1.PullIfNotPresent,
	}, nil
}

// CopilotCommandArgs builds the co-pilot entrypoint. The stow configuration — kind,
// endpoint, and credentials alike — is deliberately NOT passed here; co-pilot reads it
// from the files AddCoPilotToPod mounts. Rendering it into the command would publish the
// storage credentials in every task's pod spec, where any principal that can read pods
// could recover them without access to the Secret holding them.
//
// The whole stow config must travel together via those files. Splitting it — non-sensitive
// keys as flags, credentials in the file — does not work: `storage.stow.config` binds to
// a pflag StringToString, and once that flag is set viper resolves the key entirely from
// it rather than merging into the file's map, silently dropping the credentials.
func CopilotCommandArgs(storageConfig *storage.Config, storageConfigGlob string) ([]string, error) {
	if storageConfigGlob == "" {
		return nil, fmt.Errorf("co-pilot storage-config is not configured: set mount-path and at least " +
			"one of config-map-name/secret-name so the storage config reaches co-pilot as mounted files " +
			"instead of on the command line")
	}

	var commands = []string{
		"/bin/flyte-copilot",
		"--storage.limits.maxDownloadMBs=0",
		"--logger.level=" + strconv.Itoa(logger.GetConfig().Level),
	}
	if storageConfig.MultiContainerEnabled {
		commands = append(commands, "--storage.enable-multicontainer")
	}
	if len(storageConfig.InitContainer) > 0 {
		commands = append(commands, fmt.Sprintf("--storage.container=%s", storageConfig.InitContainer))

	}
	commands = append(commands, fmt.Sprintf("--storage.type=%s", storageConfig.Type))

	return append(commands, "--config", storageConfigGlob), nil
}

func SidecarCommandArgs(fromLocalPath string, outputPrefix, rawOutputPath storage.DataReference, uploadTimeout time.Duration, iface *core.TypedInterface) ([]string, error) {
	if iface == nil {
		return nil, fmt.Errorf("interface is required for CoPilot Sidecar")
	}
	b, err := proto.Marshal(iface)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal given core.TypedInterface")
	}
	return []string{
		"sidecar",
		"--timeout",
		uploadTimeout.String(),
		"--to-raw-output",
		rawOutputPath.String(),
		"--to-output-prefix",
		outputPrefix.String(),
		"--from-local-dir",
		fromLocalPath,
		"--interface",
		base64.StdEncoding.EncodeToString(b),
	}, nil
}

func DownloadCommandArgs(fromInputsPath, outputPrefix storage.DataReference, toLocalPath string, format core.DataLoadingConfig_LiteralMapFormat, layout core.DataLoadingConfig_FileInputLayout, inputInterface *core.VariableMap) ([]string, error) {
	if inputInterface == nil {
		return nil, fmt.Errorf("input Interface is required for CoPilot Downloader")
	}
	b, err := proto.Marshal(inputInterface)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal given input interface")
	}
	return []string{
		"download",
		"--from-remote",
		fromInputsPath.String(),
		"--to-output-prefix",
		outputPrefix.String(),
		"--to-local-dir",
		toLocalPath,
		"--format",
		format.String(),
		"--file-input-layout",
		layout.String(),
		"--input-interface",
		base64.StdEncoding.EncodeToString(b),
	}, nil
}

// keyToPaths projects each key under its own name, so the mounted filenames match the
// keys the deployment already uses and their numeric prefixes keep ordering the merge.
func keyToPaths(keys []string) []v1.KeyToPath {
	if len(keys) == 0 {
		return nil
	}
	items := make([]v1.KeyToPath, 0, len(keys))
	for _, k := range keys {
		items = append(items, v1.KeyToPath{Key: k, Path: k})
	}
	return items
}

// storageConfigVolume projects co-pilot's storage configuration into a read-only volume.
// Both sources are combined here rather than merged into one object, preserving the
// deployment's split of non-sensitive settings (ConfigMap) from credentials (Secret).
//
// Only the configured keys are projected. Both objects carry unrelated entries — plugin
// config, database credentials — which must neither reach the co-pilot containers nor be
// parsed by co-pilot's strict-mode config loader, which rejects sections it does not
// recognise. A source with no name is omitted entirely: a deployment using instance
// credentials (S3 authType=iam) renders no credential file, and naming a key that is
// never written would fail the mount.
func storageConfigVolume(cfg config.StorageConfigSources) v1.Volume {
	// Read-only for the owner: co-pilot only reads this, and the containers run as
	// whatever user the image declares.
	mode := int32(0400)
	sources := make([]v1.VolumeProjection, 0, 2)
	if cfg.ConfigMapName != "" {
		sources = append(sources, v1.VolumeProjection{
			ConfigMap: &v1.ConfigMapProjection{
				LocalObjectReference: v1.LocalObjectReference{Name: cfg.ConfigMapName},
				Items:                keyToPaths(cfg.ConfigMapKeys),
			},
		})
	}
	if cfg.SecretName != "" {
		sources = append(sources, v1.VolumeProjection{
			Secret: &v1.SecretProjection{
				LocalObjectReference: v1.LocalObjectReference{Name: cfg.SecretName},
				Items:                keyToPaths(cfg.SecretKeys),
			},
		})
	}
	return v1.Volume{
		Name: copilotConfigVolumeName,
		VolumeSource: v1.VolumeSource{
			Projected: &v1.ProjectedVolumeSource{
				Sources:     sources,
				DefaultMode: &mode,
			},
		},
	}
}

// storageConfigMount is the mount matching storageConfigVolume. It is attached to the
// co-pilot containers only — never to the primary container, which runs user code and
// must not be able to read the storage credentials.
func storageConfigMount(cfg config.StorageConfigSources) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      copilotConfigVolumeName,
		MountPath: cfg.MountPath,
		ReadOnly:  true,
	}
}

func DataVolume(name string, size *resource.Quantity) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				Medium:    v1.StorageMediumDefault,
				SizeLimit: size,
			},
		},
	}
}

func CalculateStorageSize(requirements *v1.ResourceRequirements) *resource.Quantity {
	if requirements == nil {
		return nil
	}
	s, ok := requirements.Limits[v1.ResourceStorage]
	if ok {
		return &s
	}
	s, ok = requirements.Requests[v1.ResourceStorage]
	if ok {
		return &s
	}
	return nil
}

func AddCoPilotToContainer(ctx context.Context, cfg config.FlyteCoPilotConfig, c *v1.Container, iFace *core.TypedInterface, pilot *core.DataLoadingConfig) error {
	if pilot == nil || !pilot.Enabled {
		return nil
	}
	logger.Infof(ctx, "Enabling CoPilot on main container [%s]", c.Name)
	if c.SecurityContext == nil {
		c.SecurityContext = &v1.SecurityContext{}
	}
	if c.SecurityContext.Capabilities == nil {
		c.SecurityContext.Capabilities = &v1.Capabilities{}
	}

	if iFace != nil {
		if iFace.Inputs != nil && len(iFace.Inputs.Variables) > 0 {
			inPath := cfg.DefaultInputDataPath
			if pilot.GetInputPath() != "" {
				inPath = pilot.GetInputPath()
			}

			c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
				Name:      cfg.InputVolumeName,
				MountPath: inPath,
			})
		}

		if iFace.Outputs != nil && len(iFace.Outputs.Variables) > 0 {
			outPath := cfg.DefaultOutputPath
			if pilot.GetOutputPath() != "" {
				outPath = pilot.GetOutputPath()
			}
			c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
				Name:      cfg.OutputVolumeName,
				MountPath: outPath,
			})
		}
	}
	return nil
}

func AddCoPilotToPod(ctx context.Context, cfg config.FlyteCoPilotConfig, coPilotPod *v1.PodSpec, iFace *core.TypedInterface, taskExecMetadata core2.TaskExecutionMetadata, inputPaths io.InputFilePaths, outputPaths io.OutputFilePaths, pilot *core.DataLoadingConfig) error {
	if pilot == nil || !pilot.Enabled {
		return nil
	}

	taskName := taskExecMetadata.GetTaskExecutionID().GetID().GetTaskId().GetName()
	logger.Infof(ctx, "CoPilot Enabled for task [%s]", taskName)

	cfgMount := storageConfigMount(cfg.StorageConfig)

	if iFace != nil {
		needsDownloader := iFace.Inputs != nil && len(iFace.Inputs.Variables) > 0
		needsUploader := iFace.Outputs != nil && len(iFace.Outputs.Variables) > 0

		// We only mount the volume when either downloader or uploader is required
		if needsDownloader || needsUploader {
			coPilotPod.Volumes = append(coPilotPod.Volumes, storageConfigVolume(cfg.StorageConfig))
		}

		if needsDownloader {
			inPath := cfg.DefaultInputDataPath
			if pilot.GetInputPath() != "" {
				inPath = pilot.GetInputPath()
			}

			// TODO we should calculate input volume size based on the size of the inputs which is known ahead of time. We should store that as part of the metadata
			size := CalculateStorageSize(taskExecMetadata.GetOverrides().GetResources())
			logger.Infof(ctx, "Adding Input path [%s] of Size [%v] for Task [%s]", inPath, size, taskName)
			inputsVolumeMount := v1.VolumeMount{
				Name:      cfg.InputVolumeName,
				MountPath: inPath,
			}

			format := pilot.Format
			// Lets add the InputsVolume
			coPilotPod.Volumes = append(coPilotPod.Volumes, DataVolume(cfg.InputVolumeName, size))

			// Lets add the Inputs init container
			args, err := DownloadCommandArgs(inputPaths.GetInputPath(), outputPaths.GetOutputPrefixPath(), inPath, format, pilot.GetFileInputLayout(), iFace.Inputs)
			if err != nil {
				return err
			}
			downloader, err := FlyteCoPilotContainer(flyteDownloaderContainerName, cfg, args, inputsVolumeMount, cfgMount)
			if err != nil {
				return err
			}
			coPilotPod.InitContainers = append(coPilotPod.InitContainers, downloader)
		}

		if needsUploader {
			outPath := cfg.DefaultOutputPath
			if pilot.GetOutputPath() != "" {
				outPath = pilot.GetOutputPath()
			}

			size := CalculateStorageSize(taskExecMetadata.GetOverrides().GetResources())
			logger.Infof(ctx, "Adding Output path [%s] of size [%v] for Task [%s]", outPath, size, taskName)

			outputsVolumeMount := v1.VolumeMount{
				Name:      cfg.OutputVolumeName,
				MountPath: outPath,
			}

			// Lets add the InputsVolume
			coPilotPod.Volumes = append(coPilotPod.Volumes, DataVolume(cfg.OutputVolumeName, size))

			// Lets add the Inputs init container
			args, err := SidecarCommandArgs(outPath, outputPaths.GetOutputPrefixPath(), outputPaths.GetRawOutputPrefix(), cfg.Timeout.Duration, iFace)
			if err != nil {
				return err
			}
			sidecar, err := FlyteCoPilotContainer(flyteSidecarContainerName, cfg, args, outputsVolumeMount, cfgMount)
			// Make it into sidecar container
			restartPolicy := v1.ContainerRestartPolicyAlways
			sidecar.RestartPolicy = &restartPolicy
			if err != nil {
				return err
			}
			// Let the sidecar container start before the downloader; it will ensure the signal watcher is started before the main container finishes.
			coPilotPod.InitContainers = append([]v1.Container{sidecar}, coPilotPod.InitContainers...)

			timeoutSeconds := int64(cfg.Timeout.Seconds())
			coPilotPod.TerminationGracePeriodSeconds = &timeoutSeconds
		}
	}

	return nil
}
