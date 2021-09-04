package main

import (
	"github.com/spf13/viper"
	"goms"
	"log"
)

func main() {
	goms.InitLog()
	goms.InitConfig("config/config.yml", "yml")

	var err error

	esAddr := viper.GetString("esAddr")
	err = goms.EsConnect(esAddr)
	if err != nil {
		log.Fatal("[elasticsearch] " + err.Error())
	}

	amqpAddr := viper.GetString("amqpAddr")
	err = goms.AmqpConnect(amqpAddr)
	if err != nil {
		log.Fatal("[amqp] " + err.Error())
	}

	log.Printf("[amqp] trace consumer is runner")
	err = goms.AmqpConsumerRun("trace", goms.AmqpConsumeTrace)
	if err != nil {
		log.Fatal("[amqp] " + err.Error())
	}
}
