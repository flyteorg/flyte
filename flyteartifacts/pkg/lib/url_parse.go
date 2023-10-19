package lib

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var flyteURLNameRe = regexp.MustCompile(`(?P<name>[\w/-]+)(?:(:(?P<tag>\w+))?)(?:(@(?P<version>\w+))?)`)

func ParseFlyteURL(urlStr string) (string, string, string, string, string, map[string]string, error) {

	if len(urlStr) == 0 {
		return "", "", "", "", "", nil, errors.New("URL cannot be empty")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return "", "", "", "", "", nil, err
	}
	queryValues, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", "", "", "", "", nil, err
	}
	projectDomainName := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(projectDomainName) < 3 {
		return "", "", "", "", "", nil, errors.New("Invalid URL format")
	}
	project, domain, name := projectDomainName[0], projectDomainName[1], projectDomainName[2]
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
			return "", "", "", "", "", nil, errors.New("Cannot specify tag with version or querydict")
		}
	}

	for key, values := range queryValues {
		if len(values) > 0 {
			queryDict[key] = values[0]
		}
	}

	return project, domain, name, version, tag, queryDict, nil
}
