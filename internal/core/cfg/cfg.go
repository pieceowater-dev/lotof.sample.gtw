package cfg

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"sync"
)

// Config holds the configuration values for the application.
type Config struct {
	AppPort                   string // Port for the application server.
	GrpcPort                  string // Port for the gRPC server.
	LotofSampleSvcGrpcAddress string // Address for the Lotof Sample Service gRPC.
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
			fmt.Println("No .env file found, loading from OS environment variables.")
		}

		// Initialize the Config instance with environment variables or default values.
		instance = &Config{
			AppPort:                   getEnv("APP_PORT", "8080"),
			GrpcPort:                  getEnv("GRPC_PORT", "50051"),
			LotofSampleSvcGrpcAddress: getEnv("LOTOF_SAMPLE_SVC_GRPC_ADDRESS", "localhost:50051"),
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
