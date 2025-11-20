package main

import (
	"github.com/streadway/amqp"
	"log"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"logs",   // 交换机名
		"fanout", // 交换机类型
		true,     // 持久化
		false,    // 自动删除
		false,    // 排他
		false,    // 阻塞
		nil,      // 参数
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange:", err)
	}

	q, err := ch.QueueDeclare(
		"",    // 随机队列
		false, // 是否持久化
		true,  // 是否自动删除
		true,  // 排他队列
		false, // 是否阻塞
		nil,   // 额外参数
	)
	if err != nil {
		log.Fatal("Failed to declare a queue:", err)
	}

	err = ch.QueueBind(
		q.Name, // 队列名
		"",     // 路由键
		"logs", // 交换机名
		false,  // 是否阻塞
		nil,    // 额外参数
	)
	if err != nil {
		log.Fatal("Failed to bind a queue:", err)
	}

	msgs, err := ch.Consume(
		q.Name, // 队列名
		"",     // 消费者标签
		true,   // 自动确认
		false,  // 排他
		false,  // 不等待
		false,  // 不阻塞
		nil,    // 额外参数
	)
	if err != nil {
		log.Fatal("Failed to register a consumer:", err)
	}

	for msg := range msgs {
		log.Printf("Received: %s", msg.Body)
	}
}
