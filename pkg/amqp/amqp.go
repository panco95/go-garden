package amqp

import (
	"github.com/streadway/amqp"
	"log"
)

var amqpClient *amqp.Connection

// GetClient 获取rabbitmq客户端
func GetClient() *amqp.Connection {
	return amqpClient
}

// Connect 连接到Rabbitmq
// @param address rabbitmq连接地址
func Connect(address string) error {
	var err error
	amqpClient, err = amqp.Dial(address)
	if err != nil {
		return err
	}
	return nil
}

// Publish 发布消息
// @param queue 队列名称
// @param exchange 交换机
// @param routingKey 路由键
// @param body 消息内容
func Publish(queue, exchange, routingKey, body string) error {
	ch, err := amqpClient.Channel()
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

// Consumer 启动Rabbitmq消费者
// @param queue 队列名称
// @param exchange 交换机
// @param routingKey 路由键
// @param consumeFunc 消费方法
func Consumer(queue, exchange, routingKey string, consumeFunc func(msg amqp.Delivery)) error {
	ch, err := amqpClient.Channel()
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

// ConsumeExample 消息消费示例
// @Parma msg rabbitmq消费消息体
func ConsumeExample(msg amqp.Delivery) {
	body := string(msg.Body)
	log.Print(body)
	err := msg.Ack(true)
	if err != nil {
		return
	}
	log.Print("consume success!")
}

