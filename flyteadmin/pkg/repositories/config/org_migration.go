package config

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

var (
	inventoryEntities   = []string{"task", "workflow"}
	launchPlan          = "launch_plan"
	executions          = "executions"
	nodeExecutions      = "node_executions"
	taskExecutions      = "task_executions"
	executionEvents     = "execution_events"
	nodeExecutionEvents = "node_execution_events"
	executionAdminTags  = "execution_admin_tags"
	resources           = "resources"
	signals             = "signals"
	adminTags           = "admin_tags"
	descriptionEntity   = "description_entities"
	namedEntityMetadata = "named_entity_metadata"
	projects            = "projects"
	schedulableEntities = "schedulable_entities"
)

const (
	entityTemplate = "{{entity}}"
	indexTemplate  = "{{idx}}"
)

// inventory items constants
const (
	addOrgInventoryColumnFmt = "ALTER TABLE {{entity}}s ADD COLUMN IF NOT EXISTS org text default '';"
	addOrgColumnFmt          = "ALTER TABLE {{entity}} ADD COLUMN IF NOT EXISTS org text default '';"

	addOrgInventoryPrimaryKey = "ALTER TABLE {{entity}}s DROP CONSTRAINT {{entity}}s_pkey cascade, add primary key (org, project, domain, name, version); "

	projectDomainNameIndexFmt = "{{entity}}_project_domain_name_idx"
	projectDomainIndexFmt     = "{{entity}}_project_domain_idx"

	addProjectDomainNameIndexFmt = "CREATE INDEX CONCURRENTLY new_{{idx}} ON {{entity}}s (org, project, domain, name);"
	addProjectDomainIndexFmt     = "CREATE INDEX CONCURRENTLY new_{{idx}} ON {{entity}}s (org, project, domain);"

	dropOldIndexFmt   = "DROP INDEX {{idx}};"
	renameNewIndexFmt = "ALTER INDEX new_{{idx}} RENAME TO {{idx}};"
)

const (
	dropOrgColumn          = "ALTER TABLE {{entity}} DROP COLUMN IF EXISTS org"
	dropExecutionOrgColumn = "ALTER TABLE {{entity}} DROP COLUMN IF EXISTS execution_org"
	dropOrgInventoryColumn = "ALTER TABLE {{entity}}s DROP COLUMN IF EXISTS org; "

	restoreProjectDomainNameIndexFmt = "CREATE INDEX CONCURRENTLY new_{{idx}} ON {{entity}}s (project, domain, name);"
	restoreProjectDomainIndexFmt     = "CREATE INDEX CONCURRENTLY new_{{idx}} ON {{entity}}s (project, domain);"

	dropOrgInventoryPrimaryKey = "ALTER TABLE {{entity}}s DROP CONSTRAINT {{entity}}s_pkey cascade, add primary key (project, domain, name, version); "
)

// execution items constants
const (
	addExecutionOrgColumnFmt = "ALTER TABLE {{entity}} ADD COLUMN IF NOT EXISTS execution_org text default '';"
)

func renameIndex(tx *gorm.DB, indexName string) error {
	tx1 := tx.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx1.Rollback()
		}
	}()

	if err := tx1.Exec(strings.ReplaceAll(dropOldIndexFmt, indexTemplate, indexName)).Error; err != nil {
		tx1.Rollback()
		return err
	}
	if err := tx1.Exec(strings.ReplaceAll(renameNewIndexFmt, indexTemplate, indexName)).Error; err != nil {
		tx1.Rollback()
		return err
	}

	// Commit transaction
	if err := tx1.Commit().Error; err != nil {
		return err
	}
	return nil
}

