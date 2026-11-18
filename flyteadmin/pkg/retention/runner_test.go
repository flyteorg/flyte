package retention

import (
	"context"
	"database/sql"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	repositoryconfig "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func options() Options { return Options{Age: 24 * time.Hour, BatchSize: 1, MaxRuntime: time.Minute} }
func TestRetention(t *testing.T) {
	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().Port(15439).Database("flyte_retention_test").Username("postgres").Password("postgres").RuntimePath(t.TempDir() + "/postgres"))
	require.NoError(t, pg.Start())
	t.Cleanup(func() { require.NoError(t, pg.Stop()) })
	db, err := sql.Open("pgx", "host=localhost port=15439 user=postgres password=postgres dbname=flyte_retention_test sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	// Representative V1 execution columns, with foreign keys enabled to exercise
	// launched workflows, dynamic-node parent cycles, and transactional ordering.
	_, err = db.Exec(`CREATE TABLE executions(id BIGSERIAL UNIQUE, execution_project TEXT,execution_domain TEXT,execution_name TEXT,phase TEXT,execution_updated_at TIMESTAMPTZ,source_execution_id BIGINT, parent_node_execution_id BIGINT,PRIMARY KEY(execution_project,execution_domain,execution_name));
 CREATE TABLE node_executions(id BIGSERIAL UNIQUE,execution_project TEXT,execution_domain TEXT,execution_name TEXT,node_id TEXT,phase TEXT,parent_id BIGINT REFERENCES node_executions(id),parent_task_execution_id BIGINT, PRIMARY KEY(execution_project,execution_domain,execution_name,node_id));
 CREATE TABLE task_executions(id BIGSERIAL UNIQUE,execution_project TEXT,execution_domain TEXT,execution_name TEXT,node_id TEXT,phase TEXT);
 ALTER TABLE node_executions ADD FOREIGN KEY(parent_task_execution_id) REFERENCES task_executions(id);
 ALTER TABLE executions ADD FOREIGN KEY(parent_node_execution_id) REFERENCES node_executions(id)`)
	require.NoError(t, err)
	for _, table := range []string{"execution_events", "node_execution_events", "signals", "execution_tags"} {
		_, err = db.Exec(`CREATE TABLE ` + table + ` (execution_project TEXT,execution_domain TEXT,execution_name TEXT)`)
		require.NoError(t, err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)
	for _, migration := range repositoryconfig.Migrations {
		if migration.ID == "2026-10-16-execution-retention-index" {
			require.NoError(t, migration.Migrate(gormDB))
			require.NoError(t, migration.Rollback(gormDB))
			require.NoError(t, migration.Migrate(gormDB))
		}
	}
	reset := func(t *testing.T) {
		t.Helper()
		_, err := db.Exec(`TRUNCATE executions,node_executions,task_executions,execution_events,node_execution_events,signals,execution_tags RESTART IDENTITY CASCADE`)
		require.NoError(t, err)
	}
	execution := func(t *testing.T, name, phase string, at any) int64 {
		t.Helper()
		var id int64
		require.NoError(t, db.QueryRow(`INSERT INTO executions(execution_project,execution_domain,execution_name,phase,execution_updated_at) VALUES ('p','d',$1,$2,$3) RETURNING id`, name, phase, at).Scan(&id))
		return id
	}
	count := func(t *testing.T, table string) int64 {
		t.Helper()
		var n int64
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM `+table).Scan(&n))
		return n
	}
	old := time.Now().Add(-48 * time.Hour)
	t.Run("CleanupAndDryRun", func(t *testing.T) {
		reset(t)
		execution(t, "old", "SUCCEEDED", old)
		execution(t, "old2", "FAILED", old)
		execution(t, "recent", "SUCCEEDED", time.Now())
		execution(t, "active", "RUNNING", old)
		execution(t, "missing-time", "FAILED", nil)
		_, err := db.Exec(`INSERT INTO node_executions(execution_project,execution_domain,execution_name,node_id,phase) VALUES ('p','d','old','n','RECOVERED');
   INSERT INTO task_executions(execution_project,execution_domain,execution_name,node_id,phase) VALUES ('p','d','old','n','SUCCEEDED');
   UPDATE node_executions SET parent_task_execution_id=1 WHERE execution_name='old'`)
		require.NoError(t, err)
		for _, table := range []string{"execution_events", "node_execution_events", "signals", "execution_tags"} {
			_, err = db.Exec(`INSERT INTO ` + table + ` VALUES ('p','d','old')`)
			require.NoError(t, err)
		}
		metrics, reader := testMetrics(t)
		o := options()
		o.DryRun = true
		result, err := Run(t.Context(), db, o, metrics)
		require.NoError(t, err)
		require.EqualValues(t, 2, result.Deleted["executions"])
		require.EqualValues(t, 5, count(t, "executions"))
		require.Empty(t, deletedTotals(t, reader))
		o.DryRun = false
		result, err = Run(t.Context(), db, o, metrics)
		require.NoError(t, err)
		require.EqualValues(t, 2, result.Deleted["executions"])
		require.EqualValues(t, 3, count(t, "executions"))
		for _, table := range cleanupTables {
			if table != "executions" {
				require.Zero(t, count(t, table))
			}
		}
		totals := deletedTotals(t, reader)
		require.EqualValues(t, 2, totals["executions"])
		require.EqualValues(t, 1, totals["task_executions"])
		require.EqualValues(t, 1, totals["node_executions"])
		result, err = Run(t.Context(), db, o, metrics)
		require.NoError(t, err)
		require.Empty(t, result.Deleted)
	})
	t.Run("ActiveChildrenAndIdentity", func(t *testing.T) {
		reset(t)
		execution(t, "node-active", "SUCCEEDED", old)
		execution(t, "task-active", "ABORTED", old)
		execution(t, "unknown-node", "FAILED", old)
		execution(t, "same", "SUCCEEDED", old)
		_, err := db.Exec(`INSERT INTO node_executions(execution_project,execution_domain,execution_name,node_id,phase) VALUES ('p','d','node-active','n','RUNNING'),('p','d','unknown-node','n',NULL);
   INSERT INTO task_executions(execution_project,execution_domain,execution_name,node_id,phase) VALUES ('p','d','task-active','n','RUNNING');
   INSERT INTO executions(execution_project,execution_domain,execution_name,phase) VALUES ('other','d','same','RUNNING'),('p','other','same','RUNNING')`)
		require.NoError(t, err)
		metrics, _ := testMetrics(t)
		result, err := Run(t.Context(), db, options(), metrics)
		require.NoError(t, err)
		require.EqualValues(t, 1, result.Deleted["executions"])
		require.EqualValues(t, 5, count(t, "executions"))
	})
	t.Run("ReferencedExecutionsPreserved", func(t *testing.T) {
		reset(t)
		parent := execution(t, "source", "SUCCEEDED", old)
		execution(t, "dependent", "RUNNING", time.Now())
		_, err := db.Exec(`UPDATE executions SET source_execution_id=$1 WHERE execution_name='dependent'`, parent)
		require.NoError(t, err)
		metrics, _ := testMetrics(t)
		result, err := Run(t.Context(), db, options(), metrics)
		require.NoError(t, err)
		require.Empty(t, result.Deleted)
	})
	t.Run("LaunchedWorkflowBeforeParent", func(t *testing.T) {
		reset(t)
		execution(t, "parent", "SUCCEEDED", old)
		execution(t, "child", "SUCCEEDED", old.Add(time.Hour))
		var node int64
		require.NoError(t, db.QueryRow(`INSERT INTO node_executions(execution_project,execution_domain,execution_name,node_id,phase) VALUES ('p','d','parent','launch','SUCCEEDED') RETURNING id`).Scan(&node))
		_, err := db.Exec(`UPDATE executions SET parent_node_execution_id=$1 WHERE execution_name='child'`, node)
		require.NoError(t, err)
		metrics, _ := testMetrics(t)
		result, err := Run(t.Context(), db, options(), metrics)
		require.NoError(t, err)
		require.EqualValues(t, 1, result.Deleted["executions"])
		result, err = Run(t.Context(), db, options(), metrics)
		require.NoError(t, err)
		require.EqualValues(t, 1, result.Deleted["executions"])
		require.Zero(t, count(t, "node_executions"))
	})
	t.Run("CrossExecutionNodeReferencesPreserved", func(t *testing.T) {
		reset(t)
		execution(t, "parent", "SUCCEEDED", old)
		execution(t, "other", "RUNNING", time.Now())
		var parent int64
		require.NoError(t, db.QueryRow(`INSERT INTO node_executions(execution_project,execution_domain,execution_name,node_id,phase) VALUES ('p','d','parent','n','SUCCEEDED') RETURNING id`).Scan(&parent))
		_, err := db.Exec(`INSERT INTO node_executions(execution_project,execution_domain,execution_name,node_id,phase,parent_id) VALUES ('p','d','other','n','RUNNING',$1)`, parent)
		require.NoError(t, err)
		metrics, _ := testMetrics(t)
		result, err := Run(t.Context(), db, options(), metrics)
		require.NoError(t, err)
		require.Empty(t, result.Deleted)
	})

	t.Run("GlobalLock", func(t *testing.T) {
		reset(t)
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		_, err = tx.Exec(`SELECT pg_advisory_xact_lock(714239801,1)`)
		require.NoError(t, err)
		metrics, _ := testMetrics(t)
		result, err := Run(t.Context(), db, options(), metrics)
		require.NoError(t, err)
		require.True(t, result.LockUnavailable)
	})
	t.Run("PartialFailureMetrics", func(t *testing.T) {
		reset(t)
		execution(t, "first", "SUCCEEDED", old)
		execution(t, "fail", "SUCCEEDED", old.Add(time.Hour))
		_, err := db.Exec(`INSERT INTO execution_events VALUES ('p','d','fail');
   CREATE FUNCTION fail_cleanup() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF OLD.execution_name='fail' THEN RAISE EXCEPTION 'test failure'; END IF; RETURN OLD; END; $$;
   CREATE TRIGGER fail_cleanup BEFORE DELETE ON executions FOR EACH ROW EXECUTE FUNCTION fail_cleanup()`)
		require.NoError(t, err)
		defer db.Exec(`DROP TRIGGER fail_cleanup ON executions; DROP FUNCTION fail_cleanup()`)
		metrics, reader := testMetrics(t)
		result, err := Run(t.Context(), db, options(), metrics)
		require.ErrorContains(t, err, "test failure")
		require.EqualValues(t, 1, result.Deleted["executions"])
		require.EqualValues(t, 1, count(t, "execution_events"))
		require.EqualValues(t, 1, deletedTotals(t, reader)["executions"])
	})
	t.Run("CutoffAndRecheck", func(t *testing.T) {
		reset(t)
		cutoff := time.Now().UTC().Truncate(time.Microsecond)
		execution(t, "boundary", "SUCCEEDED", cutoff)
		counts, err := processExecution(t.Context(), db, executionKey{project: "p", domain: "d", name: "boundary"}, cutoff, false)
		require.NoError(t, err)
		require.Empty(t, counts)
		_, err = db.Exec(`UPDATE executions SET phase='RUNNING'`)
		require.NoError(t, err)
		counts, err = processExecution(t.Context(), db, executionKey{project: "p", domain: "d", name: "boundary"}, cutoff.Add(time.Hour), false)
		require.NoError(t, err)
		require.Empty(t, counts)
	})
	t.Run("ConcurrentStateUpdateRechecksEligibility", func(t *testing.T) {
		reset(t)
		execution(t, "race", "SUCCEEDED", old)
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		_, err = tx.Exec(`UPDATE executions SET phase='RUNNING' WHERE execution_name='race'`)
		require.NoError(t, err)
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		type outcome struct {
			counts map[string]int64
			err    error
		}
		done := make(chan outcome, 1)
		go func() {
			counts, err := processExecution(ctx, db, executionKey{project: "p", domain: "d", name: "race"}, time.Now().Add(-24*time.Hour), false)
			done <- outcome{counts, err}
		}()
		require.Eventually(t, func() bool {
			var n int
			err := db.QueryRow(`SELECT count(*) FROM pg_stat_activity WHERE wait_event_type='Lock' AND datname=current_database()`).Scan(&n)
			return err == nil && n > 0
		}, 3*time.Second, 10*time.Millisecond)
		require.NoError(t, tx.Commit())
		result := <-done
		require.NoError(t, result.err)
		require.Empty(t, result.counts)
		require.EqualValues(t, 1, count(t, "executions"))
	})

	t.Run("RuntimeLimitRollsBack", func(t *testing.T) {
		reset(t)
		execution(t, "blocked", "SUCCEEDED", old)
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		_, err = tx.Exec(`SELECT id FROM executions FOR UPDATE`)
		require.NoError(t, err)
		metrics, _ := testMetrics(t)
		o := options()
		o.MaxRuntime = 100 * time.Millisecond
		result, err := Run(context.Background(), db, o, metrics)
		require.Error(t, err)
		require.Empty(t, result.Deleted)
		require.EqualValues(t, 1, count(t, "executions"))
	})
}
