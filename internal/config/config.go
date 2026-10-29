package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerPort                   string
	ServerHost                   string
	ServerReadTimeout            time.Duration
	ServerWriteTimeout           time.Duration
	ServerIdleTimeout            time.Duration
	CORSAllowedOrigins           []string
	DatabaseURL                  string
	DatabaseMaxConns             int32
	DatabaseMinConns             int32
	DatabaseMaxConnLifetime      time.Duration
	DatabaseMaxConnIdleTime      time.Duration
	IngestionBatchMaxSize        int
	IngestionWorkers             int
	IngestionQueueBuffer         int
	OrganizerJWTSecret           string
	OrganizerAdminEmail          string
	OrganizerAdminPassword       string
	OrganizerViewerEmail         string
	OrganizerViewerPassword      string
	SensorKeySalt                string
	SensorOfflineThreshold       time.Duration
	AlertWindowEvaluationInterval time.Duration
}

func loadDotEnv() {
	paths := []string{".env", "../.env"}
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
					val = val[1 : len(val)-1]
				}
				if os.Getenv(key) == "" {
					os.Setenv(key, val)
				}
			}
		}
		break
	}
}

func Load() *Config {
	loadDotEnv()

	cfg := &Config{
		ServerPort:                   getEnv("SERVER_PORT", "8081"),
		ServerHost:                   getEnv("SERVER_HOST", "0.0.0.0"),
		ServerReadTimeout:            getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
		ServerWriteTimeout:           getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		ServerIdleTimeout:            getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		CORSAllowedOrigins:           strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "*"), ","),
		DatabaseURL:                  os.Getenv("DATABASE_URL"),
		DatabaseMaxConns:             int32(getEnvInt("DATABASE_MAX_CONNS", 50)),
		DatabaseMinConns:             int32(getEnvInt("DATABASE_MIN_CONNS", 10)),
		DatabaseMaxConnLifetime:      getEnvDuration("DATABASE_MAX_CONN_LIFETIME", time.Hour),
		DatabaseMaxConnIdleTime:      getEnvDuration("DATABASE_MAX_CONN_IDLE_TIME", 30*time.Minute),
		IngestionBatchMaxSize:        getEnvInt("INGESTION_BATCH_MAX_SIZE", 5000),
		IngestionWorkers:             getEnvInt("INGESTION_WORKERS", 8),
		IngestionQueueBuffer:         getEnvInt("INGESTION_QUEUE_BUFFER", 50000),
		OrganizerJWTSecret:           getEnv("ORGANIZER_JWT_SECRET", "campus-monitoring-secure-demo-secret-key-32bytes-min"),
		OrganizerAdminEmail:          getEnv("ORGANIZER_ADMIN_EMAIL", "organizer@campus.edu"),
		OrganizerAdminPassword:       getEnv("ORGANIZER_ADMIN_PASSWORD", "CampusOrganizerSecure2026!"),
		OrganizerViewerEmail:         getEnv("ORGANIZER_VIEWER_EMAIL", "viewer@campus.edu"),
		OrganizerViewerPassword:      getEnv("ORGANIZER_VIEWER_PASSWORD", "CampusViewerDemo2026!"),
		SensorKeySalt:                getEnv("SENSOR_KEY_SALT", "campus-sensor-key-salt-2026-gdg"),
		SensorOfflineThreshold:       getEnvDuration("SENSOR_OFFLINE_THRESHOLD", 10*time.Minute),
		AlertWindowEvaluationInterval: getEnvDuration("ALERT_WINDOW_EVALUATION_INTERVAL", 15*time.Second),
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
