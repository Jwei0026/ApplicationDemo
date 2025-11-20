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

//如果先启动生产者，再启动消费者可能会出现消费者消费不到消息的问题：
//RabbitMQ 的消费者和生产者在消息传递时依赖于交换机和队列的声明。如果消费者在队列还没有被声明时就开始监听，那么消费者可能无法接收到生产者发送的消息。

//问题所在：如果生产者先启动，发送消息时，消费者的队列还没有被声明，消息不会被路由到正确的队列，也就无法被消费。
//解决方法：确保消费者在启动时声明队列，生产者再进行消息发布，或者在启动生产者之前就声明队列。

//如果消息发送和消费者的接收在时间上有错配，比如生产者发布消息的时机很短暂，消费者启动稍有延迟，消息可能会丢失，
//尤其是在某些交换机类型下（如 fanout 和 direct），如果消息没有存储（如未设置持久化），消息可能在消费者启动之前已经被丢弃。
//解决方法：确保消费者能及时启动，并且可以尝试使用 durable 队列和 persistent 消息，这样即使消费者启动较慢，消息仍然能够被保存。
