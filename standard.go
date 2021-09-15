package garden

import "github.com/gin-gonic/gin"

func Run(route func(r *gin.Engine), auth func() gin.HandlerFunc) {
	go InitRpc(Config.RpcPort)
	Fatal("Run", GinServer(Config.HttpPort, route, auth))
}

type Any map[string]interface{}

func GatewaySuccess(data Any) Any {
	response := Any{
		"status": true,
	}
	for k, v := range data {
		response[k] = v
	}
	return response
}

func GatewayFail() Any {
	response := Any{
		"status": false,
	}
	return response
}

func ApiResponse(code int, msg string, data interface{}) Any {
	return Any{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
