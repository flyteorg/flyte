package impl

import (
	"context"
	"testing"

	"github.com/flyteorg/flyte/v2/flytestdlib/authorization"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationTaskQueries(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepo(db)
	EnableAuthorization(repo)
	system := authorization.WithSystem(context.Background())
	for _, key := range []models.TaskKey{
		{Domain: "prod", Project: "payments", Name: "auth-a", Version: "v1"},
		{Domain: "prod", Project: "other", Name: "auth-b", Version: "v1"},
		{Domain: "dev", Project: "payments", Name: "auth-c", Version: "v1"},
		{Domain: "dev", Project: "other", Name: "auth-d", Version: "v1"},
	} {
		require.NoError(t, repo.CreateTask(system, &models.Task{TaskKey: key, TaskSpec: []byte("{}")}, nil))
	}
	ctx := authorization.WithScopes(context.Background(), []authorization.Scope{{Domain: "prod", Project: "payments"}, {Domain: "dev", Project: "other"}})
	result, err := repo.ListTasks(ctx, interfaces.ListResourceInput{Limit: 100})
	require.NoError(t, err)
	require.Len(t, result.Tasks, 2)
	require.EqualValues(t, 2, result.Total)
	require.EqualValues(t, 2, result.FilteredTotal)
	// Client OR filters must not escape the independently injected isolation predicate.
	input := interfaces.ListResourceInput{Limit: 100, Filter: NewEqualFilter("project", "payments").Or(NewEqualFilter("domain", "dev"))}
	result, err = repo.ListTasks(ctx, input)
	require.NoError(t, err)
	require.Len(t, result.Tasks, 2)
	input.Filter = NewEqualFilter("name", "auth-a")
	result, err = repo.ListTasks(ctx, input)
	require.NoError(t, err)
	require.Len(t, result.Tasks, 1)
	require.EqualValues(t, 2, result.Total)
	versions, err := repo.ListVersions(ctx, interfaces.ListResourceInput{Limit: 100})
	require.NoError(t, err)
	require.Len(t, versions, 2)
	_, err = repo.GetTask(ctx, models.TaskKey{Domain: "prod", Project: "other", Name: "auth-b", Version: "v1"})
	require.ErrorIs(t, err, authorization.ErrDenied)
	err = repo.CreateTask(ctx, &models.Task{TaskKey: models.TaskKey{Domain: "prod", Project: "other", Name: "auth-b", Version: "v1"}}, nil)
	require.ErrorIs(t, err, authorization.ErrDenied)
	_, err = repo.ListTasks(context.Background(), interfaces.ListResourceInput{Limit: 100})
	require.ErrorIs(t, err, authorization.ErrDenied)
	result, err = repo.ListTasks(authorization.WithScopes(context.Background(), nil), interfaces.ListResourceInput{Limit: 100})
	require.NoError(t, err)
	require.Empty(t, result.Tasks)
	// Authorization applies to every trigger before beginning the task transaction.
	err = repo.CreateTask(ctx, &models.Task{TaskKey: models.TaskKey{Domain: "prod", Project: "payments", Name: "auth-new", Version: "v1"}}, []*models.Trigger{{Project: "other", Domain: "prod"}})
	require.ErrorIs(t, err, authorization.ErrDenied)
}

func TestAuthorizationDeniesBeforeDatabaseAccess(t *testing.T) {
	ctx := authorization.WithScopes(context.Background(), []authorization.Scope{{Domain: "prod", Project: "payments"}})
	// A nil DB makes an accidental query observable as a panic.
	tasks := NewTaskRepo(nil)
	EnableAuthorization(tasks)
	_, err := tasks.GetTaskSpec(ctx, "guessed-digest")
	require.ErrorIs(t, err, authorization.ErrDenied)
	projects := NewProjectRepo(nil)
	EnableAuthorization(projects)
	err = projects.UpdateProject(ctx, &models.Project{Identifier: "payments"})
	require.ErrorIs(t, err, authorization.ErrDenied)
	settings := NewSettingsRepo(nil)
	EnableAuthorization(settings)
	err = settings.CreateSettings(ctx, &models.Settings{Key: "v1::prod:", Data: []byte("{}")})
	require.ErrorIs(t, err, authorization.ErrDenied)
	triggers := NewTriggerRepo(nil)
	EnableAuthorization(triggers)
	err = triggers.DeleteTriggers(ctx, []interfaces.TriggerNameKey{{Domain: "prod", Project: "payments"}, {Domain: "dev", Project: "payments"}})
	require.ErrorIs(t, err, authorization.ErrDenied)
	actions := &actionRepo{}
	actions.EnableAuthorization()
	err = actions.InsertEvents(ctx, []*models.ActionEvent{{Domain: "prod", Project: "payments"}, {Domain: "prod", Project: "other"}})
	require.ErrorIs(t, err, authorization.ErrDenied)
	_, err = actions.ListActions(context.Background(), interfaces.ListResourceInput{})
	require.ErrorIs(t, err, authorization.ErrDenied)
}

func TestAuthorizationFilterRequiresContext(t *testing.T) {
	var guard authorizationGuard
	filter, err := guard.authorizationFilter(context.Background(), false)
	require.NoError(t, err)
	require.Nil(t, filter)

	guard.EnableAuthorization()
	for _, projects := range []bool{false, true} {
		filter, err = guard.authorizationFilter(context.Background(), projects)
		require.ErrorIs(t, err, authorization.ErrDenied)
		require.Nil(t, filter)
		// A supplied client filter must not bypass context validation.
		filter, err = guard.andAuthorization(NewRawFilter("TRUE"), context.Background(), projects)
		require.ErrorIs(t, err, authorization.ErrDenied)
		require.Nil(t, filter)

		filter, err = guard.authorizationFilter(authorization.WithSystem(context.Background()), projects)
		require.NoError(t, err)
		require.Nil(t, filter)

		filter, err = guard.authorizationFilter(authorization.WithScopes(context.Background(), nil), projects)
		require.NoError(t, err)
		require.NotNil(t, filter)
		expr, err := filter.QueryExpression("")
		require.NoError(t, err)
		require.Equal(t, "FALSE", expr.Query)
	}
}
