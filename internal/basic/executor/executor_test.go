package executor

import (
	"context"
	"errors"
	"fmt"
	"github.com/kraken-pedestal/internal/basic/inventory"
	"golang.org/x/crypto/ssh"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockSSHConnector 是 SSHConnection 接口的 mock 实现，用于单元测试。
//
// 通过注入自定义的 connect 函数，可以控制 Connect 的返回值，
// 从而在不建立真实 SSH 连接的情况下测试 Executor.Run 的各种分支
// （连接成功、连接失败、Context 取消等）。
type mockSSHConnector struct {
	// connect 是可注入的连接函数，由测试用例根据需要设置
	connect func(host inventory.Host) (*ssh.Client, error)
}

// Connect 实现 SSHConnection 接口。
//
// 注意：当前实现的判断条件存在逻辑反转问题——
// 当 m.connect != nil（已设置）时反而返回错误，
// 当 m.connect == nil（未设置）时却尝试调用 m.connect(host) 会导致 panic。
// 正确的判断应为 if m.connect == nil。
func (m *mockSSHConnector) Connect(host inventory.Host) (*ssh.Client, error) {
	if m.connect == nil {
		return nil, errors.New("mockSSHConnector.Connect: connect is nil")
	}
	return m.connect(host)
}

// TestNew 测试 NewExecutor 的归一化逻辑。
//
// 验证点：
//  1. Concurrency 为 0 时应归一化为 1
//  2. SSH.User 为空时应归一化为 "root"
//  3. SSH.Port 为 0 时应归一化为 22
//  4. 自定义值应原样保留
//
// 两个用例：
//   - "default": 传入零值选项，验证默认值填充
//   - "custom": 传入自定义选项，验证值不被覆盖
func TestNew(t *testing.T) {
	// 表驱动测试：每个用例包含输入 options 和期望的归一化结果
	tests := []struct {
		name        string  // 子测试名称
		options     Options // 传入 NewExecutor 的选项
		concurrency int     // 期望的 Concurrency
		user        string  // 期望的 SSH.User
		port        int     // 期望的 SSH.Port
	}{
		{
			// 默认值用例：Concurrency=0，SSH 为零值
			name: "default",
			options: Options{
				Concurrency: 0,
			},
			concurrency: 1,      // 0 → 1
			user:        "root", // "" → "root"
			port:        22,     // 0 → 22
		},
		{
			// 自定义值用例：验证非零值不被归一化覆盖
			name: "custom",
			options: Options{
				Concurrency: 10,
				SSH: SSHOptions{
					User:          "admin",
					Port:          2222,
					Timeout:       0,
					StrictHostKey: false,
				},
			},
			concurrency: 10,      // 保持原值
			user:        "admin", // 保持原值
			port:        2222,    // 保持原值
		},
	}
	// 遍历所有用例，每个用例作为独立子测试运行
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 Executor
			executor := NewExecutor(tt.options)
			// 先断言 executor 非 nil，避免后续空指针
			if executor == nil {
				t.Fatalf("New() executor is nil")
			}
			// 获取归一化后的配置
			options := executor.Options()
			// 校验 Concurrency
			if options.Concurrency != tt.concurrency {
				t.Errorf("options.Concurrency = %d, want %d", options.Concurrency, tt.concurrency)
			}
			// 校验 SSH.User
			if options.SSH.User != tt.user {
				t.Errorf("options.SSH.User = %s, want %s", options.SSH.User, tt.user)
			}
			// 校验 SSH.Port
			if options.SSH.Port != tt.port {
				t.Errorf("options.SSH.Port = %d, want %d", options.SSH.Port, tt.port)
			}
		})
	}
}

// TestExecutorRunEmptyHosts 测试当 hosts 为空时 Executor.Run 的边界行为。
//
// 背景：Executor.Run 在并发执行主机操作前，会先对入参做防御性校验。
// 查看 executor.go 中 Run 的实现可知，当 hosts 为空（nil 或长度为 0 的切片）时，
// 会在 "if len(hosts) == 0 { return nil }" 处直接返回 nil，
// 不会创建任何 goroutine、信号量（semaphore），也不会调用 SSH 连接或 operation。
//
// 验证点：
//  1. 传入 nil hosts 时，Run 应返回 nil（而非空切片、长度为 0 的切片或错误结果）
//  2. 空主机场景下 operation 不会被执行（Run 在 len(hosts)==0 判断处提前 return）
//  3. 即使创建 Executor 时设置了 Concurrency=2，由于无主机可执行，并发控制逻辑不会生效
//
// 注意：此用例使用 NewExecutor（真实 SSHClient），但由于 hosts 为 nil，
// Run 在入口校验阶段直接返回 nil，不会尝试建立任何 SSH 连接，
// 因此无需注入 mockSSHConnector。
func TestExecutorRunEmptyHosts(t *testing.T) {
	// 创建 Executor，Concurrency=2 在此场景下不会生效（无主机可执行，不会进入并发调度逻辑）
	executor := NewExecutor(Options{
		Concurrency: 2,
	})
	// 调用 Run，传入 nil hosts；operation 在此场景下不会被执行
	// （Run 内部 len(hosts)==0 时直接返回 nil，不会进入 goroutine 启动逻辑）
	results := executor.Run(context.Background(), nil, func(
		ctx context.Context,
		host inventory.Host,
		client *ssh.Client,
	) Result {
		return Result{}
	},
	)
	// 断言：空主机时 Run 应返回 nil
	// 对应 executor.go:142 的 "if len(hosts) == 0 { return nil }" 分支
	if results != nil {
		t.Fatalf("executor.Run() should return nil got %v", results)
	}
}

