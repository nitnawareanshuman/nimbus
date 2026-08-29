package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")

	var (
		codeID int
		originalURL string
	)

	// Find the original URL
	err := h.DB.QueryRow(
		`SELECT id, original_url FROM codes WHERE short_code = $1`,
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

	// Record the click
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

	// Redirect to original URL
	c.Redirect(http.StatusFound, originalURL)
}