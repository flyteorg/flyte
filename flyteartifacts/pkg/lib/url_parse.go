package lib

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

var flyteURLNameRe = regexp.MustCompile(`(?P<name>[\w/-]+)(?:(:(?P<tag>\w+))?)(?:(@(?P<version>\w+))?)`)

func ParseFlyteURL(urlStr string) (core.ArtifactID, string, error) {
	if len(urlStr) == 0 {
		return core.ArtifactID{}, "", errors.New("URL cannot be empty")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return core.ArtifactID{}, "", err
	}
	queryValues, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return core.ArtifactID{}, "", err
	}
	projectDomainName := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(projectDomainName) < 3 {
		return core.ArtifactID{}, "", errors.New("invalid URL format")
	}
	project, domain, name := projectDomainName[0], projectDomainName[1], strings.Join(projectDomainName[2:], "/")
	version := ""
	tag := ""
	queryDict := make(map[string]string)

	if match := flyteURLNameRe.FindStringSubmatch(name); match != nil {
		name = match[1]
		if match[3] != "" {
			tag = match[3]
		}
		if match[5] != "" {
			version = match[5]
		}

		if tag != "" && (version != "" || len(queryValues) > 0) {
			return core.ArtifactID{}, "", errors.New("cannot specify tag with version or querydict")
		}
	}

	for key, values := range queryValues {
		if len(values) > 0 {
			queryDict[key] = values[0]
		}
	}

	a := core.ArtifactID{
		ArtifactKey: &core.ArtifactKey{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		Version: version,
	}

	p := models.PartitionsToIdl(queryDict)
	a.Dimensions = &core.ArtifactID_Partitions{
		Partitions: p,
	}

	return a, tag, nil
}
