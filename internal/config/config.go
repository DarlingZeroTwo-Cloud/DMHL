package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Config DMHL 全局配置
type Config struct {
	DataDir       string `json:"dataDir"`       // 游戏数据目录
	JavaPath      string `json:"javaPath"`      // 自定义 Java 路径（空=自动探测）
	MaxMemoryMB   int    `json:"maxMemoryMB"`   // 最大堆内存
	MinMemoryMB   int    `json:"minMemoryMB"`   // 最小堆内存
	UseMirror     bool   `json:"useMirror"`     // 使用国内镜像
	GCType        string `json:"gcType"`        // G1 / ZGC
	ThemeColor    string `json:"themeColor"`    // 跟随系统=auto
	BackgroundURL string `json:"backgroundUrl"` // 自定义背景
	ServerPort    int    `json:"serverPort"`    // 本地 API 端口
	AutoUpdate    bool   `json:"autoUpdate"`
}

// Default 返回默认配置
func Default() *Config {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".dmhl")
	if runtime.GOOS == "windows" {
		dir = filepath.Join(home, "AppData", "Roaming", "DMHL")
	}
	return &Config{
		DataDir:     dir,
		MaxMemoryMB: 4096,
		MinMemoryMB: 1024,
		UseMirror:   true,
		GCType:      "G1",
		ThemeColor:  "auto",
		ServerPort:  18923,
		AutoUpdate:  true,
	}
}

// Load 从磁盘读配置
func Load(path string) (*Config, error) {
	c := Default()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			_ = Save(path, c)
			return c, nil
		}
		return nil, err
	}
	_ = json.Unmarshal(b, c)
	return c, nil
}

// Save 写配置
func Save(path string, c *Config) error {
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(path, b, 0644)
}

// ResolveMemory 根据物理内存推荐堆大小
func ResolveMemoryMB(totalMB int) (min, max int) {
	switch {
	case totalMB <= 4096:
		return 512, 1536
	case totalMB <= 8192:
		return 1024, 3072
	case totalMB <= 16384:
		return 1024, 6144
	default:
		return 2048, 8192
	}
}
