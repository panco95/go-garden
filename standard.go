package garden

type MapData map[string]interface{}

var(
	syncCache []byte
)

func gatewaySuccess(data MapData) MapData {
	response := MapData{
		"status": true,
	}
	for k, v := range data {
		response[k] = v
	}
	return response
}

func gatewayFail() MapData {
	response := MapData{
		"status": false,
	}
	return response
}

func ApiResponse(code int, msg string, data interface{}) MapData {
	return MapData{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
