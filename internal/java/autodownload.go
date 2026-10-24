package java

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// DownloadedJRE 自动下载的 JRE 信息
type DownloadedJRE struct {
	MajorVersion int    `json:"majorVersion"`
	Path         string `json:"path"`
}

// JREDir DMHL 管理的 JRE 根目录
func JREDir() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Roaming", "DMHL", "runtime")
	}
	return filepath.Join(home, ".dmhl", "runtime")
}

// FindOrDownload 找 Java，找不到自动下载 Temurin JRE
func FindOrDownload(major int) (*FoundJava, error) {
	if major == 0 {
		major = 21
	}
	// 先扫本地已下载的
	if p := scanLocalJRE(major); p != "" {
		return &FoundJava{Path: p, Version: major}, nil
	}
	// 再扫系统
	p := NewProbe()
	if fj, err := p.FindByMajor(major); err == nil {
		return fj, nil
	}
	fmt.Printf("[DMHL] 未检测到 Java %d，开始自动下载 Temurin JRE...\n", major)
	jre, err := downloadJRE(major)
	if err != nil {
		return nil, fmt.Errorf("自动下载 JRE 失败: %w", err)
	}
	fmt.Printf("[DMHL] JRE 就绪: %s\n", jre.Path)
	return &FoundJava{Path: jre.Path, Version: major}, nil
}

func scanLocalJRE(major int) string {
	root := JREDir()
	binName := "java"
	if runtime.GOOS == "windows" {
		binName = "java.exe"
	}
	matches, _ := filepath.Glob(filepath.Join(root, fmt.Sprintf("%d", major), "*", "bin", binName))
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func downloadJRE(major int) (*DownloadedJRE, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	ext := ".tar.gz"
	if osName == "windows" {
		ext = ".zip"
	}
	// Adoptium API 拿真实下载地址
	apiURL := fmt.Sprintf(
		"https://api.adoptium.net/v3/binary/latest/%d/ga/%s/%s/jre/hotspot/normal/eclipse",
		major, osName, arch,
	)
	req, _ := http.NewRequest("HEAD", apiURL, nil)
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil }}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	finalURL := resp.Request.URL.String()
	resp.Body.Close()

	// 换华为云镜像
	mirrorURL := toHuaweicloud(finalURL)
	fmt.Printf("[DMHL] 下载地址: %s\n", mirrorURL)

	downloadDir := JREDir()
	_ = os.MkdirAll(downloadDir, 0755)
	archivePath := filepath.Join(downloadDir, fmt.Sprintf("jre%d%s", major, ext))

	if err := downloadFile(mirrorURL, archivePath); err != nil {
		return nil, err
	}

	extractDir := filepath.Join(downloadDir, fmt.Sprintf("%d", major))
	_ = os.MkdirAll(extractDir, 0755)
	if ext == ".zip" {
		if err := extractZip(archivePath, extractDir); err != nil {
			return nil, err
		}
	} else {
		if err := extractTarGz(archivePath, extractDir); err != nil {
			return nil, err
		}
	}

	binName := "java"
	if osName == "windows" {
		binName = "java.exe"
	}
	matches, _ := filepath.Glob(filepath.Join(extractDir, "*", "bin", binName))
	if len(matches) == 0 {
		return nil, fmt.Errorf("解压后未找到 java 可执行文件")
	}
	return &DownloadedJRE{MajorVersion: major, Path: matches[0]}, nil
}

// toHuaweicloud 把 Adoptium/GitHub 下载 URL 换成华为云镜像
func toHuaweicloud(original string) string {
	idx := strings.LastIndex(original, "/")
	if idx < 0 {
		return original
	}
	fileName := original[idx+1:]
	if fileName == "" {
		return original
	}
	return "https://mirrors.huaweicloud.com/adoptium/releases/" + fileName
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("下载失败 HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func extractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, hdr.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)) {
			continue
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(target, 0755)
		case tar.TypeReg:
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			out, _ := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			_, _ = io.Copy(out, tr)
			_ = out.Close()
		}
	}
	return nil
}

func extractZip(src, dest string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			fmt.Sprintf("Expand-Archive -LiteralPath '%s' -DestinationPath '%s' -Force", src, dest))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%w: %s", err, string(out))
		}
		return nil
	}
	cmd := exec.Command("unzip", "-o", src, "-d", dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}