// Creates a new org index with composite keys (org, project, domain, name, version)
// Migrates existing indices on project, domain, (name) to also include org
func migrateInventoryEntity(tx *gorm.DB, entity string) error {
	// add org column
	addOrgSQL := strings.ReplaceAll(addOrgInventoryColumnFmt, entityTemplate, entity)
	if err := tx.Exec(addOrgSQL).Error; err != nil {
		return err
	}

	// We can't create new indices in a transaction
	addOrgPkeySQL := strings.ReplaceAll(addOrgInventoryPrimaryKey, entityTemplate, entity)
	if err := tx.Exec(addOrgPkeySQL).Error; err != nil {
		return err
	}

	// add project domain index
	projectDomainIndex := strings.ReplaceAll(projectDomainIndexFmt, entityTemplate, entity)
	addProjectDomainIndex := strings.ReplaceAll(addProjectDomainIndexFmt, indexTemplate, projectDomainIndex)
	addProjectDomainIndex = strings.ReplaceAll(addProjectDomainIndex, entityTemplate, entity)
	err := tx.Exec(addProjectDomainIndex).Error
	if err != nil {
		return err
	}

	// add project domain name index
	projectDomainNameIndex := strings.ReplaceAll(projectDomainNameIndexFmt, entityTemplate, entity)
	addProjectDomainNameIndex := strings.ReplaceAll(addProjectDomainNameIndexFmt, indexTemplate, projectDomainNameIndex)
	addProjectDomainNameIndex = strings.ReplaceAll(addProjectDomainNameIndex, entityTemplate, entity)
	err = tx.Exec(addProjectDomainNameIndex).Error
	if err != nil {
		return err
	}

	tx1 := tx.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx1.Rollback()
		}
	}()

	for _, index := range []string{projectDomainIndex, projectDomainNameIndex} {
		if err := tx1.Exec(strings.ReplaceAll(dropOldIndexFmt, indexTemplate, index)).Error; err != nil {
			tx1.Rollback()
			return err
		}
		if err := tx1.Exec(strings.ReplaceAll(renameNewIndexFmt, indexTemplate, index)).Error; err != nil {
			tx1.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx1.Commit().Error; err != nil {
		return err
	}
	return nil
}

func migrateLaunchPlans(tx *gorm.DB) error {
	// add org column
	addOrgSQL := strings.ReplaceAll(addOrgInventoryColumnFmt, entityTemplate, launchPlan)
	if err := tx.Exec(addOrgSQL).Error; err != nil {
		return err
	}

	addOrgPkeySQL := strings.ReplaceAll(addOrgInventoryPrimaryKey, entityTemplate, launchPlan)
	if err := tx.Exec(addOrgPkeySQL).Error; err != nil {
		return err
	}

	indexName := "lp_project_domain_name_idx"
	// update task execution index
	err := tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON launch_plans (org, project, domain, name);", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	return nil
}

func migrateExecutions(tx *gorm.DB) error {
	// add org column
	if err := tx.Exec(strings.ReplaceAll(addExecutionOrgColumnFmt, entityTemplate, executions)).Error; err != nil {
		return err
	}

	// add org primary key
	err := tx.Exec("ALTER TABLE executions drop constraint executions_pkey cascade, add primary key (execution_org, execution_project, execution_domain, execution_name);").Error
	if err != nil {
		return err
	}

	return nil
}

func migrateExecutionEvents(tx *gorm.DB) error {
	// add org column
	if err := tx.Exec(strings.ReplaceAll(addExecutionOrgColumnFmt, entityTemplate, executionEvents)).Error; err != nil {
		return err
	}

	// add org primary key
	err := tx.Exec("ALTER TABLE execution_events drop constraint execution_events_pkey cascade, add primary key (execution_org, execution_project, execution_domain, execution_name, phase);").Error
	if err != nil {
		return err
	}

	return nil
}

func migrateExecutionTags(tx *gorm.DB) error {
	// add org column
	if err := tx.Exec(strings.ReplaceAll(addExecutionOrgColumnFmt, entityTemplate, executionAdminTags)).Error; err != nil {
		return err
	}

	// add org primary key
	err := tx.Exec("ALTER TABLE execution_admin_tags drop constraint execution_admin_tags_pkey cascade, add primary key (execution_org, execution_project, execution_domain, execution_name, admin_tag_id);").Error
	if err != nil {
		return err
	}

	return nil
}

func migrateAdminTags(tx *gorm.DB) error {
	if err := addOrgColumn(tx, adminTags); err != nil {
		return err
	}

	// No composite primary key for this table

	indexName := "idx_admin_tags_name"
	// update description entity index
	err := tx.Exec(fmt.Sprintf("CREATE UNIQUE INDEX CONCURRENTLY new_%s ON admin_tags (org, name)", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	return nil
}

func migrateNodeExecutions(tx *gorm.DB) error {
	// add org column
	if err := tx.Exec(strings.ReplaceAll(addExecutionOrgColumnFmt, entityTemplate, nodeExecutions)).Error; err != nil {
		return err
	}

	// add org primary key
	err := tx.Exec("ALTER TABLE node_executions drop constraint node_executions_pkey cascade, add primary key (execution_org, execution_project, execution_domain, execution_name, node_id);").Error
	if err != nil {
		return err
	}

	return nil
}

func migrateNodeExecutionEvents(tx *gorm.DB) error {
	// add org column
	if err := tx.Exec(strings.ReplaceAll(addExecutionOrgColumnFmt, entityTemplate, nodeExecutionEvents)).Error; err != nil {
		return err
	}

	// add org primary key
	err := tx.Exec("ALTER TABLE node_execution_events drop constraint node_execution_events_pkey cascade, add primary key (execution_org, execution_project, execution_domain, execution_name, node_id, phase);").Error
	if err != nil {
		return err
	}

	indexName := "idx_node_execution_events_node_id"

	// update task execution index
	err = tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON node_execution_events (execution_org, node_id);", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	return nil
}

func migrateTaskExecutions(tx *gorm.DB) error {
	// add execution_org column
	if err := tx.Exec(strings.ReplaceAll(addExecutionOrgColumnFmt, entityTemplate, taskExecutions)).Error; err != nil {
		return err
	}

	// add org column
	addOrgSQL := strings.ReplaceAll(addOrgInventoryColumnFmt, entityTemplate, "task_execution")
	if err := tx.Exec(addOrgSQL).Error; err != nil {
		return err
	}

	// add org primary key
	err := tx.Exec("ALTER TABLE task_executions drop constraint task_executions_pkey cascade, add primary key (org, project, domain, name, version, execution_org, execution_project, execution_domain, execution_name, node_id, retry_attempt);").Error
	if err != nil {
		return err
	}

	indexName := "idx_task_executions_exec"

	// update task execution index
	err = tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON task_executions (execution_org, execution_project, execution_domain, execution_name, node_id);", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	return nil
}

func addOrgColumn(tx *gorm.DB, entity string) error {
	addOrgSQL := strings.ReplaceAll(addOrgColumnFmt, entityTemplate, entity)
	return tx.Exec(addOrgSQL).Error
}

func migrateDescriptionEntities(tx *gorm.DB) error {
	if err := addOrgColumn(tx, descriptionEntity); err != nil {
		return err
	}

	if err := tx.Exec("ALTER TABLE description_entities drop constraint description_entities_pkey cascade, add primary key (resource_type, org, project, domain, name, version);").Error; err != nil {
		return err
	}
	indexName := "description_entity_project_domain_name_version_idx"
	// update description entity index
	err := tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON description_entities (resource_type, org, project, domain, name, version)", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	return nil
}

func migrateNamedEntities(tx *gorm.DB) error {
	if err := addOrgColumn(tx, namedEntityMetadata); err != nil {
		return err
	}

	if err := tx.Exec("ALTER TABLE named_entity_metadata drop constraint named_entity_metadata_pkey cascade, add primary key (resource_type, org, project, domain, name);").Error; err != nil {
		return err
	}
	indexName := "named_entity_metadata_type_project_domain_name_idx"
	// update description entity index
	err := tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON named_entity_metadata (resource_type, org, project, domain, name)", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	return nil
}

func migrateProjects(tx *gorm.DB) error {
	// add org column
	if err := addOrgColumn(tx, projects); err != nil {
		return err
	}

	if err := tx.Exec("ALTER TABLE projects drop constraint projects_pkey cascade, add primary key (org, identifier);").Error; err != nil {
		return err
	}

	return nil
}

func migrateResources(tx *gorm.DB) error {
	// add org column
	if err := addOrgColumn(tx, resources); err != nil {
		return err
	}

	// No composite primary key for this table

	indexName := "resource_idx"
	err := tx.Exec(fmt.Sprintf("CREATE UNIQUE INDEX CONCURRENTLY new_%s ON resources (org, project, domain, workflow, launch_plan, resource_type)", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	return nil
}

func migrateSignals(tx *gorm.DB) error {
	// add execution_org column
	if err := tx.Exec(strings.ReplaceAll(addExecutionOrgColumnFmt, entityTemplate, signals)).Error; err != nil {
		return err
	}

	if err := tx.Exec("ALTER TABLE signals drop constraint signals_pkey cascade, add primary key (execution_org, execution_project, execution_domain, execution_name, signal_id);").Error; err != nil {
		return err
	}

	return nil
}

func migrateSchedules(tx *gorm.DB) error {
	// add org column
	if err := addOrgColumn(tx, schedulableEntities); err != nil {
		return err
	}

	if err := tx.Exec("ALTER TABLE schedulable_entities drop constraint schedulable_entities_pkey cascade, add primary key (org, project, domain, name, version);").Error; err != nil {
		return err
	}

	return nil
}

func migrateAddOrg(tx *gorm.DB) error {

	shouldMigrate, err := shouldApplyOrgMigrations(tx)
	if err != nil {
		return err
	}
	if !shouldMigrate {
		logger.Infof(context.TODO(), "not running org migrations, detected org column present in tasks table")
		return nil
	}

	// tasks, workflows
	for _, inventoryEntity := range inventoryEntities {
		if err := migrateInventoryEntity(tx, inventoryEntity); err != nil {
			return err
		}
	}

	if err := migrateLaunchPlans(tx); err != nil {
		return err
	}

	if err := migrateExecutions(tx); err != nil {
		return err
	}

	if err := migrateExecutionEvents(tx); err != nil {
		return err
	}

	if err := migrateExecutionTags(tx); err != nil {
		return err
	}

	if err := migrateAdminTags(tx); err != nil {
		return err
	}

	if err := migrateNodeExecutions(tx); err != nil {
		return err
	}

	if err := migrateNodeExecutionEvents(tx); err != nil {
		return err
	}

	if err := migrateTaskExecutions(tx); err != nil {
		return err
	}

	if err := migrateDescriptionEntities(tx); err != nil {
		return err
	}

	if err := migrateNamedEntities(tx); err != nil {
		return err
	}

	if err := migrateProjects(tx); err != nil {
		return err
	}

	if err := migrateResources(tx); err != nil {
		return err
	}

	if err := migrateSignals(tx); err != nil {
		return err
	}

	if err := migrateSchedules(tx); err != nil {
		return err
	}
	return nil
}

func rollbackInventoryEntity(tx *gorm.DB, entity string) error {
	// restore project domain index
	projectDomainIndex := strings.ReplaceAll(projectDomainIndexFmt, entityTemplate, entity)
	restoreProjectDomainIndex := strings.ReplaceAll(restoreProjectDomainIndexFmt, indexTemplate, projectDomainIndex)
	restoreProjectDomainIndex = strings.ReplaceAll(restoreProjectDomainIndex, entityTemplate, entity)
	err := tx.Exec(restoreProjectDomainIndex).Error
	if err != nil {
		return err
	}

	// restore project domain name index
	projectDomainNameIndex := strings.ReplaceAll(projectDomainNameIndexFmt, entityTemplate, entity)
	restoreProjectDomainNameIndex := strings.ReplaceAll(restoreProjectDomainNameIndexFmt, indexTemplate, projectDomainNameIndex)
	restoreProjectDomainNameIndex = strings.ReplaceAll(restoreProjectDomainNameIndex, entityTemplate, entity)
	err = tx.Exec(restoreProjectDomainNameIndex).Error
	if err != nil {
		return err
	}

	tx1 := tx.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx1.Rollback()
		}
	}()

	for _, index := range []string{projectDomainIndex, projectDomainNameIndex} {
		if err := tx1.Exec(strings.ReplaceAll(dropOldIndexFmt, indexTemplate, index)).Error; err != nil {
			tx1.Rollback()
			return err
		}
		if err := tx1.Exec(strings.ReplaceAll(renameNewIndexFmt, indexTemplate, index)).Error; err != nil {
			tx1.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx1.Commit().Error; err != nil {
		return err
	}

	// We can't create new indices in a transaction
	addOrgPkeySQL := strings.ReplaceAll(dropOrgInventoryPrimaryKey, entityTemplate, entity)
	if err := tx.Exec(addOrgPkeySQL).Error; err != nil {
		return err
	}

	// Can't drop a column in a transaction
	dropOrgColumnSQL := strings.ReplaceAll(dropOrgInventoryColumn, entityTemplate, entity)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}
	return nil
}

func rollbackLaunchPlans(tx *gorm.DB) error {
	indexName := "lp_project_domain_name_idx"
	// update task execution index
	err := tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON launch_plans (project, domain, name);", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	// We can't create new indices in a transaction
	addOrgPkeySQL := strings.ReplaceAll(dropOrgInventoryPrimaryKey, entityTemplate, launchPlan)
	if err := tx.Exec(addOrgPkeySQL).Error; err != nil {
		return err
	}

	// Can't drop a column in a transaction
	dropOrgColumnSQL := strings.ReplaceAll(dropOrgInventoryColumn, entityTemplate, launchPlan)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}

	return nil
}

func rollbackExecutions(tx *gorm.DB) error {
	// drop org from primary key
	err := tx.Exec("ALTER TABLE executions drop constraint executions_pkey cascade, add primary key (execution_project, execution_domain, execution_name);").Error
	if err != nil {
		return err
	}

	dropOrgColumnSQL := strings.ReplaceAll(dropExecutionOrgColumn, entityTemplate, executions)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}

	return nil
}

func rollbackExecutionEvents(tx *gorm.DB) error {
	// drop org from primary key
	err := tx.Exec("ALTER TABLE execution_events drop constraint execution_events_pkey cascade, add primary key (execution_project, execution_domain, execution_name, phase);").Error
	if err != nil {
		return err
	}

	dropOrgColumnSQL := strings.ReplaceAll(dropExecutionOrgColumn, entityTemplate, executionEvents)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}

	return nil
}

func rollbackExecutionTags(tx *gorm.DB) error {
	// drop org from primary key
	err := tx.Exec("ALTER TABLE execution_admin_tags drop constraint execution_admin_tags_pkey cascade, add primary key (execution_project, execution_domain, execution_name, admin_tag_id);").Error
	if err != nil {
		return err
	}

	dropOrgColumnSQL := strings.ReplaceAll(dropExecutionOrgColumn, entityTemplate, executionAdminTags)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}

	return nil
}

