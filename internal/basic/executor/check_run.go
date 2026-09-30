package executor

import (
	"context"
	"sync"
	"time"
	
	"github.com/kraken-toolkit/internal/basic/check"
)

// CheckResult 单台主机单个检查项的结果
type CheckResult struct {
	Host    string        // 目标主机地址
	Check   string        // 检查项名称
	OK      bool          //结果判定 (true=通过,false=失败)
	Detail  string        // validate 返回的判定详情
	Error   error         // 连接或执行错误 （隐含首判时记录）
	Elapsed time.Duration // 耗时
}

// RunChecks 在每台主机上复用一条 SSH 连接，顺序执行全部检查项。
//
// 连接策略（省连接原则）：
//   - 每台主机只建立一次 SSH 连接（Connect 次数 = 主机数）
//   - 所有检查项复用同一条连接，各自开轻量 session 执行
//   - 连接失败 → 该主机所有检查项判 FAIL（隐含首判，见 README §6.2.1）
func RunChecks(ctx context.Context, hosts []Host, opts Options, checks []check.Check) []CheckResult {
	opts = normalizeOptions(opts)
	var mu sync.Mutex
	var all []CheckResult
	var wg sync.WaitGroup
	sem := make(chan struct{}, opts.Concurrency)
	for _, h := range hosts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// 并发槽位
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			
			// 连接建立一次，检查项全部复用
			client, err := NewSSHClient(opts.SSH).Connect(h)
			if err != nil {
				// 连接失败 -> 所有检查项纪录隐含首判 FAIL
				mu.Lock()
				for _, c := range checks {
					all = append(all, CheckResult{
						Host:  h.Address,
						Check: c.Name(),
						OK:    false,
						Error: err,
					})
				}
				mu.Unlock()
				return
			}
			defer client.Close()
			
			// 顺序执行所有检查项，复用同一连接
			for _, c := range checks {
				runChecks := RunShell(ctx, h, client, c.Command())
				checkResult := CheckResult{
					Host:  h.Address,
					Check: c.Name(),
				}
				if runChecks.Error != nil {
					checkResult.Error = runChecks.Error
				} else {
					checkResult.OK, checkResult.Detail = c.Validate(runChecks.Stdout)
				}
				checkResult.Elapsed = runChecks.Duration
				mu.Lock()
				all = append(all, checkResult)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return all
	
}
