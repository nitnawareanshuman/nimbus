package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"

	"nimbus/handler"
)

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func requiredEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		log.Fatalf("%s is not set", key)
	}

	return value
}

func main() {
	// PostgreSQL
	dbURL := requiredEnv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("database connection failed:", err)
	}

	// Redis / Valkey
	redisURL := requiredEnv("REDIS_URL")

	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("invalid REDIS_URL:", err)
	}

	rdb := redis.NewClient(redisOptions)
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis connection failed:", err)
	}

	// Application configuration
	baseURL := strings.TrimRight(
		getEnv("BASE_URL", "http://localhost:8080"),
		"/",
	)

	allowedOrigin := getEnv(
		"ALLOWED_ORIGIN",
		"*",
	)

	// Handler
	h := &handler.Handler{
		DB:            db,
		RDB:           rdb,
		BaseURL:       baseURL,
		AllowedOrigin: allowedOrigin,
	}

	// Router
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if allowedOrigin == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin == allowedOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header(
			"Access-Control-Allow-Methods",
			"GET, POST, OPTIONS",
		)

		c.Header(
			"Access-Control-Allow-Headers",
			"Content-Type",
		)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API
	r.POST("/shorten", h.Shorten)

	// Short URL redirect
	r.GET("/:code", h.Redirect)

	port := getEnv("PORT", "8080")

	log.Printf("Nimbus server running on port %s", port)
	log.Printf("Base URL: %s", baseURL)

	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal(err)
	}
}