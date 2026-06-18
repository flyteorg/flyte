-- Add created_by to actions: the OIDC subject of the identity that created the run.
-- Captured from the auth headers the load balancer forwards (it enforces auth),
-- and used to populate ActionMetadata.executed_by on read.
-- TEXT, not VARCHAR(n): the OIDC `sub` length is IdP-dependent and can exceed 255.
ALTER TABLE actions ADD COLUMN IF NOT EXISTS created_by TEXT;
