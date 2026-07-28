package logger

import (
	"fmt"
	"sync"
	"time"
)

// BatchLogger 批量执行日志记录器
type BatchLogger struct {
	totalHosts int
	completed  int
	success    int
	failed     int
	mu         sync.Mutex
	startTime  time.Time
}

// NewBatchLogger 创建批量执行日志记录器
func NewBatchLogger(totalHosts int) *BatchLogger {
	return &BatchLogger{
		totalHosts: totalHosts,
		startTime:  time.Now(),
	}

}

// HostStart 主机开始执行
func (b *BatchLogger) HostStart(host, command string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	Info("开始执行命令",
		"host", host,
		"command", command,
		"progress", fmt.Sprintf("%d/%d", b.completed+1, b.totalHosts),
		"elapsed", time.Since(b.startTime).Round(time.Millisecond),
	)
}

// HostComplete 主机执行完成
func (b *BatchLogger) HostComplete(host string, err error, duration time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.completed++

	if err == nil {
		b.success++
		Info("命令执行成功",
			"host", host,
			"duration", duration.Round(time.Millisecond),
			"success_count", b.success,
			"failed_count", b.failed,
			"progress", fmt.Sprintf("%d/%d", b.completed, b.totalHosts),
			"elapsed", time.Since(b.startTime).Round(time.Millisecond),
		)
	} else {
		b.failed++
		Error("命令执行失败",
			"host", host,
			"error", err.Error(),
			"duration", duration.Round(time.Millisecond),
			"success_count", b.success,
			"failed_count", b.failed,
			"progress", fmt.Sprintf("%d/%d", b.completed, b.totalHosts),
			"elapsed", time.Since(b.startTime).Round(time.Millisecond),
		)
	}

	// 全部完成时输出汇总
	if b.completed == b.totalHosts {
		totalDuration := time.Since(b.startTime)
		Info("批量执行完成",
			"total_hosts", b.totalHosts,
			"success", b.success,
			"failed", b.failed,
			"success_rate", fmt.Sprintf("%.1f%%", float64(b.success)/float64(b.totalHosts)*100),
			"total_duration", totalDuration.Round(time.Millisecond),
			"avg_duration", (totalDuration / time.Duration(b.totalHosts)).Round(time.Millisecond),
		)
	}
}
