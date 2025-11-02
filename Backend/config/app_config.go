package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig holds all application configuration
type AppConfig struct {
	// API Configuration
	APIBaseURL  string
	APIKey      string
	APIEndpoint string
	OutputFile  string

	// Database Configuration
	Database DatabaseConfig

	// CockroachDB Configuration
	CockroachDB CockroachDBConfig

	// Application Settings
	AppEnv      string
	AppDebug    bool
	AppLogLevel string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	LogLevel string
}

// CockroachDBConfig holds CockroachDB-specific configuration
type CockroachDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	CertsDir string

	// Cluster Settings
	ClusterName   string
	ClusterRegion string

	// Range Configuration
	RangeSize               int64
	RangeMaxBytes           int64
	RangeMinBytes           int64
	RangeRebalanceThreshold float64

	// Replica Configuration
	NumReplicas        int
	ReplicaConstraints string
	ReplicaLeaseholder string

	// Zone Configuration
	ZoneConfig string

	// Performance Settings
	MaxConns        int
	MinConns        int
	MaxConnLifetime string
	MaxConnIdleTime string

	// Backup Configuration
	BackupEnabled   bool
	BackupSchedule  string
	BackupRetention string

	// Monitoring Configuration
	LogLevel         string
	MetricsEnabled   bool
	ProfilingEnabled bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *AppConfig {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	return &AppConfig{
		// API Configuration
		APIBaseURL:  getEnv("API_BASE_URL", "https://api.example.com"),
		APIKey:      getEnv("API_KEY", ""),
		APIEndpoint: getEnv("API_ENDPOINT", "/data"),
		OutputFile:  getEnv("OUTPUT_FILE", "extracted_data.json"),

		// Database Configuration
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "26257"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "stock_data"),
			SSLMode:  getEnv("DB_SSLMODE", "require"),
			LogLevel: getEnv("DB_LOG_LEVEL", "info"),
		},

		// CockroachDB Configuration
		CockroachDB: CockroachDBConfig{
			Host:     getEnv("COCKROACH_HOST", "localhost"),
			Port:     getEnv("COCKROACH_PORT", "26257"),
			User:     getEnv("COCKROACH_USER", "root"),
			Password: getEnv("COCKROACH_PASSWORD", ""),
			DBName:   getEnv("COCKROACH_DB_NAME", "stock_data"),
			SSLMode:  getEnv("COCKROACH_SSL_MODE", "require"),
			CertsDir: "./db_setup/certs",

			// Cluster Settings
			ClusterName:   getEnv("COCKROACH_CLUSTER_NAME", "dataextractor-secure-cluster"),
			ClusterRegion: getEnv("COCKROACH_CLUSTER_REGION", "us-east-1"),

			// Range Configuration
			RangeSize:               getEnvAsInt64("COCKROACH_RANGE_SIZE", 67108864),
			RangeMaxBytes:           getEnvAsInt64("COCKROACH_RANGE_MAX_BYTES", 134217728),
			RangeMinBytes:           getEnvAsInt64("COCKROACH_RANGE_MIN_BYTES", 33554432),
			RangeRebalanceThreshold: getEnvAsFloat64("COCKROACH_RANGE_REBALANCE_THRESHOLD", 0.05),

			// Replica Configuration
			NumReplicas:        getEnvAsInt("COCKROACH_NUM_REPLICAS", 3),
			ReplicaConstraints: getEnv("COCKROACH_REPLICA_CONSTRAINTS", ""),
			ReplicaLeaseholder: getEnv("COCKROACH_REPLICA_LEASEHOLDER", ""),

			// Zone Configuration
			ZoneConfig: getEnv("COCKROACH_ZONE_CONFIG", ""),

			// Performance Settings
			MaxConns:        getEnvAsInt("COCKROACH_MAX_CONNS", 100),
			MinConns:        getEnvAsInt("COCKROACH_MIN_CONNS", 10),
			MaxConnLifetime: getEnv("COCKROACH_MAX_CONN_LIFETIME", "1h"),
			MaxConnIdleTime: getEnv("COCKROACH_MAX_CONN_IDLE_TIME", "30m"),

			// Backup Configuration
			BackupEnabled:   getEnvAsBool("COCKROACH_BACKUP_ENABLED", false),
			BackupSchedule:  getEnv("COCKROACH_BACKUP_SCHEDULE", "0 2 * * *"),
			BackupRetention: getEnv("COCKROACH_BACKUP_RETENTION", "7d"),

			// Monitoring Configuration
			LogLevel:         getEnv("COCKROACH_LOG_LEVEL", "info"),
			MetricsEnabled:   getEnvAsBool("COCKROACH_METRICS_ENABLED", true),
			ProfilingEnabled: getEnvAsBool("COCKROACH_PROFILING_ENABLED", false),
		},

		// Application Settings
		AppEnv:      getEnv("APP_ENV", "development"),
		AppDebug:    getEnvAsBool("APP_DEBUG", true),
		AppLogLevel: getEnv("APP_LOG_LEVEL", "info"),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 gets an environment variable as an int64 with a default value
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsFloat64 gets an environment variable as a float64 with a default value
func getEnvAsFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as a boolean with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
