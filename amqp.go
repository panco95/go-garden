package goms

import (
	"github.com/streadway/amqp"
	"log"
)

var (
	AmqpClient *amqp.Connection
)

func AmqpConnect(address string) error {
	var err error
	AmqpClient, err = amqp.Dial(address)
	if err != nil {
		return err
	}
	return nil
}

func AmqpPublish(queue, exchange, routingKey, body string) error {
	ch, err := AmqpClient.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(queue, "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Publish(exchange, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         []byte(body),
	})

	if err != nil {
		return nil
	}

	return nil
}

func AmqpConsumer(queue, exchange, routingKey string, f func(msg amqp.Delivery)) error {
	ch, err := AmqpClient.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(queue, "direct", true, false, false, false, nil)
	err = ch.QueueBind(queue, routingKey, exchange, false, nil)
	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)

	go func() {
		for msg := range msgs {
			f(msg)
		}
	}()

	log.Printf("[amqp] consumer is running")

	forever := make(chan bool)
	<-forever
	return nil
}
