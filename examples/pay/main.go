package main

import (
	"encoding/json"
	"garden"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// server init
	garden.Init()
	// server run
	garden.Run(Route, nil)
}

func Route(r *gin.Engine) {
	r.Use(garden.CheckCallSafeMiddleware())
	r.POST("order", Order)
}

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

	// call [user] service example
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

	// Parse to get the data returned by the user service, and if the user exists (exists=true), then the order is successful
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

// VOrder order api parameter validator
type VOrder struct {
	Username string `form:"username" binding:"required,max=20,min=1" `
}
