package core

import (
	"github.com/streadway/amqp"
)

type amqpMsg struct {
	Cmd  string
	Data MapData
}

func (g *Garden) connAmqp(address string) error {
	var err error
	g.amqp, err = amqp.Dial(address)
	if err != nil {
		return err
	}
	return nil
}

func (g *Garden) amqpPublish(kind, queue, exchange, routingKey, body string) error {
	ch, err := g.amqp.Channel()
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

func (g *Garden) amqpConsumer(kind, queue, exchange, routingKey string, consumeFunc func(msg amqp.Delivery)) error {
	ch, err := g.amqp.Channel()
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

	g.Log(InfoLevel, "amqp", queue+" consumer is running")

	forever := make(chan bool)
	<-forever
	return nil
}
