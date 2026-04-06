-- Runs service initial schema

CREATE TABLE IF NOT EXISTS projects (
    identifier  TEXT PRIMARY KEY,
    name        TEXT NOT NULL DEFAULT '',
    description VARCHAR(300) NOT NULL DEFAULT '',
    labels      BYTEA,
    state       INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_projects_state ON projects (state);

CREATE TABLE IF NOT EXISTS actions (
    id                  BIGSERIAL PRIMARY KEY,
    project             TEXT NOT NULL,
    domain              TEXT NOT NULL,
    run_name            TEXT NOT NULL DEFAULT '',
    name                TEXT NOT NULL,
    parent_action_name  TEXT,
    phase               INTEGER NOT NULL DEFAULT 1,
    run_source          TEXT NOT NULL DEFAULT '',
    action_type         INTEGER NOT NULL DEFAULT 0,
    action_group        TEXT,
    task_project        TEXT,
    task_domain         TEXT,
    task_name           TEXT,
    task_version        TEXT,
    task_type           TEXT NOT NULL DEFAULT '',
    task_short_name     TEXT,
    function_name       TEXT NOT NULL DEFAULT '',
    environment_name    TEXT,
    action_spec         BYTEA,
    action_details      BYTEA,
    detailed_info       BYTEA,
    run_spec            BYTEA,
    abort_requested_at  TIMESTAMPTZ,
    abort_attempt_count INTEGER NOT NULL DEFAULT 0,
    abort_reason        TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ,
    ended_at            TIMESTAMPTZ,
    duration_ms         BIGINT,
    attempts            INTEGER NOT NULL DEFAULT 0,
    cache_status        INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_actions_identifier ON actions (project, domain, run_name, name);
CREATE INDEX IF NOT EXISTS idx_actions_run_lookup ON actions (project, domain, run_name);
CREATE INDEX IF NOT EXISTS idx_actions_parent ON actions (parent_action_name);
CREATE INDEX IF NOT EXISTS idx_actions_phase ON actions (phase);
CREATE INDEX IF NOT EXISTS idx_actions_created ON actions (created_at);
CREATE INDEX IF NOT EXISTS idx_actions_abort_pending ON actions (abort_requested_at);

CREATE TABLE IF NOT EXISTS action_events (
    project     TEXT NOT NULL,
    domain      TEXT NOT NULL,
    run_name    TEXT NOT NULL,
    name        TEXT NOT NULL,
    attempt     INTEGER NOT NULL,
    phase       INTEGER NOT NULL,
    version     INTEGER NOT NULL,
    info        BYTEA,
    error_kind  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ,
    PRIMARY KEY (project, domain, run_name, name, attempt, phase, version)
);
CREATE INDEX IF NOT EXISTS idx_action_events_error_kind ON action_events (error_kind);

CREATE TABLE IF NOT EXISTS tasks (
    project                 TEXT NOT NULL,
    domain                  TEXT NOT NULL,
    name                    TEXT NOT NULL,
    version                 TEXT NOT NULL,
    environment             TEXT NOT NULL DEFAULT '',
    function_name           TEXT NOT NULL DEFAULT '',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ,
    deployed_by             TEXT NOT NULL DEFAULT '',
    trigger_name            TEXT,
    total_triggers          INTEGER NOT NULL DEFAULT 0,
    active_triggers         INTEGER NOT NULL DEFAULT 0,
    trigger_automation_spec BYTEA,
    trigger_types           BIT(10),
    task_spec               BYTEA,
    env_description         TEXT,
    short_description       TEXT,
    PRIMARY KEY (project, domain, name, version)
);
CREATE INDEX IF NOT EXISTS idx_tasks_identifier ON tasks (project, domain, name, version);

CREATE TABLE IF NOT EXISTS task_specs (
    digest      TEXT PRIMARY KEY,
    spec        BYTEA NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMPTZ
);
