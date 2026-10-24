package modrinth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const base = "https://api.modrinth.com/v2"

// Project 搜索结果项
type Project struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Downloads   int      `json:"downloads"`
	IconURL     string   `json:"icon_url"`
	Versions    []string `json:"versions"`
	Loaders     []string `json:"loaders"`
}

// Version 一个版本
type Version struct {
	VersionNumber string `json:"version_number"`
	Changelog     string `json:"changelog"`
	GameVersions  []string `json:"game_versions"`
	Loaders       []string `json:"loaders"`
	Files         []struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
		Primary  bool   `json:"primary"`
	} `json:"files"`
}

// Search 搜索 mod
func Search(query, mcVersion, loader string, limit int) ([]Project, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	facets := `[["project_type:mod"]`
	if mcVersion != "" {
		facets += `,["versions:` + mcVersion + `"]`
	}
	if loader != "" {
		facets += `,["categories:` + loader + `"]`
	}
	facets += "]"
	u := fmt.Sprintf("%s/search?query=%s&limit=%d&facets=%s",
		base, url.QueryEscape(query), limit, url.QueryEscape(facets))
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r struct {
		Hits []Project `json:"hits"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("解析 Modrinth 失败: %w\n%s", err, string(b)[:200])
	}
	return r.Hits, nil
}

// ListVersions 拿某个 mod 的所有版本
func ListVersions(slug, mcVersion, loader string) ([]Version, error) {
	u := fmt.Sprintf("%s/project/%s/version?game_versions=%s&loaders=%s",
		base, slug, url.QueryEscape(`["`+mcVersion+`"]`), url.QueryEscape(`["`+loader+`"]`))
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var versions []Version
	if err := json.Unmarshal(b, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}
