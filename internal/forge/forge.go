package forge

import (
	"dmhl/internal/download"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func get(url string) ([]byte, error) {
	m := download.New()
	return m.GetBytes(url)
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// VersionList Forge/NeoForge 版本清单
type VersionList struct {
	Versions []string `json:"versions"`
}

// ListForgeVersions 获取 Forge 版本列表（BMCLAPI）
func ListForgeVersions(mcVersion string) ([]string, error) {
	url := fmt.Sprintf("https://bmclapi2.bangbang93.com/forge/minecraft/%s", mcVersion)
	body, err := get(url)
	if err != nil {
		return nil, err
	}
	var raw []struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	versions := []string{}
	for _, v := range raw {
		versions = append(versions, v.Version)
	}
	return versions, nil
}

// ListNeoForgeVersions NeoForge 版本
func ListNeoForgeVersions(mcVersion string) ([]string, error) {
	// NeoForge 用 NeoForged API
	url := "https://maven.neoforged.net/api/maven/details/releases/net/neoforged/neoforge"
	body, err := get(url)
	if err != nil {
		return nil, err
	}
	var r struct {
		Versions []string `json:"versions"`
	}
	json.Unmarshal(body, &r)
	// 过滤匹配 mc 版本
	out := []string{}
	prefix := mcVersion
	for _, v := range r.Versions {
		// NeoForge 版本号如 21.4.190-beta, mc 1.21.4 → 21.4
		if len(v) > len(prefix) && v[:len(prefix)] == prefix {
			out = append(out, v)
		}
	}
	return out, nil
}

// InstallForge 下载 Forge installer 并跑 --installClient
// mcDir: .minecraft 目录（含 versions/libraries）
// 返回安装后的 version id
func InstallForge(mcVersion, forgeVersion, mcDir, javaPath string) (string, error) {
	// 下载 installer.jar
	installerURL := fmt.Sprintf("https://bmclapi2.bangbang93.com/forge/download?mc=%s&forge=%s&format=jar",
		mcVersion, forgeVersion)
	installerJar := filepath.Join(os.TempDir(), "forge-installer.jar")
	if err := downloadFile(installerURL, installerJar); err != nil {
		return "", fmt.Errorf("下载 installer 失败: %w", err)
	}
	// 跑 java -jar installer.jar --installClient <mcDir>
	cmd := exec.Command(javaPath, "-jar", installerJar, "--installClient", mcDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Forge installer 执行失败: %w", err)
	}
	// installer 生成了 versions/<mcVersion>-forge-<forgeVersion>/
	versionID := fmt.Sprintf("%s-forge-%s", mcVersion, forgeVersion)
	return versionID, nil
}

// InstallNeoForge NeoForge 安装（类似）
func InstallNeoForge(mcVersion, neoVersion, mcDir, javaPath string) (string, error) {
	// NeoForge installer 从 maven 下载
	installerURL := fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar",
		neoVersion, neoVersion)
	installerJar := filepath.Join(os.TempDir(), "neoforge-installer.jar")
	if err := downloadFile(installerURL, installerJar); err != nil {
		return "", fmt.Errorf("下载 NeoForge installer 失败: %w", err)
	}
	cmd := exec.Command(javaPath, "-jar", installerJar, "--installClient", mcDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("NeoForge installer 执行失败: %w", err)
	}
	versionID := fmt.Sprintf("%s-neoforge-%s", mcVersion, neoVersion)
	return versionID, nil
}
