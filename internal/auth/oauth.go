package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// MicrosoftOAuth 微软登录流程
// 1. 浏览器打开授权 URL
// 2. 回调 localhost:18923/auth/microsoft?code=xxx
// 3. code 换 token → Xbox Live → XSTS → Minecraft profile

const (
	msClientID    = "00000000402b5328" // HMCL 用的公开 client_id
	msRedirectURI = "http://localhost:18923/auth/microsoft"
)

// MicrosoftAuthURL 生成授权 URL
func MicrosoftAuthURL() string {
	return fmt.Sprintf(
		"https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize?client_id=%s&response_type=code&scope=XboxLive.signin%%20offline_access&redirect_uri=%s",
		msClientID, msRedirectURI,
	)
}

// ExchangeMicrosoftCode 用 code 换 Minecraft 账号
func ExchangeMicrosoftCode(code string) (*Account, error) {
	// 1. code → access_token
	tokenURL := "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"
	form := url.Values{
		"client_id":  {msClientID},
		"code":       {code},
		"grant_type": {"authorization_code"},
		"redirect_uri": {msRedirectURI},
	}
	resp, err := http.PostForm(tokenURL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, err
	}
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("未拿到 access_token")
	}

	// 2. Xbox Live
	xblBody, _ := json.Marshal(map[string]any{
		"Properties": map[string]string{
			"AuthMethod": "RPS",
			"SiteName":   "user.auth.xboxlive.com",
			"RpsTicket":  "d=" + tok.AccessToken,
		},
		"RelyingParty": "http://auth.xboxlive.com",
		"TokenType":    "JWT",
	})
	resp2, err := http.Post("https://user.auth.xboxlive.com/user/authenticate", "application/json", bytes.NewReader(xblBody))
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()
	var xbl struct {
		Token             string `json:"Token"`
		DisplayClaims     map[string][]map[string]string `json:"DisplayClaims"`
	}
	json.NewDecoder(resp2.Body).Decode(&xbl)
	if xbl.Token == "" {
		return nil, fmt.Errorf("Xbox Live 认证失败（可能是 XGPU 账号）")
	}

	// 3. XSTS
	xstsBody, _ := json.Marshal(map[string]any{
		"Properties": map[string]string{
			"SandboxId": "RETAIL",
			"UserToken": xbl.Token,
		},
		"RelyingParty": "rp://api.minecraftservices.com/",
		"TokenType":    "JWT",
	})
	resp3, err := http.Post("https://xsts.auth.xboxlive.com/xsts/authorize", "application/json", bytes.NewReader(xstsBody))
	if err != nil {
		return nil, err
	}
	defer resp3.Body.Close()
	var xsts struct {
		Token string `json:"Token"`
	}
	json.NewDecoder(resp3.Body).Decode(&xsts)

	// 4. Minecraft login
	uhs := xbl.DisplayClaims["xui"][0]["uhs"]
	mcBody, _ := json.Marshal(map[string]string{
		"identityToken": "XBL3.0 x=" + uhs + ";" + xsts.Token,
	})
	resp4, err := http.Post("https://api.minecraftservices.com/authentication/login_with_xbox", "application/json", bytes.NewReader(mcBody))
	if err != nil {
		return nil, err
	}
	defer resp4.Body.Close()
	var mcTok struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp4.Body).Decode(&mcTok)

	// 5. Minecraft profile
	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	req.Header.Set("Authorization", "Bearer "+mcTok.AccessToken)
	resp5, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp5.Body.Close()
	body, _ := io.ReadAll(resp5.Body)
	var profile struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	json.Unmarshal(body, &profile)
	if profile.Name == "" {
		return nil, fmt.Errorf("该微软账号未购买 Minecraft")
	}

	return &Account{
		Type:        "microsoft",
		Username:    profile.Name,
		UUID:        profile.ID,
		AccessToken: mcTok.AccessToken,
	}, nil
}

// LittleSkinLogin LittleSkin 外置登录
// baseURL 例: https://littleskin.cn/api/yggdrasil
func LittleSkinLogin(baseURL, username, password string) (*Account, error) {
	// POST {baseURL}/auth/login
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
		"clientToken": "dmhl-" + username,
	})
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LittleSkin 登录失败: %s", string(b))
	}
	var r struct {
		AccessToken     string `json:"accessToken"`
		SelectedProfile struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"selectedProfile"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &Account{
		Type:        "littleskin",
		Username:    r.SelectedProfile.Name,
		UUID:        r.SelectedProfile.ID,
		AccessToken: r.AccessToken,
		BaseURL:     baseURL,
	}, nil
}
