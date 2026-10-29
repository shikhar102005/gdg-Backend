package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/models"
)

type LatencyRecord struct {
	Duration time.Duration
	Success  bool
}

func main() {
	targetURL := flag.String("url", "http://localhost:8081/api/v1/readings/batch", "Target batch ingestion URL")
	targetRate := flag.Int("target-rate", 10000, "Target readings per second")
	batchSize := flag.Int("batch-size", 500, "Readings per batch HTTP payload")
	concurrency := flag.Int("concurrency", 20, "Number of concurrent workers")
	duration := flag.Duration("duration", 5*time.Second, "Test run duration")
	burstScenario := flag.Bool("burst", true, "Execute 10,000 readings 1-second burst scenario")
	reportFile := flag.String("report", "docs/load-test-report.md", "Output path for markdown report")
	flag.Parse()

	log.Println("=================================================================")
	log.Println("  Campus Sensor Monitoring - Benchmark & Load-Test Harness      ")
	log.Println("=================================================================")
	log.Printf("Target:           %s\n", *targetURL)
	log.Printf("Target Rate:      %d readings/sec\n", *targetRate)
	log.Printf("Batch Size:       %d readings/request\n", *batchSize)
	log.Printf("Concurrency:      %d workers\n", *concurrency)
	log.Printf("Duration:         %v\n", *duration)
	log.Printf("Burst Mode:       %v\n", *burstScenario)
	log.Println("=================================================================")

	// Pre-generate sample sensor IDs and types
	sensorCount := 50
	sensorIDs := make([]string, sensorCount)
	for i := 0; i < sensorCount; i++ {
		sensorIDs[i] = fmt.Sprintf("bench-sensor-%03d", i+1)
	}
	metricTypes := []string{"temperature", "humidity", "co2", "occupancy", "noise"}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 500,
			MaxConnsPerHost:     500,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	var totalReadingsSubmitted int64
	var totalReadingsAccepted int64
	var totalRequests int64
	var failedRequests int64

	recordsMu := sync.Mutex{}
	records := make([]LatencyRecord, 0, 50000)

	startTime := time.Now()
	var wg sync.WaitGroup

	if *burstScenario {
		log.Println("[benchmark] Starting 10,000 readings 1-second burst scenario...")
		// Generate 10,000 readings in batches of batchSize
		batchesNeeded := 10000 / *batchSize
		if batchesNeeded <= 0 {
			batchesNeeded = 1
		}

		burstWg := sync.WaitGroup{}
		burstStart := time.Now()

		for i := 0; i < batchesNeeded; i++ {
			burstWg.Add(1)
			go func(batchIdx int) {
				defer burstWg.Done()
				batch := generateBatch(*batchSize, sensorIDs, metricTypes)
				atomic.AddInt64(&totalReadingsSubmitted, int64(len(batch)))
				atomic.AddInt64(&totalRequests, 1)

				reqStart := time.Now()
				accepted, err := sendBatchRequest(client, *targetURL, batch)
				reqDur := time.Since(reqStart)

				recordsMu.Lock()
				records = append(records, LatencyRecord{Duration: reqDur, Success: err == nil})
				recordsMu.Unlock()

				if err != nil {
					atomic.AddInt64(&failedRequests, 1)
				} else {
					atomic.AddInt64(&totalReadingsAccepted, int64(accepted))
				}
			}(i)
		}
		burstWg.Wait()
		log.Printf("[benchmark] Burst completed in %v\n", time.Since(burstStart))
	} else {
		log.Printf("[benchmark] Starting sustained throughput test for %v...\n", *duration)
		stopChan := make(chan struct{})
		time.AfterFunc(*duration, func() { close(stopChan) })

		for i := 0; i < *concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stopChan:
						return
					default:
						batch := generateBatch(*batchSize, sensorIDs, metricTypes)
						atomic.AddInt64(&totalReadingsSubmitted, int64(len(batch)))
						atomic.AddInt64(&totalRequests, 1)

						reqStart := time.Now()
						accepted, err := sendBatchRequest(client, *targetURL, batch)
						reqDur := time.Since(reqStart)

						recordsMu.Lock()
						records = append(records, LatencyRecord{Duration: reqDur, Success: err == nil})
						recordsMu.Unlock()

						if err != nil {
							atomic.AddInt64(&failedRequests, 1)
						} else {
							atomic.AddInt64(&totalReadingsAccepted, int64(accepted))
						}
					}
				}
			}()
		}
		wg.Wait()
	}

	totalDuration := time.Since(startTime)
	actualReadingsSec := float64(totalReadingsAccepted) / totalDuration.Seconds()
	actualReqsSec := float64(totalRequests) / totalDuration.Seconds()

	// Compute latencies
	sort.Slice(records, func(i, j int) bool {
		return records[i].Duration < records[j].Duration
	})

	var p50, p90, p95, p99, minLat, maxLat time.Duration
	if len(records) > 0 {
		minLat = records[0].Duration
		maxLat = records[len(records)-1].Duration
		p50 = records[int(float64(len(records))*0.50)].Duration
		p90 = records[int(float64(len(records))*0.90)].Duration
		p95 = records[int(float64(len(records))*0.95)].Duration
		p99 = records[int(float64(len(records))*0.99)].Duration
	}

	successRate := 0.0
	if totalRequests > 0 {
		successRate = float64(totalRequests-failedRequests) / float64(totalRequests) * 100.0
	}

	fmt.Println("\n=================================================================")
	fmt.Println("                       BENCHMARK RESULTS                        ")
	fmt.Println("=================================================================")
	fmt.Printf("Total Elapsed:          %v\n", totalDuration)
	fmt.Printf("Readings Submitted:     %d\n", totalReadingsSubmitted)
	fmt.Printf("Readings Accepted:      %d\n", totalReadingsAccepted)
	fmt.Printf("Total HTTP Requests:    %d\n", totalRequests)
	fmt.Printf("Failed HTTP Requests:   %d (%.2f%% error rate)\n", failedRequests, 100.0-successRate)
	fmt.Printf("Achieved Throughput:    %.2f readings/second\n", actualReadingsSec)
	fmt.Printf("Request Rate:           %.2f requests/second\n", actualReqsSec)
	fmt.Printf("Latency (Min):          %v\n", minLat)
	fmt.Printf("Latency (p50):          %v\n", p50)
	fmt.Printf("Latency (p90):          %v\n", p90)
	fmt.Printf("Latency (p95):          %v\n", p95)
	fmt.Printf("Latency (p99):          %v\n", p99)
	fmt.Printf("Latency (Max):          %v\n", maxLat)
	fmt.Println("=================================================================")

	// Generate reproducible markdown report
	reportContent := fmt.Sprintf(`# Campus Sensor Monitoring & Alerts - Load Test & Benchmark Report

## Execution Metadata
- **Date & Time:** %s
- **Host OS / Arch:** %s / %s
- **CPU Cores:** %d logical CPUs
- **Go Version:** %s
- **Target URL:** %s

## Benchmark Configuration
- **Target Ingestion Goal:** %d readings/sec
- **Test Mode:** %s
- **Batch Size:** %d readings/payload
- **Worker Concurrency:** %d goroutines
- **Target Elapsed:** %v

## Measured Results
| Metric | Measured Value | Target |
| :--- | :--- | :--- |
| **Achieved Ingestion Throughput** | **%.2f readings/sec** | 10,000 readings/sec |
| **HTTP Request Rate** | **%.2f reqs/sec** | - |
| **Total Readings Submitted** | %d | 10,000+ |
| **Total Readings Accepted** | %d | 10,000+ |
| **Total Requests** | %d | - |
| **Failed Requests** | %d | 0 |
| **Success Rate** | **%.2f%%%%** | >= 99.9%%%% |
| **Latency Min** | %v | - |
| **Latency p50** | %v | < 100ms |
| **Latency p90** | %v | < 250ms |
| **Latency p95** | %v | < 500ms |
| **Latency p99** | %v | < 1000ms |
| **Latency Max** | %v | - |

## Architectural Analysis & Scaling Boundaries

### Ingestion Pipeline
1. **Durability Guarantee:** Readings are accepted and committed within the batch ingestion transaction before responding to the client.
2. **Bulk Ingestion Path:**
   - On PostgreSQL: High-throughput pgx.CopyFrom streaming binary rows into the table, bypassing individual INSERT overhead.
   - On Memory Store: Atomic sliding buffer writes with O(1) indexed idempotency map lookups.
3. **Decoupled Alert Processing:** Single-reading and time-window evaluation are processed asynchronously through a bounded worker pool channel (%d workers), preventing alert evaluation rules from blocking high-burst ingestion.
4. **Horizontal Scaling:**
   - Multiple stateless Go server instances behind a layer 4/7 reverse proxy (e.g. NGINX, Cloudflare).
   - PostgreSQL table partitioning by date range (e.g. daily or weekly partitions) prevents B-Tree index bloating on tables with millions of readings.
   - Read replicas for dashboard trends and analytic queries, keeping the primary connection pool dedicated to writes and alerts.
`,
		time.Now().UTC().Format(time.RFC3339),
		runtime.GOOS, runtime.GOARCH,
		runtime.NumCPU(),
		runtime.Version(),
		*targetURL,
		*targetRate,
		map[bool]string{true: "10,000 Burst Scenario", false: "Sustained Concurrency"}[*burstScenario],
		*batchSize,
		*concurrency,
		totalDuration,
		actualReadingsSec,
		actualReqsSec,
		totalReadingsSubmitted,
		totalReadingsAccepted,
		totalRequests,
		failedRequests,
		successRate,
		minLat, p50, p90, p95, p99, maxLat,
		runtime.NumCPU(),
	)

	_ = os.WriteFile(*reportFile, []byte(reportContent), 0644)
	log.Printf("[benchmark] Report written to %s\n", *reportFile)
}

func generateBatch(size int, sensors []string, metrics []string) []models.IngestReadingRequest {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	batch := make([]models.IngestReadingRequest, size)
	now := time.Now().UTC()

	for i := 0; i < size; i++ {
		sen := sensors[rng.Intn(len(sensors))]
		met := metrics[rng.Intn(len(metrics))]
		val := 20.0 + rng.Float64()*10.0

		batch[i] = models.IngestReadingRequest{
			SensorID:  sen,
			Type:      met,
			Value:     val,
			Timestamp: now.Format(time.RFC3339),
			Quality:   "load_test",
		}
	}
	return batch
}

func sendBatchRequest(client *http.Client, url string, batch []models.IngestReadingRequest) (int, error) {
	reqBody, err := json.Marshal(models.BatchIngestRequest{Readings: batch})
	if err != nil {
		return 0, err
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	var batchResp models.BatchIngestResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return len(batch), nil
	}
	return batchResp.AcceptedCount, nil
}
