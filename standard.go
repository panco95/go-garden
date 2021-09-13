package goms

// ApiResponse 服务响应数据格式
func ApiResponse(code int, msg string, data interface{}) Any {
	return Any{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}