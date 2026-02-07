package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BearerAuth(apiKey string) gin.HandlerFunc {
	const prefix = "Bearer "

	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing Authorization header",
			})
			return
		}

		if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid Authorization format",
			})
			return
		}

		if auth[len(prefix):] != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		c.Next()
	}
}
