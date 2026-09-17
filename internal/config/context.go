package config

// contextKey 是包内私有的类型，防止其他包意外使用或修改
type contextKey string

// ContextKeyConfig 用于在 context 中存取 *Config
const ContextKeyConfig = contextKey("config")
