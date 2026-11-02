# Graph Report - gdg-Backend  (2026-11-03)

## Corpus Check
- Corpus is ~35,359 words - fits in a single context window. You may not need a graph.

## Summary
- 579 nodes · 1550 edges · 30 communities (14 shown, 16 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 88 edges (avg confidence: 0.85)
- Token cost: 12,500 input · 2,800 output

## Community Hubs (Navigation)
- React UI & Lucide Icons
- Server Core & Auth Middleware
- CLI Tools & Simulator Harness
- Alert API & Handlers
- Ingestion & Telemetry Routes
- Docker Orchestration & Services
- Web Tooling & Test Suite
- In-Memory Store Implementation
- Alert Rules & Deduplication Logic
- Sensors & Readings Schema Migrations
- PostgreSQL & pgx Persistence
- Frontend App TypeScript Config
- Vite Node TypeScript Config
- Dashboard Summary & Aggregation
- Sensor Domain Models & State
- Time-Series Reading Ingestion
- Audit Logging System
- Trends & Latency Analytics
- Oxlint Linter Configuration
- Audit Log Schema Migration
- TypeScript Root Configuration
- Go Module Specification
- Favicon Brand Asset
- Hero Banner Asset
- Campus Icon Asset
- React Logo Asset
- Vite Logo Asset

## God Nodes (most connected - your core abstractions)
1. `Store` - 52 edges
2. `MemoryStore` - 38 edges
3. `PostgresStore` - 35 edges
4. `writeJSON()` - 30 edges
5. `Alert` - 28 edges
6. `writeError()` - 27 edges
7. `ApiClient` - 24 edges
8. `AlertRule` - 23 edges
9. `react` - 21 edges
10. `IngestionService` - 18 edges

## Surprising Connections (you probably didn't know these)
- `Deterministic Deduplication Key` --implements--> `GenerateDeduplicationKey()`  [EXTRACTED]
  docs/architecture.md → internal/alerts/deduplication.go
- `Dual Persistence Strategy` --rationale_for--> `MemoryStore`  [INFERRED]
  docs/architecture.md → internal/store/memory_store.go
- `Dual Persistence Strategy` --rationale_for--> `PostgresStore`  [INFERRED]
  docs/architecture.md → internal/store/postgres_store.go
- `CampusPulse Overview & Capabilities` --references--> `main()`  [EXTRACTED]
  README.md → cmd/server/main.go
- `Bruno Automated API Test Collection` --references--> `AlertHandler`  [EXTRACTED]
  bruno/README.md → internal/api/handlers_alerts.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **High-Throughput Batch Ingestion & Persistence** — docs_architecture_ingestion_pipeline, ingestion_ingestionservice_ingestbatch, store_postgresstore, store_memorystore [EXTRACTED 0.95]
- **Alert Engine & Incident Lifecycle** — docs_architecture_alert_engine, internal_alerts_evaluator, docs_architecture_incident_lifecycle, api_alerthandler [EXTRACTED 0.95]
- **Load Testing & Simulator Harness** — docs_load_test_report_benchmark, cmd_loadtest_main, cmd_simulator_main [EXTRACTED 0.95]

## Communities (30 total, 16 thin omitted)

### Community 0 - "React UI & Lucide Icons"
Cohesion: 0.05
Nodes (56): lucide-react, react, react-dom, dependencies, lucide-react@^1.61.0, react@^19.3.1, react-dom@^19.3.1, CampusPulse React Frontend App (+48 more)

### Community 1 - "Server Core & Auth Middleware"
Cohesion: 0.15
Nodes (4): LoginRequest, ResetSchema(), BatchIngestResponse, BatchItemResult

### Community 2 - "CLI Tools & Simulator Harness"
Cohesion: 0.06
Nodes (43): generateBatch(), main(), sendBatchRequest(), main(), isStoreEmpty(), main(), seedDemoData(), main() (+35 more)

### Community 3 - "Alert API & Handlers"
Cohesion: 0.16
Nodes (14): AlertHandler, HealthHandler, IngestHandler, RulesHandler, SensorHandler, SimulatorHandler, Bruno Automated API Test Collection, Incident Lifecycle State Machine (+6 more)

### Community 4 - "Ingestion & Telemetry Routes"
Cohesion: 0.07
Nodes (24): AuthHandler, contextKey, LoginResponse, ReadingsHandler, ServerComponents, NewAlertHandler(), NewAuthHandler(), NewHealthHandler() (+16 more)

### Community 5 - "Docker Orchestration & Services"
Cohesion: 0.06
Nodes (34): build, container_name, depends_on, environment, ports, restart, context, dockerfile (+26 more)

### Community 6 - "Web Tooling & Test Suite"
Cohesion: 0.08
Nodes (33): jsdom, oxlint, @testing-library/jest-dom, @testing-library/react, @types/node, @types/react, @types/react-dom, typescript (+25 more)

### Community 8 - "Alert Rules & Deduplication Logic"
Cohesion: 0.11
Nodes (12): TestCompareOperator(), TestComputeAggregation(), TestMatchesScope(), GenerateDeduplicationKey(), CompareOperator(), MatchesScope(), ComputeAggregation(), AlertRule (+4 more)

### Community 9 - "Sensors & Readings Schema Migrations"
Cohesion: 0.16
Nodes (17): idx_sensors_building, idx_sensors_status, idx_sensors_type, sensors, idx_readings_idempotency, idx_readings_observed_at, idx_readings_sensor_time, readings (+9 more)

### Community 11 - "Frontend App TypeScript Config"
Cohesion: 0.11
Nodes (18): compilerOptions, allowImportingTsExtensions, isolatedModules, jsx, lib, module, moduleDetection, moduleResolution (+10 more)

### Community 12 - "Vite Node TypeScript Config"
Cohesion: 0.12
Nodes (16): compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module, moduleDetection, noEmit, noFallthroughCasesInSwitch (+8 more)

### Community 13 - "Dashboard Summary & Aggregation"
Cohesion: 0.14
Nodes (6): WindowEvaluator, DashboardHandler, NewWindowEvaluator(), NewDashboardHandler(), DashboardSummary, Store

### Community 18 - "Oxlint Linter Configuration"
Cohesion: 0.33
Nodes (5): plugins, rules, react/only-export-components, react/rules-of-hooks, $schema

### Community 19 - "Audit Log Schema Migration"
Cohesion: 0.83
Nodes (3): audit_logs, idx_audit_created_at, idx_audit_entity

## Knowledge Gaps
- **8 isolated node(s):** `github.com/shikhar102005/gdg-Backend`, `pgx.CopyFrom Bulk Binary Streaming`, `CampusPulse React Frontend App`, `Asset: favicon.svg`, `Asset: icons.svg` (+3 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 117 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **16 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Store` connect `Dashboard Summary & Aggregation` to `Server Core & Auth Middleware`, `CLI Tools & Simulator Harness`, `Alert API & Handlers`, `Ingestion & Telemetry Routes`, `In-Memory Store Implementation`, `Alert Rules & Deduplication Logic`, `Sensor Domain Models & State`, `Time-Series Reading Ingestion`, `Audit Logging System`, `Trends & Latency Analytics`?**
  _High betweenness centrality (0.055) - this node is a cross-community bridge._
- **Are the 27 inferred relationships involving `writeJSON()` (e.g. with `.Acknowledge()` and `.Get()`) actually correct?**
  _`writeJSON()` has 27 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/shikhar102005/gdg-Backend`, `pgx.CopyFrom Bulk Binary Streaming`, `CampusPulse React Frontend App` to the rest of the system?**
  _8 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `React UI & Lucide Icons` be split into smaller, more focused modules?**
  _Cohesion score 0.050915750915750915 - nodes in this community are weakly interconnected._
- **Why does `MemoryStore` connect `In-Memory Store Implementation` to `Server Core & Auth Middleware`, `CLI Tools & Simulator Harness`, `Ingestion & Telemetry Routes`, `Alert Rules & Deduplication Logic`, `Dashboard Summary & Aggregation`, `Sensor Domain Models & State`, `Time-Series Reading Ingestion`, `Audit Logging System`, `Trends & Latency Analytics`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Should `Server Core & Auth Middleware` be split into smaller, more focused modules?**
  _Cohesion score 0.146448087431694 - nodes in this community are weakly interconnected._
- **Why does `Simulator` connect `Ingestion & Telemetry Routes` to `Server Core & Auth Middleware`, `CLI Tools & Simulator Harness`, `Alert API & Handlers`, `Dashboard Summary & Aggregation`, `Time-Series Reading Ingestion`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._