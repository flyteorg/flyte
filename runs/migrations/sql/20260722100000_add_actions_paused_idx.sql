-- Partial index supporting the "has paused action" run filter (ListRuns).
-- phase 9 = ACTION_PHASE_PAUSED (e.g. human-in-the-loop gate node awaiting input).
-- Paused actions are rare, so this partial index stays small and turns the
-- correlated EXISTS subquery into an index lookup keyed on the run's PK prefix.
CREATE INDEX IF NOT EXISTS idx_actions_paused
    ON actions (project, domain, run_name)
    WHERE phase = 9;
