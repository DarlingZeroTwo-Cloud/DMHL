package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ReleaseInfo GitHub/Gitee Release 信息
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

const currentVersion = "0.2.0"

// CheckGitHub 查 GitHub 最新 Release
func CheckGitHub() (*ReleaseInfo, error) {
	return check("https://api.github.com/repos/DarlingZeroTwo-Cloud/DMHL/releases/latest")
}

// CheckGitee 查 Gitee 最新 Release
func CheckGitee() (*ReleaseInfo, error) {
	return check("https://gitee.com/api/v5/repos/darling-cloud/DMHL/releases/latest")
}

func check(url string) (*ReleaseInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r ReleaseInfo
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}
	return &r, nil
}

// CurrentVersion 当前版本
func CurrentVersion() string { return currentVersion }

// NeedUpdate 比较版本号
func NeedUpdate(latest string) bool {
	return latest != "v"+currentVersion && latest != currentVersion
}
