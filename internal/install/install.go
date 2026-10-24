package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"dmhl/internal/download"
	"dmhl/internal/mirror"
	"dmhl/internal/mojang"
)

// Manager 版本安装器
type Manager struct {
	GameDir string            // 游戏根目录（含 versions/ libraries/ assets/）
	P       mirror.Provider
	Dl      *download.Manager
}

func New(gameDir string, p mirror.Provider) *Manager {
	return &Manager{
		GameDir: gameDir,
		P:       p,
		Dl:      download.New(),
	}
}

// AssetIndex 资源索引
type AssetIndex struct {
	Objects map[string]AssetObject `json:"objects"`
}

type AssetObject struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

// InstallVersion 完整安装一个 vanilla 版本
func (m *Manager) InstallVersion(vd *mojang.VersionDetail, onProgress func(string, int, int)) error {
	verID := vd.ID
	verDir := filepath.Join(m.GameDir, "versions", verID)
	_ = os.MkdirAll(verDir, 0755)

	// 1. 保存 version JSON
	vdJSON, _ := json.MarshalIndent(vd, "", "  ")
	if err := os.WriteFile(filepath.Join(verDir, verID+".json"), vdJSON, 0644); err != nil {
		return err
	}

	var items []download.Item

	// 2. client.jar
	items = append(items, download.Item{
		URL:  vd.Downloads.Client.URL,
		Dest: filepath.Join(verDir, verID+".jar"),
		SHA1: vd.Downloads.Client.SHA1,
		Size: vd.Downloads.Client.Size,
	})

	// 3. libraries
	for _, lib := range vd.Libraries {
		if lib.Downloads == nil || lib.Downloads.Artifact == nil {
			continue
		}
		// 检查 rules（OS 过滤）
		if !libAllowed(lib) {
			continue
		}
		art := lib.Downloads.Artifact
		items = append(items, download.Item{
			URL:  m.P.LibraryURL(art.Path),
			Dest: filepath.Join(m.GameDir, "libraries", art.Path),
			SHA1: art.SHA1,
		})

		// natives classifier
		if lib.Downloads.Classifiers != nil && lib.Downloads.Natives != nil {
			classifierKey := nativesKey(lib.Downloads.Natives)
			if class, ok := lib.Downloads.Classifiers[classifierKey]; ok {
				items = append(items, download.Item{
					URL:  m.P.LibraryURL(class.Path),
					Dest: filepath.Join(m.GameDir, "libraries", class.Path),
				})
			}
		}
	}

	// 4. asset index
	idxURL := vd.AssetIndex.URL
	if m.P.UseMirror {
		idxURL = strings.Replace(idxURL, "https://piston-meta.mojang.com", "https://bmclapi2.bangbang93.com", 1)
	}
	idxBytes, err := m.Dl.GetBytes(idxURL)
	if err != nil {
		return fmt.Errorf("下载 asset index 失败: %w", err)
	}
	idxDir := filepath.Join(m.GameDir, "assets", "indexes")
	_ = os.MkdirAll(idxDir, 0755)
	if err := os.WriteFile(filepath.Join(idxDir, vd.AssetIndex.ID+".json"), idxBytes, 0644); err != nil {
		return err
	}

	var ai AssetIndex
	if err := json.Unmarshal(idxBytes, &ai); err != nil {
		return err
	}

	// 5. assets objects
	for name, obj := range ai.Objects {
		_ = name
		items = append(items, download.Item{
			URL:  m.P.AssetURL(obj.Hash),
			Dest: filepath.Join(m.GameDir, "assets", "objects", obj.Hash[:2], obj.Hash),
			Size: obj.Size,
		})
	}

	if onProgress != nil {
		onProgress(fmt.Sprintf("开始下载 %d 个文件", len(items)), 0, len(items))
	}

	errs := m.Dl.Do(items, nil)
	if len(errs) > 0 {
		return fmt.Errorf("下载失败 %d 个文件，首个错误: %w", len(errs), errs[0])
	}
	return nil
}

func libAllowed(lib mojang.Library) bool {
	if len(lib.Rules) == 0 {
		return true
	}
	allowed := false
	for _, r := range lib.Rules {
		osName := r.OS.Name
		if osName == "" || osName == currentOS() {
			if r.Action == "allow" {
				allowed = true
			} else if r.Action == "disallow" {
				allowed = false
			}
		}
	}
	return allowed
}

func nativesKey(natives map[string]string) string {
	if v, ok := natives[currentOS()]; ok {
		return strings.ReplaceAll(v, "${arch}", "64")
	}
	return ""
}

func currentOS() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "osx"
	default:
		return "linux"
	}
}
