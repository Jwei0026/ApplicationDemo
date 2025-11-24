package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func main() {
	// Nacos 服务配置
	serverConfig := []constant.ServerConfig{
		{
			IpAddr: "127.0.0.1", // Nacos 服务的 IP 地址
			Port:   8848,        // Nacos 服务的端口
		},
	}

	// 客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         "public", // 默认命名空间
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "./logs", // 日志目录
	}

	// 创建动态配置客户端的另一种方式 (推荐)
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfig,
		})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("配置客户端创建成功")

	// 注册服务
	success, err := client.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: "myservice", // 服务名称（与Nacos控制台显示保持一致）
		Ip:          "127.0.0.1", // 服务实例的 IP 地址
		Port:        8080,        // 服务实例的端口号
		Weight:      1.0,         // 权重（负载均衡的基础）
		Enable:      true,        // 实例是否启用
		Ephemeral:   true,        // 是否为临时实例
		Healthy:     true,        // 实例是否健康
	})
	if err != nil {
		log.Fatalf("Error registering service instance: %v", err)
	}
	if success {
		fmt.Println("Service registered successfully!")
	} else {
		fmt.Println("Service registration failed!")
	}

	// 注册后添加延迟，确保服务实例完全生效
	fmt.Println("等待1秒让服务实例完全注册...")
	time.Sleep(1 * time.Second)

	// 先尝试使用GetService方法验证服务是否存在
	service, err := client.GetService(vo.GetServiceParam{
		ServiceName: "myservice",
		GroupName:   "DEFAULT_GROUP",
	})
	if err != nil {
		fmt.Printf("获取服务信息失败: %v\n", err)
	} else {
		fmt.Printf("服务信息: %+v\n", service)
	}
	// 通过服务名称查询实例
	go func() {
		for {
			// 获取服务实例列表
			fmt.Println("正在查询服务实例: myservice")
			// 尝试使用健康检查为false的查询方式
			res, err := client.SelectInstances(vo.SelectInstancesParam{
				ServiceName: "myservice",     // 服务名称（与Nacos控制台显示保持一致）
				GroupName:   "DEFAULT_GROUP", // 分组名称
				HealthyOnly: true,            // 只返回健康的实例
			})
			// 如果查询失败，尝试使用SelectOneHealthyInstance方法

			if err != nil {
				// 检查是否是实例列表为空的情况，这种情况在没有服务注册时是正常的
				if err.Error() == "instance list is empty!" {
					fmt.Println("未发现服务实例，请等待服务注册...")
				} else {
					log.Printf("查询服务实例错误: %v", err)
				}
				continue // 继续下一次查询
			}

			// 打印服务实例
			fmt.Println("服务实例列表:")
			if len(res) == 0 {
				fmt.Println("  [暂无可用实例]")
			} else {
				for _, instance := range res {
					fmt.Printf("  IP: %s, Port: %d\n", instance.Ip, instance.Port, instance.Weight)
				}
			}

			// 每隔 3 秒查询一次
			time.Sleep(3 * time.Second)
		}
	}()

	// 模拟服务运行，保持客户端活跃
	select {}
}
