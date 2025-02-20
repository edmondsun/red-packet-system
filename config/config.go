package config

import (
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"red-packet-system/pkg/logger"
)

// Config stores all environment variables
type Config struct {
	ServerPort      string
	DBMaster        string
	DBSlave         string
	DBUser          string
	DBPassword      string
	DBName          string
	RedisAddr       string
	RedisPassword   string
	RedisCluster    []string
	KafkaBrokers    []string
	ZookeeperBroker string
}

// Ensure singleton pattern using `sync.Once`
var (
	configInstance *Config
	once           sync.Once
)

// LoadConfig loads environment variables (ensures singleton instance)
func LoadConfig() *Config {
	log := logger.GetLogger()

	once.Do(func() {
		// Attempt to load `.env` from multiple possible locations
		possiblePaths := []string{
			"../../.env", // For `cmd/server/` execution
			"../.env",    // For `cmd/worker/` execution
			".env",       // For execution within `red-packet-system/`
		}

		envLoaded := false
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil { // Ensure the file exists
				if err := godotenv.Load(path); err == nil {
					log.Printf(".env file loaded from: %s", path)
					envLoaded = true
					break
				}
			}
		}

		if !envLoaded {
			log.Println("No .env file found, using system environment variables")
		}

		// Load configuration from environment variables
		configInstance = &Config{
			ServerPort:      os.Getenv("SERVER_PORT"),
			DBMaster:        os.Getenv("DB_MASTER"),
			DBSlave:         os.Getenv("DB_SLAVE"),
			DBUser:          os.Getenv("DB_USER"),
			DBPassword:      os.Getenv("DB_PASSWORD"),
			DBName:          os.Getenv("DB_NAME"),
			RedisAddr:       os.Getenv("REDIS_ADDR"),
			RedisPassword:   os.Getenv("REDIS_PASSWORD"),
			RedisCluster:    strings.Split(os.Getenv("REDIS_CLUSTER_NODES"), ","),
			KafkaBrokers:    strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
			ZookeeperBroker: os.Getenv("KAFKA_ZOOKEEPER_CONNECT"),
		}

		log.Printf("Config Loaded: %+v\n", configInstance)
	})

	return configInstance
}
