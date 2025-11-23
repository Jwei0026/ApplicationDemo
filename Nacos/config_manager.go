package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"gopkg.in/yaml.v2"
)

//生成级别的nacos配置管理器工具包

// ConfigManager 配置管理器，负责配置的获取、监听和更新
type ConfigManager struct {
	client       config_client.IConfigClient //客户端
	config       *ConfigData                 //配置信息
	mutex        sync.RWMutex                //读写锁
	waitGroup    sync.WaitGroup              //等待组
	stopChan     chan struct{}               //停止通道
	listeners    []ConfigChangeListener      //监听器
	initialized  bool                        //是否初始化
	configParams vo.ConfigParam              //配置参数
}

// ConfigChangeListener 配置变更监听器接口
type ConfigChangeListener interface {
	OnConfigChange(config *ConfigData)
}

// NewConfigManager 创建配置管理器实例
func NewConfigManager(client config_client.IConfigClient, dataId, group, configType string) *ConfigManager {
	return &ConfigManager{
		client:    client,
		config:    &ConfigData{},
		stopChan:  make(chan struct{}),
		listeners: make([]ConfigChangeListener, 0),
		configParams: vo.ConfigParam{
			DataId: dataId,
			Group:  group,
			Type:   vo.ConfigType(configType),
		},
	}
}

// Start 启动配置管理器
func (cm *ConfigManager) Start() error {
	if cm.initialized {
		return fmt.Errorf("配置管理器已经初始化")
	}

	// 首先获取初始配置
	err := cm.loadInitialConfig()
	if err != nil {
		return fmt.Errorf("加载初始配置失败: %v", err)
	}

	// 在单独的goroutine中启动配置监听
	cm.waitGroup.Add(1)
	go cm.listenForChanges()

	cm.initialized = true
	fmt.Println("配置管理器已成功启动")
	return nil
}

// loadInitialConfig 加载初始配置
func (cm *ConfigManager) loadInitialConfig() error {
	content, err := cm.client.GetConfig(cm.configParams)
	if err != nil {
		return err
	}

	return cm.updateConfig(content)
}

// listenForChanges 监听配置变化
func (cm *ConfigManager) listenForChanges() {
	defer cm.waitGroup.Done()

	// 启动Nacos监听
	err := cm.client.ListenConfig(vo.ConfigParam{
		DataId: cm.configParams.DataId,
		Group:  cm.configParams.Group,
		Type:   cm.configParams.Type,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("检测到配置变更，正在更新...")
			err := cm.updateConfig(data)
			if err != nil {
				fmt.Printf("配置更新失败: %v\n", err)
				return
			}

			// 通知所有监听器
			cm.notifyListeners()
			fmt.Println("配置已成功更新并通知所有监听器")
		},
	})

	if err != nil {
		fmt.Printf("启动配置监听失败: %v\n", err)
		return
	}

	// 等待停止信号
	select {
	case <-cm.stopChan:
		fmt.Println("收到停止信号，配置监听将停止")
	}
}

// updateConfig 更新配置内容
func (cm *ConfigManager) updateConfig(content string) error {
	var newConfig ConfigData
	err := yaml.Unmarshal([]byte(content), &newConfig)
	if err != nil {
		return fmt.Errorf("YAML解析失败: %v, 内容: %s", err, content)
	}

	// 加锁更新配置
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	*cm.config = newConfig
	return nil
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() *ConfigData {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 返回配置的副本，避免外部修改
	configCopy := *cm.config
	return &configCopy
}

// AddListener 添加配置变更监听器
func (cm *ConfigManager) AddListener(listener ConfigChangeListener) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.listeners = append(cm.listeners, listener)
}

// notifyListeners 通知所有监听器配置已变更
func (cm *ConfigManager) notifyListeners() {
	// 获取配置副本
	config := cm.GetConfig()

	// 在goroutine中通知每个监听器，避免阻塞
	for _, listener := range cm.listeners {
		go func(l ConfigChangeListener) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("监听器处理配置变更时发生panic: %v\n", r)
				}
			}()
			l.OnConfigChange(config)
		}(listener)
	}
}

// Stop 停止配置管理器
func (cm *ConfigManager) Stop() {
	if !cm.initialized {
		return
	}

	// 发送停止信号
	close(cm.stopChan)

	// 等待监听goroutine结束
	cm.waitGroup.Wait()

	cm.initialized = false
	fmt.Println("配置管理器已成功停止")
}

//上面的时包内的工具，下面的时外部调用的接口

// ExampleService 使用配置的示例服务
type ExampleService struct {
	configManager *ConfigManager
}

// NewExampleService 创建示例服务
func NewExampleService(configManager *ConfigManager) *ExampleService {
	service := &ExampleService{
		configManager: configManager,
	}

	// 注册为配置变更监听器
	configManager.AddListener(service)
	return service
}

// OnConfigChange 实现ConfigChangeListener接口
func (s *ExampleService) OnConfigChange(config *ConfigData) {
	fmt.Println("\nExampleService检测到配置更新:")
	fmt.Printf("  应用名称: %s\n", config.AppName)
	fmt.Printf("  服务器端口: %d\n", config.ServerPort)
	fmt.Printf("  数据库URL: %s\n", config.Database.Url)
	fmt.Printf("  数据库用户名: %s\n", config.Database.Username)
	fmt.Printf("  数据库密码: ********\n") // 不显示密码明文
	fmt.Printf("  功能列表: %v\n", config.Features)

	// 在这里可以进行服务重启、资源重新初始化等操作
}

// GetConfig 获取当前配置
func (s *ExampleService) GetConfig() *ConfigData {
	return s.configManager.GetConfig()
}

// SetupGracefulShutdown 设置优雅关闭
func SetupGracefulShutdown(managers ...*ConfigManager) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	// 监听系统信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// 创建一个带缓冲的信号通道c，缓冲区大小为1
	// 使用signal.Notify注册对os.Interrupt（Ctrl+C）和syscall.SIGTERM（kill命令默认信号）的监听
	// 缓冲区大小设为1可确保不会错过信号，即使处理信号的goroutine没有立即准备好接收

	go func() {
		<-c
		fmt.Println("\n接收到中断信号，开始优雅关闭...")

		// 创建关闭上下文，设置超时
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		// 停止所有配置管理器
		for _, manager := range managers {
			manager.Stop()
		}

		// 等待关闭完成或超时
		select {
		case <-shutdownCtx.Done():
			fmt.Println("关闭超时，强制退出")
		default:
			fmt.Println("所有组件已成功关闭")
		}

		cancel()
	}()

	return ctx
}
