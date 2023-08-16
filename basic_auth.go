package middlewares

import "github.com/gin-gonic/gin"

//key-value: username-password
func BasicAuth() gin.HandlerFunc {
	return gin.BasicAuthForRealm(gin.Accounts{
		"username": "password",
	}, "")
}
