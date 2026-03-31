package shared

// Config 全局配置
type Config struct {
	AppName    string `json:"appName"`
	Version    string `json:"version"`
	DarkTheme  bool   `json:"darkTheme"`
	AIProvider string `json:"aiProvider"` // 默认AI提供商
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		AppName:    "个人投资分析平台",
		Version:    "0.1.0",
		DarkTheme:  true,
		AIProvider: "deepseek",
	}
}
