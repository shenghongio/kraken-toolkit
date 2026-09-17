package executor

import (
	"context"
	"fmt"
	"github.com/kraken-pedestal/internal/basic/inventory"
	"golang.org/x/crypto/ssh"
	"sync"
	"time"
)

// Result 表示单个主机上一次操作的执行结果。
//
// 字段说明：
//   - Host: 目标主机信息，标识结果属于哪台主机
//   - Success: 操作是否成功（SSH 连接成功且 operation 执行无错误）
//   - ExitCode: 远程命令退出码（仅 shell/script 类操作有意义）
//   - Stdout: 远程命令标准输出
//   - Stderr: 远程命令标准错误
//   - Error: 执行过程中的错误（连接失败、操作错误等），成功时为 nil
//   - Duration: 从开始连接到操作完成的总耗时
type Result struct {
	Host     inventory.Host
	Success  bool
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
	Duration time.Duration
}

// Options 定义 executor 的运行选项。
//
// 字段说明：
//   - Concurrency: 控制同时执行操作的主机数量（并发度）。<= 0 时由 normalizeOptions 归一化为 1
//   - SSH: SSH 连接参数（用户、端口、密钥、超时等），由 normalizeSSHOptions 进一步归一化
type Options struct {
	// Concurrency 控制同时执行的主机数量
	Concurrency int

	// SSH 连接参数
	SSH SSHOptions
}

// Executor 是 basic 的执行引擎。
//
// Executor 只负责：
//
//	Host
//	  ↓
//	SSH
//	  ↓
//	Operation
//	  ↓
//	Result
//
// 不负责：
//
//	Cobra
//	Config
//	Logger
//	Printer
//
// 内部字段：
//   - options: 执行配置（并发数、SSH 参数），经 normalizeOptions 归一化
//   - ssh: SSH 连接接口，依赖接口而非具体实现，便于测试注入 mock
type Executor struct {
	options Options
	ssh     SSHConnection
}

// NewExecutor 创建一个 Executor 实例。
//
// 流程：
//  1. 调用 normalizeOptions 对传入的 options 做归一化
//     （Concurrency <= 0 置为 1，SSH 零值填充默认值）
//  2. 用归一化后的 SSH 配置创建 SSHClient 作为默认连接实现
func NewExecutor(options Options) *Executor {
	options = normalizeOptions(options)
	return &Executor{
		options: options,
		ssh:     NewSSHClient(options.SSH),
	}
}

// Options 返回 Executor 的配置副本。
//
// 防御性处理：若 receiver 为 nil，返回零值 Options{}，避免空指针解引用。
// 返回的是值拷贝，调用方修改不会影响 Executor 内部状态。
func (e *Executor) Options() Options {
	if e == nil {
		return Options{}
	}
	return e.options
}

// runWithConnector 创建一个使用自定义 SSHConnection 的 Executor。
//
// 主要用于单元测试：通过注入 mock 的 SSHConnection，
// 可以在不建立真实 SSH 连接的情况下测试 Run 的并发控制、
// 错误处理、Context 取消等逻辑。
//
// 与 NewExecutor 的区别：
//   - NewExecutor 使用默认的 SSHClient（真实 SSH 连接）
//   - runWithConnector 使用传入的 connector（可替换为 mock）
//
// 注意：此函数为包内可见（小写开头），仅供测试使用。
func runWithConnector(options Options, connector SSHConnection) *Executor {
	options = normalizeOptions(options)
	return &Executor{
		options: options,
		ssh:     connector,
	}
}

// Run 并发地在一组主机上执行 operation，并返回每个主机的执行结果。
//
// 参数：
//   - ctx: 上下文，用于取消/超时控制；传 nil 会使用 context.Background()
//   - hosts: 目标主机列表，空切片直接返回 nil
//   - operation: 具体业务操作（如 shell、copy、fetch 等），接收 ctx、host、已建立的 *ssh.Client，返回 Result
//
// 职责：
//   - 并发控制：通过带缓冲 channel 作为信号量（semaphore），限制同时执行的主机数为 e.options.Concurrency
//   - SSH 连接：为每个主机建立 SSH 连接，连接失败直接记录错误结果
//   - Context 感知：在排队等待槽位、执行前都会检查 ctx 是否取消，取消则直接返回 ctx.Err()
//   - Result 汇总：结果按 hosts 的原始下标填入 results，最终统一返回
//
// 注意：results 按下标写入（每个 goroutine 写自己的 index），不存在数据竞争；
// 所有 goroutine 结束后（wg.Wait()）才返回 results。
func (e *Executor) Run(ctx context.Context, hosts []inventory.Host, operation func(context.Context, inventory.Host, *ssh.Client) Result) []Result {
	// 防御：nil Receiver 直接返回
	if e == nil {
		return nil
	}
	// 无主机，无需执行
	if len(hosts) == 0 {
		return nil
	}
	// 未提供 ctx，使用空上下文
	if ctx == nil {
		ctx = context.Background()
	}
	// operation 为 nil 时，为每个主机生成错误结果，避免空指针调用
	if operation == nil {
		results := make([]Result, len(hosts))
		for i, host := range hosts {
			results[i] = Result{
				Host:  host,
				Error: fmt.Errorf("executor operation is nil"),
			}
		}
		return results
	}
	results := make([]Result, len(hosts))
	// Semaphore 控制并发数：容量为 Concurrency，每个 goroutine 执行前需获取一个槽位
	semaphore := make(chan struct{}, e.options.Concurrency)
	var wg sync.WaitGroup
	for index := range hosts {
		// 循环变量捕获：Go 1.22 之前循环变量在所有迭代间共享，需用局部变量固定当次迭代的值
		index := index
		host := hosts[index]
		wg.Add(1)
		//go func(index int, host inventory.Host) {}()
		go func() {
			defer wg.Done()

			// 进入时先检查 ctx 是否已取消，取消则直接记录错误并返回
			select {
			case <-ctx.Done():
				results[index] = Result{
					Host:  host,
					Error: ctx.Err(),
				}
				return
			default:
			}
			// 获取并发槽位；若槽位已满则阻塞，同时监听 ctx 取消以提前退出
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				results[index] = Result{
					Host:  host,
					Error: ctx.Err(),
				}
				return
			}
			// 执行完毕释放槽位，让其它等待中的 goroutine 继续
			defer func() {
				<-semaphore
			}()
			start := time.Now()

			// 建立 SSH 连接；失败则记录错误并返回（不调用 operation）
			client, err := e.ssh.Connect(host)
			if err != nil {
				results[index] = Result{
					Host:     host,
					Success:  false,
					Error:    err,
					Duration: time.Since(start),
				}
				return
			}
			if client == nil {
				results[index] = Result{
					Host:     host,
					Success:  false,
					Error:    fmt.Errorf("ssh client is  nil"),
					Duration: time.Since(start),
				}
				return
			}
			defer client.Close()
			// 执行具体业务操作，并补充 Host 与 Duration 字段
			result := operation(ctx, host, client)
			result.Host = host
			result.Duration = time.Since(start)
			results[index] = result
		}()
	}
	// 等待所有主机执行完成
	wg.Wait()
	return results
}

func normalizeOptions(options Options) Options {
	if options.Concurrency <= 0 {
		options.Concurrency = 1
	}
	options.SSH = normalizeSSHOptions(options.SSH)
	return options
}
