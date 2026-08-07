package config

import (
	"fmt"
	"os"
)

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
