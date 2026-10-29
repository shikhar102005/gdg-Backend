-- Migration: 000002_create_readings.up.sql
CREATE TABLE IF NOT EXISTS readings (
    id BIGSERIAL PRIMARY KEY,
    sensor_id VARCHAR(64) NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
    observed_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metric_type VARCHAR(32) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    quality VARCHAR(32) NOT NULL DEFAULT 'good',
    idempotency_key VARCHAR(128),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_readings_idempotency 
    ON readings(sensor_id, idempotency_key) 
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_readings_sensor_time 
    ON readings(sensor_id, observed_at DESC);

CREATE INDEX IF NOT EXISTS idx_readings_observed_at 
    ON readings(observed_at DESC);
