package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-ms/pkg/base/request"
	"net/http"
	"strings"
)

// 服务调用验证
func CheckCallServiceKey(c *gin.Context) {
	requestKey := c.GetHeader("Call-Service-Key")
	if strings.Compare(requestKey, viper.GetString("callServiceKey")) != 0 {
		c.JSON(http.StatusForbidden, request.MakeFailResponse())
		c.Abort()
		return
	}
}
