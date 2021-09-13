package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在这里写网关统一鉴权逻辑
		c.Next()
		log.Printf(c.Request.RequestURI)
	}
}