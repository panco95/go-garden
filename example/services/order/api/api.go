package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"goms"
	"goms/example/services/order/validate"
	"net/http"
)

// Submit 提交订单接口
// @param username 用户名
func Submit(c *gin.Context) {
	var Validate validate.Submit
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(http.StatusOK, goms.ApiResponse(1000, "非法参数", nil))
		return
	}
	username := c.DefaultPostForm("username", "")

	// 调用user服务示例
	service := "user"
	action := "exists"
	span, err := goms.GetSpan(c)
	if err != nil {
		goms.Logger.Error("get span error：" + err.Error())
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	result, err := goms.CallService(span, service, action, &goms.Request{
		Method: "POST",
		Body: goms.Any{
			"username": username,
		},
	})
	if err != nil {
		goms.Logger.Error("call service error：" + err.Error())
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	var res goms.Any
	err = json.Unmarshal([]byte(result), &res)
	if err != nil {
		goms.Logger.Error("json unmarshall error：" + err.Error())
		c.JSON(http.StatusInternalServerError, nil)
	}

	// 解析获取user服务返回的数据，如果用户存在(exists=true)，那么下单成功
	data := res["data"].(map[string]interface{})
	exists := data["exists"].(bool)
	if !exists {
		c.JSON(http.StatusOK, goms.ApiResponse(1000, "下单失败", nil))
		return
	}
	orderId := goms.NewUuid()
	c.JSON(http.StatusOK, goms.ApiResponse(0, "下单成功", goms.Any{
		"orderId": orderId,
	}))
}
