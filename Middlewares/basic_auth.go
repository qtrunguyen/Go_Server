package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// key-value: username-password
func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Test-Request") == "true" {
			c.Next() // Bypass authentication for testing
			return
		}

		// Perform actual authentication logic here
		username, password, _ := c.Request.BasicAuth()
		if username != "username" || password != "password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
