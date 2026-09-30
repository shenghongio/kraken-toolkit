package check

import "fmt"

//CheckItem 清单zhong 单个检查项，从本地文件加载，(.check.yaml的checks 元素)
type CheckItem struct {
	Name   string         `yaml:"name"`
	Type   string         `yaml:"type"`
	Params map[string]any `yaml:"params"`
}

//Check 表示一种可执行的检查类型
type Check interface {
	Name() string                                    // 检查项名称
	Command() string                                 // 目标主机执行的shell名，收集判定信息
	Validate(stdout string) (ok bool, datail string) //解析输出判定
}

// Factory 根据 params 构造 check
type Factory func(params map[string]any) Check

//Registry 按type 管理 check 构造工厂
type Registry struct {
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

func (r *Registry) Register(t string, factory Factory) {
	if r == nil || t == "" || factory == nil {
		return
	}
	r.factories[t] = factory
}
func (r *Registry) Build(item CheckItem) (Check, error) {
	if r == nil {
		return nil, fmt.Errorf("registry is nil")
	}
	factory, ok := r.factories[item.Type]
	if !ok {
		return nil, fmt.Errorf("unknown check type: %s", item.Type)
	}
	return factory(item.Params), nil
}
