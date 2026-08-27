package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisAddr      string
	GRPCSenderAddr string
	GRPCUseTLS     bool

	GRPCCertFile string
	GRPCKeyFile  string
	GRPCCAFile   string

	APIPort string
}

func LoadConfig() Config {
	_ = godotenv.Load()

	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "sns_db"),

		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		GRPCSenderAddr: getEnv("GRPC_SENDER_ADDR", "localhost:50051"),
		GRPCUseTLS:     getEnv("GRPC_USE_TLS", "false") == "true",

		GRPCCertFile: getEnv("GRPC_CERT_FILE", ""),
		GRPCKeyFile:  getEnv("GRPC_KEY_FILE", ""),
		GRPCCAFile:   getEnv("GRPC_CA_FILE", ""),

		APIPort: getEnv("API_PORT", "8080"),
	}
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
