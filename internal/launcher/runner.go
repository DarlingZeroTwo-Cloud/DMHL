package launcher

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"dmhl/internal/config"
	"dmhl/internal/instance"
	"dmhl/internal/java"
	"dmhl/internal/mirror"
	"dmhl/internal/mojang"
)

// Runner 真启动器
type Runner struct {
	Cfg       *config.Config
	Provider  mirror.Provider
	GlobalDir string // 全局共享 libraries/assets
}

// NewRunner 构造
func NewRunner(cfg *config.Config, p mirror.Provider) *Runner {
	return &Runner{Cfg: cfg, Provider: p, GlobalDir: filepath.Join(cfg.DataDir, "shared")}
}

// LaunchResult 启动结果
type LaunchResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	PID     int    `json:"pid,omitempty"`
	LogFile string `json:"logFile,omitempty"`
}

// Do 真启动 MC
func (r *Runner) Do(ins *instance.Instance, accountName string) (*LaunchResult, error) {
	gameDir := r.gameDir(ins)
	verDir := filepath.Join(gameDir, "versions", ins.VersionID)

	// 读 version JSON
	vdBytes, err := os.ReadFile(filepath.Join(verDir, ins.VersionID+".json"))
	if err != nil {
		return nil, fmt.Errorf("版本未安装: %w", err)
	}
	var vd mojang.VersionDetail
	if err := json.Unmarshal(vdBytes, &vd); err != nil {
		return nil, err
	}

	// 找 Java
	major := vd.JavaVersion.MajorVersion
	if major == 0 {
		major = 21
	}
	fj, err := java.FindOrDownload(major)
	if err != nil {
		return nil, err
	}

	// 解压 natives
	nativesDir := filepath.Join(verDir, "natives")
	_ = os.MkdirAll(nativesDir, 0755)
	if err := r.extractNatives(&vd, gameDir, nativesDir); err != nil {
		return nil, err
	}

	// 拼 classpath
	cp := r.classpath(&vd, gameDir, verDir)

	// 内存
	minMem := ins.MinMemMB
	if minMem == 0 {
		minMem = r.Cfg.MinMemoryMB
	}
	maxMem := ins.MaxMemMB
	if maxMem == 0 {
		maxMem = r.Cfg.MaxMemoryMB
	}

	args := []string{
		fmt.Sprintf("-Xms%dM", minMem),
		fmt.Sprintf("-Xmx%dM", maxMem),
		"-XX:+UnlockExperimentalVMOptions",
		"-XX:+UseG1GC",
		"-XX:-UsePerfData",
		"-Djava.library.path=" + nativesDir,
		"-cp", strings.Join(cp, string(filepath.ListSeparator)),
		vd.MainClass,
		"--username", accountName,
		"--version", ins.VersionID,
		"--gameDir", gameDir,
		"--assetsDir", filepath.Join(r.GlobalDir, "assets"),
		"--assetIndex", vd.AssetIndex.ID,
		"--uuid", "00000000-0000-0000-0000-000000000000",
		"--accessToken", "0",
		"--userType", "legacy",
	}

	if ins.WindowWidth > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", ins.WindowWidth))
	}
	if ins.WindowHeight > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", ins.WindowHeight))
	}
	if ins.ServerAddr != "" {
		args = append(args, "--server", ins.ServerAddr)
	}

	cmd := exec.Command(fj.Path, args...)
	cmd.Dir = gameDir

	// 日志文件
	logPath := filepath.Join(gameDir, "logs", "dmhl-launch.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	cmd.Stdout = io.MultiWriter(logFile, os.Stdout)
	cmd.Stderr = io.MultiWriter(logFile, os.Stderr)

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return nil, fmt.Errorf("启动失败: %w", err)
	}

	go func() {
		_ = cmd.Wait()
		logFile.Close()
	}()

	return &LaunchResult{
		OK:      true,
		Message: fmt.Sprintf("MC 已启动 (PID %d)", cmd.Process.Pid),
		PID:     cmd.Process.Pid,
		LogFile: logPath,
	}, nil
}

func (r *Runner) gameDir(ins *instance.Instance) string {
	return filepath.Join(r.Cfg.DataDir, "instances", ins.ID)
}

func (r *Runner) classpath(vd *mojang.VersionDetail, gameDir, verDir string) []string {
	var cp []string
	cp = append(cp, filepath.Join(verDir, vd.ID+".jar"))
	for _, lib := range vd.Libraries {
		if lib.Downloads == nil || lib.Downloads.Artifact == nil {
			continue
		}
		if !libAllowedOS(lib) {
			continue
		}
		cp = append(cp, filepath.Join(r.GlobalDir, "libraries", lib.Downloads.Artifact.Path))
	}
	return cp
}

func (r *Runner) extractNatives(vd *mojang.VersionDetail, gameDir, nativesDir string) error {
	for _, lib := range vd.Libraries {
		if lib.Downloads == nil || lib.Downloads.Natives == nil {
			continue
		}
		classifierKey := nativesKeyOS(lib.Downloads.Natives)
		if classifierKey == "" {
			continue
		}
		class, ok := lib.Downloads.Classifiers[classifierKey]
		if !ok {
			continue
		}
		jarPath := filepath.Join(r.GlobalDir, "libraries", class.Path)
		if _, err := os.Stat(jarPath); err != nil {
			continue
		}
		if err := unzipJar(jarPath, nativesDir); err != nil {
			return err
		}
	}
	return nil
}

func unzipJar(jarPath, dest string) error {
	zr, err := zip.OpenReader(jarPath)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		// 跳过 META-INF
		if strings.HasPrefix(f.Name, "META-INF/") {
			continue
		}
		outPath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(filepath.Clean(outPath), filepath.Clean(dest)) {
			continue
		}
		_ = os.MkdirAll(filepath.Dir(outPath), 0755)
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			rc.Close()
			return err
		}
		_, _ = io.Copy(out, rc)
		out.Close()
		rc.Close()
	}
	return nil
}

func libAllowedOS(lib mojang.Library) bool {
	if len(lib.Rules) == 0 {
		return true
	}
	allowed := false
	for _, rule := range lib.Rules {
		osName := rule.OS.Name
		if osName == "" || osName == currentOSName() {
			if rule.Action == "allow" {
				allowed = true
			} else {
				allowed = false
			}
		}
	}
	return allowed
}

func nativesKeyOS(natives map[string]string) string {
	if v, ok := natives[currentOSName()]; ok {
		return strings.ReplaceAll(v, "${arch}", "64")
	}
	return ""
}

func currentOSName() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "osx"
	default:
		return "linux"
	}
}
