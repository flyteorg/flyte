package retention

import (
	"context"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/flyteorg/flyte/v2/runs/migrations"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func options() Options {
	return Options{Age: 24 * time.Hour, BatchSize: 1, MaxRuntime: time.Minute}
}

func TestRetention(t *testing.T) {
	pg := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(15439).
			Database("flyte_retention_test").
			Username("postgres").
			Password("postgres").
			RuntimePath(t.TempDir() + "/postgres"),
	)
	require.NoError(t, pg.Start())
	t.Cleanup(func() { require.NoError(t, pg.Stop()) })

	testDB, err := sqlx.ConnectContext(t.Context(), "pgx",
		"host=localhost port=15439 user=postgres password=postgres dbname=flyte_retention_test sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testDB.Close()) })
	require.NoError(t, migrations.RunMigrations(t.Context(), testDB))
	defaultMetrics, _ := testMetrics(t)

	// Subtests share this database and run sequentially, resetting their data.
	reset := func(t *testing.T) {
		t.Helper()
		_, err := testDB.Exec(`TRUNCATE actions,action_events`)
		require.NoError(t, err)
	}

	action := func(t *testing.T, run, name string, parent any, phase int, end any, abort any) {
		t.Helper()
		_, err := testDB.Exec(`INSERT INTO actions (project,domain,run_name,name,parent_action_name,phase,ended_at,abort_requested_at) VALUES ('p','d',$1,$2,$3,$4,$5,$6)`, run, name, parent, phase, end, abort)
		require.NoError(t, err)
	}

	count := func(t *testing.T, table string) int {
		t.Helper()
		var n int
		require.NoError(t, testDB.Get(&n, `SELECT count(*) FROM `+table))
		return n
	}

	t.Run("Cleanup", func(t *testing.T) {
		reset(t)
		metrics, reader := testMetrics(t)
		old := time.Now().Add(-48 * time.Hour)
		action(t, "expired", "expired", nil, 5, old, nil)
		action(t, "expired", "child", "expired", 10, old, nil)
		action(t, "expired2", "expired2", nil, 6, old, nil)
		action(t, "recent", "recent", nil, 5, time.Now(), nil)
		action(t, "active", "active", nil, 4, old, nil)
		action(t, "paused", "paused", nil, 9, old, nil)
		action(t, "unknown", "unknown", nil, 0, old, nil)
		action(t, "noend", "noend", nil, 5, nil, nil)
		action(t, "childactive", "childactive", nil, 5, old, nil)
		action(t, "childactive", "child", "childactive", 4, nil, nil)
		action(t, "aborting", "aborting", nil, 7, old, time.Now())
		_, err := testDB.Exec(`INSERT INTO action_events(project,domain,run_name,name,attempt,phase,version) VALUES ('p','d','expired','child',0,5,0)`)
		require.NoError(t, err)
		o := options()
		o.DryRun = true
		result, err := Run(context.Background(), testDB, o, metrics)
		require.NoError(t, err)
		require.EqualValues(t, 2, result.Runs)
		require.EqualValues(t, 3, result.Actions)
		require.EqualValues(t, 1, result.Events)
		require.Equal(t, 11, count(t, "actions"))
		require.Empty(t, deletedTotals(t, reader))
		o.DryRun = false
		result, err = Run(context.Background(), testDB, o, metrics)
		require.NoError(t, err)
		require.EqualValues(t, 2, result.Runs)
		require.Equal(t, 8, count(t, "actions"))
		require.Equal(t, 0, count(t, "action_events"))
		result, err = Run(context.Background(), testDB, o, metrics)
		require.NoError(t, err)
		require.Zero(t, result.Runs)
		require.Equal(t, map[string]int64{"run": 2, "action": 3, "action_event": 1}, deletedTotals(t, reader))

	})

	t.Run("CutoffAndRecheck", func(t *testing.T) {
		reset(t)
		cutoff := time.Now().UTC().Truncate(time.Microsecond)
		action(t, "boundary", "boundary", nil, 5, cutoff, nil)
		key := runKey{Project: "p", Domain: "d", RunName: "boundary"}
		result, err := processRun(context.Background(), testDB, key, cutoff, false)
		require.NoError(t, err)
		require.Zero(t, result.Runs)
		_, err = testDB.Exec(`UPDATE actions SET phase=4 WHERE run_name='boundary'`)
		require.NoError(t, err)
		result, err = processRun(context.Background(), testDB, key, cutoff.Add(time.Hour), false)
		require.NoError(t, err)
		require.Zero(t, result.Runs)
	})

	t.Run("Rollback", func(t *testing.T) {
		reset(t)
		metrics, reader := testMetrics(t)
		old := time.Now().Add(-48 * time.Hour)
		action(t, "rollback", "rollback", nil, 5, old, nil)
		_, err := testDB.Exec(`INSERT INTO action_events(project,domain,run_name,name,attempt,phase,version) VALUES ('p','d','rollback','rollback',0,5,0)`)
		require.NoError(t, err)
		_, err = testDB.Exec(`CREATE FUNCTION fail_retention_delete() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'test delete failure'; END; $$;
	 CREATE TRIGGER fail_retention_delete BEFORE DELETE ON actions FOR EACH ROW EXECUTE FUNCTION fail_retention_delete()`)
		require.NoError(t, err)
		defer testDB.Exec(`DROP TRIGGER fail_retention_delete ON actions; DROP FUNCTION fail_retention_delete()`)
		_, err = Run(context.Background(), testDB, options(), metrics)
		require.ErrorContains(t, err, "test delete failure")
		require.Equal(t, 1, count(t, "actions"))
		require.Equal(t, 1, count(t, "action_events"))
		require.Empty(t, deletedTotals(t, reader))
	})

	t.Run("MetricsPreserveCommittedProgressOnFailure", func(t *testing.T) {
		reset(t)
		metrics, reader := testMetrics(t)
		action(t, "committed", "committed", nil, 5, time.Now().Add(-49*time.Hour), nil)
		action(t, "failed", "failed", nil, 5, time.Now().Add(-48*time.Hour), nil)
		_, err := testDB.Exec(`INSERT INTO action_events(project,domain,run_name,name,attempt,phase,version) VALUES ('p','d','committed','committed',0,5,0),('p','d','failed','failed',0,5,0)`)
		require.NoError(t, err)
		_, err = testDB.Exec(`CREATE FUNCTION fail_partial_retention_delete() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF OLD.run_name='failed' THEN RAISE EXCEPTION 'partial cleanup failure'; END IF; RETURN OLD; END; $$;
        CREATE TRIGGER fail_partial_retention_delete BEFORE DELETE ON actions FOR EACH ROW EXECUTE FUNCTION fail_partial_retention_delete()`)
		require.NoError(t, err)
		defer testDB.Exec(`DROP TRIGGER fail_partial_retention_delete ON actions; DROP FUNCTION fail_partial_retention_delete()`)
		result, err := Run(context.Background(), testDB, options(), metrics)
		require.ErrorContains(t, err, "partial cleanup failure")
		require.EqualValues(t, 1, result.Runs)
		require.Equal(t, map[string]int64{"run": 1, "action": 1, "action_event": 1}, deletedTotals(t, reader))
		require.Equal(t, 1, count(t, "actions"))
		require.Equal(t, 1, count(t, "action_events"))
	})

	t.Run("GlobalLock", func(t *testing.T) {
		reset(t)
		tx, err := testDB.Beginx()
		require.NoError(t, err)
		defer tx.Rollback()
		_, err = tx.Exec(`SELECT pg_advisory_xact_lock(714239801,1)`)
		require.NoError(t, err)
		result, err := Run(context.Background(), testDB, options(), defaultMetrics)
		require.NoError(t, err)
		require.True(t, result.LockUnavailable)
	})

	t.Run("ConcurrentStateUpdateRechecksEligibility", func(t *testing.T) {
		reset(t)
		old := time.Now().Add(-48 * time.Hour)
		action(t, "race", "race", nil, 5, old, nil)
		tx, err := testDB.Beginx()
		require.NoError(t, err)
		defer tx.Rollback()
		_, err = tx.Exec(`UPDATE actions SET phase=4 WHERE project='p' AND domain='d' AND run_name='race'`)
		require.NoError(t, err)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		type outcome struct {
			r Result
			e error
		}
		done := make(chan outcome, 1)
		go func() {
			r, e := processRun(ctx, testDB, runKey{Project: "p", Domain: "d", RunName: "race"}, time.Now().Add(-24*time.Hour), false)
			done <- outcome{r, e}
		}()
		require.Eventually(t, func() bool {
			var n int
			e := testDB.Get(&n, `SELECT count(*) FROM pg_stat_activity WHERE wait_event_type='Lock' AND datname=current_database()`)
			return e == nil && n > 0
		}, 3*time.Second, 10*time.Millisecond)
		require.NoError(t, tx.Commit())
		out := <-done
		require.NoError(t, out.e)
		require.Zero(t, out.r.Runs)
		require.Equal(t, 1, count(t, "actions"))
	})

	t.Run("RunIdentityIsolation", func(t *testing.T) {
		reset(t)
		old := time.Now().Add(-48 * time.Hour)
		action(t, "same", "same", nil, 5, old, nil)
		_, err := testDB.Exec(`INSERT INTO actions(project,domain,run_name,name,phase) VALUES ('other','d','same','same',4),('p','other','same','same',4)`)
		require.NoError(t, err)
		_, err = testDB.Exec(`INSERT INTO task_specs(digest,spec) VALUES ('retention-test-spec','\\x01') ON CONFLICT DO NOTHING`)
		require.NoError(t, err)
		defer testDB.Exec(`DELETE FROM task_specs WHERE digest='retention-test-spec'`)
		result, err := Run(context.Background(), testDB, options(), defaultMetrics)
		require.NoError(t, err)
		require.EqualValues(t, 1, result.Runs)
		require.Equal(t, 2, count(t, "actions"))
		var specs int
		require.NoError(t, testDB.Get(&specs, `SELECT count(*) FROM task_specs WHERE digest='retention-test-spec'`))
		require.Equal(t, 1, specs)
	})

	t.Run("CancellationRollsBackBlockedRun", func(t *testing.T) {
		reset(t)
		old := time.Now().Add(-48 * time.Hour)
		action(t, "blocked", "blocked", nil, 5, old, nil)
		tx, err := testDB.Beginx()
		require.NoError(t, err)
		defer tx.Rollback()
		_, err = tx.Exec(`SELECT name FROM actions WHERE project='p' AND domain='d' AND run_name='blocked' FOR UPDATE`)
		require.NoError(t, err)
		o := options()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		result, err := Run(ctx, testDB, o, defaultMetrics)
		require.Error(t, err)
		require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
		require.Zero(t, result.Runs)
		require.Equal(t, 1, count(t, "actions"))
	})

	t.Run("RuntimeLimitReturnsError", func(t *testing.T) {
		reset(t)
		action(t, "timeout", "timeout", nil, 5, time.Now().Add(-48*time.Hour), nil)
		tx, err := testDB.Beginx()
		require.NoError(t, err)
		defer tx.Rollback()
		_, err = tx.Exec(`SELECT name FROM actions WHERE project='p' AND domain='d' AND run_name='timeout' FOR UPDATE`)
		require.NoError(t, err)
		o := options()
		o.MaxRuntime = 100 * time.Millisecond
		result, err := Run(context.Background(), testDB, o, defaultMetrics)
		require.Error(t, err)
		require.Zero(t, result.Runs)
		require.Equal(t, 1, count(t, "actions"))
	})
}
