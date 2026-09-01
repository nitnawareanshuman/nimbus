package handler

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"nimbus/service"
)

const (
	maxURLLength = 2048
)

type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

type ShortenResponse struct {
	Code      string `json:"code"`
	ShortURL  string `json:"short_url"`
	CreatedAt string `json:"created_at"`
}

type Handler struct {
	DB  *sql.DB
	RDB *redis.Client

	BaseURL string

	RateLimitRequests int64
	RateLimitWindow   time.Duration
}

func (h *Handler) Shorten(c *gin.Context) {
	// --------------------------------------------------
	// Rate limiting
	// --------------------------------------------------

	clientIP := c.ClientIP()
	rateLimitKey := "rate_limit:shorten:" + clientIP

	ctx := c.Request.Context()

	count, err := h.RDB.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		log.Printf("rate limit redis error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "rate limiter unavailable",
		})
		return
	}

	if count == 1 {
		if err := h.RDB.Expire(
			ctx,
			rateLimitKey,
			h.RateLimitWindow,
		).Err(); err != nil {
			log.Printf("rate limit expiry error: %v", err)
		}
	}

	if count > h.RateLimitRequests {
		ttl, err := h.RDB.TTL(ctx, rateLimitKey).Result()

		if err != nil || ttl < 0 {
			ttl = h.RateLimitWindow
		}

		c.Header(
			"Retry-After",
			strings.TrimSuffix(ttl.Round(time.Second).String(), "0s"),
		)

		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "rate limit exceeded",
			"retry_after": int(ttl.Seconds()),
		})
		return
	}

	// --------------------------------------------------
	// Parse JSON
	// --------------------------------------------------

	var req ShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "request body must contain a valid url",
		})
		return
	}

	req.URL = strings.TrimSpace(req.URL)

	// --------------------------------------------------
	// Basic input validation
	// --------------------------------------------------

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "url cannot be empty",
		})
		return
	}

	if len(req.URL) > maxURLLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "url is too long",
		})
		return
	}

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid URL",
		})
		return
	}

	// Only HTTP and HTTPS URLs are allowed.
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only http and https URLs are allowed",
		})
		return
	}

	if parsedURL.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL must contain a host",
		})
		return
	}

	// Reject URLs containing embedded credentials.
	if parsedURL.User != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URLs containing credentials are not allowed",
		})
		return
	}

	// --------------------------------------------------
	// Generate unique short code
	// --------------------------------------------------

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

	// --------------------------------------------------
	// Insert into PostgreSQL
	// --------------------------------------------------

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

	// --------------------------------------------------
	// Write-through to Redis
	// --------------------------------------------------

	err = h.RDB.Set(
		ctx,
		code,
		req.URL,
		0,
	).Err()

	if err != nil {
		// PostgreSQL already contains the record.
		// Redis is only a cache, so don't fail the request.
		log.Printf("redis cache set failed: %v", err)
	}

	// --------------------------------------------------
	// Response
	// --------------------------------------------------

	shortURL := strings.TrimRight(h.BaseURL, "/") + "/" + code

	c.JSON(http.StatusCreated, ShortenResponse{
		Code:      code,
		ShortURL:  shortURL,
		CreatedAt: createdAt,
	})
}
