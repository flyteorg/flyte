-- Cache service initial schema

CREATE TABLE IF NOT EXISTS cache_service_outputs (
    key           VARCHAR(512) PRIMARY KEY,
    output_uri    TEXT NOT NULL DEFAULT '',
    metadata      BYTEA,
    last_updated  TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_cache_service_outputs_last_updated ON cache_service_outputs (last_updated);

CREATE TABLE IF NOT EXISTS cache_service_reservations (
    key               VARCHAR(512) PRIMARY KEY,
    owner_id          TEXT NOT NULL,
    heartbeat_seconds INTEGER NOT NULL DEFAULT 0,
    expires_at        TIMESTAMPTZ NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_cache_service_reservations_owner_id ON cache_service_reservations (owner_id);
CREATE INDEX IF NOT EXISTS idx_cache_service_reservations_expires_at ON cache_service_reservations (expires_at);
