package main

import (
	"goms"
	"log"
)

func main() {
	serviceName := "traceConsumer"
	projectName := "goms"
	goms.Init("", "", serviceName, projectName)
	log.Fatal(goms.AmqpConsumerRun("trace", goms.AmqpConsumeTrace))
}
