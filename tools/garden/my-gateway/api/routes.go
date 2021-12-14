package api

import (
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine) {
	r.POST("test", Test)
}
