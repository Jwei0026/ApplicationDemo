package main

//Go并发编程示范
import (
	"fmt"
	"sync"
)

// 定义一个函数来计算数字的平方
func calculateSquare(num int, ch chan int, wg *sync.WaitGroup, mu *sync.RWMutex) {
	defer wg.Done() // 在函数结束时通知 WaitGroup  5.并发编程首先记得释放并发组成员
	square := num * num

	// 使用读写锁来保护对 channel 的访问
	mu.Lock()    // 写锁（确保只有一个写操作）
	ch <- square // 将结果写入channel内
	mu.Unlock()  // 解锁
}

func main() {
	// 定义一个待计算的数字列表
	nums := []int{1, 2, 3, 4, 5}

	// 创建一个 channel 用于接收计算结果
	resultChannel := make(chan int, len(nums))

	// 使用 WaitGroup 来等待所有 goroutine 完成  1.建立并发组
	var wg sync.WaitGroup

	// 使用 RWMutex 来保护对 channel 的并发访问 2.建立读写锁（多读一写）保证竞态
	var mu sync.RWMutex

	// 启动多个 goroutine 来计算每个数的平方
	for _, num := range nums {
		wg.Add(1) // 增加等待计数  3.给并发组加入成功再启动go
		go calculateSquare(num, resultChannel, &wg, &mu)
	}

	// 等待所有 goroutine 完成  4.等待所有的goroutine完成
	wg.Wait()

	// 关闭 channel，表示不再发送数据
	close(resultChannel)

	// 从 channel 中读取结果并打印
	// 使用读锁来读取 channel（这里其实是从 channel 中读取没有竞争，没必要加锁，但为了演示 RWMutex 使用）
	mu.RLock() // 读锁（多个读操作可以同时进行）
	for result := range resultChannel {
		fmt.Println(result)
	}
	mu.RUnlock() // 解锁
}

//执行步骤：创建交换信息的channel --- 创建并发组 --- 并发组Add --- 启动协程 -- 协程内defer done -- Lock-Unlock -- wait并发组 -- 关闭channel -- channel读传递的数据
