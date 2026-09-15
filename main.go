package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

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

func ensureSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS codes (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(20) NOT NULL UNIQUE,
			original_url TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func connectRedis() *redis.Client {
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if redisURL == "" {
		log.Println("REDIS_URL is not set; running without Redis cache")
		return nil
	}

	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("invalid REDIS_URL; running without Redis cache: %v", err)
		return nil
	}

	rdb := redis.NewClient(redisOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Redis unavailable; running without cache: %v", err)
		_ = rdb.Close()
		return nil
	}

	log.Println("Redis connected")
	return rdb
}

func main() {
	// PostgreSQL is the source of truth and is required.
	dbURL := requiredEnv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("failed to open database: ", err)
	}
	defer db.Close()

	dbCtx, cancelDB := context.WithTimeout(context.Background(), 10*time.Second)
	if err := db.PingContext(dbCtx); err != nil {
		cancelDB()
		log.Fatal("database connection failed: ", err)
	}
	cancelDB()
	log.Println("PostgreSQL connected")

	// Keep production deployments usable even when the migration container is not run.
	schemaCtx, cancelSchema := context.WithTimeout(context.Background(), 10*time.Second)
	if err := ensureSchema(schemaCtx, db); err != nil {
		cancelSchema()
		log.Fatal("failed to ensure database schema: ", err)
	}
	cancelSchema()

	// Redis is only a cache. The API remains functional if it is absent or temporarily down.
	rdb := connectRedis()
	if rdb != nil {
		defer rdb.Close()
	}

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
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Printf("health check failed: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	r.POST("/shorten", h.Shorten)
	r.GET("/:code", h.Redirect)

	// Render automatically provides PORT. Local Docker/default development uses 8080.
	port := getEnv("PORT", "8080")

	log.Printf("Nimbus server running on port %s", port)
	log.Printf("Base URL: %s", baseURL)

	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal(err)
	}
}
