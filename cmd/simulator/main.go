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
	"os/signal"
	"syscall"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/models"
)

func main() {
	targetURL := flag.String("url", "http://localhost:8081/api/v1/readings/batch", "Target batch ingestion URL")
	rate := flag.Int("rate", 50, "Readings per second")
	burstRate := flag.Int("burst", 0, "Burst rate to inject every burst-interval")
	burstInterval := flag.Duration("burst-interval", 15*time.Second, "Interval between bursts")
	spike := flag.Bool("spike", false, "Periodically inject threshold violation spikes")
	flag.Parse()

	log.Println("=========================================================")
	log.Println("  Campus Sensor Fleet Simulator (Standalone CLI)        ")
	log.Printf("  Target URL:       %s\n", *targetURL)
	log.Printf("  Normal Rate:      %d readings/sec\n", *rate)
	if *burstRate > 0 {
		log.Printf("  Burst Rate:       %d every %v\n", *burstRate, *burstInterval)
	}
	log.Printf("  Spikes Enabled:   %v\n", *spike)
	log.Println("=========================================================")

	sensors := []struct {
		id         string
		sensorType string
	}{
		{"lib-fl2-temp", "temperature"},
		{"lib-fl1-co2", "co2"},
		{"tt-lab3-temp", "temperature"},
		{"tt-lab3-hum", "humidity"},
		{"sjt-aud-occ", "occupancy"},
		{"sjt-fl5-noise", "noise"},
		{"mb-srv-temp", "temperature"},
		{"mb-air-pm25", "pm25"},
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	client := &http.Client{Timeout: 5 * time.Second}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var burstTicker *time.Ticker
	if *burstRate > 0 {
		burstTicker = time.NewTicker(*burstInterval)
		defer burstTicker.Stop()
	}

	totalSent := 0

	for {
		select {
		case <-stopChan:
			log.Printf("[simulator] Stopped. Total readings transmitted: %d\n", totalSent)
			return

		case <-ticker.C:
			currentBatchSize := *rate
			now := time.Now().UTC()
			batch := make([]models.IngestReadingRequest, 0, currentBatchSize)

			for i := 0; i < currentBatchSize; i++ {
				sen := sensors[rng.Intn(len(sensors))]
				val := 22.0 + rng.Float64()*4.0
				if sen.sensorType == "co2" {
					val = 450.0 + rng.Float64()*150.0
				} else if sen.sensorType == "occupancy" {
					val = float64(rng.Intn(35))
				}

				batch = append(batch, models.IngestReadingRequest{
					SensorID:  sen.id,
					Type:      sen.sensorType,
					Value:     val,
					Timestamp: now.Format(time.RFC3339),
					Quality:   "simulated",
				})
			}

			if *spike && rng.Float64() < 0.2 {
				// Inject deliberate threshold spike on server room sensor
				batch = append(batch, models.IngestReadingRequest{
					SensorID:  "mb-srv-temp",
					Type:      "temperature",
					Value:     38.5,
					Timestamp: now.Format(time.RFC3339),
					Quality:   "spike_simulated",
				})
			}

			go sendBatch(client, *targetURL, batch)
			totalSent += len(batch)
			fmt.Printf("\r[simulator] Transmitted: %d readings (last batch: %d)", totalSent, len(batch))
		}
	}
}

func sendBatch(client *http.Client, url string, readings []models.IngestReadingRequest) {
	reqBody, err := json.Marshal(models.BatchIngestRequest{Readings: readings})
	if err != nil {
		return
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err == nil && resp != nil {
		_ = resp.Body.Close()
	}
}
