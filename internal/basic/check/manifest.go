package check

import (
	"fmt"
	"log/slog"
	"os"
	
	"gopkg.in/yaml.v3"
)

// 默认清单内容，仅作为首次生成 .check.yaml 的初始值
const defaultManifestYAML = `# .check.yaml
# 删除某检查项 = 不执行；修改 params = 调整阈值
checks:
  - name: ssh
    type: ssh
    params:
      hostkey_algo: ed25519
      password_auth: "no"
      permit_root_login: prohibit-password
  - name: disk-root
    type: disk
    params:
      path: "/"
      min_free_gb: 5
  - name: memory-total
    type: memory
    params:
      min_gb: 2
`

// Manifest 清单结构(.check.yaml 顶层)
type Manifest struct {
	Checks []CheckItem `yaml:"checks"`
}

// LoadManifest 加载清单检查项
func LoadManifest(path string) (*Manifest, error) {
	if path == "" {
		path = ".check.yaml"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := ensureDefaultFile(path); err != nil {
				return nil, err
			}
			slog.Debug("generated default", "file", path, "error", err)
		}
	}
	readFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read check file failed %q: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(readFile, &m); err != nil {
		return nil, fmt.Errorf("unmarshal check file failed %q: %w", path, err)
	}
	return &m, nil
}

// ensureDefaultFile 兜底生成默认清单文件
func ensureDefaultFile(path string) error {
	if err := os.WriteFile(path, []byte(defaultManifestYAML), 0o644); err != nil {
		return fmt.Errorf("write file failed %q: %w", path, err)
	}
	return nil
}
