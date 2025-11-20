package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// 这是一个使用rabbit头部交换机的示例

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

	// 发送带有 header 的消息  不一样的地方  通过头部信息进行信息过滤，消费者会匹配该头部信息
	headers := amqp.Table{
		"animal": "tiger",
		"color":  "white",
	}

	err = ch.Publish(
		"headers_logs", // 交换机
		"",             // Headers exchange 不需要路由键
		false,          // 是否立刻发布
		false,          // 是否强制发布
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("Rabbit message with headers"),
			Headers:     headers,
		},
	)
	if err != nil {
		log.Fatal("Failed to publish a message:", err)
	}

	fmt.Println(" [x] Sent 'Rabbit message with headers'")
}
