package goms

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"goms/drives"
	"log"
)

// Request HTTP请求 调试结构体
type Request struct {
	Method   string `json:"method"`
	Url      string `json:"url"`
	UrlParam string `json:"urlParam"`
	ClientIp string `json:"clientIp"`
	Headers  Any    `json:"headers"`
	Body     Any    `json:"body"`
}

// TraceLog 调试日志结构体
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
func PushTraceLog(traceLog *TraceLog) {
	str, _ := json.Marshal(traceLog)
	err := drives.AmqpPublish("trace", "trace", "trace", string(str))
	if err != nil {
		Logger.Debugf(err.Error())
	}
}

// UploadTraceLog 上传调试日志
// @Description 上传到es
func UploadTraceLog(traceLog string) error {
	_, err := drives.EsPut("trace_logs", traceLog)
	if err != nil {
		return err
	}
	return nil
}

// AmqpTraceConsume 调试日志消费逻辑
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
