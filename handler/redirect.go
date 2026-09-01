package handler

import (
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
		c.JSON(http.StatusNotFound, gin.H{
			"error": "short URL not found",
		})
		return
	}

	c.Redirect(
		http.StatusFound,
		targetURL,
	)
}