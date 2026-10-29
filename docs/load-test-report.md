# Campus Sensor Monitoring & Alerts - Load Test & Benchmark Report

## Execution Metadata
- **Date & Time:** 2026-10-25T23:20:52Z
- **Host OS / Arch:** windows / amd64
- **CPU Cores:** 24 logical CPUs
- **Go Version:** go1.26.0
- **Target URL:** http://localhost:8081/api/v1/readings/batch

## Benchmark Configuration
- **Target Ingestion Goal:** 10000 readings/sec
- **Test Mode:** Sustained Concurrency
- **Batch Size:** 500 readings/payload
- **Worker Concurrency:** 20 goroutines
- **Target Elapsed:** 3.0023735s

## Measured Results
| Metric | Measured Value | Target |
| :--- | :--- | :--- |
| **Achieved Ingestion Throughput** | **1408718.80 readings/sec** | 10,000 readings/sec |
| **HTTP Request Rate** | **2817.44 reqs/sec** | - |
| **Total Readings Submitted** | 4229500 | 10,000+ |
| **Total Readings Accepted** | 4229500 | 10,000+ |
| **Total Requests** | 8459 | - |
| **Failed Requests** | 0 | 0 |
| **Success Rate** | **100.00%%** | >= 99.9%% |
| **Latency Min** | 0s | - |
| **Latency p50** | 3.283ms | < 100ms |
| **Latency p90** | 19.8684ms | < 250ms |
| **Latency p95** | 30.9069ms | < 500ms |
| **Latency p99** | 49.0861ms | < 1000ms |
| **Latency Max** | 115.4823ms | - |

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
