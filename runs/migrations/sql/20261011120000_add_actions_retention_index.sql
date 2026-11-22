-- Support selection of completed root actions by retention cutoff.
CREATE INDEX IF NOT EXISTS idx_actions_retention
    ON actions (ended_at, project, domain, run_name)
    WHERE parent_action_name IS NULL AND phase IN (5, 6, 7, 8, 10);
