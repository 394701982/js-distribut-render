package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config 结构体定义配置文件的格式
type Config struct {
	BrowserlessURLs []string `json:"browserlessURLs"` // Browserless服务的URL列表
	ThreadPools     []int    `json:"threadPools"`     // 线程池大小列表
	MemoryLimits    []uint64 `json:"memoryLimits"`    // 内存限制列表
	LogFilePath     string   `json:"logFilePath"`     // 日志文件路径
	Screenshot      bool     `json:"screenshot"`      // 是否启用截图
}

// LoadConfig 从指定的文件路径加载配置文件
func LoadConfig(filePath string) (Config, error) {
	var config Config

	// 打开配置文件
	configFile, err := os.Open(filePath)
	if err != nil {
		return config, err // 打开文件出错
	}
	defer configFile.Close() // 确保文件在函数结束时关闭

	// 读取配置文件内容
	bytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		return config, err // 读取文件内容出错
	}

	// 将JSON数据反序列化为Config结构体
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, err // 反序列化出错
	}

	return config, nil // 返回加载的配置
}
