package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// 这是一个使用rabbit主题交换机的示例
func main() {
	// 连接到 RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	// 创建一个通道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	// 声明一个 Topic 类型的交换机
	err = ch.ExchangeDeclare(
		"topic_logs", // 交换机名称
		"topic",      // 交换机类型
		true,         // 是否持久化
		false,        // 是否自动删除
		false,        // 是否独占
		false,        // 是否等待确认
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange:", err)
	}

	// 发送消息，路由键为 'animal.rabbit'
	body := "Rabbit message"
	err = ch.Publish(
		"topic_logs",   // 交换机
		"animal.tiger", // 路由键   不一样的地方
		false,          // 是否立刻发布
		false,          // 是否强制发布
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		log.Fatal("Failed to publish a message:", err)
	}

	fmt.Println(" [x] Sent", body)
}
