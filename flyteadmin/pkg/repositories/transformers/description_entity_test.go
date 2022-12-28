package transformers

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

const shortDescription = "hello"

func TestToDescriptionEntityExecutionModel(t *testing.T) {
	longDescription := &admin.Description{IconLink: "https://flyte"}
	sourceCode := &admin.SourceCode{Link: "https://github/flyte"}

	longDescriptionBytes, err := proto.Marshal(longDescription)
	assert.Nil(t, err)

	descriptionEntity := &admin.DescriptionEntity{
		ShortDescription: shortDescription,
		LongDescription:  longDescription,
		SourceCode:       sourceCode,
	}

	id := core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "domain",
		Version:      "xyz",
	}

	model, err := CreateDescriptionEntityModel(descriptionEntity, id)
	assert.Nil(t, err)
	assert.Equal(t, shortDescription, model.ShortDescription)
	assert.Equal(t, longDescriptionBytes, model.LongDescription)
	assert.Equal(t, sourceCode.Link, model.Link)
}

func TestFromDescriptionEntityExecutionModel(t *testing.T) {
	longDescription := &admin.Description{IconLink: "https://flyte"}
	sourceCode := &admin.SourceCode{Link: "https://github/flyte"}

	longDescriptionBytes, err := proto.Marshal(longDescription)
	assert.Nil(t, err)

	descriptionEntity, err := FromDescriptionEntityModel(models.DescriptionEntity{
		DescriptionEntityKey: models.DescriptionEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		},
		ShortDescription: shortDescription,
		LongDescription:  longDescriptionBytes,
		SourceCode:       models.SourceCode{Link: "https://github/flyte"},
	})
	assert.Nil(t, err)
	assert.Equal(t, descriptionEntity.ShortDescription, shortDescription)
	assert.Equal(t, descriptionEntity.LongDescription.IconLink, longDescription.IconLink)
	assert.Equal(t, descriptionEntity.SourceCode, sourceCode)
}

func TestFromDescriptionEntityExecutionModels(t *testing.T) {
	longDescription := &admin.Description{IconLink: "https://flyte"}
	sourceCode := &admin.SourceCode{Link: "https://github/flyte"}

	longDescriptionBytes, err := proto.Marshal(longDescription)
	assert.Nil(t, err)

	descriptionEntity, err := FromDescriptionEntityModels([]models.DescriptionEntity{
		{
			DescriptionEntityKey: models.DescriptionEntityKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
				Version: "version",
			},
			ShortDescription: shortDescription,
			LongDescription:  longDescriptionBytes,
			SourceCode:       models.SourceCode{Link: "https://github/flyte"},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, descriptionEntity[0].ShortDescription, shortDescription)
	assert.Equal(t, descriptionEntity[0].LongDescription.IconLink, longDescription.IconLink)
	assert.Equal(t, descriptionEntity[0].SourceCode, sourceCode)
}
