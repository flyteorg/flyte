package impl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type projectRepo struct {
	db *sqlx.DB
}

func NewProjectRepo(db *sqlx.DB) interfaces.ProjectRepo {
	return &projectRepo{
		db: db,
	}
}

func (r *projectRepo) CreateProject(ctx context.Context, project *models.Project) error {
	now := time.Now().UTC()
	project.CreatedAt = now
	project.UpdatedAt = now

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (identifier, name, description, labels, state, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (identifier) DO NOTHING`,
		project.Identifier, project.Name, project.Description, project.Labels, project.State, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		if database.IsPgErrorWithCode(err, database.PgDuplicatedKey) {
			return fmt.Errorf("%w: %s", interfaces.ErrProjectAlreadyExists, project.Identifier)
		}
		return fmt.Errorf("failed to create project %s: %w", project.Identifier, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", interfaces.ErrProjectAlreadyExists, project.Identifier)
	}
	return nil
}

func (r *projectRepo) GetProject(ctx context.Context, identifier string) (*models.Project, error) {
	var project models.Project
	err := sqlx.GetContext(ctx, r.db, &project, "SELECT * FROM projects WHERE identifier = $1", identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", interfaces.ErrProjectNotFound, identifier)
		}
		logger.Errorf(ctx, "failed to get project %s: %v", identifier, err)
		return nil, fmt.Errorf("failed to get project %s: %w", identifier, err)
	}

	return &project, nil
}

func (r *projectRepo) UpdateProject(ctx context.Context, project *models.Project) error {
	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx,
		`UPDATE projects SET name = $1, description = $2, labels = $3, state = $4, updated_at = $5
		 WHERE identifier = $6`,
		project.Name, project.Description, project.Labels, project.State, now, project.Identifier)
	if err != nil {
		logger.Errorf(ctx, "failed to update project %s: %v", project.Identifier, err)
		return fmt.Errorf("failed to update project %s: %w", project.Identifier, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", interfaces.ErrProjectNotFound, project.Identifier)
	}

	return nil
}

func (r *projectRepo) ListProjects(ctx context.Context, input interfaces.ListResourceInput) ([]*models.Project, error) {
	var queryBuilder strings.Builder
	var args []interface{}
	argIdx := 1

	queryBuilder.WriteString("SELECT * FROM projects")

	if input.Filter != nil {
		expr, err := input.Filter.QueryExpression("")
		if err != nil {
			return nil, fmt.Errorf("failed to build filter: %w", err)
		}
		rewritten, rewrittenArgs := rewritePlaceholders(expr.Query, expr.Args, argIdx)
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(rewritten)
		args = append(args, rewrittenArgs...)
		argIdx += len(rewrittenArgs)
	}

	if len(input.SortParameters) > 0 {
		queryBuilder.WriteString(" ORDER BY ")
		for i, sp := range input.SortParameters {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(sp.GetOrderExpr())
		}
	}

	if input.Limit > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d", argIdx))
		args = append(args, input.Limit)
		argIdx++
	}

	if input.Offset > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", argIdx))
		args = append(args, input.Offset)
		argIdx++
	}

	var projects []*models.Project
	if err := sqlx.SelectContext(ctx, r.db, &projects, queryBuilder.String(), args...); err != nil {
		logger.Errorf(ctx, "failed to list projects: %v", err)
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}

// rewritePlaceholders converts ? placeholders to $N PostgreSQL placeholders starting at startIdx.
func rewritePlaceholders(query string, args []interface{}, startIdx int) (string, []interface{}) {
	var result strings.Builder
	idx := startIdx
	for _, ch := range query {
		if ch == '?' {
			result.WriteString(fmt.Sprintf("$%d", idx))
			idx++
		} else {
			result.WriteRune(ch)
		}
	}
	return result.String(), args
}
