package main

import (
	"goms"
	"goms/drives"
	"log"
)

func main() {
	serviceName := "traceConsumer"
	projectName := "goms"
	goms.Init("", "", serviceName, projectName)
	log.Fatal(drives.AmqpConsumer("trace", "trace", "trace", goms.AmqpTraceConsume))
}
