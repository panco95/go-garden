package main

import (
	"garden"
	"github.com/gin-gonic/gin"
)

func main() {
	// 服务初始化
	garden.Init()
	// 服务启动
	garden.Run(garden.GatewayRoute, Auth)
}

// Auth 自定义全局中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 前置逻辑
		c.Next()
		// 后置逻辑
	}
}
