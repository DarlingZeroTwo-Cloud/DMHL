package mojang

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dmhl/internal/mirror"
)

// VersionManifest Mojang 版本清单
type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []VersionEntry `json:"versions"`
}

// VersionEntry 单个版本条目
type VersionEntry struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // release / snapshot / old_beta / old_alpha
	URL         string    `json:"url"`
	Time        time.Time `json:"time"`
	ReleaseTime time.Time `json:"releaseTime"`
}

// FetchManifest 拉取版本清单
func FetchManifest(p mirror.Provider) (*VersionManifest, error) {
	c := &http.Client{Timeout: 30 * time.Second}
	resp, err := c.Get(p.VersionManifestURL())
	if err != nil {
		return nil, fmt.Errorf("拉取版本清单失败: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var m VersionManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("解析版本清单失败: %w", err)
	}
	return &m, nil
}

// VersionDetail 版本详情（启动时需要）
type VersionDetail struct {
	ID                 string `json:"id"`
	MainClass          string `json:"mainClass"`
	MinecraftArguments string `json:"minecraftArguments,omitempty"`
	Arguments          *struct {
		Game    []json.RawMessage `json:"game"`
		JVM     []json.RawMessage `json:"jvm"`
	} `json:"arguments"`
	Libraries []Library `json:"libraries"`
	AssetIndex struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Hash string `json:"sha1"`
	} `json:"assetIndex"`
	Downloads struct {
		Client struct {
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
			Size int64  `json:"size"`
		} `json:"client"`
	} `json:"downloads"`
	JavaVersion struct {
		MajorVersion int    `json:"majorVersion"`
		Component    string `json:"component"`
	} `json:"javaVersion"`
}

// Library 依赖库
type Library struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	Downloads *struct {
		Artifact *struct {
			Path string `json:"path"`
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
		} `json:"artifact"`
		Classifiers map[string]struct {
			Path string `json:"path"`
			URL  string `json:"url"`
		} `json:"classifiers,omitempty"`
		Natives map[string]string `json:"natives,omitempty"`
	} `json:"downloads,omitempty"`
	Rules []struct {
		Action string `json:"action"`
		OS     struct {
			Name string `json:"name"`
		} `json:"os,omitempty"`
	} `json:"rules,omitempty"`
}

// FetchVersionDetail 拉取版本详情
func FetchVersionDetail(url string, p mirror.Provider) (*VersionDetail, error) {
	c := &http.Client{Timeout: 30 * time.Second}
	finalURL := p.VersionJSONURL(url)
	resp, err := c.Get(finalURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var vd VersionDetail
	if err := json.Unmarshal(b, &vd); err != nil {
		return nil, fmt.Errorf("解析版本详情失败: %w\n%s", err, string(b)[:200])
	}
	return &vd, nil
}
