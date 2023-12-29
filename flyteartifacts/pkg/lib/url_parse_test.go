package lib

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
)

func TestURLParseWithTag(t *testing.T) {
	artifactID, tag, err := ParseFlyteURL("flyte://av0.1/project/domain/name:tag")
	assert.NoError(t, err)

	assert.Equal(t, "project", artifactID.ArtifactKey.Project)
	assert.Equal(t, "domain", artifactID.ArtifactKey.Domain)
	assert.Equal(t, "name", artifactID.ArtifactKey.Name)
	assert.Equal(t, "", artifactID.Version)
	assert.Equal(t, "tag", tag)
	assert.Nil(t, artifactID.GetPartitions())
}

func TestURLParseWithVersionAndPartitions(t *testing.T) {
	artifactID, tag, err := ParseFlyteURL("flyte://av0.1/project/domain/name@version?foo=bar&ham=spam")
	expPartitions := map[string]string{"foo": "bar", "ham": "spam"}
	assert.NoError(t, err)

	assert.Equal(t, "project", artifactID.ArtifactKey.Project)
	assert.Equal(t, "domain", artifactID.ArtifactKey.Domain)
	assert.Equal(t, "name", artifactID.ArtifactKey.Name)
	assert.Equal(t, "version", artifactID.Version)
	assert.Equal(t, "", tag)
	p := artifactID.GetPartitions()
	mapP := models.PartitionsFromIdl(context.TODO(), p)
	assert.Equal(t, expPartitions, mapP)
}

func TestURLParseFailsWithBothTagAndPartitions(t *testing.T) {
	_, _, err := ParseFlyteURL("flyte://av0.1/project/domain/name:tag?foo=bar&ham=spam")
	assert.Error(t, err)
}

func TestURLParseWithBothTagAndVersion(t *testing.T) {
	_, _, err := ParseFlyteURL("flyte://av0.1/project/domain/name:tag@version")
	assert.Error(t, err)
}

func TestURLParseNameWithSlashes(t *testing.T) {
	artifactID, tag, err := ParseFlyteURL("flyte://av0.1/project/domain/name/with/slashes")
	assert.NoError(t, err)
	assert.Equal(t, "project", artifactID.ArtifactKey.Project)
	assert.Equal(t, "domain", artifactID.ArtifactKey.Domain)
	assert.Equal(t, "name/with/slashes", artifactID.ArtifactKey.Name)
	assert.Equal(t, "", tag)

	artifactID, _, err = ParseFlyteURL("flyte://av0.1/project/domain/name/with/slashes?ds=2020-01-01")
	assert.NoError(t, err)
	assert.Equal(t, "name/with/slashes", artifactID.ArtifactKey.Name)
	assert.Equal(t, "project", artifactID.ArtifactKey.Project)
	assert.Equal(t, "domain", artifactID.ArtifactKey.Domain)

	expPartitions := map[string]string{"ds": "2020-01-01"}

	assert.Equal(t, expPartitions, models.PartitionsFromIdl(context.TODO(), artifactID.GetPartitions()))
}
