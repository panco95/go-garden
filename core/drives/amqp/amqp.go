package amqp

import (
	"github.com/streadway/amqp"
	"log"
)

var client *amqp.Connection

// Client get class
func Client() *amqp.Connection {
	return client
}

// Connect amqp server like rabbitmq
func Connect(address string) error {
	var err error
	client, err = amqp.Dial(address)
	if err != nil {
		return err
	}
	return nil
}

// Publish message
func Publish(queue, exchange, routingKey, body string) error {
	ch, err := client.Channel()
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

// Consumer message
func Consumer(queue, exchange, routingKey string, consumeFunc func(msg amqp.Delivery)) error {
	ch, err := client.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(queue, "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatal("ExchangeDeclare：" + err.Error())
	}
	q, err := ch.QueueDeclare(queue, false, false, true, false, nil)
	if err != nil {
		log.Fatal("QueueDeclare：" + err.Error())
	}
	err = ch.QueueBind(q.Name, routingKey, exchange, false, nil)
	if err != nil {
		log.Fatal("QueueBind：" + err.Error())
	}
	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.Print("Consume：" + err.Error())
	}

	go func() {
		for msg := range msgs {
			consumeFunc(msg)
		}
	}()

	log.Printf("[amqp] consumer is running")

	forever := make(chan bool)
	<-forever
	return nil
}

// ConsumeExample test
func ConsumeExample(msg amqp.Delivery) {
	body := string(msg.Body)
	log.Print(body)
	err := msg.Ack(true)
	if err != nil {
		return
	}
	log.Print("consume success!")
}
