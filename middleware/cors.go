package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Only allow the configured extension origin.
		if origin != "" && origin == allowedOrigin {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			c.Header("Vary", "Origin")
		}

		// Handle browser preflight requests.
		if c.Request.Method == http.MethodOptions {
			if origin == allowedOrigin {
				c.Status(http.StatusNoContent)
				return
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error": "origin not allowed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
