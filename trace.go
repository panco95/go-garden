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
	err := UploadTraceLog(string(msg.Body))
	if err != nil {
		Logger.Error("[amqp trace consumer error] " + err.Error())
	}
	log.Print("amqp trace consumer success")
}
