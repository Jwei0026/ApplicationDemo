package main

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"sync"
	"time"
)

func main() {
	// Kafka writer 配置
	writer := kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "demo-topic",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	var wg sync.WaitGroup
	numMessages := 10 // 生产 10 条消息

	// 启动多个 goroutine 来并发发送消息
	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			msg := kafka.Message{
				Key:   []byte("Key-A"),
				Value: []byte("Hello Kafka " + time.Now().Format(time.RFC3339)),
			}

			err := writer.WriteMessages(context.Background(), msg)
			if err != nil {
				log.Fatalf("failed to write message: %v", err)
			}
			log.Printf("Message %d sent", i+1)
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	log.Println("All messages sent")
}
