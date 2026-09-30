package builtins

import (
	"fmt"
	
	"github.com/kraken-toolkit/internal/basic/check"
)

// 类型断言工具，提供 all builtins 复用 (yaml 数值可能是 int 或者float64)

// registry 全局检查注册表，供builtins的init()注册工厂
var registry = check.NewRegistry()

// Registry 返回全局注册标 (供命令层binding与聚合使用)
func Registry() *check.Registry { return registry }

//getString 从 params 取值,缺省或类型不符返回默认值
func getString(v any, def string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return def
}

//getFloat 从params取数值，兼容 float64/int 缺省返回def
func getFloat(v any, def float64) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case string:
		var f float64
		if _, err := fmt.Sscanf(t, "%f", &f); err == nil {
			return f
		}
	}
	return def
}

// getBool 从 params 取布尔，缺省 false。
func getBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "on"
	}
	return false
}
