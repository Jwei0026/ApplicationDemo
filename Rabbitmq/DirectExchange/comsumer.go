package main

import (
	"github.com/streadway/amqp"
	"log"
)

func main() {
	//1.连接客户端
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	//2.在客户端的基础上建立通道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	//3.声明交换机
	err = ch.ExchangeDeclare(
		"direct_logs", // 交换机名
		"direct",      // 交换机类型
		true,          // 持久化
		false,         // 自动删除
		false,         // 排他
		false,         // 阻塞
		nil,           // 参数
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange:", err)
	}

	//4.声明接收的队列
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

	//5.绑定交换机与队列
	err = ch.QueueBind(
		q.Name,        // 队列名
		"info",        // 路由键
		"direct_logs", // 交换机名
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to bind a queue:", err)
	}

	//6.从通道内获取消费信息
	msgs, err := ch.Consume(
		q.Name, // 队列名
		"",     // 消费者标签
		true,   // 自动确认
		false,  // 不排他
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
