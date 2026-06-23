-- Add created_by to actions: the OIDC subject of the run's creator, captured from the
-- auth headers the load balancer forwards. An indexed, queryable scalar for filtering and
-- listing runs by owner — complements executed_by, which holds the full serialized
-- EnrichedIdentity (subject plus name/email) for display. NULL for runs created without
-- an authenticated identity.
ALTER TABLE actions ADD COLUMN IF NOT EXISTS created_by TEXT;
CREATE INDEX IF NOT EXISTS idx_actions_created_by ON actions (created_by);
