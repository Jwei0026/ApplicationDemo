package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"gopkg.in/yaml.v2"
)

// 定义解析yaml文件装载的结构体
// 定义一个结构体用于存储YAML配置
type ConfigData struct {
	AppName    string   `yaml:"appName"`
	ServerPort int      `yaml:"serverPort"`
	Database   Database `yaml:"database"`
	Features   []string `yaml:"features"`
}

// 数据库配置结构体
type Database struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func main() {
	fmt.Println("Nacos非阻塞配置热更新示例")
	//创建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         "", // 使用public命名空间，确保发布和读取配置在同一命名空间
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "./tmp/nacos/log",
		CacheDir:            "./tmp/nacos/cache",
		LogLevel:            "debug",
	}

	//创建服务端配置
	// 至少一个ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "localhost",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
		{
			IpAddr:      "localhost",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}

	// 创建动态配置客户端的另一种方式 (推荐)
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("配置客户端创建成功")
	// PublishConfig(configClient)

	// 初始读取配置
	read(configClient)
	rollbackConfig()

	// 在独立的goroutine中启动配置监听，不阻塞主程序运行
	go ListenConfig(configClient)

	// 主程序继续运行其他业务逻辑
	fmt.Println("\n主程序继续执行业务逻辑...")

	// 模拟业务逻辑持续运行
	for i := 1; ; i++ {
		fmt.Printf("业务逻辑执行第 %d 次\n", i)
		// 每秒执行一次业务逻辑
		time.Sleep(10 * time.Second)
	}

	// 注意：由于上面的循环是无限的，下面的阻塞代码实际上不会执行到
	// 如果需要让程序能够优雅退出，可以使用信号处理机制
	// blockChan := make(chan struct{})
	// <-blockChan
}

func read(configClient config_client.IConfigClient) {
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: "dataId",
		Group:  "group",
		Type:   "yaml",
	})
	if err != nil {
		fmt.Println("获取配置出错:", err.Error())
		return
	}
	fmt.Println("原始配置内容:", content)

	// 将配置保存到config.yaml文件
	err = os.WriteFile("config.yaml", []byte(content), 0644)
	if err != nil {
		fmt.Println("保存配置到文件出错:", err.Error())
	} else {
		fmt.Println("配置已成功保存到config.yaml文件")
	}

	// 检查是否是YAML格式
	fmt.Println("\n解析YAML配置...")
	var config ConfigData
	err = yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		fmt.Println("YAML解析失败:", err.Error())
		// 如果不是YAML格式，可能是普通文本
		fmt.Println("配置可能不是YAML格式，使用字符串方式处理")
		if content == "hello world!222222" {
			fmt.Println("获取配置成功")
		}
		return
	}

	// 配置已成功解析并保存在config变量中
	fmt.Println("YAML解析成功，配置已保存到变量中")
	fmt.Println("\n解析后的配置内容:")
	fmt.Printf("  应用名称: %s\n", config.AppName)
	fmt.Printf("  服务器端口: %d\n", config.ServerPort)
	fmt.Printf("  数据库URL: %s\n", config.Database.Url)
	fmt.Printf("  数据库用户名: %s\n", config.Database.Username)
	fmt.Printf("  数据库密码: %s\n", config.Database.Password)
	fmt.Printf("  功能列表: %v\n", config.Features)

	// 现在可以使用这个config变量在程序的其他地方
	fmt.Println("\n配置已成功保存到本地变量和config.yaml文件，可以在程序中使用")
}

func PublishConfig(configClient config_client.IConfigClient) {
	// 创建一个YAML格式的配置字符串
	yamlContent := `appName: demo-service
serverPort: 8080
database:
  url: jdbc:mysql://localhost:3306/mydb
  username: admin
  password: password123
features:
  - login
  - dashboard
  - reporting
`

	fmt.Println("准备发布YAML格式配置...")
	success, err := configClient.PublishConfig(vo.ConfigParam{
		DataId:  "dataId",
		Group:   "group",
		Content: yamlContent,
		Type:    "yaml"}) // 发布YAML格式的配置
	if err != nil {
		fmt.Println("发布配置出错:", err.Error())
		return
	}
	if success {
		fmt.Println("YAML格式配置发布成功")
		fmt.Println("发布的配置内容:")
		fmt.Println(yamlContent)
	} else {
		fmt.Println("发布配置失败")
	}
}

func ListenConfig(configClient config_client.IConfigClient) {
	fmt.Println("开始在独立goroutine中监听配置变化...")

	// 尝试启动配置监听
	configClient.ListenConfig(vo.ConfigParam{
		DataId: "dataId",
		Group:  "group",
		Type:   "yaml",
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("\n==========================================")
			fmt.Println("[配置热更新触发]")
			fmt.Printf("检测到配置变化 - dataId: %s, group: %s\n", dataId, group)
			// 重新读取配置并更新文件
			read(configClient)
			fmt.Println("配置热更新完成")
			fmt.Println("==========================================\n")
		},
	})

	fmt.Println("配置监听已启动，将在后台独立运行...")
	// 保持goroutine运行
	select {}
}

// 版本管理demo
func rollbackConfig() {
	//进行版本管理可以进行如下操作：
	//1.通过给dataId添加版本号来实现配置的版本管理
	//2.通过使用searchConfig函数来搜索配置版本，找到需要回滚的版本号
	//3.调用publishConfig函数来发布回滚后的配置

	//searcgConfig使用实例如下：
	// searchResult, err := client.SearchConfig(vo.SearchConfigParam{
	// 	Search:   "blur",    // 搜索关键字
	// 	DataId:   "",        // 不指定数据 ID，表示搜索所有配置
	// 	Group:    "",        // 不指定配置分组，表示搜索所有分组
	// 	PageNo:   1,         // 页码，默认从第 1 页开始
	// 	PageSize: 10,        // 每页返回 10 个配置项
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to search config: %v", err)
	// }
	// 输出搜索结果
	// fmt.Println("Search result:")
	// fmt.Printf("Total count: %d\n", searchResult.Count)
	// for _, config := range searchResult.Configs {
	// 	fmt.Printf("DataId: %s, Group: %s, Content: %s\n", config.DataId, config.Group, config.Content)
	// }
	//返回的结果：
	// Search result:
	// Total count: 2
	// DataId: app-config, Group: DEFAULT_GROUP, Content: blur effect
	// DataId: image-config, Group: DEFAULT_GROUP, Content: background blur
}
