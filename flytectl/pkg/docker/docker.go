package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate mockery -all -case=underscore

type Docker interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
	ContainerWait(ctx context.Context, container string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)
	ImageList(ctx context.Context, listOption types.ImageListOptions) ([]image.Summary, error)
	ContainerStatPath(ctx context.Context, containerID, path string) (types.ContainerPathStat, error)
	CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, types.ContainerPathStat, error)
	VolumeCreate(ctx context.Context, options volume.CreateOptions) (volume.Volume, error)
	VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
}

type FlyteDocker struct {
	*client.Client
}

//go:generate enumer -type=ImagePullPolicy -trimprefix=ImagePullPolicy --json
type ImagePullPolicy int

const (
	ImagePullPolicyAlways ImagePullPolicy = iota
	ImagePullPolicyIfNotPresent
	ImagePullPolicyNever
)

// Set implements PFlag's Value interface to attempt to set the value of the flag from string.
func (i *ImagePullPolicy) Set(val string) error {
	policy, err := ImagePullPolicyString(val)
	if err != nil {
		return err
	}

	*i = policy
	return nil
}

// Type implements PFlag's Value interface to return type name.
func (i ImagePullPolicy) Type() string {
	return "ImagePullPolicy"
}

type ImagePullOptions struct {
	RegistryAuth string `json:"registryAuth" pflag:",The base64 encoded credentials for the registry."`
	Platform     string `json:"platform" pflag:",Forces a specific platform's image to be pulled.'"`
}
