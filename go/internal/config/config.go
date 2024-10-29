package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
    Port        string
    DBHost      string
    DBPort      string
    DBUser      string
    DBName      string
    DBPassword  string
    DatabaseURL string
    RedisURL    string
    JWTSecret   string
    Environment string
}

func LoadConfig() *Config {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: .env file not loaded: %v", err)
    }

    // Get environment variables with defaults
    dbHost := getEnvOrDefault("DB_HOST", "localhost")
    dbPort := getEnvOrDefault("DB_PORT", "5432")
    dbUser := getEnvOrDefault("DB_USER", "user")
    dbPass := getEnvOrDefault("DB_PASSWORD", "password")
    dbName := getEnvOrDefault("DB_NAME", "chatdb")
    
    // Construct database URL
    dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
        dbUser, dbPass, dbHost, dbPort, dbName)

    return &Config{
        Port:        getEnvOrDefault("PORT", ":8080"),
        DatabaseURL: dbURL,
        RedisURL:    fmt.Sprintf("%s:%s", 
            getEnvOrDefault("REDIS_HOST", "localhost"),
            getEnvOrDefault("REDIS_PORT", "6379")),
        JWTSecret:   getEnvOrDefault("JWT_SECRET", "your_development_secret_key_123"),
        Environment: getEnvOrDefault("ENV", "development"),
    }
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}