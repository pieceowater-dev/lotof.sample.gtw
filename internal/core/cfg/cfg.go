package cfg

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"sync"
)

type Config struct {
	AppPort                   string
	GrpcPort                  string
	LotofSampleSvcGrpcAddress string
}

var (
	once     sync.Once
	instance *Config
)

func Inst() *Config {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			fmt.Println("No .env file found, loading from OS environment variables.")
		}

		instance = &Config{
			AppPort:                   getEnv("APP_PORT", "8080"),
			GrpcPort:                  getEnv("GRPC_PORT", "50051"),
			LotofSampleSvcGrpcAddress: getEnv("LOTOF_SAMPLE_SVC_GRPC_ADDRESS", "localhost:50051"),
		}
	})
	return instance
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
