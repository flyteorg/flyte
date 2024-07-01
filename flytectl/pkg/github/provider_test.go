package github

import (
	"testing"

	"github.com/flyteorg/flyte/flytectl/pkg/github/mocks"
	go_github "github.com/google/go-github/v42/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLatestFlytectlVersion(t *testing.T) {
	t.Run("Get latest release", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		// return a list of github releases
		mockGh.OnListReleasesMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			[]*go_github.RepositoryRelease{
				{TagName: go_github.String("flytectl/1.2.4")},
				{TagName: go_github.String("flytectl/1.2.3")},
				{TagName: go_github.String("other-1.0.0")},
			},
			nil,
			nil,
		)
		mockProvider := &GHProvider{
			RepositoryURL: flytectlRepository,
			ArchiveName:   getFlytectlAssetName(),
			ghRepo:        mockGh,
		}

		latestVersion, err := mockProvider.GetLatestVersion()
		assert.Nil(t, err)
		assert.Equal(t, "flytectl/1.2.4", latestVersion)
		cleanVersion, err := mockProvider.GetCleanLatestVersion()
		assert.Nil(t, err)
		assert.Equal(t, "1.2.4", cleanVersion)
	})
}

func TestGetFlytectlReleases(t *testing.T) {
	t.Run("Get releases", func(t *testing.T) {
		mockGh := &mocks.GHRepoService{}
		allReleases := []*go_github.RepositoryRelease{
			{TagName: go_github.String("flytectl/1.2.4")},
			{TagName: go_github.String("flytectl/1.2.3")},
			{TagName: go_github.String("other-1.0.0")},
		}
		releases := []*go_github.RepositoryRelease{
			{TagName: go_github.String("flytectl/1.2.4")},
			{TagName: go_github.String("flytectl/1.2.3")},
		}
		// return a list of github releases
		mockGh.OnListReleasesMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			allReleases,
			nil,
			nil,
		)
		mockProvider := &GHProvider{
			RepositoryURL: flytectlRepository,
			ArchiveName:   getFlytectlAssetName(),
			ghRepo:        mockGh,
		}

		flytectlReleases, err := mockProvider.getReleases()
		assert.Nil(t, err)
		assert.Equal(t, releases, flytectlReleases)
	})
}
