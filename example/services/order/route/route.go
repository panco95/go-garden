package route

import (
	"github.com/gin-gonic/gin"
	"goms"
	"goms/example/services/order/api"
)

func Route(r *gin.Engine) {
	r.Use(goms.CheckCallSafeMiddleware())
	r.POST("submit", api.Submit)
}
