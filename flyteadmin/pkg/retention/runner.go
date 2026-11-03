// Package retention deletes expired workflow execution history from PostgreSQL.
package retention

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type Options struct {
	Age        time.Duration
	BatchSize  int
	MaxRuntime time.Duration
	DryRun     bool
}

func (o Options) Validate() error {
	if o.Age <= 0 || o.BatchSize <= 0 || o.MaxRuntime <= 0 {
		return errors.New("retention age, batch size and maximum runtime must be positive")
	}
	return nil
}

type Result struct {
	Cutoff          time.Time        `json:"cutoff"`
	Deleted         map[string]int64 `json:"deleted"`
	DryRun          bool             `json:"dryRun"`
	LockUnavailable bool             `json:"lockUnavailable"`
}

type executionKey struct {
	project, domain, name string
	updatedAt             time.Time
}

const keyWhere = `execution_project=$1 AND execution_domain=$2 AND execution_name=$3`
const matchExecution = `n.execution_project=e.execution_project AND n.execution_domain=e.execution_domain AND n.execution_name=e.execution_name`

// V1 records completion in execution_updated_at, rather than an ended_at column.
// Protect executions still needed by recovery/relaunches or launched workflows.
const eligible = `e.phase IN ('SUCCEEDED','FAILED','ABORTED','TIMED_OUT') AND e.execution_updated_at < $1
 AND NOT EXISTS (SELECT 1 FROM node_executions n WHERE ` + matchExecution + ` AND (n.phase IS NULL OR n.phase NOT IN ('SUCCEEDED','FAILED','ABORTED','TIMED_OUT','SKIPPED','RECOVERED')))
 AND NOT EXISTS (SELECT 1 FROM task_executions n WHERE ` + matchExecution + ` AND (n.phase IS NULL OR n.phase NOT IN ('SUCCEEDED','FAILED','ABORTED')))
 AND NOT EXISTS (SELECT 1 FROM executions dependent WHERE dependent.source_execution_id=e.id)
 AND NOT EXISTS (SELECT 1 FROM executions dependent WHERE dependent.parent_node_execution_id IN (SELECT n.id FROM node_executions n WHERE ` + matchExecution + `))
 AND NOT EXISTS (SELECT 1 FROM node_executions child WHERE
   (child.execution_project,child.execution_domain,child.execution_name) <> (e.execution_project,e.execution_domain,e.execution_name)
   AND (child.parent_id IN (SELECT n.id FROM node_executions n WHERE ` + matchExecution + `)
     OR child.parent_task_execution_id IN (SELECT n.id FROM task_executions n WHERE ` + matchExecution + `)))`

// Run deletes each execution in its own transaction and records committed cleanup.
// The caller records invocation duration/result to include connection setup.
// It requires an already migrated PostgreSQL database and never performs DDL.
func Run(ctx context.Context, db *sql.DB, o Options, metrics *Metrics) (Result, error) {
	result := Result{Cutoff: time.Now().UTC().Add(-o.Age), Deleted: map[string]int64{}, DryRun: o.DryRun}
	if err := o.Validate(); err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, o.MaxRuntime)
	defer cancel()
	lock, err := db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer lock.Rollback()
	var locked bool
	if err = lock.QueryRowContext(ctx, `SELECT pg_try_advisory_xact_lock(714239801,1)`).Scan(&locked); err != nil {
		return result, err
	}
	if !locked {
		result.LockUnavailable = true
		return result, nil
	}
	var cursor *executionKey
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		query := `SELECT e.execution_project,e.execution_domain,e.execution_name,e.execution_updated_at FROM executions e WHERE ` + eligible
		args := []any{result.Cutoff}
		if cursor != nil {
			query += ` AND (e.execution_updated_at,e.execution_project,e.execution_domain,e.execution_name)>($2,$3,$4,$5)`
			args = append(args, cursor.updatedAt, cursor.project, cursor.domain, cursor.name)
		}
		query += fmt.Sprintf(` ORDER BY e.execution_updated_at,e.execution_project,e.execution_domain,e.execution_name LIMIT $%d`, len(args)+1)
		args = append(args, o.BatchSize)
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return result, err
		}
		var keys []executionKey
		for rows.Next() {
			var key executionKey
			if err = rows.Scan(&key.project, &key.domain, &key.name, &key.updatedAt); err != nil {
				rows.Close()
				return result, err
			}
			keys = append(keys, key)
		}
		scanErr := rows.Err()
		closeErr := rows.Close()
		if scanErr != nil {
			return result, scanErr
		}
		if closeErr != nil {
			return result, closeErr
		}
		if len(keys) == 0 {
			return result, nil
		}
		for _, key := range keys {
			counts, err := processExecution(ctx, db, key, result.Cutoff, o.DryRun)
			if err != nil {
				return result, fmt.Errorf("purge %s/%s/%s: %w", key.project, key.domain, key.name, err)
			}
			for table, n := range counts {
				result.Deleted[table] += n
			}
			metrics.recordCleanup(ctx, counts, o.DryRun)
		}
		cursor = &keys[len(keys)-1]
		logger.Infof(ctx, "retention progress: dry_run=%t cutoff=%s counts=%v", o.DryRun, result.Cutoff.Format(time.RFC3339), result.Deleted)
	}
}

// Delete child history before its owning execution. Node/task parent pointers
// can form cycles (dynamic workflows), so clear only pointers in this execution
// within the same transaction before deleting the task and node rows.
var cleanupTables = []string{"execution_events", "node_execution_events", "signals", "execution_tags", "task_executions", "node_executions", "executions"}

func processExecution(ctx context.Context, db *sql.DB, key executionKey, cutoff time.Time, dry bool) (map[string]int64, error) {
	counts := map[string]int64{}
	isolation := sql.LevelReadCommitted
	if dry {
		isolation = sql.LevelRepeatableRead
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: isolation, ReadOnly: dry})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	args := []any{key.project, key.domain, key.name}
	if !dry {
		for _, table := range []string{"executions", "node_executions", "task_executions"} {
			rows, err := tx.QueryContext(ctx, `SELECT id FROM `+table+` WHERE `+keyWhere+` ORDER BY id FOR UPDATE`, args...)
			if err != nil {
				return nil, err
			}
			for rows.Next() {
			}
			scanErr := rows.Err()
			closeErr := rows.Close()
			if scanErr != nil {
				return nil, scanErr
			}
			if closeErr != nil {
				return nil, closeErr
			}
		}
	}
	var ok bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM executions e WHERE `+eligible+` AND e.execution_project=$2 AND e.execution_domain=$3 AND e.execution_name=$4)`, cutoff, key.project, key.domain, key.name).Scan(&ok); err != nil {
		return nil, err
	}
	if !ok {
		return counts, nil
	}
	if !dry {
		if _, err = tx.ExecContext(ctx, `UPDATE node_executions SET parent_id=NULL,parent_task_execution_id=NULL WHERE `+keyWhere, args...); err != nil {
			return nil, err
		}
	}
	for _, table := range cleanupTables {
		if dry {
			var n int64
			if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE `+keyWhere, args...).Scan(&n); err != nil {
				return nil, err
			}
			counts[table] = n
		} else {
			deleted, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE `+keyWhere, args...)
			if err != nil {
				return nil, err
			}
			n, err := deleted.RowsAffected()
			if err != nil {
				return nil, err
			}
			counts[table] = n
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return counts, nil
}
