-- Store the run creator's subject (the OIDC `sub` the load balancer forwards) as a
-- queryable, indexed column. Display identity (name/email) is resolved from the subject
-- at read time by the deployment's identity service, so it never goes stale. Surfaced as
-- ActionMetadata.executed_by. NULL for runs created without an authenticated identity.
ALTER TABLE actions ADD COLUMN IF NOT EXISTS created_by TEXT;
CREATE INDEX IF NOT EXISTS idx_actions_created_by ON actions (created_by);
