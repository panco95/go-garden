package main

import (
	"encoding/json"
	"garden"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// 服务初始化
	garden.Init()
	// 服务启动
	garden.Run(Route, nil)
}

// Route gin路由
func Route(r *gin.Engine) {
	r.Use(garden.CheckCallSafeMiddleware())
	r.POST("order", Order)
}

// Order 下单接口
// @param username 用户名
func Order(c *gin.Context) {
	span, err := garden.GetSpan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		garden.Logger.Errorf("[%s] %s", "GetSpan", err)
		return
	}

	var Validate VOrder
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(http.StatusOK, garden.ApiResponse(1000, "非法参数", nil))
		return
	}
	username := c.DefaultPostForm("username", "")

	// 调用user服务示例
	service := "user"
	action := "exists"
	result, err := garden.CallService(span, service, action, &garden.Request{
		Method: "POST",
		Body: garden.Any{
			"username": username,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		garden.Logger.Errorf("[%s] %s", "CallService", err)
		span.SetTag("CallService", err)
		return
	}
	var res garden.Any
	err = json.Unmarshal([]byte(result), &res)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		garden.Logger.Errorf("[%s] %s", "JsonUnmarshall", err)
		span.SetTag("JsonUnmarshall", err)
	}

	// 解析获取user服务返回的数据，如果用户存在(exists=true)，那么下单成功
	data := res["data"].(map[string]interface{})
	exists := data["exists"].(bool)
	if !exists {
		c.JSON(http.StatusOK, garden.ApiResponse(1000, "下单失败", nil))
		return
	}
	orderId := garden.NewUuid()
	c.JSON(http.StatusOK, garden.ApiResponse(0, "下单成功", garden.Any{
		"orderId": orderId,
	}))
}

// VOrder 下单接口参数验证器
type VOrder struct {
	Username string `form:"username" binding:"required,max=20,min=1" `
}
