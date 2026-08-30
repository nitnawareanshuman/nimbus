package handler

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")
	ctx := c.Request.Context()

	var (
		codeID     int
		originalURL string
	)

	// 1. Check Redis first
	originalURL, err := h.RDB.Get(ctx, code).Result()

	if err == nil {
		// Redis HIT
		log.Println("Redis HIT:", code)

		// We still need codeID to record the click.
		err = h.DB.QueryRow(
			`SELECT id FROM codes WHERE short_code = $1`,
			code,
		).Scan(&codeID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
			})
			return
		}

	} else if err == redis.Nil {
		// Redis MISS
		log.Println("Redis MISS:", code)

		// 2. Fall back to PostgreSQL
		err = h.DB.QueryRow(
			`SELECT id, original_url
			FROM codes
			WHERE short_code = $1`,
			code,
		).Scan(&codeID, &originalURL)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "short code not found",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
			})
			return
		}

		// 3. Repopulate Redis
		err = h.RDB.Set(
			ctx,
			code,
			originalURL,
			0,
		).Err()

		if err != nil {
			log.Printf("redis cache set failed: %v", err)
		}

	} else {
		// Actual Redis error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "redis error",
		})
		return
	}

	// 4. Record the click
	_, err = h.DB.Exec(
		`INSERT INTO clicks (code_id, clicked_at, reference)
		VALUES ($1, NOW(), $2)`,
		codeID,
		c.GetHeader("Referer"),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record click",
		})
		return
	}

	// 5. Redirect
	c.Redirect(http.StatusFound, originalURL)
}