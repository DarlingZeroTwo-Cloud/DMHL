package instance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// Instance 一个 MC 实例
type Instance struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	VersionID   string `json:"versionId"`   // 如 1.21.4
	Loader      string `json:"loader"`      // vanilla / fabric / forge / neoforge / quilt
	LoaderVer   string `json:"loaderVer"`   // Fabric 版本号
	JavaPath    string `json:"javaPath,omitempty"`
	MaxMemMB    int    `json:"maxMemMb"`
	MinMemMB    int    `json:"minMemMb"`
	JVMArgs     string `json:"jvmArgs,omitempty"`
	GameArgs    string `json:"gameArgs,omitempty"`
	Icon        string `json:"icon,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
	LastLaunch  int64  `json:"lastLaunch"`
}

// Manager 实例管理器
type Manager struct {
	BaseDir string // <dataDir>/instances
}

func New(baseDir string) *Manager {
	return &Manager{BaseDir: filepath.Join(baseDir, "instances")}
}

func (m *Manager) dir(id string) string {
	return filepath.Join(m.BaseDir, id)
}

func (m *Manager) metaFile(id string) string {
	return filepath.Join(m.dir(id), "instance.json")
}

// List 列出所有实例
func (m *Manager) List() ([]*Instance, error) {
	entries, err := os.ReadDir(m.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*Instance
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		b, err := os.ReadFile(m.metaFile(e.Name()))
		if err != nil {
			continue
		}
		var ins Instance
		if json.Unmarshal(b, &ins) == nil {
			out = append(out, &ins)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastLaunch > out[j].LastLaunch })
	return out, nil
}

// Create 新建实例
func (m *Manager) Create(ins *Instance) error {
	if err := os.MkdirAll(m.dir(ins.ID), 0755); err != nil {
		return err
	}
	// versions / mods / resourcepacks 子目录
	for _, sub := range []string{"versions", "mods", "resourcepacks", "saves", "shaderpacks"} {
		_ = os.MkdirAll(filepath.Join(m.dir(ins.ID), sub), 0755)
	}
	return m.Save(ins)
}

// Save 保存
func (m *Manager) Save(ins *Instance) error {
	_ = os.MkdirAll(m.dir(ins.ID), 0755)
	b, _ := json.MarshalIndent(ins, "", "  ")
	return os.WriteFile(m.metaFile(ins.ID), b, 0644)
}

// Get 取一个
func (m *Manager) Get(id string) (*Instance, error) {
	b, err := os.ReadFile(m.metaFile(id))
	if err != nil {
		return nil, err
	}
	var ins Instance
	if err := json.Unmarshal(b, &ins); err != nil {
		return nil, err
	}
	return &ins, nil
}

// Delete 删除
func (m *Manager) Delete(id string) error {
	return os.RemoveAll(m.dir(id))
}
