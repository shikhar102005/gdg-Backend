-- Migration: 000005_create_audit_logs.up.sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(32) NOT NULL, -- 'alert', 'rule', 'sensor', 'simulator'
    entity_id VARCHAR(64) NOT NULL,
    action VARCHAR(32) NOT NULL, -- 'CREATE', 'UPDATE', 'ENABLE', 'DISABLE', 'ACKNOWLEDGE', 'RESOLVE', 'REOPEN'
    actor VARCHAR(128) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs(created_at DESC);
