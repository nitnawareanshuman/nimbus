package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"nimbus/handler"
)

func main() {
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

	r := gin.Default()

	h := &handler.Handler{
		DB: db,
	}

	r.POST("/shorten", h.Shorten)
	r.GET("/:code", h.Redirect)

	log.Println("Server running on: 8080")
	r.Run(":8080")
}
