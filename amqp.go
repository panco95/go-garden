package goms

import (
	"github.com/streadway/amqp"
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

func AmqpPublish(queue, body string) error {
	ch, err := AmqpClient.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(queue, false, false, false, false, nil)

	err = ch.Publish("", q.Name, false, false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})
	if err != nil {
		return nil
	}

	return nil
}

func AmqpConsumerRun(queue string, f func(msg amqp.Delivery)) error {
	ch, err := AmqpClient.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	forever := make(chan bool)

	q, err := ch.QueueDeclare(queue, false, false, false, false, nil)
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)

	go func() {
		for msg := range msgs {
			f(msg)
		}
	}()

	<-forever
	return nil
}