func rollbackAdminTags(tx *gorm.DB) error {
	// No composite primary key for this table
	indexName := "idx_admin_tags_name"
	// update description entity index
	err := tx.Exec(fmt.Sprintf("CREATE UNIQUE INDEX CONCURRENTLY new_%s ON admin_tags (name)", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}
	dropOrgColumnSQL := strings.ReplaceAll(dropOrgColumn, entityTemplate, adminTags)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}

	return nil
}

func rollbackNodeExecutions(tx *gorm.DB) error {
	// drop org from primary key
	err := tx.Exec("ALTER TABLE node_executions drop constraint node_executions_pkey cascade, add primary key (execution_project, execution_domain, execution_name, node_id);").Error
	if err != nil {
		return err
	}

	dropOrgColumnSQL := strings.ReplaceAll(dropExecutionOrgColumn, entityTemplate, nodeExecutions)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}

	return nil
}

func rollbackNodeExecutionEvents(tx *gorm.DB) error {
	indexName := "idx_node_execution_events_node_id"

	// update task execution index
	err := tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON node_execution_events (node_id);", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	// drop org from primary key
	err = tx.Exec("ALTER TABLE node_execution_events drop constraint node_execution_events_pkey cascade, add primary key (execution_project, execution_domain, execution_name, node_id, phase);").Error
	if err != nil {
		return err
	}

	dropOrgColumnSQL := strings.ReplaceAll(dropExecutionOrgColumn, entityTemplate, nodeExecutionEvents)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}

	return nil
}

