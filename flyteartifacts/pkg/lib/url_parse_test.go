package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestURLParseWithTag(t *testing.T) {
	project, domain, name, version, tag, queryDict, err := ParseFlyteURL("flyte://av0.1/project/domain/name:tag")
	assert.NoError(t, err)

	assert.Equal(t, "project", project)
	assert.Equal(t, "domain", domain)
	assert.Equal(t, "name", name)
	assert.Equal(t, "", version)
	assert.Equal(t, "tag", tag)
	assert.Equal(t, map[string]string{}, queryDict)
}

func TestURLParseWithVersionAndPartitions(t *testing.T) {
	project, domain, name, version, tag, queryDict, err := ParseFlyteURL("flyte://av0.1/project/domain/name@version?foo=bar&ham=spam")
	expPartitions := map[string]string{"foo": "bar", "ham": "spam"}
	assert.NoError(t, err)

	assert.Equal(t, "project", project)
	assert.Equal(t, "domain", domain)
	assert.Equal(t, "name", name)
	assert.Equal(t, "version", version)
	assert.Equal(t, "", tag)
	assert.Equal(t, expPartitions, queryDict)
}

func TestURLParseFailsWithBothTagAndPartitions(t *testing.T) {
	_, _, _, _, _, _, err := ParseFlyteURL("flyte://av0.1/project/domain/name:tag?foo=bar&ham=spam")
	assert.Error(t, err)
}

func TestURLParseWithBothTagAndVersion(t *testing.T) {
	_, _, _, _, _, _, err := ParseFlyteURL("flyte://av0.1/project/domain/name:tag@version")
	assert.Error(t, err)
}

func TestURLParseNameWithSlashes(t *testing.T) {
	res := ParseFlyteURL("flyte://av0.1/project/domain/name/with/slashes")
	expected := ("project", "domain", "name/with/slashes", "", "", map[string]string{})
	if res != expected {
		t.Errorf("Expected %v, but got %v", expected, res)
	}

	res = ParseFlyteURL("flyte://av0.1/project/domain/name/with/slashes?ds=2020-01-01")
	expected = ("project", "domain", "name/with/slashes", "", "", map[string]string{"ds": "2020-01-01"})
	if res != expected {
		t.Errorf("Expected %v, but got %v", expected, res)
	}
}
