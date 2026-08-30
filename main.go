package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"

	"nimbus/handler"
)

func main() {
	// PostgreSQL
	db, err := sql.Open(
		"postgres",
		"postgres://nimbus:nimbuspass@localhost:5432/nimbus?sslmode=disable",
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

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