package validation

import (
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestValidateProjectDomainAttributesUpdateRequest(t *testing.T) {
	err := ValidateProjectDomainAttributesUpdateRequest(admin.ProjectDomainAttributesUpdateRequest{})
	assert.EqualError(t, err, "missing project_domain")

	err = ValidateProjectDomainAttributesUpdateRequest(admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{},
	})
	assert.EqualError(t, err, "missing project")

	err = ValidateProjectDomainAttributesUpdateRequest(admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project: "project",
		},
	})
	assert.EqualError(t, err, "missing domain")

	err = ValidateProjectDomainAttributesUpdateRequest(admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project: "project",
			Domain:  "domain",
		},
	})
	assert.Nil(t, err)
}