func rollbackTaskExecutions(tx *gorm.DB) error {
	indexName := "idx_task_executions_exec"

	// update task execution index
	err := tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON task_executions (execution_project, execution_domain, execution_name, node_id);", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	// drop org from primary key
	err = tx.Exec("ALTER TABLE task_executions drop constraint task_executions_pkey cascade, add primary key (project, domain, name, version, execution_project, execution_domain, execution_name, node_id, retry_attempt);").Error
	if err != nil {
		return err
	}

	dropExecutionOrgColumnSQL := strings.ReplaceAll(dropExecutionOrgColumn, entityTemplate, taskExecutions)
	if err := tx.Exec(dropExecutionOrgColumnSQL).Error; err != nil {
		return err
	}

	dropOrgColumnSQL := strings.ReplaceAll(dropOrgColumn, entityTemplate, taskExecutions)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}

	return nil

}

func execDropOrgColumn(tx *gorm.DB, entity string) error {
	dropOrgColumnSQL := strings.ReplaceAll(dropOrgColumn, entityTemplate, entity)
	return tx.Exec(dropOrgColumnSQL).Error
}

func rollbackDescriptionEntities(tx *gorm.DB) error {
	indexName := "description_entity_project_domain_name_version_idx"
	// update description entity index
	err := tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON description_entities (resource_type, project, domain, name, version)", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	// drop org from primary key
	if err := tx.Exec("ALTER TABLE description_entities drop constraint description_entities_pkey cascade, add primary key (resource_type, project, domain, name, version);").Error; err != nil {
		return err
	}

	if err := execDropOrgColumn(tx, descriptionEntity); err != nil {
		return err
	}

	return nil
}

