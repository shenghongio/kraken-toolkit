package utils

import (
	"fmt"
	"github.com/kraken-pedestal/internal/basic/executor"
	"strings"
	"time"
)

// PrintResult 自动选择输出模式
//   - 所有结果输出都简短 [<60字符,<= 1行] -> 表格形式输出
//   - 大段输出 ->  分组详情模式
func PrintResult(results []executor.Result) bool {
	if len(results) == 0 {
		return true
	}

	// 判断是否适合表格模式
	tableMode := true
	for _, res := range results {
		if res.Stdout != "" {
			if strings.Count(res.Stdout, "\n") > 1 || len(res.Stdout) > 60 {
				tableMode = false
				break
			}
		}
	}
	if tableMode {
		printTable(results)
	} else {
		printVerbose(results)
	}
	for _, r := range results {
		if r.Error != nil {
			return false
		}
	}
	return true
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%ds", int(d/time.Minute), int(d%time.Second)%60)
}

func printTable(results []executor.Result) {
	//计算 host 列表
	maxHostLen := len("HOST")
	for _, r := range results {
		if len(r.Host.Address) > maxHostLen {
			maxHostLen = len(r.Host.Address)
		}
	}
	// 表头
	headerFormat := fmt.Sprintf("%%-%ds STATUS     DURATION     OUTPUT", maxHostLen)
	fmt.Printf(headerFormat+"\n", "HOST")
	fmt.Printf("%s  ------   --------  ------\n", strings.Repeat("-", maxHostLen))
	// 数据行
	var okCount, failCount int
	for _, r := range results {
		host := fmt.Sprintf("%-*s", maxHostLen, r.Host.Address)
		dur := formatDuration(r.Duration)

		if r.Error != nil {
			fmt.Printf("%s    FAIL     %-8s   %s\n", host, dur, r.Error)
			failCount++
		} else if r.Message != "" {
			fmt.Printf("%s    OK     %-8s   %s\n", host, dur, r.Message)
			okCount++
		} else if r.Stdout != "" {
			firstLine := strings.SplitN(strings.TrimRight(r.Stdout, "\n"), "\n", 2)[0]
			fmt.Printf("%s    OK     %-8s   %s\n", host, dur, firstLine)
		} else {
			fmt.Printf("%s    OK     %-8s", host, dur)
			okCount++
		}
	}
	// 汇总行
	total := float64(0)
	for _, r := range results {
		total += r.Duration.Seconds()
	}
	fmt.Printf("\n✓ %d succeeded | ✗ %d failed | total: %d (%.1fs)\n",
		okCount, failCount, len(results), total)
}

func printVerbose(results []executor.Result) {
	var okResults, failResults []executor.Result
	for _, r := range results {
		if r.Error != nil {
			failResults = append(failResults, r)
		} else {
			okResults = append(okResults, r)
		}
	}
	// 成功组
	if len(okResults) > 0 {
		fmt.Printf("━━━ Success ─────────────────────────────────\n")
		for _, r := range okResults {
			dur := formatDuration(r.Duration)
			if r.Message != "" {
				fmt.Printf("  %s  (%s)\n    %s\n", r.Host.Address, dur, r.Message)
			} else if r.Stdout != "" {
				fmt.Printf("  %s  (%s)\n", r.Host.Address, dur)
				for _, line := range strings.Split(strings.TrimRight(r.Stdout, "\n"), "\n") {
					fmt.Printf("    %s\n", line)
				}
			} else {
				fmt.Printf("  %s  (%s)\n", r.Host.Address, dur)
			}
		}
	}

	// 失败组
	if len(failResults) > 0 {
		if len(okResults) > 0 {
			fmt.Println()
		}
		fmt.Printf("━━━ Failed ──────────────────────────────────\n")
		for _, r := range failResults {
			dur := formatDuration(r.Duration)
			fmt.Printf("  %s  %s  (%s)\n", r.Host.Address, r.Error, dur)
		}
	}

	// 汇总行
	fmt.Printf("\n✓ %d succeeded | ✗ %d failed | total: %d\n",
		len(okResults), len(failResults), len(results))

}

func PrintResultAndCheck(results []executor.Result) error {
	PrintResult(results)
	for _, r := range results {
		if r.Error != nil {
			return fmt.Errorf("")
		}
	}
	return nil
}
