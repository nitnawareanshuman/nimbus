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
		log.Fatal("failed to open database: ", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("database connection failed: ", err)
	}

	log.Println("PostgreSQL connected")

	// Redis
	redisURL := requiredEnv("REDIS_URL")

	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("invalid REDIS_URL: ", err)
	}

	rdb := redis.NewClient(redisOptions)
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis connection failed: ", err)
	}

	log.Println("Redis connected")

	// Public URL used when creating shortened links
	baseURL := strings.TrimRight(
		getEnv("BASE_URL", "http://localhost:8080"),
		"/",
	)

	h := &handler.Handler{
		DB:      db,
		RDB:     rdb,
		BaseURL: baseURL,
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.POST("/shorten", h.Shorten)
	r.GET("/:code", h.Redirect)

	// Render automatically provides PORT.
	// Local Docker/default development uses 8080.
	port := getEnv("PORT", "8080")

	log.Printf("Nimbus server running on port %s", port)
	log.Printf("Base URL: %s", baseURL)

	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal(err)
	}
}