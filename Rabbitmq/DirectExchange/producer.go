package main

//这是一个使用rabbit直接交换机的示例

import (
	"github.com/streadway/amqp"
	"log"
)

func main() {
	// 1.连接到 RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	//2.连接通道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	// 3.声明交换机 (核心一：设置交换机)
	err = ch.ExchangeDeclare(
		"direct_logs", // 交换机名
		"direct",      // 交换机类型
		true,          // 是否持久化
		false,         // 是否自动删除
		false,         // 是否排他
		false,         // 是否阻塞
		nil,           // 额外参数
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange:", err)
	}

	// 发送消息（核心二：发送消息）
	routingKey := "info"
	body := "Hello World!"
	err = ch.Publish(
		"direct_logs", // 交换机
		routingKey,    // 路由键
		false,         // 是否立即发送
		false,         // 是否不等待
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
