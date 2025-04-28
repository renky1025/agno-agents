package config

import (
	"fmt"
	"os"
)

// Config 表示应用程序配置
type Config struct {
	// Kubeconfig 文件路径
	KubeconfigPath string
}

// NewConfig 从命令行参数创建配置
func NewConfig(kubeconfigPath string) *Config {
	return &Config{
		KubeconfigPath: kubeconfigPath,
	}
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	// 检查 kubeconfig 是否可访问
	if c.KubeconfigPath != "" {
		_, err := os.Stat(c.KubeconfigPath)
		if err != nil {
			return fmt.Errorf("无法访问 kubeconfig 文件: %w", err)
		}
	}
	return nil
}
