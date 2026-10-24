package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"dmhl/internal/auth"
	"dmhl/internal/config"
	"dmhl/internal/fabric"
	"dmhl/internal/instance"
	"dmhl/internal/install"
	"dmhl/internal/launcher"
	"dmhl/internal/modrinth"
	"dmhl/internal/mirror"
	"dmhl/internal/mojang"
	"dmhl/internal/skin"
	"dmhl/internal/update"
)

// Server HTTP API 服务
type Server struct {
	Cfg       *config.Config
	InstMgr   *instance.Manager
	AuthStore *auth.Store
	SkinMgr   *skin.Manager
}

func New(cfg *config.Config) *Server {
	return &Server{
		Cfg:       cfg,
		InstMgr:   instance.New(cfg.DataDir),
		AuthStore: auth.NewStore(filepath.Join(cfg.DataDir, "accounts.json")),
		SkinMgr:   skin.New(cfg.DataDir),
	}
}

func (s *Server) provider() mirror.Provider {
	return mirror.Provider{UseMirror: s.Cfg.UseMirror}
}

// Routes 注册路由
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "name": "DMHL", "version": "0.2.0"})
	})
	mux.HandleFunc("/api/versions", s.handleVersions)
	mux.HandleFunc("/api/version/detail", s.handleVersionDetail)
	mux.HandleFunc("/api/instances", s.handleInstances)
	mux.HandleFunc("/api/instances/create", s.handleCreateInstance)
	mux.HandleFunc("/api/instances/install", s.handleInstall)
	mux.HandleFunc("/api/instances/launch", s.handleLaunch)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/auth/offline", s.handleOfflineLogin)
	mux.HandleFunc("/api/accounts", s.handleAccounts)
	mux.HandleFunc("/api/skin/import", s.handleSkinImport)
	mux.HandleFunc("/api/fabric/loaders", s.handleFabricLoaders)
	mux.HandleFunc("/api/modrinth/search", s.handleModrinthSearch)
	mux.HandleFunc("/api/modrinth/download", s.handleModrinthDownload)
	mux.HandleFunc("/api/instances/delete", s.handleDeleteInstance)
	mux.HandleFunc("/api/update/check", s.handleUpdateCheck)
	return mux
}

// Start 启动服务
func (s *Server) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.Cfg.ServerPort)
	log.Printf("[DMHL] API 服务启动: http://%s", addr)
	return http.ListenAndServe(addr, s.Routes())
}

// === 版本 ===
func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
	m, err := mojang.FetchManifest(s.provider())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, m)
}

func (s *Server) handleVersionDetail(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("url")
	if u == "" {
		writeError(w, 400, "缺少 url")
		return
	}
	vd, err := mojang.FetchVersionDetail(u, s.provider())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, vd)
}

// === 实例 ===
func (s *Server) handleInstances(w http.ResponseWriter, r *http.Request) {
	list, _ := s.InstMgr.List()
	if list == nil {
		list = []*instance.Instance{}
	}
	writeJSON(w, list)
}

func (s *Server) handleCreateInstance(w http.ResponseWriter, r *http.Request) {
	var ins instance.Instance
	_ = json.NewDecoder(r.Body).Decode(&ins)
	if ins.ID == "" {
		ins.ID = fmt.Sprintf("inst_%d", os.Getpid())
	}
	if err := s.InstMgr.Create(&ins); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, ins)
}

// POST /api/instances/install { id }
func (s *Server) handleInstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	ins, err := s.InstMgr.Get(req.ID)
	if err != nil {
		writeError(w, 404, "实例不存在")
		return
	}
	manifest, err := mojang.FetchManifest(s.provider())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	var verURL string
	for _, v := range manifest.Versions {
		if v.ID == ins.VersionID {
			verURL = v.URL
			break
		}
	}
	if verURL == "" {
		writeError(w, 404, "未找到版本 "+ins.VersionID)
		return
	}
	vd, err := mojang.FetchVersionDetail(verURL, s.provider())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	gameDir := filepath.Join(s.Cfg.DataDir, "instances", ins.ID)
	im := install.New(gameDir, s.provider())
	if err := im.InstallVersion(vd, func(msg string, done, total int) {
		log.Printf("[DMHL] 安装: %s", msg)
	}); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true, "message": fmt.Sprintf("版本 %s 安装完成", ins.VersionID)})
}

