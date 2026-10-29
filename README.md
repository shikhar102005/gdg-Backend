# CampusPulse &bull; Campus Sensor Monitoring & Alerts Engine

A complete, high-throughput IoT environmental monitoring and incident command system engineered with a **Go backend**, **PostgreSQL / Supabase durable persistence**, and a modern, high-density **React + TypeScript + Vite frontend**.

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React Version](https://img.shields.io/badge/React-19.3-61DAFB?style=flat&logo=react)](https://react.dev/)
[![Vite](https://img.shields.io/badge/Vite-8.3-646CFF?style=flat&logo=vite)](https://vitejs.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## Highlights & Implemented Capabilities

- **High-Burst Ingestion Pipeline:** Benchmarked and measured to handle bursts targeting **10,000 readings per second** (achieved **297,811+ readings/sec** on burst load testing and **1.4M+ readings/sec** sustained).
- **Dual Persistent Architecture:**
  - **Durable PostgreSQL / Supabase:** Native `pgx/v5` connection pooler with high-speed `pgx.CopyFrom` streaming binary bulk ingestion, versioned migrations, and Row Level Security (RLS) policies.
  - **Zero-Dependency Fallback Store:** High-fidelity concurrent in-memory store allowing immediate evaluation without external database setup.
- **Alert Rules Engine:**
  - **Single-Reading Rules:** Immediate threshold evaluation (e.g. Temperature > 28°C in server room).
  - **Time-Window Rules:** Rolling aggregate computation (`avg`, `min`, `max`, `count`) over configurable lookback windows (e.g. 10m Average CO2 > 750 ppm).
  - **Deterministic Deduplication:** Correlation key prevents alert spam while incidents remain active.
  - **Incident Lifecycle:** `OPEN` $\rightarrow$ `ACKNOWLEDGED` $\rightarrow$ `RESOLVED` (with root-cause resolution notes) $\rightarrow$ `REOPEN`.
- **Campus Portal Frontend:** High-density academic dashboard featuring collegiate navy styling, compact information tables, live SVG telemetry trend charts, alert lifecycle modals, and simulator controls.
- **Sensor Simulator:** Built-in background daemon and standalone CLI tool generating realistic telemetry, drift, and controlled threshold spikes.
- **Bruno Collection:** Native Bruno API collection in `bruno/` with full endpoint coverage, environment configuration, and variable chaining.
- **Reproducible Load-Test Harness:** Benchmark tool recording p50, p90, p95, and p99 latencies, generating markdown reports.

---

## Repository Structure

```
gdg-Backend/
├── cmd/
│   ├── server/       # Main backend application entrypoint
│   ├── seed/         # Standalone database and demo data seeder CLI
│   ├── simulator/    # Standalone sensor fleet simulator CLI
│   └── loadtest/     # Benchmark & load-testing harness (10K burst target)
├── internal/
│   ├── alerts/       # Single-reading, window aggregators, deduplication logic
│   ├── api/          # Chi router, handlers, CORS, and auth middleware
│   ├── auth/         # JWT generation/validation & SHA-256 sensor key hashing
│   ├── config/       # Environment configuration loader
│   ├── database/     # SQL migration runner
│   ├── ingestion/    # High-throughput batch ingestion pipeline
│   ├── models/       # Domain data structures & API requests/responses
│   ├── simulator/    # Telemetry generation engine & spike injector
│   └── store/        # PostgreSQL store (pgxpool) & thread-safe memory store
├── migrations/       # Versioned SQL migrations (Sensors, Readings, Rules, Alerts, Audit, RLS)
├── web/              # React 18, TypeScript, Vite campus dashboard
│   ├── src/
│   │   ├── api/      # API client for backend communication
│   │   ├── components/ # Navbar, MetricCard, TrendsChart, AlertActionModal, RuleFormModal
│   │   ├── pages/    # Overview, Sensors, Trends, Alerts, Rules, Simulator
│   │   └── test/     # Vitest unit & component tests
├── bruno/            # Complete Bruno API collection with chained environment
├── docs/             # OpenAPI specification, architecture docs, and benchmark reports
├── docker-compose.yml# Local PostgreSQL development orchestration
├── .env.example      # Documented environment variables
└── README.md
```

---

## Quick Start (Zero-Dependency Local Run)

You can run both backend and frontend locally in less than 60 seconds without installing or configuring PostgreSQL.

### 1-Click Startup (Recommended)

You can launch both the Go backend and React frontend with a single command:

**Windows (Batch / Double-click):**
```cmd
start.bat
```
*(Or simply double-click `start.bat` in Windows Explorer)*

**Windows (PowerShell):**
```powershell
.\start.ps1
```

To stop all services later, run `stop.bat` or `.\stop.ps1`.

---

### Manual Startup

#### 1. Start the Go Backend
```bash
# From workspace root
go run ./cmd/server
```
The server will start on `http://localhost:8081`, auto-seed demo sensors and alert rules, and initialize the telemetry simulator.

### 2. Start the React Frontend

```bash
# In a new terminal
cd web
npm install
npm run dev
```
Open **`http://localhost:5173`** in your browser to access the campus dashboard.

### Demo Organizer Credentials
- **Admin Role:** `organizer@campus.edu` / `CampusOrganizerSecure2026!`
- **Viewer Role:** `viewer@campus.edu` / `CampusViewerDemo2026!`

---

## Supabase & Local PostgreSQL Setup

### Using Docker Compose (Local PostgreSQL)
```bash
docker compose up -d postgres
```

### Using Supabase PostgreSQL
1. Create a project at [supabase.com](https://supabase.com).
2. Copy your PostgreSQL connection string from **Project Settings > Database** (use Transaction Pooler or Direct connection).
3. Set `DATABASE_URL` in your environment or `.env`:
   ```bash
   DATABASE_URL="postgres://postgres.[ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres?sslmode=require"
   ```
4. Run the seed tool to apply migrations and seed data:
   ```bash
   go run ./cmd/seed
   ```
5. Start the server:
   ```bash
   go run ./cmd/server
   ```

---

## Testing & Quality Assurance

### Run Backend Unit & Integration Tests
```bash
go test -v -count=1 ./...
```
*Executes validation tests, auth token verification, single-reading rule operators, time-window aggregations (`avg`, `min`, `max`, `count`), deduplication state transitions, and API endpoints.*

### Run Frontend Component Tests & Build
```bash
cd web
npm test       # Executes Vitest suite
npm run build  # Validates TypeScript compilation and generates production bundle
```

---

## Performance Benchmark & Load Testing

A reproducible benchmark tool is provided in `cmd/loadtest`. It can simulate bursts of 10,000 readings/sec, measure achieved throughput, and record latency percentiles (p50, p90, p95, p99).

### 1. 10,000 Readings 1-Second Burst Scenario
```bash
go run ./cmd/loadtest -target-rate 10000 -batch-size 500 -concurrency 20 -burst=true
```

### 2. Sustained Concurrency Benchmark
```bash
go run ./cmd/loadtest -target-rate 10000 -batch-size 500 -concurrency 20 -duration 5s -burst=false
```

### Measured Results (Actual Run on 24-Core Host)
| Scenario | Readings Submitted | Readings Accepted | Duration | Achieved Throughput | Latency p50 | Latency p95 | Latency p99 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **10K Burst** | 10,000 | 10,000 (100%) | 33.58 ms | **297,811/sec** | 29.7 ms | 33.6 ms | 33.6 ms |
| **Sustained** | 4,229,500 | 4,229,500 (100%) | 3.00 sec | **1,408,718/sec** | 3.28 ms | 30.9 ms | 49.1 ms |

*Detailed benchmark reports are generated in [`docs/load-test-report.md`](docs/load-test-report.md).*

---

## Bruno API Collection

A full Bruno collection is included in [`bruno/`](bruno/) covering every route:
- **Health & Readiness:** `/healthz`, `/readyz`
- **Authentication:** Admin login, viewer login, invalid credentials error, `/auth/me`
- **Ingestion:** Single reading, batch reading burst, invalid reading validation error
- **Sensors:** List sensors, filter by building/status, get by ID, get reading history, upsert sensor
- **Trends & Analytics:** Time-bucketed aggregates for temperature, humidity, CO2, occupancy, etc.
- **Alert Rules:** Create single-reading and time-window rules, get, update, toggle
- **Alert Lifecycle:** List alerts, get details, acknowledge, resolve with notes, reopen
- **Simulator:** Status, start, stop, trigger threshold violation spike

Refer to [`bruno/README.md`](bruno/README.md) for step-by-step chaining instructions.

---

## Architecture & Scaling Boundaries

Refer to [`docs/architecture.md`](docs/architecture.md) for full architectural analysis:
1. **Durability Boundary:** Readings are committed inside the batch transaction before HTTP 200/202 is returned.
2. **Alert Worker Decoupling:** Bounded in-memory queues (50,000 buffer) prevent slow alert notifications from blocking ingestion writes.
3. **Database Scaling:** PostgreSQL date range partitioning (e.g. daily partitions) and read replicas for dashboard queries.