package main

import (
	"github.com/panco95/go-garden"
	"github.com/gin-gonic/gin"
)

func main() {
	// server init
	garden.Init()
	// server run
	garden.Run(garden.GatewayRoute, Auth)
}

// Auth Customize the global middleware
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before logic
		c.Next()
		// after logic
	}
}
