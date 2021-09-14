package garden

import "github.com/gin-gonic/gin"

// Run 服务启动
// 目前仅支持gin框架
func Run(route func(r *gin.Engine), auth func() gin.HandlerFunc) {
	go InitRpc(Config.RpcPort)
	Fatal("Run", GinServer(Config.HttpPort, route, auth))
}

// Any 抽象map封装
type Any map[string]interface{}

// GatewaySuccess 网关统一成功响应
func GatewaySuccess(data Any) Any {
	response := Any{
		"status": true,
	}
	for k, v := range data {
		response[k] = v
	}
	return response
}

// GatewayFail 网关统一失败响应
func GatewayFail() Any {
	response := Any{
		"status": false,
	}
	return response
}

// ApiResponse 服务响应数据格式
func ApiResponse(code int, msg string, data interface{}) Any {
	return Any{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
