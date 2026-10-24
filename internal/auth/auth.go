package auth

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Account 登录账号
type Account struct {
	Type       string `json:"type"` // offline / microsoft / littleskin
	Username   string `json:"username"`
	UUID       string `json:"uuid"`
	AccessToken string `json:"accessToken,omitempty"`
	BaseURL    string `json:"baseUrl,omitempty"` // LittleSkin 域名
}

// Store 账号存储
type Store struct {
	path string
}

func NewStore(path string) *Store { return &Store{path: path} }

func (s *Store) Load() ([]*Account, error) {
	b, err := readFile(s.path)
	if err != nil {
		return nil, err
	}
	var list []*Account
	_ = json.Unmarshal(b, &list)
	return list, nil
}

func (s *Store) Save(list []*Account) error {
	b, _ := json.MarshalIndent(list, "", "  ")
	return writeFile(s.path, b)
}

// OfflineLogin 离线登录（生成随机 UUID）
func OfflineLogin(username string) *Account {
	return &Account{
		Type:     "offline",
		Username: username,
		UUID:     offlineUUID(username),
	}
}

func offlineUUID(username string) string {
	// 离线 UUID：MD5("OfflinePlayer:" + username) 的变体
	// 简化：随机 UUID
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// LittleSkinLogin 简易外置登录（BaseURL 由用户提供）
// 完整 OAuth 流程需要浏览器授权，这里预留结构
type LittleSkinSession struct {
	AccessToken string `json:"accessToken"`
	ClientToken string `json:"clientToken"`
	SelectedProfile struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"selectedProfile"`
}

// MicrosoftLoginURL 微软登录 URL（需要 azure 应用 ID，这里给占位流程）
func MicrosoftLoginURL(clientID string) string {
	return fmt.Sprintf(
		"https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize?client_id=%s&response_type=code&scope=XboxLive.signin%20offline_access&redirect_uri=http://localhost:18923/auth/microsoft",
		clientID,
	)
}

// PingHTTP 测试 URL 可达
func PingHTTP(url string) error {
	c := &http.Client{Timeout: 5 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("服务端错误 %d", resp.StatusCode)
	}
	return nil
}

// readFile / writeFile 薄包装
func readFile(p string) ([]byte, error) {
	if strings.HasPrefix(p, "~") {
		return nil, fmt.Errorf("路径需展开 ~")
	}
	return os.ReadFile(p)
}

func writeFile(p string, b []byte) error {
	return os.WriteFile(p, b, 0644)
}
