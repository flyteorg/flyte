// Package retention implements bounded, transactional cleanup of completed runs.
package retention

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/jmoiron/sqlx"
)

// Options are fixed for one invocation. Age is measured from root ended_at.
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

// Result counts committed deletes, or candidates and associated rows in dry-run.
type Result struct {
	Cutoff          time.Time `json:"cutoff"`
	Runs            int64     `json:"runs"`
	Actions         int64     `json:"actions"`
	Events          int64     `json:"events"`
	DryRun          bool      `json:"dryRun"`
	LockUnavailable bool      `json:"lockUnavailable"`
}

type runKey struct {
	Project string    `db:"project"`
	Domain  string    `db:"domain"`
	RunName string    `db:"run_name"`
	EndedAt time.Time `db:"ended_at"`
}

// Explicit enum list includes RECOVERED, but excludes PAUSED and unknown phases.
const eligible = `r.parent_action_name IS NULL AND r.phase IN (5,6,7,8,10)
 AND r.ended_at < $1
 AND NOT EXISTS (SELECT 1 FROM actions c
   WHERE c.project=r.project AND c.domain=r.domain AND c.run_name=r.run_name
   AND (c.phase NOT IN (5,6,7,8,10) OR c.abort_requested_at IS NOT NULL))`
const runWhere = `project=$1 AND domain=$2 AND run_name=$3`

// Run requires migrations to have been applied by the runs service. It never
// starts service workers or performs DDL, including during dry-run.
func Run(ctx context.Context, db *sqlx.DB, o Options) (Result, error) {
	result := Result{Cutoff: time.Now().UTC().Add(-o.Age), DryRun: o.DryRun}
	if err := o.Validate(); err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, o.MaxRuntime)
	defer cancel()
	// Hold the binary's single advisory lock until the job exits. A transaction
	// lock releases automatically on rollback or connection loss.
	lockTx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer lockTx.Rollback()
	var locked bool
	if err = lockTx.GetContext(ctx, &locked, `SELECT pg_try_advisory_xact_lock(714239801, 1)`); err != nil {
		return result, err
	}
	if !locked {
		result.LockUnavailable = true
		return result, nil
	}

	var cursor *runKey
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		query := `SELECT r.project,r.domain,r.run_name,r.ended_at FROM actions r WHERE ` + eligible
		args := []any{result.Cutoff}
		if cursor != nil {
			query += ` AND (r.ended_at,r.project,r.domain,r.run_name) > ($2,$3,$4,$5)`
			args = append(args, cursor.EndedAt, cursor.Project, cursor.Domain, cursor.RunName)
		}
		query += fmt.Sprintf(` ORDER BY r.ended_at,r.project,r.domain,r.run_name LIMIT $%d`, len(args)+1)
		args = append(args, o.BatchSize)
		var keys []runKey
		if err := db.SelectContext(ctx, &keys, query, args...); err != nil {
			return result, err
		}
		if len(keys) == 0 {
			return result, nil
		}
		for _, key := range keys {
			counts, err := processRun(ctx, db, key, result.Cutoff, o.DryRun)
			if err != nil {
				return result, fmt.Errorf("purge %s/%s/%s: %w", key.Project, key.Domain, key.RunName, err)
			}
			result.Runs += counts.Runs
			result.Actions += counts.Actions
			result.Events += counts.Events
		}
		cursor = &keys[len(keys)-1]
		logger.Infof(ctx, "retention progress: dry_run=%t runs=%d actions=%d events=%d cutoff=%s", o.DryRun, result.Runs, result.Actions, result.Events, result.Cutoff.Format(time.RFC3339))
	}
}

func processRun(ctx context.Context, db *sqlx.DB, key runKey, cutoff time.Time, dry bool) (Result, error) {
	result := Result{}
	isolation := sql.LevelReadCommitted
	if dry {
		isolation = sql.LevelRepeatableRead
	}
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: isolation, ReadOnly: dry})
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	args := []any{key.Project, key.Domain, key.RunName}
	if !dry {
		// Lock existing rows before checking phases so a
		// concurrent state update cannot make an eligible run active during deletion.
		rows, err := tx.QueryxContext(ctx, `SELECT name FROM actions WHERE `+runWhere+` ORDER BY name FOR UPDATE`, args...)
		if err != nil {
			return result, err
		}
		// Drain without accumulating an arbitrarily large run in memory.
		for rows.Next() {
		}
		scanErr := rows.Err()
		closeErr := rows.Close()
		if scanErr != nil {
			return result, scanErr
		}
		if closeErr != nil {
			return result, closeErr
		}
	}
	var ok bool
	if err = tx.GetContext(ctx, &ok, `SELECT EXISTS(SELECT 1 FROM actions r WHERE `+eligible+` AND r.project=$2 AND r.domain=$3 AND r.run_name=$4)`, cutoff, key.Project, key.Domain, key.RunName); err != nil {
		return result, err
	}
	if !ok {
		return result, nil
	}
	if dry {
		if err = tx.GetContext(ctx, &result.Actions, `SELECT count(*) FROM actions WHERE `+runWhere, args...); err != nil {
			return result, err
		}
		if err = tx.GetContext(ctx, &result.Events, `SELECT count(*) FROM action_events WHERE `+runWhere, args...); err != nil {
			return result, err
		}
	} else {
		for _, table := range []string{"action_events", "actions"} {
			deleted, e := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE `+runWhere, args...)
			if e != nil {
				return result, e
			}
			n, e := deleted.RowsAffected()
			if e != nil {
				return result, e
			}
			if table == "actions" {
				result.Actions = n
			} else {
				result.Events = n
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return Result{}, err
	}
	result.Runs = 1
	return result, nil
}
