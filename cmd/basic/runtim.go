package basic

// loadBasicRuntime 加载 Basic 运行时。
//
// 负责：
//   1. 读取 Basic 配置文件
//   2. 合并本次 CLI 参数
//   3. 创建 internal/basic.Runtime
//
// 注意：
// 这里属于 cmd 层，所以可以读取 Cobra/CLI 参数；
// internal/basic 不应该反过来依赖 cmd/basic。

func loadBasicRuntime() ()
