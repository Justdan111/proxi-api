package config

import (
    "log"
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type Config struct {
    AppEnv          string
    Port            string
    MongoURI        string
    MongoDBName     string
    JWTSecret       string
    JWTExpiryHours  int
    AllowedOrigins  string
}

func Load() *Config {
    // Load .env file (ignore error in production — env vars will be set directly)
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }

    jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "72"))

    return &Config{
        AppEnv:         getEnv("APP_ENV", "development"),
        Port:           getEnv("PORT", "8080"),
        MongoURI:       getEnv("MONGODB_URI", "mongodb://localhost:27017"),
        MongoDBName:    getEnv("MONGODB_NAME", "proxi"),
        JWTSecret:      getEnv("JWT_SECRET", "change-me"),
        JWTExpiryHours: jwtExpiry,
        AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
    }
}

func getEnv(key, fallback string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return fallback
}