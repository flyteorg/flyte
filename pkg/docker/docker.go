package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate mockery -all -case=underscore

type Docker interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *specs.Platform, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error)
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)
	ImageList(ctx context.Context, listOption types.ImageListOptions) ([]types.ImageSummary, error)
	ContainerStatPath(ctx context.Context, containerID, path string) (types.ContainerPathStat, error)
	CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, types.ContainerPathStat, error)
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
