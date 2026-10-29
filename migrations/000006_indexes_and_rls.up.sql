-- Migration: 000006_indexes_and_rls.up.sql

-- Enable Row Level Security on all core tables for Supabase environments
ALTER TABLE sensors ENABLE ROW LEVEL SECURITY;
ALTER TABLE readings ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

-- Service role bypasses RLS in Postgres/Supabase by default.
-- For standard authenticated or anonymous roles (if direct client querying is enabled):
-- Allow read access to sensors, readings, alert rules, alerts, and audit logs
CREATE POLICY "Allow public/anon read sensors" ON sensors FOR SELECT USING (true);
CREATE POLICY "Allow public/anon read readings" ON readings FOR SELECT USING (true);
CREATE POLICY "Allow public/anon read alert_rules" ON alert_rules FOR SELECT USING (true);
CREATE POLICY "Allow public/anon read alerts" ON alerts FOR SELECT USING (true);
CREATE POLICY "Allow public/anon read audit_logs" ON audit_logs FOR SELECT USING (true);

-- Backend service role (or postgres superuser) handles writes.
-- In case app connects with authenticated role:
CREATE POLICY "Allow service/authenticated insert readings" ON readings FOR INSERT WITH CHECK (true);
CREATE POLICY "Allow service/authenticated manage sensors" ON sensors FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow service/authenticated manage alert_rules" ON alert_rules FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow service/authenticated manage alerts" ON alerts FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow service/authenticated insert audit_logs" ON audit_logs FOR INSERT WITH CHECK (true);
