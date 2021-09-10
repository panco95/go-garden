package goms

import (
	"github.com/streadway/amqp"
	"log"
)

// AmqpClient Rabbitmq消息队列客户端
var (
	AmqpClient *amqp.Connection
)

// AmqpConnect 连接到Rabbitmq
// @param address rabbitmq连接地址
func AmqpConnect(address string) error {
	var err error
	AmqpClient, err = amqp.Dial(address)
	if err != nil {
		return err
	}
	return nil
}

// AmqpPublish 发布消息
// @param queue 队列名称
// @param exchange 交换机
// @param routingKey 路由键
// @param body 消息内容
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

// AmqpConsumer 启动Rabbitmq消费者
// @param queue 队列名称
// @param exchange 交换机
// @param routingKey 路由键
// @param consumeFunc 消费方法
func AmqpConsumer(queue, exchange, routingKey string, consumeFunc func(msg amqp.Delivery)) error {
	ch, err := AmqpClient.Channel()
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
