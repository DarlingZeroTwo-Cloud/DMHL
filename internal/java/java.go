package java

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Probe 探测本机可用 Java
type Probe struct {
	SearchPaths []string
}

// FoundJava 探测结果
type FoundJava struct {
	Path    string `json:"path"`
	Version int    `json:"version"` // 17 / 21 / 8 ...
}

// NewProbe 默认探测器
func NewProbe() *Probe {
	p := &Probe{}
	p.buildPaths()
	return p
}

func (p *Probe) buildPaths() {
	p.SearchPaths = []string{"java", "javaw"}
	if runtime.GOOS == "windows" {
		for _, root := range []string{`C:\Program Files\Java`, `C:\Program Files\Eclipse Adoptium`, `C:\Program Files\Microsoft\jdk-*`} {
			matches, _ := filepath.Glob(root)
			p.SearchPaths = append(p.SearchPaths, matches...)
		}
	} else {
		p.SearchPaths = append(p.SearchPaths, "/usr/bin/java", "/usr/local/bin/java")
	}
	if home, err := os.UserHomeDir(); err == nil {
		p.SearchPaths = append(p.SearchPaths, filepath.Join(home, ".dmhl", "runtime", "*"))
	}
}

// FindBest 找最高版本 Java
func (p *Probe) FindBest() (*FoundJava, error) {
	var best *FoundJava
	for _, path := range p.SearchPaths {
		ver, err := probeVersion(path)
		if err != nil {
			continue
		}
		fj := &FoundJava{Path: path, Version: ver}
		if best == nil || fj.Version > best.Version {
			best = fj
		}
	}
	if best == nil {
		return nil, &NoJavaError{}
	}
	return best, nil
}

// FindByMajor 找指定主版本
func (p *Probe) FindByMajor(major int) (*FoundJava, error) {
	for _, path := range p.SearchPaths {
		ver, err := probeVersion(path)
		if err != nil {
			continue
		}
		if ver == major {
			return &FoundJava{Path: path, Version: ver}, nil
		}
	}
	return nil, &NoJavaError{Major: major}
}

func probeVersion(path string) (int, error) {
	cmd := exec.Command(path, "-version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	s := string(out)
	// java version "17.0.1" 或 "21" 或 "1.8.0_381"
	if strings.Contains(s, `"1.8`) {
		return 8, nil
	}
	start := strings.Index(s, `"`)
	if start < 0 {
		return 0, nil
	}
	rest := s[start+1:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return 0, nil
	}
	verStr := rest[:end]
	parts := strings.Split(verStr, ".")
	if len(parts) == 0 {
		return 0, nil
	}
	var major int
	fmtSscanf(parts[0], &major)
	return major, nil
}

func fmtSscanf(s string, n *int) {
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		*n = *n*10 + int(c-'0')
	}
}

type NoJavaError struct{ Major int }

func (e *NoJavaError) Error() string {
	if e.Major > 0 {
		return "未找到 Java " + itoa(e.Major)
	}
	return "未找到任何 Java，请安装 JDK 17 或 21"
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [12]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
