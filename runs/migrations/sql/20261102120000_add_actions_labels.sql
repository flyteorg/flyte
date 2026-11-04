-- Add labels to actions: the run's labels (RunSpec.labels) on its root action, as a JSONB
-- object so runs can be filtered by label. The run_spec column holds the same labels but
-- serialized, so it cannot be queried. NULL on child actions and on existing runs.
ALTER TABLE actions ADD COLUMN IF NOT EXISTS labels JSONB;
