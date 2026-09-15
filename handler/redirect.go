package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nimbus/service"
)

func (h *Handler) Redirect(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "code is required",
		})
		return
	}

	targetURL, err := service.GetURL(
		c.Request.Context(),
		h.DB,
		h.RDB,
		code,
	)

	if err != nil {
		if errors.Is(err, service.ErrCodeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "short URL not found",
			})
			return
		}

		log.Printf("GetURL failed for short code %q: %v", code, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to resolve short URL",
		})
		return
	}

	c.Redirect(http.StatusFound, targetURL)
}
