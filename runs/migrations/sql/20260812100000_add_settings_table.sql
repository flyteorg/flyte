-- Settings: one row per settings scope (instance, domain, or project).
-- data holds the Settings proto as protojson; version implements
-- optimistic locking. See the Settings Service RFC (#7775).

CREATE TABLE IF NOT EXISTS settings (
    id         BIGSERIAL   PRIMARY KEY,
    key        TEXT        NOT NULL UNIQUE,
    data       JSONB       NOT NULL DEFAULT '{}',
    version    BIGINT      NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
