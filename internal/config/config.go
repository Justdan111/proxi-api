package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv         string
	Port           string
	MongoURI       string
	MongoDBName    string
	JWTSecret      string
	JWTExpiryHours int
	AllowedOrigins string
}

func Load() *Config {
	// Load .env file if it exists, but don't fail if it's missing
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	appEnv := strings.ToLower(getEnv("APP_ENV", "development"))
	isDevLike := appEnv == "development" || appEnv == "dev" || appEnv == "local" || appEnv == "test"

	mongoURI := getEnv("MONGODB_URI", "")
	if mongoURI == "" {
		if isDevLike {
			mongoURI = "mongodb://localhost:27017"
			log.Println("MONGODB_URI not set, using local default mongodb://localhost:27017")
		} else {
			log.Fatal("MONGODB_URI is required in non-development environments")
		}
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		if isDevLike {
			jwtSecret = "change-me"
			log.Println("JWT_SECRET not set, using development fallback")
		} else {
			log.Fatal("JWT_SECRET is required in non-development environments")
		}
	}

	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "72"))

	return &Config{
		AppEnv:         appEnv,
		Port:           getEnv("PORT", "8080"),
		MongoURI:       mongoURI,
		MongoDBName:    getEnv("MONGODB_NAME", "proxi"),
		JWTSecret:      jwtSecret,
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
