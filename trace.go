package goms

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
)

type Request struct {
	Method   string `json:"method"`
	Url      string `json:"url"`
	UrlParam string `json:"urlParam"`
	ClientIp string `json:"clientIp"`
	Headers  Any    `json:"headers"`
	Body     Any    `json:"body"`
}

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

func PushTraceLog(traceLog *TraceLog) {
	str, _ := json.Marshal(traceLog)
	err := AmqpPublish("trace", string(str))
	if err != nil {
		Logger.Debugf(err.Error())
	}
}

func UploadTraceLog(traceLog string) error {
	_, err := EsPut("trace_logs", traceLog)
	if err != nil {
		return err
	}
	return nil
}

func AmqpConsumeTrace(msg amqp.Delivery) {
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
