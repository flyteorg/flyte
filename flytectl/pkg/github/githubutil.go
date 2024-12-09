package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/flyteorg/flyte/flytectl/pkg/platformutil"
	"github.com/flyteorg/flyte/flytectl/pkg/util"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	stdlibversion "github.com/flyteorg/flyte/flytestdlib/version"
	"github.com/google/go-github/v42/github"
	"github.com/mouuff/go-rocket-update/pkg/updater"
	"golang.org/x/oauth2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	owner                   = "flyteorg"
	flyte                   = "flyte"
	flytectl                = "flytectl"
	sandboxSupportedVersion = "v0.10.0"
	flytectlRepository      = "github.com/flyteorg/flyte"
	commonMessage           = "\n A new release of flytectl is available: %s → %s \n"
	brewMessage             = "To upgrade, run: brew update && brew upgrade flytectl \n"
	linuxMessage            = "To upgrade, run: flytectl upgrade \n"
	darwinMessage           = "To upgrade, run: flytectl upgrade \n"
	releaseURL              = "https://github.com/flyteorg/flyte/releases/tag/%s \n"
	brewInstallDirectory    = "/Cellar/flytectl"
)

var Client GHRepoService

// FlytectlReleaseConfig represent the updater config for flytectl binary
var FlytectlReleaseConfig = &updater.Updater{
	Provider: &GHProvider{
		RepositoryURL: flytectlRepository,
		ArchiveName:   getFlytectlAssetName(),
		ghRepo:        GetGHRepoService(),
	},
	ExecutableName: flytectl,
	Version:        stdlibversion.Version,
}

var (
	arch = platformutil.Arch(runtime.GOARCH)
)

//go:generate mockery -name=GHRepoService -case=underscore

type GHRepoService interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
	GetCommitSHA1(ctx context.Context, owner, repo, ref, lastSHA string) (string, *github.Response, error)
}

// GetLatestRelease returns the latest non-prerelease version of provided repoName, as
// described in https://docs.github.com/en/rest/reference/releases#get-the-latest-release
func GetLatestRelease(repoName string, g GHRepoService) (*github.RepositoryRelease, error) {
	release, _, err := g.GetLatestRelease(context.Background(), owner, repoName)
	if err != nil {
		return nil, err
	}
	return release, err
}

// ListReleases returns the list of release of provided repoName
func ListReleases(repoName string, g GHRepoService, filterPrefix string) ([]*github.RepositoryRelease, error) {
	releases, _, err := g.ListReleases(context.Background(), owner, repoName, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		return nil, err
	}
	var filteredReleases []*github.RepositoryRelease
	for _, release := range releases {
		if !strings.HasPrefix(release.GetTagName(), filterPrefix) {
			filteredReleases = append(filteredReleases, release)
		}
	}
	return filteredReleases, err
}

// GetReleaseByTag returns the provided tag release if tag exist in repository
func GetReleaseByTag(repoName, tag string, g GHRepoService) (*github.RepositoryRelease, error) {
	release, _, err := g.GetReleaseByTag(context.Background(), owner, repoName, tag)
	if err != nil {
		return nil, err
	}
	return release, err
}

// GetCommitSHA1 returns sha hash against the version
func GetCommitSHA1(repoName, version string, g GHRepoService) (string, error) {
	sha, _, err := g.GetCommitSHA1(context.Background(), owner, repoName, version, "")
	if err != nil {
		return "", err
	}
	return sha, err
}

// GetAssetFromRelease returns the asset using assetName from github release with tag
func GetAssetFromRelease(tag, assetName, repoName string, g GHRepoService) (*github.ReleaseAsset, error) {
	release, _, err := g.GetReleaseByTag(context.Background(), owner, repoName, tag)
	if err != nil {
		return nil, err
	}
	for _, v := range release.Assets {
		if v.GetName() == assetName {
			return v, nil
		}
	}
	return nil, fmt.Errorf("assest is not found in %s[%s] release", repoName, tag)
}

