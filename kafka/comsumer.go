package main

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"sync"
	"time"
)

//kafka也是要先启动消费者创建对应的主题和分区然后再启动生产者产生消息

func consumeMessages(reader *kafka.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	// 设定一个上下文和超时，避免永久阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 30秒超时
	defer cancel()
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("could not read message: %v", err)
			return
		}
		log.Printf("Message received: key=%s value=%s offset=%d", string(m.Key), string(m.Value), m.Offset)
	}
}

func main() {
	// Kafka reader 配置
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"}, //设置实例brokers
		Topic:   "demo-topic",               //设置主题
		GroupID: "demo-group",               //对应的消费者组
	})
	defer reader.Close()

	var wg sync.WaitGroup
	numConsumers := 3 // 启动 3 个并发消费者

	// 启动多个消费者 goroutine
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go consumeMessages(reader, &wg)
	}

	// 等待所有消费者完成
	wg.Wait()
	log.Println("All consumers have finished")
}
