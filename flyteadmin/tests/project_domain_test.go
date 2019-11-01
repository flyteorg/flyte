// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/gormimpl"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"

	"github.com/stretchr/testify/assert"

	databaseConfig "github.com/lyft/flyteadmin/pkg/repositories/config"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

func TestUpdateProjectDomain(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	req := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project: "admintests",
			Domain:  "development",
			Attributes: map[string]string{
				"foo": "bar",
			},
		},
	}

	_, err := client.UpdateProjectDomainAttributes(ctx, &req)
	assert.Nil(t, err)

	// If we ever expose get/list ProjectDomainAttributes APIs update the test below to call those instead.
	db := databaseConfig.OpenDbConnection(databaseConfig.NewPostgresConfigProvider(getDbConfig(), adminScope))
	defer db.Close()

	errorsTransformer := errors.NewPostgresErrorTransformer(adminScope.NewSubScope("errors"))
	projectDomainRepo := gormimpl.NewProjectDomainRepo(db, errorsTransformer, adminScope.NewSubScope("project_domain"))

	attributes, err := projectDomainRepo.Get(ctx, "admintests", "development")
	assert.Nil(t, err)

	projectDomain, err := transformers.FromProjectDomainModel(attributes)

	assert.EqualValues(t, map[string]string{
		"foo": "bar",
	}, projectDomain.Attributes)
}
