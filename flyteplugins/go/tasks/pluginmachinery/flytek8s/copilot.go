package flytek8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	core2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const (
	flyteSidecarContainerName = "uploader"
	flyteDownloaderContainerName    = "downloader"
)

var pTraceCapability = v1.Capability("SYS_PTRACE")

func FlyteCoPilotContainer(name string, cfg config.FlyteCoPilotConfig, args []string, volumeMounts ...v1.VolumeMount) (v1.Container, error) {
	cpu, err := resource.ParseQuantity(cfg.CPU)
	if err != nil {
		return v1.Container{}, err
	}

	mem, err := resource.ParseQuantity(cfg.Memory)
	if err != nil {
		return v1.Container{}, err
	}
	return v1.Container{
		Name:       cfg.NamePrefix + name,
		Image:      cfg.Image,
		Command:    CopilotCommandArgs(storage.GetConfig()),
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

func CopilotCommandArgs(storageConfig *storage.Config) []string {
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
		for key, val := range storageConfig.Stow.Config {
			commands = append(commands, "--storage.stow.config")
			commands = append(commands, fmt.Sprintf("%s=%s", key, val))
		}
		return append(commands, fmt.Sprintf("--storage.stow.kind=%s", storageConfig.Stow.Kind))
	}
	return append(commands, []string{
		fmt.Sprintf("--storage.connection.secret-key=%s", storageConfig.Connection.SecretKey),
		fmt.Sprintf("--storage.connection.access-key=%s", storageConfig.Connection.AccessKey),
		fmt.Sprintf("--storage.connection.auth-type=%s", storageConfig.Connection.AuthType),
		fmt.Sprintf("--storage.connection.region=%s", storageConfig.Connection.Region),
		fmt.Sprintf("--storage.connection.endpoint=%s", storageConfig.Connection.Endpoint.String()),
	}...)
}

func SidecarCommandArgs(fromLocalPath string, outputPrefix, rawOutputPath storage.DataReference, startTimeout time.Duration, iface *core.TypedInterface) ([]string, error) {
	if iface == nil {
		return nil, fmt.Errorf("interface is required for CoPilot Sidecar")
	}
	b, err := proto.Marshal(iface)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal given core.TypedInterface")
	}
	return []string{
		"sidecar",
		"--start-timeout",
		startTimeout.String(),
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

func DownloadCommandArgs(fromInputsPath, outputPrefix storage.DataReference, toLocalPath string, format core.DataLoadingConfig_LiteralMapFormat, inputInterface *core.VariableMap) ([]string, error) {
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
		"--input-interface",
		base64.StdEncoding.EncodeToString(b),
	}, nil
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
	if pilot == nil || !pilot.GetEnabled() {
		return nil
	}
	logger.Infof(ctx, "Enabling CoPilot on main container [%s]", c.Name)
	if c.SecurityContext == nil {
		c.SecurityContext = &v1.SecurityContext{}
	}
	if c.SecurityContext.Capabilities == nil {
		c.SecurityContext.Capabilities = &v1.Capabilities{}
	}
	c.SecurityContext.Capabilities.Add = append(c.SecurityContext.Capabilities.Add, pTraceCapability)

	if iFace != nil {
		if iFace.GetInputs() != nil && len(iFace.GetInputs().GetVariables()) > 0 {
			inPath := cfg.DefaultInputDataPath
			if pilot.GetInputPath() != "" {
				inPath = pilot.GetInputPath()
			}

			c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
				Name:      cfg.InputVolumeName,
				MountPath: inPath,
			})
		}

		if iFace.GetOutputs() != nil && len(iFace.GetOutputs().GetVariables()) > 0 {
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

func AddCoPilotToPod(ctx context.Context, cfg config.FlyteCoPilotConfig, coPilotPod *v1.PodSpec, iFace *core.TypedInterface, taskExecMetadata core2.TaskExecutionMetadata, inputPaths io.InputFilePaths, outputPaths io.OutputFilePaths, pilot *core.DataLoadingConfig) (string, error) {
	if pilot == nil || !pilot.GetEnabled() {
		return "", nil
	}

	//nolint:protogetter
	logger.Infof(ctx, "CoPilot Enabled for task [%s]", taskExecMetadata.GetTaskExecutionID().GetID().TaskId.GetName())
	shareProcessNamespaceEnabled := true
	coPilotPod.ShareProcessNamespace = &shareProcessNamespaceEnabled
	primaryInitContainerName := ""
	if iFace != nil {
		if iFace.GetInputs() != nil && len(iFace.GetInputs().GetVariables()) > 0 {
			inPath := cfg.DefaultInputDataPath
			if pilot.GetInputPath() != "" {
				inPath = pilot.GetInputPath()
			}

			// TODO we should calculate input volume size based on the size of the inputs which is known ahead of time. We should store that as part of the metadata
			size := CalculateStorageSize(taskExecMetadata.GetOverrides().GetResources())
			//nolint:protogetter
			logger.Infof(ctx, "Adding Input path [%s] of Size [%d] for Task [%s]", inPath, size, taskExecMetadata.GetTaskExecutionID().GetID().TaskId.GetName())
			inputsVolumeMount := v1.VolumeMount{
				Name:      cfg.InputVolumeName,
				MountPath: inPath,
			}

			format := pilot.GetFormat()
			// Lets add the InputsVolume
			coPilotPod.Volumes = append(coPilotPod.Volumes, DataVolume(cfg.InputVolumeName, size))

			// Lets add the Inputs init container
			args, err := DownloadCommandArgs(inputPaths.GetInputPath(), outputPaths.GetOutputPrefixPath(), inPath, format, iFace.GetInputs())
			if err != nil {
				return primaryInitContainerName, err
			}
			downloader, err := FlyteCoPilotContainer(flyteDownloaderContainerName, cfg, args, inputsVolumeMount)
			if err != nil {
				return primaryInitContainerName, err
			}
			coPilotPod.InitContainers = append(coPilotPod.InitContainers, downloader)
			primaryInitContainerName = downloader.Name
		}

		if iFace.GetOutputs() != nil && len(iFace.GetOutputs().GetVariables()) > 0 {
			outPath := cfg.DefaultOutputPath
			if pilot.GetOutputPath() != "" {
				outPath = pilot.GetOutputPath()
			}

			size := CalculateStorageSize(taskExecMetadata.GetOverrides().GetResources())
			//nolint:protogetter
			logger.Infof(ctx, "Adding Output path [%s] of size [%d] for Task [%s]", size, outPath, taskExecMetadata.GetTaskExecutionID().GetID().TaskId.GetName())

			outputsVolumeMount := v1.VolumeMount{
				Name:      cfg.OutputVolumeName,
				MountPath: outPath,
			}

			// Lets add the InputsVolume
			coPilotPod.Volumes = append(coPilotPod.Volumes, DataVolume(cfg.OutputVolumeName, size))

			// Lets add the Inputs init container
			args, err := SidecarCommandArgs(outPath, outputPaths.GetOutputPrefixPath(), outputPaths.GetRawOutputPrefix(), cfg.StartTimeout.Duration, iFace)
			if err != nil {
				return primaryInitContainerName, err
			}
			sidecar, err := FlyteCoPilotContainer(flyteSidecarContainerName, cfg, args, outputsVolumeMount)
			// Make it into sidecar container
			sidecar.RestartPolicy = ptr.To(v1.ContainerRestartPolicyAlways)
			if err != nil {
				return primaryInitContainerName, err
			}
			coPilotPod.InitContainers = append(coPilotPod.InitContainers, sidecar)
		}
	}

	return primaryInitContainerName, nil
}
