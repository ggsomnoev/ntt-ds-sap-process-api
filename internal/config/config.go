package config

import (
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	AppEnv        string
	APIPort       string
	ProcessCfgDir string

	WebAPICertFile string
	WebAPIKeyFile  string

	DBConnectionURL   string
	DBMaxConnLifetime time.Duration
	DBMaxConnIdleTime time.Duration
	DBHealthCheck     time.Duration
	DBMinConns        int32
	DBMaxConns        int32

	RabbitMQConnURL  string
	RabbitMQQueue    string
	RabbitMQCAFile   string
	RabbitMQCertFile string
	RabbitMQKeyFile  string
}

func Load() (*Config, error) {
	return &Config{
		AppEnv:            getEnv("APP_ENV", "local"),
		APIPort:           getEnv("API_PORT", "8080"),
		ProcessCfgDir:     getEnv("PROCESS_CFG_DIR", "/app/process_configs"),
		WebAPICertFile:    getEnv("WEBAPI_TLS_CERT_FILE", ""),
		WebAPIKeyFile:     getEnv("WEBAPI_TLS_KEY_FILE", ""),
		DBConnectionURL:   getEnv("DB_CONNECTION_URL", "postgres://user:pass@processdb:5432/processdb"),
		DBMaxConnLifetime: getDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute),
		DBMaxConnIdleTime: getDuration("DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
		DBHealthCheck:     getDuration("DB_HEALTH_CHECK_PERIOD", 1*time.Minute),
		DBMinConns:        getInt32("DB_MIN_CONNS", 1),
		DBMaxConns:        getInt32("DB_MAX_CONNS", 5),
		RabbitMQConnURL:   getEnv("RABBITMQ_CONN_URL", ""),
		RabbitMQQueue:     getEnv("RABBITMQ_QUEUE", ""),
		RabbitMQCAFile:    getEnv("RABBITMQ_CA_FILE", ""),
		RabbitMQCertFile:  getEnv("RABBITMQ_CERT_FILE", ""),
		RabbitMQKeyFile:   getEnv("RABBITMQ_KEY_FILE", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		d, err := time.ParseDuration(val)
		if err == nil {
			return d
		}
	}
	return fallback
}

func getInt32(key string, fallback int32) int32 {
	if val := os.Getenv(key); val != "" {
		i, err := strconv.Atoi(val)
		if err == nil {
			return int32(i)
		}
	}
	return fallback
}
