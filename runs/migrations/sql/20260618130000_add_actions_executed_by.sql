-- Add executed_by to actions: the serialized common.EnrichedIdentity of the run's
-- creator, captured from the OIDC claims the load balancer forwards (subject, plus
-- name/email when present). Surfaced as ActionMetadata.executed_by.
ALTER TABLE actions ADD COLUMN IF NOT EXISTS executed_by BYTEA;
