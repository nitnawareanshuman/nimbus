package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"

	"nimbus/handler"
)

func main() {
	// PostgreSQL
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Redis
	redisURL := os.Getenv("REDIS_URL")

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