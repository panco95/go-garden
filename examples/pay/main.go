package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"math/rand"
	"time"
)

var service *core.Garden

func main() {
	service = core.New()
	service.Run(route, nil)
}

func route(r *gin.Engine) {
	r.Use(service.CheckCallSafeMiddleware())
	r.POST("order", order)
}

func order(c *gin.Context) {
	span, err := core.GetSpan(c)
	if err != nil {
		c.JSON(500, nil)
		service.Log(core.ErrorLevel, "GetSpan", err)
		return
	}

	var Validate vOrder
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(200, apiResponse(1000, "非法参数", nil))
		return
	}
	username := c.DefaultPostForm("username", "")

	// call [user] service example
	code, result, err := service.CallService(span, "user", "exists", &core.Request{
		Method: "POST",
		Body: core.MapData{
			"username": username,
		},
	})
	if err != nil {
		c.JSON(code, nil)
		service.Log(core.ErrorLevel, "CallService", err)
		span.SetTag("CallService", err)
		return
	}

	var res core.MapData
	err = json.Unmarshal([]byte(result), &res)
	if err != nil {
		c.JSON(500, nil)
		service.Log(core.ErrorLevel, "JsonUnmarshall", err)
		span.SetTag("JsonUnmarshall", err)
	}

	// Parse to get the data returned by the user service, and if the user exists (exists=true), then the order is successful
	data := res["data"].(map[string]interface{})
	exists := data["exists"].(bool)
	if !exists {
		c.JSON(code, apiResponse(1000, "下单失败", nil))
		return
	}
	orderId := fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
	c.JSON(code, apiResponse(0, "下单成功", core.MapData{
		"orderId": orderId,
	}))
}

type vOrder struct {
	Username string `form:"username" binding:"required,max=20,min=1" `
}

// apiResponse format response
func apiResponse(code int, msg string, data interface{}) core.MapData {
	return core.MapData{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
