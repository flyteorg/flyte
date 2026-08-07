package flytek8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/flyteorg/stow/s3"
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
	// copilotConfigVolumeName holds the projected storage configuration. Mounted on the
	// co-pilot containers only.
	copilotConfigVolumeName = "flyte-copilot-storage-config"
)

func FlyteCoPilotContainer(ctx context.Context, name string, cfg config.FlyteCoPilotConfig, args []string, volumeMounts ...v1.VolumeMount) (v1.Container, error) {
	cpu, err := resource.ParseQuantity(cfg.CPU)
	if err != nil {
		return v1.Container{}, err
	}

	mem, err := resource.ParseQuantity(cfg.Memory)
	if err != nil {
		return v1.Container{}, err
	}

	storageCfg := storage.GetConfig()
	// Use override value if provideds
	overridden := cfg.StorageConfigOverride != nil
	if overridden {
		storageCfg = cfg.StorageConfigOverride
	}

	command, err := CopilotCommandArgs(ctx, storageCfg, cfg.StorageConfig, overridden)
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

// Both exposures below are a property of the deployment's configuration, not of any one
// task, so they are reported once per process rather than on every pod build.
var (
	warnOverrideExposure     sync.Once
	warnUnconfiguredExposure sync.Once
)

// CopilotCommandArgs builds the co-pilot entrypoint.
//
// The stow config — endpoint and credentials alike — normally does NOT appear here.
// Rendering it would publish the storage credentials in every task's pod spec, where any
// principal that can read pods could recover them without access to the Secret holding
// them. Instead AddCoPilotToPod projects the deployment's own storage ConfigMap and
// Secret into the co-pilot containers and this function points co-pilot at them with
// --config, which flytestdlib globs and merges (one viper per file).
//
// The remaining flags are all scalars, so they still override the mounted files key by
// key — viper resolves each key from the highest-precedence source that has it. That is
// what keeps --storage.container and friends working as operator-facing overrides.
// storage.stow.config is the exception: it binds to a StringToString pflag, and viper
// returns a set map flag whole rather than merging it into the file's map, so passing it
// would silently drop every key it does not itself carry — including the credentials.
// Hence it is all-or-nothing, and the two cases where it is rendered take all of it:
//
//   - No sources configured. Deployments that predate this setting keep working exactly
//     as before, at the cost of the exposure above — hence the warning. It closes when
//     they set co-pilot.storage-config.
//   - StorageConfigOverride set. The operator has deliberately given co-pilot a different
//     storage config, and the mounted files cannot express it. Honouring the override
//     means putting all of it on the command line, warning included.
func CopilotCommandArgs(ctx context.Context, storageConfig *storage.Config, sources config.StorageConfigSources, overridden bool) ([]string, error) {
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

	if sources.Configured() {
		commands = append(commands, "--config", filepath.Join(sources.MountPath, "*.yaml"))
		if !overridden {
			return commands, nil
		}
		warnOverrideExposure.Do(func() {
			logger.Warnf(ctx, "co-pilot storage-config-override is set, so the stow config is rendered "+
				"into the task pod spec and any credentials it carries are readable by anyone who can read "+
				"pods. Remove the override to have co-pilot read the mounted configuration instead.")
		})
	} else {
		warnUnconfiguredExposure.Do(func() {
			logger.Warnf(ctx, "co-pilot storage-config is not configured, so the stow config is rendered "+
				"into the task pod spec and any credentials it carries are readable by anyone who can read "+
				"pods. Set co-pilot.storage-config.config-map-name and secret-name to close this.")
		})
	}

	if len(storageConfig.Stow.Config) > 0 && len(storageConfig.Stow.Kind) > 0 {
		isS3 := storageConfig.Stow.Kind == s3.Kind
		for key, val := range storageConfig.Stow.Config {
			// secret_key_path is never forwarded: it names a file in the deployment's own
			// container, which does not exist in a task pod, and newStowRawStore stats it
			// regardless of auth_type — so passing it stops co-pilot from starting at all.
			if isS3 && key == storage.ConfigSecretKeyPath {
				continue
			}
			commands = append(commands, "--storage.stow.config")
			commands = append(commands, fmt.Sprintf("%s=%s", key, val))
		}
		return append(commands, fmt.Sprintf("--storage.stow.kind=%s", storageConfig.Stow.Kind)), nil
	}

	return commands, fmt.Errorf("no stow.config or stow.kind specified")
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

// storageConfigVolume projects co-pilot's storage configuration into a read-only volume.
// Both sources are combined here rather than merged into one object, preserving the
// deployment's split of non-sensitive settings (ConfigMap) from credentials (Secret).
//
// Only the configured keys are projected. Both objects carry unrelated entries — plugin
// config, database credentials — which must neither reach the co-pilot containers nor be
// parsed by co-pilot's strict-mode config loader, which rejects sections it does not
// recognise. A source with no name is omitted entirely: a deployment using instance
// credentials (S3 authType=iam) renders no credential file, and naming a key that is
// never written would leave the pod stuck in ContainerCreating.
func storageConfigVolume(sources config.StorageConfigSources) v1.Volume {
	// Read-only for the owner: co-pilot only reads this, and the containers run as
	// whatever user the image declares.
	mode := int32(0400)
	projections := make([]v1.VolumeProjection, 0, 2)
	if sources.ConfigMapName != "" {
		projections = append(projections, v1.VolumeProjection{
			ConfigMap: &v1.ConfigMapProjection{
				LocalObjectReference: v1.LocalObjectReference{Name: sources.ConfigMapName},
				Items:                keyToPaths(sources.ConfigMapKeys),
			},
		})
	}
	if sources.SecretName != "" {
		projections = append(projections, v1.VolumeProjection{
			Secret: &v1.SecretProjection{
				LocalObjectReference: v1.LocalObjectReference{Name: sources.SecretName},
				Items:                keyToPaths(sources.SecretKeys),
			},
		})
	}
	return v1.Volume{
		Name: copilotConfigVolumeName,
		VolumeSource: v1.VolumeSource{
			Projected: &v1.ProjectedVolumeSource{
				Sources:     projections,
				DefaultMode: &mode,
			},
		},
	}
}

// storageConfigMount is the mount matching storageConfigVolume. It is attached to the
// co-pilot containers only — never to the primary container, which runs user code and
// must not be able to read the storage credentials.
func storageConfigMount(sources config.StorageConfigSources) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      copilotConfigVolumeName,
		MountPath: sources.MountPath,
		ReadOnly:  true,
	}
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

	if iFace != nil {
		needsDownloader := iFace.Inputs != nil && len(iFace.Inputs.Variables) > 0
		needsUploader := iFace.Outputs != nil && len(iFace.Outputs.Variables) > 0

		// The storage config is shared by both co-pilot containers, so the volume is added
		// once here and mounted from each. It is deliberately absent from the primary
		// container's mounts: that one runs user code.
		var copilotMounts []v1.VolumeMount
		if (needsDownloader || needsUploader) && cfg.StorageConfig.Configured() {
			coPilotPod.Volumes = append(coPilotPod.Volumes, storageConfigVolume(cfg.StorageConfig))
			copilotMounts = append(copilotMounts, storageConfigMount(cfg.StorageConfig))
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
			downloader, err := FlyteCoPilotContainer(ctx, flyteDownloaderContainerName, cfg, args,
				append([]v1.VolumeMount{inputsVolumeMount}, copilotMounts...)...)
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
			sidecar, err := FlyteCoPilotContainer(ctx, flyteSidecarContainerName, cfg, args,
				append([]v1.VolumeMount{outputsVolumeMount}, copilotMounts...)...)
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
