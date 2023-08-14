package config

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	schedulerModels "github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// TODO: add a way to get these list of tables directly from the gorm loaded models
var (
	tables = []string{"execution_events", "executions", "launch_plans", "named_entity_metadata",
		"node_execution_events", "node_executions", "projects", "resources", "schedulable_entities",
		"schedule_entities_snapshots", "task_executions", "tasks", "workflows", "description_entities"}
)

var LegacyMigrations = []*gormigrate.Migration{
	// Create projects table.
	{
		ID: "2019-05-22-projects",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Project{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("projects")
		},
	},
	// Create Task
	{
		ID: "2018-05-23-tasks",
		Migrate: func(tx *gorm.DB) error {
			// The gormigrate library recommends that we copy the actual struct into here for record-keeping but after
			// some internal discussion we've decided that that's not necessary. Just a history of what we've touched
			// when should be sufficient.
			return tx.AutoMigrate(&models.Task{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("tasks")
		},
	},
	// Create Workflow
	{
		ID: "2018-05-23-workflows",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Workflow{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("workflows")
		},
	},
	// Create Launch Plan table
	{
		ID: "2019-05-23-lp",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.LaunchPlan{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("launch_plans")
		},
	},

	// Create executions table
	{
		ID: "2019-05-23-executions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("executions")
		},
	},
	// Create executions events table
	{
		ID: "2019-01-29-executions-events",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.ExecutionEvent{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("executions_events")
		},
	},

	// Create node executions table
	{
		ID: "2019-04-17-node-executions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&NodeExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("node_executions")
		},
	},
	// Create node executions events table
	{
		ID: "2019-01-29-node-executions-events",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.NodeExecutionEvent{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("node_executions_events")
		},
	},
	// Create task executions table
	{
		ID: "2019-03-16-task-executions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&TaskExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("task_executions")
		},
	},
	// Update node executions with null parent values
	{
		ID: "2019-04-17-node-executions-backfill",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec("update node_executions set parent_task_execution_id = NULL where parent_task_execution_id = 0").Error
		},
	},
	// Update executions table to add cluster
	{
		ID: "2019-09-27-executions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE executions DROP COLUMN IF EXISTS cluster").Error
		},
	},
	// Update projects table to add description column
	{
		ID: "2019-10-09-project-description",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Project{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE projects DROP COLUMN IF EXISTS description").Error
		},
	},
	// Add offloaded URIs to table
	{
		ID: "2019-10-15-offload-inputs",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE executions DROP COLUMN IF EXISTS InputsURI, DROP COLUMN IF EXISTS UserInputsURI").Error
		},
	},
	// Create named_entity_metadata table.
	{
		ID: "2019-11-05-named-entity-metadata",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.NamedEntityMetadata{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("named_entity_metadata")
		},
	},
	// Add ProjectAttributes with custom resource attributes.
	{
		ID: "2020-01-10-resource",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Resource{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("resources")
		},
	},
	// Add Type to Task model.
	{
		ID: "2020-03-17-task-type",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Task{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE tasks DROP COLUMN IF EXISTS type").Error
		},
	},
	// Add state to name entity model
	{
		ID: "2020-04-03-named-entity-state",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.NamedEntityMetadata{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Table("named_entity_metadata").Migrator().DropColumn(&models.NamedEntityMetadata{}, "state")
		},
	},
	// Set default state value for workflow model
	{
		ID: "2020-04-03-named-entity-state-default",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec("UPDATE named_entity_metadata SET state = 0").Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("UPDATE named_entity_metadata set state = NULL").Error
		},
	},
	// Modify the workflows table, if necessary
	{
		ID: "2020-04-03-workflow-state",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE workflows DROP COLUMN IF EXISTS state").Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE workflows ADD COLUMN IF NOT EXISTS state integer;").Error
		},
	},
	// Modify the executions & node_execution table, if necessary
	{
		ID: "2020-04-29-executions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{}, &models.NodeExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Model(&models.Execution{}).Migrator().DropColumn(&models.Execution{}, "error_code"); err != nil {
				return err
			}
			if err := tx.Model(&models.Execution{}).Migrator().DropColumn(&models.Execution{}, "error_kind"); err != nil {
				return err
			}
			if err := tx.Model(&models.NodeExecution{}).Migrator().DropColumn(&models.NodeExecution{}, "error_code"); err != nil {
				return err
			}
			if err := tx.Model(&models.NodeExecution{}).Migrator().DropColumn(&models.NodeExecution{}, "error_kind"); err != nil {
				return err
			}
			return nil
		},
	},
	// Add TaskID to Execution model.
	{
		ID: "2020-04-14-task-type",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE executions DROP COLUMN IF EXISTS task_id").Error
		},
	},

	// NodeExecutions table has CacheStatus for Task nodes
	{
		ID: "2020-07-27-cachestatus",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.NodeExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Model(&models.NodeExecution{}).Migrator().DropColumn(&models.NodeExecution{}, "cache_status")
		},
	},
	{
		ID: "2020-07-31-node-execution",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.NodeExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Model(&models.NodeExecution{}).Migrator().DropColumn(&models.NodeExecution{}, "parent_id"); err != nil {
				return err
			}
			return tx.Model(&models.NodeExecution{}).Migrator().DropColumn(&models.NodeExecution{}, "node_execution_metadata")
		},
	},
	{
		ID: "2020-08-17-labels-addition",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Project{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Model(&models.Project{}).Migrator().DropColumn(&models.Project{}, "labels")
		},
	},
	{
		ID: "2020-09-01-task-exec-idx",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&TaskExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Model(&TaskExecution{}).Migrator().DropIndex(&TaskExecution{}, "idx_task_executions_exec")
		},
	},
	{
		ID: "2020-11-03-project-state-addition",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Project{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Model(&models.Project{}).Migrator().DropColumn(&models.Project{}, "state")
		},
	},
	{
		ID: "2020-11-03-project-state-default",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec("UPDATE projects set state = 0").Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("UPDATE projects set state = NULL").Error
		},
	},
	{
		ID: "2021-01-22-execution-user",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Model(&models.Execution{}).Migrator().DropColumn(&models.Execution{}, "user")
		},
	},
	{
		ID: "2021-04-19-node-execution_dynamic-workflow",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.NodeExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Model(&models.NodeExecution{}).Migrator().DropColumn(&models.NodeExecution{}, "dynamic_workflow_remote_closure_reference")
		},
	},

	{
		ID: "2021-07-22-schedulable_entities",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&schedulerModels.SchedulableEntity{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&schedulerModels.SchedulableEntity{}, "schedulable_entities")
		},
	},

	{
		ID: "2021-08-05-schedulable_entities_snapshot",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&schedulerModels.ScheduleEntitiesSnapshot{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&schedulerModels.ScheduleEntitiesSnapshot{}, "schedulable_entities_snapshot")
		},
	},

	// For any new table, Please use the following pattern due to a bug
	// in the postgres gorm layer https://github.com/go-gorm/postgres/issues/65
	{
		ID: "2022-01-11-id-to-bigint",
		Migrate: func(tx *gorm.DB) error {
			db, err := tx.DB()
			if err != nil {
				return err
			}
			return alterTableColumnType(db, "id", "bigint")
		},
		Rollback: func(tx *gorm.DB) error {
			db, err := tx.DB()
			if err != nil {
				return err
			}
			return alterTableColumnType(db, "id", "int")
		},
	},

	// Add state to execution model.
	{
		ID: "2022-01-11-execution-state",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Table("execution").Migrator().DropColumn(&models.Execution{}, "state")
		},
	},
	// Add internal data to node execution model.
	{
		ID: "2022-03-29-node-execution-internal-data",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&NodeExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Table("node_execution").Migrator().DropColumn(&NodeExecution{}, "internal_data")
		},
	},
	// Add created_at index to the execution model.
	{
		ID: "2022-04-04-execution-created-at-index",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Model(&models.Execution{}).Migrator().DropIndex(&models.Execution{}, "idx_executions_created_at")
		},
	},
	// Create description entities table
	{
		ID: "2022-09-13-description-entities",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.DescriptionEntity{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("description_entities")
		},
	},
	// Modify the tasks table, if necessary
	{
		ID: "2020-09-13-task-short_description",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE tasks ADD COLUMN IF NOT EXISTS short_description varchar(4000)").Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE tasks DROP COLUMN IF EXISTS short_description").Error
		},
	},
	// Modify the workflows table, if necessary
	{
		ID: "2020-09-13-workflow-short_description",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE workflows ADD COLUMN IF NOT EXISTS short_description varchar(4000)").Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE workflows DROP COLUMN IF EXISTS short_description").Error
		},
	},
	// Create signals table.
	{
		ID: "2022-04-11-signals",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Signal{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("signals")
		},
	},
	// Add the launch_type resource to the execution model
	{
		ID: "2022-12-09-execution-launch-type",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Model(&models.Execution{}).Migrator().DropColumn(&models.Execution{}, "launch_entity")
		},
	},
}