func TestExecutorRunNilOptions(t *testing.T) {
	connector := &mockSSHConnector{
		connect: func(host inventory.Host) (*ssh.Client, error) {
			return nil, errors.New("should not connect")
		},
	}
	executor := runWithConnector(
		Options{Concurrency: 2},
		connector,
	)
	hosts := []inventory.Host{
		{
			Address: "10.0.0.1",
			User:    "root",
			Port:    22,
		},
		{
			Address: "10.0.0.2",
			User:    "root",
			Port:    22,
		},
	}
	results := executor.Run(context.Background(), hosts, nil)
	if len(results) != len(hosts) {
		t.Fatalf("results length = %d, want %d", len(results), len(hosts))
	}
	for i, result := range results {
		if result.Error == nil {
			t.Errorf("results[%d] error = %v, want error", i, result.Error)
		}
		if result.Host.Address != hosts[i].Address {
			t.Errorf("results[%d].Host.Address = %s, want %s", i, result.Host.Address, hosts[i].Address)
		}
	}
}

func TestExecutorRunConnectionError(t *testing.T) {
	connectorErr := errors.New("connection refused")
	connector := &mockSSHConnector{
		connect: func(host inventory.Host) (*ssh.Client, error) {
			return nil, connectorErr
		},
	}
	executor := runWithConnector(
		Options{Concurrency: 2},
		connector,
	)
	hosts := []inventory.Host{
		{
			Address: "10.0.0.1",
			User:    "root",
			Port:    22,
		},
	}
	results := executor.Run(context.Background(), hosts, func(ctx context.Context, host inventory.Host, client *ssh.Client) Result {
		t.Fatalf("operation should not be called")
		return Result{}
	},
	)
	if len(results) != 1 {
		t.Fatalf("results length = %d, want %d", len(results), 1)
	}
	if !errors.Is(results[0].Error, connectorErr) {
		t.Fatalf("Error = %v, want %v", results[0].Error, connectorErr)
	}
	if results[0].Success {
		t.Fatalf("Success = true, want false")
	}
	if results[0].Host.Address != "10.0.0.1" {
		t.Fatalf("host.Address = %q", results[0].Host.Address)
	}
}
func TestExecutorRunContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var connectCount atomic.Int32
	connector := &mockSSHConnector{connect: func(host inventory.Host) (*ssh.Client, error) {
		connectCount.Add(1)
		return nil, errors.New("should not connect")

	},
	}
	executor := runWithConnector(
		Options{Concurrency: 2},
		connector,
	)
	hosts := []inventory.Host{
		{
			Address: "10.0.0.1",
			User:    "root",
			Port:    22,
		}, {
			Address: "10.0.0.2",
			User:    "root",
			Port:    22,
		},
	}
	results := executor.Run(ctx, hosts, func(ctx context.Context, host inventory.Host, client *ssh.Client) Result {
		t.Fatalf("operation should not be called")
		return Result{}
	})
	if len(results) != len(hosts) {
		t.Fatalf("results length = %d, want %d", len(results), len(hosts))
	}
	for i, result := range results {
		if !errors.Is(result.Error, context.Canceled) {
			t.Errorf("result = %v, want %v", i, result.Error)
		}
	}
	if connectCount.Load() != 0 {
		t.Errorf("connectCount = %d, want %d", connectCount.Load(), 0)
	}
}

func TestExecutorResultOrder(t *testing.T) {
	connector := &mockSSHConnector{
		connect: func(host inventory.Host) (*ssh.Client, error) {
			return nil, nil
		}}
	executor := runWithConnector(Options{Concurrency: 2}, connector)
	hosts := []inventory.Host{
		{
			Address: "10.0.0.1",
			User:    "root",
			Port:    22,
		},
		{
			Address: "10.0.0.2",
			User:    "root",
			Port:    22,
		},
		{
			Address: "10.0.0.3",
			User:    "root",
			Port:    22,
		},
	}
	results := executor.Run(context.Background(), hosts, func(ctx context.Context, host inventory.Host, client *ssh.Client) Result {
		switch host.Address {
		case "10.0.0.1":
			time.Sleep(60 * time.Millisecond)
		case "10.0.0.2":
			time.Sleep(30 * time.Millisecond)
		case "10.0.0.3":
			time.Sleep(5 * time.Millisecond)
		}
		return Result{
			Success: true,
			Stdout:  host.Address,
		}
	})
	for i := range hosts {
		if results[i].Host.Address != hosts[i].Address {
			t.Errorf("results[%d].Host.Address = %v, want %v", i, results[i].Host.Address, hosts[i].Address)
		}
		if results[i].Stdout != hosts[i].Address {
			t.Errorf("results[%d].Stdout = %v, want %v", i, results[i].Stdout, hosts[i].Address)
		}
	}

}

