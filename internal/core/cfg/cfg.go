package cfg

import (
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

// Config holds the configuration values for the application.
type Config struct {
	AppPort                   string  // Port for the application server.
	GrpcPort                  string  // Port for the gRPC server.
	LotofSampleSvcGrpcAddress string  // Address for the Lotof Sample Service gRPC.
	Environment               string  // Environment (e.g., dev, staging, prod).
	OtlpEndpoint              string  // OpenTelemetry collector endpoint.
	TraceSampleRatio          float64 // Trace sample ratio.
	LogLevel                  string  // Log level (debug, info, warn, error).
}

var (
	once     sync.Once
	instance *Config
)

// Inst initializes the configuration instance if it hasn't been already and returns it.
func Inst() *Config {
	once.Do(func() {
		// Load environment variables from .env file if it exists.
		err := godotenv.Load()
		if err != nil {
			slog.Warn("No .env file found, loading from OS environment variables.")
		}

		// Initialize the Config instance with environment variables or default values.
		instance = &Config{
			AppPort:                   getEnv("APP_PORT", "8080"),
			GrpcPort:                  getEnv("GRPC_PORT", "50051"),
			LotofSampleSvcGrpcAddress: getEnv("LOTOF_SAMPLE_SVC_GRPC_ADDRESS", "localhost:50051"),
			Environment:               getEnv("ENVIRONMENT", "development"),
			OtlpEndpoint:              getEnv("OTLP_ENDPOINT", "localhost:4317"),
			TraceSampleRatio:          getEnvFloat("TRACE_SAMPLE_RATIO", 0.1),
			LogLevel:                  getEnv("LOG_LEVEL", "info"),
		}
	})
	return instance
}

// getEnv retrieves the value of the environment variable named by the key or returns the default value if the variable is not present.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvFloat retrieves the float value of the environment variable or returns the default value.
func getEnvFloat(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		f, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return f
		}
	}
	return defaultValue
}