var NoopMigrations = []*gormigrate.Migration{
	/* The following is a series of Postgres specific migrations. They should mirror the state
	   of the database as of 2023 March. The rollback is a noop for everything because the migration itself should
	   be a noop.
	*/

	{
		ID: "pg-noop-2023-03-31-noop-project-3",
		Migrate: func(tx *gorm.DB) error {
			type Project struct {
				ID          uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt   time.Time  `gorm:"type:time"`
				UpdatedAt   time.Time  `gorm:"type:time"`
				DeletedAt   *time.Time `gorm:"index"`
				Identifier  string     `gorm:"primary_key"`
				Name        string     `valid:"length(0|255)"` // Human-readable name, not a unique identifier.
				Description string     `gorm:"type:varchar(300)"`
				Labels      []byte
				// GORM doesn't save the zero value for ints, so we use a pointer for the State field
				State *int32 `gorm:"default:0;index"`
			}
			return tx.AutoMigrate(&Project{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	// ALTER TABLE "projects" ALTER COLUMN "id" DROP NOT NULL otherwise.

	{
		ID: "pg-noop-2023-03-31-noop-task-2",
		Migrate: func(tx *gorm.DB) error {
			type Task struct {
				ID        uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time  `gorm:"type:time"`
				UpdatedAt time.Time  `gorm:"type:time"`
				DeletedAt *time.Time `gorm:"index"`
				Project   string     `gorm:"primary_key;index:task_project_domain_name_idx;index:task_project_domain_idx" valid:"length(0|255)"`
				Domain    string     `gorm:"primary_key;index:task_project_domain_name_idx;index:task_project_domain_idx" valid:"length(0|255)"`
				Name      string     `gorm:"primary_key;index:task_project_domain_name_idx" valid:"length(0|255)"`
				Version   string     `gorm:"primary_key" valid:"length(0|255)"`
				Closure   []byte     `gorm:"not null"`
				// Hash of the compiled task closure
				Digest []byte
				// Task type (also stored in the closure put promoted as a column for filtering).
				Type string `gorm:"" valid:"length(0|255)"`
				// ShortDescription for the task.
				ShortDescription string
			}
			return tx.AutoMigrate(&Task{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-workflow",
		Migrate: func(tx *gorm.DB) error {
			type Workflow struct {
				ID                      uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt               time.Time  `gorm:"type:time"`
				UpdatedAt               time.Time  `gorm:"type:time"`
				DeletedAt               *time.Time `gorm:"index"`
				Project                 string     `gorm:"primary_key;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
				Domain                  string     `gorm:"primary_key;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
				Name                    string     `gorm:"primary_key;index:workflow_project_domain_name_idx"  valid:"length(0|255)"`
				Version                 string     `gorm:"primary_key"`
				TypedInterface          []byte
				RemoteClosureIdentifier string `gorm:"not null" valid:"length(0|255)"`
				// Hash of the compiled workflow closure
				Digest []byte
				// ShortDescription for the workflow.
				ShortDescription string
			}
			return tx.AutoMigrate(&Workflow{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-launchplan",
		Migrate: func(tx *gorm.DB) error {
			type LaunchPlanScheduleType string

			type LaunchPlan struct {
				ID         uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt  time.Time  `gorm:"type:time"`
				UpdatedAt  time.Time  `gorm:"type:time"`
				DeletedAt  *time.Time `gorm:"index"`
				Project    string     `gorm:"primary_key;index:lp_project_domain_name_idx,lp_project_domain_idx" valid:"length(0|255)"`
				Domain     string     `gorm:"primary_key;index:lp_project_domain_name_idx,lp_project_domain_idx" valid:"length(0|255)"`
				Name       string     `gorm:"primary_key;index:lp_project_domain_name_idx" valid:"length(0|255)"`
				Version    string     `gorm:"primary_key" valid:"length(0|255)"`
				Spec       []byte     `gorm:"not null"`
				WorkflowID uint       `gorm:"index"`
				Closure    []byte     `gorm:"not null"`
				// GORM doesn't save the zero value for ints, so we use a pointer for the State field
				State *int32 `gorm:"default:0"`
				// Hash of the launch plan
				Digest       []byte
				ScheduleType LaunchPlanScheduleType
			}
			return tx.AutoMigrate(&LaunchPlan{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-namedentitymetadata",
		Migrate: func(tx *gorm.DB) error {
			type NamedEntityMetadata struct {
				ID           uint              `gorm:"index;autoIncrement;not null"`
				CreatedAt    time.Time         `gorm:"type:time"`
				UpdatedAt    time.Time         `gorm:"type:time"`
				DeletedAt    *time.Time        `gorm:"index"`
				ResourceType core.ResourceType `gorm:"primary_key;index:named_entity_metadata_type_project_domain_name_idx" valid:"length(0|255)"`
				Project      string            `gorm:"primary_key;index:named_entity_metadata_type_project_domain_name_idx" valid:"length(0|255)"`
				Domain       string            `gorm:"primary_key;index:named_entity_metadata_type_project_domain_name_idx" valid:"length(0|255)"`
				Name         string            `gorm:"primary_key;index:named_entity_metadata_type_project_domain_name_idx" valid:"length(0|255)"`
				Description  string            `gorm:"type:varchar(300)"`
				// GORM doesn't save the zero value for ints, so we use a pointer for the State field
				State *int32 `gorm:"default:0"`
			}

			return tx.AutoMigrate(&NamedEntityMetadata{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-execution",
		Migrate: func(tx *gorm.DB) error {
			type ExecutionKey struct {
				Project string `gorm:"primary_key;column:execution_project" valid:"length(0|255)"`
				Domain  string `gorm:"primary_key;column:execution_domain" valid:"length(0|255)"`
				Name    string `gorm:"primary_key;column:execution_name" valid:"length(0|255)"`
			}

			type Execution struct {
				ID        uint `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time
				UpdatedAt time.Time
				DeletedAt *time.Time `gorm:"index"`
				ExecutionKey
				LaunchPlanID uint   `gorm:"index"`
				WorkflowID   uint   `gorm:"index"`
				TaskID       uint   `gorm:"index"`
				Phase        string `valid:"length(0|255)"`
				Closure      []byte
				Spec         []byte `gorm:"not null"`
				StartedAt    *time.Time
				// Corresponds to the CreatedAt field in the Execution closure.
				// Prefixed with Execution to avoid clashes with gorm.Model CreatedAt
				ExecutionCreatedAt *time.Time `gorm:"index:idx_executions_created_at"`
				// Corresponds to the UpdatedAt field in the Execution closure
				// Prefixed with Execution to avoid clashes with gorm.Model UpdatedAt
				ExecutionUpdatedAt *time.Time
				Duration           time.Duration
				// In the case of an aborted execution this string may be non-empty.
				// It should be ignored for any other value of phase other than aborted.
				AbortCause string `valid:"length(0|255)"`
				// Corresponds to the execution mode used to trigger this execution
				Mode int32
				// The "parent" execution (if there is one) that is related to this execution.
				SourceExecutionID uint
				// The parent node execution if this was launched by a node
				ParentNodeExecutionID uint
				// Cluster where execution was triggered
				Cluster string `valid:"length(0|255)"`
				// Offloaded location of inputs LiteralMap. These are the inputs evaluated and contain applied defaults.
				InputsURI storage.DataReference
				// User specified inputs. This map might be incomplete and not include defaults applied
				UserInputsURI storage.DataReference
				// Execution Error Kind. nullable
				ErrorKind *string `gorm:"index"`
				// Execution Error Code nullable
				ErrorCode *string `valid:"length(0|255)"`
				// The user responsible for launching this execution.
				// This is also stored in the spec but promoted as a column for filtering.
				User string `gorm:"index" valid:"length(0|255)"`
				// GORM doesn't save the zero value for ints, so we use a pointer for the State field
				State *int32 `gorm:"index;default:0"`
				// The resource type of the entity used to launch the execution, one of 'launch_plan' or 'task'
				LaunchEntity string
			}

			return tx.AutoMigrate(&Execution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-taskexecution",
		Migrate: func(tx *gorm.DB) error {
			type TaskKey struct {
				Project string `gorm:"primary_key"`
				Domain  string `gorm:"primary_key"`
				Name    string `gorm:"primary_key"`
				Version string `gorm:"primary_key"`
			}
			type TaskExecutionKey struct {
				TaskKey
				Project string `gorm:"primary_key;column:execution_project;index:idx_task_executions_exec"`
				Domain  string `gorm:"primary_key;column:execution_domain;index:idx_task_executions_exec"`
				Name    string `gorm:"primary_key;column:execution_name;index:idx_task_executions_exec"`
				NodeID  string `gorm:"primary_key;index:idx_task_executions_exec;index"`
				// *IMPORTANT* This is a pointer to an int in order to allow setting an empty ("0") value according to gorm convention.
				// Because RetryAttempt is part of the TaskExecution primary key is should *never* be null.
				RetryAttempt *uint32 `gorm:"primary_key;AUTO_INCREMENT:FALSE"`
			}
			type TaskExecution struct {
				ID        uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time  `gorm:"type:time"`
				UpdatedAt time.Time  `gorm:"type:time"`
				DeletedAt *time.Time `gorm:"index"`
				TaskExecutionKey
				Phase        string `gorm:"type:text"`
				PhaseVersion uint32
				InputURI     string `gorm:"type:text"`
				Closure      []byte
				StartedAt    *time.Time
				// Corresponds to the CreatedAt field in the TaskExecution closure
				// This field is prefixed with TaskExecution because it signifies when
				// the execution was createdAt, not to be confused with gorm.Model.CreatedAt
				TaskExecutionCreatedAt *time.Time
				// Corresponds to the UpdatedAt field in the TaskExecution closure
				// This field is prefixed with TaskExecution because it signifies when
				// the execution was UpdatedAt, not to be confused with gorm.Model.UpdatedAt
				TaskExecutionUpdatedAt *time.Time
				Duration               time.Duration
				// The child node executions (if any) launched by this task execution.
				ChildNodeExecution []NodeExecution `gorm:"foreignkey:ParentTaskExecutionID;references:ID"`
			}

			return tx.AutoMigrate(&TaskExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		// This migration handles the necessary setup to change the type of the `parent_id` column in the node_executions table.
		ID: "pg-2023-05-02-fix-parentid-type-phase-1",
		Migrate: func(tx *gorm.DB) error {
			shouldMigrate, err := shouldApplyFixParentidMigration(tx)
			if err != nil {
				return err
			}
			if !shouldMigrate {
				return nil
			}

			// Alter table and add new column
			if err := tx.Exec("ALTER TABLE node_executions ADD COLUMN new_parent_id BIGINT;").Error; err != nil {
				return err
			}

			// Create trigger function
			triggerFunction := `
			CREATE FUNCTION set_new_parent_id() RETURNS TRIGGER AS
			$BODY$
			BEGIN
				NEW.new_parent_id := NEW.parent_id;
				RETURN NEW;
			END
			$BODY$ LANGUAGE PLPGSQL;
			`
			if err := tx.Exec(triggerFunction).Error; err != nil {
				return err
			}

			// Create trigger
			if err := tx.Exec("CREATE TRIGGER set_new_parent_id_trigger BEFORE INSERT OR UPDATE ON node_executions FOR EACH ROW EXECUTE PROCEDURE set_new_parent_id();").Error; err != nil {
				return err
			}

			// Update table
			if err := tx.Exec("UPDATE node_executions SET new_parent_id = parent_id WHERE parent_id is not null;").Error; err != nil {
				return err
			}

			// Create new index
			if err := tx.Exec("CREATE INDEX idx_node_executions_new_parent_id ON public.node_executions USING btree (new_parent_id);").Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop trigger and function
			if err := tx.Exec("DROP TRIGGER IF EXISTS set_new_parent_id_trigger ON node_executions;").Error; err != nil {
				return err
			}
			if err := tx.Exec("DROP FUNCTION IF EXISTS set_new_parent_id();").Error; err != nil {
				return err
			}
			// Drop column iff exists
			if err := tx.Exec("ALTER TABLE node_executions DROP COLUMN IF EXISTS new_parent_id;").Error; err != nil {
				return err
			}
			// Drop index idx_node_executions_new_parent_id
			if err := tx.Exec("DROP INDEX IF EXISTS idx_node_executions_new_parent_id;").Error; err != nil {
				return err
			}

			return nil
		},
	},
	{
		// This migration actually changes the type of the `parent_id` column in the node_executions table in a transaction.
		ID: "pg-2023-05-02-fix-parentid-type-phase-2",
		Migrate: func(tx *gorm.DB) error {
			shouldMigrate, err := shouldApplyFixParentidMigration(tx)
			if err != nil {
				return err
			}
			if !shouldMigrate {
				return nil
			}

			// Start transaction
			tx1 := tx.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx1.Rollback()
				}
			}()

			// Lock table
			if err := tx1.Exec("LOCK TABLE node_executions IN EXCLUSIVE MODE;").Error; err != nil {
				tx1.Rollback()
				return err
			}

			// DropIndex and create a new one
			if err := tx1.Exec("DROP INDEX idx_node_executions_parent_id;").Error; err != nil {
				tx1.Rollback()
				return err
			}

			// Drop and rename columns
			if err := tx1.Exec("ALTER TABLE node_executions DROP COLUMN parent_id;").Error; err != nil {
				tx1.Rollback()
				return err
			}

			// Rename idx_node_executions_new_parent_id to idx_node_executions_parent_id
			if err := tx1.Exec("ALTER INDEX idx_node_executions_new_parent_id RENAME TO idx_node_executions_parent_id;").Error; err != nil {
				tx1.Rollback()
				return err
			}

			if err := tx1.Exec("ALTER TABLE node_executions RENAME COLUMN new_parent_id TO parent_id;").Error; err != nil {
				tx1.Rollback()
				return err
			}

			// Drop trigger and function
			if err := tx1.Exec("DROP TRIGGER IF EXISTS set_new_parent_id_trigger ON node_executions;").Error; err != nil {
				tx1.Rollback()
				return err
			}
			if err := tx1.Exec("DROP FUNCTION IF EXISTS set_new_parent_id();").Error; err != nil {
				tx1.Rollback()
				return err
			}

			// Commit transaction
			if err := tx1.Commit().Error; err != nil {
				return err
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "pg-noop-2023-03-31-noop-nodeexecution",
		Migrate: func(tx *gorm.DB) error {
			type ExecutionKey struct {
				Project string `gorm:"primary_key;column:execution_project"`
				Domain  string `gorm:"primary_key;column:execution_domain"`
				Name    string `gorm:"primary_key;column:execution_name"`
			}

			type NodeExecutionKey struct {
				ExecutionKey
				NodeID string `gorm:"primary_key;index"`
			}
			type NodeExecution struct {
				ID        uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time  `gorm:"type:time"`
				UpdatedAt time.Time  `gorm:"type:time"`
				DeletedAt *time.Time `gorm:"index"`
				NodeExecutionKey
				// Also stored in the closure, but defined as a separate column because it's useful for filtering and sorting.
				Phase     string
				InputURI  string
				Closure   []byte
				StartedAt *time.Time
				// Corresponds to the CreatedAt field in the NodeExecution closure
				// Prefixed with NodeExecution to avoid clashes with gorm.Model CreatedAt
				NodeExecutionCreatedAt *time.Time
				// Corresponds to the UpdatedAt field in the NodeExecution closure
				// Prefixed with NodeExecution to avoid clashes with gorm.Model UpdatedAt
				NodeExecutionUpdatedAt *time.Time
				Duration               time.Duration
				// The task execution (if any) which launched this node execution.
				// TO BE DEPRECATED - as we have now introduced ParentID
				ParentTaskExecutionID uint `sql:"default:null" gorm:"index"`
				// The workflow execution (if any) which this node execution launched
				LaunchedExecution models.Execution `gorm:"foreignKey:ParentNodeExecutionID;references:ID"`
				// In the case of dynamic workflow nodes, the remote closure is uploaded to the path specified here.
				DynamicWorkflowRemoteClosureReference string
				// Metadata that is only relevant to the flyteadmin service that is used to parse the model and track additional attributes.
				InternalData          []byte
				NodeExecutionMetadata []byte
				// Parent that spawned this node execution - value is empty for executions at level 0
				ParentID *uint `sql:"default:null" gorm:"index"`
				// List of child node executions - for cases like Dynamic task, sub workflow, etc
				ChildNodeExecutions []NodeExecution `gorm:"foreignKey:ParentID;references:ID"`
				// Execution Error Kind. nullable, can be one of core.ExecutionError_ErrorKind
				ErrorKind *string `gorm:"index"`
				// Execution Error Code nullable. string value, but finite set determined by the execution engine and plugins
				ErrorCode *string
				// If the node is of Type Task, this should always exist for a successful execution, indicating the cache status for the execution
				CacheStatus *string
			}

			return tx.AutoMigrate(&NodeExecution{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-execution-event",
		Migrate: func(tx *gorm.DB) error {
			type ExecutionKey struct {
				Project string `gorm:"primary_key;column:execution_project" valid:"length(0|127)"`
				Domain  string `gorm:"primary_key;column:execution_domain" valid:"length(0|127)"`
				Name    string `gorm:"primary_key;column:execution_name" valid:"length(0|127)"`
			}
			type ExecutionEvent struct {
				ID        uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time  `gorm:"type:time"`
				UpdatedAt time.Time  `gorm:"type:time"`
				DeletedAt *time.Time `gorm:"index"`
				ExecutionKey
				RequestID  string `valid:"length(0|255)"`
				OccurredAt time.Time
				Phase      string `gorm:"primary_key"`
			}

			return tx.AutoMigrate(&ExecutionEvent{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-node-execution-event",
		Migrate: func(tx *gorm.DB) error {
			type ExecutionKey struct {
				Project string `gorm:"primary_key;column:execution_project" valid:"length(0|127)"`
				Domain  string `gorm:"primary_key;column:execution_domain" valid:"length(0|127)"`
				Name    string `gorm:"primary_key;column:execution_name" valid:"length(0|127)"`
			}
			type NodeExecutionKey struct {
				ExecutionKey
				NodeID string `gorm:"primary_key;index" valid:"length(0|180)"`
			}
			type NodeExecutionEvent struct {
				ID        uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time  `gorm:"type:time"`
				UpdatedAt time.Time  `gorm:"type:time"`
				DeletedAt *time.Time `gorm:"index"`
				NodeExecutionKey
				RequestID  string
				OccurredAt time.Time
				Phase      string `gorm:"primary_key"`
			}

			return tx.AutoMigrate(&NodeExecutionEvent{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-description-entity-2",
		Migrate: func(tx *gorm.DB) error {
			type DescriptionEntityKey struct {
				ResourceType core.ResourceType `gorm:"primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
				Project      string            `gorm:"primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
				Domain       string            `gorm:"primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
				Name         string            `gorm:"primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
				Version      string            `gorm:"primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
			}

			// SourceCode Database model to encapsulate a SourceCode.
			type SourceCode struct {
				Link string `valid:"length(0|255)"`
			}

			// DescriptionEntity Database model to encapsulate a DescriptionEntity.
			type DescriptionEntity struct {
				DescriptionEntityKey
				ID        uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time  `gorm:"type:time"`
				UpdatedAt time.Time  `gorm:"type:time"`
				DeletedAt *time.Time `gorm:"index"`
				SourceCode
				ShortDescription string
				LongDescription  []byte
			}

			return tx.AutoMigrate(&DescriptionEntity{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-signal",
		Migrate: func(tx *gorm.DB) error {
			type SignalKey struct {
				ExecutionKey
				SignalID string `gorm:"primary_key;index" valid:"length(0|255)"`
			}

			type Signal struct {
				ID        uint       `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time  `gorm:"type:time"`
				UpdatedAt time.Time  `gorm:"type:time"`
				DeletedAt *time.Time `gorm:"index"`
				SignalKey
				Type  []byte `gorm:"not null"`
				Value []byte
			}

			return tx.AutoMigrate(&Signal{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-resource",
		Migrate: func(tx *gorm.DB) error {
			type ResourcePriority int32

			// In this model, the combination of (Project, Domain, Workflow, LaunchPlan, ResourceType) is unique
			type Resource struct {
				ID           int64 `gorm:"AUTO_INCREMENT;column:id;primary_key;not null"`
				CreatedAt    time.Time
				UpdatedAt    time.Time
				DeletedAt    *time.Time `sql:"index"`
				Project      string     `gorm:"uniqueIndex:resource_idx" valid:"length(0|255)"`
				Domain       string     `gorm:"uniqueIndex:resource_idx" valid:"length(0|255)"`
				Workflow     string     `gorm:"uniqueIndex:resource_idx" valid:"length(0|255)"`
				LaunchPlan   string     `gorm:"uniqueIndex:resource_idx" valid:"length(0|255)"`
				ResourceType string     `gorm:"uniqueIndex:resource_idx" valid:"length(0|255)"`
				Priority     ResourcePriority
				// Serialized flyteidl.admin.MatchingAttributes.
				Attributes []byte
			}

			return tx.AutoMigrate(&Resource{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-schedulable_entities",
		Migrate: func(tx *gorm.DB) error {
			type SchedulableEntityKey struct {
				Project string `gorm:"primary_key"`
				Domain  string `gorm:"primary_key"`
				Name    string `gorm:"primary_key"`
				Version string `gorm:"primary_key"`
			}
			type SchedulableEntity struct {
				ID        uint `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time
				UpdatedAt time.Time
				DeletedAt *time.Time `gorm:"index"`
				SchedulableEntityKey
				CronExpression      string
				FixedRateValue      uint32
				Unit                admin.FixedRateUnit
				KickoffTimeInputArg string
				Active              *bool
			}

			return tx.AutoMigrate(&SchedulableEntity{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "pg-noop-2023-03-31-noop-schedulable_entities-snapshot",
		Migrate: func(tx *gorm.DB) error {
			type ScheduleEntitiesSnapshot struct {
				ID        uint `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time
				UpdatedAt time.Time
				DeletedAt *time.Time `gorm:"index"`
				Snapshot  []byte     `gorm:"column:snapshot" schema:"-"`
			}

			return tx.AutoMigrate(&ScheduleEntitiesSnapshot{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "2023-06-19-id-sequence-to-bigint",
		Migrate: func(tx *gorm.DB) error {
			db, err := tx.DB()
			if err != nil {
				return err
			}
			return alterIDSequenceType(db)
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "2023-08-04-admin-tags",
		Migrate: func(tx *gorm.DB) error {
			type AdminTag struct {
				gorm.Model
				Name string `gorm:"index:,unique;size:255"`
			}

			return tx.AutoMigrate(&AdminTag{})
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},

	{
		ID: "2023-08-04-execution-admin-tags", // A join table used to associate executions with tags
		Migrate: func(tx *gorm.DB) error {
			type AdminTag struct {
				gorm.Model
				Name string `gorm:"index:,unique;size:255"`
			}

			type ExecutionKey struct {
				Project string `gorm:"primary_key;column:execution_project" valid:"length(0|255)"`
				Domain  string `gorm:"primary_key;column:execution_domain" valid:"length(0|255)"`
				Name    string `gorm:"primary_key;column:execution_name" valid:"length(0|255)"`
			}

			type Execution struct {
				ID        uint `gorm:"index;autoIncrement;not null"`
				CreatedAt time.Time
				UpdatedAt time.Time
				DeletedAt *time.Time `gorm:"index"`
				ExecutionKey
				LaunchPlanID uint   `gorm:"index"`
				WorkflowID   uint   `gorm:"index"`
				TaskID       uint   `gorm:"index"`
				Phase        string `valid:"length(0|255)"`
				Closure      []byte
				Spec         []byte `gorm:"not null"`
				StartedAt    *time.Time
				// Corresponds to the CreatedAt field in the Execution closure.
				// Prefixed with Execution to avoid clashes with gorm.Model CreatedAt
				ExecutionCreatedAt *time.Time `gorm:"index:idx_executions_created_at"`
				// Corresponds to the UpdatedAt field in the Execution closure
				// Prefixed with Execution to avoid clashes with gorm.Model UpdatedAt
				ExecutionUpdatedAt *time.Time
				Duration           time.Duration
				// In the case of an aborted execution this string may be non-empty.
				// It should be ignored for any other value of phase other than aborted.
				AbortCause string `valid:"length(0|255)"`
				// Corresponds to the execution mode used to trigger this execution
				Mode int32
				// The "parent" execution (if there is one) that is related to this execution.
				SourceExecutionID uint
				// The parent node execution if this was launched by a node
				ParentNodeExecutionID uint
				// Cluster where execution was triggered
				Cluster string `valid:"length(0|255)"`
				// Offloaded location of inputs LiteralMap. These are the inputs evaluated and contain applied defaults.
				InputsURI storage.DataReference
				// User specified inputs. This map might be incomplete and not include defaults applied
				UserInputsURI storage.DataReference
				// Execution Error Kind. nullable
				ErrorKind *string `gorm:"index"`
				// Execution Error Code nullable
				ErrorCode *string `valid:"length(0|255)"`
				// The user responsible for launching this execution.
				// This is also stored in the spec but promoted as a column for filtering.
				User string `gorm:"index" valid:"length(0|255)"`
				// GORM doesn't save the zero value for ints, so we use a pointer for the State field
				State *int32 `gorm:"index;default:0"`
				// The resource type of the entity used to launch the execution, one of 'launch_plan' or 'task'
				LaunchEntity string
				// Tags associated with the execution
				Tags []AdminTag `gorm:"many2many:execution_admin_tags;"`
			}

			return tx.AutoMigrate(&Execution{})
		},
	},
}

var Migrations = append(LegacyMigrations, NoopMigrations...)

func alterTableColumnType(db *sql.DB, columnName, columnType string) error {
	var err error
	for _, table := range tables {
		if _, err = db.Exec(fmt.Sprintf(`ALTER TABLE IF EXISTS %s ALTER COLUMN "%s" TYPE %s`, table, columnName,
			columnType)); err != nil {
			return err
		}
	}
	return nil
}

func alterIDSequenceType(db *sql.DB) error {
	var err error
	for _, table := range tables {
		if _, err = db.Exec(fmt.Sprintf(`ALTER SEQUENCE IF EXISTS %s_id_seq AS bigint NO MAXVALUE`, table)); err != nil {
			return err
		}
	}
	return nil
}

func shouldApplyFixParentidMigration(db *gorm.DB) (bool, error) {
	// This only applies to postgres and in the case of the node_executions table contains a
	// column named parent_id of type `integer` instead of `bigint`.
	if db.Dialector.Name() != "postgres" {
		return false, nil
	}

	// We should only apply this migration in case the type of the parent_id column is integer
	var columnType string
	query := `
	SELECT data_type
	FROM information_schema.columns
	WHERE table_name = ? AND column_name = ?;
	`
	err := db.Raw(query, "node_executions", "parent_id").Scan(&columnType).Error
	if err != nil {
		return false, err
	}
	if columnType == "bigint" {
		return false, nil
	}

	return true, nil
}
