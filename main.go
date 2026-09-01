package main

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"nimbus/handler"
	nimbusMiddleware "nimbus/middleware"
)

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("invalid %s value %q, using default %d", key, value, fallback)
		return fallback
	}

	return parsed
}

func main() {
	// --------------------------------------------------
	// Configuration
	// --------------------------------------------------

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL is not set")
	}

	baseURL := getEnv("BASE_URL", "http://localhost:8080")
	extensionOrigin := os.Getenv("EXTENSION_ORIGIN")

	if extensionOrigin == "" {
		log.Fatal("EXTENSION_ORIGIN is not set")
	}

	rateLimitRequests := getEnvInt("RATE_LIMIT_REQUESTS", 10)
	rateLimitWindowSeconds := getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60)

	if rateLimitRequests <= 0 {
		log.Fatal("RATE_LIMIT_REQUESTS must be greater than 0")
	}

	if rateLimitWindowSeconds <= 0 {
		log.Fatal("RATE_LIMIT_WINDOW_SECONDS must be greater than 0")
	}

	// Validate BASE_URL during startup.
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" {
		log.Fatal("invalid BASE_URL")
	}

	// --------------------------------------------------
	// PostgreSQL
	// --------------------------------------------------

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("database connection failed:", err)
	}

	// --------------------------------------------------
	// Redis
	// --------------------------------------------------

	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("invalid REDIS_URL:", err)
	}

	rdb := redis.NewClient(redisOptions)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis connection failed:", err)
	}

	defer db.Close()
	defer rdb.Close()

	// --------------------------------------------------
	// Handler
	// --------------------------------------------------

	h := &handler.Handler{
		DB:                db,
		RDB:               rdb,
		BaseURL:           baseURL,
		RateLimitRequests: int64(rateLimitRequests),
		RateLimitWindow:   time.Duration(rateLimitWindowSeconds) * time.Second,
	}

	// --------------------------------------------------
	// Router
	// --------------------------------------------------

	r := gin.Default()

	// CORS is scoped to the configured Chrome extension.
	r.Use(nimbusMiddleware.CORS(extensionOrigin))

	r.POST("/shorten", h.Shorten)
	r.GET("/:code", h.Redirect)

	log.Println("Server running on: 8080")
	log.Println("Base URL:", baseURL)
	log.Println("Extension origin:", extensionOrigin)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
