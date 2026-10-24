package skin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// Manager 皮肤管理器
type Manager struct {
	BaseDir string // <dataDir>/skins
}

func New(baseDir string) *Manager {
	return &Manager{BaseDir: filepath.Join(baseDir, "skins")}
}

// ImportPNG 从 PNG 文件导入皮肤
func (m *Manager) ImportPNG(srcPath, playerUUID string) (string, error) {
	_ = os.MkdirAll(m.BaseDir, 0755)
	b, err := os.ReadFile(srcPath)
	if err != nil {
		return "", err
	}
	// 校验是 PNG
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("不是有效的 PNG: %w", err)
	}
	_ = img
	// 皮肤规格：64x64 或 64x32
	outPath := filepath.Join(m.BaseDir, playerUUID+".png")
	if err := os.WriteFile(outPath, b, 0644); err != nil {
		return "", err
	}
	return outPath, nil
}

// GetSkinPath 取玩家皮肤路径
func (m *Manager) GetSkinPath(playerUUID string) string {
	p := filepath.Join(m.BaseDir, playerUUID+".png")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

// PreviewData 预览用（base64）
func (m *Manager) PreviewData(playerUUID string) (string, error) {
	p := m.GetSkinPath(playerUUID)
	if p == "" {
		return "", fmt.Errorf("未设置皮肤")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b), nil
}

// ApplyToInstance 把皮肤注入到实例（离线模式通过 authlib-injector 或本地替换）
// 简化：返回要加的 JVM 参数
func (m *Manager) ApplyToInstance(playerUUID string) []string {
	// 离线模式皮肤需要 authlib-injector + 本地皮肤 API
	// 这里返回占位参数，后续接入
	return []string{
		// "-javaagent:authlib-injector.jar=https://littleskin.cn/api/yggdrasil/",
	}
}

// DecodePNG 解码尺寸
func DecodePNG(path string) (image.Rectangle, error) {
	f, err := os.Open(path)
	if err != nil {
		return image.Rectangle{}, err
	}
	defer f.Close()
	img, err := png.DecodeConfig(f)
	if err != nil {
		return image.Rectangle{}, err
	}
	return image.Rect(0, 0, img.Width, img.Height), nil
}

// SkinMeta 皮肤元信息
type SkinMeta struct {
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Model  string `json:"model"` // slim / default
}

// List 列出所有皮肤
func (m *Manager) List() ([]*SkinMeta, error) {
	entries, err := os.ReadDir(m.BaseDir)
	if err != nil {
		return nil, err
	}
	var out []*SkinMeta
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".png" {
			continue
		}
		out = append(out, &SkinMeta{
			UUID:  e.Name()[:len(e.Name())-4],
			Path:  filepath.Join(m.BaseDir, e.Name()),
			Model: "default",
		})
	}
	return out, nil
}

// 保证 json 被使用
var _ = json.Marshal
