package modpack

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// MRPack Modrinth 整合包格式
type MRPack struct {
	Game           string `json:"game"`
	VersionID      string `json:"versionId"`
	Name           string `json:"name"`
	Summary        string `json:"summary"`
	Files          []MRPackFile `json:"files"`
	Dependencies   map[string]string `json:"dependencies"`
}

type MRPackFile struct {
	Path   string `json:"path"`
	Hashes struct {
		SHA1 string `json:"sha1"`
		SHA512 string `json:"sha512"`
	} `json:"hashes"`
	URL    string `json:"url"`
	FileSize int64 `json:"size"`
}

// ImportMRPack 导入 .mrpack 整合包
// zip 结构：
//   modrinth.index.json
//   overrides/   → 复制到实例目录
//   client-overrides/ → 客户端覆盖
func ImportMRPack(zipPath, instanceDir string) (*MRPack, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	var pack MRPack

	for _, f := range zr.File {
		switch {
		case f.Name == "modrinth.index.json":
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			rc.Close()
			if err := json.Unmarshal(b, &pack); err != nil {
				return nil, fmt.Errorf("解析 modrinth.index.json 失败: %w", err)
			}
		case strings.HasPrefix(f.Name, "overrides/"):
			rel := strings.TrimPrefix(f.Name, "overrides/")
			dest := filepath.Join(instanceDir, rel)
			if f.FileInfo().IsDir() {
				os.MkdirAll(dest, 0755)
				continue
			}
			if err := extractZipFile(f, dest); err != nil {
				return nil, err
			}
		}
	}

	// 下载整合包里的 mod 文件
	for _, file := range pack.Files {
		dest := filepath.Join(instanceDir, file.Path)
		if _, err := os.Stat(dest); err == nil {
			continue
		}
		os.MkdirAll(filepath.Dir(dest), 0755)
		resp, err := http.Get(file.URL)
		if err != nil {
			continue
		}
		f, _ := os.Create(dest)
		io.Copy(f, resp.Body)
		f.Close()
		resp.Body.Close()
	}

	return &pack, nil
}

func extractZipFile(f *zip.File, dest string) error {
	os.MkdirAll(filepath.Dir(dest), 0755)
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

// ListResourcePacks 列出实例的资源包
func ListResourcePacks(instanceDir string) ([]string, error) {
	dir := filepath.Join(instanceDir, "resourcepacks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".zip") {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// ListShaderPacks 列出实例的光影包
func ListShaderPacks(instanceDir string) ([]string, error) {
	dir := filepath.Join(instanceDir, "shaderpacks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}
