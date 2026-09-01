package handler

import (
	"database/sql"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"nimbus/service"
)

type Handler struct {
	DB            *sql.DB
	RDB           *redis.Client
	BaseURL       string
	AllowedOrigin string
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code      string `json:"code"`
	ShortURL  string `json:"short_url"`
	TargetURL string `json:"target_url"`
}

func (h *Handler) Shorten(c *gin.Context) {
	var req shortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body",
		})
		return
	}

	req.URL = strings.TrimSpace(req.URL)

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "url is required",
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

	if parsedURL.Scheme != "http" &&
		parsedURL.Scheme != "https" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL must use http or https",
		})
		return
	}

	if parsedURL.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL must contain a valid host",
		})
		return
	}

	code, err := service.CreateCode(
		c.Request.Context(),
		h.DB,
		h.RDB,
		req.URL,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create short URL",
		})
		return
	}

	shortURL := strings.TrimRight(h.BaseURL, "/") + "/" + code

	c.JSON(http.StatusCreated, shortenResponse{
		Code:      code,
		ShortURL:  shortURL,
		TargetURL: req.URL,
	})
}