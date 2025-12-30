package transformers

import (
	"context"
	"database/sql"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// NewTaskKey creates a models.TaskKey from a task.TaskIdentifier
func ToTaskKey(taskId *task.TaskIdentifier) models.TaskKey {
	return models.TaskKey{
		Org:     taskId.GetOrg(),
		Project: taskId.GetProject(),
		Domain:  taskId.GetDomain(),
		Name:    taskId.GetName(),
		Version: taskId.GetVersion(),
	}
}

// ToTaskName creates a models.TaskName from a task.TaskName
func ToTaskName(taskName *task.TaskName) models.TaskName {
	return models.TaskName{
		Org:     taskName.GetOrg(),
		Project: taskName.GetProject(),
		Domain:  taskName.GetDomain(),
		Name:    taskName.GetName(),
	}
}

func NewTaskModel(ctx context.Context, taskId *task.TaskIdentifier, spec *task.TaskSpec) (*models.Task, error) {
	var deployedBy string
	// TODO(nary): Get real identity subject after adding auth
	deployedBy = "mock-subject"
	// if subject, err := authorization.GetCallerIdentitySubject(ctx); err != nil {
	// 	logger.Warnf(ctx, "Failed to get caller identity subject. Error: %v", err)
	// } else {
	// 	deployedBy = subject
	// }

	specBytes, err := proto.Marshal(spec)
	if err != nil {
		return nil, err
	}

	var environment string
	var envDescription string
	if env := spec.GetEnvironment(); env != nil {
		environment = env.GetName()
		envDescription = env.GetDescription()
	}
	var shortDescription string
	if documentation := spec.GetDocumentation(); documentation != nil {
		shortDescription = documentation.GetShortDescription()
	}

	m := &models.Task{
		TaskKey:          ToTaskKey(taskId),
		Environment:      environment,
		FunctionName:     ExtractFunctionName(ctx, taskId.GetName(), environment),
		DeployedBy:       deployedBy,
		TaskSpec:         specBytes,
		EnvDescription:   newNullString(envDescription),
		ShortDescription: newNullString(shortDescription),
	}
	return m, nil
}

func TaskModelsToTaskDetailsWithoutIdentity(ctx context.Context, taskModels []*models.Task) ([]*task.TaskDetails, error) {
	// TODO(nary): Add identity enrichment after adding auth
	return TaskModelsToTaskDetails(ctx, taskModels)
}

// TaskModelsToTaskDetails transforms task database models to task.TaskDetails objects
func TaskModelsToTaskDetails(ctx context.Context, taskModels []*models.Task) ([]*task.TaskDetails, error) {
	// TODO(nary): Add identity enrichment after adding auth
	// subjectToIdentity, err := getEnrichedIdentities(ctx, taskModels, identityCache, enrichIdentities, enrichIdentitiesRemoteFallback)
	// if err != nil {
	// 	return nil, err
	// }

	tasks := make([]*task.TaskDetails, 0, len(taskModels))
	for _, m := range taskModels {
		taskIdl := &task.TaskDetails{
			TaskId: taskIdentifier(m),
			Metadata: &task.TaskMetadata{
				// TODO(nary): We should either store ShortName or rename this to FunctionName in the future
				ShortName:       m.FunctionName,
				EnvironmentName: m.Environment,
				DeployedAt:      timestamppb.New(m.CreatedAt),
				TriggersSummary: taskTriggersSummaryFromModel(m),
			},
		}
		if m.ShortDescription.Valid {
			taskIdl.Metadata.ShortDescription = m.ShortDescription.String
		}
		// TODO(nary): Add the deployed by identity if available after adding auth
		// if identity, ok := subjectToIdentity[m.DeployedBy]; ok {
		// 	taskIdl.Metadata.DeployedBy = identity
		// }

		// Parse task spec if available
		var spec task.TaskSpec
		err := proto.Unmarshal(m.TaskSpec, &spec)
		if err != nil {
			logger.Errorf(ctx, "failed to unmarshal task spec for task %v: %v", m.TaskKey, err)
			return nil, err
		}
		taskIdl.Spec = &spec
		tasks = append(tasks, taskIdl)
	}
	return tasks, nil
}

// TaskModelsToTasks uses the above TaskModelsToTaskDetails function to convert task models
// to task.TaskDetails objects, which are then simplified to task.Task objects.
func TaskModelsToTasks(ctx context.Context, taskModels []*models.Task, latestRuns map[models.TaskName]*models.Action) ([]*task.Task, error) {
	// TODO(nary): Add identity enrichment after adding auth
	// subjectToIdentity, err := getEnrichedIdentities(ctx, taskModels, identityCache, enrichIdentities, enrichIdentitiesRemoteFallback)
	// if err != nil {
	// 	return nil, err
	// }

	tasks := make([]*task.Task, 0, len(taskModels))
	for _, m := range taskModels {
		t := &task.Task{
			TaskId: taskIdentifier(m),
			Metadata: &task.TaskMetadata{
				// TODO(nary): We should either store ShortName or rename this to FunctionName in the future
				ShortName:       m.FunctionName,
				EnvironmentName: m.Environment,
				DeployedAt:      timestamppb.New(m.CreatedAt),
				TriggersSummary: taskTriggersSummaryFromModel(m),
			},
			TaskSummary: &task.TaskSummary{},
		}
		if m.ShortDescription.Valid {
			t.Metadata.ShortDescription = m.ShortDescription.String
		}
		// TODO(nary): Add the deployed by identity if available after adding auth
		// if identity, ok := subjectToIdentity[m.DeployedBy]; ok {
		// 	t.Metadata.DeployedBy = identity
		// }

		// Add latest run if available
		if latestRuns != nil {
			taskName := models.TaskName{
				Org:     m.Org,
				Project: m.Project,
				Domain:  m.Domain,
				Name:    m.Name,
			}
			if latestRun, ok := latestRuns[taskName]; ok {
				// TODO(nary): Implement ActionToLatestRunSummary after adding action transformers
				_ = latestRun
				// t.TaskSummary.LatestRun = ActionToLatestRunSummary(latestRun)
			}
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func taskTriggersSummaryFromModel(taskModel *models.Task) *task.TaskTriggersSummary {
	if taskModel.TotalTriggers == 0 {
		return nil
	}

	// TODO(nary): Add back trigger automation spec unmarshaling after adding trigger support
	if taskModel.TotalTriggers == 1 {
		// automationSpec, err := UnmarshalAutomationSpec(taskModel.TriggerAutomationSpec, "")
		// if err != nil {
		// 	logger.Errorf(context.Background(), "failed to unmarshal trigger automation spec: %v", err)
		// 	return nil
		// }

		return &task.TaskTriggersSummary{
			Summary: &task.TaskTriggersSummary_Details{
				Details: &task.TaskTriggersSummary_TriggerDetails{
					Name:   taskModel.TriggerName.String,
					Active: taskModel.ActiveTriggers > 0,
					// AutomationSpec: automationSpec,
				},
			},
		}
	}
	return &task.TaskTriggersSummary{
		Summary: &task.TaskTriggersSummary_Stats{
			Stats: &task.TaskTriggersSummary_TriggerStats{
				Total:  taskModel.TotalTriggers,
				Active: taskModel.ActiveTriggers,
			},
		},
	}
}

// ExtractFunctionName extracts the function name from a task name.
// If environment is set, it strips the "environment." prefix from the task name.
// Otherwise, it falls back to taking the last segment after splitting on ".".
func ExtractFunctionName(ctx context.Context, taskName, environment string) string {
	if environment != "" {
		prefix := environment + "."
		if functionName, ok := strings.CutPrefix(taskName, prefix); ok {
			return functionName
		}
		logger.Warnf(ctx, "Task name '%s' does not start with environment prefix '%s', falling back to extracting last segment", taskName, prefix)
	}
	// Fallback: take the last segment after splitting on "."
	segments := strings.Split(taskName, ".")
	return segments[len(segments)-1]
}

// VersionModelsToVersionResponses transforms version database models to task.ListVersionsResponse_VersionResponse objects.
func VersionModelsToVersionResponses(versionModels []*models.TaskVersion) []*task.ListVersionsResponse_VersionResponse {
	versions := make([]*task.ListVersionsResponse_VersionResponse, 0, len(versionModels))
	for _, m := range versionModels {
		versions = append(versions, &task.ListVersionsResponse_VersionResponse{
			Version:    m.Version,
			DeployedAt: timestamppb.New(m.CreatedAt),
		})
	}
	return versions
}

// Helper function to create a task identifier from a task model
func taskIdentifier(m *models.Task) *task.TaskIdentifier {
	return &task.TaskIdentifier{
		Org:     m.Org,
		Project: m.Project,
		Domain:  m.Domain,
		Name:    m.Name,
		Version: m.Version,
	}
}

// TaskListResultToTasksAndMetadata transforms a TaskListResult into tasks and metadata.
// Returns nil tasks and metadata if the input is nil.
func TaskListResultToTasksAndMetadata(ctx context.Context, result *models.TaskListResult,
	latestRuns map[models.TaskName]*models.Action, _ interface{}, _, _ bool) ([]*task.Task, *task.ListTasksResponse_ListTasksMetadata, error) {

	var taskModels []*models.Task
	var metadata *task.ListTasksResponse_ListTasksMetadata

	if result != nil {
		taskModels = result.Tasks
		metadata = &task.ListTasksResponse_ListTasksMetadata{
			Total:         result.Total,
			FilteredTotal: result.FilteredTotal,
		}
	}

	// TODO(nary): Add identity enrichment after adding auth (last 3 params)
	tasks, err := TaskModelsToTasks(ctx, taskModels, latestRuns)
	if err != nil {
		return nil, nil, err
	}

	return tasks, metadata, nil
}

// newNullString creates a sql.NullString from a string value
func newNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
