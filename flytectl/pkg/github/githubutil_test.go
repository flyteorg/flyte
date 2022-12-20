package github

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/flyteorg/flytectl/pkg/github/mocks"
	"github.com/flyteorg/flytectl/pkg/platformutil"
	stdlibversion "github.com/flyteorg/flytestdlib/version"
	"github.com/google/go-github/v42/github"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var sandboxImageName = "cr.flyte.org/flyteorg/flyte-sandbox"

func TestGetLatestVersion(t *testing.T) {
	t.Run("Get latest release with wrong url", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		mockGh.OnGetLatestReleaseMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("failed"))
		_, err := GetLatestRelease("fl", mockGh)
		assert.NotNil(t, err)
	})
	t.Run("Get latest release", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		mockGh.OnGetLatestReleaseMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, nil)
		_, err := GetLatestRelease("flytectl", mockGh)
		assert.Nil(t, err)
	})
}

func TestGetLatestRelease(t *testing.T) {
	mockGh := &mocks.GHRepoService{}
	tag := "v1.0.0"
	mockGh.OnGetLatestReleaseMatch(mock.Anything, mock.Anything, mock.Anything).Return(&github.RepositoryRelease{
		TagName: &tag,
	}, nil, nil)
	release, err := GetLatestRelease("flyte", mockGh)
	assert.Nil(t, err)
	assert.Equal(t, true, strings.HasPrefix(release.GetTagName(), "v"))
}

func TestCheckVersionExist(t *testing.T) {
	t.Run("Invalid Tag", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		mockGh.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("failed"))
		_, err := GetReleaseByTag("v100.0.0", "flyte", mockGh)
		assert.NotNil(t, err)
	})
	t.Run("Valid Tag", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		tag := "v1.0.0"
		mockGh.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&github.RepositoryRelease{
			TagName: &tag,
		}, nil, nil)
		release, err := GetReleaseByTag(tag, "flyte", mockGh)
		assert.Nil(t, err)
		assert.Equal(t, true, strings.HasPrefix(release.GetTagName(), "v"))
	})
}

func TestGetFullyQualifiedImageName(t *testing.T) {
	t.Run("Get tFully Qualified Image Name ", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		tag := "v0.15.0"
		isPreRelease := false
		releases := []*github.RepositoryRelease{{
			TagName:    &tag,
			Prerelease: &isPreRelease,
		}}
		mockGh.OnListReleasesMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(releases, nil, nil)
		mockGh.OnGetCommitSHA1Match(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sandboxImageName, nil, nil)
		image, tag, err := GetFullyQualifiedImageName("dind", "", sandboxImageName, false, mockGh)
		assert.Nil(t, err)
		assert.Equal(t, true, strings.HasPrefix(tag, "v"))
		assert.Equal(t, true, strings.HasPrefix(image, sandboxImageName))
	})
	t.Run("Get Fully Qualified Image Name with pre release", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		tag := "v0.15.0-pre"
		isPreRelease := true
		releases := []*github.RepositoryRelease{{
			TagName:    &tag,
			Prerelease: &isPreRelease,
		}}
		mockGh.OnListReleasesMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(releases, nil, nil)
		mockGh.OnGetCommitSHA1Match(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sandboxImageName, nil, nil)
		image, tag, err := GetFullyQualifiedImageName("dind", "", sandboxImageName, isPreRelease, mockGh)
		assert.Nil(t, err)
		assert.Equal(t, true, strings.HasPrefix(tag, "v"))
		assert.Equal(t, true, strings.HasPrefix(image, sandboxImageName))
	})
	t.Run("Get Fully Qualified Image Name with specific version", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		tag := "v0.19.0"
		isPreRelease := true
		release := &github.RepositoryRelease{
			TagName:    &tag,
			Prerelease: &isPreRelease,
		}
		mockGh.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(release, nil, nil)
		mockGh.OnGetCommitSHA1Match(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sandboxImageName, nil, nil)
		image, tag, err := GetFullyQualifiedImageName("dind", "v0.19.0", sandboxImageName, isPreRelease, mockGh)
		assert.Nil(t, err)
		assert.Equal(t, "v0.19.0", tag)
		assert.Equal(t, true, strings.HasPrefix(image, sandboxImageName))
	})
}

func TestGetSHAFromVersion(t *testing.T) {
	t.Run("Invalid Tag", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		mockGh.OnGetCommitSHA1Match(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil, fmt.Errorf("failed"))
		_, err := GetCommitSHA1("v100.0.0", "flyte", mockGh)
		assert.NotNil(t, err)
	})
	t.Run("Valid Tag", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		mockGh.OnGetCommitSHA1Match(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("v1.15.0", nil, nil)
		release, err := GetCommitSHA1("v0.15.0", "flyte", mockGh)
		assert.Nil(t, err)
		assert.Greater(t, len(release), 0)
	})
}

func TestGetAssetsFromRelease(t *testing.T) {
	t.Run("Successful get assets", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		tag := "v0.15.0"
		sandboxManifest := "flyte_sandbox_manifest.yaml"
		mockGh.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&github.RepositoryRelease{
			TagName: &tag,
			Assets: []*github.ReleaseAsset{{
				Name: &sandboxManifest,
			},
			},
		}, nil, nil)
		assets, err := GetAssetFromRelease(tag, sandboxManifest, flyte, mockGh)
		assert.Nil(t, err)
		assert.NotNil(t, assets)
		assert.Equal(t, sandboxManifest, *assets.Name)
	})

	t.Run("Failed get assets with wrong name", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		mockGh.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("failed"))
		assets, err := GetAssetFromRelease("v0.15.0", "test", flyte, mockGh)
		assert.NotNil(t, err)
		assert.Nil(t, assets)
	})
}

func TestGetAssetsName(t *testing.T) {
	t.Run("Get Assets name", func(t *testing.T) {
		expected := fmt.Sprintf("flytectl_%s_386.tar.gz", cases.Title(language.English).String(runtime.GOOS))
		arch = platformutil.Arch386
		assert.Equal(t, expected, getFlytectlAssetName())
	})
}

func TestCheckBrewInstall(t *testing.T) {
	symlink, err := CheckBrewInstall(platformutil.Darwin)
	assert.Nil(t, err)
	assert.Equal(t, len(symlink), 0)
	symlink, err = CheckBrewInstall(platformutil.Linux)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(symlink))
}

func TestGetUpgradeMessage(t *testing.T) {
	var darwin = platformutil.Darwin
	var linux = platformutil.Linux
	var windows = platformutil.Linux

	var version = "v0.2.20"
	stdlibversion.Version = "v0.2.10"
	message, err := GetUpgradeMessage(version, darwin)
	assert.Nil(t, err)
	assert.Equal(t, 157, len(message))

	version = "v0.2.09"
	message, err = GetUpgradeMessage(version, darwin)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(message))

	version = "v"
	message, err = GetUpgradeMessage(version, darwin)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(message))

	version = "v0.2.20"
	message, err = GetUpgradeMessage(version, windows)
	assert.Nil(t, err)
	assert.Equal(t, 157, len(message))

	version = "v0.2.20"
	message, err = GetUpgradeMessage(version, linux)
	assert.Nil(t, err)
	assert.Equal(t, 157, len(message))
}
