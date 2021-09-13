package route

import (
	"github.com/gin-gonic/gin"
	"goms"
	"goms/example/services/user/api"
)

func Route(r *gin.Engine) {
	r.Use(goms.CheckCallSafeMiddleware())
	r.POST("login", api.Login)
	r.POST("exists", api.Exists)
}
