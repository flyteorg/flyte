-- Index runs' labels for label filters. Its own migration, so the ACCESS EXCLUSIVE lock
-- adding the column takes is released before the build scans the table. Partial, so child
-- actions, which never carry labels, stay out of the index.
CREATE INDEX IF NOT EXISTS idx_actions_labels ON actions USING GIN (labels jsonb_path_ops)
    WHERE labels IS NOT NULL;
