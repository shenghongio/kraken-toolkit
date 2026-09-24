package kubeadm

import (
	"context"
	"fmt"
)

// Tip:
//  - Phase 接口的四个方法分别对应：阶段名、描述、要执行的命令、依赖的前置阶段
//  - PhaseConfig 携带执行所需的全部上下文（配置文件路径、kubeadm 二进制路径、目标节点等）
//  - ExecuteFunc 是执行策略的抽象：本地调试传入本地执行函数，生产环境传入 SSH 远程执行函数

// Phase 是 kubeadm 单个阶段的统一抽象
// 每个kubeadm phase 都需要实现这个接口
type Phase interface {
	Name() string
	Description() string
	Command(cfg PhaseConfig) []string
	Dependencies() []string
}

// PhaseConfig 传递给 每个Phase 的执行上下文
type PhaseConfig struct {
	KubeadmConfigPath string
	KubeadmBinary     string
	NodeAddress       string
	DryRun            bool
	ExtraArgs         map[string]string
}

// ExcuteFunc 是实际执行命令的函数类型
// 由调用方注入，(本地执行或ssh远程执行)
type ExecuteFunc func(ctx context.Context, phase Phase, cfg PhaseConfig) error

// Tip:
//  - phases 用 map[string]Phase 存储，支持 O(1) 按名称查找
//  - order 是排序缓存——每次 Register 后失效，只有调用 Sorted() 时重新计算
//  - Register 允许覆盖，方便调试时替换某个阶段实现

// PhaseRegistry 管理阶段的注册，排序和批量执行
type PhaseRegistry struct {
	phases map[string]Phase
	order  []string // 拓扑排序缓存，注册新阶段后失效
}

// NewPhaseRegistry 创建一个新的阶段注册器
func NewPhaseRegistry() *PhaseRegistry {
	return &PhaseRegistry{
		phases: make(map[string]Phase),
	}
}

// Register 注册一个阶段，同名阶段会被覆盖，覆盖后排序缓存失效
func (r *PhaseRegistry) Register(p Phase) {
	r.phases[p.Name()] = p
	r.order = nil // 失效缓存
}

// Get 按名称获取单个阶段
func (r *PhaseRegistry) Get(name string) Phase {
	return r.phases[name]
}

// List 返回所有已注册阶段(未排序)
func (r *PhaseRegistry) List() []Phase {
	result := make([]Phase, 0, len(r.phases))
	for _, p := range r.phases {
		result = append(result, p)
	}
	return result
}

/*

核心算法，使用kahn 算法对阶段依赖DAG进行排序
	初始入度:    preflight=0, certs=1, etcd=1, kubeconfig=1
	第一轮:      取出 preflight（入度0）→ certs入度-1=0, kubelet-start入度-1=0
	第二轮:      取出 certs（入度0）→ etcd入度-1=0, kubeconfig入度-1=0, control-plane入度-1=0
	第三轮:      并行取出 etcd, kubeconfig, control-plane, kubelet-start（入度均为0）
	...
	如果最后 len(sorted) != len(r.phases)，说明有循环依赖（如 A→B→C→A），会返回错误。

*/

// Sorted 按拓扑顺序返回阶段列表
// 使用 Kahn算法，如果存在循环依赖则返回错误
func (r *PhaseRegistry) Sorted() ([]Phase, error) {
	if r.order != nil {
		return r.orderToPhases(), nil
	}
	
	// 计算每个节点的入度
	inDegree := make(map[string]int)
	for n := range r.phases {
		inDegree[n] = 0
	}
	for _, p := range r.phases {
		for _, dep := range p.Dependencies() {
			inDegree[p.Name()]++
			// 检查依赖的阶段是否存在
			if _, ok := r.phases[dep]; !ok {
				return nil, fmt.Errorf("phase %q depends on unknown phase %q", p.Name(), dep)
			}
		}
	}
	
	// kahn 算法： 不断取出入度为0的节点
	var sorted []string
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		sorted = append(sorted, name)
		
		// 减少所有后继口节点的入度
		for _, p := range r.phases {
			for _, dep := range p.Dependencies() {
				if dep == name {
					inDegree[p.Name()]--
					if inDegree[p.Name()] == 0 {
						queue = append(queue, p.Name())
					}
				}
			}
		}
	}
	
	// 如果有节点未被排序，说明存在循环依赖
	if len(sorted) != len(r.phases) {
		return nil, fmt.Errorf("circular dependency detected in phases")
	}
	r.order = sorted
	return r.orderToPhases(), nil
}

func (r *PhaseRegistry) orderToPhases() []Phase {
	result := make([]Phase, 0, len(r.order))
	for _, name := range r.order {
		result = append(result, r.phases[name])
	}
	return result
}

// RunAll 按照拓扑顺序执行所有节点，任意阶段失败则终止
func (r *PhaseRegistry) RunAll(ctx context.Context, cfg PhaseConfig, exec ExecuteFunc) error {
	sorted, err := r.Sorted()
	if err != nil {
		return fmt.Errorf("phase sorting failed %w", err)
	}
	for _, phase := range sorted {
		if cfg.DryRun {
			fmt.Printf(" [DRT-RUN] %-20s -> %s\n", phase.Name(), phase.Dependencies())
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := exec(ctx, phase, cfg); err != nil {
			return fmt.Errorf("phase %q failed %w", phase.Name(), err)
		}
	}
	return nil
}

// RunOne 执行指定名称的阶段机器所有前置依赖
func (r *PhaseRegistry) RunOne(ctx context.Context, name string, cfg PhaseConfig, exec ExecuteFunc) error {
	target := r.Get(name)
	if target == nil {
		return fmt.Errorf("phase %q not found", name)
	}
	// 收集目标阶段的所有依赖 (BFS)
	deps := r.collectDependencies(name)
	
	// 按拓扑顺序执行依赖链 + 目标阶段
	sorted, err := r.Sorted()
	if err != nil {
		return fmt.Errorf("phase sorting failed: %w", err)
	}
	
	// 过滤出需要执行的阶段
	deps[name] = true
	for _, phase := range sorted {
		if !deps[phase.Name()] {
			continue
		}
		if cfg.DryRun {
			fmt.Printf(" [DRT-RUN] %-20s -> %s\n", phase.Name(), phase.Dependencies())
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := exec(ctx, phase, cfg); err != nil {
			return fmt.Errorf("phase %q failed %w", phase.Name(), err)
		}
		
	}
	return nil
	
}

func (r *PhaseRegistry) collectDependencies(name string) map[string]bool {
	result := make(map[string]bool)
	queue := []string{name}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		p := r.Get(current)
		if p == nil {
			continue
		}
		for _, dep := range p.Dependencies() {
			if !result[dep] {
				result[dep] = true
				queue = append(queue, dep)
			}
		}
	}
	return result
}
