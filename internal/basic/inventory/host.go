package inventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	DefaultSSHPort = 22
	DefaultSSHUser = "root"
)

var groupRegex = regexp.MustCompile(`^\[([^\[\]]+)\]$`)

// Host 表示一个SSH目标主机
type Host struct {
	Address string
	User    string
	Port    int
	Passwd  string
}

func ParseFile(path string) ([]Host, error) {
	openFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open inventory file %q: %w", path, err)
	}
	defer openFile.Close()

	var (
		hosts        []Host
		currentGroup string
		lineNumber   int
	)

	scanner := bufio.NewScanner(openFile)
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// 空行
		if line == "" {
			continue
		}

		// 注释
		if strings.HasPrefix(line, "#") {
			continue
		}
		// group

		if submatch := groupRegex.FindStringSubmatch(line); submatch != nil {
			currentGroup = strings.TrimSpace(submatch[1])

			if currentGroup == "" {
				return nil, fmt.Errorf("inventory %q line %d: empty group name", path, lineNumber)
			}
			continue
		}
		hostLine, err := parseHostLine(line, currentGroup)
		if err != nil {
			return nil, fmt.Errorf("inventory %q line %d: %w", path, lineNumber, err)
		}
		hosts = append(hosts, hostLine)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan inventory file %q: %w", path, err)
	}

	return hosts, nil
}

func parseHostLine(line, group string) (Host, error) {
	parts := strings.Fields(line)

	if len(parts) == 0 {
		return Host{}, fmt.Errorf("empty host line")
	}

	host := Host{
		Address: parts[0],
		User:    DefaultSSHUser,
		Port:    DefaultSSHPort,
		//Group:   group,
	}

	if host.Address == "" {
		return Host{}, fmt.Errorf("empty host address")
	}
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return Host{}, fmt.Errorf("invalid paramenter %q expected key=value", part)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			return Host{}, fmt.Errorf("empty paramenter key in %q", part)
		}
		switch key {
		case "user":
			if value == "" {
				return Host{}, fmt.Errorf("user cannot be empty")
			}
			host.User = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil {
				return Host{}, fmt.Errorf("invalid port in %q : %w", value, err)
			}
			if port < 0 || port > 65535 {
				return Host{}, fmt.Errorf("port out of range: %d", port)
			}
			host.Port = port
		case "password":
			host.Passwd = value

		default:

			// 位置参数暂时忽略

			//TODO: fastdp 遇到：192.168.1.10 xxx 会触发警告然后哦忽略 xxx并且继续运行
			//      我建议 Kraken 改成：错误出现时指出 inventory 文件 + 行号 然后停止解析，因为 inventory 属于基础配置，配置写错最好尽早失败。
			continue
		}
	}
	return host, nil
}
