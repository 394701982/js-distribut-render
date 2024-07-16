package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"js-distribut-render/config"
	"js-distribut-render/monitor"
	"js-distribut-render/scanner"
)

func main() {
	// 定义命令行标志参数
	url := flag.String("url", "", "URL to render")                                       // 单个URL
	file := flag.String("file", "", "File containing list of URLs to render")            // 包含URL列表的文件
	configFilePath := flag.String("config", "config.json", "Path to configuration file") // 配置文件路径
	flag.Parse()                                                                         // 解析命令行标志参数

	// 检查是否提供了URL或文件
	if *url == "" && *file == "" {
		log.Fatal("Must provide URL or file containing URLs") // 必须提供URL或包含URL的文件
	}

	// 加载配置文件
	config, err := config.LoadConfig(*configFilePath)
	if err != nil {
		log.Fatalf("Error loading config file: %v", err) // 加载配置文件出错
	}

	// 打开日志文件
	logFile, err := os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err) // 打开日志文件出错
	}
	defer logFile.Close()  // 确保在程序结束时关闭日志文件
	log.SetOutput(logFile) // 将日志输出定向到日志文件

	// 读取URL列表
	var urls []string
	if *file != "" {
		content, err := ioutil.ReadFile(*file)
		if err != nil {
			log.Fatalf("Error reading file: %v", err) // 读取文件出错
		}
		urls = strings.Split(string(content), "\r\n") // 将文件内容按行分割成URL列表
	} else {
		urls = append(urls, *url) // 将单个URL添加到列表
	}

	// 定义WaitGroup和停止通道
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	// 启动内存监控协程
	for i := range config.BrowserlessURLs {
		memoryThreshold := config.MemoryLimits[i]
		go monitor.MonitorMemory(memoryThreshold, stopCh)
	}

	// 处理每个URL
	for i := range config.BrowserlessURLs {
		sem := make(chan struct{}, config.ThreadPools[i%len(config.ThreadPools)]) // 定义信号量，控制并发线程数

		for _, url := range urls {
			if url == "" {
				continue // 跳过空URL
			}
			select {
			case <-stopCh:
				log.Println("Memory threshold reached, stopping new tasks") // 内存阈值达到，停止新任务
				break
			default:
				wg.Add(1)         // 增加WaitGroup计数
				sem <- struct{}{} // 占用信号量
				go func(url string, browserlessURL string) {
					defer wg.Done()          // 完成任务后减少WaitGroup计数
					defer func() { <-sem }() // 释放信号量

					// 扫描URL
					result, err := scanner.Scan(url, config.Screenshot, browserlessURL)
					if err != nil {
						log.Printf("Error rendering URL %s: %v", url, err) // 渲染URL出错
						return
					}

					// 保存扫描结果
					err = scanner.SaveResult(url, result, config.Screenshot)
					if err != nil {
						log.Fatalf("Error saving result: %v", err) // 保存结果出错
					}
				}(url, config.BrowserlessURLs[i%len(config.BrowserlessURLs)])
			}
		}
	}

	wg.Wait() // 等待所有任务完成
}
