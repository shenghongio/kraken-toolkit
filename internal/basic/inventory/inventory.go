package inventory

import (
	"fmt"
	"strings"
)

// Inventory 表示已经加载的主机清单
type Inventory struct {
	hosts []Host
}

// Load 从文件加载 Inventory
func Load(path string) (*Inventory, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("inventory path is empty")
	}
	hosts, err := ParseFile(path)
	if err != nil {
		return nil, err
	}
	return &Inventory{hosts: hosts}, nil
}

// Hosts ,返回inventory 中的全主机
// 返回副本，避免调用方直接修改Inventory内部数据
func (i *Inventory) Hosts() []Host {
	if i == nil || len(i.hosts) == 0 {
		return nil
	}
	result := make([]Host, len(i.hosts))
	copy(result, i.hosts)
	return result
}

// Resolve 根据target解析目标主机
// target可以是一下：
//
//	"group"
//	"host"
//	"all"
//
// 例如:
//
//	Resolve([]string{"web"})
//	Resolve([]string{"192.168.0.1"})
//	Resolve([]string{"web","db"})
//	Resolve([]string{"all"})
func (i *Inventory) Resolve(targets []string) ([]Host, error) {
	if i == nil || len(targets) == 0 {
		return nil, fmt.Errorf("inventory is nil and no inventory target specified")
	}

	groupMap := make(map[string][]Host)
	hostMap := make(map[string][]Host)

	for _, host := range i.hosts {
		if host.Group != "" {
			groupMap[host.Group] = append(groupMap[host.Group], host)
		}
		hostMap[host.Address] = append(hostMap[host.Address], host)
	}

	var result []Host

	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		// all优先级最高
		if target == "all" {
			return Deduplicate(i.Hosts()), nil
		}

		// 优先匹配 group
		if hosts, ok := groupMap[target]; ok {
			result = append(result, hosts...)
			continue
		}

		// 再匹配 host
		if hosts, ok := hostMap[target]; ok {
			result = append(result, hosts...)
			continue
		}
		return nil, fmt.Errorf("unknown inventory target: %q", target)
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid inventory  targets specified")
	}
	return Deduplicate(result), nil
}

// Deduplicate 对主机进行去重，不能单纯使用address 去重，因为:
//
//	root@10.0.0.0:22
//	admin@10.0.0.0:22
//
// 可能是两个不同的SSH endpoint
func Deduplicate(hosts []Host) []Host {
	if len(hosts) <= 1 {
		return hosts
	}

	seen := make(map[string]struct{}, len(hosts))
	result := make([]Host, 0, len(hosts))

	for _, host := range hosts {
		key := hostKey(host)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, host)
	}
	return result
}

func hostKey(host Host) string {
	return fmt.Sprintf("%s@%s:%d", host.User, host.Address, host.Port)
}
