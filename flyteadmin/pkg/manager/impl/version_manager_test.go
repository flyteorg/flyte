package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	adminversion "github.com/flyteorg/flyte/flytestdlib/version"
)

var (
	build      = "fa70a14"
	buildTime  = "2021-03-25"
	appversion = "v0.3.7-47-gfa70a14"
)

func TestVersionManager_GetVersion(t *testing.T) {
	adminversion.Build = build
	adminversion.BuildTime = buildTime
	adminversion.Version = appversion
	vmanager := NewVersionManager()

	v, err := vmanager.GetVersion(context.Background(), &admin.GetVersionRequest{})
	assert.Nil(t, err)
	assert.Equal(t, v.ControlPlaneVersion.BuildTime, buildTime)
	assert.Equal(t, v.ControlPlaneVersion.Build, build)
	assert.Equal(t, v.ControlPlaneVersion.Version, appversion)
}
