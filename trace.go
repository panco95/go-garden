package goms

import "encoding/json"

type Req struct {
	Method   string `json:"method"`
	Url      string `json:"url"`
	UrlParam string `json:"urlParam"`
	ClientIp string `json:"clientIp"`
	Headers  Any    `json:"headers"`
	Body     Any    `json:"body"`
}

// ReqTrace 请求上下文封装
type ReqTrace struct {
	RequestId   string `json:"requestId"`
	Req         Req    `json:"req"`
	Event       string `json:"event"`
	Time        string `json:"time"`
	ServiceName string `json:"serviceName"`
	ServiceId   string `json:"serviceId"`
	ProjectName string `json:"projectName"`
	Trace       Any    `json:"trace"`
}

func RemoteTrace(rc *ReqTrace) {
	logJson, _ := json.Marshal(rc)
	_, err := EsPut("logs", string(logJson))
	if err != nil {
		Logger.Errorf("[elasticsearch] %s", err)
	}
}
