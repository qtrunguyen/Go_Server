package middlewares

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// key-value: username-password
// func BasicAuth() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		if c.GetHeader("X-Test-Request") == "true" {
// 			c.Next() // Bypass authentication for testing
// 			return
// 		}

// 		// Perform actual authentication logic here
// 		username, password, _ := c.Request.BasicAuth()
// 		if username != "username" || password != "password" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 			c.Abort()
// 			return
// 		}

// 		c.Next()
// 	}
// }

func AuthMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		authHeader := context.GetHeader("Authorization")
		if authHeader == "" {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			context.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			context.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Parse and validate the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			const secretKey = "vcsbackend"
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			context.Abort()
			return
		}

		// Extract user information from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token claims"})
			context.Abort()
			return
		}

		// Attach user information to the context for later use
		context.Set("user", claims)

		// Continue processing the request
		context.Next()
	}
}
