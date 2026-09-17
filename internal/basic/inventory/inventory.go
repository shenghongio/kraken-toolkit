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

// Resolve 根据目标字符串列表解析出对应的主机集合。
//
// 每个 target 会按以下优先级依次匹配，命中即停止该 target 的解析：
//
//  1. "all"：匹配 inventory 中的全部主机，优先级最高，一旦出现会直接返回
//     去重后的全部主机，忽略其它 target。
//  2. 分组名（Group）：按 host.Group 匹配，返回该组内的所有主机。
//  3. 主机地址（Address）：按 host.Address 匹配，返回地址对应的所有主机。
//
// 多个 target 命中的主机会合并后再去重（参见 Deduplicate）。
//
// 参数:
//   - targets: 目标字符串列表，支持 "group"、"host"、"all" 三种形式，
//     元素会被去除首尾空白；空白元素会被跳过。
//
// 返回值:
//   - []Host: 解析得到的去重后的主机列表。
//   - error: 当 inventory 为 nil、targets 为空、或某个 target 既不是分组
//     也不是已知主机地址时返回错误。
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
	// groupMap: 以分组名(Group)为 key，映射到该分组下的所有主机
	groupMap := make(map[string][]Host)
	// hostMap: 以主机地址(Address)为 key，映射到该地址对应的所有主机
	hostMap := make(map[string][]Host)

	// 遍历所有主机，构建分组索引和地址索引
	for _, host := range i.hosts {
		// 主机属于某个分组时，加入分组索引
		//if host.Group != "" {
		//	groupMap[host.Group] = append(groupMap[host.Group], host)
		//}
		// 无论是否有分组，都按地址加入地址索引
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

// NewFromHosts 从主机列表创建 Inventory 实例
func NewFromHosts(host []Host) *Inventory {
	return &Inventory{hosts: host}
}
