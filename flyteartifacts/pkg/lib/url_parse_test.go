package lib

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
)

func TestURLParseWithVersionAndPartitions(t *testing.T) {
	artifactID, err := ParseFlyteURL("flyte://av0.1/project/domain/name@version?foo=bar&ham=spam")
	expPartitions := map[string]string{"foo": "bar", "ham": "spam"}
	assert.NoError(t, err)

	assert.Equal(t, "project", artifactID.ArtifactKey.Project)
	assert.Equal(t, "domain", artifactID.ArtifactKey.Domain)
	assert.Equal(t, "name", artifactID.ArtifactKey.Name)
	assert.Equal(t, "version", artifactID.Version)
	p := artifactID.GetPartitions()
	mapP := models.PartitionsFromIdl(context.TODO(), p)
	assert.Equal(t, expPartitions, mapP)
}

func TestURLParseWithSlashVersionAndPartitions(t *testing.T) {
	artifactID, err := ParseFlyteURL("flyte://av0.1/project/domain/name/more@version/abc/0/o0?foo=bar&ham=spam")
	expPartitions := map[string]string{"foo": "bar", "ham": "spam"}
	assert.NoError(t, err)

	assert.Equal(t, "project", artifactID.ArtifactKey.Project)
	assert.Equal(t, "domain", artifactID.ArtifactKey.Domain)
	assert.Equal(t, "name/more", artifactID.ArtifactKey.Name)
	assert.Equal(t, "version/abc/0/o0", artifactID.Version)
	p := artifactID.GetPartitions()
	mapP := models.PartitionsFromIdl(context.TODO(), p)
	assert.Equal(t, expPartitions, mapP)
}

func TestURLParseNameWithSlashes(t *testing.T) {
	artifactID, err := ParseFlyteURL("flyte://av0.1/project/domain/name/with/slashes")
	assert.NoError(t, err)
	assert.Equal(t, "project", artifactID.ArtifactKey.Project)
	assert.Equal(t, "domain", artifactID.ArtifactKey.Domain)
	assert.Equal(t, "name/with/slashes", artifactID.ArtifactKey.Name)

	artifactID, err = ParseFlyteURL("flyte://av0.1/project/domain/name/with/slashes?ds=2020-01-01")
	assert.NoError(t, err)
	assert.Equal(t, "name/with/slashes", artifactID.ArtifactKey.Name)
	assert.Equal(t, "project", artifactID.ArtifactKey.Project)
	assert.Equal(t, "domain", artifactID.ArtifactKey.Domain)

	expPartitions := map[string]string{"ds": "2020-01-01"}

	assert.Equal(t, expPartitions, models.PartitionsFromIdl(context.TODO(), artifactID.GetPartitions()))
}
