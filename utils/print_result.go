package utils

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/kraken-toolkit/internal/basic/executor"
)

type checkRow struct {
	ok     bool
	datail string
}

func PrintDisplayResult(results []executor.Result) bool {
	if len(results) == 0 {
		return true
	}
	
	var okCount, failCount int
	for _, r := range results {
		dur := formatDisplayDuration(r.Duration)
		ts := time.Now().Format("15:04:05")
		
		if r.Error != nil {
			errMsg := strings.ReplaceAll(r.Error.Error(), "\n", " ")
			fmt.Printf("%s  ✗  %-15s  %-6s  %s\n", ts, r.Host.Address, dur, errMsg)
			failCount++
			continue
		}
		
		okCount++
		output := r.Stdout
		if output == "" && r.Message != "" {
			output = r.Message
		}
		
		if output == "" && r.Stderr == "" {
			fmt.Printf("%s  ✓  %-15s  %-6s\n", ts, r.Host.Address, dur)
			continue
		}
		
		// 合并 stdout + stderr
		var allLines []string
		if output != "" {
			allLines = append(allLines, strings.Split(strings.TrimRight(output, "\n"), "\n")...)
		}
		if r.Stderr != "" {
			if len(allLines) > 0 {
				allLines = append(allLines, "") // 分隔行
			}
			allLines = append(allLines, strings.Split(strings.TrimRight(r.Stderr, "\n"), "\n")...)
		}
		
		if len(allLines) <= 1 {
			line := ""
			if len(allLines) > 0 {
				line = allLines[0]
			}
			fmt.Printf("%s  ✓  %-15s  %-6s  %s\n", ts, r.Host.Address, dur, line)
		} else {
			fmt.Printf("%s  ✓  %-15s  %-6s\n", ts, r.Host.Address, dur)
			for _, line := range allLines {
				if line == "" {
					fmt.Println()
				} else {
					fmt.Printf("  %s\n", line)
				}
			}
		}
	}
	
	// 汇总行
	total := 0.0
	for _, r := range results {
		total += r.Duration.Seconds()
	}
	fmt.Printf("\n✓ %d  ✗ %d  |  total: %d  (%.1fs)\n",
		okCount, failCount, len(results), total)
	
	return failCount == 0
}

func formatDisplayDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%ds", int(d/time.Minute), int(d%time.Second)%60)
}

func PrintResultAndCheck(results []executor.Result) error {
	PrintDisplayResult(results)
	for _, r := range results {
		if r.Error != nil {
			return fmt.Errorf("")
		}
	}
	return nil
}

// PrintCheckResults 以 host×check 矩阵打印检查结果；失败项连续汇总。
func PrintCheckResults(results []executor.CheckResult) error {
	if len(results) == 0 {
		return nil
	}
	
	// 收集唯一 host 与 check 顺序
	var hosts []string
	var checkOrder []string
	hostSeen := map[string]bool{}
	checkSeen := map[string]bool{}
	matrix := map[string]map[string]checkRow{}
	
	for _, r := range results {
		if !hostSeen[r.Host] {
			hostSeen[r.Host] = true
			hosts = append(hosts, r.Host)
		}
		if !checkSeen[r.Check] {
			checkSeen[r.Check] = true
			checkOrder = append(checkOrder, r.Check)
		}
		if matrix[r.Host] == nil {
			matrix[r.Host] = map[string]checkRow{}
		}
		datail := r.Detail
		if r.Error != nil {
			datail = r.Error.Error()
		}
		matrix[r.Host][r.Check] = checkRow{ok: r.OK, datail: datail}
	}
	
	// 表头
	fmt.Printf("%-18s", "host")
	for _, c := range checkOrder {
		fmt.Printf("  %-7s", c)
	}
	fmt.Println()
	
	// 数据行
	failCount := 0
	for _, h := range hosts {
		fmt.Printf("%-18s", h)
		for _, c := range checkOrder {
			row := matrix[h][c]
			if row.ok {
				fmt.Printf("  %-7s", "PASS")
			} else {
				fmt.Printf("  %-7s", "FAIL")
				failCount++
			}
		}
		fmt.Println()
	}
	
	// 失败详情
	if failCount > 0 {
		fmt.Println("\n失败详情: ")
		for _, h := range hosts {
			for _, c := range checkOrder {
				row := matrix[h][c]
				if !row.ok {
					fmt.Printf("  %s / %s: %s\n", h, c, row.datail)
				}
			}
		}
		return fmt.Errorf("%d check(s) failed", failCount)
	}
	fmt.Println("\na checks passed")
	return nil
}
