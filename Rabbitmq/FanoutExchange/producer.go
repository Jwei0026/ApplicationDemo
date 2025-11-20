package main

import (
	"github.com/streadway/amqp"
	"log"
)

//rabbitmq的广播交换机

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
		true,     // 是否持久化
		false,    // 是否自动删除
		false,    // 是否排他
		false,    // 是否阻塞
		nil,      // 额外参数
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange:", err)
	}

	body := "Broadcast message!"
	err = ch.Publish(
		"logs", // 交换机
		"",     // 路由键，扇形交换机不使用
		false,  // 是否立即发送
		false,  // 是否不等待
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		log.Fatal("Failed to publish a message:", err)
	}
	log.Printf("Sent: %s", body)
}
