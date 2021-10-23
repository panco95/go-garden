package auth

import "github.com/gin-gonic/gin"

// Auth Customize the auth middleware
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before logic
		c.Next()
		// after logic
	}
}
