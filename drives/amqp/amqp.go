package amqp

import (
	"github.com/streadway/amqp"
)

// Conn address format: amqp://guest:guest@192.168.125.186:5672
func Conn(address string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Publish message
func Publish(conn *amqp.Connection, kind, queue, exchange, routingKey, body string) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(queue, kind, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Publish(exchange, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         []byte(body),
	})

	if err != nil {
		return err
	}

	return nil
}

// Consumer listener
func Consumer(conn *amqp.Connection, kind, queue, exchange, routingKey string, consumeFunc func(msg amqp.Delivery)) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(queue, kind, true, false, false, false, nil)
	if err != nil {
		return err
	}
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	if kind != "fanout" {
		err = ch.QueueBind(q.Name, routingKey, exchange, false, nil)
		if err != nil {
			return err
		}
	}

	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			consumeFunc(msg)
		}
	}()

	forever := make(chan bool)
	<-forever
	return nil
}
