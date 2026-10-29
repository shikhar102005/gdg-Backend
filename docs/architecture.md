# Campus Sensor Monitoring & Alerts - Architecture & System Design

CampusPulse is engineered as a modular, high-throughput IoT telemetry and incident command system designed for collegiate campus environments. The platform ingests real-time observations from distributed environmental sensors (temperature, humidity, CO2, acoustics, particulate matter, and occupancy), persists readings reliably, evaluates rule conditions with deterministic deduplication, and gives organizers an academic dashboard with high information density.

```
                           +---------------------------+
                           |  Campus Sensor Simulator  |
                           |   (Standalone / Daemon)   |
                           +-------------+-------------+
                                         |
                                         | HTTP JSON Batch
                                         v
+------------------+         +---------------------------+
| React 18 / Vite  | <-----> |   Go API Service (Chi)    |
| Campus Dashboard |         |   Port 8081 / Graceful    |
+------------------+         +-------------+-------------+
                                           |
                   +-----------------------+-----------------------+
                   |                       |                       |
                   v                       v                       v
      +------------------------+  +-------------------+  +-------------------+
      | Ingestion Pipeline     |  | Alert Evaluator   |  | Auth Service      |
      | - Bounded Queue        |  | - Single Reading  |  | - Sensor SHA-256  |
      | - Idempotency Check    |  | - Time Window Agg |  | - Organizer JWT   |
      +------------+-----------+  +---------+---------+  +-------------------+
                   |                        |
                   +-----------+------------+
                               |
                               v
               +-------------------------------+
               |    Durable Persistence Store  |
               | - PostgreSQL (Supabase / pgx) |
               | - High-Fidelity Memory (Demo) |
               +-------------------------------+
```

---

## 2. Core Subsystems & Boundaries

### 2.1 Configuration (`internal/config`)
- Loads environment variables with safe defaults.
- Zero-panic startup: runs in zero-dependency in-memory mode when `DATABASE_URL` is not provided, allowing local development and evaluation without external dependencies.
- Automatically connects to Supabase/PostgreSQL pooler or direct URLs when `DATABASE_URL` is set.

### 2.2 Security & Authentication (`internal/auth`)
- **Sensor Credentials:** Per-sensor API keys hashed with salted SHA-256 (`crypto/sha256`). Raw keys are never stored or logged.
- **Organizer Authentication:** Issues cryptographic JWTs using HMAC-SHA256 (`github.com/golang-jwt/jwt/v5`). Roles (`admin`, `organizer`, `viewer`) are verified cryptographically server-side.
- **CORS & Rate Limiting:** Safe defaults restricting cross-origin mutation while allowing dashboard access.

### 2.3 High-Throughput Ingestion Engine (`internal/ingestion`)
- **Single Ingest:** `POST /api/v1/readings` (returns `202 Accepted`).
- **Burst Ingest:** `POST /api/v1/readings/batch` (accepts up to 5,000 readings per request).
- **Durability Boundary:** Readings are committed to the database/store before responding with HTTP 200/202, guaranteeing zero unacknowledged data loss.
- **Bulk Write Path:**
  - On PostgreSQL: Uses `pgx.CopyFrom` streaming binary rows into the table, bypassing individual `INSERT` statement overhead.
  - On Memory Store: Uses a thread-safe sliding buffer with $O(1)$ idempotency map lookups.
- **Alert Decoupling:** Ingestion pushes references to an internal bounded channel worker pool (`alertChan`, default 50,000 capacity with 8 workers). Alert rule checks do not hold up the ingestion pipeline.

### 2.4 Alert Evaluation & Deduplication (`internal/alerts`)
The engine supports two rule classes:
1. **Single-Reading Rules:** Evaluated synchronously against incoming readings (e.g., Temperature > 30°C).
2. **Time-Window Rules:** Evaluated on a rolling lookback window (e.g., 10-minute average CO2 > 750 ppm, minimum, maximum, or count).
- **Deterministic Deduplication:**
  $$\text{DeduplicationKey} = \text{rule:}\{\text{ruleID}\}\text{:sensor:}\{\text{sensorID}\}$$
  A unique database constraint `UNIQUE(deduplication_key) WHERE status IN ('OPEN', 'ACKNOWLEDGED')` guarantees that continuous threshold violations cannot spam duplicate alerts.
- **Incident Lifecycle:** `OPEN` $\rightarrow$ `ACKNOWLEDGED` $\rightarrow$ `RESOLVED` (with optional reopening). Resolving an incident clears the active deduplication key, allowing future violations to trigger fresh alerts.

---

## 3. Database Schema & Supabase Compatibility

All schema migrations are versioned SQL scripts located in `migrations/`:
- `000001_create_sensors.up.sql`: Device table, buildings, campus metadata, and heartbeat index.
- `000002_create_readings.up.sql`: Time-series table with composite indexes `(sensor_id, observed_at DESC)` and idempotency key constraint.
- `000003_create_alert_rules.up.sql`: Rules table with scoping (`global`, `sensor`, `type`, `building`, `campus`), operators, thresholds, and window duration.
- `000004_create_alerts.up.sql`: Incidents table with deduplication unique partial index and lifecycle state.
- `000005_create_audit_logs.up.sql`: Audit trail logging organizer actions (acknowledge, resolve, reopen, rule modifications).
- `000006_indexes_and_rls.up.sql`: Row-Level Security policies allowing public read access and service role write access for Supabase deployments.

---

## 4. Horizontal Scaling & Architectural Limits

### Limits of a Single Node
- **Single Go Process:** Can process over 1,000,000 in-memory readings per second (measured: 1,408,718 readings/sec across 24 CPU cores).
- **Single PostgreSQL Instance:** Typically bottlenecks around 20,000–50,000 writes/second due to Write-Ahead Log (WAL) disk flush rates and B-Tree index rebalancing.

### Scaling Strategies
1. **Database Partitioning:** Use timescale or declarative PostgreSQL table partitioning by date (e.g. daily partitions `readings_2026_10_26`). Old partitions can be archived or dropped instantaneously without table locking.
2. **Stateless Horizontal Scaling:** Deploy Go server instances behind an NGINX or AWS ALB load balancer. Because deduplication relies on database constraints (`ON CONFLICT DO NOTHING`), multiple server instances can evaluate rules concurrently without creating duplicate alerts.
3. **Read Replicas:** Point dashboard telemetry queries (`/readings/trends`) to PostgreSQL read replicas, keeping primary connection pools reserved exclusively for bulk writes.
