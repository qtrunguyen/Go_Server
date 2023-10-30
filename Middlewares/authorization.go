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

// func VerifyJWT(tokenString, secretKey string) (*jwt.Token, error) {
//     token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//         // Make sure to use the same secret key that you used for signing the token
//         return []byte(secretKey), nil
//     })

//     if err != nil {
//         return nil, err
//     }

//     return token, nil
// }

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Get the token from the request header
// 		tokenString := c.GetHeader("Authorization")
// 		if tokenString == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization token is missing"})
// 			c.Abort()
// 			return
// 		}

// 		// Verify and decode the token
// 		token, err := VerifyJWT(tokenString, "your-secret-key")
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
// 			c.Abort()
// 			return
// 		}

// 		// Token is valid, extract user information from claims
// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok {
// 			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
// 			c.Abort()
// 			return
// 		}

// 		userID, ok := claims["userID"].(float64)
// 		if !ok {
// 			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
// 			c.Abort()
// 			return
// 		}

// 		// Now you have the user's ID from the token. You can use it for authentication and authorization.

// 		// Continue with the request
// 		c.Next()
// 	}
// }
