package monitor

import (
	"log"
	"runtime"
	"time"
)

// MonitorMemory 监控内存使用情况，当内存分配达到指定阈值时触发停止信号。
func MonitorMemory(threshold uint64, stopCh chan struct{}) {
	var m runtime.MemStats

	// 循环监控内存使用情况
	for {
		runtime.ReadMemStats(&m)  // 获取当前内存统计信息
		if m.Alloc >= threshold { // 如果当前分配的内存超过阈值
			log.Printf("Memory usage is high: %v bytes, triggering stop signal and GC", m.Alloc) // 记录高内存使用情况，并触发停止信号和垃圾回收
			runtime.GC()                                                                         // 执行强制垃圾回收
			stopCh <- struct{}{}                                                                 // 向停止通道发送信号
			return                                                                               // 结束监控
		}
		time.Sleep(1 * time.Second) // 每秒检查一次内存使用情况
	}
}
