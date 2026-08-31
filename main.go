package main

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"

	"nimbus/handler"
)

func main() {
	// PostgreSQL configuration
	dbURI := os.Getenv("DB_URI")
	dbPassword := os.Getenv("DB_PASSWORD")

	if dbURI == "" {
		log.Fatal("DB_URI is not set")
	}

	// Add the password from the Kubernetes Secret to the DB URI.
	dbURL, err := url.Parse(dbURI)
	if err != nil {
		log.Fatal("invalid DB_URI:", err)
	}

	username := dbURL.User.Username()
	dbURL.User = url.UserPassword(username, dbPassword)

	db, err := sql.Open("postgres", dbURL.String())
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("database connection failed:", err)
	}

	// Redis
	redisHost := os.Getenv("REDIS_HOST")

	if redisHost == "" {
		log.Fatal("REDIS_HOST is not set")
	}

	redisURL := "redis://" + redisHost

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(opt)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis connection failed:", err)
	}

	defer db.Close()
	defer rdb.Close()

	// Handler
	h := &handler.Handler{
		DB:  db,
		RDB: rdb,
	}

	// Router
	r := gin.Default()

	r.POST("/shorten", h.Shorten)
	r.GET("/:code", h.Redirect)

	log.Println("Server running on: 8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}