-- App conditions history table.
-- Stores the append-only condition log for each app.
-- KService CRD remains the source of truth for live status (spec, replicas, ingress).
-- This table only persists the conditions array for historical display.

CREATE TABLE IF NOT EXISTS app_conditions (
    project    TEXT NOT NULL,
    domain     TEXT NOT NULL,
    name       TEXT NOT NULL,
    conditions BYTEA NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project, domain, name)
);
