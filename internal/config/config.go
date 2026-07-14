package config

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all the operational parameters for the Xomoi-Core edge node.
type Config struct {
	DBPath           string
	APIPort          string
	MQTTPort         string
	SignalingURL     string
	STUNServers      []string
	IngestionWorkers int
	FactorySecret    string
	FlushIntervalSec int
	MemoryLimitMB    int
	LogFormat        string // "json" or "text"
}

// Load parses environment variables and applies sensible defaults.
func Load() *Config {
	// Attempt to load .env file and OVERRIDE any stale global OS variables
	if err := godotenv.Overload(); err != nil && !os.IsNotExist(err) {
		log.Println("No .env file found or unable to load. Falling back to environment variables.")
	}

	return &Config{
		DBPath:           getEnv("XOMOI_DB_PATH", "xomoi.db"),
		APIPort:          getEnv("XOMOI_API_PORT", "8085"),
		MQTTPort:         getEnv("XOMOI_MQTT_PORT", "1883"),
		SignalingURL:     getEnv("XOMOI_SIGNALING_URL", "ws://localhost:8086/ws"),
		STUNServers:      getEnvSlice("XOMOI_STUN_SERVERS", []string{"stun:stun.l.google.com:19302"}),
		IngestionWorkers: getEnvInt("XOMOI_INGESTION_WORKERS", runtime.NumCPU()),
		FactorySecret:    getEnv("XOMOI_FACTORY_SECRET", "xomoi-factory-secret"),
		FlushIntervalSec: getEnvInt("XOMOI_FLUSH_INTERVAL_SEC", 30),
		MemoryLimitMB:    getEnvInt("XOMOI_MEMORY_LIMIT_MB", 250),
		LogFormat:        getEnv("XOMOI_LOG_FORMAT", "json"),
	}
}

// Helpers
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		// e.g. "stun:a.com,stun:b.com"
		return strings.Split(value, ",")
	}
	return fallback
}
