-- Migration: 000001_create_sensors.up.sql
CREATE TABLE IF NOT EXISTS sensors (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    campus VARCHAR(64) NOT NULL DEFAULT 'Main Campus',
    building VARCHAR(64) NOT NULL,
    location VARCHAR(128) NOT NULL,
    sensor_type VARCHAR(32) NOT NULL,
    unit VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    api_key_hash VARCHAR(128) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sensors_building ON sensors(building);
CREATE INDEX IF NOT EXISTS idx_sensors_type ON sensors(sensor_type);
CREATE INDEX IF NOT EXISTS idx_sensors_status ON sensors(status);
