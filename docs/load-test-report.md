# Campus Sensor Monitoring & Alerts - Load Test & Benchmark Report

## Execution Metadata
- **Date & Time:** 2026-10-29T08:31:49Z
- **Host OS / Arch:** windows / amd64
- **CPU Cores:** 24 logical CPUs
- **Go Version:** go1.26.0
- **Target URL:** http://localhost:8081/api/v1/readings/batch

## Benchmark Configuration
- **Target Ingestion Goal:** 10000 readings/sec
- **Test Mode:** 10,000 Burst Scenario
- **Batch Size:** 500 readings/payload
- **Worker Concurrency:** 20 goroutines
- **Target Elapsed:** 4.011199s

## Measured Results
| Metric | Measured Value | Target |
| :--- | :--- | :--- |
| **Achieved Ingestion Throughput** | **2493.02 readings/sec** | 10,000 readings/sec |
| **HTTP Request Rate** | **4.99 reqs/sec** | - |
| **Total Readings Submitted** | 10000 | 10,000+ |
| **Total Readings Accepted** | 10000 | 10,000+ |
| **Total Requests** | 20 | - |
| **Failed Requests** | 0 | 0 |
| **Success Rate** | **100.00%%** | >= 99.9%% |
| **Latency Min** | 1.1230344s | - |
| **Latency p50** | 1.6323496s | < 100ms |
| **Latency p90** | 2.0647564s | < 250ms |
| **Latency p95** | 4.0106669s | < 500ms |
| **Latency p99** | 4.0106669s | < 1000ms |
| **Latency Max** | 4.0106669s | - |

## Architectural Analysis & Scaling Boundaries

### Ingestion Pipeline
1. **Durability Guarantee:** Readings are accepted and committed within the batch ingestion transaction before responding to the client.
2. **Bulk Ingestion Path:**
   - On PostgreSQL: High-throughput pgx.CopyFrom streaming binary rows into the table, bypassing individual INSERT overhead.
   - On Memory Store: Atomic sliding buffer writes with O(1) indexed idempotency map lookups.
3. **Decoupled Alert Processing:** Single-reading and time-window evaluation are processed asynchronously through a bounded worker pool channel (24 workers), preventing alert evaluation rules from blocking high-burst ingestion.
4. **Horizontal Scaling:**
   - Multiple stateless Go server instances behind a layer 4/7 reverse proxy (e.g. NGINX, Cloudflare).
   - PostgreSQL table partitioning by date range (e.g. daily or weekly partitions) prevents B-Tree index bloating on tables with millions of readings.
   - Read replicas for dashboard trends and analytic queries, keeping the primary connection pool dedicated to writes and alerts.