func rollbackNamedEntities(tx *gorm.DB) error {
	indexName := "named_entity_metadata_type_project_domain_name_idx"
	// update description entity index
	err := tx.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY new_%s ON named_entity_metadata (resource_type, project, domain, name)", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	// drop org from primary key
	if err := tx.Exec("ALTER TABLE named_entity_metadata drop constraint named_entity_metadata_pkey cascade, add primary key (resource_type, project, domain, name);").Error; err != nil {
		return err
	}

	if err := execDropOrgColumn(tx, namedEntityMetadata); err != nil {
		return err
	}

	return nil
}

func rollbackProjects(tx *gorm.DB) error {

	// drop org from primary key
	if err := tx.Exec("ALTER TABLE projects drop constraint projects_pkey cascade, add primary key (identifier);").Error; err != nil {
		return err
	}

	if err := execDropOrgColumn(tx, projects); err != nil {
		return err
	}
	return nil
}

func rollbackResources(tx *gorm.DB) error {
	indexName := "resource_idx"
	err := tx.Exec(fmt.Sprintf("CREATE UNIQUE INDEX CONCURRENTLY new_%s ON resources (project, domain, workflow, launch_plan, resource_type)", indexName)).Error
	if err != nil {
		return err
	}
	if err := renameIndex(tx, indexName); err != nil {
		return err
	}

	// No composite primary key for this table

	if err := execDropOrgColumn(tx, resources); err != nil {
		return err
	}
	return nil
}

