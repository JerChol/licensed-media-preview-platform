package config

import (
	"fmt"
	"os"
)

// RedisConfig holds everything needed to connect to Redis.
type RedisConfig struct {
	Addr     string
	Password string
	UseTLS   bool
}

// LoadRedisConfig reads Redis connection settings from environment variables, falling back to local dev defaults
func LoadRedisConfig() RedisConfig {
	addr := getEnvOrDefault("REDIS_ADDR", "localhost:6379")
	password := getEnvOrDefault("REDIS_PASSWORD", "")
	useTLS := os.Getenv("REDIS_TLS") == "true"

	return RedisConfig{
		Addr:     addr,
		Password: password,
		UseTLS:   useTLS,
	}
}

// S3Config holds settings for connecting to S3-compatible storage.
type S3Config struct {
	BucketName string
	Region     string
}

// LoadS3Config reads S3 settings from environment variables.
func LoadS3Config() S3Config {
	return S3Config{
		BucketName: getEnvOrDefault("S3_BUCKET_NAME", ""),
		Region:     getEnvOrDefault("AWS_REGION", "us-east-2"),
	}
}

// PostgresDSN returns the connection string for Postgres, reading from an environment variable if set, falling back to a local dev default otherwise
func PostgresDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "pass")
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	dbName := getEnvOrDefault("DB_NAME", "lmpp_dev")

	if password == "" {
		return fmt.Sprintf("postgres://%s@%s:%s/%s", user, host, port, dbName)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbName)
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
