package goms

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"goms/drives"
	"log"
)

// Request HTTP请求 调试结构体
// Method 请求方式
// Url 请求地址
// UrlParam 请求query参数
// ClientIP 请求客户端IP
// Headers 请求头map
// Body 请求体map
type Request struct {
	Method   string `json:"method"`
	Url      string `json:"url"`
	UrlParam string `json:"urlParam"`
	ClientIp string `json:"clientIp"`
	Headers  Any    `json:"headers"`
	Body     Any    `json:"body"`
}

// TraceLog 调试日志结构体
// RequestId 请求唯一标识id
// Request 请求结构体
// Event 事件名称
// Time 时间
// ServiceName 服务名称
// ServiceId 服务id
// ProjectName 项目名称
// Trace 调试记录信息map
type TraceLog struct {
	RequestId   string  `json:"requestId"`
	Request     Request `json:"request"`
	Event       string  `json:"event"`
	Time        string  `json:"time"`
	ServiceName string  `json:"serviceName"`
	ServiceId   string  `json:"serviceId"`
	ProjectName string  `json:"projectName"`
	Trace       Any     `json:"trace"`
}

// PushTraceLog 推送调试日志
// @Description 推送日志到消息队列，消息队列再存储到es
// traceLog 调试日志结构体
func PushTraceLog(traceLog *TraceLog) {
	str, _ := json.Marshal(traceLog)
	err := drives.AmqpPublish("trace", "trace", "goms", string(str))
	if err != nil {
		Logger.Debugf(err.Error())
	}
}

// UploadTraceLog 上传调试日志
// @Description 上传到es
// traceLog 调试日志结构体
func UploadTraceLog(traceLog string) error {
	_, err := drives.EsPut("trace_logs", traceLog)
	if err != nil {
		return err
	}
	return nil
}

// AmqpTraceConsume 调试日志消费逻辑
// @Parma msg rabbitmq消费消息体
func AmqpTraceConsume(msg amqp.Delivery) {
	body := string(msg.Body)
	err := UploadTraceLog(body)
	if err != nil {
		log.Print("consume fail: " + err.Error())
		Logger.Error("[trace consume fail] " + err.Error())
		Logger.Error("[trace consume fail body] " + body)
	}
	err = msg.Ack(true)
	if err != nil {
		return
	}
	log.Print("consume success: " + body)
}