func rollbackSignals(tx *gorm.DB) error {
	// drop org from primary key
	if err := tx.Exec("ALTER TABLE signals drop constraint signals_pkey cascade, add primary key (execution_project, execution_domain, execution_name, signal_id);").Error; err != nil {
		return err
	}

	dropOrgColumnSQL := strings.ReplaceAll(dropExecutionOrgColumn, entityTemplate, signals)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}
	return nil
}

func rollbackSchedules(tx *gorm.DB) error {
	// drop org from primary key
	if err := tx.Exec("ALTER TABLE schedulable_entities drop constraint schedulable_entities_pkey cascade, add primary key (project, domain, name, version);").Error; err != nil {
		return err
	}

	dropOrgColumnSQL := strings.ReplaceAll(dropOrgColumn, entityTemplate, schedulableEntities)
	if err := tx.Exec(dropOrgColumnSQL).Error; err != nil {
		return err
	}
	return nil
}

func rollbackAddOrg(tx *gorm.DB) error {
	for _, inventoryEntity := range inventoryEntities {
		if err := rollbackInventoryEntity(tx, inventoryEntity); err != nil {
			return err
		}
	}

	if err := rollbackLaunchPlans(tx); err != nil {
		return err
	}

	if err := rollbackExecutions(tx); err != nil {
		return err
	}
	if err := rollbackExecutionEvents(tx); err != nil {
		return err
	}
	if err := rollbackExecutionTags(tx); err != nil {
		return err
	}
	if err := rollbackAdminTags(tx); err != nil {
		return err
	}
	if err := rollbackNodeExecutions(tx); err != nil {
		return err
	}
	if err := rollbackNodeExecutionEvents(tx); err != nil {
		return err
	}
	if err := rollbackTaskExecutions(tx); err != nil {
		return err
	}
	if err := rollbackDescriptionEntities(tx); err != nil {
		return err
	}
	if err := rollbackNamedEntities(tx); err != nil {
		return err
	}
	if err := rollbackProjects(tx); err != nil {
		return err
	}
	if err := rollbackResources(tx); err != nil {
		return err
	}
	if err := rollbackSignals(tx); err != nil {
		return err
	}
	if err := rollbackSchedules(tx); err != nil {
		return err
	}
	return nil
}

// shouldApplyOrgMigrations ascertains whether models have been automigrated with or without the org column (depending on when this flyteadmin was provisioned)
// and returns whether the org column should be added.
func shouldApplyOrgMigrations(db *gorm.DB) (bool, error) {
	if db.Dialector.Name() != "postgres" {
		return false, nil
	}

	// Use tasks as a sentinel schema for whether orgs need to be added to all table schemas.
	var columnType string
	query := `
	SELECT data_type
	FROM information_schema.columns
	WHERE table_name = ? AND column_name = ?;
	`
	err := db.Raw(query, "tasks", "org").Scan(&columnType).Error
	if err != nil {
		return false, err
	}
	if columnType == "text" {
		return false, nil
	}

	return true, nil
}
