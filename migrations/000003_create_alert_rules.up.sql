-- Migration: 000003_create_alert_rules.up.sql
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    scope_type VARCHAR(32) NOT NULL DEFAULT 'global', -- 'global', 'sensor', 'type', 'building', 'campus'
    scope_value VARCHAR(64) NOT NULL DEFAULT '',
    metric VARCHAR(32) NOT NULL,
    rule_class VARCHAR(32) NOT NULL DEFAULT 'single_reading', -- 'single_reading', 'time_window'
    operator VARCHAR(8) NOT NULL, -- '>', '>=', '<', '<=', '==', '!='
    threshold DOUBLE PRECISION NOT NULL,
    window_duration_seconds INTEGER NOT NULL DEFAULT 600, -- default 10 minutes
    window_aggregation VARCHAR(16) NOT NULL DEFAULT 'avg', -- 'avg', 'min', 'max', 'count'
    severity VARCHAR(16) NOT NULL DEFAULT 'warning', -- 'info', 'warning', 'critical'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_rules_metric ON alert_rules(metric);
