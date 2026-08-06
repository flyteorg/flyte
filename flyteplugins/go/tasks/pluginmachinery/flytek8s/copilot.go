package flytek8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/flyteorg/stow/s3"
	"github.com/golang/protobuf/proto" //nolint: staticcheck
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"

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
	// s3AuthTypeIAM is stow's "resolve credentials ambiently" mode — it covers the
	// environment, the shared credentials file, and instance/pod roles, despite the name.
	// stow keeps its own constant unexported, so it is repeated here.
	s3AuthTypeIAM = "iam"
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

	var storageCfg *storage.Config
	if cfg.StorageConfigOverride != nil {
		storageCfg = cfg.StorageConfigOverride
	} else {
		storageCfg = storage.GetConfig()
	}

	command, err := CopilotCommandArgs(ctx, storageCfg, cfg.StorageCredentials)
	if err != nil {
		return v1.Container{}, err
	}

	// Supplies the credentials withheld from the command above. envFrom puts only the
	// Secret's name in the pod spec, so reading the values requires access to the Secret.
	// Attached to the co-pilot containers alone — never to the primary container, which
	// runs user code.
	var envFrom []v1.EnvFromSource
	if cfg.StorageCredentials.SecretName != "" {
		envFrom = append(envFrom, v1.EnvFromSource{
			SecretRef: &v1.SecretEnvSource{
				LocalObjectReference: v1.LocalObjectReference{Name: cfg.StorageCredentials.SecretName},
			},
		})
	}

	return v1.Container{
		Name:       cfg.NamePrefix + name,
		Image:      cfg.Image,
		Command:    command,
		Args:       args,
		EnvFrom:    envFrom,
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

// s3CredentialStowKeys are the stow config keys that carry S3 credentials. They are
// withheld from the co-pilot command line when a credentials Secret is configured — see
// CopilotCommandArgs.
var s3CredentialStowKeys = sets.NewString(
	s3.ConfigAccessKeyID,
	s3.ConfigSecretKey,
)

// CopilotCommandArgs builds the co-pilot entrypoint.
//
// When a credentials Secret is configured, the S3 credentials are withheld from the
// rendered stow config: they would otherwise appear in every task's pod spec, where any
// principal that can read pods could recover them without access to the Secret. They reach
// co-pilot as environment variables instead, injected by FlyteCoPilotContainer.
//
// auth_type is then forced to iam, because stow only consults ambient credentials
// (environment, shared credentials file, instance role) in that mode; under accesskey it
// builds a static credentials provider from the config map and never looks at the
// environment — and its makefn rejects the config outright once the keys are absent.
// Overriding it here rather than in the deployment's own storage config keeps the change
// scoped to co-pilot: the main binary keeps reading its credentials from file.
//
// With no Secret configured the credentials are rendered as before, so deployments that
// have not adopted the new setting keep working. This leaves them exposed, hence the
// warning; the exposure closes when they set storage-credentials.secret-name.
//
// Non-S3 backends are always rendered unchanged. Their credentials have no
// ambient-resolution path in stow (Azure and Swift read the key straight from the config
// map), so withholding them would break those deployments rather than secure them.
func CopilotCommandArgs(ctx context.Context, storageConfig *storage.Config, credentials config.StorageCredentialsConfig) ([]string, error) {
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

	if len(storageConfig.Stow.Config) > 0 && len(storageConfig.Stow.Kind) > 0 {
		isS3 := storageConfig.Stow.Kind == s3.Kind
		withholdCredentials := isS3 && credentials.SecretName != ""

		for key, val := range storageConfig.Stow.Config {
			// secret_key_path is never forwarded, in either mode: it names a file in the
			// deployment's own container, which does not exist in a task pod, and
			// newStowRawStore stats it regardless of auth_type — so passing it stops
			// co-pilot from starting at all.
			if isS3 && key == storage.ConfigSecretKeyPath {
				continue
			}
			// TODO(alex): we still write s3 auth info into command when !withholdCredentials due to backward compatibility
			// We should remove it in the future and all auth info should be passed into copilot with secret volume
			if withholdCredentials && (s3CredentialStowKeys.Has(key) || key == s3.ConfigAuthType) {
				continue
			}
			commands = append(commands, "--storage.stow.config")
			commands = append(commands, fmt.Sprintf("%s=%s", key, val))
		}
		if withholdCredentials {
			commands = append(commands, "--storage.stow.config", fmt.Sprintf("%s=%s", s3.ConfigAuthType, s3AuthTypeIAM))
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
			downloader, err := FlyteCoPilotContainer(ctx, flyteDownloaderContainerName, cfg, args, inputsVolumeMount)
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
			sidecar, err := FlyteCoPilotContainer(ctx, flyteSidecarContainerName, cfg, args, outputsVolumeMount)
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