// GetSandboxImageSha returns the sha as per input
func GetSandboxImageSha(tag string, pre bool, g GHRepoService) (string, string, error) {
	var release *github.RepositoryRelease
	if len(tag) == 0 {
		// Only fetch Flyte releases
		releases, err := ListReleases(flyte, g, "flytectl/")
		if err != nil {
			return "", release.GetTagName(), err
		}
		for _, v := range releases {
			// When pre-releases are allowed, simply choose the latest release
			if pre {
				release = v
				break
			} else if !*v.Prerelease {
				release = v
				break
			}
		}
		logger.Infof(context.Background(), "starting with release %s", release.GetTagName())
	} else if len(tag) > 0 {
		r, err := GetReleaseByTag(flyte, tag, g)
		if err != nil {
			return "", r.GetTagName(), err
		}
		release = r
	}
	isGreater, err := util.IsVersionGreaterThan(release.GetTagName(), sandboxSupportedVersion)
	if err != nil {
		return "", release.GetTagName(), err
	}
	if !isGreater {
		return "", release.GetTagName(), fmt.Errorf("version flag only supported with flyte %s+ release", sandboxSupportedVersion)
	}
	sha, err := GetCommitSHA1(flyte, release.GetTagName(), g)
	if err != nil {
		return "", release.GetTagName(), err
	}
	return sha, release.GetTagName(), nil
}

func getFlytectlAssetName() string {
	if arch == platformutil.ArchAmd64 {
		arch = platformutil.ArchX86
	} else if arch == platformutil.ArchX86 {
		arch = platformutil.Archi386
	}
	return fmt.Sprintf("flytectl_%s_%s.tar.gz", cases.Title(language.English).String(runtime.GOOS), arch.String())
}

// GetUpgradeMessage return the upgrade message
func GetUpgradeMessage(latest string, goos platformutil.Platform) (string, error) {
	isGreater, err := util.IsVersionGreaterThan(latest, stdlibversion.Version)
	if err != nil {
		return "", err
	}

	if !isGreater {
		return "", err
	}
	message := fmt.Sprintf(commonMessage, stdlibversion.Version, latest)

	symlink, err := CheckBrewInstall(goos)
	if err != nil {
		return "", err
	}
	if len(symlink) > 0 {
		message += brewMessage
	} else if goos == platformutil.Darwin {
		message += darwinMessage
	} else if goos == platformutil.Linux {
		message += linuxMessage
	}
	message += fmt.Sprintf(releaseURL, latest)

	return message, nil
}

// CheckBrewInstall returns the path of symlink if flytectl is installed from brew
func CheckBrewInstall(goos platformutil.Platform) (string, error) {
	if goos.String() == platformutil.Darwin.String() {
		executable, err := FlytectlReleaseConfig.GetExecutable()
		if err != nil {
			return executable, err
		}
		if symlink, err := filepath.EvalSymlinks(executable); err != nil {
			return symlink, err
		} else if len(symlink) > 0 {
			if strings.Contains(symlink, brewInstallDirectory) {
				return symlink, nil
			}
		}
	}
	return "", nil
}

// GetFullyQualifiedImageName Returns the sandbox image, version and error
// if no version is specified then the Latest release of cr.flyte.org/flyteorg/flyte-sandbox:dind-{SHA} is used
// else cr.flyte.org/flyteorg/flyte-sandbox:dind-{SHA}, where sha is derived from the version.
// If pre release is true then use latest pre release of Flyte, In that case User don't need to pass version

func GetFullyQualifiedImageName(prefix, version, image string, pre bool, g GHRepoService) (string, string, error) {
	sha, version, err := GetSandboxImageSha(version, pre, g)
	if err != nil {
		return "", version, err
	}

	return fmt.Sprintf("%s:%s", image, fmt.Sprintf("%s-%s", prefix, sha)), version, nil
}

// GetGHRepoService returns the initialized github repo service client.
func GetGHRepoService() GHRepoService {
	if Client == nil {
		var gh *github.Client
		if len(os.Getenv("GITHUB_TOKEN")) > 0 {
			gh = github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
			)))
			if _, err := ListReleases(flyte, gh.Repositories, ""); err != nil {
				logger.Warnf(context.Background(), "Found GITHUB_TOKEN but failed to fetch releases. Using empty http.Client: %s.", err)
				gh = nil
			}
		}
		if gh == nil {
			gh = github.NewClient(&http.Client{})
		}
		return gh.Repositories
	}
	return Client
}
