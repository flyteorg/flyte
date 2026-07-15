-- Composite index supporting the WatchActions initial-snapshot keyset pagination,
-- which pages a run's actions ordered by (created_at, name) ascending. Without it,
-- keyset/offset paging re-scans the run's rows on every page (~O(n^2)); with it each
-- page is an index range-scan (seek to the cursor, read the next page), making the
-- full snapshot O(n) and keeping large runs (tens of thousands of actions) under the
-- client stream timeout.
CREATE INDEX IF NOT EXISTS idx_actions_run_created_name
    ON actions (project, domain, run_name, created_at, name);
