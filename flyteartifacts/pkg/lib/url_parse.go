package lib

import (
	"errors"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"net/url"
	"regexp"
)

var flyteURLNameRe = regexp.MustCompile(`(?P<project>[\w-]+)/(?P<domain>[\w-]+)/(?P<name>[\w/-]+)(@(?P<version>[\w/-]+))?`)

func ParseFlyteURL(urlStr string) (core.ArtifactID, error) {
	if len(urlStr) == 0 {
		return core.ArtifactID{}, errors.New("URL cannot be empty")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return core.ArtifactID{}, err
	}
	queryValues, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return core.ArtifactID{}, err
	}
	//projectDomainName := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	//if len(projectDomainName) < 3 {
	//	return core.ArtifactID{}, errors.New("invalid URL format")
	//}
	//project, domain, name := projectDomainName[0], projectDomainName[1], strings.Join(projectDomainName[2:], "/")
	var project, domain, name, version string
	queryDict := make(map[string]string)

	if match := flyteURLNameRe.FindStringSubmatch(parsed.Path); match != nil {
		if len(match) < 4 {
			return core.ArtifactID{}, fmt.Errorf("insufficient components specified %s", parsed.Path)
		}
		project = match[1]
		domain = match[2]
		name = match[3]
		if len(match) > 5 {
			version = match[5]
		}
	} else {
		return core.ArtifactID{}, fmt.Errorf("unable to parse %s", parsed.Path)
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

	return a, nil
}
