package launcher

import (
	"fmt"
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

// Builder 启动命令构造器
type Builder struct {
	Cfg      *config.Config
	Provider mirror.Provider
}

// BuildArgs 生成 java 命令参数
func (b *Builder) BuildArgs(ins *instance.Instance, vd *mojang.VersionDetail, javaPath string) []string {
	var args []string

	// 内存
	minMem := ins.MinMemMB
	if minMem == 0 {
		minMem = b.Cfg.MinMemoryMB
	}
	maxMem := ins.MaxMemMB
	if maxMem == 0 {
		maxMem = b.Cfg.MaxMemoryMB
	}
	args = append(args, fmt.Sprintf("-Xms%dM", minMem))
	args = append(args, fmt.Sprintf("-Xmx%dM", maxMem))

	// GC
	if b.Cfg.GCType == "ZGC" {
		args = append(args, "-XX:+UseZGC", "-XX:+ZGenerational")
	} else {
		args = append(args, "-XX:+UseG1GC")
	}
	args = append(args, "-XX:+UnlockExperimentalVMOptions", "-XX:+UseAdaptiveGCBoundary")

	// 优化
	args = append(args,
		"-XX:-UsePerfData",
		"-XX:+AggressiveOpts",
		"-Djava.net.preferIPv4Stack=true",
	)

	// 自定义 JVM 参数
	if ins.JVMArgs != "" {
		args = append(args, strings.Fields(ins.JVMArgs)...)
	}

	// classpath（简化：这里只放 client.jar + 已下载 libraries）
	gameDir := instanceGameDir(ins)
	versionsDir := filepath.Join(gameDir, "versions", vd.ID)
	clientJar := filepath.Join(versionsDir, vd.ID+".jar")

	var cp []string
	cp = append(cp, clientJar)
	libsDir := filepath.Join(gameDir, "libraries")
	// 递归收集所有 .jar
	_ = filepath.Walk(libsDir, func(p string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(p, ".jar") {
			cp = append(cp, p)
		}
		return nil
	})

	args = append(args, "-cp", strings.Join(cp, filepath.ListSeparator))
	args = append(args, vd.MainClass)

	// 游戏参数（简化版，完整版需解析 arguments.game）
	args = append(args,
		"--username", "Player",
		"--version", vd.ID,
		"--gameDir", gameDir,
		"--assetsDir", filepath.Join(gameDir, "assets"),
		"--assetIndex", vd.AssetIndex.ID,
		"--uuid", "00000000-0000-0000-0000-000000000000",
		"--accessToken", "0",
		"--userType", "legacy",
	)
	if ins.GameArgs != "" {
		args = append(args, strings.Fields(ins.GameArgs)...)
	}
	return args
}

// Launch 真正启动 MC
func (b *Builder) Launch(ins *instance.Instance, vd *mojang.VersionDetail, javaPath string) (*exec.Cmd, error) {
	if javaPath == "" {
		fj, err := java.NewProbe().FindBest()
		if err != nil {
			return nil, err
		}
		javaPath = fj.Path
	}
	args := b.BuildArgs(ins, vd, javaPath)
	cmd := exec.Command(javaPath, args...)
	cmd.Dir = instanceGameDir(ins)
	// Linux 下可能需要 DISPLAY，这里不处理
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &SysProcAttrNoWin{}
	}
	return cmd, nil
}

func instanceGameDir(ins *instance.Instance) string {
	base := filepath.Join(".", "instances", ins.ID)
	abs, _ := filepath.Abs(base)
	return abs
}

// SysProcAttrNoWin 占位
type SysProcAttrNoWin struct{}
