package fabric

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"dmhl/internal/mirror"
)

// LoaderEntry Fabric loader 版本
type LoaderEntry struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

// LoaderList 从 meta.fabricmc.net 返回的列表是数组
type LoaderList []struct {
	Loader struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	} `json:"loader"`
}

// Profile Fabric profile（合并到原版）
type Profile struct {
	ID         string `json:"id"`
	MainClass  string `json:"mainClass"`
	Libraries  []struct {
		Name string `json:"name"`
		URL  string `json:"url,omitempty"`
	} `json:"libraries"`
	Arguments struct {
		JVM    []string `json:"jvm"`
		Game   []string `json:"game"`
	} `json:"arguments"`
}

// ListLoaders 查某 MC 版本可用的 Fabric loader
func ListLoaders(mcVersion string, p mirror.Provider) ([]LoaderEntry, error) {
	url := "https://meta.fabricmc.net/v2/versions/loader/" + mcVersion
	if p.UseMirror {
		url = "https://bmclapi2.bangbang93.com/fabric-meta/v2/versions/loader/" + mcVersion
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var raw LoaderList
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("解析 Fabric loader 列表失败: %w\n%s", err, string(b)[:200])
	}
	out := make([]LoaderEntry, 0, len(raw))
	for _, item := range raw {
		out = append(out, LoaderEntry{Version: item.Loader.Version, Stable: item.Loader.Stable})
	}
	return out, nil
}

// FetchProfile 拿 Fabric profile JSON
func FetchProfile(mcVersion, loaderVersion string, p mirror.Provider) (*Profile, error) {
	url := fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json", mcVersion, loaderVersion)
	if p.UseMirror {
		url = fmt.Sprintf("https://bmclapi2.bangbang93.com/fabric-meta/v2/versions/loader/%s/%s/profile/json", mcVersion, loaderVersion)
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var profile Profile
	if err := json.Unmarshal(b, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// MavenCoordToPath 把 maven coordinate 转成 path
// net.fabricmc:fabric-loader:0.16.5 -> net/fabricmc/fabric-loader/0.16.5/fabric-loader-0.16.5.jar
func MavenCoordToPath(coord string) string {
	parts := strings.Split(coord, ":")
	if len(parts) < 3 {
		return ""
	}
	group := strings.ReplaceAll(parts[0], ".", "/")
	name := parts[1]
	version := parts[2]
	return fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, name, version, name, version)
}