// POST /api/instances/launch { id, username }
func (s *Server) handleLaunch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Username == "" {
		req.Username = "Player"
	}
	ins, err := s.InstMgr.Get(req.ID)
	if err != nil {
		writeError(w, 404, "实例不存在")
		return
	}
	runner := launcher.NewRunner(s.Cfg, s.provider())
	res, err := runner.Do(ins, req.Username)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, res)
}

// === 配置 ===
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, s.Cfg)
		return
	}
	var cfg config.Config
	_ = json.NewDecoder(r.Body).Decode(&cfg)
	s.Cfg = &cfg
	_ = config.Save(configPath(), &cfg)
	writeJSON(w, cfg)
}

// === 账号 ===
func (s *Server) handleOfflineLogin(w http.ResponseWriter, r *http.Request) {
	var req struct{ Username string `json:"username"` }
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Username == "" {
		writeError(w, 400, "缺少 username")
		return
	}
	acc := auth.OfflineLogin(req.Username)
	list, _ := s.AuthStore.Load()
	list = append(list, acc)
	_ = s.AuthStore.Save(list)
	writeJSON(w, acc)
}

func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	list, _ := s.AuthStore.Load()
	if list == nil {
		list = []*auth.Account{}
	}
	writeJSON(w, list)
}

// === 皮肤 ===
func (s *Server) handleSkinImport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
		UUID string `json:"uuid"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	out, err := s.SkinMgr.ImportPNG(req.Path, req.UUID)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": out})
}

// === Fabric ===
func (s *Server) handleFabricLoaders(w http.ResponseWriter, r *http.Request) {
	mc := r.URL.Query().Get("mc")
	if mc == "" {
		writeError(w, 400, "缺少 mc 参数")
		return
	}
	list, err := fabric.ListLoaders(mc, s.provider())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, list)
}

// === Modrinth ===
func (s *Server) handleModrinthSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	mc := r.URL.Query().Get("mc")
	loader := r.URL.Query().Get("loader")
	hits, err := modrinth.Search(q, mc, loader, 20)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, hits)
}

// POST /api/modrinth/download { slug, mc, loader, instanceId }
func (s *Server) handleModrinthDownload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Slug       string `json:"slug"`
		MC         string `json:"mc"`
		Loader     string `json:"loader"`
		InstanceID string `json:"instanceId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	versions, err := modrinth.ListVersions(req.Slug, req.MC, req.Loader)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if len(versions) == 0 {
		writeError(w, 404, "无匹配版本")
		return
	}
	v := versions[0]
	if len(v.Files) == 0 {
		writeError(w, 404, "无下载文件")
		return
	}
	fileURL := v.Files[0].URL
	fileName := v.Files[0].Filename
	modsDir := filepath.Join(s.Cfg.DataDir, "instances", req.InstanceID, "mods")
	_ = os.MkdirAll(modsDir, 0755)
	dest := filepath.Join(modsDir, fileName)
	if _, err := os.Stat(dest); err == nil {
		writeJSON(w, map[string]any{"ok": true, "path": dest, "message": "已存在"})
		return
	}
	resp, err := http.Get(fileURL)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer resp.Body.Close()
	f, _ := os.Create(dest)
	defer f.Close()
	_, _ = io.Copy(f, resp.Body)
	writeJSON(w, map[string]any{"ok": true, "path": dest, "message": "mod 已下载"})
}

// POST /api/instances/delete { id }
func (s *Server) handleDeleteInstance(w http.ResponseWriter, r *http.Request) {
	var req struct{ ID string `json:"id"` }
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := s.InstMgr.Delete(req.ID); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// GET /api/update/check
func (s *Server) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	gi, _ := update.CheckGitHub()
	ge, _ := update.CheckGitee()
	writeJSON(w, map[string]any{
		"current": update.CurrentVersion(),
		"github":  gi,
		"gitee":   ge,
	})
}

// helpers
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	writeJSON(w, map[string]any{"error": msg})
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dmhl", "config.json")
}
