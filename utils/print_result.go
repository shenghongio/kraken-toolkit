package utils

import (
	"fmt"
	"github.com/kraken-pedestal/internal/basic/executor"
	"strings"
	"time"
)

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
