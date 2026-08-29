package handler

import (
	"database/sql"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"nimbus/service"
)

type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

type ShortenResponse struct {
	Code string `json:"code"`
	ShortURL string `json:"short_code"`
	CreatedAt string `json:"created_at"`
}

type Handler struct {
	DB *sql.DB
}

func (h *Handler) Shorten (c *gin.Context) {
	var req ShortenRequest

	// Parse JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H {
			"error": "url is required",
		})
		return
	}

	// Validate URL
	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid URL",
		})
		return
	}

	// Generate unique short code
	var code string

	for {
		code = service.GenerateCode()

		var exists bool
		err := h.DB.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM codes WHERE short_code = $1)",
			code,
		).Scan(&exists)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
			})
			return
		}

		if !exists {
			break
		}
	}

	// Insert into database
	var createdAt string

	err = h.DB.QueryRow(`
		INSERT INTO codes (short_code, original_url)
		VALUES ($1, $2)
		RETURNING created_at
	`, code, req.URL).Scan(&createdAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create short URL",
		})
		return
	}

	c.JSON(http.StatusCreated, ShortenResponse{
		Code:      code,
		ShortURL:  "http://localhost:8080/" + code,
		CreatedAt: createdAt,
	})
}