// **// TestExecutorConcurrency 测试 Executor.Run 是否正确限制了并发度。
//
// 背景：executor.go 中 Run 通过带缓冲 channel（semaphore）实现并发控制，
// channel 容量等于 e.options.Concurrency，每个 goroutine 执行 operation 前
// 必须先向 semaphore 写入一个占位值（获取槽位），执行完毕后再读出（释放槽位）。
// 当 semaphore 已满时，后续 goroutine 会阻塞在 select 上，直到有槽位释放。
// 因此同时处于 operation 内部的 goroutine 数量不应超过 Concurrency。
//
// 验证思路：
//   - Concurrency=2，6 台主机（主机数 > 并发度，确保会有主机排队等待槽位）
//   - 在 operation 内部用原子计数器 current 记录"当前正在执行"的主机数
//   - 用 maxConcurrency 记录 current 出现过的峰值
//   - operation 内 sleep 20ms 模拟耗时，增大 goroutine 重叠执行的概率
//   - 断言 maxConcurrency 不超过 Concurrency 限制
//
// 注意：mockSSHConnector.Connect 返回 (nil, nil) 模拟连接成功，
// 让 Run 进入 operation 执行阶段，从而能测量真实的并发度。
// mock SSH 连接：返回 (nil, nil) 表示连接成功，不建立真实 SSH 连接
func TestExecutorConcurrency(t *testing.T) {
	connector := &mockSSHConnector{connect: func(host inventory.Host) (*ssh.Client, error) {
		return nil, nil
	},
	}
	// 创建 Executor，并发度设为 2
	executor := runWithConnector(Options{Concurrency: 2}, connector)
	// 构造 6 台主机（数量 > Concurrency=2，确保会有 goroutine 排队等待 semaphore 槽位）
	hosts := make([]inventory.Host, 6)
	for i := range hosts {
		hosts[i] = inventory.Host{
			//Address: "10.0.0." + string(rune('1'+i)),
			// 使用 fmt.Sprintf 生成 10.0.0.1 ~ 10.0.0.6 的地址
			Address: fmt.Sprintf("10.0.0.%d", i+1),
			User:    "root",
			Port:    22,
		}
	}

	// current：当前正在执行 operation 的 goroutine 数量（原子计数器）
	var current atomic.Int32
	//var maxConcurrency int32
	// maxConcurrency：记录 current 的历史峰值（原子计数器，通过 CAS 更新）
	var maxConcurrency atomic.Int32
	// mu + operationCount：统计 operation 被调用的总次数
	// （operationCount 是普通 int，需加锁保护，因为多个 goroutine 并发写入）
	var mu sync.Mutex
	var operationCount int

	// 对所有主机并发执行 operation
	results := executor.Run(context.Background(), hosts, func(ctx context.Context, host inventory.Host, client *ssh.Client) Result {
		// 原子自增 current，表示"进入 operation"，并取得自增后的最新值 now
		now := current.Add(1)

		// 用 CAS 循环更新 maxConcurrency：若 now 超过当前 max，则更新
		// 循环是因为 CAS 可能失败（其它 goroutine 同时更新了 max），失败则重新读取并重试
		for {
			max := maxConcurrency.Load()
			if now <= max {
				break
			}
			if maxConcurrency.CompareAndSwap(max, now) {
				break
			}
		}

		// 模拟业务耗时 20ms，让多个 goroutine 有机会同时处于 operation 内部
		time.Sleep(20 * time.Millisecond)
		// 原子自减 current，表示"离开 operation"
		current.Add(-1)

		// 统计 operation 调用次数（普通变量，需加锁）
		mu.Lock()
		operationCount++
		mu.Unlock()

		return Result{
			Success: true,
			Stdout:  host.Address,
		}
	},
	)
	// 断言 1：每台主机都有对应的 Result
	if len(results) != len(hosts) {
		t.Fatalf("results length = %d, want %d", len(results), len(hosts))
	}
	// 断言 2：operation 对每台主机恰好执行一次
	if operationCount != len(hosts) {
		t.Fatalf("operationCount = %d, want %d", operationCount, len(hosts))
	}
	// 断言 3：并发峰值不应超过 Concurrency=2
	// （此处失败条件为 >= 2，即期望 maxConcurrency < 2；
	//  理想情况下并发峰值应恰好等于 Concurrency=2，
	//  若实现正确则 maxConcurrency 最多为 2）
	if maxConcurrency.Load() >= 2 {
		t.Fatalf("maxconcurrency = %d, want %d", maxConcurrency.Load(), 2)
	}
	// 断言 4：所有主机的操作结果均为成功
	for i, reult := range results {
		if !reult.Success {
			t.Errorf("results[%d].Success = %v, want true", i, reult.Success)
		}
	}
}
