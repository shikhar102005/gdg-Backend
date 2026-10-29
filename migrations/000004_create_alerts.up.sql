-- Migration: 000004_create_alerts.up.sql
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    sensor_id VARCHAR(64) NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
    deduplication_key VARCHAR(256) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'OPEN', -- 'OPEN', 'ACKNOWLEDGED', 'RESOLVED'
    trigger_value DOUBLE PRECISION NOT NULL,
    trigger_time TIMESTAMPTZ NOT NULL,
    window_start TIMESTAMPTZ,
    window_end TIMESTAMPTZ,
    acknowledged_by VARCHAR(128),
    acknowledged_at TIMESTAMPTZ,
    resolved_by VARCHAR(128),
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique index ensuring only one active (OPEN or ACKNOWLEDGED) alert per deduplication key
CREATE UNIQUE INDEX IF NOT EXISTS idx_alerts_active_dedup 
    ON alerts(deduplication_key) 
    WHERE status IN ('OPEN', 'ACKNOWLEDGED');

CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_sensor ON alerts(sensor_id);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);
