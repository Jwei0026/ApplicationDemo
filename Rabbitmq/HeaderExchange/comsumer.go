package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	// 连接到 RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	// 创建一个频道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	// 声明一个 Headers 类型的交换机
	err = ch.ExchangeDeclare(
		"headers_logs", // 交换机名称
		"headers",      // 交换机类型
		true,           // 是否持久化
		false,          // 是否自动删除
		false,          // 是否独占
		false,          // 是否等待确认
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange:", err)
	}

	// 创建一个匿名队列
	q, err := ch.QueueDeclare(
		"",    // 随机队列名称
		false, // 是否持久化
		true,  // 是否自动删除
		true,  // 是否独占
		false, // 是否等待确认
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare a queue:", err)
	}

	// 绑定队列到交换机，设置头部匹配条件
	err = ch.QueueBind(
		q.Name,         // 队列名
		"",             // Headers exchange 不使用路由键
		"headers_logs", // 交换机名
		false,
		map[string]interface{}{
			"x-match": "any", // 选择 'all' 或 'any'  会匹配该头部信息来过滤元素
			"animal":  "rabbit",
			"color":   "white",
		},
	)
	if err != nil {
		log.Fatal("Failed to bind a queue:", err)
	}

	// 定义回调函数
	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	msgs, err := ch.Consume(
		q.Name, // 队列名称
		"",     // 消费者名称
		true,   // 是否自动确认
		false,  // 是否独占
		false,  // 是否等待确认
		false,  // 是否等待
		nil,
	)
	if err != nil {
		log.Fatal("Failed to register a consumer:", err)
	}

	// 消费消息
	for msg := range msgs {
		fmt.Printf(" [x] Received %s with headers %v\n", msg.Body, msg.Headers)
	}
}
